package glui

import (
)

//go:generate ./gen_element tabPage "CalcDepth AContainer PaddingContainer SpacingContainer"

// we don't want to export tabPage, but instead let it be used as a Container
type tabPage struct {
  ElementData

  tabbed *Tabbed
}

func newTabPage(tabbed *Tabbed) *tabPage {
  e := &tabPage{
    NewElementData(tabbed.Root, 9*2, 0),
    tabbed,
  }

  e.Padding(10)

  e.Show()

  return e
}

func (e *tabPage) Show() {
  e.SetButtonStyle()

  e.ElementData.Show()
}

func (e *tabPage) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {

  w, h := e.CalcPosChildren(maxWidth, maxHeight, maxZIndex)

  if w > maxWidth {
    maxWidth = w
  }

  if h > maxHeight {
    maxHeight = h
  }

  e.width = maxWidth
  e.height = maxHeight

  e.SetButtonPos(maxWidth, maxHeight, maxZIndex)

  return e.InitRect(maxWidth, maxHeight)
}
