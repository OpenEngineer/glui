package glui

import (
)

//go:generate ./gen_element Body "A CalcDepth Padding Spacing"

type Body struct {
  ElementData
}

// windows can't be made transparent like this sadly, so alpha stays 255
func newBody(frame *Frame, isFirst bool) *Body {
  if isFirst && false{
    return &Body{
      newElementData(frame, 0, 0),
    }
  } else {
    e := &Body{
      newElementData(frame, 9*2, 0),
    }

    e.setStyle()

    return e
  }
}

func (e *Body) borderT() int {
  if len(e.p1Tris) == 0 {
    return 0
  } else {
    return e.Root.P1.Skin.ButtonBorderThickness()
  }
}

func (e *Body) setStyle() {
  if e.borderT() != 0 {
    e.SetButtonStyle()
  }
}

func (e *Body) Show() {
  e.setStyle()

  e.ElementData.Show()
}

func (e *Body) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  e.ElementData.CalcPosChildren(maxWidth - 2*e.borderT(), maxHeight - 2*e.borderT(), maxZIndex)

  if e.borderT() != 0 {
    e.SetBorderedElementPos(maxWidth, maxHeight, e.borderT(), maxZIndex)

    for _, child := range e.children {
      child.Translate(e.borderT(), e.borderT())
    }
  }

  return e.InitRect(maxWidth, maxHeight)
}
