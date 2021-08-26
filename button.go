package glui

//go:generate ./gen_element Button "A CalcDepth On Size Padding"

type Button struct {
  ElementData

  flat   bool
  sticky bool

  // state
  down    bool
  inside  bool

  onClick func()
}

func NewButton(root *Root) *Button {
  return newButton(root, false, false)
}

func NewFlatButton(root *Root) *Button {
  return newButton(root, true, false)
}

func NewStickyFlatButton(root *Root) *Button {
  return newButton(root, true, true)
}

func newButton(root *Root, flat bool, sticky bool) *Button {
  e := &Button{
    NewElementData(root, 9*2, 0), 
    flat, sticky, 
    false, false,
    nil,
  }

  e.Size(200, 50)

  e.setTypesAndTCoords(false)

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

func (e *Button) Cursor() int {
  return e.ButtonCursor(e.enabled)
}

func (e *Button) OnClick(fn func()) {
  e.onClick = fn
}

func (e *Button) setState(down bool, inside bool) {
  curPressed := e.down && e.inside

  e.down = down
  oldInside := e.inside
  e.inside = inside

  newPressed := e.down && e.inside

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
}

func (e *Button) onMouseUp(evt *Event) {
  e.setState(false, e.inside)
}

func (e *Button) onMouseClick(evt *Event) {
  if e.onClick != nil {
    e.onClick()
  }
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
  if evt.Key == "space" || evt.Key == "return" {
    curPressed := e.down
    e.down = true

    if !curPressed {
      e.setTypesAndTCoords(true)
    }
  }
}

func (e *Button) onKeyUp(evt *Event) {
  if evt.Key == "space" || evt.Key == "return" {
    curPressed := e.down
    e.down = false

    if curPressed {
      e.setTypesAndTCoords(false)
    }
  }
}

func (e *Button) Show() {
  e.setTypesAndTCoords(e.down && e.inside)

  e.ElementData.Show()
}

func (e *Button) setTypesAndTCoords(pressed bool) {
  t := e.Root.P1.Skin.ButtonBorderThickness()

  if pressed {
    x0, y0 := e.Root.P1.Skin.ButtonPressedOrigin()

    setBorderedElementTypesAndTCoords(e.Root, e.p1Tris, x0, y0, t, e.Root.P1.Skin.BGColor())
  } else if e.flat {
    if e.inside {
      x0, y0 := e.Root.P1.Skin.InsetOrigin()

      setBorderedElementTypesAndTCoords(e.Root, e.p1Tris, x0, y0, t, e.Root.P1.Skin.BGColor())
    } else {
      for _, tri := range e.p1Tris {
        e.Root.P1.Type.Set1Const(tri, VTYPE_PLAIN)
        e.Root.P1.SetColorConst(tri, e.Root.P1.Skin.BGColor())
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
  e.setState(false, false)

  e.ElementData.Hide()
}
