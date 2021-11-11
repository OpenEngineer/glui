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

func (tb *TextureBuilder) HaveFreeSpace(w, h int) bool {
  b := false

  for _, f := range tb.free {
    if f.W >= w && f.H >= h {
      b = true
      break
    }
  }

  return b
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

    tb.filterEmptyFree()

    return f.X, f.Y
  } else {
    if horFailCount > verFailCount {
      tb.growRight()
    } else {
      tb.growDown()
    }

    tb.dirty = true

    return tb.Build(data, w, h)
  }
}

func (tb *TextureBuilder) filterEmptyFree() {
  free := make([]Rect, 0)

  for _, fr := range tb.free {
    if fr.W > 0 && fr.H > 0 {
      free = append(free, fr)
    }
  }

  tb.free = free
}

func (tb *TextureBuilder) BuildBordered(data []byte, t int) (int, int) {
  s := 2*t+1

  return tb.Build(data, s, s)
}

func (tb *TextureBuilder) Free(x, y, w, h int) {
  if w > 0 && h > 0 {
    tb.free = append(tb.free, Rect{x, y, w, h})
  }

  tb.Defrag()
}

func (tb *TextureBuilder) dumpFree(fname string) {
  f, err := os.Create(fname)
  if err != nil {
    panic("unable to debug defrag")
  }

  for _, fr := range tb.free {
    fmt.Fprintf(f, "%d %d\n", fr.X, fr.Y)
    fmt.Fprintf(f, "%d %d\n", fr.Right(), fr.Y)
    fmt.Fprintf(f, "%d %d\n", fr.Right(), fr.Bottom())
    fmt.Fprintf(f, "%d %d\n", fr.X, fr.Bottom())
    fmt.Fprintf(f, "%d %d\n\n", fr.X, fr.Y)
  }
}

func (tb *TextureBuilder) Defrag() {
  tb.free = defragFreeRects(tb.free)
}

func aIsBetterThanB(a []Rect, b []Rect) bool {
  a = sortRects(a[:])
  b = sortRects(b[:])

  for i := 0; i < len(a); i++ {
    if i >= len(b) {
      return false
    }

    rA := a[i]
    rB := b[i]

    rAWH := rA.W*rA.H
    rBWH := rB.W*rB.H
    if rAWH > rBWH {
      return true
    } else if rBWH > rAWH {
      return false
    }
  }

  return true
}

func defragFreeRects(free []Rect) []Rect {
  // no further improvements possible
  if len(free) < 1 {
    return free
  }

  removeTwo := func(lst []Rect, i, j int) []Rect {
    if i > j {
      i, j = j, i
    }

    newLst := make([]Rect, len(lst) - 2)

    copy(newLst[0:i], lst[0:i])
    copy(newLst[i:j-1], lst[i+1:j])
    if j < len(lst) - 1 {
      copy(newLst[j-1:], lst[j+1:])
    }

    return newLst
  }

  // first simple search for common edges
  for i, r := range free {
    for j, r_ := range free {
      if i == j {
        continue
      }

      rNew, ok := r.MergeAlongEdge(r_)
      if ok {
        if r.Area() + r_.Area() != rNew.Area() {
          fmt.Println(rNew, r_, r)
          panic("area not preserved")
        }

        free = append(removeTwo(free, i, j), rNew)

        // this is a guaranteed improvement, and we return now
        return defragFreeRects(free)
      }
    }
  }

  // now search for partially overlapping edges
  improved := false
  freeTries := make([][]Rect, 0) // different possibilities
  for i, r := range free {
    for j, r_ := range free {
      if i == j {
        continue
      }

      rNew, rRem, ok := r.MergeAlongPartialEdge(r_)
      if ok {
        freeTries = append(freeTries, append(removeTwo(free, i, j), rNew, rRem))
      }
    }
  }

  for _, try := range freeTries {
    if aIsBetterThanB(try, free) {
      free = try

      improved = true
    }
  }

  if improved {
    return defragFreeRects(free)
  } else { 
    return free
  }
}

func (tb *TextureBuilder) setData(x int, y int, data []byte, w int, h int) {
  // newer method with larger contiguous copies
  for i := x; i < x+w; i++ {
    dst0 := (i*tb.height + y)*tb.nComp
    dst1 := (i*tb.height + y+h)*tb.nComp

    src0 := ((i-x)*h + (y - y))*tb.nComp
    src1 := ((i-x)*h + (y + h - y))*tb.nComp

    copy(tb.data[dst0 : dst1 : dst1], data[src0 : src1 : src1])
  }

  //old method for reference
  // for i := x; i < x+w; i++ {
  //   for j := y; j < y+h; j++ {
  //     for c := 0; c < tb.nComp; c++ {
  //       dst := (i*tb.height + j)*tb.nComp + c
  //       src := ((i-x)*h + (j - y))*tb.nComp + c
  //       tb.data[dst] = data[src]
  //     }
  //   }
  // }

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

  tb.Defrag()
}

func (tb *TextureBuilder) growDown() {
  oldHeight := tb.height
  tb.height = oldHeight*2

  oldData := tb.data
  tb.data = make([]byte, tb.nComp*tb.width*tb.height)

  tb.setData(0, 0, oldData, tb.width, oldHeight)

  tb.free = append(tb.free, Rect{0, oldHeight, tb.width, oldHeight})
  tb.dirty = true

  tb.Defrag()
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
