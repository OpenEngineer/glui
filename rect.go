package glui

import (
  "math"
  "sort"
)

// XXX: chain parent rects everytime absolute positioning is specified
type Rect struct {
  X int // coordinates of top left
  Y int
  W int
  H int
}

type rectEdge struct {
  x0 int
  y0 int
  x1 int
  y1 int
}

type RectF struct {
  X float64
  Y float64
  W float64
  H float64
}

func (r Rect) Right() int {
  return r.X + r.W
}

func (r Rect) Bottom() int {
  return r.Y + r.H
}

func (r Rect) Hit(x, y int) bool {
  return (x >= r.X) && (x < r.Right()) && (y >= r.Y) && (y < r.Bottom())
}

func (r Rect) Translate(x, y int) Rect {
  return Rect{r.X + x, r.Y + y, r.W, r.H}
}

func (r Rect) Area() int {
  return r.W*r.H
}

func (r Rect) Pos(fx, fy float64) (int, int) {
  x := int(math.Round(float64(r.X) + fx*float64(r.W)))
  y := int(math.Round(float64(r.Y) + fy*float64(r.H)))

  return x, y
}

// other is appliead to this
func (r Rect) Common(other Rect) Rect {
  smallestXMax := int(math.Min(float64(r.Right()), float64(other.Right())))
  largestXMin := int(math.Max(float64(r.X), float64(other.X)))

  smallestYMax := int(math.Min(float64(r.Bottom()), float64(other.Bottom())))
  largestYMin := int(math.Max(float64(r.Y), float64(other.Y)))

  if smallestXMax > largestXMin && smallestYMax > largestYMin {
    return Rect{largestXMin, largestYMin, smallestXMax - largestXMin, smallestYMax - largestYMin}
  } else {
    return Rect{}
  }
}

func (r Rect) Merge(other Rect) Rect {
  x0 := r.X
  if other.X < x0 {
    x0 = other.X
  }

  y0 := r.Y
  if other.Y < y0 {
    y0 = other.Y
  }

  x1 := r.Right()
  if other.Right() > x1 {
    x1 = other.Right()
  }

  y1 := r.Bottom()
  if other.Bottom() > y1 {
    y1 = other.Bottom()
  }


  return Rect{x0, y0, x1 - x0, y1 - y0}
}

func (r Rect) edges() []rectEdge {
  return []rectEdge{
    rectEdge{r.X, r.Y, r.Right(), r.Y},
    rectEdge{r.Right(), r.Y, r.Right(), r.Bottom()},
    rectEdge{r.X, r.Bottom(), r.Right(), r.Bottom()},
    rectEdge{r.X, r.Y, r.X, r.Bottom()},
  }
}

func (e rectEdge) LenSq() int {
  dx := (e.x1 - e.x0)
  dy := (e.y1 - e.y0)

  return dx*dx + dy*dy
}

func (e rectEdge) Eq(other rectEdge) bool {
  return e.x0 == other.x0 && e.y0 == other.y0 && e.x1 == other.x1 && e.y1 == other.y1
}

func (e rectEdge) isHor() bool {
  return e.y0 == e.y1
}

func (e rectEdge) isVer() bool {
  return e.x0 == e.x1
}

func (e rectEdge) overlaps(other rectEdge) (rectEdge, bool, bool) {
  if e.isHor() && other.isHor() && e.y0 == other.y0 {
    y := e.y0

    if e.x0 == other.x0 {
      if other.x1 < e.x1 {
        return rectEdge{e.x0, y, other.x1, y}, true, true
      } else {
        return rectEdge{e.x0, y, e.x1, y}, false, true
      }
    } else if e.x1 == other.x1 {
      if other.x0 < e.x0 {
        return rectEdge{e.x0, y, e.x1, y}, false, true
      } else {
        return rectEdge{other.x0, y, e.x1, y}, true, true
      }
    }
  } else if e.isVer() && other.isVer() && e.x0 == other.x0 {
    x := e.x0

    if e.y0 == other.y0 {
      if other.y1 < e.y1 {
        return rectEdge{x, e.y0, x, other.y1}, true, true
      } else {
        return rectEdge{x, e.y0, x, e.y1}, false, true
      }
    } else if e.y1 == other.y1 {
      if other.y0 < e.y0 {
        return rectEdge{x, e.y0, x, e.y1}, false, true
      } else {
        return rectEdge{x, other.y0, x, e.y1}, true, true
      }
    }
  } 

  return rectEdge{}, false, false
}

func (r Rect) hasCommonEdge(other Rect) bool {
  aEdges := r.edges()
  bEdges := other.edges()

  for _, a := range aEdges {
    for _, b := range bEdges {
      if a.Eq(b) {
        return true
      }
    }
  }

  return false
}

