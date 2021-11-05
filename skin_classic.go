package glui

import (
  "fmt"
  "math"

  "github.com/veandco/go-sdl2/sdl"
)

const (
  RADIO_SIZE = 20 // includes border
  RADIO_DOT_RADIUS = 0.4 // fraction of RADIO_SIZE/2
  TICK_SIZE = 16 // doesn't include border
  TICK_THICKNESS = 0.3 // fraction of the TICK_SIZE/2
)

var (
  DEBUG = false
  DEBUG_MORE = false
)

type ClassicSkin struct {
}

func setColor(d []byte, pixId int, r, g, b, a byte) {
  d[pixId*4+0] = r
  d[pixId*4+1] = g
  d[pixId*4+2] = b
  d[pixId*4+3] = a
}

func setTransparent(d []byte, pixId int) {
  d[pixId*4+3] = 0x00
}

// i is horizontal coord
// j is vertical coord
func ijToPix5x5(i, j int) int {
  return i*5 + j
}

func ijToPix3x3(i, j int) int {
  return i*3 + j
}

func setColor5x5(d []byte, i int, j int, r, g, b, a byte) {
  setColor(d, ijToPix5x5(i, j), r, g, b, a)
}

func setColor3x3(d []byte, i, j int, r, g, b, a byte) {
  setColor(d, ijToPix3x3(i, j), r, g, b, a)
}

func setTransparent5x5(d []byte, i int, j int) {
  setTransparent(d, ijToPix5x5(i, j))
}

func setColorGray(d []byte, pixId int, gray byte) {
  setColor(d, pixId, gray, gray, gray, 0xff)
}

func setColor5x5Gray(d []byte, i int, j int, gray byte) {
  setColor5x5(d, i, j, gray, gray, gray, 0xff)
}

func setColor3x3Gray(d []byte, i, j int, gray byte) {
  setColor3x3(d, i, j, gray, gray, gray, 0xff)
}

// f==0 -> a, f==1.0  -> b
func interpShade(f float64, a, b byte) byte {
  if f < 0.0 {
    f = 0.0
  }
  if f > 1.0 {
    f = 1.0
  }

  res := f*float64(int(b)) + (1.0 - f)*float64(int(a))

  if res < 0.0 {
    return 0
  } else if res >= 255.0 {
    return 255
  } else {
    return byte(int(res))
  }
}

func setColorSDL(d []byte, pixId int, c sdl.Color) {
  d[pixId*4+0] = c.R
  d[pixId*4+1] = c.G
  d[pixId*4+2] = c.B
  d[pixId*4+3] = c.A
}

func swapColor(d []byte, pixId0 int, pixId1 int) {
  r := d[pixId0*4+0] 
  g := d[pixId0*4+1]
  b := d[pixId0*4+2]
  a := d[pixId0*4+3]

  d[pixId0*4+0] = d[pixId1*4+0]
  d[pixId0*4+1] = d[pixId1*4+1]
  d[pixId0*4+2] = d[pixId1*4+2]
  d[pixId0*4+3] = d[pixId1*4+3]

  d[pixId1*4+0] = r
  d[pixId1*4+1] = g
  d[pixId1*4+2] = b
  d[pixId1*4+3] = a
}

// in place
func flipX(d []byte, w int, h int) {
  for i := 0; i < w/2; i++ {
    for j := 0; j < h; j++ {
      a := i*h + j
      b := (w - 1 - i)*h + j

      swapColor(d, a, b)
    }
  }
}

func flipY(d []byte, w int, h int) {
  for i := 0; i < w; i++ {
    for j := 0; j < h/2; j++ {
      a := i*h + j
      b := i*h + (h - 1 - j)

      swapColor(d, a, b)
    }
  }
}

func (s *ClassicSkin) BGColor() sdl.Color {
  return sdl.Color{0xc0, 0xc0, 0xc0, 255}
}

func (s *ClassicSkin) SelColor() sdl.Color {
  return sdl.Color{0x00, 0x00, 196, 255}
}

