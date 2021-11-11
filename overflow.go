package glui

import (
)

//go:generate ./gen_element Overflow "CalcDepth appendChild A Size H W On"

// element with horizontal and vertical scrolling
type Overflow struct {
  ElementData
}

func NewOverflow() *Overflow {
  e := &Overflow{
    NewElementData(0, 0),
  }

  horSB := NewScrollbar(HOR)
  verSB := NewScrollbar(VER)

  e.appendChild(horSB)
  e.appendChild(verSB)

  e.On("wheel", e.onWheel)

  return e
}

func (e *Overflow) horScrollbar() *Scrollbar {
  sb, ok := e.children[0].(*Scrollbar)
  if !ok {
    panic("expected scrollbar")
  }

  return sb
}

func (e *Overflow) verScrollbar() *Scrollbar {
  sb, ok := e.children[1].(*Scrollbar)
  if !ok {
    panic("expected scrollbar")
  }

  return sb
}

func (e *Overflow) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  if e.width != -1 {
    maxWidth = e.width
  } 

  if e.height != -1 {
    maxHeight = e.height
  }

  horSB  := e.horScrollbar()
  verSB  := e.verScrollbar()

  childrenBck := e.children[:]
  e.children = e.children[2:]
  innerW, innerH := e.CalcPosChildren(maxWidth, maxHeight, maxZIndex)

  sbTrackSize := e.Root.P1.Skin.ScrollbarTrackSize()
  crop := false

  dx := float32(0.0)
  dy := float32(0.0)

  if innerW > maxWidth - sbTrackSize {
    horSB.Show()
    horSB.SetSliderLength(int(float64(maxWidth - 3*sbTrackSize)*float64(maxWidth - sbTrackSize)/float64(innerW)))
    horSB.CalcPos(maxWidth - sbTrackSize, sbTrackSize, maxZIndex)
    horSB.Translate(0.0, maxHeight - sbTrackSize)
    pos := horSB.Pos()
    crop = true
    dx = -pos*float32(innerW)
  } else {
    horSB.Hide()
    maxWidth = innerW + sbTrackSize
  }

  if innerH > maxHeight - sbTrackSize {
    verSB.Show()
    verSB.SetSliderLength(int(float64(maxHeight - 3*sbTrackSize)*float64(maxHeight - sbTrackSize)/float64(innerH)))
    verSB.CalcPos(sbTrackSize, maxHeight - sbTrackSize, maxZIndex)
    verSB.Translate(maxWidth - sbTrackSize, 0.0)
    crop = true
    pos := verSB.Pos()
    dy = -pos*float32(innerH)
  } else {
    verSB.Hide()
    maxHeight = innerH + sbTrackSize
  }

  if crop {
    e.Translate(int(dx), int(dy))
    e.Crop(Rect{0, 0, maxWidth - sbTrackSize, maxHeight - sbTrackSize})
  }

  e.children = childrenBck

  return e.InitRect(maxWidth, maxHeight)
}

func (e *Overflow) onWheel(evt *Event) {
  horSB := e.horScrollbar()
  verSB := e.verScrollbar()

  if horSB.Visible() {
    horSB.MoveBy(evt.XRel)
  }

  if verSB.Visible() {
    verSB.MoveBy(evt.YRel)
  }
}
