package glui

import (
)

//go:generate ./gen_element radioItem "On CalcDepth appendChild"
//go:generate ./gen_element RadioGroup "On CalcDepth appendChild"

type radioItem struct {
  ElementData

  group *RadioGroup
  selected bool
}

type RadioGroup struct {
  ElementData
  
  options []string
  fillHor bool // false -> vertical fill
}

func NewRadioGroup(options []string, fillHor bool) *RadioGroup {
  e := &RadioGroup{
    NewElementData(0, 0),
    options,
    fillHor,
  }

  for _, option := range options {
    item := newRadioItem(e, option, false)

    e.appendChild(item)
  }

  e.spacing = 10

  e.On("focus", e.onFocus)
  e.On("blur", e.onBlur)
  e.On("keypress", e.onKeyPress)

  return e
}

func newRadioItem(group *RadioGroup, caption string, selected bool) *radioItem {
  e := &radioItem{
    NewElementData(2, 0),
    group,
    selected,
  }

  e.setTypesAndTCoords()

  txt := NewSans(caption, 10)

  e.appendChild(txt)
  e.spacing = 10

  e.On("click", e.onMouseClick)

  return e
}

func (e *radioItem) Cursor() int {
  return e.ButtonCursor(e.enabled)
}

func (e *radioItem) onMouseClick(evt *Event) {
  e.group.selectItem(e)
}

func (e *radioItem) Show() {
  e.setTypesAndTCoords()

  e.ElementData.Show()
}

func (e *radioItem) select_() {
  e.selected = true

  e.setTypesAndTCoords()
}

func (e *radioItem) deselect() {
  e.selected = false

  e.setTypesAndTCoords()
}

func (e *radioItem) text() *Text {
  if e.nChildren() != 1 {
    panic("expected 1 child")
  }

  txt, ok := e.children[0].(*Text)
  if !ok {
    panic("child isn't text")
  }

  return txt
}

func (e *radioItem) radioSize() int {
  return e.Root.P1.Skin.RadioSize()
}

func (e *radioItem) getSkinCoords() ([4]int, [4]int) {
  var (
    x0 int
    y0 int
    x [4]int
    y [4]int
  )

  if e.selected {
    x0, y0 = e.Root.P1.Skin.RadioOnOrigin()
  } else {
    x0, y0 = e.Root.P1.Skin.RadioOffOrigin()
  }

  x[0] = x0
  x[1] = x0 + e.radioSize()

  y[0] = y0
  y[1] = y0 + e.radioSize()

  return x, y
}

func (e *radioItem) setTypesAndTCoords() {
  tri0 := e.p1Tris[0]
  tri1 := e.p1Tris[1]

  e.Root.P1.SetTriType(tri0, VTYPE_SKIN)
  e.Root.P1.SetTriType(tri1, VTYPE_SKIN)

  e.Root.P1.Color.Set4Const(tri0, 1.0, 1.0, 1.0, 1.0)
  e.Root.P1.Color.Set4Const(tri1, 1.0, 1.0, 1.0, 1.0)

  x, y := e.getSkinCoords()

  e.Root.P1.setQuadSkinCoords(tri0, tri1, 0, 0, x, y)
}

func (e *radioItem) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  x := e.radioSize() + e.spacing

  tri0 := e.p1Tris[0]
  tri1 := e.p1Tris[1]

  z := e.Z(maxZIndex)
  e.Root.P1.SetQuadPos(tri0, tri1, Rect{0, 0, e.radioSize(), e.radioSize()}, z)

  txt := e.text()

  txtW, txtH := txt.CalcPos(maxWidth - x, maxHeight, maxZIndex)

  h := e.radioSize()
  dy := (h - txtH)/2

  if txtH > h {
    h = txtH
  }

  txt.Translate(x, dy)


  return e.InitRect(x + txtW, h)
}

func (e *RadioGroup) item(i int) *radioItem {
  if i < 0 || i >= len(e.options) {
    panic("i out of range")
  }

  item, ok := e.children[i].(*radioItem)
  if !ok {
    panic("child not radioItem")
  }

  return item
}

func (e *RadioGroup) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  x := e.padding[3]
  y := e.padding[0]

  if e.fillHor {
    lineH := 0
    totalW := 0
    for _, child := range e.children {
      cw, ch := child.CalcPos(maxWidth - x - e.padding[1], maxHeight - y - e.padding[2], maxZIndex)

      if lineH == 0 {
        // first item of line
        child.Translate(x, y)

        lineH = ch 
      } else {
        if ch > lineH {
          lineH = ch
        }

        if x + cw > (maxWidth -e.padding[1]) {
          x = e.padding[3]
          y += lineH
          lineH = 0
        }

        child.Translate(x, y)
      }

      x += cw

      if x > totalW {
        totalW = x
      }
    }

    return e.InitRect(totalW + e.padding[1], y + lineH + e.padding[2])
  } else {
    colW := 0
    totalH := 0
    for i, child := range e.children {
      cw, ch := child.CalcPos(maxWidth - x - e.padding[1], maxHeight - y - e.padding[2], maxZIndex)

      if colW == 0 {
        // first item of column
        child.Translate(x, y)

        colW = cw
      } else {
        if cw > colW {
          colW = cw
        }
        
        if y + ch > (maxHeight - e.padding[2]) {
          x += colW
          y = e.padding[0]

          colW = 0
        }

        child.Translate(x, y)
      }

      y += ch
      if i < e.nChildren() - 1 {
        y += e.spacing
      }

      if y > totalH {
        totalH = y
      }
    }

    return e.InitRect(x + colW + e.padding[1], totalH + e.padding[2])
  }
}

func (e *RadioGroup) selectItem(item *radioItem) {
  for i := 0; i < e.nChildren(); i++ {
    otherItem := e.item(i)

    if item == otherItem {
      otherItem.select_()
    } else {
      otherItem.deselect()
    }
  }
}

func (e *RadioGroup) onFocus(evt *Event) {
  if evt.IsKeyboardEvent() {
    e.Root.FocusRect.Show(e)
  }
}

func (e *RadioGroup) onBlur(evt *Event) {
  e.Root.FocusRect.Hide()
}

func (e *RadioGroup) Value() string {
  i := e.Index()
  if i < 0 {
    return ""
  } else {
    return e.options[i]
  }
}

func (e *RadioGroup) Index() int {
  for i := 0; i < e.nChildren(); i++ {
    item := e.item(i)

    if item.selected {
      return i
    }
  }

  return -1
}

func (e *RadioGroup) selectIndex(i int) {
  item := e.item(i)

  e.selectItem(item)
}

func (e *RadioGroup) onKeyPress(evt *Event) {
  if evt.Key == "down" {
    i := e.Index()

    if i == -1 || i == len(e.options) - 1 {
      e.selectIndex(0)
    } else {
      e.selectIndex(i + 1)
    }
  } else if evt.Key == "up" {
    i := e.Index()

    if i == -1 || i == 0 {
      e.selectIndex(len(e.options)-1)
    } else {
      e.selectIndex(i - 1)
    }
  }
}
