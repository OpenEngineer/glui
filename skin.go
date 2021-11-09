package glui

import (
  "math"

  "github.com/veandco/go-sdl2/sdl"
)

// each skin type returns a matrix of pixel colors (i.e. a bitmap)
// the byte order is row major, col minor, then color channels

type Skin interface {
  BGColor()       sdl.Color
  SelColor()      sdl.Color

  Button()        []byte // len(Button()) = N => 2*n+1 = sqrt(N)
  ButtonPressed() []byte // same length as Button()
  Corner()        []byte // for tab lips
  Input()         []byte
  Focus()         []byte
  Inset()         []byte
  Bar()           []byte // vertical bar, transposed to form horizontal bar

  RadioOff() []byte // square shape
  RadioOn()  []byte // square shape
  Tick()     []byte // square shape, determines size of checbox

  ScrollbarTrack() []byte // 1xn
}

func calcSquareSkinSize(d []byte) int {
  sqrtN := math.Sqrt(float64(len(d)/4))
  if math.Mod(sqrtN, 1.0) != 0.0 {
    panic("incorrect skin size")
  }

  return int(sqrtN)
}

func calcSkinThickness(d []byte) int {
  sqrtN := calcSquareSkinSize(d)

  t := (sqrtN - 1)/2

  return t
}

func calcSkinThicknessCheckRef(d []byte, tRef int) int {

  t := calcSkinThickness(d)

  if t != tRef {
    panic("skin border not equal to reference")
  }

  return t
}
