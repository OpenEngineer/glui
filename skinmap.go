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
  buttonN int
  buttonPressedX int
  buttonPressedY int

  loc uint32
  tid uint32
}

func NewSkinMap(s Skin) *SkinMap {
  d := &SkinMap{} // zero construct, because number of fields of SkinMap increases a lot

  d.skin = s
  d.genData(s)

  return d
}

func (sd *SkinMap) genData(s Skin) {
  tb := NewTextureBuilder(4, 128, 128)

  button := s.Button()
  sqrtN := math.Sqrt(float64(len(button)/4))
  if math.Mod(sqrtN, 1.0) != 0.0 {
    panic("button border skin incorrect size")
  }

  nButton := (int(sqrtN) - 1)/2

  sd.buttonX, sd.buttonY = tb.Build(button, 2*nButton+1, 2*nButton+1)
  sd.buttonN = nButton

  buttonPressed := s.ButtonPressed()
  if len(buttonPressed) != len(button) {
    panic("buttonPressed not same length as button")
  }

  sd.buttonPressedX, sd.buttonPressedY = tb.Build(buttonPressed, 2*nButton+1, 2*nButton+1)

  //if err := tb.ToImage("debug.png"); err != nil {
    //panic(err)
  //}

  sd.data = tb.data
  sd.width = tb.width
  sd.height = tb.height
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
  return s.buttonN
}

func (s *SkinMap) BGColor() sdl.Color {
  return s.skin.BGColor()
}
