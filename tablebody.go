package glui

import (
  "strconv"
)

//go:generate ./gen_element tableBody "CalcDepth appendChild On"

type tableBody struct {
  ElementData

  selState []bool
  selPivot int
  showSel  bool
}

func newTableBody() *tableBody {
  return &tableBody{
    NewElementData(0, 0),
    make([]bool, 0),
    -1,
    false,
  }
}

func (e *tableBody) lineHeight() int {
  return 25
}

func (e *tableBody) getColumn(i int) Column {
  c_ := e.children[i]

  c, ok := c_.(Column)
  if !ok {
    panic("child of Table is not a Column")
  }

  return c
}

func (e *tableBody) table() *Table {
  p := e.Parent()

  if t, ok := p.(*Table); ok {
    return t
  } else {
    panic("parent of tableBody must be Table")
  }
}

func (e *tableBody) nColumns() int {
  return e.nChildren()
}

func (e *tableBody) numSelected() int {
  count := 0

  for _, s := range e.selState {
    if s {
      count++
    }
  }

  return count
}

func (e *tableBody) nRows() int {
  return len(e.selState)
}

func (e *tableBody) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  nc := e.nColumns()
  mc := e.table().masterCol

  if mc < 0 || mc > nc - 1 {
    // no masterColumn, each column is the same width
    wpc := maxWidth/nc

    for i := 0; i < nc; i++ {
      c := e.getColumn(i)

      c.CalcPos(wpc, maxHeight, maxZIndex)

      c.Translate(i*wpc, 0)
    }
  } else {
    // masterColumn gets the remaining space
    remWidth := maxWidth

    colWidths := make([]int, nc)

    for i := 0; i < nc; i++ {
      if i != mc {
        c := e.getColumn(i)

        actColWidth, _ := c.CalcPos(1, maxHeight, maxZIndex)

        colWidths[i] = actColWidth

        remWidth -= actColWidth
      }
    }

    c := e.getColumn(mc)

    actColWidth, _ := c.CalcPos(remWidth, maxHeight, maxZIndex)

    colWidths[mc] = actColWidth

    maxWidth = 0 
    cumW := 0
    for i, colW := range colWidths {
      maxWidth += colW

      if i > 0 {
        c := e.getColumn(i)

        c.Translate(cumW, 0)
      }

      cumW += colW
    }
  }


  selI := 0
  for i, selState := range e.selState {
    if selState {
      tri0 := e.p1Tris[selI*2]
      tri1 := e.p1Tris[selI*2+1]

      r := Rect{0, i*e.lineHeight(), maxWidth, e.lineHeight()}
      z := e.Z(maxZIndex - 1)

      e.Root.P1.SetQuadPos(tri0, tri1, r, z)
      
      selI++
    }
  }

  return e.InitRect(maxWidth, maxHeight)
}

func (e *tableBody) addRow(data ...interface{}) {
  if len(data) != e.nColumns() {
    panic("expected " + strconv.Itoa(e.nColumns()) + " row entries, got " + strconv.Itoa(len(data)))
  }

  for i, x := range data {
    c := e.getColumn(i)

    c.AddRow(x)
  }

  e.selState = append(e.selState, false)
}

func (e *tableBody) syncSelection() {
  numSel := e.numSelected()
  e.p1Tris = e.Root.P1.Resize(e.p1Tris, numSel*2)
  selI := 0
  for i, selState := range e.selState {
    if selState {
      tri0 := e.p1Tris[selI*2]
      tri1 := e.p1Tris[selI*2+1]

      // set tri color

      if e.showSel {
        e.Root.P1.Type.Set1Const(tri0, VTYPE_PLAIN)
        e.Root.P1.Type.Set1Const(tri1, VTYPE_PLAIN)

        e.Root.P1.SetColorConst(tri0, e.Root.P1.Skin.SelColor())
        e.Root.P1.SetColorConst(tri1, e.Root.P1.Skin.SelColor())

        for ic := 0; ic < e.nColumns(); ic++ {
          c := e.getColumn(ic)

          c.Select(i, true)
        }
      } else {
        e.Root.P1.Type.Set1Const(tri0, VTYPE_HIDDEN)
        e.Root.P1.Type.Set1Const(tri1, VTYPE_HIDDEN)
      }

      selI++
    }
  }

  if !e.showSel {
    for ic := 0; ic < e.nColumns(); ic++ {
      c := e.getColumn(ic)

      c.ClearSelection()
    }
  }
}

