package glui

import (
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

func newMenu(frame *Frame) *Menu {
  e := &Menu{
    newElementData(frame, 9*2, 0),
    nil,
    0.0,
    0.0,
  }

  e.setTypesAndTCoords()

  e.Padding(e.Root.P1.Skin.ButtonBorderThickness())

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
      e.Root.P1.SetTriType(tri, VTYPE_PLAIN)
    } else {
      e.Root.P1.SetTriType(tri, VTYPE_SKIN)
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

  e.SetBorderedElementPos(w, h, t, maxZIndex)

  e.ElementData.CalcPosChildren(w, h, maxZIndex)

  e.InitRect(w, h)

  r := e.anchor.Rect()
  x_, y_ := r.Pos(e.anchorX, e.anchorY)

  x := x_ - int(e.anchorX*float64(w))
  y := y_ - int((1.0 - e.anchorY)*float64(h))

  // bound by window
  W, H := e.Root.GetSize()

  if x < 0 {
    x = 0
  } else if x + w > W {
    x = W - w
  }

  if y < 0 {
    y = 0
  } else if y + h > H {
    y = H - h
  }

  e.Translate(x, y)

  return 0, 0
}

func (e *Menu) ClearChildren() {
  e.height = 2*e.Root.P1.Skin.ButtonBorderThickness()

  e.ElementData.ClearChildren()
}

func (e *Menu) AddItem(item *MenuItem, enabled bool, selected bool) {
  // add mouseupeventlistener to close the menu
  item.On("mouseup", func(evt *Event) {
    if e.Visible() {
      e.Hide()
    } else {
      evt.stopPropagation = true
    }
  })

  item.Padding(0, 10)

  if !enabled {
    item.Disable()
  }

  if selected {
    item.Select()
  }

  e.A(item)

  e.height += item.height
}

func (e *Menu) IsOwnedBy(el Element) bool {
  return e.anchor == el && e.Visible()
}

func (e *Menu) loopMenuItems(fn func(i int, item *MenuItem)) int {
  c := 0

  for _, child_ := range e.children {
    if child, ok := child_.(*MenuItem); ok && child.enabled {
      fn(c, child)
      c += 1
    }
  }

  return c
}

func (e *Menu) countMenuItems() int {
  return e.loopMenuItems(func(i int, item *MenuItem) {})
}

func (e *Menu) SelectedIndex() int {
  found := -1

  e.loopMenuItems(func(i int, item *MenuItem) {
    if found == -1 && item.Selected() {
      found = i
    }
  })

  return found
}

func (e *Menu) unselectOtherMenuItems(item *MenuItem) {
  e.loopMenuItems(func(_ int, otherItem *MenuItem) {
    if otherItem != item {
      otherItem.Unselect()
    }
  })
}

func (e *Menu) SelectNext() {
  oldSel := e.SelectedIndex()

  e.loopMenuItems(func(i int, item *MenuItem) {
    if oldSel == -1 {
      item.Select()
      oldSel = -2
    } else if i == oldSel + 1 {
      item.Select()
    } else {
      item.Unselect()
    }
  })
}

func (e *Menu) SelectPrev() {
  oldSel := e.SelectedIndex()

  if oldSel == -1 {
    c := e.countMenuItems()

    e.loopMenuItems(func(i int, item *MenuItem) {
      if i == c -1 {
        item.Select()
      } else {
        item.Unselect()
      }
    })
  } else if oldSel == 0 {
    e.unselectOtherMenuItems(nil)
  } else {
    e.loopMenuItems(func(i int, item *MenuItem) {
      if i == oldSel - 1 {
        item.Select()
      } else {
        item.Unselect()
      }
    })
  }
}

func (e *Menu) ClickSelected() {
  if e.Visible() {
    // onClick must be called after menu is hidden (to avoid double click issues), but we must find the selected item before the menu is hidden (because the selected state is reset upon hiding the menu)
    var fn func() = nil

    e.loopMenuItems(func(i int, item *MenuItem) {
      if item.Selected() {
        fn = item.onClick
      }
    })

    e.Hide()

    if fn != nil {
      fn()
    }
  }
}
