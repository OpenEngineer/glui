package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

type Body struct {
  ElementData
  iTest   int
  bgColor sdl.Color
}

// windows can't be made transparent like this sadly, so alpha stays 255
func NewBody() *Body {
  return &Body{
    newElementData(),
    0,
    sdl.Color{0,0,0,255},
  }
}

func (e *Body) RegisterParent(_ Element) {
  panic("can't register body parent")
}

func (e *Body) AppendChild(child Element) {
  e.ElementData.appendChild(child)

  child.RegisterParent(e)
}

func (e *Body) BGColor() sdl.Color {
  return e.bgColor
}

// test function
func (e *Body) IncrementBGColor() {
  e.iTest += 1

  c := uint8(e.iTest*10%256)

  e.bgColor = sdl.Color{c, c, c, 255}
}

func (e *Body) OnResize(this Rect) {
  // default block positioning
  n := len(e.children)

  if n > 0 {
    h := this.H/n

    for i := 0; i < n; i++ {
      e.children[i].OnResize(Rect{this.X, this.Y + h*i, this.W, h})
    }
  }

  e.bb = this
}
