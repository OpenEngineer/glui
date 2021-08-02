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

//go:generate ./A Body

func (e *Body) RegisterParent(_ Element) {
  panic("can't register body parent")
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

func (e *Body) OnResize(maxWidth, maxHeight int) (int, int) {
  e.bb = Rect{0, 0, maxWidth, maxHeight}

  e.ElementData.resizeChildren(maxWidth, maxHeight)

  return e.InitBB(maxWidth, maxHeight)

  // default block positioning
  /*n := len(e.children)

  if n > 0 {
    h := maxHeight/n

    for i := 0; i < n; i++ {
      child := e.children[i]

      child.OnResize(maxWidth, h)Rect{this.X, this.Y + h*i, this.W, h})
    }
  }

  e.bb = this*/
}
