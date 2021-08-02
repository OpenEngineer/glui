package glui

import (
  "fmt"
  "reflect"

  "github.com/veandco/go-sdl2/sdl"
)

type animationEvent struct {
  force bool
  tick  uint64
}

// runs on main loop, must handle quit, and wm events
func (app *App) detectUserEvents() error {
  running := true
  for running {
    event_ := sdl.WaitEvent()

    switch event := event_.(type) {
    case *sdl.SysWMEvent:
      if err := HandleSysWMEvent(app, event); err != nil{
        return err
      }
      break
    case *sdl.QuitEvent:
      running = false
      break
    default:
      app.eventCh <- event_
      break
    }
  }

  return nil

}

func (app *App) getScreenSize() (int, int, error) {
  dm, err := sdl.GetCurrentDisplayMode(0)
  if err != nil {
    return 0, 0, err
  }

  return int(dm.W), int(dm.H), nil
}

func (app *App) emitAnimationTicks() {
  someOffscreenBecameVisible := func(oldX int, x int, w int, W int) bool {
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

    return b
  }

  var tick uint64 = 0

  for true {
    x, y := app.window.GetPosition()
    w := app.dd.W
    h := app.dd.H
    W, H, err := app.getScreenSize()
    if err != nil {
      panic(err)
    }

    bX := someOffscreenBecameVisible(app.x, int(x), int(w), int(W))
    bY := someOffscreenBecameVisible(app.y, int(y), int(h), int(H))

    app.x = int(x)
    app.y = int(y)

    event := &animationEvent{bX || bY, tick}

    app.eventCh <- event

    tick++

    sdl.Delay(ANIMATION_LOOP_INTERVAL)
  }
}

func (app *App) initEventLoop() {
  for {
    event_ := <- app.eventCh

    switch event := event_.(type) {
    case *animationEvent:
      app.body.OnTick(event.tick)

      if event.force {
        app.Draw()
      } else {
        app.DrawIfDirty()
      }

      break
    case *sdl.MouseMotionEvent:
      if !app.state.outside {
        app.UpdateActive(int(event.X), int(event.Y))
      }
      break
    case *sdl.MouseButtonEvent:
      if event.Button == sdl.BUTTON_LEFT {
        if event.Type == sdl.MOUSEBUTTONDOWN {
          app.state.lastDown = app.state.active
          app.TriggerEvent("mousedown", NewMouseEvent(int(event.X), int(event.Y)))
        } else if event.Type == sdl.MOUSEBUTTONUP {
          if !app.state.outside {
            if app.state.lastDown != nil {
              tmp := app.state.active
              app.state.active = app.state.lastDown

              app.TriggerEvent("mouseup", NewMouseEvent(int(event.X), int(event.Y)))

              app.state.active = tmp
            }

            app.TriggerEvent("mouseup", NewMouseEvent(int(event.X), int(event.Y)))
          } else {
            app.state.active = app.state.lastDown

            app.TriggerEvent("mouseup", NewMouseEvent(int(event.X), int(event.Y)))

            app.state.active = nil
          }
          app.state.lastDown = nil
        }
      }
      break
    case *sdl.TextInputEvent:
      app.body.IncrementBGColor()
      app.Draw()
      break
    case *sdl.KeyboardEvent:
      break
    case *sdl.WindowEvent:
      switch event.Event {
      case sdl.WINDOWEVENT_SHOWN:
        app.OnShowOrResize()
        break
      case sdl.WINDOWEVENT_EXPOSED:
        app.OnShowOrResize()
        break
      case sdl.WINDOWEVENT_RESIZED:
        app.OnShowOrResize()
        break
      case sdl.WINDOWEVENT_MAXIMIZED:
        app.OnShowOrResize()
        break
      case sdl.WINDOWEVENT_RESTORED:
        app.OnShowOrResize()
        break
      case sdl.WINDOWEVENT_LEAVE:
        if !app.state.outside {
          app.TriggerEvent("mouseleave", NewMouseEvent(-1, -1))
        }
        app.state.active = nil
        app.state.outside = true
        app.state.cursor = -1
        break
      case sdl.WINDOWEVENT_ENTER:
        if app.mouseInWindow() {
          app.state.outside = false
          app.UpdateActive(-1, -1)
        }
        break
      }
    default:
      fmt.Println("event: ", reflect.TypeOf(event_).String())
    }
  }
}
