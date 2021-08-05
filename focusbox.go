package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

type FocusBox struct {
  t int

  tris []uint32 // 8 tris

  dd *DrawData
}

func newFocusBox(dd *DrawData) *FocusBox {
  tris := dd.P1.Alloc(9*2) // the center quad is hidden, but allows for easy indexing

  e := &FocusBox{0, tris, dd}

  e.setTypesAndTCoords()

  return e
}

func (e *FocusBox) setTypesAndTCoords() {
  x0, y0 := e.dd.P1.Skin.FocusOrigin()
  e.t = e.dd.P1.Skin.FocusThickness()

  setBorderElementTypesAndTCoords(e.dd, e.tris, x0, y0, e.t, sdl.Color{0xff, 0xff, 0xff, 0xff})

  for _, tri := range e.tris {
    e.dd.P1.Type.Set1Const(tri, VTYPE_HIDDEN)
  }
  tri0 := e.tris[8]
  tri1 := e.tris[9]

  e.dd.P1.Type.Set1Const(tri0, VTYPE_HIDDEN)
  e.dd.P1.Type.Set1Const(tri1, VTYPE_HIDDEN)
}

func (e *FocusBox) Show(r Rect) {
  setBorderElementPosZ(e.dd, e.tris, r.W + e.t*2, r.H + e.t*2, e.t, 0.45)

  for i, tri := range e.tris {
    if i < 8 || i > 9 {
      e.dd.P1.Type.Set1Const(tri, VTYPE_SKIN)
    }

    e.dd.P1.TranslateTri(tri, r.X - e.t, r.Y - e.t, 0.0)
  }
}

func (e *FocusBox) Hide() {
  for _, tri := range e.tris {
    e.dd.P1.Type.Set1Const(tri, VTYPE_HIDDEN)
  }
}

func (e *FocusBox) Translate(dx, dy int, dz float32) {
  for _, tri := range e.tris {
    e.dd.P1.TranslateTri(tri, dx, dy, dz)
  }
}
