package glui

import (
  "errors"
  "fmt"
  "os"
  "sync"

  "github.com/veandco/go-sdl2/sdl"
  "github.com/go-gl/gl/v4.1-core/gl"
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

  drawCh chan bool
  eventCh chan interface{}
  body   *Body
  window *sdl.Window
  framebuffers [2]uint32 // for windows thumbnail drawing
  program1 uint32
  program2 uint32

  ctx    sdl.GLContext
  debug  *os.File

  dd      *DrawData
  state   AppState
}

type AppState struct {
  active   Element
  cursor   int
  lastDown Element
  outside  bool
}

func newAppState() AppState {
  return AppState{
    nil,
    -1,
    nil,
    false,
  }
}

func NewApp(name string, skin Skin, glyphs map[string]*Glyph) *App {
  debug, err := os.Create(name + ".log")
  if err != nil {
    panic(err)
  }

  fmt.Fprintf(debug, "#starting log\n")

  body := NewBody()

  if glyphs == nil {
    glyphs = make(map[string]*Glyph)
  }

  return &App{
    name,
    0, 0,
    make(chan bool),
    make(chan interface{}),
    body,
    nil,
    [2]uint32{0, 0},
    0,
    0,
    nil,
    debug,
    NewDrawData(skin, glyphs),
    newAppState(),
  }
}

func (app *App) Run() {
  if err := app.run(); err != nil {
    fmt.Fprintf(os.Stderr, "%s\n", err.Error())
  }
}

func (app *App) Body() *Body {
  return app.body
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
    app.initRenderLoop(m)
  }(m)

  sdl.Delay(START_DELAY)

  m.Lock()

  m.Unlock()

  go func() {
    app.emitAnimationTicks()
  }()

  go func() {
    app.initEventLoop() // serializes all the events
  }()

  // this is the main thread and must be used to detect user events
  return app.detectUserEvents()
}

func (app *App) UpdateActive(x, y int) {
  if x < 0 {
    x_, y_, _ := sdl.GetMouseState()

    x = int(x_)
    y = int(y_)
  }

  oldActive := app.state.active
  if oldActive == nil {
    oldActive = app.body
  }

  newActive, isSameOrChildOfOld := findActive(oldActive, x, y)

  // trigger mouse leave event if new active isn't child of old
  if app.state.active != nil && !isSameOrChildOfOld {
    evt := NewMouseEvent(x, y)
    if app.state.active.Parent() != nil {
      evt.stopBubblingWhenElementReached(app.state.active.Parent())
    }
    app.TriggerEvent("mouseleave", evt)
  }

  if app.state.active == nil {
    evt := NewMouseEvent(x, y)
    app.state.active = newActive
    app.TriggerEvent("mouseenter", evt)
  } else if app.state.active != newActive {
    evt := NewMouseEvent(x, y)

    ca := commonAncestor(app.state.active, newActive)

    evt.stopBubblingWhenElementReached(ca)

    app.state.active = newActive
    if ca != newActive {
      app.TriggerEvent("mouseenter", evt)
    }
  }

  if app.state.active.Cursor() != app.state.cursor {
    app.state.cursor = app.state.active.Cursor()

    if app.state.cursor >= 0 && app.state.cursor < sdl.NUM_SYSTEM_CURSORS {
      sdl.ShowCursor(sdl.ENABLE)

      oldCursor := sdl.GetCursor()

      c := sdl.CreateSystemCursor((sdl.SystemCursor)(app.state.cursor))

      sdl.SetCursor(c)

      sdl.FreeCursor(oldCursor) // free the previous
    } else {
      panic("not custom cursors defined yet")
    }
  }
}

// the window enter or leave events might be called spuriously
func (app *App) mouseInWindow() bool {
  x0, y0 := app.window.GetPosition()
  w, h := app.window.GetSize()

  x, y, _ := sdl.GetGlobalMouseState()

  r := Rect{int(x0), int(y0), int(w), int(h)}

  b := r.Hit(int(x), int(y))

  return b
}

func (app *App) OnShowOrResize() {
  app.dd.SyncSize(app.window)

  app.body.OnResize(app.dd.W, app.dd.H)

  if app.mouseInWindow() {
    app.UpdateActive(-1, -1)
  }

  app.DrawIfDirty()
}

func (app *App) TriggerEvent(name string, evt *Event) {
  //fmt.Printf("triggering %s event, %d %d, %p\n", name, evt.X, evt.Y, app.state.active)

  e := app.state.active 
  for e != nil {
    l := e.GetEventListener(name)
    
    if l != nil {
      l(evt)
    }

    if evt.stopBubbling {
      break
    }

    // bubble
    e = e.Parent()

    if evt.stopBubblingElement == e {
      break
    }
  }

  app.DrawIfDirty()
}

func (app *App) DrawIfDirty() {
  if app.dd.Dirty() {
    app.Draw()
  }
}

func (app *App) Draw() {
  app.drawCh <- true
}

func (app *App) initRenderLoop(m *sync.Mutex) {
  m.Lock()

  ctx, err := app.window.GLCreateContext()
  if err != nil {
    fmt.Fprintf(app.debug, "unable to create context: %s\n", err.Error())
    panic(err)
  }

  app.ctx = ctx

  if err := app.window.GLMakeCurrent(ctx); err != nil {
    fmt.Fprintf(app.debug, "unable to make current in render: %s\n", err.Error())
    panic(err)
  }

  if err := gl.Init(); err != nil {
    fmt.Fprintf(app.debug, "render gl.Init error: %s\n", err.Error())
    panic(err)
  }

  glVersion := gl.GoStr(gl.GetString(gl.VERSION))
  if glVersion == "" {
    err := errors.New("empty OpenGL version")
    fmt.Fprintf(app.debug, "%s\n", err.Error())
    panic(err)
  }

  app.program1, err = compileProgram1()
  if err != nil {
    fmt.Fprintf(app.debug, "failed to compile OpenGL program1: %s\n", err.Error())
    panic(err)
  }

  app.program2, err = compileProgram2()
  if err != nil {
    fmt.Fprintf(app.debug, "failed to compile OpenGL program2: %s\n", err.Error())
    panic(err)
  }

  app.dd.InitGL(app.program1, app.program2)

  //gl.CreateFramebuffers(1, &(app.framebuffers[0]))
  //gl.CreateFramebuffers(1, &(app.framebuffers[1]))

  gl.GenFramebuffers(1, &(app.framebuffers[0]))
  gl.GenFramebuffers(1, &(app.framebuffers[1]))

  x, y := app.window.GetPosition()
  app.x = int(x)
  app.y = int(y)

  if err := app.window.GLMakeCurrent(nil); err != nil {
    fmt.Fprintf(app.debug, "unable to unmake current in render: %s\n", err.Error())
    return
  }

  //app.draw()

  m.Unlock()

  for true {
    //fmt.Println("in render loop")
    //for _ = range app.drawCh {
    //}
    <- app.drawCh

    app.draw()

    sdl.Delay(RENDER_LOOP_INTERVAL)

    /*draining := true
    for draining {
      select {
      case <- app.drawCh:
        continue
      default:
        draining = false
        break
      }
    }*/

    //}
    //default:
    //fmt.Println("nothing done in render loop")
  }

  sdl.GLDeleteContext(ctx)
}

func (app *App) DrawData() *DrawData {
  return app.dd
}
