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
    w,h  := app.root.GetSize()
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

func (app *App) initMainEventLoop(m *sync.Mutex) {
  app.initDrawLoop(m)

  for {
    event_ := <- app.eventCh

    switch event := event_.(type) {
    case *animationEvent:
      app.onTick(event)
      break
    case *sdl.MouseMotionEvent:
      app.onMouseMove(event)
      break
    case *sdl.MouseButtonEvent:
      if app.root.Menu.Visible() && !app.root.Menu.IsHit(int(event.X), int(event.Y)) && !app.root.Menu.IsOwnedBy(app.state.mouseElement) {
        if hasEvent(app.root.Menu.anchor, "mousebuttonoutsidemenu") {
          TriggerEvent(app.root.Menu.anchor, "mousebuttonoutsidemenu", 
            NewMouseEvent(int(event.X), int(event.Y)))
        } else {
          app.root.Menu.Hide()
        }
      }

      if event.Type == sdl.MOUSEBUTTONDOWN {
        app.onMouseDown(event)
      } else if event.Type == sdl.MOUSEBUTTONUP {
        app.onMouseUp(event)
      }

      break
    case *sdl.TextInputEvent:
      app.onTextInput(event)
      break
    case *sdl.KeyboardEvent:
      // tab and shift-tab cycle through the focusable elements
      if event.Keysym.Sym == sdl.K_TAB && event.State == sdl.PRESSED {
        app.onTab(event)
      } else {
        app.onKeyPress(event)
      }

      break
    case *sdl.WindowEvent:
      switch event.Event {
      case sdl.WINDOWEVENT_SHOWN:
        //fmt.Println("  window shown event")
        app.onShowOrResize()
        break
      case sdl.WINDOWEVENT_EXPOSED:
        //fmt.Println("  window exposed event")
        app.onShowOrResize()
        break
      case sdl.WINDOWEVENT_RESIZED:
        //fmt.Println("  window resized event")
        app.onShowOrResize()
        break
      case sdl.WINDOWEVENT_MAXIMIZED:
        //fmt.Println("  window maximized event")
        app.onShowOrResize()
        break
      case sdl.WINDOWEVENT_RESTORED:
        //fmt.Println("  window restored event")
        app.onShowOrResize()
        break
      case sdl.WINDOWEVENT_FOCUS_LOST:
        //fmt.Println("  window focus lost event")
        app.onBlur()
        break
      case sdl.WINDOWEVENT_FOCUS_GAINED:
        //fmt.Println("  window focus gained event")
        app.onFocus()
        break
      case sdl.WINDOWEVENT_LEAVE:
        //fmt.Println("  window leave event")
        app.onLeave()
        break
      case sdl.WINDOWEVENT_ENTER:
        //fmt.Println("  window enter event")
        app.onEnter()
        break
      }
    default:
      fmt.Println("unhandled event ", reflect.TypeOf(event_).String())
    }

    app.DrawIfDirty()
  }

  app.endDrawLoop()
}

func (app *App) onTick(event *animationEvent) {
  app.state.lastTick = event.tick

  app.root.Animate(event.tick)

  if event.force {
    app.root.ForcePosDirty()
  } 
}

func (app *App) onMouseMove(event *sdl.MouseMotionEvent) {
  if !app.state.outside {
    app.updateMouseElement(int(event.X), int(event.Y), int(event.XRel), int(event.YRel))
  }
}

func (app *App) onMouseDown(event *sdl.MouseButtonEvent) {
  if app.state.mouseElement == app.root.Menu || app.state.mouseElement == nil {
    // eg. on edge of menu
    return
  }

  app.state.lastDown = app.state.mouseElement

  if event.Button == sdl.BUTTON_LEFT {
    app.state.lastDownX = int(event.X)
    app.state.lastDownY = int(event.Y)
    app.state.mouseMoveSumX = 0
    app.state.mouseMoveSumY = 0

    app.triggerHitEvent("mousedown", NewMouseEvent(int(event.X), int(event.Y)))
  } else if event.Button == sdl.BUTTON_RIGHT {
    app.triggerHitEvent("rightclick", NewMouseEvent(int(event.X), int(event.Y)))
  }

  if !hasAncestor(app.state.mouseElement, app.root.Menu) {
    newFocusable := findFocusable(app.state.mouseElement)

    blurEvt := NewMouseEvent(int(event.X), int(event.Y))
    focusEvt := NewMouseEvent(int(event.X), int(event.Y))

    fmt.Println("changing focuselement upon mouseclick", dumpElement(newFocusable), dumpElement(app.state.mouseElement))
    app.changeFocusElement(newFocusable, blurEvt, focusEvt)
  }
}