func (s *ClassicSkin) twoPxOutsetColorBorder(c0 sdl.Color, c1 sdl.Color, c2 sdl.Color, c3 sdl.Color) []byte {
  d := make([]byte, 25*4)

  // left and top side
  for i := 0; i < 4; i++ {
    setColorSDL(d, i, c0)
    setColorSDL(d, i*5, c0)
  }

  // middel top left
  for i := 1; i < 3; i++ {
    for j := 1; j < 3; j++ {
      pixId := i*5+j
      setColorSDL(d, pixId, c1)
    }
  }

  // middel bottom right
  for i := 1; i < 4; i++ {
    setColorSDL(d, 3*5 + i, c2)
    setColorSDL(d, i*5 + 3, c2)
  }

  // bottom and right
  for i := 0; i < 5; i++ {
    setColorSDL(d, 4*5 + i, c3)
    setColorSDL(d, i*5 + 4, c3)
  }

  return d
}

func (s *ClassicSkin) twoPxOutsetBorder(c0 byte, c1 byte, c2 byte, c3 byte) []byte {
  d := make([]byte, 25*4)

  // left and top side
  for i := 0; i < 4; i++ {
    setColor5x5Gray(d, 0, i, c0)
    setColor5x5Gray(d, i, 0, c0)
  }

  // middel top left
  for i := 1; i < 3; i++ {
    for j := 1; j < 3; j++ {
      setColor5x5Gray(d, i, j, c1)
    }
  }

  // middel bottom right
  for i := 1; i < 4; i++ {
    setColor5x5Gray(d, 3, i, c2)
    setColor5x5Gray(d, i, 3, c2)
  }

  // bottom and right
  for i := 0; i < 5; i++ {
    setColor5x5Gray(d, 4, i, c3)
    setColor5x5Gray(d, i, 4, c3)
  }

  return d
}

func (s *ClassicSkin) Button() []byte {
  return s.twoPxOutsetBorder(0xff, 0xc0, 0x80, 0x00)
}

func (s *ClassicSkin) ButtonPressed() []byte {
  d := s.Button()

  flipX(d, 5, 5)
  flipY(d, 5, 5)

  return d
}

func (s *ClassicSkin) Input() []byte {
  d := s.twoPxOutsetBorder(0x80, 0x00, 0xc0, 0xff)

  setColor5x5Gray(d, 2, 2, 0xff)

  return d
}

func (s *ClassicSkin) Focus() []byte {
  c := s.SelColor()

  return s.twoPxOutsetColorBorder(c, c, c, c)
}

func (s *ClassicSkin) Inset() []byte {
  d := s.twoPxOutsetBorder(0x80, 0xc0, 0xc0, 0xff)

  // only outer rim has non-zero alpha

  for i := 1; i <= 3; i++ {
    for j := 1; j <= 3; j++ {
      setTransparent5x5(d, i, j)
    }
  }

  return d
}

func (s *ClassicSkin) Corner() []byte {
  d := make([]byte, 25*4)

  c0 := byte(0xff)
  c1 := byte(0xc0)
  c2 := byte(0x80)
  c3 := byte(0x00)

  // top left of cross
  setColor5x5Gray(d, 0, 0, c0)
  setColor5x5Gray(d, 1, 0, c1)
  setColor5x5Gray(d, 0, 1, c1)
  setColor5x5Gray(d, 1, 1, c1)

  // left middle (same as button)
  setColor5x5Gray(d, 0, 2, c0)
  setColor5x5Gray(d, 1, 2, c1)

  // top right of cross
  setColor5x5Gray(d, 3, 0, c2)
  setColor5x5Gray(d, 4, 0, c0)
  setColor5x5Gray(d, 3, 1, c1)
  setColor5x5Gray(d, 4, 0, c1)

  // top middle (sams as button)
  setColor5x5Gray(d, 2, 0, c0)
  setColor5x5Gray(d, 2, 1, c1)

  // right middle (same as button)
  setColor5x5Gray(d, 3, 2, c2)
  setColor5x5Gray(d, 4, 2, c3)

  return d
}

