package glui

import (
  "fmt"
  "reflect"
  "time"

  "github.com/veandco/go-sdl2/sdl"
)

func delay(d int) {
  time.Sleep(time.Duration(d)*time.Millisecond)
  //sdl.Delay(uint32(d))
}

func TriggerEvent(e Element, name string, evt *Event) {
  for elementNotNil(e) {
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
}

func (app *App) triggerHitEvent(name string, evt *Event) {
  frame := app.ActiveFrame()

  if elementNotNil(frame.state.mouseElement) {
    TriggerEvent(frame.state.mouseElement, name, evt)
  }
}

func (app *App) changeFocusElement(newFocusable Element, blurEvt, focusEvt *Event) {
  frame := app.ActiveFrame()

  if newFocusable != frame.state.focusElement {
    if elementNotNil(frame.state.focusElement) {
      TriggerEvent(frame.state.focusElement, "blur", blurEvt)
    }

    frame.state.focusElement = newFocusable

    if elementNotNil(newFocusable) {
      TriggerEvent(frame.state.focusElement, "focus", focusEvt)
    }
  } else if !frame.FocusRect.IsOwnedBy(frame.state.focusElement) && focusEvt.IsKeyboardEvent() {
    // retrigger the focus event
    TriggerEvent(frame.state.focusElement, "focus", focusEvt)
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

func (app *App) updateMouseElement(x, y int, dx, dy int) {
  frame := app.ActiveFrame()

  if x < 0 {
    x, y = currentMousePos()
  }

  newMouseElement, isSameOrChildOfOld := frame.findMouseElement(frame.state.mouseElement, x, y)


  // trigger mouse leave event if new mouseElement isn't child of old
  if elementNotNil(frame.state.mouseElement) && !isSameOrChildOfOld {
    evt := NewMouseEvent(x, y)

    ca := commonAncestor(frame.state.mouseElement, newMouseElement)

    evt.stopBubblingWhenElementReached(ca)

    app.triggerHitEvent("mouseleave", evt)
  }


  if !elementNotNil(frame.state.mouseElement) {
    evt := NewMouseEvent(x, y)
    frame.state.mouseElement = newMouseElement
    app.triggerHitEvent("mouseenter", evt)
  } else if frame.state.mouseElement != newMouseElement {
    evt := NewMouseEvent(x, y)

    ca := commonAncestor(frame.state.mouseElement, newMouseElement)

    evt.stopBubblingWhenElementReached(ca)

    frame.state.mouseElement = newMouseElement
    if frame.state.mouseElement == nil {
      fmt.Println("mouseElement is nil")
    } else {
      fmt.Println("mouseElement is ", reflect.TypeOf(frame.state.mouseElement).String())

      if oflow, ok := frame.state.mouseElement.(*Overflow); ok {
        fmt.Println(oflow.Rect())
        for _, child := range oflow.children {
          fmt.Println(reflect.TypeOf(child).String(), child.Rect())
        }
      }
    }

    if ca != newMouseElement {
      app.triggerHitEvent("mouseenter", evt)
    }
  }

  mouseMove := dx != 0 || dy != 0
  // mousemove event must be triggered before cursor is updated, as mousemove might change cursor
  if mouseMove {
    if elementNotNil(frame.state.mouseElement) {
      frame.state.mouseMoveSumX += dx
      frame.state.mouseMoveSumY += dy

      if frame.state.lastDownX + frame.state.mouseMoveSumX != x {
        dx += x - frame.state.lastDownX - frame.state.mouseMoveSumX
        frame.state.mouseMoveSumX = x
      }

      if frame.state.lastDownY + frame.state.mouseMoveSumY != y {
        dy += y - frame.state.lastDownY - frame.state.mouseMoveSumY
        frame.state.mouseMoveSumY = y
      }

      if elementNotNil(frame.state.lastDown) && frame.state.lastDown != frame.state.mouseElement {
        TriggerEvent(frame.state.lastDown, "mousemove", NewMouseMoveEvent(x, y, dx, dy))
      }

      app.triggerHitEvent("mousemove", NewMouseMoveEvent(x, y, dx, dy))
    }
  }

  app.updateCursor()
}

func (app *App) updateCursor() {
  frame := app.ActiveFrame()

  cursor := -1
  e := frame.state.mouseElement

  x, y := currentMousePos()

  for cursor < 0 && elementNotNil(e) {
    cursor = e.Cursor(x, y)
    e = e.Parent()
  }

  if cursor < 0 {
    cursor = sdl.SYSTEM_CURSOR_ARROW
  }

  if cursor != frame.state.cursor {
    frame.state.cursor = cursor

    if frame.state.cursor >= 0 && frame.state.cursor < sdl.NUM_SYSTEM_CURSORS {
      sdl.ShowCursor(sdl.ENABLE)

      oldCursor := sdl.GetCursor()

      c := sdl.CreateSystemCursor((sdl.SystemCursor)(frame.state.cursor))

      sdl.SetCursor(c)

      sdl.FreeCursor(oldCursor) // free the previous
    } else {
      panic("not custom cursors defined yet")
    }
  }
}

func (app *App) hideMenuIfVisible() {
  frame := app.ActiveFrame()

  if frame.Menu.Visible() {
    frame.Menu.Hide()
  }
}
