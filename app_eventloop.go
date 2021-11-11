package glui

import (
  "C"

  "fmt"
  "math"
  "reflect"
  "sync"

  "github.com/veandco/go-sdl2/sdl"
)

type animationEvent struct {
  force bool
  tick  uint64
}

// runs on main loop, must handle quit, and wm events
func (app *App) forwardSystemAndUserEvents() error {
  running := true
  for running {
    event_ := sdl.WaitEvent()

    switch event := event_.(type) {
    case *sdl.SysWMEvent:
      if err := HandleSysWMEvent(app, event); err != nil{
        return err
      }
    case *sdl.QuitEvent:
      app.eventCh <- event_
      running = false
      delay(START_DELAY) // give the draw loop some time to exit cleanly
    default:
      app.eventCh <- event_
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

func (app *App) getWindowSize() (int, int) {
  return app.winW, app.winH
}

func (app *App) syncWindowSize() {
  w_, h_ := app.window.GLGetDrawableSize()
  w, h := int(w_), int(h_)

  app.winW, app.winH = w, h

  for i, frame := range app.frames {
    if i == 0 {
      frame.maxW, frame.maxH = w, h
    }

    frame.syncWindowSize(w, h)
  }
}

func (app *App) emitAnimationEvents() {
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
    w, h  := app.winW, app.winH
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

    delay(ANIMATION_LOOP_INTERVAL)
  }
}

func (app *App) initMainEventLoop(m *sync.Mutex) {
  app.initDrawLoop(m)

  Outer:
  for {
    event_ := <- app.eventCh

    frame := app.ActiveFrame()

    switch event := event_.(type) {
    case *animationEvent:
      app.onTick(event)
    case *sdl.MouseMotionEvent:
      app.onMouseMove(event)
    case *sdl.MouseButtonEvent:
      if frame.state.blockNextMouseButtonEvent {
        frame.state.blockNextMouseButtonEvent = false
      } else {
        evt := NewMouseEvent(int(event.X), int(event.Y))

        if frame.Menu.Visible() && !frame.Menu.IsHit(int(event.X), int(event.Y)) && !frame.Menu.IsOwnedBy(frame.state.mouseElement) {
          if hasEvent(frame.Menu.anchor, "mousebuttonoutsidemenu") {
            TriggerEvent(frame.Menu.anchor, "mousebuttonoutsidemenu", evt)
          } else {
            frame.Menu.Hide()
          }
        }

        if !evt.stopPropagation { // TODO: also propagate to subsequent MOUSEBUTTONUP
          if event.Type == sdl.MOUSEBUTTONDOWN {
            app.onMouseDown(event)
          } else if event.Type == sdl.MOUSEBUTTONUP {
            app.onMouseUp(event)
          }
        } else if event.Type == sdl.MOUSEBUTTONDOWN {
          frame.state.blockNextMouseButtonEvent = true
        }
      }
    case *sdl.MouseWheelEvent:
      app.onMouseWheel(event)
    case *sdl.TextInputEvent:
      app.onTextInput(event)
    case *sdl.KeyboardEvent:
      // tab and shift-tab cycle through the focusable elements
      if event.Keysym.Sym == sdl.K_TAB && event.State == sdl.PRESSED {
        app.onTab(event)
      } else if event.Keysym.Sym == sdl.K_F4 && event.State == sdl.PRESSED && (event.Keysym.Mod & sdl.KMOD_ALT > 0) {
        Quit() // which throws another event!
      } else if event.Keysym.Sym == sdl.K_q && event.State == sdl.PRESSED && (event.Keysym.Mod & sdl.KMOD_CTRL > 0) {
        Quit() // which throws another event!
      } else {
        app.onKeyPress(event)
      }
    case *sdl.WindowEvent:
      switch event.Event {
      case sdl.WINDOWEVENT_SHOWN:
        app.onShowOrResize()
      case sdl.WINDOWEVENT_EXPOSED:
        app.onShowOrResize()
      case sdl.WINDOWEVENT_RESIZED:
        app.onShowOrResize()
      case sdl.WINDOWEVENT_MAXIMIZED:
        app.onShowOrResize()
      case sdl.WINDOWEVENT_RESTORED:
        app.onShowOrResize()
      case sdl.WINDOWEVENT_FOCUS_LOST:
        app.onBlur()
      case sdl.WINDOWEVENT_FOCUS_GAINED:
        app.onFocus()
      case sdl.WINDOWEVENT_LEAVE:
        frame.state.blockNextMouseButtonEvent = false
        app.onLeave()
      case sdl.WINDOWEVENT_ENTER:
        app.onEnter()
      }
    case *sdl.QuitEvent:
      // TODO: optionally catch this event with an eventlistener that can still decide not to quit
      break Outer
    default:
      fmt.Println("unhandled event ", reflect.TypeOf(event_).String())
    }

    app.DrawIfDirty()
  }

  app.endDrawLoop()
}

func (app *App) onTick(event *animationEvent) {
  // only animate the active frame, but update all the frame states
  for i, frame := range app.frames {
    frame.state.lastTick = event.tick

    if i == app.activeFrame {
      frame.Animate(event.tick)

      if event.force {
        frame.ForcePosDirty()
      } 
    }
  }
}

func (app *App) onMouseMove(event *sdl.MouseMotionEvent) {
  frame := app.ActiveFrame()

  if !frame.state.outside {
    app.updateMouseElement(int(event.X), int(event.Y), int(event.XRel), int(event.YRel))
  }
}

func (app *App) onMouseWheel(event *sdl.MouseWheelEvent) {
  frame := app.ActiveFrame()

  if !frame.state.outside && elementNotNil(frame.state.mouseElement) {
    // TODO: smart scaling depending on the platform
    // TODO: smart direction depending on the platform
    TriggerEvent(frame.state.mouseElement, "wheel", NewMouseWheelEvent(-int(event.X)*5, -int(event.Y)*5))
  }
}

func (app *App) onMouseDown(event *sdl.MouseButtonEvent) {
  frame := app.ActiveFrame()

  if frame.state.mouseElement == frame.Menu || frame.state.mouseElement == nil {
    // eg. on edge of menu
    return
  }

  frame.state.lastDown = frame.state.mouseElement

  frame.state.lastDownX = int(event.X)
  frame.state.lastDownY = int(event.Y)
  frame.state.mouseMoveSumX = 0
  frame.state.mouseMoveSumY = 0

  if event.Button == sdl.BUTTON_LEFT {
    app.triggerHitEvent("mousedown", NewMouseEvent(int(event.X), int(event.Y)))
  } else if event.Button == sdl.BUTTON_RIGHT {
    app.triggerHitEvent("rightmousedown", NewMouseEvent(int(event.X), int(event.Y)))
  }

  if !hasAncestor(frame.state.mouseElement, frame.Menu) {
    newFocusable := findFocusable(frame.state.mouseElement)

    blurEvt := NewMouseEvent(int(event.X), int(event.Y))
    focusEvt := NewMouseEvent(int(event.X), int(event.Y))

    app.changeFocusElement(newFocusable, blurEvt, focusEvt)
  }
}

func (app *App) onMouseUp(event *sdl.MouseButtonEvent) {
  frame := app.ActiveFrame()

  fnTrigger := func() {
    if event.Button == sdl.BUTTON_LEFT {
      evt := NewMouseEvent(int(event.X), int(event.Y))
      app.triggerHitEvent("mouseup", evt)

      if !evt.stopPropagation {
        app.detectClick(int(event.X), int(event.Y)) // turn mouseup into click, doubleclick or tripleclick
      }
    } else if event.Button == sdl.BUTTON_RIGHT {
      evt := NewMouseEvent(int(event.X), int(event.Y))
      app.triggerHitEvent("rightmouseup", evt)

      if !evt.stopPropagation {
        app.triggerHitEvent("rightclick", NewMouseEvent(int(event.X), int(event.Y)))
      }
    }
  }

  if !frame.state.outside {
    // lastdown also gets a mouseup/rightclick event
    if elementNotNil(frame.state.lastDown) && frame.state.lastDown != frame.state.mouseElement {
      tmp := frame.state.mouseElement
      frame.state.mouseElement = frame.state.lastDown

      fnTrigger()

      frame.state.mouseElement = tmp
    }

    fnTrigger()
  } else {
    frame.state.mouseElement = frame.state.lastDown

    fnTrigger()

    frame.state.mouseElement = nil
  }

  frame.state.mouseMoveSumX = 0
  frame.state.mouseMoveSumY = 0
  frame.state.lastDown = nil
}

// turn a mouseup into a click, doubleclick or tripleclick
func (app *App) detectClick(x, y int) {
  frame := app.ActiveFrame()

  xPrev := frame.state.lastUpX
  yPrev := frame.state.lastUpY

  // use manhattan distance
  dr := math.Abs(float64(xPrev - x)) + math.Abs(float64(yPrev - y)) 
  dt := math.Abs(float64(frame.state.lastTick - frame.state.lastUpTick))
  if dr < 1.0 && dt < 15.0 {
    frame.state.upCount = frame.state.upCount + 1
  } else {
    frame.state.upCount = 1
  }

  frame.state.lastUpTick = frame.state.lastTick
  frame.state.lastUpX = x
  frame.state.lastUpY = y
  frame.state.mouseMoveSumX = 0
  frame.state.mouseMoveSumY = 0

  eName := "click"
  switch frame.state.upCount {
  case 0:
    panic("shouldn't be possible")
  case 1:
  case 2:
    if elementNotNil(frame.state.mouseElement) {
      if ancestorHasEvent(frame.state.mouseElement, "doubleclick") {
        eName = "doubleclick"
      }
    }
  default:
    if elementNotNil(frame.state.mouseElement) {
      if ancestorHasEvent(frame.state.mouseElement, "tripleclick") {
        eName = "tripleclick"
      }
    }
  }

  TriggerEvent(frame.state.mouseElement, eName, NewMouseEvent(x, y))
}

func (app *App) onTextInput(event *sdl.TextInputEvent) {
  frame := app.ActiveFrame()

  str := event.GetText()

  TriggerEvent(frame.state.focusElement, "textinput", NewTextInputEvent(str))
}

func (app *App) onTab(event *sdl.KeyboardEvent) {
  frame := app.ActiveFrame()

  app.hideMenuIfVisible()

  var newFocusable Element

  shift := event.Keysym.Mod & sdl.KMOD_SHIFT > 0

  if elementNotNil(frame.state.focusElement) && !frame.FocusRect.IsOwnedBy(frame.state.focusElement) {
    newFocusable = frame.state.focusElement
  } else {
    if shift {
      if elementNotNil(frame.state.focusElement) {
        newFocusable = findPrevFocusable(frame.state.focusElement)
      } else {
        newFocusable = findPrevFocusable(frame.Body)
      }
    } else {
      if elementNotNil(frame.state.focusElement) {
        newFocusable = findNextFocusable(frame.state.focusElement)
      } else {
        newFocusable = findNextFocusable(frame.Body)
      }
    }
  }

  blurEvt := NewKeyboardEvent("tab", false, shift, false)
  focusEvt := NewKeyboardEvent("tab", false, shift, false)

  app.changeFocusElement(newFocusable, blurEvt, focusEvt)
}

func (app *App) onKeyPress(event *sdl.KeyboardEvent) {
  frame := app.ActiveFrame()

  eType, kType, ctrl, shift, alt := extractKeyboardEventDetails(event)
  if eType != "" && kType != "" {
    TriggerEvent(frame.state.focusElement, eType, NewKeyboardEvent(kType, ctrl, shift, alt))

    if eType == "keydown" {
      TriggerEvent(frame.state.focusElement, "keypress", NewKeyboardEvent(kType, ctrl, shift, alt))
    }
  } else {
    fmt.Println("unhandled keyboardevent ", event.Keysym.Sym)
  }
}

func (app *App) onShowOrResize() {
  app.syncWindowSize()

  for i, frame := range app.frames {
    if i <= app.activeFrame {
      frame.ForcePosDirty()
    }
  }
}

func (app *App) onBlur() {
  frame := app.ActiveFrame()

  TriggerEvent(frame.state.focusElement, "blur", NewMouseEvent(-1, -1))
}

func (app *App) onFocus() {
  frame := app.ActiveFrame()

  TriggerEvent(frame.state.focusElement, "focus", NewMouseEvent(-1, -1))
}

func (app *App) onLeave() {
  frame := app.ActiveFrame()

  if !frame.state.outside {
    app.triggerHitEvent("mouseleave", NewMouseEvent(-1, -1))
  }

  frame.state.mouseElement = nil
  frame.state.outside = true
  frame.state.cursor = -1
}

func (app *App) onEnter() {
  frame := app.ActiveFrame()

  if app.mouseInWindow() {
    frame.state.outside = false
    app.updateMouseElement(-1, -1, 0, 0)
  }
}
