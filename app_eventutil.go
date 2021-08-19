package glui

import (
  "fmt"
  "reflect"

  "github.com/veandco/go-sdl2/sdl"
)

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
}

func (app *App) triggerHitEvent(name string, evt *Event) {
  app.triggerEvent(app.state.mouseElement, name, evt)
}

// XXX: one focus element per root?
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

// the window enter or leave events might be called spuriously
func (app *App) mouseInWindow() bool {
  x0, y0 := app.window.GetPosition()
  w, h := app.window.GetSize()

  x, y, _ := sdl.GetGlobalMouseState()

  r := Rect{int(x0), int(y0), int(w), int(h)}

  b := r.Hit(int(x), int(y))

  return b
}

func (app *App) updateMouseElement(x, y int) {
  if x < 0 {
    x_, y_, _ := sdl.GetMouseState()

    x = int(x_)
    y = int(y_)
  }

  newMouseElement, isSameOrChildOfOld := app.root.findMouseElement(app.state.mouseElement, x, y)

  // trigger mouse leave event if new mouseElement isn't child of old
  if app.state.mouseElement != nil && !isSameOrChildOfOld {
    evt := NewMouseEvent(x, y)

    ca := commonAncestor(app.state.mouseElement, newMouseElement)

    evt.stopBubblingWhenElementReached(ca)

    app.triggerHitEvent("mouseleave", evt)
  }

  if app.state.mouseElement == nil {
    evt := NewMouseEvent(x, y)
    app.state.mouseElement = newMouseElement
    app.triggerHitEvent("mouseenter", evt)
  } else if app.state.mouseElement != newMouseElement {
    evt := NewMouseEvent(x, y)

    ca := commonAncestor(app.state.mouseElement, newMouseElement)

    evt.stopBubblingWhenElementReached(ca)

    app.state.mouseElement = newMouseElement
    if app.state.mouseElement == nil {
      fmt.Println("mouseElement is nil")
    } else {
      fmt.Println("mouseElement is ", reflect.TypeOf(app.state.mouseElement).String())
    }

    if ca != newMouseElement {
      app.triggerHitEvent("mouseenter", evt)
    }
  }

  cursor := -1
  e := app.state.mouseElement
  for cursor < 0 && e != nil {
    cursor = e.Cursor()
    e = e.Parent()
  }

  if cursor < 0 {
    cursor = sdl.SYSTEM_CURSOR_ARROW
  }

  if cursor != app.state.cursor {
    app.state.cursor = cursor

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

func (app *App) hideMenuIfVisible() {
  if app.root.Menu.Visible() {
    app.root.Menu.Hide()
  }
}
