package glui

import (
  "fmt"
)

//go:generate ./gen_element Menu "A CalcDepth Padding Spacing"

// styled the same as a button, and can be filled with arbitrary children
type Menu struct {
  ElementData

  // state 
  anchor  Element // XXX: what if anchor is deleted?
  anchorX float64
  anchorY float64
}

func newMenu(root *Root) *Menu {
  e := &Menu{
    NewElementData(root, 9*2, 0),
    nil,
    0.0,
    0.0,
  }

  e.setTypesAndTCoords()

  e.Hide()

  return e
}

func (e *Menu) setTypesAndTCoords() {
  t := e.Root.P1.Skin.ButtonBorderThickness()

  x0, y0 := e.Root.P1.Skin.ButtonOrigin()

  setBorderedElementTypesAndTCoords(e.Root, e.p1Tris, x0, y0, t, e.Root.P1.Skin.BGColor())

  e.Hide()
}

// anchorX, anchorY are between 0.0 and 1.0
// anchorX, anchorY specify a point in the anchor elements rect
// the positioning of the menu will try to place the target point of the menu as close as 
//  possible to the anchor point
func (e *Menu) ShowAt(anchor Element, anchorX, anchorY float64, width, height int) {
  e.anchor  = anchor
  e.anchorX = anchorX
  e.anchorY = anchorY
  e.width   = width
  e.height  = height

  e.ElementData.Show()

  for i, tri := range e.p1Tris {
    if i == 8 || i == 9 {
      e.Root.P1.Type.Set1Const(tri, VTYPE_PLAIN)
    } else {
      e.Root.P1.Type.Set1Const(tri, VTYPE_SKIN)
    }
  }

  e.Root.ForcePosDirty()
}

func (e *Menu) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  if !e.Visible() {
    return 0, 0
  }

  t := e.Root.P1.Skin.ButtonBorderThickness()

  w, h := e.GetSize()

  fmt.Println("menu z index: ", e.ZIndex())
  e.SetBorderedElementPos(w, h, t, maxZIndex)

  e.ElementData.CalcPosChildren(w, h, maxZIndex)

  e.InitRect(w, h)

  r := e.anchor.Rect()
  x, y := r.Pos(e.anchorX, e.anchorY)

  // bound by window
  W, H := e.Root.GetSize()

  if x < 0 {
    x = 0
  } else if x + r.W > W {
    x = W - r.W
  }

  if y < 0 {
    y = 0
  } else if y + r.H > H {
    y = H - r.H
  }

  e.Translate(x, y)

  return 0, 0
}
