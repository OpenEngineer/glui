package main

import (
  "encoding/base64"
  "encoding/json"
  "errors"
  "fmt"
  "io/ioutil"
  "math"
  "os"
  "path/filepath"
  "strconv"
  "strings"

  canvaslib "github.com/tdewolff/canvas"
  fontlib   "github.com/tdewolff/canvas/font"
)

const (
  M = 24
  N = 24
)

// takes a glyphs.json file, and generates valid go source (stdout)
//
// var GlyphNames = ["fontFace:a", "fontFace:b", ...]
//
// var GlyphOffsets = [0, 1.., ..] // offsets of each glyph in the blob
//
// var GlyphBlob = "..." // base64 encoded data

// 127.5 is 0 distance to nearest glyph boundary
// 0 is far away outside the glyph
// 255 is far away inside the glyph

// each glyph is 24px x 24px
// if an edge were to align with a pixel boundary, then at least two pixels on either side give a range of 127.5
// because distance values are defined in pixel centers, the gradient is always 127.5/1.5 => 85/px

// for each column state how many pixels have an unexpected value, followed by that number of (j,d) pairs
// the value of the first pixel of each col is implicitely 0, any unspecified pixels are either 0 or 255 (whichever value is closer to the previously specified pixel)

type Config struct {
  Fonts  []FontsConfig  `json:"fonts"`
  Glyphs []GlyphsConfig `json:"glyphs"`
}

type FontsConfig struct {
  Path string `json:"path"`
  Name string `json:"name"`
}

type GlyphsConfig struct {
  Font  float64 `json:"font"`
  Svg   string  `json:"svg"`
  Index float64 `json:"index"`
  Name  string  `json:"name"`
}

type Glyph struct {
  name string
  data []byte // full dataset
}

func mainInner() error {
  args := os.Args[1:]

  if len(args) != 1 {
    return errors.New("expected 1 arg")
  }

  b, err := ioutil.ReadFile(args[0])
  if err != nil {
    return err
  }

  cfgDir := filepath.Dir(args[0])
  if !filepath.IsAbs(cfgDir) {
    cfgDir, err = filepath.Abs(cfgDir)
    if err != nil {
      return err
    }
  }

  cfg := Config{}
  if err := json.Unmarshal(b, &cfg); err != nil {
    return err
  }

  // read the fonts using canvas fonts
  fonts := make([]*fontlib.SFNT, len(cfg.Fonts))
  for i, fontCfg := range cfg.Fonts {
    if fontCfg.Name == "" {
      return errors.New("font name unset")
    }

    path := fontCfg.Path
    if path == "" {
      return errors.New("font path not set in config file")
    }

    if !filepath.IsAbs(path) {
      path = filepath.Join(cfgDir, path)
    }

    b, err := ioutil.ReadFile(path)
    if err != nil {
      return err
    }

    fonts[i], err = fontlib.ParseFont(b, 0)
    if err != nil {
      return err
    }
  }

  uniqueNames := make(map[string]bool)

  glyphs := make([]*Glyph, len(cfg.Glyphs))
  for i, glyphCfg := range cfg.Glyphs {
    if glyphCfg.Name == "" {
      return errors.New("glyph name unset")
    }

    if glyphCfg.Svg == "" {
      idx := uint16(glyphCfg.Index)

      f := int(glyphCfg.Font)

      if f < 0 || f >= len(fonts) {
        return errors.New("bad font index")
      }

      name := cfg.Fonts[f].Name + ":" + glyphCfg.Name
      if _, ok := uniqueNames[name]; ok {
        return errors.New("glyph name " + name + " already used")
      }

      uniqueNames[name] = true

      glyphs[i], err = createFontGlyph(fonts[f], idx, name)
      if err != nil {
        return err
      }
    } else {
      if glyphCfg.Index != 0.0 {
        return errors.New("can't have both svg and index")
      }

      if glyphCfg.Font != 0.0 {
        return errors.New("can't have both svg and font")
      }

      name := glyphCfg.Name
      if _, ok := uniqueNames[name]; ok {
        return errors.New("glyph name " + name + " already used")
      }

      uniqueNames[name] = true

      g := glyphCfg.Svg

      glyphs[i], err = createSvgGlyph(g, name)
      if err != nil {
        return err
      }
    }
  }

  return glyphsToSource(glyphs)
}

func main() {
  if err := mainInner(); err != nil {
    fmt.Fprintf(os.Stderr, "%s\n", err.Error())
  }
}

