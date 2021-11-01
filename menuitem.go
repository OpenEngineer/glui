package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

//go:generate ./gen_element MenuItem "appendChild CalcDepth On Size Padding H W"

type MenuItemConfig struct {
  Caption  string
  Callback func()
  Width    int
}

// a plain button
type MenuItem struct {
  ElementData

  menu    *Menu
  caption *Caption

  // state
  selected bool

  onClick func()
}

func NewMenuItem(captionText string, callback func()) *MenuItem {
  menu := ActiveFrame().Menu

  caption := NewSansCaption(captionText, 10)

  e := &MenuItem{
    NewElementData(2, 0),
    menu,
    caption,
    false,
    callback,
  }

  e.height = 30
  e.width = 200

  e.setTypesAndColor()

  e.On("mouseup",    e.onMouseClick)
  e.On("mouseleave", e.onMouseLeave)
  e.On("mouseenter", e.onMouseEnter)
  
  e.appendChild(NewHor(START, CENTER, 0).H(-1).A(caption))

  return e
}

func newMenuItemFromConfig(cfg MenuItemConfig) *MenuItem {
  item := NewMenuItem(cfg.Caption, cfg.Callback)

  if cfg.Width != 0 {
    item.W(cfg.Width)
  }

  return item
}

func (e *MenuItem) setTypesAndColor() {
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

func (e *MenuItem) setState(selected bool) {
  if e.enabled && e.selected != selected {
    e.selected = selected

    e.setTypesAndColor()
  }
}

func (e *MenuItem) onMouseEnter(evt *Event) {
  e.menu.unselectOtherMenuItems(e)

  e.Select()
}

func (e *MenuItem) onMouseLeave(evt *Event) {
  e.Unselect()
}

func (e *MenuItem) Select() {
  e.setState(true)
}

func (e *MenuItem) Unselect() {
  e.setState(false)
}

func (e *MenuItem) Selected() bool {
  return e.selected
}

func (e *MenuItem) onMouseClick(evt *Event) {
  if e.onClick != nil {
    e.onClick()
  }
}

func (e *MenuItem) Cursor() int {
  if e.enabled {
    return sdl.SYSTEM_CURSOR_HAND
  } else {
    return -1
  }
}

func (e *MenuItem) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  w := maxWidth

  e.Root.P1.SetQuadPos(e.p1Tris[0], e.p1Tris[1], Rect{0, 0, w, e.height}, e.Z(maxZIndex))

  e.CalcPosChildren(w, e.height, maxZIndex)

  return e.InitRect(w, e.height)
}

func (e *MenuItem) Hide() {
  e.setState(false)

  e.ElementData.Hide()
}
