package glui

import (
  "fmt"
  "reflect"
)

//go:generate ./gen_element tabPage "CalcDepth AContainer PaddingContainer SpacingContainer"

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
  e.width = maxWidth
  e.height = maxHeight

  e.SetButtonPos(maxWidth, maxHeight, maxZIndex)

  e.CalcPosChildren(maxWidth, maxHeight, maxZIndex)

  return e.InitRect(maxWidth, maxHeight)
}

func (e *tabPage) Hide() {
  fmt.Println("hiding tab: ", e.tabbed.tabIndex(e))

  for _, child := range e.children {
    fmt.Println("might hide ", reflect.TypeOf(child).String())
  }

  e.ElementData.Hide()
}
