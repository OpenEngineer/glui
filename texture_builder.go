package glui

import (
  "image"
  "image/color"
  "image/png"
  "os"
)

type TextureBuilder struct {
  nComp  int
  width  int
  height int

  free   []Rect

  data   []byte
}

func NewTextureBuilder(nComp int, initWidth int, initHeight int) *TextureBuilder {
  return &TextureBuilder{
    nComp,
    initWidth,
    initHeight,
    []Rect{Rect{0, 0, initWidth, initHeight}},
    make([]byte, nComp*initWidth*initHeight),
  }
}

func (tb *TextureBuilder) Build(data []byte, w int, h int) (int, int) {
  // look for a free that is able to accomodate
  horFailCount := 0
  verFailCount := 0

  for i, f := range tb.free {
    horFail := f.W < w
    verFail := f.H < h
    if horFail {
      horFailCount++
    }
    if verFail {
      verFailCount++
    }

    if !horFail && !verFail {
      tb.setData(f.X, f.Y, data, w, h)

      if f.W == w && f.H == h {
        if len(tb.free) == 1 {
          tb.free = []Rect{}
        } else {
          tb.free = append(tb.free[0:i], tb.free[i+1:]...)
        }
      } else if (f.W - w) < (f.H - h) {
        tb.free[i] = Rect{f.X, f.Y + h, w, f.H - h}
        tb.free = append(tb.free, Rect{f.X + w, f.Y, f.W - w, f.H})
      } else {
        tb.free[i] = Rect{f.X + w, f.Y, f.W - w, h}
        tb.free = append(tb.free, Rect{f.X, f.Y + h, f.W, f.H - h})
      }

      return f.X, f.Y
    }
  }

  if horFailCount > verFailCount {
    tb.growRight()
  } else {
    tb.growDown()
  }

  return tb.Build(data, w, h)
}

func (tb *TextureBuilder) setData(x int, y int, data []byte, w int, h int) {
  for i := x; i < x+w; i++ {
    for j := y; j < y+h; j++ {
      for c := 0; c < tb.nComp; c++ {
        dst := (i*tb.height + j)*tb.nComp + c
        src := ((i-x)*h + (j - y))*tb.nComp + c
        tb.data[dst] = data[src]
      }
    }
  }
}

func (tb *TextureBuilder) growRight() {
  oldWidth := tb.width
  tb.width = oldWidth*2

  oldData := tb.data
  tb.data = make([]byte, tb.nComp*tb.width*tb.height)

  tb.setData(0, 0, oldData, oldWidth, tb.height)

  tb.free = append(tb.free, Rect{oldWidth, 0, oldWidth, tb.height})
}

func (tb *TextureBuilder) growDown() {
  oldHeight := tb.height
  tb.height = oldHeight*2

  oldData := tb.data
  tb.data = make([]byte, tb.nComp*tb.width*tb.height)

  tb.setData(0, 0, oldData, tb.width, oldHeight)

  tb.free = append(tb.free, Rect{0, oldHeight, tb.width, oldHeight})
}

func (tb *TextureBuilder) ToImage(fname string) error {
  f, err := os.Create(fname)
  if err != nil {
    return err
  }

  r := image.Rect(0, 0, tb.width, tb.height)

  img := image.NewRGBA(r)

  for i := 0; i < tb.width; i++ {
    for j := 0; j < tb.height; j++ {
      src := i*tb.height + j

      c := color.RGBA{
        tb.data[src*4+0], 
        tb.data[src*4+1], 
        tb.data[src*4+2], 
        tb.data[src*4+3],
      }

      img.Set(i, j, c)
    }
  }

  return png.Encode(f, img)
}
