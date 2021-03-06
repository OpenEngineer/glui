package glui

import (
  "math"
)

//go:generate ./gen_element Hor "A CalcDepth Padding Spacing H"

// special element that is just used for horizontal positioning of children
type Hor struct {
  ElementData

  hAlign  Align
  vAlign  Align
}

func NewHor(hAlign, vAlign Align, spacing int) *Hor {
  e := &Hor{
    NewElementData(0, 0), 
    hAlign, 
    vAlign,
  }

  if vAlign == STRETCH {
    panic("vAlign == STRETCH not supported in Hor")
  }

  e.spacing = spacing

  return e
}

// z is irrelevant here
func (e *Hor) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  // first space the children inline

  x := e.padding[3]
  maxChildH := 0

  childHs := make([]int, len(e.children))
  childWs := make([]int, len(e.children))

  for i, child := range e.children {
    if i > 0 {
      x += e.spacing
    }

    childW, childH := child.CalcPos(
      maxWidth - x - e.padding[1], 
      maxHeight - e.padding[0] - e.padding[2], 
      maxZIndex)

    childWs[i] = childW
    childHs[i] = childH

    child.Translate(x, 0)

    x += childW

    if childH > maxChildH {
      maxChildH = childH
    }
  }

  // dx[0] is 0 for STRETCH, dx[0..end] is a general translation for CENTER and END, dx[0..end] is 0 for START
  dx := make([]int, len(e.children)) 
  someDXSet := false
  if x < (maxWidth - e.padding[1]) {
    switch e.hAlign {
    case CENTER:
      dxAll := (maxWidth - x - e.padding[1])/2
      for i, _ := range e.children {
        dx[i] = dxAll
      }
      someDXSet = true
    case END:
      dxAll := maxWidth - x - e.padding[1]
      for i, _ := range e.children {
        dx[i] = dxAll
      }
      someDXSet = true
    case STRETCH:
      rem := maxWidth - x - e.padding[1]
      remPerChild := float64(rem)/float64(len(e.children) - 1)

      for i, _ := range e.children {
        dx[i] = int(math.Floor(float64(i)*remPerChild))
      }
      someDXSet = true
    }
  }

  h := e.height
  if h < 0 {
    h = maxHeight
  } else if h < maxChildH + e.padding[0] + e.padding[2] {
    h = maxChildH + e.padding[0] + e.padding[2]
  }

  if someDXSet || e.vAlign != START {
    for i, child := range e.children {
      dy := 0

      switch e.vAlign {
      case CENTER:
        dy = (h - childHs[i] - e.padding[0] - e.padding[2])/2
      case END:
        dy = (h - childHs[i] - e.padding[2])
      }

      child.Translate(dx[i], dy)
    }
  } 

  totalWidth := x + e.padding[1]
  if len(dx) > 0 {
    totalWidth += dx[len(dx)-1]
  }

  return e.InitRect(totalWidth, h)
}
