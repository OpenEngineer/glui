package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

type Button struct {
  ElementData

  // first 2 tris form the hear to the button
  tris []uint32
  dd   *DrawData

  down bool
  inside bool
}

func NewButton(dd *DrawData) *Button {
  tris := dd.P1.Alloc(9*2)

  e := &Button{newElementData(), tris, dd, false, false}

  e.setTypesAndTCoords(false)

  e.SetEventListener("mousedown", e.onMouseDown)
  e.SetEventListener("mouseup", e.onMouseUp)
  e.SetEventListener("mouseleave", e.onMouseLeave)
  e.SetEventListener("mouseenter", e.onMouseEnter)

  return e
}

//go:generate ./A Button

func (e *Button) Cursor() int {
  return sdl.SYSTEM_CURSOR_HAND
}

func (e *Button) setState(down bool, inside bool) {
  curPressed := e.down && e.inside

  e.down = down
  e.inside = inside

  newPressed := e.down && e.inside

  if curPressed != newPressed {
    e.setTypesAndTCoords(newPressed)
  }
}

func (e *Button) onMouseDown(evt *Event) {
  e.setState(true, e.inside)
}

func (e *Button) onMouseUp(evt *Event) {
  e.setState(false, e.inside)
}

func (e *Button) onMouseLeave(evt *Event) {
  e.setState(e.down, false)
}

func (e *Button) onMouseEnter(evt *Event) {
  e.setState(e.down, true)
}

func (e *Button) setTypesAndTCoords(pressed bool) {
  t := e.dd.P1.Skin.ButtonBorderThickness()

  if pressed {
    x0, y0 := e.dd.P1.Skin.ButtonPressedOrigin()

    setBorderElementTypesAndTCoords(e.dd, e.tris, x0, y0, t, e.dd.P1.Skin.BGColor())
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
        dd.P1.TCoord.Set2Const(tri0, 0.0, 0.0)

        dd.P1.Type.Set1Const(tri1, VTYPE_PLAIN)
        dd.P1.SetColorConst(tri1, bgColor)
        dd.P1.TCoord.Set2Const(tri1, 0.0, 0.0)
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
  width  := 200
  height := 50

  t := e.dd.P1.Skin.ButtonBorderThickness()

  setBorderElementPos(e.dd, e.tris, width, height, t)

  e.ElementData.resizeChildren(width, height)

  return e.InitBB(width, height)
}

func setBorderElementPos(dd *DrawData, tris []uint32, width, height, t int) {
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

      dd.P1.SetPos(tri0, 0, x[i], y[j], 0.5)
      dd.P1.SetPos(tri0, 1, x[i+1], y[j], 0.5)
      dd.P1.SetPos(tri0, 2, x[i], y[j+1], 0.5)

      dd.P1.SetPos(tri1, 0, x[i+1], y[j+1], 0.5)
      dd.P1.SetPos(tri1, 1, x[i+1], y[j], 0.5)
      dd.P1.SetPos(tri1, 2, x[i], y[j+1], 0.5)
    }
  }
}

func (e *Button) Translate(dx, dy int) {
  for _, tri := range e.tris {
    e.dd.P1.TranslateTri(tri, dx, dy)
  }

  e.ElementData.Translate(dx, dy)
}
