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

  "github.com/urfave/cli"
  canvaslib "github.com/tdewolff/canvas"
  fontlib   "github.com/tdewolff/canvas/font"
)

const (
  RESOLUTION = 24
  PADDING    = 2
  D_PER_PX   = 85.0
)

var (
  PACKAGE_NAME = "main"
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

type Kerning struct {
  next    rune
  advance float64
}

type Glyph struct {
  name string
  data []byte
  angle []byte
  hints []float64 // top, right, bottom, left
  scale float64
  advance float64
  originX float64
  originY float64
  kernings []Kerning
}


func swapByte(d []byte, a int, b int) {
  x := d[a] 

  d[a] = d[b]

  d[b] = x
}

func flipY(d []byte, w int, h int) {
  for i := 0; i < w; i++ {
    for j := 0; j < h/2; j++ {
      a := i*h + j
      b := i*h + (h - 1 - j)

      swapByte(d, a, b)
    }
  }
}

func mainInner(c *cli.Context) error {
  args := c.Args()

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
    if glyphCfg.Svg == "" {
      if glyphCfg.Name != "" {
        return errors.New("unexpected glyph name")
      }

      idx := rune(glyphCfg.Index)

      f := int(glyphCfg.Font)

      if f < 0 || f >= len(fonts) {
        return errors.New("bad font index")
      }

      name := fmt.Sprintf("%s:%d", cfg.Fonts[f].Name, idx)
      if _, ok := uniqueNames[name]; ok {
        return errors.New("glyph name " + name + " already used")
      }

      uniqueNames[name] = true

      glyphs[i], err = createFontGlyph(fonts[f], idx, name)
      if err != nil {
        return err
      }
    } else {
      if glyphCfg.Name == "" {
        return errors.New("glyph name unset")
      }

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

  return glyphsToSource(glyphs, PACKAGE_NAME)
}

func main() {
  app := cli.NewApp()
  app.Name = "glyph_maker"
  app.Usage = "Generate a go source files containing the glyphs specified in a json config"
  app.Version = "0.1.0"
  app.Action = mainInner
  app.Flags = []cli.Flag{
    cli.StringFlag{
      Name: "p,package",
      Destination: &PACKAGE_NAME,
      Value: "main",
    },
  }

  if err := app.Run(os.Args); err != nil {
    fmt.Fprintf(os.Stderr, "%s\n", err.Error())
  }
}

func glyphsToSource(glyphs []*Glyph, packageName string) error {
  final := make([]byte, 0)
  offsets := make([]int, len(glyphs))

  var b strings.Builder
  b.WriteString("package ")
  b.WriteString(packageName)
  b.WriteString("\n\n")

  b.WriteString("import (\n")
  b.WriteString("  \"encoding/base64\"\n\n")
  b.WriteString("  \"github.com/computeportal/glui\"\n\n")
  b.WriteString(")\n")

  b.WriteString("func MakeGlyphs() map[string]*glui.Glyph {\n")

  b.WriteString("var glyphNames_ = []string{\n")

  // first the names
  for _, g := range glyphs {
    b.WriteString("\"")
    b.WriteString(g.name)
    b.WriteString("\",\n")
  }
  b.WriteString("}\n\n")

  // write the hint box
  b.WriteString("var glyphHints_ = []float64{\n")
  for _, g := range glyphs {
    for _, h := range g.hints {
      b.WriteString(fmt.Sprintf("%.03f,\n", h))
    }
  }
  b.WriteString("}\n\n")

  // write the scales
  b.WriteString("var glyphScales_ = []float64{\n")
  for _, g := range glyphs {
    b.WriteString(fmt.Sprintf("%g", g.scale))
    b.WriteString(",\n")
  }
  b.WriteString("}\n\n")

  // write the advance
  b.WriteString("var glyphAdvances_ = []float64{\n")
  for _, g := range glyphs {
    b.WriteString(fmt.Sprintf("%g", g.advance))
    b.WriteString(",\n")
  }
  b.WriteString("}\n\n")

  // write the origins
  b.WriteString("var glyphOrigins_ = []float64{\n")
  for _, g := range glyphs {
    b.WriteString(fmt.Sprintf("%.03f,%.03f,\n", g.originX, g.originY))
  }
  b.WriteString("}\n\n")

  kOffsets := make([]int, len(glyphs) + 1)
  kOffsets[0] = 0
  b.WriteString("var glyphKernings_ = []glui.GlyphKerning{\n")
  for i, g := range glyphs {
    nKernings := len(g.kernings)
    for _, k := range g.kernings {
      b.WriteString("glui.GlyphKerning{")
      b.WriteString(strconv.Itoa(int(k.next)))
      b.WriteString(",")
      b.WriteString(fmt.Sprintf("%g", k.advance))
      b.WriteString("},\n")
    }

    kOffsets[i+1] = kOffsets[i] + nKernings
  }
  b.WriteString("}\n\n")

  // write kerning offsets
  b.WriteString("var glyphKerningOffsets_ = []int{\n")
  for i := 0; i < len(kOffsets); i++ {
    b.WriteString(strconv.Itoa(kOffsets[i]))
    b.WriteString(",\n")
  }
  b.WriteString("}\n\n")

  for i, g := range glyphs {
    //g.PrintData()
    //os.Exit(1)
    data := append(g.data[:], g.angle[:]...)
    //data, err := g.Compress() // compression doesnt actually save any data
    //if err != nil {
      //return err
    //}

    offsets[i] = len(final)
    final = append(final, data...)
  }


  // write the offsets
  b.WriteString("var glyphDataOffsets_ = []int{\n")
  for _, o := range offsets {
    b.WriteString(strconv.Itoa(o))
    b.WriteString(",\n")
  }
  b.WriteString("}\n\n")


  // write the data (distance+angles)
  b.WriteString("var glyphData_ = \"")
  b.WriteString(base64.StdEncoding.EncodeToString(final))
  b.WriteString("\"\n\n")

  b.WriteString("  b, err := base64.StdEncoding.DecodeString(glyphData_)\n")
  b.WriteString("  if err != nil {\n")
  b.WriteString("     panic(err)\n")
  b.WriteString("  }\n")
  b.WriteString("  m := make(map[string]*glui.Glyph)\n")
  b.WriteString("  offsetk1 := 0\n")
  b.WriteString("  for i, name := range glyphNames_ {\n")
  b.WriteString("    offsetd1 := glyphDataOffsets_[i]\n")
  b.WriteString("    offsetd2 := offsetd1 + glui.GlyphResolution*glui.GlyphResolution\n")
  b.WriteString("    data := b[offsetd1:offsetd2]\n")
  b.WriteString("    angle := b[offsetd2:offsetd2 + glui.GlyphResolution*glui.GlyphResolution]\n")
  b.WriteString("    hints := glyphHints_[i*4:(i+1)*4]\n")
  b.WriteString("    offsetk2 := glyphKerningOffsets_[i+1]\n")
  b.WriteString("    kernings := glyphKernings_[offsetk1:offsetk2]\n")
  b.WriteString("    offsetk1 = offsetk2\n")
  b.WriteString("    m[name] = &glui.Glyph{data, angle, hints, glyphScales_[i], glyphAdvances_[i], glyphOrigins_[i*2], glyphOrigins_[i*2+1], kernings, 0}\n")
  b.WriteString("  }\n\n")
  b.WriteString("  return m\n")
  b.WriteString("}\n\n")

  fmt.Println(b.String())

  return nil
}

func (g *Glyph) PrintData() {
  for i := 0; i < RESOLUTION; i++ {
    for j := 0; j < RESOLUTION; j++ {
      k := i*RESOLUTION + j

      x := float64(i) + 0.5
      y := float64(j) + 0.5

      fmt.Println(x, y, g.data[k])
    }

    fmt.Println()
  }
}

func (g *Glyph) Compress() ([]byte, error) {
  res := make([]byte, 0)

  for i := 0; i < RESOLUTION; i++ {
    started := false

    col := make([]byte, 0)

    for j := 0; j < RESOLUTION; j++ {
      k := i*RESOLUTION + j

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

      if j == RESOLUTION - 1 && d >= 128 {
        return nil, errors.New("unexpected glyph data")
      }
    }

    res = append(res, byte(len(col)/2))
    res = append(res, col...)
  }

  return res, nil
}

func createFontGlyph(font *fontlib.SFNT, rIdx rune, name string) (*Glyph, error) {
  gIdx := font.GlyphIndex(rIdx)

  /*if strings.HasSuffix(name, "G") {
    dumpKerning(font, rIdx)
  }*/

  pTest := &canvaslib.Path{}
  if err := font.GlyphPath(pTest, gIdx, 0, 0, 0, 1.0, fontlib.NoHinting); err != nil {
    return nil, err
  }

  r := pTest.Bounds()

  pFinal := &canvaslib.Path{}

  var (
    scale float64
    x0 float64
    y0 float64
  )
  if r.W < r.H { // narrow rect, horizontally center
    scale = float64(RESOLUTION - 2*PADDING)/r.H
    y0 = float64(PADDING)
    x0 = (float64(RESOLUTION) - scale*r.W)*0.5
  } else { // wide tect, vertically center
    scale = float64(RESOLUTION - 2*PADDING)/r.W
    y0 = (float64(RESOLUTION) - scale*r.H)*0.5
    x0 = float64(PADDING)
  }

  if (strings.HasSuffix(name, "q")) {
    fmt.Fprintf(os.Stderr, "%f %f\n", r.X, r.Y)
  }

  originX := x0/scale - r.X
  originY := y0/scale - r.Y
  if err := font.GlyphPath(pFinal, gIdx, 0, int32(originX), int32(originY), scale, fontlib.NoHinting); err != nil {
    return nil, err
  }

  advance := font.GlyphAdvance(gIdx)

  glyph, err := createPathGlyph(pFinal, name, scale, scale*float64(advance), originX*scale, float64(RESOLUTION) - originY*scale, listKernings(font, gIdx, scale))
  if err != nil {
    return nil, err
  }

  flipY(glyph.data,   RESOLUTION, RESOLUTION)
  flipY(glyph.angle, RESOLUTION, RESOLUTION)

  return glyph, nil
}

func createSvgGlyph(g string, name string) (*Glyph, error) {
  p, err := canvaslib.ParseSVG(g)
  if err != nil {
    return nil, err
  }

  return createPathGlyph(p, name, 1.0, float64(RESOLUTION), 0.0, float64(RESOLUTION), make([]Kerning, 0))
}

func distanceAndAngleToFiniteLine(xa, ya, xb, yb, x, y float64) (float64, float64) {
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
    return math.Sqrt(dx*dx + dy*dy), math.Atan2(dy, dx)
  } else if p >= tl {
    // distance to b
    dx = x - xb
    dy = y - yb

    return math.Sqrt(dx*dx + dy*dy), math.Atan2(dy, dx)
  } else {
    // normal distance to line

    px := xa + tx*p
    py := ya + ty*p

    dx = x - px
    dy = y - py

    return math.Sqrt(dx*dx + dy*dy), math.Atan2(dy, dx)
  }
}

func distanceAndAngleToPath(segments []canvaslib.Segment, x float64, y float64) (float64, float64) {
  // assume path is closed
  dClosest := float64(RESOLUTION)
  aClosest := 0.0

  for _, s := range segments {
    if s.Cmd != 1.0 {
      a := s.Start
      b := s.End

      d, angle := distanceAndAngleToFiniteLine(a.X, a.Y, b.X, b.Y, x, y)
      if d < dClosest {
        dClosest = d
        
        if d > 1e-4 {
          aClosest = angle
        }
      }
    }
  }

  return dClosest, aClosest
}

func printSegments(segments []canvaslib.Segment) {
  for _, s := range segments {
    if s.Cmd != 1.0 {
      fmt.Println(s.Start.X, s.Start.Y, s.Cmd)
      fmt.Println(s.End.X, s.End.Y, s.Cmd)
      fmt.Println()
    }
  }

  os.Exit(1)
}

func createPathGlyph(p *canvaslib.Path, name string, scale, advance, originX, originY float64, kernings []Kerning) (*Glyph, error) {
  canvaslib.Tolerance = 0.01

  pFlat := p.Flatten()

  segments := pFlat.Segments()
  data := make([]byte, RESOLUTION*RESOLUTION)

  angle := make([]byte, RESOLUTION*RESOLUTION)

  /*if strings.Contains(name, ":") {
    printSegments(segments)
    os.Exit(1)
  }*/

  for i := 0; i < RESOLUTION; i++ {
    for j := 0; j < RESOLUTION; j++ {
      k := i*RESOLUTION + j
      x := float64(i) + 0.5
      y := float64(j) + 0.5

      d, a := distanceAndAngleToPath(segments, x, y)

      dByte := d*D_PER_PX


      aInit := a  // DEBUG
      for a < 0.0 {
        a += 2.0*math.Pi 
      }

      a = math.Mod(a, 0.5*math.Pi)
      if (a > 0.25*math.Pi) {
        a  = 0.5*math.Pi - a
      }

      if a < 0.0 {
        fmt.Fprintf(os.Stderr, "%g -> %g\n", aInit, a)
        panic("unexpected negative angle")
      }

      aByte := byte((a/(0.25*math.Pi))*255.0)
      angle[k] = aByte

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

  bb := p.Bounds()

  hints := []float64{bb.Y, bb.X + bb.W, bb.Y + bb.H, bb.X}

  return &Glyph{name, data, angle, hints, scale, advance, originX, originY, kernings}, nil
}

func dumpKerning(font *fontlib.SFNT, rIdx rune) {
  gIdx := font.GlyphIndex(rIdx)

  for i := 33; i < 127; i++ {
    other := rune(i)
    otherIdx := font.GlyphIndex(other)

    kerning := font.Kerning(gIdx, otherIdx)

    fmt.Printf("%s: %d\n", string([]rune{rIdx, other}), kerning)
  }
}

func listKernings(font *fontlib.SFNT, gIdx uint16, scale float64) []Kerning {
  res := make([]Kerning, 0)

  for i := 33; i < 127; i++ {
    other := rune(i)
    otherIdx := font.GlyphIndex(other)

    kerning := font.Kerning(gIdx, otherIdx)

    if kerning != 0 {
      res = append(res, Kerning{other, float64(kerning)*scale})
    }
  }

  return res
}
