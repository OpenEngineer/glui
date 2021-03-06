package glui

import (
)

//go:generate ./gen_element Button "A CalcDepth On Size Padding H W"

type Button struct {
  ElementData

  flat   bool
  sticky bool

  // state
  down    bool
  inside  bool

  onClick func()
}

func NewButton() *Button {
  return newButton(false, false)
}

func NewFlatButton() *Button {
  return newButton(true, false)
}

func NewFlatIconButton(iconName string, iconSize int) *Button {
  icon := NewIcon(iconName, iconSize)

  button := NewFlatButton()
  button.A(NewHor(CENTER, CENTER, 0).H(-1).A(icon))

  return button
}

func NewIconButton(iconName string, iconSize int, iconOrientation Orientation) *Button {
  icon := NewIcon(iconName, iconSize)

  if iconOrientation != HOR {
    icon.SetOrientation(iconOrientation)
  }

  button := NewButton()
  button.A(NewHor(CENTER, CENTER, 0).H(-1).A(icon))

  return button
}

func NewStickyFlatButton() *Button {
  return newButton(true, true)
}

func NewCaptionButton(caption string, optArgs ...interface{}) *Button {
  b := NewButton()

  var (
    hAlign Align = CENTER
    vAlign Align = CENTER
    lrPadding int = 0
    ok bool
  )

  if len(optArgs) >= 1 {
    hAlign, ok = optArgs[0].(Align)
    if !ok {
      panic("expected Align for first arg")
    }
  }

  if len(optArgs) >= 2 {
    vAlign, ok = optArgs[1].(Align)
    if !ok {
      panic("expected Align for second arg")
    }
  }

  if len(optArgs) >= 3 {
    lrPadding, ok = optArgs[2].(int)
    if !ok {
      panic("expected int for third arg")
    }
  }

  if len(optArgs) >= 4 {
    panic("unexpected number of args")
  }

  hor := NewHor(hAlign, vAlign, 0).H(-1)

  hor.A(NewSans(caption, 10))
  hor.Padding(0, lrPadding, 0, lrPadding)

  b.A(hor)

  return b
}

func newButton(flat bool, sticky bool) *Button {
  e := &Button{
    NewElementData(9*2, 0), 
    flat, sticky, 
    false, false,
    nil,
  }

  e.Size(200, 50)

  e.Show()

  e.On("mousedown", e.onMouseDown)
  e.On("mouseup", e.onMouseUp)
  e.On("click", e.onMouseClick)
  e.On("mouseleave", e.onMouseLeave)
  e.On("mouseenter", e.onMouseEnter)
  e.On("focus", e.onFocus)
  e.On("blur", e.onBlur)
  e.On("keydown", e.onKeyDown)
  e.On("keyup", e.onKeyUp)

  return e
}

func (e *Button) Cursor(x, y int) int {
  return e.ButtonCursor(x, y, e.enabled)
}

// set onClick here, so it can also be accessed by the keypresses
func (e *Button) OnClick(fn func()) *Button {
  e.onClick = fn

  return e
}

func (e *Button) setState(down bool, inside bool) {
  curPressed := e.down && (e.inside || e.sticky)

  e.down = down
  oldInside := e.inside
  e.inside = inside

  newPressed := e.down && (e.inside || e.sticky)

  if e.enabled {
    if curPressed != newPressed || (e.flat && e.inside != oldInside) {
      e.setTypesAndTCoords(newPressed)
    }
  }
}

func (e *Button) Disable() {
  e.setState(false, e.inside)

  e.ElementData.Disable()
}

func (e *Button) Enable() {
  e.ElementData.Enable()
}

func (e *Button) onMouseDown(evt *Event) {
  e.setState(true, e.inside)

  evt.StopBubbling()
}

func (e *Button) onMouseUp(evt *Event) {
  if !e.sticky {
    e.setState(false, e.inside)
  }
}

func (e *Button) triggerClick(evt *Event) {
  if e.onClick != nil {
    e.onClick()

    evt.StopBubbling()
  }
}

func (e *Button) onMouseClick(evt *Event) {
  e.triggerClick(evt)
}

func (e *Button) onMouseLeave(evt *Event) {
  e.setState(e.down, false)
}

func (e *Button) onMouseEnter(evt *Event) {
  e.setState(e.down, true)
}

func (e *Button) focused() bool {
  return e.Root.FocusRect.IsOwnedBy(e)
}

func (e *Button) onFocus(evt *Event) {
  if evt.IsKeyboardEvent() {
    e.Root.FocusRect.Show(e)
  }
}

func (e *Button) onBlur(evt *Event) {
  if e.focused() {
    e.setState(false, false)

    e.setTypesAndTCoords(false)

    e.Root.FocusRect.Hide()
  }
}

func (e *Button) onKeyDown(evt *Event) {
  if evt.IsReturnOrSpace() {
    curPressed := e.down
    e.down = true

    if !curPressed {
      e.setTypesAndTCoords(true)
    }
  }
}

func (e *Button) onKeyUp(evt *Event) {
  if evt.IsReturnOrSpace() {
    curPressed := e.down

    if !e.sticky {
      e.down = false
    }

    if curPressed && !e.sticky {
      e.setTypesAndTCoords(false)
    }

    if curPressed {
      e.triggerClick(evt)
    }
  }
}

func (e *Button) Show() {
  e.setTypesAndTCoords(e.down && e.inside)

  e.ElementData.Show()
}

func (e *Button) setTypesAndTCoords(pressed bool) {
  t := e.Root.P1.Skin.ButtonBorderThickness()

  if len(e.p1Tris) == 0 {
    panic("no p1Tris available")
  }

  if pressed {
    x0, y0 := e.Root.P1.Skin.ButtonPressedOrigin()

    e.Root.P1.setBorderedElementTypesAndTCoords(e.p1Tris, x0, y0, t, e.Root.P1.Skin.BGColor())
  } else if e.flat {
    if e.inside {
      x0, y0 := e.Root.P1.Skin.InsetOrigin()

      e.Root.P1.setBorderedElementTypesAndTCoords(e.p1Tris, x0, y0, t, e.Root.P1.Skin.BGColor())

      // hide the center tris
      tri0 := e.p1Tris[(3+1)*2 + 0]
      tri1 := e.p1Tris[(3+1)*2 + 1]

      e.Root.P1.SetTriType(tri0, VTYPE_HIDDEN)
      e.Root.P1.SetTriType(tri1, VTYPE_HIDDEN)
    } else {
      for _, tri := range e.p1Tris {
        // hide the tris, leave the color up the underlying element
        e.Root.P1.SetTriType(tri, VTYPE_HIDDEN)
        //e.Root.P1.SetColorConst(tri, e.Root.P1.Skin.BGColor())
      }
    }
  } else {
    e.SetButtonStyle()
  }
}

func (e *Button) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  return e.SetButtonPos(maxWidth, maxHeight, maxZIndex)
}

func (e *Button) Hide() {
  //e.setState(false, false)

  e.ElementData.Hide()
}

func (e *Button) Unstick() {
  if e.sticky {
    e.setState(false, e.inside)
  }
}
