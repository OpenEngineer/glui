package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

//go:generate ./gen_element VSplit "A CalcDepth Spacing Padding On"

const TRIS_PER_BAR = 6

// like Hor with START, but with additional moveable splitbars between the children
type VSplit struct {
  ElementData

  intervals    []int // split equally if interval is unknown (i.e. interval==-1)
  minIntervals []int
  hover        bool

  activeBar    int
  startX       int
  startY       int
  startLeft    int
  startRight   int
}

func NewVSplit() *VSplit {
  e := &VSplit{
    NewElementData(0, 0),
    make([]int, 0),
    nil,
    false,
    -1, 0,0,0,0,
  }

  e.spacing = 5

  e.On("mousemove", e.onMouseMove)
  e.On("mousedown", e.onMouseDown)
  e.On("mouseup",   e.onMouseUp)

  return e
}

func (e *VSplit) MinIntervals(minInterv []int) {
  e.minIntervals = minInterv
}

func (e *VSplit) nBars() int {
  n := e.nChildren() - 1

  if n < 0 {
    return 0
  } else {
    return n
  }
}

func (e *VSplit) barPos(i int) int {
  x := e.Rect().X

  for i_ := 0; i_ <= i; i++ {
    x += e.intervals[i_]

    if i_ < i {
      x += e.childSpacing()
    }
  }

  return x + e.spacing
}

func (e *VSplit) hitBar(evt *Event) int {
  x0 := e.rect.X
  y0 := e.rect.Y
  h := e.rect.H

  margin := 1
  w := e.Root.P1.Skin.BarThickness() + margin*2
  // check if we are hovering over any vsplit bar

  for i := 0; i < e.nBars(); i++ {
    x0 += e.intervals[i] + e.spacing

    r := Rect{x0 - margin, y0, w, h}

    if r.Hit(evt.X, evt.Y) {
      return i
    }

    x0 += e.childSpacing()
  }

  return -1
}

func (e *VSplit) onMouseMove(evt *Event) {
  i := e.hitBar(evt)
  if i > -1 || e.activeBar > -1 {
    e.hover = true

    if e.activeBar > -1 {
      e.moveActiveBar(evt.X - e.startX)
    }
  } else {
    e.hover = false
  }
}

// bound by mininterval
func (e *VSplit) setInterval(i int, dLeft int, dRight int) {
  if e.minIntervals != nil {
    if i >= 0 && i < len(e.minIntervals) {
      if dLeft < e.minIntervals[i] {
        diff := e.minIntervals[i] - dLeft
        dLeft = e.minIntervals[i]
        dRight -= diff
      }
    }

    if i+1 < len(e.minIntervals) {
      if dRight < e.minIntervals[i+1] {
        diff := e.minIntervals[i+1] - dRight
        dRight = e.minIntervals[i+1]
        dLeft -= diff
      }
    }
  }

  e.intervals[i] = dLeft
  e.intervals[i+1] = dRight
}

func (e *VSplit) onMouseDown(evt *Event) {
  i := e.hitBar(evt)
  if i > -1 {
    e.activeBar = i
    e.startX = evt.X
    e.startY = evt.Y
    e.startLeft = e.intervals[i]
    e.startRight = e.intervals[i+1]
  }
}

func (e *VSplit) onMouseUp(evt *Event) {
  e.activeBar = -1

  e.onMouseMove(evt)
}

// its unlikely that a resize of the window is triggered while the mouse is down
func (e *VSplit) moveActiveBar(delta int) {
  e.setInterval(e.activeBar, e.startLeft + delta, e.startRight - delta)

  e.Root.ForcePosDirty()
}

func (e *VSplit) childSpacing() int {
  return e.spacing*2 + e.Root.P1.Skin.BarThickness()
}

