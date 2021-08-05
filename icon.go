package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

type Icon struct {
  ElementData

  name string
  size int
  mainColor sdl.Color

  shadow      bool
  shadowColor sdl.Color

  tris  []uint32
  dd    *DrawData
  glyph *Glyph
}

func NewIcon(dd *DrawData, name string, size int) *Icon {
  glyph := dd.P2.Glyphs.GetGlyph(name)

  color := sdl.Color{0x00, 0x00, 0x00, 0xff}

  shadowColor := sdl.Color{0xff, 0xff, 0xff, 0xff}

  // tri 0 and 1 are used for shadow, 2 and 3 are the actual icon
  tris := dd.P2.Alloc(2*2)

  e := &Icon{newElementData(), name, size, color, false, shadowColor, tris, dd, glyph}

  e.setTypeColorAndTCoord()

  return e
}

func (e *Icon) setTypeColorAndTCoord() {
  e.dd.P2.Type.Set1Const(e.tris[2], VTYPE_GLYPH)
  e.dd.P2.Type.Set1Const(e.tris[3], VTYPE_GLYPH)

  e.dd.P2.SetColorConst(e.tris[2], e.mainColor)
  e.dd.P2.SetColorConst(e.tris[3], e.mainColor)

  e.dd.P2.SetGlyphCoords(e.tris[2], e.tris[3], e.name)

  if e.shadow {
    e.dd.P2.Type.Set1Const(e.tris[0], VTYPE_GLYPH)
    e.dd.P2.Type.Set1Const(e.tris[1], VTYPE_GLYPH)
  } else {
    e.dd.P2.Type.Set1Const(e.tris[0], VTYPE_HIDDEN)
    e.dd.P2.Type.Set1Const(e.tris[1], VTYPE_HIDDEN)
  }

  e.dd.P2.SetColorConst(e.tris[0], e.shadowColor)
  e.dd.P2.SetColorConst(e.tris[1], e.shadowColor)

  e.dd.P2.SetGlyphCoords(e.tris[0], e.tris[1], e.name)

  scale := float64(e.size)/float64(GlyphResolution)
  for _, tri := range e.tris {
    e.dd.P2.Param.Set1Const(tri, float32(scale))
  }
}

func (e *Icon) OnResize(maxWidth, maxHeight int) (int, int) {
  // shadow tris are always set, even if they are hidden
  e.dd.P2.SetQuadPos(e.tris[2], e.tris[3], Rect{0, 0, e.size, e.size}, 0.5)

  e.dd.P2.SetQuadPos(e.tris[0], e.tris[1], Rect{2, 2, e.size, e.size}, 0.5)

  return e.size, e.size
}

func (e *Icon) Translate(dx, dy int, dz float32) {
  for _, tri := range e.tris {
    e.dd.P2.TranslateTri(tri, dx, dy, dz)
  }

  e.ElementData.Translate(dx, dy, dz)
}
