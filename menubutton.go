package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

//go:generate ./gen_element menuButton "appendChild CalcDepth On Size Padding"

// a plain button
type menuButton struct {
  ElementData

  menu    *Menu
  caption *Caption

  // state
  selected bool

  onClick func()
}

func newMenuButton(menu *Menu, captionText string, callback func()) *menuButton {
  root := menu.Root

  caption := NewSansCaption(root, captionText, 10)

  e := &menuButton{
    NewElementData(root, 2, 0),
    menu,
    caption,
    false,
    callback,
  }

  e.height = 30

  e.setTypesAndColor()

  e.On("mousedown",  e.onMouseClick)
  e.On("mouseup",    e.onMouseClick)
  e.On("mouseleave", e.onMouseLeave)
  e.On("mouseenter", e.onMouseEnter)
  
  e.appendChild(NewHor(root, START, CENTER, 0).A(caption))

  return e
}

func (e *menuButton) setTypesAndColor() {
  var c sdl.Color

  if e.selected && e.enabled {
    c = e.Root.P1.Skin.SelColor()
    e.caption.SetColor(sdl.Color{0xff, 0xff, 0xff, 0xff})
  } else {
    c = e.Root.P1.Skin.BGColor()
    e.caption.SetColor(sdl.Color{0x00, 0x00, 0x00, 0xff})
  }

  for _, tri := range e.p1Tris {
    e.Root.P1.Type.Set1Const(tri, VTYPE_PLAIN)
    e.Root.P1.SetColorConst(tri, c)
  }
}

func (e *menuButton) setState(selected bool) {
  if e.enabled && e.selected != selected {
    e.selected = selected

    e.setTypesAndColor()
  }
}

func (e *menuButton) onMouseEnter(evt *Event) {
  e.menu.unselectOtherMenuButtons(e)

  e.Select()
}

func (e *menuButton) onMouseLeave(evt *Event) {
  e.Unselect()
}

func (e *menuButton) Select() {
  e.setState(true)
}

func (e *menuButton) Unselect() {
  e.setState(false)
}

func (e *menuButton) Selected() bool {
  return e.selected
}

func (e *menuButton) onMouseClick(evt *Event) {
  if e.onClick != nil {
    e.onClick()
  }
}

func (e *menuButton) Cursor() int {
  if e.enabled {
    return sdl.SYSTEM_CURSOR_HAND
  } else {
    return -1
  }
}

func (e *menuButton) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  w := maxWidth

  e.Root.P1.SetQuadPos(e.p1Tris[0], e.p1Tris[1], Rect{0, 0, w, e.height}, e.Z(maxZIndex))

  e.CalcPosChildren(w, e.height, maxZIndex)

  return e.InitRect(w, e.height)
}

func (e *menuButton) Hide() {
  e.setState(false)

  e.ElementData.Hide()
}
