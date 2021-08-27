package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

type ClassicSkin struct {
}

func setColor(d []byte, pixId int, r, g, b, a byte) {
  d[pixId*4+0] = r
  d[pixId*4+1] = g
  d[pixId*4+2] = b
  d[pixId*4+3] = a
}

func setTransparent(d []byte, pixId int) {
  d[pixId*4+3] = 0x00
}

// i is horizontal coord
// j is vertical coord
func ijToPix5x5(i, j int) int {
  return i*5 + j
}

func ijToPix3x3(i, j int) int {
  return i*3 + j
}

func setColor5x5(d []byte, i int, j int, r, g, b, a byte) {
  setColor(d, ijToPix5x5(i, j), r, g, b, a)
}

func setColor3x3(d []byte, i, j int, r, g, b, a byte) {
  setColor(d, ijToPix3x3(i, j), r, g, b, a)
}

func setTransparent5x5(d []byte, i int, j int) {
  setTransparent(d, ijToPix5x5(i, j))
}

func setColorGray(d []byte, pixId int, gray byte) {
  setColor(d, pixId, gray, gray, gray, 0xff)
}

func setColor5x5Gray(d []byte, i int, j int, gray byte) {
  setColor5x5(d, i, j, gray, gray, gray, 0xff)
}

func setColor3x3Gray(d []byte, i, j int, gray byte) {
  setColor3x3(d, i, j, gray, gray, gray, 0xff)
}

func setColorSDL(d []byte, pixId int, c sdl.Color) {
  d[pixId*4+0] = c.R
  d[pixId*4+1] = c.G
  d[pixId*4+2] = c.B
  d[pixId*4+3] = c.A
}

func swapColor(d []byte, pixId0 int, pixId1 int) {
  r := d[pixId0*4+0] 
  g := d[pixId0*4+1]
  b := d[pixId0*4+2]
  a := d[pixId0*4+3]

  d[pixId0*4+0] = d[pixId1*4+0]
  d[pixId0*4+1] = d[pixId1*4+1]
  d[pixId0*4+2] = d[pixId1*4+2]
  d[pixId0*4+3] = d[pixId1*4+3]

  d[pixId1*4+0] = r
  d[pixId1*4+1] = g
  d[pixId1*4+2] = b
  d[pixId1*4+3] = a
}

// in place
func flipX(d []byte, w int, h int) {
  for i := 0; i < w/2; i++ {
    for j := 0; j < h; j++ {
      a := i*h + j
      b := (w - 1 - i)*h + j

      swapColor(d, a, b)
    }
  }
}

func flipY(d []byte, w int, h int) {
  for i := 0; i < w; i++ {
    for j := 0; j < h/2; j++ {
      a := i*h + j
      b := i*h + (h - 1 - j)

      swapColor(d, a, b)
    }
  }
}

func (s *ClassicSkin) BGColor() sdl.Color {
  return sdl.Color{0xc0, 0xc0, 0xc0, 255}
}

func (s *ClassicSkin) SelColor() sdl.Color {
  return sdl.Color{0x00, 0x00, 196, 255}
}

func (s *ClassicSkin) twoPxOutsetColorBorder(c0 sdl.Color, c1 sdl.Color, c2 sdl.Color, c3 sdl.Color) []byte {
  d := make([]byte, 25*4)

  // left and top side
  for i := 0; i < 4; i++ {
    setColorSDL(d, i, c0)
    setColorSDL(d, i*5, c0)
  }

  // middel top left
  for i := 1; i < 3; i++ {
    for j := 1; j < 3; j++ {
      pixId := i*5+j
      setColorSDL(d, pixId, c1)
    }
  }

  // middel bottom right
  for i := 1; i < 4; i++ {
    setColorSDL(d, 3*5 + i, c2)
    setColorSDL(d, i*5 + 3, c2)
  }

  // bottom and right
  for i := 0; i < 5; i++ {
    setColorSDL(d, 4*5 + i, c3)
    setColorSDL(d, i*5 + 4, c3)
  }

  return d
}

func (s *ClassicSkin) twoPxOutsetBorder(c0 byte, c1 byte, c2 byte, c3 byte) []byte {
  d := make([]byte, 25*4)

  // left and top side
  for i := 0; i < 4; i++ {
    setColor5x5Gray(d, 0, i, c0)
    setColor5x5Gray(d, i, 0, c0)
  }

  // middel top left
  for i := 1; i < 3; i++ {
    for j := 1; j < 3; j++ {
      setColor5x5Gray(d, i, j, c1)
    }
  }

  // middel bottom right
  for i := 1; i < 4; i++ {
    setColor5x5Gray(d, 3, i, c2)
    setColor5x5Gray(d, i, 3, c2)
  }

  // bottom and right
  for i := 0; i < 5; i++ {
    setColor5x5Gray(d, 4, i, c3)
    setColor5x5Gray(d, i, 4, c3)
  }

  return d
}

func (s *ClassicSkin) Button() []byte {
  return s.twoPxOutsetBorder(0xff, 0xc0, 0x80, 0x00)
}

func (s *ClassicSkin) ButtonPressed() []byte {
  d := s.Button()

  flipX(d, 5, 5)
  flipY(d, 5, 5)

  return d
}

func (s *ClassicSkin) Input() []byte {
  d := s.twoPxOutsetBorder(0x80, 0x00, 0xc0, 0xff)

  setColor5x5Gray(d, 2, 2, 0xff)

  return d
}

func (s *ClassicSkin) Focus() []byte {
  c := s.SelColor()

  return s.twoPxOutsetColorBorder(c, c, c, c)
}

func (s *ClassicSkin) Inset() []byte {
  d := s.twoPxOutsetBorder(0x80, 0xc0, 0xc0, 0xff)

  // only outer rim has non-zero alpha

  for i := 1; i <= 3; i++ {
    for j := 1; j <= 3; j++ {
      setTransparent5x5(d, i, j)
    }
  }

  return d
}

func (s *ClassicSkin) Corner() []byte {
  d := make([]byte, 25*4)

  c0 := byte(0xff)
  c1 := byte(0xc0)
  c2 := byte(0x80)

  // top left of cross
  setColor5x5Gray(d, 0, 0, c0)
  setColor5x5Gray(d, 1, 0, c1)
  setColor5x5Gray(d, 0, 1, c1)
  setColor5x5Gray(d, 1, 1, c1)

  // left middle (same as button)
  setColor5x5Gray(d, 0, 2, c0)
  setColor5x5Gray(d, 1, 2, c1)

  // top right of cross
  setColor5x5Gray(d, 3, 0, c2)
  setColor5x5Gray(d, 4, 0, c0)
  setColor5x5Gray(d, 3, 1, c1)
  setColor5x5Gray(d, 4, 0, c1)

  // top middle (sams as button)
  setColor5x5Gray(d, 2, 0, c0)
  setColor5x5Gray(d, 2, 1, c1)

  return d
}

func (s *ClassicSkin) Bar() []byte {
  d := make([]byte, 9*4)

  c0 := byte(0xff)
  c1 := byte(0xc0)
  c2 := byte(0x80)

  setColor3x3Gray(d, 0, 0, c0)
  setColor3x3Gray(d, 1, 0, c0)
  setColor3x3Gray(d, 2, 0, c0)

  setColor3x3Gray(d, 0, 1, c0)
  setColor3x3Gray(d, 1, 1, c1)
  setColor3x3Gray(d, 2, 1, c2)

  setColor3x3Gray(d, 0, 2, c2)
  setColor3x3Gray(d, 1, 2, c2)
  setColor3x3Gray(d, 2, 2, c2)

  return d
}
