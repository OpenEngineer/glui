package glui

import (
  "fmt"
  "os"
  "sync"

  "github.com/veandco/go-sdl2/sdl"
)

const (
  START_DELAY             = 10 // ms
  ANIMATION_LOOP_INTERVAL = 2*16 // ms
)

// app is stored in global variable because this makes it easier to access to active frame when creating new elements
var _app *App = nil

type App struct {
  name string

  x int
  y int
  winW int
  winH int

  drawCh  chan bool
  eventCh chan interface{}

  window   *sdl.Window
  programs *Programs
  skinMap  *SkinMap
  glyphMap *GlyphMap

  frames      []*Frame
  activeFrame int
  quitPending bool

  ctx    sdl.GLContext
  debug  *os.File
}

func NewApp(name string, skin Skin, glyphs map[string]*Glyph, nFrames int) {
  debug, err := os.Create(name + ".log")
  if err != nil {
    panic(err)
  }

  fmt.Fprintf(debug, "#starting log\n")

  if glyphs == nil {
    glyphs = make(map[string]*Glyph)
  }

  skinMap := newSkinMap(skin)
  glyphMap := newGlyphMap(glyphs)

  if nFrames < 1 {
    panic("need at least one frame")
  }

  frames := make([]*Frame, nFrames)
  for i := 0; i < nFrames; i++ {
    // skinMap and glyphMap are shared across frames
    frames[i] = newFrame(i == 0, skinMap, glyphMap)
  }

  if _app != nil {
    panic("app already initialized")
  }

  // saved in a global variable
  _app = &App{
    name,
    0, 0, 0, 0,
    make(chan bool),
    make(chan interface{}),
    nil,
    &Programs{},
    skinMap,
    glyphMap,
    frames,
    0,
    false,
    nil,
    debug,
  }
}

func getApp() *App {
  if _app == nil {
    panic("app not yet initialized (hint: call NewApp(name, skin, glyphs, nFrames)")
  }

  return _app
}

func (app *App) ActiveFrame() *Frame {
  return app.frames[app.activeFrame]
}

func ActiveFrame() *Frame {
  app := getApp()

  return app.ActiveFrame()
}


func ActiveBody() *Body {
  app := getApp()

  frame := app.ActiveFrame()

  return frame.Body
}

func Run() {
  app := getApp()

  if err := app.run(); err != nil {
    fmt.Fprintf(os.Stderr, "%s\n", err.Error())
  }
}

// additional frames are always displayed in the center
// elements can only be added after this! (otherwise activeFrame is still the previous frame)
func PushFrame(maxW, maxH int) {
  app := getApp()

  app.activeFrame++

  if app.activeFrame >= len(app.frames) {
    panic("not enough frames allocated")
  }

  f := app.ActiveFrame()
  f.maxW, f.maxH = maxW, maxH
}

func PopFrame() {
  app := getApp()

  app.ActiveFrame().Clear()

  app.activeFrame--

  if app.activeFrame < 0 {
    panic("already at base frame")
  }

  newActiveFrame := app.ActiveFrame()

  newActiveFrame.ForceAllDirty()

  app.Draw()
}

func (app *App) run() error {
  if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
    return err
  }

  defer sdl.Quit()

  // can we get the size of the displayable area before creating the window?
  /*dm, err := sdl.GetCurrentDisplayMode(0)
  if err != nil {
    return err
  }*/

  var err error
  app.window, err = sdl.CreateWindow(
    app.name, 
    sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
    0, 0, 
    sdl.WINDOW_RESIZABLE | sdl.WINDOW_OPENGL | sdl.WINDOW_MAXIMIZED, // doesn't work in ratpoison wm manager for some reason,
  )
  if err != nil {
    return err
  }

  app.window.SetMinimumSize(1024, 764)

  defer app.window.Destroy()

  if err := InitOS(app.window); err != nil {
    return err
  }

  // give opengl some time to initialize
  delay(START_DELAY)

  app.ActiveFrame().show()

  m := &sync.Mutex{}

  // both animation and system/user events are serialized by a separate thread
  go func(m_ *sync.Mutex) {
    app.initMainEventLoop(m_)
  }(m)

  // make sure the mutex is locked/unlocked by initMainEventLoop()
  delay(START_DELAY)

  // once we are able to unlock the mutex here, we can start emitting events
  m.Lock()
  m.Unlock()
  go func() {
    app.emitAnimationEvents()
  }()

  // here we are in the main thread and this thread must be used to detect system and user events, 
  // which are forwarded into the main event loop (separate thread)
  return app.forwardSystemAndUserEvents()
}

func (app *App) quit() {
  if app.quitPending {
    return
  }

  app.quitPending = true

  callback := func(args ...interface{}) {
    app.quitPending = false

    if len(args) == 1 {
      if arg, ok := args[0].(bool); ok {
        if arg {
          // XXX: does the timestamp matter?
          sdl.PushEvent(&sdl.QuitEvent{sdl.QUIT, 0})
        }
      }
    }
  }

  for i := app.activeFrame; i >= 0; i-- {
    frame := app.frames[i]

    body := frame.Body

    if hasEvent(body, "quit") {
      evt := NewAppEvent("quit", callback)

      TriggerEvent(body, "quit", evt)

      return 
    }
  }

  callback(true)
}

func Quit() {
  app := getApp()

  app.quit()
}
