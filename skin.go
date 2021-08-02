package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

// rgba bitmaps! TODO: row major or col major
type Skin interface {
  BGColor()       sdl.Color
  Button()        []byte // len(Button()) = N => 2*n+1 = sqrt(N)
  ButtonPressed() []byte // same length as Button()
  Input()         []byte
}
