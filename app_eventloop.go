package glui

import (
  "C"

  "fmt"
  "math"
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
      app.state.lastTick = event.tick
      app.body.OnTick(event.tick)

      if event.force {
        app.Draw()
      } else {
        app.DrawIfDirty()
      }

      break
    case *sdl.MouseMotionEvent:
      if !app.state.outside {
        app.updateMouseElement(int(event.X), int(event.Y))

        if app.state.mouseElement != nil {
          if app.state.lastDown != nil && app.state.lastDown != app.state.mouseElement {
            app.triggerEvent(app.state.lastDown, "mousemove", NewMouseEvent(int(event.X), int(event.Y)))
          }

          app.triggerHitEvent("mousemove", NewMouseEvent(int(event.X), int(event.Y)))
        }
      }
      break
    case *sdl.MouseButtonEvent:
      if app.dd.Menu.isVisible() && !app.dd.Menu.Hit(int(event.X), int(event.Y)) {
        app.dd.Menu.Hide()
      }

      if event.Type == sdl.MOUSEBUTTONDOWN {
        app.state.lastDown = app.state.mouseElement

        if event.Button == sdl.BUTTON_LEFT {
          app.triggerHitEvent("mousedown", NewMouseEvent(int(event.X), int(event.Y)))
        } else if event.Button == sdl.BUTTON_RIGHT {
          app.triggerHitEvent("rightclick", NewMouseEvent(int(event.X), int(event.Y)))
        }

        newFocusable := findFocusable(app.state.mouseElement)

        blurEvt := NewMouseEvent(int(event.X), int(event.Y))
        focusEvt := NewMouseEvent(int(event.X), int(event.Y))

        app.changeFocusElement(newFocusable, blurEvt, focusEvt)
      } else if event.Type == sdl.MOUSEBUTTONUP {
        if !app.state.outside && event.Button == sdl.BUTTON_LEFT {
          if app.state.lastDown != nil && app.state.lastDown != app.state.mouseElement {
            tmp := app.state.mouseElement
            app.state.mouseElement = app.state.lastDown

            app.triggerHitEvent("mouseup", NewMouseEvent(int(event.X), int(event.Y)))

            app.state.mouseElement = tmp
          }

          app.triggerHitEvent("mouseup", NewMouseEvent(int(event.X), int(event.Y)))

          app.handleMouseUp(int(event.X), int(event.Y))
        } else {
          if event.Button == sdl.BUTTON_LEFT {
            app.state.mouseElement = app.state.lastDown

            app.triggerHitEvent("mouseup", NewMouseEvent(int(event.X), int(event.Y)))

            app.handleMouseUp(int(event.X), int(event.Y))
          }

          app.state.mouseElement = nil
        }

        app.state.lastDown = nil
      }

      break
    case *sdl.TextInputEvent:
      str := event.GetText()

      app.triggerEvent(app.state.focusElement, "textinput", NewTextInputEvent(str))

      break
    case *sdl.KeyboardEvent:
      // tab and shift-tab cycle through the focusable elements
      if event.Keysym.Sym == sdl.K_TAB && event.State == sdl.PRESSED {
        var newFocusable Element

        shift := event.Keysym.Mod & sdl.KMOD_SHIFT > 0
        if shift {
          if app.state.focusElement != nil {
            newFocusable = findPrevFocusable(app.state.focusElement)
          } else {
            newFocusable = findPrevFocusable(app.body)
          }
        } else {
          if app.state.focusElement != nil {
            newFocusable = findNextFocusable(app.state.focusElement)
          } else {
            newFocusable = findNextFocusable(app.body)
          }
        }

        blurEvt := NewKeyboardEvent("tab", false, shift, false)
        focusEvt := NewKeyboardEvent("tab", false, shift, false)

        app.changeFocusElement(newFocusable, blurEvt, focusEvt)
      } else {
        eType, kType, ctrl, shift, alt := extractKeyboardEventDetails(event)
        if eType != "" && kType != "" {
          app.triggerEvent(app.state.focusElement, eType, NewKeyboardEvent(kType, ctrl, shift, alt))

          if eType == "keydown" {
            app.triggerEvent(app.state.focusElement, "keypress", NewKeyboardEvent(kType, ctrl, shift, alt))
          }
        } else {
          fmt.Println("unhandled keyboardevent: ", event.Keysym.Sym)
        }
      }

      app.DrawIfDirty()

      break
    case *sdl.WindowEvent:
      switch event.Event {
      case sdl.WINDOWEVENT_SHOWN:
        app.onShowOrResize()
        break
      case sdl.WINDOWEVENT_EXPOSED:
        app.onShowOrResize()
        break
      case sdl.WINDOWEVENT_RESIZED:
        app.onShowOrResize()
        break
      case sdl.WINDOWEVENT_MAXIMIZED:
        app.onShowOrResize()
        break
      case sdl.WINDOWEVENT_RESTORED:
        app.onShowOrResize()
        break
      case sdl.WINDOWEVENT_FOCUS_LOST:
        app.triggerEvent(app.state.focusElement, "blur", NewMouseEvent(-1, -1))
        break
      case sdl.WINDOWEVENT_FOCUS_GAINED:
        app.triggerEvent(app.state.focusElement, "focus", NewMouseEvent(-1, -1))
        break
      case sdl.WINDOWEVENT_LEAVE:
        if !app.state.outside {
          app.triggerHitEvent("mouseleave", NewMouseEvent(-1, -1))
        }

        app.state.mouseElement = nil
        app.state.outside = true
        app.state.cursor = -1

        break
      case sdl.WINDOWEVENT_ENTER:
        if app.mouseInWindow() {
          app.state.outside = false
          app.updateMouseElement(-1, -1)
        }
        break
      }

      app.DrawIfDirty()
    default:
      fmt.Println("event: ", reflect.TypeOf(event_).String())
    }
  }
}

