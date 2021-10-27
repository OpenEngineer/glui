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
  hover        bool
  activeBar    int
  barDelta     int

  startX       int
  startY       int
  startLeft    int
  startRight   int
}

func NewVSplit(root *Root) *VSplit {
  e := &VSplit{
    NewElementData(root, 0, 0),
    make([]int, 0),
    false,
    -1, 0,
    0,0,0,0,
  }

  e.spacing = 5

  e.On("mousemove", e.onMouseMove)
  e.On("mousedown", e.onMouseDown)
  e.On("mouseup",   e.onMouseUp)

  return e
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
  return e.children[i].Rect().Right() + e.spacing
}

func (e *VSplit) hitBar(evt *Event) int {
  y0 := e.rect.Y
  h := e.rect.H

  margin := 1
  w := e.Root.P1.Skin.BarThickness() + margin*2
  // check if we are hovering over any 
  for i := 0; i < e.nBars(); i++ {
    x0 := e.barPos(i) - margin

    r := Rect{x0, y0, w, h}

    if r.Hit(evt.X, evt.Y) {
      return i
    }
  }

  return -1
}

func (e *VSplit) onMouseMove(evt *Event) {
  i := e.hitBar(evt)
  if i > -1 || e.activeBar > -1 {
    e.hover = true

    if e.activeBar > -1 {
      e.moveActiveBar(evt.XRel) // TODO: dont use XRel, but the actual start pos
    }
  } else {
    e.hover = false
  }
}

func (e *VSplit) onMouseDown(evt *Event) {
  i := e.hitBar(evt)
  if i > -1 {
    e.activeBar = i
    e.startX = evt.X
    e.startY = evt.Y
    e.startLeft = e.intervals[i]
    e.startRight = e.intervals[i]
  }
}

func (e *VSplit) onMouseUp(evt *Event) {
  e.activeBar = -1

  e.onMouseMove(evt)
}

// its unlikely that a resize of the window is triggered while the mouse is down
func (e *VSplit) moveActiveBar(delta int) {
  e.barDelta = delta

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

    texX_, texY := getBarSkinCoords(e.Root)
    texX := [4]int{texX_[0], texX_[1], 0, 0}

    for barI := 0; barI < n; barI++ {
      for j := 0; j < 3; j++ {
        tri0 := e.p1Tris[barI*TRIS_PER_BAR + j*2 + 0]
        tri1 := e.p1Tris[barI*TRIS_PER_BAR + j*2 + 1]

        e.Root.P1.Type.Set1Const(tri0, VTYPE_SKIN)
        e.Root.P1.Type.Set1Const(tri1, VTYPE_SKIN)
        e.Root.P1.SetColorConst(tri0, sdl.Color{0xff, 0xff, 0xff, 0xff})
        e.Root.P1.SetColorConst(tri1, sdl.Color{0xff, 0xff, 0xff, 0xff})

        setQuadSkinCoords(e.Root, tri0, tri1, 0, j, texX, texY)
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

func (e *VSplit) addToIntervalsAfter(iSmallerThanExpected int, diff int) {
  nAfter := len(e.intervals) - 1 - iSmallerThanExpected

  if nAfter > 0 {
    diffPerInterval := diff/nAfter
    lastIntervalDiff := diffPerInterval + (diff - diffPerInterval*nAfter)

    for i := iSmallerThanExpected + 1; i < len(e.intervals); i++ {

      if i == len(e.intervals) - 1 {
        e.intervals[i] += lastIntervalDiff
      } else {
        e.intervals[i] += diffPerInterval
      }
    }
  }
}

// TODO: this function is kind of slow due to the inner loop
func (e *VSplit) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  e.fillUnsetIntervals(maxWidth)

  barSpace := e.childSpacing()

  x := e.padding[3]

  children := e.children
  cHeight := 0
  cWidths := make([]int, len(children))

  // incorporate the intended movement of the bar
  iMovePrev := -1
  iMoveNext := -1
  if e.activeBar != -1 {
    iMovePrev = e.activeBar
    iMoveNext = e.activeBar + 1
  }

  intervalsResolved := false

  backup := make([]int, len(e.intervals))
  for i, in := range e.intervals {
    backup[i] = in
  }

  for ; !intervalsResolved; {
    intervalsResolved = true

    for i, child := range children {
      wantedInterval := e.intervals[i]
      if i == len(children) - 1 {
        wantedInterval = maxWidth - e.padding[1] - x // last child always gets the remaining
      } else if i == iMovePrev {
        wantedInterval = e.intervals[i] + e.barDelta // child before moving bar gets less/more
      } else if i == iMoveNext {
        wantedInterval = e.intervals[i] - e.barDelta // child after moving bar gets more/less
      }

      w, h := child.CalcPos(wantedInterval, maxHeight - e.padding[0] - e.padding[2], maxZIndex)

      if h > cHeight {
        cHeight = h
      }

      cWidths[i] = w
      e.intervals[i] = w

      // actual size of child might be less, in which case remaining children get more
      if w < wantedInterval {
        diff := wantedInterval - w

        e.addToIntervalsAfter(i, diff)
      } else if w > wantedInterval && e.barDelta != 0 {
        intervalsResolved = false
        e.barDelta = 0

        for i, bin := range backup {
          e.intervals[i] = bin
        }

        x = e.padding[3]

        break
      }

      child.Translate(x, 0)

      x += w + barSpace
    }
  }

  // now we know the innerheight we can calculate the bar positions
  x = e.padding[3]
  for i, _ := range children {
    x += cWidths[i]

    if i < len(children) - 1 {
      e.calcPosBar(i, x, e.padding[0], cHeight, maxZIndex)

      x += barSpace
    }
  }

  e.barDelta = 0

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
