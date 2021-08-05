package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

type Button struct {
  ElementData

  width  int
  height int
  flat   bool
  sticky bool

  // first 2 tris form the hear to the button
  tris []uint32
  dd   *DrawData

  down bool
  inside bool
  focused bool // explicit tab focus
  onClick func()
}

func NewButton(dd *DrawData) *Button {
  return newButton(dd, false, false)
}

func NewFlatButton(dd *DrawData) *Button {
  return newButton(dd, true, false)
}

func NewStickyFlatButton(dd *DrawData) *Button {
  return newButton(dd, true, true)
}

func newButton(dd *DrawData, flat bool, sticky bool) *Button {
  tris := dd.P1.Alloc(9*2)

  e := &Button{newElementData(), 200, 50, flat, sticky, tris, dd, false, false, false, nil}

  e.setTypesAndTCoords(false)

  e.SetEventListener("mousedown", e.onMouseDown)
  e.SetEventListener("mouseup", e.onMouseUp)
  e.SetEventListener("mouseleave", e.onMouseLeave)
  e.SetEventListener("mouseenter", e.onMouseEnter)
  e.SetEventListener("focus", e.onFocus)
  e.SetEventListener("blur", e.onBlur)
  e.SetEventListener("keydown", e.onKeyDown)
  e.SetEventListener("keyup", e.onKeyUp)

  return e
}

//go:generate ./A Button

func (e *Button) Cursor() int {
  return sdl.SYSTEM_CURSOR_HAND
}

func (e *Button) OnClick(fn func()) {
  e.onClick = fn
}

func (e *Button) SetSize(width, height int) {
  e.width = width
  e.height = height
  
  // TODO: trigger redraw
}

func (e *Button) setState(down bool, inside bool) {
  curPressed := e.down && e.inside

  e.down = down
  oldInside := e.inside
  e.inside = inside

  newPressed := e.down && e.inside

  if curPressed != newPressed || (e.flat && e.inside != oldInside) {
    e.setTypesAndTCoords(newPressed)
  }
}

func (e *Button) onMouseDown(evt *Event) {
  e.setState(true, e.inside)
}

