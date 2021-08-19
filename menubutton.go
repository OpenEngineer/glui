package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

//go:generate ./gen_element menuButton "appendChild CalcDepth On Size Padding"

// a plain button
type menuButton struct {
  ElementData

  caption *Caption

  // state
  focused bool

  onClick func()
}

func newMenuButton(root *Root, captionText string, callback func()) *menuButton {
  caption := NewSansCaption(root, captionText, 10)

  e := &menuButton{
    NewElementData(root, 2, 0),
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

  if e.focused && e.enabled {
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

func (e *menuButton) setState(focused bool) {
  if e.enabled && e.focused != focused {
    e.focused = focused

    e.setTypesAndColor()
  }
}

func (e *menuButton) onMouseEnter(evt *Event) {
  e.setState(true)
}

func (e *menuButton) onMouseLeave(evt *Event) {
  e.setState(false)
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