func (s *ClassicSkin) Bar() []byte {
  d := make([]byte, 9*4)

  c0 := byte(0xff)
  c1 := byte(0xc0)
  c2 := byte(0x80)

  setColor3x3Gray(d, 0, 0, c0)
  setColor3x3Gray(d, 1, 0, c0)
  setColor3x3Gray(d, 2, 0, c0)

  setColor3x3Gray(d, 0, 1, c0)
  setColor3x3Gray(d, 1, 1, c1)
  setColor3x3Gray(d, 2, 1, c2)

  setColor3x3Gray(d, 0, 2, c2)
  setColor3x3Gray(d, 1, 2, c2)
  setColor3x3Gray(d, 2, 2, c2)

  return d
}

type radioCircle struct {
  r float64
}

func (c *radioCircle) inside(x, y float64) bool {
  return x*x + y*y <= c.r*c.r
}

func (c *radioCircle) integral(a, b float64) float64 {
  ta := math.Asin(a/c.r)
  tb := math.Asin(b/c.r)

  return c.r*c.r*((tb - ta)*0.5 + 0.25*(math.Sin(2.0*tb) - math.Sin(2.0*ta)))
}

func (c *radioCircle) intersect(a float64) float64 {
  if a >= c.r {
    return 0.0
  } else {
    return math.Sqrt(c.r*c.r - a*a)
  }
}

func (c *radioCircle) coverage(x0_, y0_, x1_, y1_ float64) float64 {
  x0 := math.Min(math.Abs(x0_), math.Abs(x1_))
  x1 := math.Max(math.Abs(x0_), math.Abs(x1_))

  y0 := math.Min(math.Abs(y0_), math.Abs(y1_))
  y1 := math.Max(math.Abs(y0_), math.Abs(y1_))

  bl := c.inside(x0, y0)
  br := c.inside(x1, y0)
  tr := c.inside(x1, y1)
  tl := c.inside(x0, y1)

  if !bl {
    return 0.0
  } else if tr {
    return (x1 - x0)*(y1 - y0)
  } else if tl && br {
    xc := c.intersect(y1)
    return (xc - x0)*(y1 - y0) + c.integral(xc, x1) - (x1 - xc)*y0
  } else if tl && !br {
    return c.integral(y0, y1) - (y1 - y0)*x0
  } else if !tl && br {
    return c.integral(x0, x1) - (x1 - x0)*y0
  } else {
    xc := c.intersect(y0)
    return c.integral(x0, xc) - (xc - x0)*y0
  }
}

func interpCircleColor(a float64, tl byte, br byte) byte {
  if a > 0.5*math.Pi {
    return tl
  } else if a < 0.0 && a > -0.5*math.Pi {
    return br
  } else if a > 0.0 {
    f := 2.0*a/math.Pi
    return interpShade(f, br, tl)
  } else {
    f := 2.0*(a + math.Pi)/math.Pi
    return interpShade(f, tl, br)
  }
}

func floatToByte(f float64) byte {
  if f < 0.0 {
    return 0
  } else if f >= 255.0 {
    return 255
  } else {
    return byte(int(f))
  }
}

func avgByte(bs []byte, ws []float64) byte {
  wTot := 0.0

  for _, w := range ws {
    wTot += w
  }

  if len(ws) == len(bs) - 1 && wTot <= 1.0 {
    ws = append(ws, 1.0 - wTot)
    wTot = 1.0
  }

  sum := 0.0
  for i, b := range bs {
    sum += ws[i]*float64(int(b))/wTot
  }

  return floatToByte(sum)
}

