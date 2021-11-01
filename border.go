package glui

import (
)

func getBorderedSkinCoords(x0, y0 int, t int) ([4]int, [4]int) {
  var (
    x [4]int
    y [4]int
  )

  x[0] = x0
  x[1] = x0 + t
  x[2] = x0 + t+1
  x[3] = x0 + 2*t+1

  y[0] = y0
  y[1] = y0 + t
  y[2] = y0 + t+1
  y[3] = y0 + 2*t+1

  return x, y
}