func (e *tableBody) swap(i, j int) {
  e.selState[i], e.selState[j] = e.selState[j], e.selState[i]

  for ic := 0; ic < e.nColumns(); ic++ {
    c := e.getColumn(ic)

    c.Swap(i, j)
  }

  if e.selPivot == i {
    e.selPivot = j
  } else if e.selPivot == j {
    e.selPivot = i
  }
}

func (e *tableBody) selectRow(i int, keepOld bool) {
  if i < 0 || i >= e.nRows() {
    return
  }

  if e.selState[i] {
    if keepOld {
      // deselect
      e.selState[i] = false

      for ic := 0; ic < e.nColumns(); ic++ {
        c := e.getColumn(ic)

        c.Deselect(i)
      }

    }  else {
      for is := 0; is < e.nRows(); is++ {
        e.selState[is] = is == i
      }

      for ic := 0; ic < e.nColumns(); ic++ {
        c := e.getColumn(ic)

        c.Select(i, false)
      }
    }
  } else {
    // only select  one row for now
    for is := 0; is < e.nRows(); is++ {
      if is == i {
        e.selState[is] = true
      } else if !keepOld {
        e.selState[is] = false
      }
    }

    // broadcast to each column as wel, so text color can be inverted
    for ic := 0; ic < e.nColumns(); ic++ {
      c := e.getColumn(ic)

      c.Select(i, keepOld)
    }
  }

  // update selPivot
  e.selPivot = i

  if keepOld {
    if e.pivotAtTop() {
      for e.selPivot < e.nRows() && !e.selState[e.selPivot] {
        e.selPivot++
      }

      if e.selPivot == e.nRows() {
        e.selPivot = -1
      }
    } else if e.pivotAtBottom() {
      for e.selPivot > -1 && !e.selState[e.selPivot] {
        e.selPivot--
      }
    }
  }

  e.syncSelection()
}

func (e *tableBody) deselectRow(i int) {
  if i < 0 || i >= e.nRows() {
    return
  }

  if e.selState[i] {
    e.selState[i] = false

    // broadcast to each column as wel, so text color can be inverted
    for ic := 0; ic < e.nColumns(); ic++ {
      c := e.getColumn(ic)

      c.Deselect(i)
    }

    e.syncSelection()
  }
}

func (e *tableBody) pivotAtTop() bool {
  if e.selPivot == -1 {
    return false
  }

  for i := 0; i < e.selPivot; i++ {
    if e.selState[i] {
      return false
    }
  }

  return true
}

func (e *tableBody) pivotAtBottom() bool {
  if e.selPivot == -1 {
    return false
  }

  for i := e.selPivot + 1; i < e.nRows(); i++ {
    if e.selState[i] {
      return false
    }
  }

  return true
}

func (e *tableBody) findNextSelected(iRow int) int {
  for i := iRow+1; i < e.nRows(); i++ {
    if e.selState[i] {
      return i
    }
  }

  return -1
}

func (e *tableBody) findPrevSelected(iRow int) int {
  for i := iRow-1; i >= 0; i-- {
    if e.selState[i] {
      return i
    }
  }

  return -1
}

func (e *tableBody) selectNextRow(keepOld bool) {
  if e.nRows() == 0 {
    return
  }

  cur := e.selPivot

  if cur == -1 {
    e.selectRow(0, keepOld)
  } else if cur == e.nRows() - 1 {
    if !keepOld {
      e.selectRow(0, false)
    }
  } else {
    if keepOld {
      if e.pivotAtBottom() {
        e.selectRow(cur+1, true)
      } else {
        e.selPivot = e.findNextSelected(cur)

        for i := 0; i < e.selPivot; i++ {
          e.deselectRow(i)
        }
      }
    } else {
      e.selectRow(cur+1, false)
    }
  } 


  e.syncSelection()
}

func (e *tableBody) selectPrevRow(keepOld bool) {
  n := e.nRows()
  if n == 0 {
    return
  }

  cur := e.selPivot

  if cur == -1 {
    e.selectRow(n-1, keepOld)
  } else if cur > 0 {
    if keepOld {
      if e.pivotAtTop() {
        e.selectRow(cur - 1, true)
      } else {
        e.selPivot = e.findPrevSelected(cur)

        for i := e.nRows() - 1; i > e.selPivot; i-- {
          e.deselectRow(i)
        }
      }
    } else {
      e.selectRow(cur-1, false)
    }
  } else if !keepOld {
    e.selectRow(n - 1, false)
  }

  e.syncSelection()
}

func (e *tableBody) blurSelection() {
  e.showSel = false

  e.syncSelection()
}

func (e *tableBody) showSelection() {
  e.showSel = true

  e.syncSelection()
}