func (app *App) changeFocusElement(newFocusable Element, blurEvt, focusEvt *Event) {
  if newFocusable != app.state.focusElement {
    if app.state.focusElement != nil {
      app.triggerEvent(app.state.focusElement, "blur", blurEvt)
    }

    app.state.focusElement = newFocusable

    if newFocusable != nil {
      app.triggerEvent(app.state.focusElement, "focus", focusEvt)
    }
  }
}

// turn a mouseup into a click, doubleclick or tripleclick
func (app *App) handleMouseUp(x, y int) {
  xPrev := app.state.lastUpX
  yPrev := app.state.lastUpY

  // use manhattan distance
  dr := math.Abs(float64(xPrev - x)) + math.Abs(float64(yPrev - y)) 
  dt := math.Abs(float64(app.state.lastTick - app.state.lastUpTick))
  if dr < 1.0 && dt < 15.0 {
    app.state.upCount = app.state.upCount + 1
  } else {
    app.state.upCount = 1
  }

  app.state.lastUpTick = app.state.lastTick
  app.state.lastUpX = x
  app.state.lastUpY = y

  var eName string
  switch app.state.upCount {
  case 0:
    panic("shouldn't be possible")
  case 1:
    eName = "click"
    break
  case 2:
    eName = "doubleclick"
    break
  default:
    eName = "tripleclick"
  }

  app.triggerEvent(app.state.mouseElement, eName, NewMouseEvent(x, y))
}

func (app *App) triggerEvent(e Element, name string, evt *Event) {
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

func (app *App) triggerHitEvent(name string, evt *Event) {
  app.triggerEvent(app.state.mouseElement, name, evt)
}

func (app *App) onShowOrResize() {
  app.dd.SyncSize(app.window)

  app.body.OnResize(app.dd.W, app.dd.H)

  if app.mouseInWindow() {
    app.updateMouseElement(-1, -1)
  }

  app.DrawIfDirty()
}
