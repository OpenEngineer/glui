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

  e.Padding(root.P1.Skin.ButtonBorderThickness())

  e.Hide()

  return e
}

func (e *Menu) setTypesAndTCoords() {
  e.SetButtonStyle()

  e.Hide()
}

// anchorX, anchorY are between 0.0 and 1.0
// anchorX, anchorY specify a point in the anchor elements rect
// the positioning of the menu will try to place the target point of the menu as close as 
//  possible to the anchor point
func (e *Menu) ShowAt(anchor Element, anchorX, anchorY float64, width int) {
  e.anchor  = anchor
  e.anchorX = anchorX
  e.anchorY = anchorY
  e.width   = width

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

func (e *Menu) ClearChildren() {
  e.height = 2*e.Root.P1.Skin.ButtonBorderThickness()

  e.ElementData.ClearChildren()
}

func (e *Menu) AddButton(caption string, enabled bool, selected bool, height int, callback func()) {
  button := newMenuButton(e, caption, func() {
    callback()
    e.Hide()
  }).Padding(0, 10)

  button.height = height

  if !enabled {
    button.Disable()
  }

  if selected {
    button.Select()
  }

  e.A(button)

  e.height += height
}

func (e *Menu) IsOwnedBy(el Element) bool {
  return e.anchor == el && e.Visible()
}

func (e *Menu) loopMenuButtons(fn func(i int, mb *menuButton)) int {
  c := 0

  for _, child_ := range e.children {
    if child, ok := child_.(*menuButton); ok && child.enabled {
      fn(c, child)
      c += 1
    }
  }

  return c
}

func (e *Menu) countMenuButtons() int {
  return e.loopMenuButtons(func(i int, mb *menuButton) {})
}

func (e *Menu) SelectedIndex() int {
  found := -1

  e.loopMenuButtons(func(i int, mb *menuButton) {
    if found == -1 && mb.Selected() {
      found = i
    }
  })

  return found
}

func (e *Menu) unselectOtherMenuButtons(b *menuButton) {
  e.loopMenuButtons(func(_ int, mb *menuButton) {
    if mb != b {
      mb.Unselect()
    }
  })
}

func (e *Menu) SelectNext() {
  oldSel := e.SelectedIndex()

  e.loopMenuButtons(func(i int, mb *menuButton) {
    if oldSel == -1 {
      mb.Select()
      oldSel = -2
    } else if i == oldSel + 1 {
      mb.Select()
    } else {
      mb.Unselect()
    }
  })
}

func (e *Menu) SelectPrev() {
  oldSel := e.SelectedIndex()

  if oldSel == -1 {
    c := e.countMenuButtons()

    e.loopMenuButtons(func(i int, mb *menuButton) {
      if i == c -1 {
        mb.Select()
      } else {
        mb.Unselect()
      }
    })
  } else if oldSel == 0 {
    e.unselectOtherMenuButtons(nil)
  } else {
    e.loopMenuButtons(func(i int, mb *menuButton) {
      if i == oldSel - 1 {
        mb.Select()
      } else {
        mb.Unselect()
      }
    })
  }
}

func (e *Menu) ClickSelected() {
  e.loopMenuButtons(func(i int, mb *menuButton) {
    if mb.Selected() {
      mb.onClick()
    }
  })
  
  e.Hide()
}
