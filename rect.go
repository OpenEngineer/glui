package glui

import (
  "math"
)

// XXX: chain parent rects everytime absolute positioning is specified
type Rect struct {
  X int // coordinates of top left
  Y int
  W int
  H int
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

func (r Rect) Pos(fx, fy float64) (int, int) {
  x := int(math.Round(float64(r.X) + fx*float64(r.W)))
  y := int(math.Round(float64(r.Y) + fy*float64(r.H)))

  return x, y
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