func (s *ClassicSkin) RadioOff() []byte {
  nSide := RADIO_SIZE

  d := make([]byte, nSide*nSide*4)

  c0 := &radioCircle{float64(nSide)/2.0}
  c1 := &radioCircle{float64(nSide)/2.0 - 1.0} // one pixel smaller than outer
  c2 := &radioCircle{float64(nSide)/2.0 - 2.0} // two pixels smamller than outer

  for i := 0; i < nSide; i++ {
    for j := 0; j < nSide; j++ {
      x0 := float64(i - nSide/2)
      y0 := float64(j - nSide/2)
      x1 := x0 + 1.0
      y1 := y0 + 1.0


      A0 := c0.coverage(x0, y0, x1, y1)
      A1 := c1.coverage(x0, y0, x1, y1)
      A2 := c2.coverage(x0, y0, x1, y1)
      
      outerA := A0 - A1
      innerA := A1 - A2

      //alpha := floatToByte(A0*255.0)
      alpha := floatToByte(255.0)
      
      var shade byte
      if alpha == 0 {
        shade = 0
      } else {
        a := math.Atan2(0.5*(y0 + y1), 0.5*(x0 + x1))
        outerShade := interpCircleColor(a, 0x80, 0xff)
        innerShade := interpCircleColor(a, 0x00, 0xc0)

        shade = avgByte([]byte{outerShade, innerShade, 0xff, 0xc0}, []float64{outerA, innerA, A2})
      }

      setColor(d, i*nSide + (nSide - 1 - j), shade, shade, shade, alpha)
    }
  }

  return d
}

func (s *ClassicSkin) RadioOn() []byte {
  nSide := RADIO_SIZE

  d := make([]byte, nSide*nSide*4)

  c0 := &radioCircle{float64(nSide)/2.0}
  c1 := &radioCircle{float64(nSide)/2.0 - 1.0} // one pixel smaller than outer
  c2 := &radioCircle{float64(nSide)/2.0 - 2.0} // two pixels smamller than outer
  c3 := &radioCircle{RADIO_DOT_RADIUS*float64(nSide)/2.0}

  for i := 0; i < nSide; i++ {
    for j := 0; j < nSide; j++ {
      x0 := float64(i - nSide/2)
      y0 := float64(j - nSide/2)
      x1 := x0 + 1.0
      y1 := y0 + 1.0


      A0 := c0.coverage(x0, y0, x1, y1)
      A1 := c1.coverage(x0, y0, x1, y1)
      A2 := c2.coverage(x0, y0, x1, y1)
      A3 := c3.coverage(x0, y0, x1, y1)
      
      A01 := A0 - A1 // outer edge of boundary
      A12 := A1 - A2 // inner edge of boundary
      A23 := A2 - A3 // white space
      // A3 is black dot

      alpha := floatToByte(255.0)
      
      var shade byte
      if alpha == 0 {
        shade = 0
      } else {
        a := math.Atan2(0.5*(y0 + y1), 0.5*(x0 + x1))
        outerShade := interpCircleColor(a, 0x80, 0xff)
        innerShade := interpCircleColor(a, 0x00, 0xc0)

        shade = avgByte([]byte{outerShade, innerShade, 0xff, 0x00, 0xc0}, []float64{A01, A12, A23, A3})
      }

      setColor(d, i*nSide + (nSide - 1 - j), shade, shade, shade, alpha)
    }
  }

  return d
}


type vec2 struct {
  x float64
  y float64
}

type line2 struct {
  a vec2
  b vec2
}

type ten2 struct {
  c0 vec2
  c1 vec2
}

type poly2 struct {
  ps []vec2
}

// left side of either line
const eps = 1e-6

func diff(a, b vec2) vec2 {
  return vec2{b.x - a.x, b.y - a.y}
}

func (v vec2) sub(w vec2) vec2 {
  return vec2{v.x - w.x, v.y - w.y}
}

func (v vec2) add(w vec2) vec2 {
  return vec2{v.x + w.x, v.y + w.y}
}

func (v vec2) dot(w vec2) float64 {
  return v.x*w.x + v.y*w.y
}

func (v vec2) len() float64 {
  return math.Sqrt(v.dot(v))
}

func (v vec2) normalize() vec2 {
  d := v.len()

  return vec2{v.x/d, v.y/d}
}

func (v vec2) rot90() vec2 {
  return vec2{-v.y, v.x}
}

func (v vec2) scale(s float64) vec2 {
  return vec2{v.x*s, v.y*s}
}

func (v vec2) dist(w vec2) float64 {
  return w.sub(v).len()
}

