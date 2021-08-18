package glui

import (
  "fmt"

  "github.com/veandco/go-sdl2/sdl"
)

//go:generate ./gen_element Icon "CalcDepth"

type Icon struct {
  ElementData

  name string
  size int
  mainColor sdl.Color

  shadow      bool
  shadowColor sdl.Color

  glyph *Glyph
}

func NewIcon(root *Root, name string, size int) *Icon {
  glyph := root.P2.Glyphs.GetGlyph(name)

  color := sdl.Color{0x00, 0x00, 0x00, 0xff}

  shadowColor := sdl.Color{0xff, 0xff, 0xff, 0xff}

  // tri 0 and 1 are used for shadow, 2 and 3 are the actual icon
  e := &Icon{
    NewElementData(root, 0, 2*2), 
    name, 
    size, 
    color, 
    false, 
    shadowColor, 
    glyph,
  }

  e.setTypeColorAndTCoord()

  return e
}

func (e *Icon) setTypeColorAndTCoord() {
  tri0 := e.p2Tris[0]
  tri1 := e.p2Tris[1]
  tri2 := e.p2Tris[2]
  tri3 := e.p2Tris[3]

  e.Root.P2.Type.Set1Const(tri2, VTYPE_GLYPH)
  e.Root.P2.Type.Set1Const(tri3, VTYPE_GLYPH)

  e.Root.P2.SetColorConst(tri2, e.mainColor)
  e.Root.P2.SetColorConst(tri3, e.mainColor)

  e.Root.P2.SetGlyphCoords(tri2, tri3, e.name)

  if e.shadow {
    e.Root.P2.Type.Set1Const(tri0, VTYPE_GLYPH)
    e.Root.P2.Type.Set1Const(tri1, VTYPE_GLYPH)
  } else {
    e.Root.P2.Type.Set1Const(tri0, VTYPE_HIDDEN)
    e.Root.P2.Type.Set1Const(tri1, VTYPE_HIDDEN)
  }

  e.Root.P2.SetColorConst(tri0, e.shadowColor)
  e.Root.P2.SetColorConst(tri1, e.shadowColor)

  e.Root.P2.SetGlyphCoords(tri0, tri1, e.name)

  scale := float64(e.size)/float64(GlyphResolution)
  for _, tri := range e.p2Tris {
    e.Root.P2.Param.Set1Const(tri, float32(scale))
  }
}

func (e *Icon) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  fmt.Println("icon z index: ", e.ZIndex())
  z := e.Z(maxZIndex)
  s := e.size

  // shadow tris are always set, even if they are hidden
  e.Root.P2.SetQuadPos(e.p2Tris[2], e.p2Tris[3], Rect{0, 0, s, s}, z)

  e.Root.P2.SetQuadPos(e.p2Tris[0], e.p2Tris[1], Rect{2, 2, s, s}, z)

  return s, s
}
