package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

//go:generate ./gen_element Caption "CalcDepth appendChild"

// text with shadow
type Caption struct {
  ElementData

  main *Text
  back  *Text

  mainColor sdl.Color
  backColor sdl.Color
}

func NewSansCaption(root *Root, content string, size float64) *Caption {
  return NewCaption(root, content, DEFAULT_SANS, size)
}

func NewCaption(root *Root, content string, font string, size float64) *Caption {
  e := &Caption{
    NewElementData(root, 0, 0),
    NewText(root, content, font, size),
    NewText(root, content, font, size),
    sdl.Color{0x00, 0x00, 0x00, 0xff},
    sdl.Color{0xff, 0xff, 0xff, 0xff},
  }

  e.main.SetColor(e.mainColor)
  e.back.SetColor(e.backColor)

  e.back.Hide()

  e.appendChild(e.back)
  e.appendChild(e.main)

  e.main.closerThan = append(e.main.closerThan, e.back)

  return e
}

func (e *Caption) Disable() {
  e.main.SetColor(sdl.Color{0x70, 0x70, 0x70, 0xff})
  e.back.Show()

  e.ElementData.Disable()
}

func (e *Caption) Enable() {
  e.main.SetColor(e.mainColor)
  e.back.Hide()

  e.ElementData.Enable()
}

func (e *Caption) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  w, h := e.main.CalcPos(maxWidth, maxHeight, maxZIndex)

  if !e.enabled {
    e.back.CalcPos(maxWidth, maxHeight, maxZIndex)

    e.back.Translate(1, 1)
  }

  return e.InitRect(w, h)
}