func (r Rect) hasCommonPartialEdge(other Rect) (rectEdge, bool, bool) {
  aEdges := r.edges()
  bEdges := other.edges()

  anyFound := false
  longestOverlap := rectEdge{}
  thisIsOwner := false

  for _, a := range aEdges {
    for _, b := range bEdges {
      overlap, aIsOwner, ok := a.overlaps(b)
      if ok {
        if overlap.LenSq() > longestOverlap.LenSq() {
          anyFound = true
          longestOverlap = overlap
          thisIsOwner = aIsOwner
        }
      }
    }
  }

  return longestOverlap, thisIsOwner, anyFound
}

func (r Rect) MergeAlongEdge(other Rect) (Rect, bool) {
  if r.hasCommonEdge(other) {
    res := r.Merge(other)

    if res.Area() != r.Area() + other.Area() {
      panic("shouldn't be possible")
      return Rect{}, false
    } else {
      return res, true
    }
  } else {
    return Rect{}, false
  }
}

func (r Rect) cutByPartialEdge(e rectEdge) (Rect, Rect) {
  var (
    cut Rect
    rem Rect
  )

  if e.isHor() {
    if e.x0 == r.X {
      cut = Rect{r.X, r.Y, e.x1 - e.x0, r.H}
      rem = Rect{e.x1, r.Y, r.W - (e.x1 - e.x0), r.H}
    } else if e.x1 == r.Right() {
      cut = Rect{e.x0, r.Y, e.x1 - e.x0, r.H}
      rem = Rect{r.X, r.Y, r.W - (e.x1 - e.x0), r.H}
    } else {
      panic("neither side aligns")
    }
  } else {
    if e.y0 == r.Y {
      cut = Rect{r.X, r.Y, r.W, e.y1 - e.y0}
      rem = Rect{r.X, e.y1, r.W, r.H - (e.y1 - e.y0)}
    } else if e.y1 == r.Bottom() {
      cut = Rect{r.X, e.y0, r.W, e.y1 - e.y0}
      rem = Rect{r.X, r.Y, r.W, r.H - (e.y1 - e.y0)}
    } else {
      panic("neither side aligns")
    }
  }

  if cut.Area() + rem.Area() != r.Area() {
    panic("area not preserved")
  }

  return cut, rem
}

func (r Rect) MergeAlongPartialEdge(other Rect) (Rect, Rect, bool) {
  if common, thisIsOwner, ok := r.hasCommonPartialEdge(other); ok {
    if thisIsOwner {
      cut, rem := r.cutByPartialEdge(common)
      cut = cut.Merge(other)

      return cut, rem, true
    } else {
      cut, rem := other.cutByPartialEdge(common)
      cut = cut.Merge(r)

      return cut, rem, true
    }
  } else {
    return Rect{}, Rect{}, false
  }
}

type rectSorter struct {
  rects []Rect
}

func (s *rectSorter) Len() int {
  return len(s.rects)
}

func (s *rectSorter) Swap(i, j int) {
  s.rects[i], s.rects[j] = s.rects[j], s.rects[i]
}

func (s *rectSorter) Less(i, j int) bool {
  rI := s.rects[i]
  rJ := s.rects[j]

  return rI.W*rI.H > rJ.W*rJ.H
}

func sortRects(r []Rect) []Rect {
  s := &rectSorter{r}

  sort.Sort(s)

  return s.rects
}

func (r RectF) Right() float64 {
  return r.X + r.W
}

func (r RectF) Bottom() float64 {
  return r.Y + r.H
}

func (r RectF) Hit(x, y float64) bool {
  return (x >= r.X) && (x < r.Right()) && (y >= r.Y) && (y < r.Bottom())
}

func (r RectF) Round() RectF {
  x := math.Round(r.X)
  //y := math.Round(r.Y)

  //return RectF{x, y, math.Round(r.X + r.W) - x, math.Round(r.Y + r.H) - y}
  //return RectF{x, y, r.W, r.H}

  return RectF{x, r.Y, r.W, math.Round(r.Y + r.H) - r.Y}
}

func (r RectF) Scale(inner0 RectF, inner1 RectF) RectF {
  sw := inner1.W/inner0.W
  sh := inner1.H/inner0.H

  dxBefore := inner0.X*(sw - 1.0)
  dyBefore := inner0.Y*(sh - 1.0)

  dxAfter := inner1.X - inner0.X
  dyAfter := inner1.Y - inner0.Y

  return RectF{r.X - dxBefore + dxAfter, r.Y - dyBefore + dyAfter, r.W*sw, r.H*sh}
}
