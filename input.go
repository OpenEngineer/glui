package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

type Input struct {
  ElementData

  tris []uint32
  dd   *DrawData
}

func NewInput(dd *DrawData) *Input {
  tris := dd.P1.Alloc(9*2) // so first 18 tris are used for border

  e := &Input{newElementData(), tris, dd}

  e.setTypesAndTCoords()

  return e
}

func (e *Input) Cursor() int {
  return sdl.SYSTEM_CURSOR_IBEAM
}

func (e *Input) setTypesAndTCoords() {
  x0, y0 := e.dd.P1.Skin.InputOrigin()
  t := e.dd.P1.Skin.InputBorderThickness()

  setBorderElementTypesAndTCoords(e.dd, e.tris, x0, y0, t, e.dd.P1.Skin.InputBGColor())
}

func (e *Input) OnResize(maxWidth, maxHeight int) (int, int) {
  width := 200
  height := 50

  t := e.dd.P1.Skin.InputBorderThickness()

  setBorderElementPos(e.dd, e.tris, width, height, t)

  return e.InitBB(width, height)
}

func (e *Input) Translate(dx, dy int) {
  for _, tri := range e.tris {
    e.dd.P1.TranslateTri(tri, dx, dy)
  }

  e.ElementData.Translate(dx, dy)
}
