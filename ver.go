package glui

import (
  "math"
)

//go:generate ./gen_element Ver "A CalcDepth Padding Spacing W"

// special element that is just used for vertical positioning of children
type Ver struct {
  ElementData

  vAlign Align
  hAlign Align
}

func NewVer(vAlign, hAlign Align, spacing int) *Ver {
  e := &Ver{
    NewElementData(0, 0),
    vAlign,
    hAlign,
  }

  if hAlign == STRETCH {
    panic("hAlign == STRETCH not supported in Ver")
  }

  e.spacing = spacing

  return e
}

func (e *Ver) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  y := e.padding[0]
  maxChildW := 0

  childWs := make([]int, len(e.children))
  childHs := make([]int, len(e.children))

  for i, child := range e.children {
    if i > 0 {
      y += e.spacing
    }

    childW, childH := child.CalcPos(
      maxWidth - e.padding[1] - e.padding[3],
      maxHeight - y - e.padding[2],
      maxZIndex)

    childHs[i] = childH
    childWs[i] = childW

    child.Translate(0, y)

    y += childH

    if childW > maxChildW {
      maxChildW = childW
    }
  }

  dy := make([]int, len(e.children))
  someDYSet := false
  if y < (maxHeight - e.padding[2]) {
    switch e.vAlign {
    case CENTER:
      dyAll := (maxHeight - y - e.padding[2])/2
      for i, _ := range e.children {
        dy[i] = dyAll
      }
      someDYSet = true
    case END:
      dyAll := maxHeight - y - e.padding[2]
      for i, _ := range e.children {
        dy[i] = dyAll
      }
      someDYSet = true
    case STRETCH:
      rem := maxHeight - y -e.padding[2]
      remPerChild := float64(rem)/float64(len(e.children)-1)

      for i, _ := range e.children {
        dy[i] = int(math.Floor(float64(i)*remPerChild))
      }
      someDYSet = true
    }
  }

  w := e.width
  if w < 0 {
    w = maxWidth
  } else if w < maxChildW + e.padding[1] + e.padding[3] {
    w = maxChildW + e.padding[1] + e.padding[3]
  }

  if someDYSet || e.hAlign != START {
    for i, child := range e.children {
      dx := 0

      switch e.hAlign {
      case CENTER:
        dx = (w - childWs[i] - e.padding[1] - e.padding[3])/2
      case END:
        dx = (w - childWs[i] - e.padding[1])
      }

      child.Translate(dx, dy[i])
    }
  }

  totalHeight := y + e.padding[2]
  if len(dy) > 0 {
    totalHeight += dy[len(dy)-1]
  }

  return e.InitRect(w, totalHeight)
}
