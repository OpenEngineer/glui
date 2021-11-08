package glui

import (
  "fmt"
  "math"

  "github.com/veandco/go-sdl2/sdl"
)

//go:generate ./gen_element Text "CalcDepth On"

var (
  DEFAULT_MONO = "dejavumono"
  DEFAULT_SANS = "dejavusans"
  BLACK = sdl.Color{0x00, 0x00, 0x00, 0xff}
  WHITE = sdl.Color{0xff, 0xff, 0xff, 0xff}
)

type Text struct {
  ElementData

  content string
  font    string
  size    float64
  color   sdl.Color

  refGlyph *Glyph
}

func NewSans(content string, size float64) *Text {
  return NewText(content, DEFAULT_SANS, size)
}

func NewMono(content string, size float64) *Text {
  return NewText(content ,DEFAULT_MONO, size)
}

func NewText(content string, font string, size float64) *Text {
  e := &Text{NewElementData(0, 0), "", font, size, BLACK, nil}

  e.refGlyph = e.Root.P2.Glyphs.GetGlyph(fmt.Sprintf("%s:%d", font, 'a')) 

  e.SetContent(content)

  return e
}

func (e *Text) Value() string {
  return e.content
}

func (e *Text) SetColor(c sdl.Color) {
  e.color = c

  for _, tri := range e.p2Tris {
    e.Root.P2.SetColorConst(tri, c)
  }
}

func countNonWhitespace(s string) int {
  count := 0

  for _, c := range s {
    if !isWhitespace(rune(c)) {
      count += 1
    }
  }

  return count
}

func (e *Text) SetContent(content string) {
  e.content = content

  e.Show()
}

func (e *Text) Show() {
  n := countNonWhitespace(e.content)

  e.p2Tris = e.Root.P2.Resize(e.p2Tris, n*2)
  /*nDiff := n - len(e.p2Tris)/2 // old code
  if nDiff > 0 {
    e.p2Tris = append(e.p2Tris, e.Root.P2.Alloc(nDiff*2)...)
  } else if nDiff < 0 {
    remove := e.p2Tris[n*2:]
    e.Root.P2.Dealloc(remove)
    e.p2Tris = e.p2Tris[0:n*2]
  }*/

  for i := 0; i < n; i++ {
    tri0 := e.p2Tris[i*2+0]
    tri1 := e.p2Tris[i*2+1]

    e.Root.P2.SetTriType(tri0, VTYPE_GLYPH)
    e.Root.P2.SetTriType(tri1, VTYPE_GLYPH)
    e.Root.P2.SetColorConst(tri0, e.color)
    e.Root.P2.SetColorConst(tri1, e.color)
  }

  i := 0
  for _, c := range []rune(e.content) {
    if isWhitespace(c) {
      continue
    } else {
      tri0 := e.p2Tris[i*2+0]
      tri1 := e.p2Tris[i*2+1]

      e.Root.P2.SetGlyphCoords(tri0, tri1, fmt.Sprintf("%s:%d", e.font, c))

      i++
    }
  }
}

func isWhitespace(r rune) bool {
  return r == ' ' || r == '\n' || r == '\t'
}

func isDelimiter(r rune) bool {
  return isWhitespace(r) || r == '"' || r =='\'' || r == '('  || r == ')' || r == '[' || r == ']' || r == '{' || r == '}' || r == '|' || r == ':' || r == ',' || r == ';' || r == '`' || r == '/' || r == '\\' || r == '.' || r == '-'
}

// useful for mono fonts in Input
func (e *Text) RefAdvance() float64 {
  return math.Ceil(e.refGlyph.Advance*e.size/float64(GlyphResolution))
}

// TODO: multiline depending on maxWidth
func (e *Text) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  z := float32(maxZIndex - e.zIndex)/float32(maxZIndex)
  baseline := e.size

  x := 0.0

  refScale := e.refGlyph.Scale
  space := math.Ceil(e.refGlyph.Advance*e.size/float64(GlyphResolution))

  i := 0
  runes := []rune(e.content)
  for _, c := range runes {
    if isWhitespace(c) {
      x += space
    } else {
      tri0 := e.p2Tris[i*2+0]
      tri1 := e.p2Tris[i*2+1]

      g := e.Root.P2.Glyphs.GetGlyph(fmt.Sprintf("%s:%d", e.font, c))

      size := e.size*refScale/g.Scale
      scale := size/float64(GlyphResolution)

      offsetX := g.OriginX*scale
      offsetY := g.OriginY*scale

      var advance float64
      if i < len(runes) - 1 {
        advance = math.Ceil(g.GetAdvance(runes[i+1])*scale)
      } else {
        advance = math.Ceil(g.Advance*scale)
      }

      r := RectF{
        x - offsetX,
        baseline - offsetY,
        size,
        size,
      }
        
      e.Root.P2.SetQuadPosF(tri0, tri1, r, z)

      e.Root.P2.Param.Set1Const(tri0, float32(scale))
      e.Root.P2.Param.Set1Const(tri1, float32(scale))

      x += advance
      i++
    }
  }

  return e.InitRect(int(math.Ceil(x)), int(e.size))
}
