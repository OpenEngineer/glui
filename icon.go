package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

//go:generate ./gen_element Icon "CalcDepth"

type Icon struct {
  ElementData

  name string
  size int
  orientation Orientation
  mainColor sdl.Color

  shadow      bool
  shadowColor sdl.Color

  glyph *Glyph
}

func NewIcon(name string, size int) *Icon {
  color := sdl.Color{0x00, 0x00, 0x00, 0xff}

  shadowColor := sdl.Color{0xff, 0xff, 0xff, 0xff}

  // tri 0 and 1 are used for shadow, 2 and 3 are the actual icon
  e := &Icon{
    NewElementData(0, 2*2), 
    name, 
    size, 
    HOR,
    color, 
    false, 
    shadowColor, 
    nil,
  }

  e.glyph = e.Root.P2.Glyphs.GetGlyph(name)

  e.Show()

  return e
}

func (e *Icon) SetOrientation(or Orientation) *Icon {
  if or != e.orientation {
    e.orientation = or

    e.updateGlyphCoords()
  }

  return e
}

func (e *Icon) ChangeGlyph(name string) {
  e.name = name
  e.glyph = e.Root.P2.Glyphs.GetGlyph(name)

  e.updateGlyphCoords()
}

func (e *Icon) updateGlyphCoords() {
  tri0 := e.p2Tris[0]
  tri1 := e.p2Tris[1]
  tri2 := e.p2Tris[2]
  tri3 := e.p2Tris[3]

  if e.orientation == HOR {
    e.Root.P2.SetGlyphCoords(tri0, tri1, e.name)
    e.Root.P2.SetGlyphCoords(tri2, tri3, e.name)
  } else {
    e.Root.P2.SetGlyphCoordsT(tri0, tri1, e.name)
    e.Root.P2.SetGlyphCoordsT(tri2, tri3, e.name)
  }
}

func (e *Icon) Show() {
  tri0 := e.p2Tris[0]
  tri1 := e.p2Tris[1]
  tri2 := e.p2Tris[2]
  tri3 := e.p2Tris[3]

  e.Root.P2.SetTriType(tri2, VTYPE_GLYPH)
  e.Root.P2.SetTriType(tri3, VTYPE_GLYPH)

  e.Root.P2.SetColorConst(tri2, e.mainColor)
  e.Root.P2.SetColorConst(tri3, e.mainColor)


  if e.shadow {
    e.Root.P2.SetTriType(tri0, VTYPE_GLYPH)
    e.Root.P2.SetTriType(tri1, VTYPE_GLYPH)
  } else {
    e.Root.P2.SetTriType(tri0, VTYPE_HIDDEN)
    e.Root.P2.SetTriType(tri1, VTYPE_HIDDEN)
  }

  e.Root.P2.SetColorConst(tri0, e.shadowColor)
  e.Root.P2.SetColorConst(tri1, e.shadowColor)

  e.updateGlyphCoords()

  scale := float64(e.size)/float64(GlyphResolution)
  for _, tri := range e.p2Tris {
    e.Root.P2.Param.Set1Const(tri, float32(scale))
  }
}

func (e *Icon) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  z := e.Z(maxZIndex)
  s := e.size

  // shadow tris are always set, even if they are hidden
  e.Root.P2.SetQuadPos(e.p2Tris[2], e.p2Tris[3], Rect{0, 0, s, s}, z)

  e.Root.P2.SetQuadPos(e.p2Tris[0], e.p2Tris[1], Rect{2, 2, s, s}, z)

  return s, s
}
