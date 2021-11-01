package glui

import (
  "unsafe"

  "github.com/go-gl/gl/v4.1-core/gl"
)

type GlyphMap struct {
  glyphs map[string]*Glyph

  size int
  data []byte

  loc uint32
  tid uint32
  tunit uint32
}

func glyphTextureSize(nGlyphs int) int {
  n := 0
  size := 64

  for n*n < nGlyphs {
    size = size*2

    n = size/GlyphResolution
  }

  return size
}

func newGlyphMap(glyphs map[string]*Glyph) *GlyphMap {
  texSize := glyphTextureSize(len(glyphs))
  
  data := make([]byte, texSize*texSize*4)

  n := texSize/GlyphResolution

  glyphLst := make([]*Glyph, 0)
  for _, g := range glyphs {
    g.TexId = len(glyphLst)
    glyphLst = append(glyphLst, g)
  }

  for i := 0; i < n; i++ {
    for j := 0; j < n; j++ {
      k := i*n + j

      if k < len(glyphs) {
        g := glyphLst[k]

        for ii := 0; ii < GlyphResolution; ii++ {
          for jj := 0; jj < GlyphResolution; jj++ {
            src := ii*GlyphResolution + jj
            dst := (i*GlyphResolution + ii)*texSize + j*GlyphResolution + jj

            data[dst*4+0] = g.Distances[src]
            data[dst*4+1] = g.Angles[src]
            data[dst*4+2] = 0
            data[dst*4+3] = 255
          }
        }
      }
    }
  }

  return &GlyphMap{glyphs, texSize, data, 0, 0, 0}
}

func (g *GlyphMap) initGL(uTexLoc uint32, texID uint32, texUnit uint32) {
  g.loc = uTexLoc
  g.tid = texID
  g.tunit = texUnit

  gl.ActiveTexture(g.tunit)
  gl.BindTexture(gl.TEXTURE_2D, g.tid)
  gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
  gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

  gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(g.size), int32(g.size), 0, gl.RGBA, 
    gl.UNSIGNED_BYTE, unsafe.Pointer(&(g.data[0])))

  gl.BindTexture(gl.TEXTURE_2D, 0)

  g.ToImage("glyphs.png")
  checkGLError()
}

func (g *GlyphMap) bind() {
  checkGLError()
  gl.ActiveTexture(g.tunit)
  gl.BindTexture(gl.TEXTURE_2D, g.tid)
  checkGLError()
}

func (g *GlyphMap) GetGlyph(name string) *Glyph {
  glyph, ok := g.glyphs[name]

  if !ok {
    for _, glyph_ := range g.glyphs {
      return glyph_
    }

    panic("no glyphs found (hint: use glyph_maker)")
  } else {
    return glyph
  }
}

func (g *GlyphMap) GetRect(glyphId int) Rect {
  // top left corner
  n := g.size/GlyphResolution

  i := glyphId/n
  j := glyphId%n

  return Rect{i*GlyphResolution, j*GlyphResolution, GlyphResolution, GlyphResolution}
}

func (g *GlyphMap) ToImage(fname string) error {
  return DataToImage(g.data, g.size, g.size, fname)
}