func (v vec2) eq(w vec2) bool {
  return v.dist(w) < eps
}

func (v vec2) dump() {
  fmt.Println(v.x, v.y)
}

func (t ten2) det() float64 {
  return t.c0.x*t.c1.y - t.c1.x*t.c0.y
}

func (l line2) vec2() vec2 {
  return l.b.sub(l.a)
}

func (l line2) t() vec2 {
  return l.vec2().normalize()
}

func (l line2) len() float64 {
  return l.vec2().len()
}

func (l line2) contains(v vec2) bool {
  //fmt.Println("l contains: ", l.dist(v))
  return l.dist(v) >= -eps
}

func (l line2) dist(v vec2) float64 {
  n := l.t().rot90()

  d := v.sub(l.a)

  return n.dot(d)
}

func (l line2) eq(k line2) bool {
  return (l.a.eq(k.a) && l.b.eq(k.b)) || (l.a.eq(k.b) && l.b.eq(k.a))
}

func (l line2) intersect(k line2) (c vec2, ok bool) {
  sys := ten2{l.vec2(), k.vec2().scale(-1.0)}
  rhs := k.a.sub(l.a)

  d := sys.det()

  if math.Abs(d) < eps {
    ok = false
    return
  }

  alpha := ten2{rhs, sys.c1}.det()/d
  beta := ten2{l.vec2(), rhs}.det()/d

  if alpha < 0.0 - eps || alpha > 1.0 + eps {
    ok = false
  } else {
    ok = true
  }

  c = vec2{
    (1.0 - beta)*k.a.x + beta*k.b.x,
    (1.0 - beta)*k.a.y + beta*k.b.y,
  }

  cCheck := vec2{
    (1.0 - alpha)*l.a.x + alpha*l.b.x,
    (1.0 - alpha)*l.a.y + alpha*l.b.y,
  }

  if !c.eq(cCheck) {
    panic("failed to solve intersect system")
  }

  return
}

func (p poly2) len() int {
  return len(p.ps)
}

func (p poly2) get(i int) vec2 {
  for i < 0 {
    i += p.len()
  }

  for i >= p.len() {
    i -= p.len()
  }

  return p.ps[i]
}

func (p poly2) line(i int) line2 {
  a := p.get(i)
  b := p.get(i+1)

  return line2{a, b}
}

type vec2sorter struct {
  p vec2
  vs []vec2
  lineI []int
  inside []bool
}

func (s *vec2sorter) Len() int  {
  return len(s.vs)
}

func (s *vec2sorter) Less(i, j int) bool {
  di := s.p.dist(s.vs[i])
  dj := s.p.dist(s.vs[j])

  return di < dj
}

func (s *vec2sorter) Swap(i, j int) {
  s.vs[j], s.vs[i] = s.vs[i], s.vs[j]

  if s.lineI != nil {
    s.lineI[j], s.lineI[i] = s.lineI[i], s.lineI[j]
  }

  if s.inside != nil {
    s.inside[j], s.inside[i] = s.inside[i], s.inside[j]
  }
}

func (p poly2) intersect(l line2) (vec2, vec2, int, int) {
  res := make([]vec2, 0)
  lineI := make([]int, 0)

  for i := 0; i < p.len(); i++ {
    k := p.line(i)

    c, ok := l.intersect(k)
    if ok {
      res = append(res, c)
      lineI = append(lineI, i)
    }
  }

  // 1 cut touches corner and is treated as no cut
  if len(res) == 0 || len(res) == 1 {
    return vec2{}, vec2{}, -1, -1
  }

  if len(res) != 2 {
    panic("expected 0, 1 or 2 cuts") 
  }

  firstI := lineI[0]
  lastI := lineI[1]
  
  return res[0], res[1], firstI, lastI
}

