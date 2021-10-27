package glui

import (
  "sort"
)

//go:generate ./gen_element Table "CalcDepth appendChild On Size Padding"

type Table struct {
  ElementData

  head []*Button
  body *tableBody
  headFocusNoTab bool
  masterCol      int

  currentSortCol Column
  currentSortDir SortState
}

// the table itself is styled like an input
func NewTable(root *Root) *Table {
  e := &Table{
    NewElementData(root, 9*2, 0), // after the first 18 tris come the sel tris
    make([]*Button, 0),
    newTableBody(root),
    false,
    -1,
    nil, UNSORTED,
  }

  e.width, e.height = 400, 400

  e.setTypesAndTCoords()

  e.On("click", e.onMouseClick)

  e.appendChild(e.body)

  e.body.On("focus", e.onFocusBody)
  e.body.On("blur",  e.onBlurBody)
  e.body.On("keypress", e.onKeyBody)

  return e
}

func (e *Table) MasterColumn(mc int) *Table {
  e.masterCol = mc

  return e
}

func (e *Table) LineHeight() int {
  return e.body.lineHeight()
}

func (e *Table) setTypesAndTCoords() {
  setInputLikeElementTypesAndTCoords(e.Root, e.p1Tris)
}

func (e *Table) borderT() int {
  return e.Root.P1.Skin.InputBorderThickness()
}

func (e *Table) NumSelected() int {
  return e.body.numSelected()
}

func (e *Table) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  w, h := 400, 400 //e.GetSize()

  if w > maxWidth {
    w = maxWidth
  }
  
  if h > maxHeight {
    h = maxHeight
  }

  e.body.CalcPos(w - 2*e.borderT(), h - 2*e.borderT() - e.LineHeight(), maxZIndex)

  e.body.Translate(e.borderT(), e.borderT() + e.LineHeight())

  nc := e.body.nColumns()

  x := 0
  for i := 0; i < nc; i++ {
    cw := e.body.getColumn(i).Rect().W

    hb := e.head[i]

    hb.Size(cw, e.LineHeight())

    hb.CalcPos(cw, e.LineHeight(), maxZIndex)

    hb.Translate(e.borderT() + x, e.borderT())

    x += cw
  }

  e.SetBorderedElementPos(w, h, e.borderT(), maxZIndex)

  return e.InitRect(w, h)
}

func (e *Table) Show() {
  e.body.syncSelection()

  e.ElementData.Show()
}

func (e *Table) onFocusBody(evt *Event) {
  // *tableBody should be last
  if evt.IsTab() {
    e.Root.FocusRect.Show(e.body)
  }

  e.body.showSelection()
}

func (e *Table) onBlurBody(evt *Event) {
  e.Root.FocusRect.Hide()

  e.body.blurSelection()
}

func (e *Table) onFocusHead(evt *Event) {
  if !evt.IsTab() {
    e.headFocusNoTab = true
    e.body.showSelection()
  } 
}

func (e *Table) onBlurHead(evt *Event) {
  e.body.blurSelection()
  e.headFocusNoTab = false
}

func (e *Table) onKeyHead(evt *Event) {
  if e.headFocusNoTab {
    e.onKeyBody(evt)
  }
}

// TODO: first two children should be scrollbars
func (e *Table) A(children ...Column) *Table {
  for _, child := range children {
    head := child.Head()
    e.head = append(e.head, head)
    e.appendChild(head)
    e.body.appendChild(child)
    head.On("focus", e.onFocusHead)
    head.On("blur", e.onBlurHead)
    head.On("keypress", e.onKeyHead)
  }

  // find the body element, and move to last place, which makes more sense from focus perspective
  found := false
  for i, child := range e.children {
    if _, ok := child.(*tableBody); ok {
      if i == 0 {
        e.children = append(e.children[1:], child)
      } else if i < len(e.children) - 1 {
        e.children = append(e.children[0:i], e.children[i+1:]...)
        e.children = append(e.children, child)
      }

      found = true
      break
    }
  }

  if !found {
    panic("no *tableBody child")
  }

  return e
}

func (e *Table) AddRow(data ...interface{}) *Table {
  e.body.addRow(data...)

  return e
}

func (e *Table) SortByColumn(c Column, dir SortState) {
  e.currentSortCol = c
  e.currentSortDir = dir

  sort.Sort(e)

  for i := 0; i < e.body.nColumns(); i++ {
    c_ := e.body.getColumn(i)
    if c_ == c {
      c_.SetSortState(dir)
    } else {
      c_.SetSortState(UNSORTED)
    }
  }

  if e.NumSelected() > 0 {
    e.body.syncSelection()
  }
}

func (e *Table) Len() int {
  return e.body.nRows()
}

func (e *Table) Less(i, j int) bool {
  if e.currentSortDir == ASCENDING {
    return e.currentSortCol.Less(i, j)
  } else {
    return e.currentSortCol.Less(j, i)
  }
}

func (e *Table) Swap(i, j int) {
  e.body.swap(i, j)
}

func (e *Table) Select(i int) {
  e.body.selectRow(i, false)
}

func (e *Table) onMouseClick(evt *Event) {
  _, y := evt.RelPos(e.Rect())

  i := ((y - e.borderT())/ e.LineHeight()) - 1

  if !evt.Shift {
    e.body.selectRow(i, evt.Ctrl)
  } else {
    prev := e.body.selPivot

    e.body.selectRow(i, false)

    // prev must be selected last, so it remains the pivot
    if prev < i {
      for i_ := i - 1; i_ >= prev; i_-- {
        e.body.selectRow(i_, true)
      }
    } else if prev > i {
      for i_ := i + 1; i_ <= prev; i_++ {
        e.body.selectRow(i_, true)
      }
    }
  }
}

func (e *Table) onKeyBody(evt *Event) {
  switch evt.Key {
  case "down":
    e.body.selectNextRow(evt.Shift)
  case "up":
    e.body.selectPrevRow(evt.Shift)
  }
}
