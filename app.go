package glui

import (
  "fmt"
  "os"
  "reflect"
  "sync"

  "github.com/veandco/go-sdl2/sdl"
  "github.com/go-gl/gl/v4.1-core/gl"
)

const (
  START_DELAY = 10 // ms
  EVENT_LOOP_INTERVAL = 13 // ms
  RENDER_LOOP_INTERVAL = 13 // ms
)

type App struct {
  name string

  dirty bool
  x int
  y int

  mutex  *sync.Mutex
  drawCh chan bool
  root   *Root
  window *sdl.Window

  debug  *os.File
}

func NewApp(name string) *App {
  debug, err := os.Create(name + ".log")
  if err != nil {
    panic(err)
  }

  fmt.Fprintf(debug, "#starting log\n")
  return &App{
    name,
    false,
    0, 0,
    &sync.Mutex{},
    make(chan bool),
    NewRoot(),
    nil,
    debug,
  }
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
    sdl.WINDOW_SHOWN | sdl.WINDOW_RESIZABLE | sdl.WINDOW_OPENGL | sdl.WINDOW_MAXIMIZED,
  )
  if err != nil {
    return err
  }

  defer app.window.Destroy()

  app.window.SetMinimumSize(1024, 768)

  sdl.Delay(START_DELAY)

  app.window.Maximize()

  go func() {
    app.render()
  }()

  sdl.Delay(START_DELAY)

  app.mutex.Lock()

  app.mutex.Unlock()

  go func() {
    app.detectOffscreenBecomesVisible()
  }()

  return app.loopEvents()
}

func (app *App) getScreenSize() (int, int, error) {
  dm, err := sdl.GetCurrentDisplayMode(0)
  if err != nil {
    return 0, 0, err
  }

  return int(dm.W), int(dm.H), nil
}

func (app *App) loopEvents() error {
  running := true
  for running {
    event_ := sdl.WaitEvent()

    switch event := event_.(type) {
    case *sdl.MouseMotionEvent: // most common event first
      break

    case *sdl.QuitEvent:
      running = false
      break
    case *sdl.TextInputEvent:
      app.root.IncrementBGColor()
      fmt.Println("sending data into draw ch")
      app.drawCh <- true
      break
    case *sdl.KeyboardEvent:
      break
    case *sdl.WindowEvent:
      if false {
        fmt.Println("window event type: ", event.Type)
      }
      app.drawCh <- true
      break
    default:
      fmt.Println("event: ", reflect.TypeOf(event_).String())
    }

    //sdl.Delay(EVENT_LOOP_INTERVAL)
  }

  return nil
}

func (app *App) detectOffscreenBecomesVisible() {
  for true {
    x, y := app.window.GetPosition()
    w, h := app.window.GetSize()
    W, H, err := app.getScreenSize()
    if err != nil {
      panic(err)
    }

    bX := someOffscreenBecameVisible(app.x, int(x), int(w), int(W))
    bY := someOffscreenBecameVisible(app.y, int(y), int(h), int(H))

    app.x = int(x)
    app.y = int(y)

    if bX || bY {
      fmt.Println("detected offscreen becoming visible")
      app.drawCh <- true
    }

    sdl.Delay(RENDER_LOOP_INTERVAL)
  }
}

func (app *App) render() {
  app.mutex.Lock()

  ctx, err := app.window.GLCreateContext()
  if err != nil {
    fmt.Fprintf(app.debug, "unable to create context: %s\n", err.Error())
    return
  }

  if err := app.window.GLMakeCurrent(ctx); err != nil {
    fmt.Fprintf(app.debug, "unable to make current in render: %s\n", err.Error())
    return
  }

  if err := gl.Init(); err != nil {
    fmt.Fprintf(app.debug, "render gl.Init error: %s\n", err.Error())
    return
  }

  glVersion := gl.GoStr(gl.GetString(gl.VERSION))
  if glVersion == "" {
    fmt.Fprintf(app.debug, "empty OpenGL version\n")
    return
  }

  x, y := app.window.GetPosition()
  app.x = int(x)
  app.y = int(y)

  if err := app.window.GLMakeCurrent(nil); err != nil {
    fmt.Fprintf(app.debug, "unable to unmake current in render: %s\n", err.Error())
    return
  }

  app.draw(ctx)

  app.mutex.Unlock()

  for true {
    //fmt.Println("in render loop")
    //for _ = range app.drawCh {
    //}
    <- app.drawCh

    app.draw(ctx)

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

func someOffscreenBecameVisible(oldX int, x int, w int, W int) bool {
  b := false

  if x < 0.0 && x + w > 0.0 {
    if x > oldX {
      b = true
    }
  }

  if x < W && x + w > W {
    if x < oldX {
      b = true
    }
  }

  // DEBUG
  //if x != oldX {
    //b = true
  //}

  return b
}

func (app *App) draw(ctx sdl.GLContext) {
  if err := app.window.GLMakeCurrent(ctx); err != nil {
    fmt.Fprintf(app.debug, "unable to make current: %s\n", err.Error())
    return
  }

  w, h := app.window.GetSize()
  gl.Viewport(0, 0, w, h)

  //app.root.IncrementBGColor()
  color := app.root.BGColor()

  gl.ClearColor(
    float32(color.R)/float32(256),
    float32(color.G)/float32(256),
    float32(color.B)/float32(256),
    float32(color.A)/float32(256),
  )

  gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

  //fmt.Println("drawing...")

  app.window.GLSwap()

  if err := app.window.GLMakeCurrent(nil); err != nil {
    fmt.Fprintf(app.debug, "unable to unmake current: %s\n", err.Error())
  }
}
