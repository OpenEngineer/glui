package glui

import (
  "fmt"
  "os"
  "sync"

  "github.com/veandco/go-sdl2/sdl"
)

const (
  START_DELAY             = 10 // ms
  EVENT_LOOP_INTERVAL     = 13 // ms
  ANIMATION_LOOP_INTERVAL = 16 // ms
  RENDER_LOOP_INTERVAL    = 16 // ms
)

type App struct {
  name string

  x int
  y int

  drawCh  chan bool
  eventCh chan interface{}

  root   *Root
  window *sdl.Window
  framebuffers [2]uint32 // for windows thumbnail drawing
  program1 uint32
  program2 uint32

  ctx    sdl.GLContext
  debug  *os.File

  state   AppState
}

type AppState struct {
  mouseElement   Element
  focusElement   Element
  cursor         int
  lastDown       Element
  outside        bool
  lastUpX        int
  lastUpY        int
  upCount        int // limited to three
  lastTick       uint64
  lastUpTick     uint64
}

func newAppState() AppState {
  return AppState{
    nil,
    nil,
    -1,
    nil,
    false,
    0,0,0,
    0,0,
  }
}

func NewApp(name string, skin Skin, glyphs map[string]*Glyph) *App {
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
  root := newRoot(skinMap, glyphMap)

  return &App{
    name,
    0, 0,
    make(chan bool),
    make(chan interface{}),
    root,
    nil,
    [2]uint32{0, 0},
    0,
    0,
    nil,
    debug,
    newAppState(),
  }
}

func (app *App) Root() *Root {
  return app.root
}

func (app *App) Body() *Body {
  return app.root.Body
}

func (app *App) Run() {
  if err := app.run(); err != nil {
    fmt.Fprintf(os.Stderr, "%s\n", err.Error())
  }
}

func (app *App) run() error {
  if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
    return err
  }

  defer sdl.Quit()

  var err error
  app.window, err = sdl.CreateWindow(
    app.name, 
    sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
    0, 0, 
    sdl.WINDOW_SHOWN | sdl.WINDOW_RESIZABLE | sdl.WINDOW_OPENGL | sdl.WINDOW_MAXIMIZED | sdl.WINDOW_ALLOW_HIGHDPI,
  )
  if err != nil {
    return err
  }

  defer app.window.Destroy()

  app.window.SetMinimumSize(1024, 768)

  if err := InitOS(app.window); err != nil {
    return err
  }

  sdl.Delay(START_DELAY)

  app.window.Maximize()

  m := &sync.Mutex{}

  go func(m_ *sync.Mutex) {
    app.initDrawLoop(m)
  }(m)

  sdl.Delay(START_DELAY)

  m.Lock()

  m.Unlock()

  go func() {
    app.emitAnimationEvents()
  }()

  // both animation and system/user events are serialized by a separate thread
  go func() {
    app.initMainEventLoop()
  }()

  // here we are in the main thread and must this thread must be used to detect 
  //  system and user events, which are forwarded into the main event loop (separate thread)
  return app.forwardSystemAndUserEvents()
}