func glyphsToSource(glyphs []*Glyph) error {
  final := make([]byte, 0)
  offsets := make([]int, len(glyphs))

  var b strings.Builder
  b.WriteString("var GlyphNames = [\n")

  // first the names
  for _, g := range glyphs {
    b.WriteString("\"")
    b.WriteString(g.name)
    b.WriteString("\",\n")
  }
  b.WriteString("]\n\n")

  for i, g := range glyphs {
    //g.PrintData()
    //os.Exit(1)
    data := g.data
    //data, err := g.Compress() // compression doesnt actually save any data
    //if err != nil {
      //return err
    //}

    offsets[i] = len(final)
    final = append(final, data...)
  }

  // write the offsets
  b.WriteString("var GlyphOffsets = [\n")
  for _, o := range offsets {
    b.WriteString(strconv.Itoa(o))
    b.WriteString(",\n")
  }
  b.WriteString("]\n\n")

  // write the data
  b.WriteString("var GlyphData = \"")
  b.WriteString(base64.StdEncoding.EncodeToString(final))
  b.WriteString("\"\n")

  fmt.Println(b.String())

  return nil
}

func (g *Glyph) PrintData() {
  for i := 0; i < M; i++ {
    for j := 0; j < N; j++ {
      k := i*N + j

      x := float64(i) + 0.5
      y := float64(j) + 0.5

      fmt.Println(x, y, g.data[k])
    }

    fmt.Println()
  }
}

func (g *Glyph) Compress() ([]byte, error) {
  res := make([]byte, 0)

  for i := 0; i < M; i++ {
    started := false

    col := make([]byte, 0)

    for j := 0; j < N; j++ {
      k := i*N + j

      d := g.data[k]

      if j == 0 && d >= 128 {

        return nil, errors.New("unexpected glyph data")
      }

      if !started {
        if d != 0 {
          col = append(col, byte(j), d)
          started = true
        }
      } else {
        dPrev := g.data[k-1]

        if d != 0 && d != 255 {
          col = append(col, byte(j), d)
        } else if dPrev != d {
          if d == 255 && dPrev <= 127 {
            col = append(col, byte(j), d)
          } else if d == 0 && dPrev >= 128 {
            col = append(col, byte(j), d)
          }
        }
      }

      if j == N - 1 && d >= 128 {
        return nil, errors.New("unexpected glyph data")
      }
    }

    res = append(res, byte(len(col)/2))
    res = append(res, col...)
  }

  return res, nil
}

func createFontGlyph(font *fontlib.SFNT, gIdx uint16, name string) (*Glyph, error) {
  p := &canvaslib.Path{}

  if err := font.GlyphPath(p, gIdx, uint16(N), 0, 0, 1.0, fontlib.NoHinting); err != nil {
    return nil, err
  }

  return createPathGlyph(p, name)
}

func createSvgGlyph(g string, name string) (*Glyph, error) {
  p, err := canvaslib.ParseSVG(g)
  if err != nil {
    return nil, err
  }

  return createPathGlyph(p, name)
}

func distanceToFiniteLine(xa, ya, xb, yb, x, y float64) float64 {
  dx := x - xa
  dy := y - ya

  tx := xb - xa
  ty := yb - ya
  tl := math.Sqrt(tx*tx + ty*ty)
  tx = tx/tl
  ty = ty/tl

  p := dx*tx + dy*ty

  if p <= 0.0 {
    // distance to a
    return math.Sqrt(dx*dx + dy*dy)
  } else if p >= tl {
    // distance to b
    dx = x - xb
    dy = y - yb

    return math.Sqrt(dx*dx + dy*dy)
  } else {
    // normal distance to line

    px := xa + tx*p
    py := ya + ty*p

    dx = x - px
    dy = y - py

    return math.Sqrt(dx*dx + dy*dy)
  }
}

func distanceToPath(segments []canvaslib.Segment, x float64, y float64) float64 {
  // assume path is closed
  dClosest := float64(M)

  for _, s := range segments {
    if s.Cmd == 2.0 {
      a := s.Start
      b := s.End

      d := distanceToFiniteLine(a.X, a.Y, b.X, b.Y, x, y)
      if d < dClosest {
        dClosest = d
      }
    }
  }

  return dClosest
}

func printSegments(segments []canvaslib.Segment) {
  for _, s := range segments {
    if s.Cmd == 2.0 {
      fmt.Println(s.Start.X, s.Start.Y, s.Cmd)
      fmt.Println(s.End.X, s.End.Y, s.Cmd)
      fmt.Println()
    }
  }

  os.Exit(1)
}

func createPathGlyph(p *canvaslib.Path, name string) (*Glyph, error) {
  canvaslib.Tolerance = 0.01

  pFlat := p.Flatten()

  segments := pFlat.Segments()
  data := make([]byte, M*N)

  //printSegments(segments)

  for i := 0; i < M; i++ {
    for j := 0; j < N; j++ {
      k := i*N + j
      x := float64(i) + 0.5
      y := float64(j) + 0.5

      d := distanceToPath(segments, x, y)

      dByte := d*85.0

      if p.Interior(x, y, canvaslib.NonZero) {
        b := dByte + 127.5
        if b > 255.0 {
          b = 255.0
        }
        data[k] = byte(b)
      } else {
        b := 127.5 - dByte
        if b < 0.0 {
          b = 0.0
        }
        data[k] = byte(b)
      }
    }
  }

  return &Glyph{name, data}, nil
}
