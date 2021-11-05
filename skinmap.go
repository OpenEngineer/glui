package glui

import (
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

  focusX int
  focusY int
  focusT int

  insetX int
  insetY int

  cornerX int
  cornerY int

  barX int
  barY int
  barT int // not the border thickness of the bar, but the actual intended bar thickness

  radioOffX int
  radioOffY int
  radioOnX  int
  radioOnY  int
  radioSize int // size of one side

  tickX int
  tickY int
  tickSize int // internal size of one side

  loc uint32
  tid uint32
  tunit uint32
}

func newSkinMap(s Skin) *SkinMap {
  d := &SkinMap{} // zero construct, because number of fields of SkinMap increases a lot

  d.skin = s
  d.genData(s)

  return d
}

func (sm *SkinMap) genData(s Skin) {
  tb := NewTextureBuilder(4, 128, 128)

  sm.genButtonData(s, tb)

  sm.genInputData(s, tb)

  sm.genFocusData(s, tb)

  sm.genInsetData(s, tb)

  sm.genCornerData(s, tb)

  sm.genBarData(s, tb)

  sm.genRadioOffData(s, tb)

  sm.genRadioOnData(s, tb)

  sm.genTickData(s, tb)

  if err := tb.ToImage("skin.png"); err != nil {
    panic(err)
  }

  sm.data = tb.data
  sm.width = tb.width
  sm.height = tb.height
}

func (sm *SkinMap) genButtonData(s Skin, tb *TextureBuilder) {
  sm.buttonX, sm.buttonY, sm.buttonT = sm.genBordered(s.Button(), tb, false)

  sm.buttonPressedX, sm.buttonPressedY, _ = sm.genBordered(s.ButtonPressed(), tb, true)
}

func (sm *SkinMap) genInputData(s Skin, tb *TextureBuilder) {
  sm.inputX, sm.inputY, sm.inputT = sm.genBordered(s.Input(), tb, false)
}

func (sm *SkinMap) genFocusData(s Skin, tb *TextureBuilder) {
  sm.focusX, sm.focusY, sm.focusT = sm.genBordered(s.Focus(), tb, false)
}

func (sm *SkinMap) genInsetData(s Skin, tb *TextureBuilder) {
  sm.insetX, sm.insetY, _ = sm.genBordered(s.Inset(), tb, true)
}

func (sm *SkinMap) genCornerData(s Skin, tb *TextureBuilder) {
  sm.cornerX, sm.cornerY, _ = sm.genBordered(s.Corner(), tb, true)
}

func (sm *SkinMap) genBarData(s Skin, tb *TextureBuilder) {
  bar := s.Bar()

  n := calcSquareSkinSize(bar)

  sm.barX, sm.barY = tb.Build(bar, n, n)
  sm.barT = n
}

func (sm *SkinMap) genRadioOffData(s Skin, tb *TextureBuilder) {
  d := s.RadioOff()

  n := calcSquareSkinSize(d)

  sm.radioOffX, sm.radioOffY = tb.Build(d, n, n)
  sm.radioSize = n
}

func (sm *SkinMap) genRadioOnData(s Skin, tb *TextureBuilder) {
  d := s.RadioOn()

  n := calcSquareSkinSize(d)

  sm.radioOnX, sm.radioOnY = tb.Build(d, n, n)

  if sm.radioSize != n {
    panic("inconsistent radio button size")
  }
}

func (sm *SkinMap) genTickData(s Skin, tb *TextureBuilder) {
  d := s.Tick()

  n := calcSquareSkinSize(d)

  sm.tickX, sm.tickY = tb.Build(d, n, n)
  sm.tickSize = n
}

func (sm *SkinMap) genBordered(d []byte, tb *TextureBuilder, checkT bool) (int, int, int) {
  var t int

  if checkT {
    t = calcSkinThicknessCheckRef(d, sm.buttonT)
  } else {
    t = calcSkinThickness(d)
  }

  x, y := tb.BuildBordered(d, t)

  return x, y, t
}

func (s *SkinMap) initGL(uTexLoc uint32, texID uint32, texUnit uint32) {
  s.loc = uTexLoc
  s.tid = texID
  s.tunit = texUnit

  checkGLError()
  gl.ActiveTexture(s.tunit)
  checkGLError()
  gl.BindTexture(gl.TEXTURE_2D, s.tid)
  checkGLError()
  gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
  checkGLError()
  gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
  checkGLError()

  gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(s.width), int32(s.height), 0, gl.RGBA, 
    gl.UNSIGNED_BYTE, unsafe.Pointer(&(s.data[0])))
  checkGLError()

  gl.BindTexture(gl.TEXTURE_2D, 0)
  checkGLError()
}

func (s *SkinMap) bind() {
  checkGLError()
  gl.ActiveTexture(s.tunit)
  gl.BindTexture(gl.TEXTURE_2D, s.tid)
  checkGLError()
}

func (s *SkinMap) ButtonOrigin() (int, int) {
  return s.buttonX, s.buttonY
}

func (s *SkinMap) ButtonPressedOrigin() (int, int) {
  return s.buttonPressedX, s.buttonPressedY
}

func (s *SkinMap) CornerOrigin() (int, int) {
  return s.cornerX, s.cornerY
}

func (s *SkinMap) ButtonBorderThickness() int {
  return s.buttonT
}

func (s *SkinMap) BGColor() sdl.Color {
  return s.skin.BGColor()
}

func (s *SkinMap) SelColor() sdl.Color {
  return s.skin.SelColor()
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

func (s *SkinMap) FocusOrigin() (int, int) {
  return s.focusX, s.focusY
}

func (s *SkinMap) FocusThickness() int {
  return s.focusT
}

func (s *SkinMap) InsetOrigin() (int, int) {
  return s.insetX, s.insetY
}

func (s *SkinMap) BarOrigin() (int, int) {
  return s.barX, s.barY
}

func (s *SkinMap) BarThickness() int {
  return s.barT
}

func (s SkinMap) TickOrigin() (int, int) {
  return s.tickX, s.tickY
}

func (s *SkinMap) TickSize() int {
  return s.tickSize
}

func (s *SkinMap) getButtonCoords() ([4]int, [4]int) {
  t := s.ButtonBorderThickness()

  x0, y0 := s.ButtonOrigin()

  return getBorderedSkinCoords(x0, y0, t)
}

func (s *SkinMap) getCornerCoords() ([4]int, [4]int) {
  t := s.ButtonBorderThickness()

  x0, y0 := s.CornerOrigin()

  return getBorderedSkinCoords(x0, y0, t)
}

func (s *SkinMap) getBarCoords() ([2]int, [4]int) {
  t := s.BarThickness()

  x0, y0 := s.BarOrigin()

  var (
    x [2]int
    y [4]int
  )

  x[0] = x0
  x[1] = x0 + t

  dt := (t - 1)/2
  y[0] = y0 
  y[1] = y0 + dt
  y[2] = y0 + dt + 1
  y[3] = y0 + t

  return x, y
}
