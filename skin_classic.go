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
    setColor(d, i, c0, c0, c0, 0xff)
    setColor(d, i*5, c0, c0, c0, 0xff)
  }

  // middel top left
  for i := 1; i < 3; i++ {
    for j := 1; j < 3; j++ {
      pixId := i*5+j
      setColor(d, pixId, c1, c1, c1, 0xff)
    }
  }

  // middel bottom right
  for i := 1; i < 4; i++ {
    setColor(d, 3*5 + i, c2, c2, c2, 0xff)
    setColor(d, i*5 + 3, c2, c2, c2, 0xff)
  }

  // bottom and right
  for i := 0; i < 5; i++ {
    setColor(d, 4*5 + i, c3, c3, c3, 0xff)
    setColor(d, i*5 + 4, c3, c3, c3, 0xff)
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

  i := 2
  j := 2

  setColor(d, i*5 + j, 0xff, 0xff, 0xff, 0xff)

  return d
}

func (s *ClassicSkin) Focus() []byte {
  c := s.SelColor()

  return s.twoPxOutsetColorBorder(c, c, c, c)
}

func (s *ClassicSkin) Inset() []byte {
  return s.twoPxOutsetBorder(0x80, 0xc0, 0xc0, 0xff)
}
