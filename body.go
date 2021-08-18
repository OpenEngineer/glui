package glui

import (
)

//go:generate ./gen_element Body "A CalcDepth Padding Spacing"

type Body struct {
  ElementData
}

// windows can't be made transparent like this sadly, so alpha stays 255
func newBody(root *Root) *Body {
  return &Body{
    NewElementData(root, 0, 0),
  }
}

func (e *Body) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  e.ElementData.CalcPosChildren(maxWidth, maxHeight, maxZIndex)

  return e.InitRect(maxWidth, maxHeight)
}