func (e *VSplit) Show() {
  n := e.nBars()

  // n is the number of bars, not the number of children
  if n >= 1 {
    // make sure there are enough vbars
    nDiff := n - len(e.p1Tris)/TRIS_PER_BAR

    if nDiff > 0 {
      e.p1Tris = append(e.p1Tris, e.Root.P1.Alloc(nDiff*TRIS_PER_BAR)...)
    } else if nDiff < 0 {
      remove := e.p1Tris[n*TRIS_PER_BAR:]
      e.Root.P1.Dealloc(remove)
      e.p1Tris = e.p1Tris[0:n*TRIS_PER_BAR]
    }

    texX_, texY := e.Root.P1.Skin.getBarCoords()
    texX := [4]int{texX_[0], texX_[1], 0, 0}

    for barI := 0; barI < n; barI++ {
      for j := 0; j < 3; j++ {
        tri0 := e.p1Tris[barI*TRIS_PER_BAR + j*2 + 0]
        tri1 := e.p1Tris[barI*TRIS_PER_BAR + j*2 + 1]

        e.Root.P1.Type.Set1Const(tri0, VTYPE_SKIN)
        e.Root.P1.Type.Set1Const(tri1, VTYPE_SKIN)
        e.Root.P1.SetColorConst(tri0, sdl.Color{0xff, 0xff, 0xff, 0xff})
        e.Root.P1.SetColorConst(tri1, sdl.Color{0xff, 0xff, 0xff, 0xff})

        e.Root.P1.setQuadSkinCoords(tri0, tri1, 0, j, texX, texY)
      }
    }
  }

  e.ElementData.Show()
}

func (e *VSplit) fillUnsetIntervals(maxWidth int) {
  n := len(e.children)

  nDiff := n - len(e.intervals)

  if nDiff > 0 {
    for i := 0; i < nDiff; i++ {
      e.intervals = append(e.intervals, -1)
    }
  } else if nDiff < 0 {
    e.intervals = e.intervals[0:n]
  }

  // calculate the unset intervals
  barSpace := e.childSpacing()

  if maxWidth > -1 {
    setIntervalSum := 0
    unsetCount := 0
    for _, interval := range e.intervals {
      if interval != -1 {
        setIntervalSum += interval
      } else {
        unsetCount += 1
      }
    }

    if unsetCount > 0 {
      availableForUnset := maxWidth - e.padding[1] - e.padding[3] - setIntervalSum - e.nBars()*barSpace
      if availableForUnset < 0 {
        availableForUnset = 0
      }

      intervalPerUnset := availableForUnset/unsetCount

      lastIntervalIfUnset := intervalPerUnset + (availableForUnset - intervalPerUnset*unsetCount)

      for i, interval := range e.intervals {
        if interval == -1 {
          if i == len(e.intervals) - 1 {
            e.intervals[i] = lastIntervalIfUnset
          } else {
            e.intervals[i] = intervalPerUnset
          }
        }
      }
    }
  }
}

func (e *VSplit) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  e.fillUnsetIntervals(maxWidth)

  barSpace := e.childSpacing()

  x := e.padding[3]

  children := e.children
  cHeight := 0

  for i, child := range children {
    w := e.intervals[i]
    if i == len(children) - 1 {
      // last child gets remaining width
      w = maxWidth - x
      e.intervals[i] = w
    }

    _, h := child.CalcPos(w, maxHeight - e.padding[0] - e.padding[2], maxZIndex)
    if h > cHeight {
      cHeight = h
    }

    child.Translate(x, 0)

    x += w + barSpace
  }

  // now we know the innerheight we can calculate the bar positions
  x = e.padding[3]
  for i, _ := range children {
    x += e.intervals[i]

    if i < len(children) - 1 {
      e.calcPosBar(i, x, e.padding[0], cHeight, maxZIndex)

      x += barSpace
    }
  }

  return e.InitRect(x + e.padding[1], cHeight + e.padding[0] + e.padding[2])
}

func (e *VSplit) calcPosBar(barI int, xLeft int, yTop int, h int, maxZIndex int) {
  xLeft += e.spacing

  w := e.Root.P1.Skin.BarThickness()

  dt := (w - 1)/2

  var (
    y [4]int
  )

  y[0] = yTop
  y[1] = yTop + dt
  y[2] = yTop + h - dt
  y[3] = yTop + h

  z := e.Z(maxZIndex)

  for j := 0; j < 3; j++ {
    tri0 := e.p1Tris[barI*TRIS_PER_BAR + j*2 + 0]
    tri1 := e.p1Tris[barI*TRIS_PER_BAR + j*2 + 1]

    e.Root.P1.SetQuadPos(tri0, tri1, Rect{xLeft, y[j], w, y[j+1] - y[j]}, z)
  }
}

func (e *VSplit) Cursor() int {
  if e.hover {
    return sdl.SYSTEM_CURSOR_SIZEWE
  } else {
    return -1
  }
}
