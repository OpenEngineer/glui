package glui

import (
  "fmt"
)

const HIDE_SHIFT = 2.0

// styled the same as a button, and can be filled with arbitrary children
type Menu struct {
  ElementData

  tris []uint32
  dd   *DrawData

  // state 
  visible bool

  onHide func()
}

func newMenu(dd *DrawData) *Menu {
  tris := dd.P1.Alloc(9*2)

  e := &Menu{newElementData(), tris, dd, false, nil}

  e.setTypesAndTCoords()

  return e
}

func (e *Menu) isVisible() bool {
  return e.visible
}

func (e *Menu) setTypesAndTCoords() {
  t := e.dd.P1.Skin.ButtonBorderThickness()

  x0, y0 := e.dd.P1.Skin.ButtonOrigin()

  setBorderElementTypesAndTCoords(e.dd, e.tris, x0, y0, t, e.dd.P1.Skin.BGColor())

  e.Hide()
}

func (e *Menu) OnHide(fn func()) {
  e.onHide = fn
}

func (e *Menu) A(children ...Element) Element {
  for _, child := range children {
    if !e.visible {
      child.Translate(0, 0, HIDE_SHIFT)
    } 

    e.ElementData.appendChild(child)
    child.RegisterParent(e)
  }

  return e
}

func (e *Menu) Hide() {
  if e.visible {
    // delete all children
    e.ElementData.Translate(0, 0, HIDE_SHIFT)

    for _, tri := range e.tris {
      e.dd.P1.Type.Set1Const(tri, VTYPE_HIDDEN)
    }

    e.visible = false

    if e.onHide != nil {
      e.onHide()
      e.onHide = nil
    }
  }
}

func (e *Menu) OnResize(maxWidth, maxHeight int) (int, int) {
  if e.visible {
    //e.Show(Rect{0, 0, maxWidth, maxHeight})

  }

  // doesnt push any siblings of body
  return 0, 0
}

func (e *Menu) Show(r Rect) {
  fmt.Println("attempting to show dialog")

  t := e.dd.P1.Skin.ButtonBorderThickness()

  z := float32(0.0)
  if !e.visible {
    z += HIDE_SHIFT
  }
  
  setBorderElementPosZ(e.dd, e.tris, r.W, r.H, t, z)
  for i, tri := range e.tris {
    if i == 8 || i == 9 {
      e.dd.P1.Type.Set1Const(tri, VTYPE_PLAIN)
    } else {
      e.dd.P1.Type.Set1Const(tri, VTYPE_SKIN)
    }
  }

  e.ElementData.resizeChildren(r.W, r.H)

  e.InitBB(r.W, r.H)

  dz := float32(0.0)
  if !e.visible {
    dz = -HIDE_SHIFT
  }


  dx, dy := r.X, r.Y
  fmt.Println("showing dialog: ", e.visible, dx, dy)
  for _, tri := range e.tris {
    e.dd.P1.TranslateTri(tri, dx, dy, dz)
  }

  e.ElementData.Translate(dx, dy, 0.0)

  e.visible = true
}

func (e *Menu) Clear() {
  // this method actually needs to do something
}

func (e *Menu) Translate(dx, dy int, dz float32) {
  fmt.Println("dialog z-index: ", e.dd.P1.Pos.Get(e.tris[0], 0, 2))
  for _, tri := range e.tris {
    e.dd.P1.TranslateTri(tri, dx, dy, dz)
  }

  e.ElementData.Translate(dx, dy, dz)
}
