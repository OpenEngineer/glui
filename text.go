package glui

import (
  "fmt"
  "math"

  "github.com/veandco/go-sdl2/sdl"
)

type Text struct {
  ElementData

  content string
  font    string
  size    float64

  // 2 tris per character (whitespace obviously doesnt get any tris)
  tris []uint32
  dd   *DrawData
}

func NewText(dd *DrawData, content string, font string, size float64) *Text {
  e := &Text{newElementData(), "", font, size, []uint32{}, dd}

  e.SetContent(content)

  return e
}

func (e *Text) A(children ...Element) Element {
  panic("text can't have children")
}

func countNonWhitespace(s string) int {
  count := 0

  for _, c := range s {
    if c != ' ' && c != '\n' && c != '\t' {
      count += 1
    }
  }

  return count
}

func (e *Text) SetContent(content string) {
  e.content = content

  n := countNonWhitespace(e.content)

  nDiff := n - len(e.tris)/2
  if nDiff > 0 {
    e.tris = append(e.tris, e.dd.P2.Alloc(nDiff*2)...)
  } else if nDiff < 0 {
    e.dd.P2.Dealloc(e.tris[n*2:])
    e.tris = e.tris[0:n*2]
  }

  for i := 0; i < n; i++ {
    tri0 := e.tris[i*2+0]
    tri1 := e.tris[i*2+1]

    e.dd.P2.Type.Set1Const(tri0, VTYPE_GLYPH)
    e.dd.P2.Type.Set1Const(tri1, VTYPE_GLYPH)
    e.dd.P2.SetColorConst(tri0, sdl.Color{0x00, 0x00, 0x00, 0xff})
    e.dd.P2.SetColorConst(tri1, sdl.Color{0x00, 0x00, 0x00, 0xff})
  }

  i := 0
  for _, c := range []rune(e.content) {
    if isWhitespace(c) {
      continue
    } else {
      tri0 := e.tris[i*2+0]
      tri1 := e.tris[i*2+1]

      e.dd.P2.SetGlyphCoords(tri0, tri1, fmt.Sprintf("%s:%d", e.font, c))

      i++
    }
  }
}

func isWhitespace(r rune) bool {
  return r == ' ' || r == '\n' || r == '\t'
}

// TODO: multiline depending on maxWidth
func (e *Text) OnResize(maxWidth, maxHeight int) (int, int) {
  baseline := e.size

  x := 0.0

  refG := e.dd.P2.Glyphs.GetGlyph(fmt.Sprintf("%s:%d", e.font, 'a'))
  refScale := refG.Scale
  space := refG.Advance*e.size/float64(GlyphResolution)

  i := 0
  runes := []rune(e.content)
  for _, c := range runes {
    if isWhitespace(c) {
      x += space
    } else {
      tri0 := e.tris[i*2+0]
      tri1 := e.tris[i*2+1]

      g := e.dd.P2.Glyphs.GetGlyph(fmt.Sprintf("%s:%d", e.font, c))

      size := e.size*refScale/g.Scale
      scale := size/float64(GlyphResolution)
      //size = float64(e.size) // DEBUG

      offsetX := g.OriginX*scale
      offsetY := g.OriginY*scale

      //DEBUG
      //offsetX = 0.0
      //offsetY = 0.0

      var advance float64
      if i < len(runes) - 1 {
        advance = g.GetAdvance(runes[i+1])*scale
      } else {
        advance = 0
      }

      //advance = size // DEBUG

      r := RectF{
        x - offsetX,
        baseline - offsetY,
        size,
        size,
      }

      /*rInner := RectF{
        g.Hints[3]*scale,
        g.Hints[0]*scale,
        (g.Hints[1] - g.Hints[3])*scale,
        (g.Hints[2] - g.Hints[0])*scale,
      }*/

      //r = r.Scale(rInner, rInner.Round())
        
      z := float32(x)*(0.001)

      e.dd.P2.SetQuadPosF(tri0, tri1, r, z)

      e.dd.P2.Param.Set1Const(tri0, float32(scale))
      e.dd.P2.Param.Set1Const(tri1, float32(scale))

      x += advance
      i++
    }
  }

  return e.ElementData.InitBB(int(math.Ceil(x)), int(e.size))
}

func (e *Text) Translate(dx, dy int) {
  for _, tri := range e.tris {
    e.dd.P2.TranslateTri(tri, dx, dy)
  }

  e.ElementData.Translate(dx, dy)
}