/*func cutPolys(p poly2, q poly2) (poly2, bool) {
  ps := make([]vec2, 0)

  i := 0

  tested := make([]line2, 0)
  inTested := func(l line2) bool {
    for _, t := range tested {
      if t.eq(l) {
        return true
      }
    }
    return false
  }

  Outer:
  for {
    l := p.line(i)

    if inTested(l) {
      break Outer
    } else {
      tested = append(tested, l)
    }

    fmt.Println("cutting", l, q)

    cs, lastQLine, lastInside := q.intersect(l)

    fmt.Println("cut result:", cs)

    for _, c := range cs {
      if len(ps) > 0 && c.eq(ps[0]) {
        break Outer
      }

      fmt.Println("appending", c)
      ps = append(ps, c)
    }

    if len(cs) > 0 && !lastInside {
      p, q = q, p
      i = lastQLine
    } else if len(ps) > 0 {
      fmt.Println("appending", l.b)
      ps = append(ps, l.b)
      i++
    }
  }

  if len(ps) > 0 {
    fmt.Println("poly cut with", len(ps), " points")
  }

  return poly2{ps}, len(ps) > 2
}*/

func (p poly2) contains(q poly2) bool {
  for i := 0; i < p.len(); i++ {
    l := p.line(i)

    for j := 0; j < q.len(); j++ {
      if l.dist(q.get(j)) < 0.0 {
        return false
      }
    }
  }

  return true
}

func (p poly2) containsVec2(v vec2) bool {
  return p.contains(poly2{[]vec2{v}})
}

func (p poly2) cutWithLine(l line2) poly2 {
  ps := make([]vec2, 0)

  //fmt.Println("cutting poly", p, " with line", l)
  foundFirstCut := false
  foundLastCut := false
  firstCut := -1
  for i := 0; i < p.len(); i++ {
    v := p.get(i)

    if l.contains(v) {
      if DEBUG_MORE {
        fmt.Println("line contains", v)
      }
      if foundFirstCut && !foundLastCut {
        /*if foundLastCut {
          panic("cuts more than twice")
        }*/

        c0, ok0 := p.line(firstCut).intersect(l)
        c1, ok1 := p.line(i-1).intersect(l)

        if ok0 && ok1 {
          ps = append(ps, c0, c1)

          foundLastCut = true
        } else if DEBUG {
          if !ok0 {
            fmt.Println("bad cut0", p.line(firstCut), l)
          }

          if !ok1 {
            fmt.Println("bad cut1", p.line(i), l)
          }

        }
      }

      //fmt.Println("line contains", v)

      ps = append(ps, v)
    } else if !foundFirstCut {
      if DEBUG {
        fmt.Println("found first cut", v)
      }
      foundFirstCut = true
      firstCut = i-1
    } else if DEBUG {
      fmt.Println("another point not contained:", v)
    }
  }

  if !foundLastCut {
    c0, ok0 := p.line(firstCut).intersect(l)
    c1, ok1 := p.line(-1).intersect(l)

    if ok0 && ok1 {
      ps = append(ps, c0, c1)
    }
    //fmt.Println("#line:", l)
    //p.dump()
    //panic("last cut not found")
    //c0, ok0 := p.line(firstCut).intersect(l)
    //c1, ok1 := p.line(0).intersect(l)
  }

  if len(ps) > 0 && len(ps) < 3 {
    //fmt.Println(ps)
    //panic("incomplete cut")
  }

  return poly2{ps}
}

func (p poly2) cutWithLine_(l line2) poly2 {
  ps := make([]vec2, 0)

  for i := 0; i < p.len(); i++ {
    v := p.get(i)

    if l.contains(v) {
      ps = append(ps, v)
    }

    if vc, ok := p.line(i).intersect(l); ok {
      ps = append(ps, vc)
    }
  }

  return poly2{ps}
}

func (p poly2) cut(q poly2) poly2 {
  for i := 0; i < p.len(); i++ {
    l := p.line(i)

    q_ := q
    q = q.cutWithLine_(l)

    if DEBUG {
      fmt.Println("#cut with", l)
      q.dump()

      if q.len() == 2 {
        DEBUG_MORE = true
        q_.cutWithLine(l)
      }
    }

    if q.len() == 0 {
      return q
    }
  }

  return q
}

