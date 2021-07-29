package glui

var (
  GlyphResolution = 24
  GlyphDPerPx     = 85
)

type GlyphKerning struct {
  Next    rune
  Advance float64 // relative to regular Advance
}

type Glyph struct {
  Distances []byte
  Angles    []byte
  Hints     []float64
  Scale     float64
  Advance   float64
  OriginX   float64
  OriginY   float64
  Kerning   []GlyphKerning
  TexId     int  // set by GlyphMap
}

func (g *Glyph) GetAdvance(next rune) float64 {
  a := g.Advance
  for _, k := range g.Kerning {
    if k.Next == next {
      return a + k.Advance
    }
  }

  return a
}
