package main

import (
  "errors"
  "fmt"
  "image"
  "image/color"
  "image/png"
  "math"
  "os"
  "strconv"

  "github.com/computeportal/glui"
)

// creates a png image containing the resulting render
func mainInner() error {
  args := os.Args[1:]

  if len(args) != 3 {
    return errors.New("expected glyph_maker glyphName xSize ySize")
  }

  name := args[0]

  glyphs := MakeGlyphs()

  g, ok := glyphs[name]
  if !ok {
    return errors.New("glyph \"" + name + "\" not found")
  }

  xSizeStr := args[1]
  xSize, err := strconv.Atoi(xSizeStr)
  if err != nil {
    return err
  }

  ySizeStr := args[2]
  ySize, err := strconv.Atoi(ySizeStr)
  if err != nil {
    return err
  }

  rect := image.Rect(0, 0, xSize, ySize)
  img := image.NewRGBA(rect)

  minSize := xSize 
  x0 := 0.0
  y0 := 0.0
  if ySize < xSize {
    minSize = ySize
    x0 = float64(xSize - ySize)/2.0
    y0 = 0.0
  } else {
    x0 = 0.0
    y0 = float64(ySize - xSize)/2.0
  }

  scale := float64(minSize)/float64(glui.GlyphResolution)

  for i := 0; i < xSize; i++ {
    for j := 0; j < ySize; j++ {
      j_ := ySize - 1 - j
      xAbs := float64(i) + 0.5
      yAbs := float64(j) + 0.5

      xRel := (xAbs - x0)/scale
      yRel := (yAbs - y0)/scale

      if xRel < 0.0 || xRel > float64(glui.GlyphResolution) || yRel < 0.0 || yRel > float64(glui.GlyphResolution) {
        img.Set(i, j_, color.RGBA{0x00, 0x00, 0x00, 0x00})
        fmt.Println(xAbs, yAbs, 0.0)
      } else {
        d := (interpBilinear(g.Distances, xRel, yRel) - 127.5)*scale

        a := interpBilinear(g.Angles, xRel, yRel)/255.0*0.25*math.Pi

        if d < 0.0 {
          dUnit := -d/float64(glui.GlyphDPerPx)

          if dUnit < 0.5*math.Sqrt(2.0) {
            // partially outside
            A := calcPixelCoverage(dUnit, a)
            if A > 0.5 {
              panic("unexpected")
            }

            alpha := pixelCoverageToAlpha(A)
            img.Set(i, j_, color.RGBA{0x00, 0x00, 0x00, alpha})
          } else {
            // outside
            img.Set(i, j_, color.RGBA{0x00, 0x00, 0x00, 0x00})
          }
        } else {
          dUnit := d/float64(glui.GlyphDPerPx)

          if dUnit < 0.5*math.Sqrt(2.0) {
            A := 1.0 - calcPixelCoverage(dUnit, a)
            if A < 0.5 {
              panic("unexpected")
            }

            alpha := pixelCoverageToAlpha(A)
            img.Set(i, j_, color.RGBA{0x00, 0x00, 0x00, alpha})
          } else {
            // inside
            img.Set(i, j_, color.RGBA{0x00, 0x00, 0x00, 0xff})
          }
        }
      }
    }
  }

  imgF, err := os.Create(name + ".png")
  if err != nil {
    return err
  }

  return png.Encode(imgF, img)
}

func calcPixelCoverage(d, a float64) float64 {
  h1 := 0.5 + math.Tan(a)*0.5 - d/math.Cos(a)

  if math.IsNaN(h1) {
    panic("h1 is nan")
  }

  if h1 < 1e-5 {
    return 0.0
  } else if a >= math.Atan(h1) {
    w := h1/math.Tan(a)
    if math.IsNaN(w) {
      panic("w is nan")
    }
    return 0.5*w*h1
  } else {
    h2 := 0.5 - math.Tan(a)*0.5 - d/math.Cos(a)
    if h2 < 0.0 || h2 > h1{
      fmt.Println(d, a, h1, h2, math.Atan(h1))
      panic("unexpected")
    }

    if math.IsNaN(h2) {
      panic("h2 is nan")
    }

    return 0.5*(h1 + h2)
  }
}

func pixelCoverageToAlpha(A float64) byte {
  if A < 0.0 || A > 1.0 {
    fmt.Println(A)
    panic("unexpected A")
  } else {
    start := 0.1
    end :=   0.9

    if A < start {
      return 0x00
    } else if A > end {
      return 0xff
    } else {
      f := (A - start)/(end - start)

      if f <= 0.0 { 
        return 0x00
      } else if f > 1.0 {
        return 0xff
      } else {
        return byte(f*255)
      }
    }

    //res := math.Max(math.Min(A*255, 254.0), 1.0)

    /*fmt.Println(A)
    if A > 0.5 {//&& res < 255.0 {
      //res = 255.0
      return 0xff
    } else {
      return 0x00
    }

    //return byte(res)

    if A < 0.5 {
      return 0x00
    } else {
      return 0xff
    }*/
  }
}
func interpBilinear(data []byte, xRel float64, yRel float64) float64 {
  il := int(math.Floor(xRel))
  ir := int(math.Ceil(xRel))

  if ir >= glui.GlyphResolution || il < 0 {
    return 0.0
  }

  jt := int(math.Floor(yRel))
  jb := int(math.Ceil(yRel))

  if jb >= glui.GlyphResolution || jt < 0 {
    return 0.0
  }

  fx := xRel - float64(il)
  fy := yRel - float64(jt)

  dtl := float64(data[il*glui.GlyphResolution + jt])
  dtr := float64(data[ir*glui.GlyphResolution + jt])
  dbr := float64(data[ir*glui.GlyphResolution + jb])
  dbl := float64(data[il*glui.GlyphResolution + jb])

  db := fx*dbr + (1.0 - fx)*dbl
  dt := fx*dtr + (1.0 - fx)*dtl

  d := fy*db + (1.0 - fy)*dt

  return d
}

func main() {
  if err := mainInner(); err != nil {
    fmt.Fprintf(os.Stderr, "%s\n", err.Error())
  }
}
