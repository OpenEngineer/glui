package glui

import (
  "fmt"
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

  dirty  bool
}

func NewTextureBuilder(nComp int, initWidth int, initHeight int) *TextureBuilder {
  return &TextureBuilder{
    nComp,
    initWidth,
    initHeight,
    []Rect{Rect{0, 0, initWidth, initHeight}},
    make([]byte, nComp*initWidth*initHeight),
    true,
  }
}

func (tb *TextureBuilder) Build(data []byte, w int, h int) (int, int) {
  // look for a free that is able to accomodate
  horFailCount := 0
  verFailCount := 0

  // free rect with smallest waste wins
  bestFree := -1
  bestWaste := w*h
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
      waste := (f.W - w)*h + f.W*(f.H - h)
      if bestFree == -1 || waste < bestWaste {
        bestFree = i
        bestWaste = waste
      }
    }
  }

  if bestFree != -1 {
    i := bestFree
    f := tb.free[i]

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
  } else {
    if horFailCount > verFailCount {
      fmt.Println("failed to find free rect, growing right", tb.free)
      tb.growRight()
    } else {
      fmt.Println("failed to find free rect, growing down", tb.free)
      tb.growDown()
    }

    tb.dirty = true

    return tb.Build(data, w, h)
  }
}

func (tb *TextureBuilder) BuildBordered(data []byte, t int) (int, int) {
  s := 2*t+1

  return tb.Build(data, s, s)
}

func (tb *TextureBuilder) Free(x, y, w, h int) {
  tb.free = append(tb.free, Rect{x, y, w, h})
}

// TODO
func (tb *TextureBuilder) Defrag() {
}

func (tb *TextureBuilder) setData(x int, y int, data []byte, w int, h int) {
  for i := x; i < x+w; i++ {
    dst0 := (i*tb.height + y)*tb.nComp
    dst1 := (i*tb.height + y+h)*tb.nComp

    src0 := ((i-x)*h + (y - y))*tb.nComp
    src1 := ((i-x)*h + (y + h - y))*tb.nComp

    copy(tb.data[dst0 : dst1 : dst1], data[src0 : src1 : src1])
  }

  tb.dirty = true
}

func (tb *TextureBuilder) growRight() {
  oldWidth := tb.width
  tb.width = oldWidth*2

  oldData := tb.data
  tb.data = make([]byte, tb.nComp*tb.width*tb.height)

  tb.setData(0, 0, oldData, oldWidth, tb.height)

  tb.free = append(tb.free, Rect{oldWidth, 0, oldWidth, tb.height})
  tb.dirty = true
}

func (tb *TextureBuilder) growDown() {
  oldHeight := tb.height
  tb.height = oldHeight*2

  oldData := tb.data
  tb.data = make([]byte, tb.nComp*tb.width*tb.height)

  tb.setData(0, 0, oldData, tb.width, oldHeight)

  tb.free = append(tb.free, Rect{0, oldHeight, tb.width, oldHeight})
  tb.dirty = true
}

func (tb *TextureBuilder) ToImage(fname string) error {
  return DataToImage(tb.data, tb.width, tb.height, fname)
}

func DataToImage(data []byte, width, height int, fname string) error {
  f, err := os.Create(fname)
  if err != nil {
    return err
  }

  r := image.Rect(0, 0, width, height)

  img := image.NewRGBA(r)

  for i := 0; i < width; i++ {
    for j := 0; j < height; j++ {
      src := i*height + j

      c := color.RGBA{
        data[src*4+0], 
        data[src*4+1], 
        data[src*4+2], 
        data[src*4+3],
      }

      img.Set(i, j, c)
    }
  }

  return png.Encode(f, img)
}
