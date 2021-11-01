package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

type FocusRect struct {
  t      int

  anchor Element

  tris []uint32 // 8 tris

  Root *Frame
}

func newFocusRect(frame *Frame) *FocusRect {
  tris := frame.P1.Alloc(9*2) // the center quad is hidden, but allows for easy indexing

  e := &FocusRect{0, nil, tris, frame}

  e.setTypesAndTCoords()

  return e
}

func (e *FocusRect) setTypesAndTCoords() {
  x0, y0 := e.Root.P1.Skin.FocusOrigin()
  e.t = e.Root.P1.Skin.FocusThickness()

  e.Root.P1.setBorderedElementTypesAndTCoords(e.tris, x0, y0, e.t, sdl.Color{0xff, 0xff, 0xff, 0xff})

  for _, tri := range e.tris {
    e.Root.P1.Type.Set1Const(tri, VTYPE_HIDDEN)
  }
  tri0 := e.tris[8]
  tri1 := e.tris[9]

  e.Root.P1.Type.Set1Const(tri0, VTYPE_HIDDEN)
  e.Root.P1.Type.Set1Const(tri1, VTYPE_HIDDEN)
}

func (e *FocusRect) Show(anchor Element) {
  e.anchor = anchor

  for i, tri := range e.tris {
    if i < 8 || i > 9 {
      e.Root.P1.Type.Set1Const(tri, VTYPE_SKIN)
    }
  }
}

func (e *FocusRect) Hide() {
  e.anchor = nil

  for _, tri := range e.tris {
    e.Root.P1.Type.Set1Const(tri, VTYPE_HIDDEN)
  }
}

func (e *FocusRect) CalcPos(maxWidth, maxHeight, maxZIndex int) {
  if e.anchor == nil || e.anchor.Deleted() {
    e.Hide()

    return
  } 
  
  // z should be pretty high
  zId := e.anchor.ZIndex()
  zId = e.Root.Menu.ZIndex() - 1
  z := normalizeZIndex(zId, maxZIndex)

  r := e.anchor.Rect()

  e.Root.P1.setBorderedElementPos(e.tris, r.W + e.t*2, r.H + e.t*2, e.t, z)

  for _, tri := range e.tris {
    e.Root.P1.TranslateTri(tri, r.X - e.t, r.Y - e.t, 0.0)
  }
}

func (e *FocusRect) Translate(dx, dy int) {
  for _, tri := range e.tris {
    e.Root.P1.TranslateTri(tri, dx, dy, 0.0)
  }
}

func (e *FocusRect) Animate(tick uint64) {
  // doesnt do anything yet
}

func (e *FocusRect) IsOwnedBy(el Element) bool {
  return e.anchor == el
}
