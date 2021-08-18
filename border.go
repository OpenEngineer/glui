package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

func showBorderedElement(root *Root, tris []uint32) {
  for i := 0; i < 3; i++ {
    for j := 0; j < 3; j++ {
      tri0 := tris[(i*3 + j)*2 + 0]
      tri1 := tris[(i*3 + j)*2 + 1]

      if (i == 1 && j == 1) {
        root.P1.Type.Set1Const(tri0, VTYPE_PLAIN)
        root.P1.Type.Set1Const(tri1, VTYPE_PLAIN)
      } else {
        root.P1.Type.Set1Const(tri0, VTYPE_SKIN)
        root.P1.Type.Set1Const(tri1, VTYPE_SKIN)
      }
    }
  }
}

// also used by input
func setBorderedElementTypesAndTCoords(root *Root, tris []uint32, x0, y0 int, t int, bgColor sdl.Color) {
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

  for i := 0; i < 3; i++ {
    for j := 0; j < 3; j++ {
      tri0 := tris[(i*3 + j)*2 + 0]
      tri1 := tris[(i*3 + j)*2 + 1]

      if (i == 1 && j == 1) {
        root.P1.Type.Set1Const(tri0, VTYPE_PLAIN)
        root.P1.SetColorConst(tri0, bgColor)
        //root.P1.TCoord.Set2Const(tri0, 0.0, 0.0)

        root.P1.Type.Set1Const(tri1, VTYPE_PLAIN)
        root.P1.SetColorConst(tri1, bgColor)
        //root.P1.TCoord.Set2Const(tri1, 0.0, 0.0)
      } else {
        root.P1.Type.Set1Const(tri0, VTYPE_SKIN)
        root.P1.Color.Set4Const(tri0, 1.0, 1.0, 1.0, 1.0)
        root.P1.SetSkinCoord(tri0, 0, x[i], y[j])
        root.P1.SetSkinCoord(tri0, 1, x[i+1], y[j])
        root.P1.SetSkinCoord(tri0, 2, x[i], y[j+1])

        root.P1.Type.Set1Const(tri1, VTYPE_SKIN)
        root.P1.Color.Set4Const(tri1, 1.0, 1.0, 1.0, 1.0)
        root.P1.SetSkinCoord(tri1, 0, x[i+1], y[j+1])
        root.P1.SetSkinCoord(tri1, 1, x[i+1], y[j])
        root.P1.SetSkinCoord(tri1, 2, x[i], y[j+1])
      }
    }
  }
}

func setBorderedElementPos(root *Root, tris []uint32, width, height, t int, z float32) {
  var (
    x [4]int
    y [4]int
  )

  x[0] = 0
  x[1] = x[0] + t
  x[2] = x[0] + width - t
  x[3] = x[0] + width

  y[0] = 0
  y[1] = y[0] + t
  y[2] = y[0] + height - t
  y[3] = y[0] + height

  for i := 0; i < 3; i++ {
    for j := 0; j < 3; j++ {
      tri0 := tris[(i*3 + j)*2 + 0]
      tri1 := tris[(i*3 + j)*2 + 1]

      root.P1.SetPos(tri0, 0, x[i], y[j], z)
      root.P1.SetPos(tri0, 1, x[i+1], y[j], z)
      root.P1.SetPos(tri0, 2, x[i], y[j+1], z)

      root.P1.SetPos(tri1, 0, x[i+1], y[j+1], z)
      root.P1.SetPos(tri1, 1, x[i+1], y[j], z)
      root.P1.SetPos(tri1, 2, x[i], y[j+1], z)
    }
  }
}