func (p poly2) dump() {
  for i := 0; i <= p.len(); i++ {
    p.get(i).dump()
  }

  fmt.Println()
}

func (p poly2) area() float64 {
  if p.len() < 3 {
    return 0.0
  } else if p.len() == 3 {
    b := p.line(0)

    if b.len() < eps {
      return 0.0
    }

    h := b.dist(p.get(2))
    if h < 0.0 {
      h *= -1.0
      //p.dump()
      //panic("h should be positive")
    }

    return b.len()*h*0.5
  } else {
    sum := 0.0

    // split into triangles
    o := p.get(0)
    for i := 0; i < p.len() - 2; i++ {
      a := p.get(1+i)
      b := p.get(2+i)

      q := poly2{[]vec2{
        o, a, b,
      }}

      sum += q.area()
    }

    if math.IsNaN(sum) {
      panic("area sum is nan")
    }
    return sum
  }
}

type tickTri struct {
  a vec2
  b vec2
  c vec2
}

func (t *tickTri) poly2() poly2 {
  return poly2{[]vec2{t.a, t.b, t.c}}
}

func (t *tickTri) area() float64 {
  return t.poly2().area()
}

func (t *tickTri) coverage(x0, y0, x1, y1 float64) float64 {
  p := t.poly2()
  q := poly2{[]vec2{
    vec2{x0, y0},
    vec2{x1, y0},
    vec2{x1, y1},
    vec2{x0, y1},
  }}

  res := p.cut(q)
  if res.len() == 0 {
    return 0.0
  } else {
    a := res.area()
    if math.IsNaN(a) {
      res.dump()
      panic("area of this is nan")
    }

    return a
  }
}

func (s *ClassicSkin) Tick() []byte {
  nSide := TICK_SIZE
  d := make([]byte, nSide*nSide*4)

  size := float64(nSide/2)
  t := TICK_THICKNESS
  d12 := 0.4
  d23 := 0.8

  v0 := vec2{-(t + (d12 + d23)/2.0), 0.0}.scale(size)
  v1 := v0.add(vec2{t, t}.scale(size))
  v2 := v1.add(vec2{d12, -d12}.scale(size))
  v3 := v2.add(vec2{d23, d23}.scale(size))
  v4 := v3.add(vec2{t, -t}.scale(size))
  v5 := v2.add(vec2{0, -2.0*t}.scale(size))

  t0 := &tickTri{v2, v1, v0}
  t1 := &tickTri{v4, v3, v2}
  t2 := &tickTri{v4, v2, v5}
  t3 := &tickTri{v2, v0, v5}

  //fmt.Println(v0, v1, v2, v3, v4, v5)

  for i := 0; i < nSide; i++ {
    for j := 0; j < nSide; j++ {
      x0 := float64(i - nSide/2)
      y0 := float64(j - nSide/2)
      x1 := x0 + 1.0
      y1 := y0 + 1.0

      A0 := t0.coverage(x0, y0, x1, y1)
      A1 := t1.coverage(x0, y0, x1, y1)
      A2 := t2.coverage(x0, y0, x1, y1)
      A3 := t3.coverage(x0, y0, x1, y1)

      A := A0 + A1 + A2 + A3

      if A > 1.0 + eps {
        //fmt.Println("got A:", A)
        panic("area can't be bigger than 1")
      } /*else if i==7 && j == 9 {
        DEBUG = true

        sq := poly2{[]vec2{
          vec2{x0, y0},
          vec2{x1, y0},
          vec2{x1, y1},
          vec2{x0, y1},
        }}

        cut := t3.poly2().cut(sq)

        t3.poly2().dump()
        sq.dump()
        cut.dump()
        if math.IsNaN(A) {
          panic("area is nan")
        }

        fmt.Println(A)
        panic("stop")
      }*/

      //fmt.Println(0.5*(x0 + x1), 0.5*(y0 + y1), A)

      shade := interpShade(A, 0xff, 0x00)

      alpha := byte(255)

      setColor(d, i*nSide + (nSide - 1 - j), shade, shade, shade, alpha)
    }
  }

  return d
}
