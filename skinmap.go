package glui

import (
  "math"
  "unsafe"

  "github.com/go-gl/gl/v4.1-core/gl"
  "github.com/veandco/go-sdl2/sdl"
)

type SkinMap struct {
  skin   Skin

  data   []byte
  width  int
  height int

  buttonX int
  buttonY int
  buttonT int
  buttonPressedX int
  buttonPressedY int

  inputX int
  inputY int
  inputT int

  loc uint32
  tid uint32
}

func NewSkinMap(s Skin) *SkinMap {
  d := &SkinMap{} // zero construct, because number of fields of SkinMap increases a lot

  d.skin = s
  d.genData(s)

  return d
}

func (sm *SkinMap) genData(s Skin) {
  tb := NewTextureBuilder(4, 128, 128)

  sm.genButtonData(s, tb)

  sm.genInputData(s, tb)

  //if err := tb.ToImage("debug.png"); err != nil {
    //panic(err)
  //}

  sm.data = tb.data
  sm.width = tb.width
  sm.height = tb.height
}

func (sm *SkinMap) genButtonData(s Skin, tb *TextureBuilder) {
  button := s.Button()
  sqrtN := math.Sqrt(float64(len(button)/4))
  if math.Mod(sqrtN, 1.0) != 0.0 {
    panic("button border skin incorrect size")
  }

  tButton := (int(sqrtN) - 1)/2

  sm.buttonX, sm.buttonY = tb.Build(button, 2*tButton+1, 2*tButton+1)
  sm.buttonT = tButton

  buttonPressed := s.ButtonPressed()
  if len(buttonPressed) != len(button) {
    panic("buttonPressed not same length as button")
  }

  sm.buttonPressedX, sm.buttonPressedY = tb.Build(buttonPressed, 2*tButton+1, 2*tButton+1)
}

func (sm *SkinMap) genInputData(s Skin, tb *TextureBuilder) {
  input := s.Input()
  sqrtN := math.Sqrt(float64(len(input)/4))
  if math.Mod(sqrtN, 1.0) != 0.0 {
    panic("button border skin incorrect size")
  }

  tInput := (int(sqrtN) - 1)/2

  sm.inputX, sm.inputY = tb.Build(input, 2*tInput+1, 2*tInput+1)
  sm.inputT = tInput
}

func (s *SkinMap) InitGL(loc uint32) {
  s.loc = loc

  gl.GenTextures(1, &s.tid)

  gl.ActiveTexture(gl.TEXTURE0)
  gl.Uniform1i(int32(s.loc), 0)

  gl.BindTexture(gl.TEXTURE_2D, s.tid)
  gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
  gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

  gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(s.width), int32(s.height), 0, gl.RGBA, 
    gl.UNSIGNED_BYTE, unsafe.Pointer(&(s.data[0])))

  gl.BindTexture(gl.TEXTURE_2D, 0)
}

func (s *SkinMap) bind() {
  gl.ActiveTexture(gl.TEXTURE0)
  gl.BindTexture(gl.TEXTURE_2D, s.tid)
}

func (s *SkinMap) ButtonOrigin() (int, int) {
  return s.buttonX, s.buttonY
}

func (s *SkinMap) ButtonPressedOrigin() (int, int) {
  return s.buttonPressedX, s.buttonPressedY
}

func (s *SkinMap) ButtonBorderThickness() int {
  return s.buttonT
}

func (s *SkinMap) BGColor() sdl.Color {
  return s.skin.BGColor()
}

func (s *SkinMap) InputOrigin() (int, int) {
  return s.inputX, s.inputY
}

func (s *SkinMap) InputBorderThickness() int {
  return s.inputT
}

func (s *SkinMap) InputBGColor() sdl.Color {
  i, j := s.InputOrigin()

  i += s.inputT
  j += s.inputT

  k := i*s.height + j

  r := s.data[k*4+0]
  g := s.data[k*4+1]
  b := s.data[k*4+2]
  a := s.data[k*4+3]

  return sdl.Color{r, g, b, a}
}