func (e *Button) onMouseUp(evt *Event) {
  e.setState(false, e.inside)

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

func (e *Button) onFocus(evt *Event) {
  if evt.IsKeyboardEvent() {
    e.focused = true

    e.dd.FocusBox.Show(e.bb)
  }
}

func (e *Button) onBlur(evt *Event) {
  if e.focused {
    e.focused = false

    e.setState(false, false)

    e.setTypesAndTCoords(false)

    e.dd.FocusBox.Hide()
  }
}

func (e *Button) onKeyDown(evt *Event) {
  if e.focused && (evt.Key == "space" || evt.Key == "return") {
    curPressed := e.down
    e.down = true

    if !curPressed {
      e.setTypesAndTCoords(true)
    }
  }
}

func (e *Button) onKeyUp(evt *Event) {
  if e.focused && (evt.Key == "space" || evt.Key == "return") {
    curPressed := e.down
    e.down = false

    if curPressed {
      e.setTypesAndTCoords(false)
    }
  }
}

func (e *Button) setTypesAndTCoords(pressed bool) {
  t := e.dd.P1.Skin.ButtonBorderThickness()

  if pressed {
    x0, y0 := e.dd.P1.Skin.ButtonPressedOrigin()

    setBorderElementTypesAndTCoords(e.dd, e.tris, x0, y0, t, e.dd.P1.Skin.BGColor())
  } else if e.flat {
    if e.inside {
      x0, y0 := e.dd.P1.Skin.InsetOrigin()

      setBorderElementTypesAndTCoords(e.dd, e.tris, x0, y0, t, e.dd.P1.Skin.BGColor())
    } else {
      for _, tri := range e.tris {
        e.dd.P1.Type.Set1Const(tri, VTYPE_PLAIN)
        e.dd.P1.SetColorConst(tri, e.dd.P1.Skin.BGColor())
      }
    }
  } else {
    x0, y0 := e.dd.P1.Skin.ButtonOrigin()

    setBorderElementTypesAndTCoords(e.dd, e.tris, x0, y0, t, e.dd.P1.Skin.BGColor())
  }
}

// also used by input
func setBorderElementTypesAndTCoords(dd *DrawData, tris []uint32, x0, y0 int, t int, bgColor sdl.Color) {
  var (
    x [4]int
    y [4]int
  )

  x[0] = x0
  x[1] = x0 + t
  x[2] = x0 + t+1
  x[3] = x0 + 2*t+1

  y[0] = y0
  y[1] = y0 + t
  y[2] = y0 + t+1
  y[3] = y0 + 2*t+1

  for i := 0; i < 3; i++ {
    for j := 0; j < 3; j++ {
      tri0 := tris[(i*3 + j)*2 + 0]
      tri1 := tris[(i*3 + j)*2 + 1]

      if (i == 1 && j == 1) {
        dd.P1.Type.Set1Const(tri0, VTYPE_PLAIN)
        dd.P1.SetColorConst(tri0, bgColor)
        //dd.P1.TCoord.Set2Const(tri0, 0.0, 0.0)

        dd.P1.Type.Set1Const(tri1, VTYPE_PLAIN)
        dd.P1.SetColorConst(tri1, bgColor)
        //dd.P1.TCoord.Set2Const(tri1, 0.0, 0.0)
      } else {
        dd.P1.Type.Set1Const(tri0, VTYPE_SKIN)
        dd.P1.Color.Set4Const(tri0, 1.0, 1.0, 1.0, 1.0)
        dd.P1.SetSkinCoord(tri0, 0, x[i], y[j])
        dd.P1.SetSkinCoord(tri0, 1, x[i+1], y[j])
        dd.P1.SetSkinCoord(tri0, 2, x[i], y[j+1])

        dd.P1.Type.Set1Const(tri1, VTYPE_SKIN)
        dd.P1.Color.Set4Const(tri1, 1.0, 1.0, 1.0, 1.0)
        dd.P1.SetSkinCoord(tri1, 0, x[i+1], y[j+1])
        dd.P1.SetSkinCoord(tri1, 1, x[i+1], y[j])
        dd.P1.SetSkinCoord(tri1, 2, x[i], y[j+1])
      }
    }
  }
}

func (e *Button) OnResize(maxWidth, maxHeight int) (int, int) {
  t := e.dd.P1.Skin.ButtonBorderThickness()

  setBorderElementPosZ(e.dd, e.tris, e.width, e.height, t, e.z)

  e.ElementData.resizeChildren(e.width, e.height)

  return e.InitBB(e.width, e.height)
}

func setBorderElementPosZ(dd *DrawData, tris []uint32, width, height, t int, z float32) {
  var (
    x [4]int
    y [4]int
  )

  x[0] = 0
  x[1] = x[0] + t
  x[2] = x[0] + width - t
  x[3] = x[0] + width

  y[0] = 0
  y[1] = y[0] + t
  y[2] = y[0] + height - t
  y[3] = y[0] + height

  for i := 0; i < 3; i++ {
    for j := 0; j < 3; j++ {
      tri0 := tris[(i*3 + j)*2 + 0]
      tri1 := tris[(i*3 + j)*2 + 1]

      dd.P1.SetPos(tri0, 0, x[i], y[j], z)
      dd.P1.SetPos(tri0, 1, x[i+1], y[j], z)
      dd.P1.SetPos(tri0, 2, x[i], y[j+1], z)

      dd.P1.SetPos(tri1, 0, x[i+1], y[j+1], z)
      dd.P1.SetPos(tri1, 1, x[i+1], y[j], z)
      dd.P1.SetPos(tri1, 2, x[i], y[j+1], z)
    }
  }
}

func (e *Button) Translate(dx, dy int, dz float32) {
  for _, tri := range e.tris {
    e.dd.P1.TranslateTri(tri, dx, dy, dz)
  }

  e.ElementData.Translate(dx, dy, dz)
}
