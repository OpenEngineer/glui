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
