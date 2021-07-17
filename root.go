package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

type Root struct {
  iTest   int
  bgColor sdl.Color
}

func NewRoot() *Root {
  return &Root{
    0,
    sdl.Color{0,0,0,128},
  }
}

func (r *Root) BGColor() sdl.Color {
  return r.bgColor
}

// test function
func (r *Root) IncrementBGColor() {
  r.iTest += 1

  c := uint8(r.iTest*10%256)

  r.bgColor = sdl.Color{c, c, c, 128}
}
