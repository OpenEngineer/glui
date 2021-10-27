package glui

import (
  "time"
)

const (
  DATE_FMT = "2006-01-02"
  DEFAULT_COLUMN_PADDING = 3
)

// IMPORTANT: only concrete types are allowed to implement appendChild!

//go:generate ./gen_element TextColumn "appendChild CalcDepth"
//go:generate ./gen_element DateColumn "appendChild CalcDepth"

type Column interface {
  Element

  AddRow(x interface{})

  SetSortState(ss SortState)

  Less(i, j int) bool
  Len() int
  Swap(i, j int)

  Select(i int, keepOld bool)
  Deselect(i int)
  ClearSelection()

  Head() *Button
}

type BasicColumn struct {
  ElementData

  align Align

  sortState SortState

  table *Table // reference to parent Table
  head  *Button
  arrow *Icon
  body  []*Text
}

func (e *BasicColumn) lineH() int {
  return e.table.LineHeight()
}

type TextColumn struct {
  BasicColumn
  rows []string
}

type DateColumn struct {
  BasicColumn
  rows []time.Time
}

func NewBasicColumn(root *Root, align Align) *BasicColumn {
  c := &BasicColumn{
    NewElementData(root, 0, 0),
    align,
    UNSORTED,
    nil, // registered later
    nil,
    nil,
    make([]*Text, 0),
  }

  c.padding = [4]int{0, DEFAULT_COLUMN_PADDING, 0, DEFAULT_COLUMN_PADDING}

  return c
}

func (e *BasicColumn) Head() *Button {
  return e.head
}

func (e *BasicColumn) Select(i int, keepOld bool) {
  for j, txt := range e.body {
    if i == j {
      txt.SetColor(WHITE)
    } else if !keepOld {
      txt.SetColor(BLACK)
    }
  }
}

func (e *BasicColumn) Deselect(i int) {
  e.body[i].SetColor(BLACK)
}

func (e *BasicColumn) ClearSelection() {
  for _, txt := range e.body {
    txt.SetColor(BLACK)
  }
}

func (e *BasicColumn) SetSortState(ss SortState) {
  e.sortState = ss

  switch ss {
  case ASCENDING:
    e.arrow.ChangeGlyph("arrow-down-drop")
    e.arrow.Show()
  case DESCENDING:
    e.arrow.ChangeGlyph("arrow-up-drop")
    e.arrow.Show()
  default:
    e.arrow.Hide()
  }
}

func (e *BasicColumn) RegisterParent(parent Element) {
  tb, ok := parent.(*tableBody)
  if !ok {
    panic("parent of Column must be tableBody")
  }

  e.table = tb.table()

  e.ElementData.RegisterParent(parent)
}

func (e *BasicColumn) Len() int {
  return len(e.body)
}

func (e *BasicColumn) Show() {
  e.ElementData.Show()

  if e.sortState != ASCENDING && e.sortState != DESCENDING {
    e.arrow.Hide()
  } 
}

func (e *BasicColumn) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  y := 0

  textWidths := make([]int, len(e.body))
  for i, textElem := range e.body {
    w, h := textElem.CalcPos(maxWidth - e.padding[1] - e.padding[3], e.lineH(), maxZIndex)

    wPlusPadding := w + e.padding[1] + e.padding[3]

    if wPlusPadding > maxWidth {
      maxWidth = wPlusPadding
    }

    textWidths[i] = w

    textElem.Translate(0, y + (e.lineH() - h)/2)

    y += e.lineH()
  }

  for i, textElem := range e.body {
    w := textWidths[i]
    dx := e.padding[3]

    switch e.align {
    case CENTER:
      dx = (maxWidth - w)/2
    case END:
      dx = maxWidth - w - e.padding[1]
    }

    textElem.Translate(dx, 0)
  }

  return e.InitRect(maxWidth, y)
}

func newHeadButton(root *Root, caption string) (*Button, *Icon) {
  b := NewButton(root)

  hor := NewHor(root, STRETCH, CENTER, 0)
  hor.Padding(0, DEFAULT_COLUMN_PADDING, 0, DEFAULT_COLUMN_PADDING)

  icon := NewIcon(root, "arrow-down-drop", 10)
  icon.Hide() // i.e. UNSORTED

  hor.A(NewSans(root, caption, 10), icon)

  b.A(hor)

  return b, icon
}

func NewTextColumn(root *Root, caption string) *TextColumn {
  e_ := NewBasicColumn(root, START)

  e := &TextColumn{
    *e_,
    make([]string, 0),
  }

  e.head, e.arrow = newHeadButton(root, caption)
  e.head.OnClick(e.onClickSort)

  return e
}

func (e *TextColumn) AddRow(x_ interface{}) {
  x, ok := x_.(string)
  if !ok {
    panic("row entry is not a string")
  }

  txt := NewSans(e.Root, x, 10)
  e.rows = append(e.rows, x)
  e.body = append(e.body, txt)
  e.appendChild(txt)
}

func (e *TextColumn) onClickSort() {
  e.table.SortByColumn(e, NextSortState(e.sortState))
}

func (e *BasicColumn) swap(i, j int) {
  e.body[i], e.body[j] = e.body[j], e.body[i]

  e.children[i], e.children[j] = e.children[j], e.children[i]
}

func (e *TextColumn) Less(i, j int) bool {
  return e.rows[i] < e.rows[j]
}

func (e *TextColumn) Swap(i, j int) {
  e.rows[i], e.rows[j] = e.rows[j], e.rows[i]

  e.BasicColumn.swap(i, j)
}

func NewDateColumn(root *Root, caption string) *DateColumn {
  e_ := NewBasicColumn(root, END)

  e := &DateColumn{
    *e_, 
    make([]time.Time, 0),
  }

  e.head, e.arrow = newHeadButton(root, caption)
  e.head.OnClick(e.onClickSort)

  return e
}

func (e *DateColumn) AddRow(x_ interface{}) {
  var t time.Time

  switch x := x_.(type) {
  case time.Time:
    t = x
  case string:
    var err error
    t, err = time.Parse(DATE_FMT, x)
    if err != nil {
      panic(err)
    }

  default:
    panic("row entry is not time.Time")
  }

  e.rows = append(e.rows, t)
  txt := NewMono(e.Root, t.Format(DATE_FMT), 10)
  e.body = append(e.body, txt)
  e.appendChild(txt)
}

func (e *DateColumn) onClickSort() {
  e.table.SortByColumn(e, NextSortState(e.sortState))
}

func (e *DateColumn) Less(i, j int) bool {
  return e.rows[i].Before(e.rows[j])
}

func (e *DateColumn) Swap(i, j int) {
  e.rows[i], e.rows[j] = e.rows[j], e.rows[i]

  e.BasicColumn.swap(i, j)
}
