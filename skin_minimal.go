package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

type MinimalSkin struct {
}

func (s *MinimalSkin) BGColor() sdl.Color {
  return sdl.Color{255, 255, 255, 255}
}

func (s *MinimalSkin) Button() []byte {
  d := make([]byte, 9*4)

  // a simple black line
  for i := 0; i < 9; i++ {
    d[i*4+0] = 0
    d[i*4+1] = 0
    d[i*4+2] = 0
    d[i*4+3] = 255
  }

  return d
}

func (s *MinimalSkin) ButtonPressed() []byte {
  d := make([]byte, 9*4)

  // a simple black line
  for i := 0; i < 9; i++ {
    d[i*4+0] = 0x80
    d[i*4+1] = 0x80
    d[i*4+2] = 0x80
    d[i*4+3] = 255
  }

  return d
}