func (app *App) onMouseUp(event *sdl.MouseButtonEvent) {
  if !app.state.outside && event.Button == sdl.BUTTON_LEFT {
    if elementNotNil(app.state.lastDown) && app.state.lastDown != app.state.mouseElement {
      tmp := app.state.mouseElement
      app.state.mouseElement = app.state.lastDown

      app.triggerHitEvent("mouseup", NewMouseEvent(int(event.X), int(event.Y)))

      app.state.mouseElement = tmp
    }

    app.triggerHitEvent("mouseup", NewMouseEvent(int(event.X), int(event.Y)))

    app.detectClick(int(event.X), int(event.Y))

    //app.updateCursor()
  } else {
    if event.Button == sdl.BUTTON_LEFT {
      app.state.mouseElement = app.state.lastDown

      app.triggerHitEvent("mouseup", NewMouseEvent(int(event.X), int(event.Y)))

      app.detectClick(int(event.X), int(event.Y))
    }

    app.state.mouseElement = nil
  }

  app.state.mouseMoveSumX = 0
  app.state.mouseMoveSumY = 0
  app.state.lastDown = nil
}

// turn a mouseup into a click, doubleclick or tripleclick
func (app *App) detectClick(x, y int) {
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
  app.state.mouseMoveSumX = 0
  app.state.mouseMoveSumY = 0

  eName := "click"
  switch app.state.upCount {
  case 0:
    panic("shouldn't be possible")
  case 1:
  case 2:
    if elementNotNil(app.state.mouseElement) {
      if hasEvent(app.state.mouseElement, "doubleclick") {
        eName = "doubleclick"
      }
    }
  default:
    if elementNotNil(app.state.mouseElement) {
      if hasEvent(app.state.mouseElement, "tripleclick") {
        eName = "tripleclick"
      }
    }
  }

  TriggerEvent(app.state.mouseElement, eName, NewMouseEvent(x, y))
}

func (app *App) onTextInput(event *sdl.TextInputEvent) {
  str := event.GetText()

  TriggerEvent(app.state.focusElement, "textinput", NewTextInputEvent(str))
}

func (app *App) onTab(event *sdl.KeyboardEvent) {
  app.hideMenuIfVisible()

  var newFocusable Element

  shift := event.Keysym.Mod & sdl.KMOD_SHIFT > 0
  if shift {
    if elementNotNil(app.state.focusElement) {
      newFocusable = findPrevFocusable(app.state.focusElement)
    } else {
      newFocusable = findPrevFocusable(app.root.Body)
    }
  } else {
    if elementNotNil(app.state.focusElement) {
      newFocusable = findNextFocusable(app.state.focusElement)

      fmt.Println("found next focusable", dumpElement(newFocusable))
    } else {
      newFocusable = findNextFocusable(app.root.Body)
    }
  }

  blurEvt := NewKeyboardEvent("tab", false, shift, false)
  focusEvt := NewKeyboardEvent("tab", false, shift, false)

  app.changeFocusElement(newFocusable, blurEvt, focusEvt)
}

func (app *App) onKeyPress(event *sdl.KeyboardEvent) {
  eType, kType, ctrl, shift, alt := extractKeyboardEventDetails(event)
  if eType != "" && kType != "" {
    TriggerEvent(app.state.focusElement, eType, NewKeyboardEvent(kType, ctrl, shift, alt))

    if eType == "keydown" {
      TriggerEvent(app.state.focusElement, "keypress", NewKeyboardEvent(kType, ctrl, shift, alt))
    }
  } else {
    fmt.Println("unhandled keyboardevent ", event.Keysym.Sym)
  }
}

func (app *App) onShowOrResize() {
  app.root.syncSize(app.window)

  app.root.ForcePosDirty()
}

func (app *App) onBlur() {
  TriggerEvent(app.state.focusElement, "blur", NewMouseEvent(-1, -1))
}

func (app *App) onFocus() {
  TriggerEvent(app.state.focusElement, "focus", NewMouseEvent(-1, -1))
}

func (app *App) onLeave() {
  if !app.state.outside {
    app.triggerHitEvent("mouseleave", NewMouseEvent(-1, -1))
  }

  app.state.mouseElement = nil
  app.state.outside = true
  app.state.cursor = -1
}

func (app *App) onEnter() {
  if app.mouseInWindow() {
    app.state.outside = false
    app.updateMouseElement(-1, -1, 0, 0)
  }
}
