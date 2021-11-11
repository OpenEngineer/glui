package glui

import (
  "fmt"
  "math"
  "strings"

  "github.com/veandco/go-sdl2/sdl"
  "github.com/go-gl/gl/v4.1-core/gl"
)

const (
  N_INIT_TRIS     = 2
  GROW_FACTOR     = 2.0

  VTYPE_HIDDEN = 0
  VTYPE_PLAIN  = 1
  VTYPE_SKIN   = 2
  VTYPE_GLYPH  = 3
  VTYPE_IMAGE  = 4
  VTYPE_DUMMY  = 5 // so that aParam isn't optimized out
)

type Float32Buffer struct {
  nComp int // 1, 2, 3 or 4

  loc uint32
  vao uint32
  vbo uint32

  data []float32 // 3 coords per vertex, 3 vertices per tri

  dirty bool // data not in sync with gpu
}

type UInt8Buffer struct {
  nComp int // 1, 2, 3 or 4

  loc uint32
  vao uint32
  vbo uint32

  data []uint8

  dirty bool
}

type DrawPassData struct {
  winW int // needed here to turn pixels coordinates into gl coordinates
  winH int

  // width and height of buffered texture might differ from memory
  texWidth  int 
  texHeight int

  free []uint32

  Pos    *Float32Buffer
  Type   *Float32Buffer
  Param  *Float32Buffer
  Color  *Float32Buffer
  TCoord *Float32Buffer
}

type DrawPass1Data struct {
  DrawPassData

  Skin *SkinMap


  imageTris  []uint32 // so we don't need to search Type for VTYPE_IMAGE
  images     []*ImageData // same pointers are assumed to be same images
  imageInfos map[*ImageData]ImageInfo
}

type DrawPass2Data struct {
  DrawPassData

  Glyphs *GlyphMap
}

func NewFloat32Buffer(nComp int) *Float32Buffer {
  b := &Float32Buffer{
    nComp, 
    0, 0, 0, 
    make([]float32, N_INIT_TRIS*3*nComp), 
    true,
  }

  return b
}

func (b *Float32Buffer) initGL(loc uint32, vao uint32, vbo uint32) {
  b.loc = loc
  b.vao = vao
  b.vbo = vbo

  b.dirty = true
}

func NewUInt8Buffer(nComp int) *UInt8Buffer {
  b := &UInt8Buffer{nComp, 0, 0, 0, make([]uint8, N_INIT_TRIS*3*nComp), true}

  return b
}

func (b *UInt8Buffer) initGL(location uint32, vao uint32, vbo uint32) {
  b.loc = location
  b.vao = vao
  b.vbo = vbo

  b.dirty = true
}

func newDrawPassData() DrawPassData {
  free := make([]uint32, N_INIT_TRIS)
  for i := 0; i < N_INIT_TRIS; i++ {
    free[i] = uint32(i)
  }

  types := NewFloat32Buffer(1)

  // set all types to zero
  for i := 0; i < len(types.data); i++ {
    types.data[i] = VTYPE_HIDDEN
  }

  return DrawPassData{
    0, 0,
    0, 0,
    free,
    NewFloat32Buffer(3),
    types,
    NewFloat32Buffer(1),
    NewFloat32Buffer(4),
    NewFloat32Buffer(2),
  }
}

func newDrawPass1Data(skin *SkinMap) *DrawPass1Data {
  dd := &DrawPass1Data{
    newDrawPassData(), 
    skin, 
    make([]uint32, 0),
    make([]*ImageData, 0),
    make(map[*ImageData]ImageInfo),
  }

  dd.texWidth, dd.texHeight = skin.width, skin.height

  return dd
}

func newDrawPass2Data(glyphs *GlyphMap) *DrawPass2Data {
  dd := &DrawPass2Data{newDrawPassData(), glyphs}

  dd.texWidth, dd.texHeight = glyphs.size, glyphs.size

  return dd
}

func (d *DrawPassData) initGL(
  aPosLoc, aTypeLoc, aParamLoc, aColorLoc, aTCoordLoc uint32,
  aPosVAO, aTypeVAO, aParamVAO, aColorVAO, aTCoordVAO uint32,
  aPosVBO, aTypeVBO, aParamVBO, aColorVBO, aTCoordVBO uint32,
) {
  checkGLError()

  d.Pos.initGL(aPosLoc, aPosVAO, aPosVBO)
  d.Type.initGL(aTypeLoc, aTypeVAO, aTypeVBO)
  d.Param.initGL(aParamLoc, aParamVAO, aParamVBO)
  d.Color.initGL(aColorLoc, aColorVAO, aColorVBO)
  d.TCoord.initGL(aTCoordLoc, aTCoordVAO, aTCoordVBO)

  checkGLError()
}

// number of tris
func (d *DrawPassData) Len() int {
  return len(d.Type.data)/3
}

func (d *DrawPassData) nTris() int {
  return d.Len() - len(d.free)
}

func (d *DrawPassData) Grow() {
  nTrisOld := d.Len()
  nTrisNew := int(float64(nTrisOld)*GROW_FACTOR)

  oldNFree := len(d.free)

  d.free = append(d.free, make([]uint32, nTrisNew - nTrisOld)...)

  for i := 0; i < nTrisNew - nTrisOld; i++ {
    d.free[oldNFree + i] = uint32(nTrisOld + i)
  }

  d.Pos.grow(nTrisNew)
  d.Type.grow(nTrisNew)
  d.Param.grow(nTrisNew)
  d.Color.grow(nTrisNew)
  d.TCoord.grow(nTrisNew)
}

func (d *DrawPassData) GetTri2DPos(tri uint32) (x0 float32, y0 float32, x1 float32, y1 float32, x2 float32, y2 float32) {
  x0 = d.Pos.Get(tri, 0, 0)
  y0 = d.Pos.Get(tri, 0, 1)

  x1 = d.Pos.Get(tri, 1, 0)
  y1 = d.Pos.Get(tri, 1, 1)

  x2 = d.Pos.Get(tri, 2, 0)
  y2 = d.Pos.Get(tri, 2, 1)

  return
}

func (b *Float32Buffer) grow(nTrisNew int) {
  oldData := b.data
  b.data = make([]float32, nTrisNew*3*b.nComp)

  for i, old := range oldData {
    b.data[i] = old
  }

  if b.vbo != 0 {
    checkGLError()
    gl.DeleteBuffers(1, &b.vbo)
    gl.GenBuffers(1, &b.vbo)
    checkGLError()
  }

  b.dirty = true
}

// cant wait for generics :)
func (b *UInt8Buffer) grow(nTrisNew int) {
  oldData := b.data
  b.data = make([]uint8, nTrisNew*3*b.nComp)

  for i, old := range oldData {
    b.data[i] = old
  }

  if b.vbo != 0 {
    checkGLError()
    gl.DeleteBuffers(1, &b.vbo)
    gl.GenBuffers(1, &b.vbo)
    checkGLError()
  }

  b.dirty = true
}

func (d *DrawPassData) Alloc(nTris int) []uint32 {
  offsets := make([]uint32, nTris)

  for i := 0; i < nTris; i++ {
    if len(d.free) > 0 {
      offsets[i] = d.free[0]
      d.free = d.free[1:]
    } else {
      d.Grow()

      if len(d.free) > 0 {
        offsets[i] = d.free[0]
        d.free = d.free[1:]
      } else {
        panic("buffer should've grown")
      }
    }
  }

  return offsets
}

func (d *DrawPassData) Dealloc(offsets []uint32) {
  for _, tri := range offsets {
    d.SetTriType(tri, VTYPE_HIDDEN)
  }

  oldFree := d.free

  d.free = make([]uint32, len(oldFree) + len(offsets))

  i := 0
  j := 0
  k := 0

  for ;i < len(oldFree) || j < len(offsets); {
    if i == len(oldFree) {
      d.free[k] = offsets[j]
      j++
      k++
    } else if j == len(offsets) {
      d.free[k] = oldFree[i]
      i++
      k++
    } else if oldFree[i] < offsets[j] {
      d.free[k] = oldFree[i]
      i++
      k++
    } else {
      d.free[k] = offsets[j]
      j++
      k++
    }
  }

  if k != len(d.free) {
    panic("k should be == len(b.free)")
  }
}

// if diff is
func (d *DrawPassData) Resize(tris []uint32, nNew int) []uint32 {
  nDiff := nNew - len(tris)

  if nDiff > 0 {
    return append(tris, d.Alloc(nDiff)...)
  } else if nDiff < 0 {
    remove := tris[nNew:]
    d.Dealloc(remove)
    return tris[0:nNew]
  } else {
    return tris
  }
}

func (b *Float32Buffer) Get(triId uint32, vertexId uint32, compId uint32) float32 {
  offset := (triId*3 + vertexId)*uint32(b.nComp)

  return b.data[offset + compId]
}

func (b *Float32Buffer) Get2(triId uint32, vertexId uint32) (float32, float32) {
  offset := (triId*3 + vertexId)*uint32(b.nComp)

  return b.data[offset], b.data[offset+1]
}

func (b *Float32Buffer) Set(triId uint32, vertexId uint32, compId uint32, value float32) {
  offset := (triId*3 + vertexId)*uint32(b.nComp)

  b.data[offset + compId] = value

  // dirty even if value is the same, so we can use that as a kind of forceRecalcPos trigger
  b.dirty = true
}

func (b *Float32Buffer) Set1(triId uint32, vertexId uint32, value float32) {
  offset := (triId*3 + vertexId)*uint32(b.nComp)

  b.data[offset + 0] = value

  b.dirty = true
}

func (b *Float32Buffer) Set1Const(triId uint32, value float32) {
  offset := triId*3

  b.data[(offset + 0)*uint32(b.nComp)] = value
  b.data[(offset + 1)*uint32(b.nComp)] = value
  b.data[(offset + 2)*uint32(b.nComp)] = value

  b.dirty = true
}

func (b *Float32Buffer) Set2(triId uint32, vertexId uint32, value0 float32, value1 float32) {
  offset := (triId*3 + vertexId)*uint32(b.nComp)

  b.data[offset + 0] = value0
  b.data[offset + 1] = value1

  b.dirty = true
}

func (b *Float32Buffer) Set2Const(triId uint32, value0 float32, value1 float32) {
  offset := triId*3

  b.data[(offset + 0)*uint32(b.nComp) + 0] = value0
  b.data[(offset + 0)*uint32(b.nComp) + 1] = value1

  b.data[(offset + 1)*uint32(b.nComp) + 0] = value0
  b.data[(offset + 1)*uint32(b.nComp) + 1] = value1

  b.data[(offset + 2)*uint32(b.nComp) + 0] = value0
  b.data[(offset + 2)*uint32(b.nComp) + 1] = value1

  b.dirty = true
}

func (b *Float32Buffer) Set3(triId uint32, vertexId uint32, value0 float32, value1 float32, value2 float32) {
  offset := (triId*3 + vertexId)*uint32(b.nComp)

  b.data[offset + 0] = value0
  b.data[offset + 1] = value1
  b.data[offset + 2] = value2

  b.dirty = true
}

func (b *Float32Buffer) Set4(triId uint32, vertexId uint32, 
  value0 float32, value1 float32, value2 float32, value3 float32) {
  offset := (triId*3 + vertexId)*uint32(b.nComp)

  b.data[offset + 0] = value0
  b.data[offset + 1] = value1
  b.data[offset + 2] = value2
  b.data[offset + 3] = value3

  b.dirty = true
}

func (b *Float32Buffer) Set4Const(triId uint32, value0 float32, value1 float32, value2 float32, value3 float32) {
  b.Set4(triId, 0, value0, value1, value2, value3)
  b.Set4(triId, 1, value0, value1, value2, value3)
  b.Set4(triId, 2, value0, value1, value2, value3)
}

func (b *Float32Buffer) dumpComp(compId int) string {
  var sb strings.Builder

  nTris := len(b.data)/b.nComp

  sb.WriteString("[")
  
  for triId := 0; triId < nTris; triId++ {
    c := b.data[triId*b.nComp + compId]

    sb.WriteString(fmt.Sprintf("%g", c))

    if triId < nTris - 1 {
      sb.WriteString(", ")
    }
  }

  sb.WriteString("]")

  return sb.String()
}

func (b *Float32Buffer) dumpLines() string {
  var sb strings.Builder

  nTris := len(b.data)/b.nComp

  sb.WriteString("[")
  
  for triId := 0; triId < nTris; triId++ {
    sb.WriteString("\n  ")

    for iComp := 0; iComp < b.nComp; iComp++ {
      c := b.data[triId*b.nComp + iComp]
      sb.WriteString(fmt.Sprintf("%g", c))

      if iComp < b.nComp - 1 {
        sb.WriteString(" ")
      }

    }
  }

  sb.WriteString("\n]")

  return sb.String()
}

func (b *UInt8Buffer) Set(triId uint32, vertexId uint32, compId uint32, value uint8) {
  offset := (triId*3 + vertexId)*uint32(b.nComp)

  b.data[offset + compId] = value

  b.dirty = true
}

func (b *UInt8Buffer) Set1(triId uint32, vertexId uint32, value uint8) {
  offset := (triId*3 + vertexId)*uint32(b.nComp)

  b.data[offset + 0] = value

  b.dirty = true
}

func (b *UInt8Buffer) Set1Const(triId uint32, value uint8) {
  offset := triId*3

  b.data[(offset + 0)*uint32(b.nComp)] = value
  b.data[(offset + 1)*uint32(b.nComp)] = value
  b.data[(offset + 2)*uint32(b.nComp)] = value

  b.dirty = true
}

func (b *UInt8Buffer) Set4(triId uint32, vertexId uint32, value0 uint8, value1 uint8, value2 uint8, value3 uint8) {
  offset := (triId*3 + vertexId)*uint32(b.nComp)

  b.data[offset + 0] = value0
  b.data[offset + 1] = value1
  b.data[offset + 2] = value2
  b.data[offset + 3] = value3

  b.dirty = true
}

func (b *Float32Buffer) sync() {
  if b.dirty {
    b.bind()

    gl.BufferData(gl.ARRAY_BUFFER, 4*len(b.data), gl.Ptr(b.data), gl.STATIC_DRAW)

    checkGLError()

    b.dirty = false
  }
}

func (b *UInt8Buffer) sync() {
  if b.dirty {
    b.bind()

    gl.BufferData(gl.ARRAY_BUFFER, len(b.data), gl.Ptr(b.data), gl.STATIC_DRAW)

    checkGLError()

    b.dirty = false
  }
}

func (b *Float32Buffer) bind() {
  checkGLError()

  // XXX: I could not get this to work with vaos, which shows how shit the OpenGL API has been designed
  //gl.EnableVertexArrayAttrib(b.vao, b.loc)

  gl.EnableVertexAttribArray(b.loc)
  gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
  gl.VertexAttribPointer(b.loc, int32(b.nComp), gl.FLOAT, false, 0, nil)

  checkGLError()
}

func (b *UInt8Buffer) bind() {
  checkGLError()

  //gl.EnableVertexArrayAttrib(b.vao, b.loc)

  gl.EnableVertexAttribArray(b.loc)
  gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
  gl.VertexAttribPointer(b.loc, int32(b.nComp), gl.UNSIGNED_BYTE, false, 0, nil)

  checkGLError()
}

func (d *DrawPassData) SetPos(triId uint32, vertexId uint32, x_ int, y_ int, z float32) {
  x := 2.0*float32(x_)/float32(d.winW) - 1.0
  y := 1.0 - 2.0*float32(y_)/float32(d.winH)

  d.Pos.Set3(triId, vertexId, x, y, z)
}

func assertReal(x float32, name string) {
  if math.IsNaN(float64(x)) {
    panic(name + " can't be NaN")
  } else if math.IsInf(float64(x), 0) {
    panic(name + " can't be Inf")
  }
}

func (d *DrawPassData) translatePos(triId uint32, vertexId uint32, dx float32, dy float32, dz float32) {
  if dx != 0.0 {
    assertReal(dx, "dx")

    xStart := d.Pos.Get(triId, vertexId, 0)
    d.Pos.Set(triId, vertexId, 0, xStart+dx)
  }

  if dy != 0.0 {
    assertReal(dy, "dy")

    yStart := d.Pos.Get(triId, vertexId, 1)
    d.Pos.Set(triId, vertexId, 1, yStart+dy)
  }

  if dz != 0.0 {
    assertReal(dz, "dz")

    zStart := d.Pos.Get(triId, vertexId, 2)
    d.Pos.Set(triId, vertexId, 2, zStart+dz)
  }
}

func (d *DrawPassData) TranslatePos(triId uint32, vertexId uint32, dx_ int, dy_ int, dz float32) {
  dx := 2.0*float32(dx_)/float32(d.winW)
  dy := -2.0*float32(dy_)/float32(d.winH)

  d.translatePos(triId, vertexId, dx, dy, dz)
}

func (d *DrawPassData) TranslateTri(triId uint32, dx_ int, dy_ int, dz float32) {
  dx := 2.0*float32(dx_)/float32(d.winW)
  dy := -2.0*float32(dy_)/float32(d.winH)

  d.translatePos(triId, 0, dx, dy, dz)
  d.translatePos(triId, 1, dx, dy, dz)
  d.translatePos(triId, 2, dx, dy, dz)
}

func (d *DrawPassData) CropTri(tri uint32, r_ Rect) {
  xMin := 2.0*float32(r_.X)/float32(d.winW) - 1.0
  xMax := 2.0*float32(r_.Right())/float32(d.winW) - 1.0

  yMax := 1.0 - 2.0*float32(r_.Y)/float32(d.winH)
  yMin := 1.0 - 2.0*float32(r_.Bottom())/float32(d.winH)

  du_dx :=  0.5*float32(d.winW)/float32(d.texWidth)
  dv_dy := -0.5*float32(d.winH)/float32(d.texHeight)

  // pos is impacted, but texture pos too!
  d.cropPosAndTCoord(tri, 0, xMin, yMin, xMax, yMax, du_dx, dv_dy)
  d.cropPosAndTCoord(tri, 1, xMin, yMin, xMax, yMax, du_dx, dv_dy)
  d.cropPosAndTCoord(tri, 2, xMin, yMin, xMax, yMax, du_dx, dv_dy)
}

func (d *DrawPassData) cropPosAndTCoord(tri uint32, vertex uint32, xMin, yMin, xMax, yMax float32, du_dx, dv_dy float32) {
  x, y := d.Pos.Get2(tri, vertex)
  u, v := d.GetTCoord(tri, vertex)

  dx := float32(0.0)
  dy := float32(0.0)

  if x < xMin {
    dx = xMin - x
  } else if x > xMax {
    dx = xMax - x
  }

  if y < yMin {
    dy = yMin - y
  } else if y > yMax {
    dy = yMax - y
  }

  if dx != float32(0.0) || dy != (0.0) {
    du := du_dx*dx
    dv := dv_dy*dy

    d.Pos.Set2(tri, vertex, x + dx, y + dy)
    d.TCoord.Set2(tri, vertex, v + dv, u + du)
  }
}

func (d *DrawPassData) SetPosF(triId uint32, vertexId uint32, x_ float64, y_ float64, z float32) {
  x := 2.0*float32(x_)/float32(d.winW) - 1.0
  y := 1.0 - 2.0*float32(y_)/float32(d.winH)

  d.Pos.Set3(triId, vertexId, x, y, z)
}

func (d *DrawPassData) SetQuadPos(tri0 uint32, tri1 uint32, r Rect, z float32) {
  d.SetPos(tri0, 0, r.X, r.Y, z)
  d.SetPos(tri0, 1, r.Right(), r.Y, z)
  d.SetPos(tri0, 2, r.X, r.Bottom(), z)

  d.SetPos(tri1, 0, r.Right(), r.Bottom(), z)
  d.SetPos(tri1, 1, r.Right(), r.Y, z)
  d.SetPos(tri1, 2, r.X, r.Bottom(), z)
}

func (d *DrawPassData) SetQuadPosF(tri0 uint32, tri1 uint32, r RectF, z float32) {
  d.SetPosF(tri0, 0, r.X, r.Y, z)
  d.SetPosF(tri0, 1, r.Right(), r.Y, z)
  d.SetPosF(tri0, 2, r.X, r.Bottom(), z)

  d.SetPosF(tri1, 0, r.Right(), r.Bottom(), z)
  d.SetPosF(tri1, 1, r.Right(), r.Y, z)
  d.SetPosF(tri1, 2, r.X, r.Bottom(), z)
}

func (d *DrawPass1Data) SetSkinCoord(triId uint32, vertexId uint32, x_ int, y_ int) {
  x := float32(x_)/float32(d.texWidth)
  y := float32(y_)/float32(d.texHeight)

  d.SetTCoord(triId, vertexId, x, y)
}

// if size of texture changes, all the coords need to change
func (d *DrawPass1Data) syncSkinSize() {
  w, h := d.Skin.width, d.Skin.height

  if w == d.texWidth && h == d.texHeight {
    return
  }

  xScale := float32(d.texWidth)/float32(w)
  yScale := float32(d.texHeight)/float32(h)

  for triId := 0; triId < d.Len(); triId++ {
    for vertexId := 0; vertexId < 3; vertexId++ {
      // scale regardless of Type (a tri might be hidden, and then switch to another Type later on)
      x, y := d.GetTCoord(uint32(triId), uint32(vertexId))

      x = x*xScale
      y = y*yScale

      d.SetTCoord(uint32(triId), uint32(vertexId), x, y)
    }
  }

  d.TCoord.dirty = true
  d.texWidth = w
  d.texHeight = h
}

func (d *DrawPassData) SetTriType(triId uint32, value float32) {
  d.Type.Set1Const(triId, value)
}

func (d *DrawPass1Data) SetQuadImage(tri0, tri1 uint32, img *ImageData) {
  d.Type.Set1Const(tri0, VTYPE_IMAGE)
  d.Type.Set1Const(tri1, VTYPE_IMAGE)

  tri0Index := -1
  tri1Index := -1
  for i, tri := range d.imageTris {
    if tri == tri0 {
      tri0Index = i
    }

    if tri == tri1 {
      tri1Index = i
    }

    if tri0Index >= 0 && tri1Index >= 0 {
      break
    }
  }

  if tri0Index == -1 {
    d.imageTris = append(d.imageTris, tri0)
    d.images = append(d.images, img)
  } else {
    d.images[tri0Index] = img
  }

  if tri1Index == -1 {
    d.imageTris = append(d.imageTris, tri1)
    d.images = append(d.images, img)
  } else {
    d.images[tri1Index] = img
  }

  d.setQuadImageRelTCoords(tri0, tri1, img.W, img.H)
}

func (d *DrawPassData) setQuadImageRelTCoords(tri0, tri1 uint32, w_, h_ int) {
  w := float32(w_)/float32(d.texWidth)
  h := float32(h_)/float32(d.texHeight)

  d.SetTCoord(tri0, 0, 0, 0)
  d.SetTCoord(tri0, 1, w, 0) 
  d.SetTCoord(tri0, 2, 0, h)

  d.SetTCoord(tri1, 0, w, h)
  d.SetTCoord(tri1, 1, w, 0)
  d.SetTCoord(tri1, 2, 0, h)
}

func (d *DrawPassData) SetTCoord(triId uint32, vertexId uint32, x, y float32) {
  /*if x < 0.0 || y < 0.0 {
    panic("can't be negative")
  }*/

  // XXX transpose for some reason
  d.TCoord.Set2(triId, vertexId, y, x)
}

func (d *DrawPassData) GetTCoord(triId uint32, vertexId uint32) (float32, float32) {
  // XXX transpose for some reason
  y, x := d.TCoord.Get2(triId, vertexId)

  return x, y
}

func (d *DrawPassData) SetColorConst(triId uint32, c sdl.Color) {
  r := float32(c.R)/float32(256)
  g := float32(c.G)/float32(256)
  b := float32(c.B)/float32(256)
  a := float32(c.A)/float32(256)

  d.Color.Set4Const(triId, r, g, b, a)
}

func (d *DrawPassData) SetQuadColorLinearVGrad(tri0 uint32, tri1 uint32, cTop, cBottom sdl.Color) {
  rT := float32(cTop.R)/float32(256)
  gT := float32(cTop.G)/float32(256)
  bT := float32(cTop.B)/float32(256)
  aT := float32(cTop.A)/float32(256)

  rB := float32(cBottom.R)/float32(256)
  gB := float32(cBottom.G)/float32(256)
  bB := float32(cBottom.B)/float32(256)
  aB := float32(cBottom.A)/float32(256)

  d.Color.Set4(tri0, 0, rT, gT, bT, aT)
  d.Color.Set4(tri0, 1, rT, gT, bT, aT)
  d.Color.Set4(tri0, 2, rB, gB, bB, aB)

  d.Color.Set4(tri1, 0, rB, gB, bB, aB)
  d.Color.Set4(tri1, 1, rT, gT, bT, aT)
  d.Color.Set4(tri1, 2, rB, gB, bB, aB)
}

func (d *DrawPass2Data) SetGlyphCoord(triId uint32, vertexId uint32, x_ int, y_ int) {
  x := float32(x_)/float32(d.Glyphs.size)
  y := float32(y_)/float32(d.Glyphs.size)

  d.TCoord.Set2(triId, vertexId, y, x)
}

func (d *DrawPass2Data) SetGlyphCoords(tri0, tri1 uint32, name string) {
  g := d.Glyphs.GetGlyph(name)

  k := g.TexId
  
  r := d.Glyphs.GetRect(k)

  d.SetGlyphCoord(tri0, 0, r.X, r.Y)
  d.SetGlyphCoord(tri0, 1, r.Right(), r.Y)
  d.SetGlyphCoord(tri0, 2, r.X, r.Bottom())

  d.SetGlyphCoord(tri1, 0, r.Right(), r.Bottom())
  d.SetGlyphCoord(tri1, 1, r.Right(), r.Y)
  d.SetGlyphCoord(tri1, 2, r.X, r.Bottom())
}

// transposed version
func (d *DrawPass2Data) SetGlyphCoordsT(tri0, tri1 uint32, name string) {
  g := d.Glyphs.GetGlyph(name)

  k := g.TexId
  
  r := d.Glyphs.GetRect(k)

  d.SetGlyphCoord(tri0, 0, r.X, r.Y)
  d.SetGlyphCoord(tri0, 1, r.X, r.Bottom())
  d.SetGlyphCoord(tri0, 2, r.Right(), r.Y)

  d.SetGlyphCoord(tri1, 0, r.Right(), r.Bottom())
  d.SetGlyphCoord(tri1, 1, r.X, r.Bottom())
  d.SetGlyphCoord(tri1, 2, r.Right(), r.Y)
}

func (d *DrawPassData) ForceAllDirty() {
  d.Pos.dirty = true
  d.Type.dirty = true
  d.Param.dirty = true
  d.Color.dirty = true
  d.TCoord.dirty = true
}

func (d *DrawPassData) SyncAndBind() {
  d.Pos.sync()
  d.Type.sync()
  d.Param.sync()
  d.Color.sync()
  d.TCoord.sync()

  d.Pos.bind()
  d.Type.bind()
  d.Param.bind()
  d.Color.bind()
  d.TCoord.bind()
}

func (d *DrawPass1Data) SyncAndBind() {
  d.Skin.bind()

  d.syncSkinSize()

  d.DrawPassData.SyncAndBind()
}

func (d *DrawPass2Data) SyncAndBind() {
  d.DrawPassData.SyncAndBind()

  d.Glyphs.bind()
}

// XXX: should we make a distinction between posDirty and dirty?
func (d *DrawPassData) dirty() bool {
  return d.Pos.dirty || d.Type.dirty || d.Param.dirty || d.TCoord.dirty || d.Color.dirty
}

func (d *DrawPassData) posDirty() bool {
  return d.Type.dirty // a change of type indicates that some elements became visible/hidden, and thus affect positioning of siblings etc.
}

func (d *DrawPassData) forcePosDirty() {
  d.Type.dirty = true
}

func (d *DrawPassData) clearPosDirty() {
  d.Type.dirty = false
}

func (d *DrawPass1Data) showBorderedElement(tris []uint32) {
  for i := 0; i < 3; i++ {
    for j := 0; j < 3; j++ {
      tri0 := tris[(i*3 + j)*2 + 0]
      tri1 := tris[(i*3 + j)*2 + 1]

      if (i == 1 && j == 1) {
        d.SetTriType(tri0, VTYPE_PLAIN)
        d.SetTriType(tri1, VTYPE_PLAIN)
      } else {
        d.SetTriType(tri0, VTYPE_SKIN)
        d.SetTriType(tri1, VTYPE_SKIN)
      }
    }
  }
}

// (i,j) is top left of tri
func (d *DrawPass1Data) setTopLeftTriSkinCoords(tri uint32, i int, j int, x [4]int, y [4]int) {
  d.SetSkinCoord(tri, 0, x[i],   y[j])
  d.SetSkinCoord(tri, 1, x[i+1], y[j])
  d.SetSkinCoord(tri, 2, x[i],   y[j+1])
}

// transposed version
func (d *DrawPass1Data) setTopLeftTriSkinCoordsT(tri uint32, i int, j int, x [4]int, y [4]int) {
  d.SetSkinCoord(tri, 0, x[i],   y[j])
  d.SetSkinCoord(tri, 1, x[i], y[j+1])
  d.SetSkinCoord(tri, 2, x[i+1],   y[j])
}

// (i,j) is top left of corresponding top left tri
func (d *DrawPass1Data) setBottomRightTriSkinCoords(tri uint32, i int, j int, x [4]int, y [4]int) {
  d.SetSkinCoord(tri, 0, x[i+1], y[j+1])
  d.SetSkinCoord(tri, 1, x[i+1], y[j])
  d.SetSkinCoord(tri, 2, x[i],   y[j+1])
}

// transposed version
func (d *DrawPass1Data) setBottomRightTriSkinCoordsT(tri uint32, i int, j int, x [4]int, y [4]int) {
  d.SetSkinCoord(tri, 0, x[i+1], y[j+1])
  d.SetSkinCoord(tri, 1, x[i], y[j+1])
  d.SetSkinCoord(tri, 2, x[i+1],   y[j])
}

func (d *DrawPass1Data) setQuadSkinCoords(topRightTri uint32, bottomLeftTri uint32, i, j int, x [4]int, y [4]int) {
  d.setTopLeftTriSkinCoords(topRightTri, i, j, x, y)
  d.setBottomRightTriSkinCoords(bottomLeftTri, i, j, x, y)
}

// transposed version
func (d *DrawPass1Data) setQuadSkinCoordsT(topRightTri uint32, bottomLeftTri uint32, i, j int, x [4]int, y [4]int) {
  d.setTopLeftTriSkinCoordsT(topRightTri, i, j, x, y)
  d.setBottomRightTriSkinCoordsT(bottomLeftTri, i, j, x, y)
}

// also used by input
func (d *DrawPass1Data) setBorderedElementTypesAndTCoords(tris []uint32, x0, y0 int, t int, bgColor sdl.Color) {
  x, y := getBorderedSkinCoords(x0, y0, t)

  for i := 0; i < 3; i++ {
    for j := 0; j < 3; j++ {
      tri0 := tris[(i*3 + j)*2 + 0]
      tri1 := tris[(i*3 + j)*2 + 1]

      if (i == 1 && j == 1) {
        d.SetTriType(tri0, VTYPE_PLAIN)
        d.SetColorConst(tri0, bgColor)
        //d.TCoord.Set2Const(tri0, 0.0, 0.0)

        d.SetTriType(tri1, VTYPE_PLAIN)
        d.SetColorConst(tri1, bgColor)
        //d.TCoord.Set2Const(tri1, 0.0, 0.0)
      } else {
        d.SetTriType(tri0, VTYPE_SKIN)
        d.Color.Set4Const(tri0, 1.0, 1.0, 1.0, 1.0)
        //d.SetSkinCoord(tri0, 0, x[i], y[j])
        //d.SetSkinCoord(tri0, 1, x[i+1], y[j])
        //d.SetSkinCoord(tri0, 2, x[i], y[j+1])

        d.SetTriType(tri1, VTYPE_SKIN)
        d.Color.Set4Const(tri1, 1.0, 1.0, 1.0, 1.0)

        d.setQuadSkinCoords(tri0, tri1, i, j, x, y)
        //d.SetSkinCoord(tri1, 0, x[i+1], y[j+1])
        //d.SetSkinCoord(tri1, 1, x[i+1], y[j])
        //d.SetSkinCoord(tri1, 2, x[i], y[j+1])
      }
    }
  }
}

func (d *DrawPass1Data) setInputLikeElementTypesAndTCoords(tris []uint32) {
  borderT := d.Skin.InputBorderThickness()

  x0, y0 := d.Skin.InputOrigin()

  d.setBorderedElementTypesAndTCoords(tris, x0, y0, borderT, d.Skin.InputBGColor())
}

func (d *DrawPass1Data) setBorderedElementPos(tris []uint32, width, height, t int, z float32) {
  var (
    x [4]int
    y [4]int
  )

  x[0] = 0
  x[1] = x[0] + t
  x[2] = x[0] + width - t
  x[3] = x[0] + width

  y[0] = 0
  y[1] = y[0] + t
  y[2] = y[0] + height - t
  y[3] = y[0] + height

  for i := 0; i < 3; i++ {
    for j := 0; j < 3; j++ {
      tri0 := tris[(i*3 + j)*2 + 0]
      tri1 := tris[(i*3 + j)*2 + 1]

      d.SetPos(tri0, 0, x[i], y[j], z)
      d.SetPos(tri0, 1, x[i+1], y[j], z)
      d.SetPos(tri0, 2, x[i], y[j+1], z)

      d.SetPos(tri1, 0, x[i+1], y[j+1], z)
      d.SetPos(tri1, 1, x[i+1], y[j], z)
      d.SetPos(tri1, 2, x[i], y[j+1], z)
    }
  }
}

func (d *DrawPass1Data) setButtonStyle(tris []uint32) {
  x0, y0 := d.Skin.ButtonOrigin()

  c := d.Skin.BGColor()

  t := d.Skin.ButtonBorderThickness()

  d.setBorderedElementTypesAndTCoords(tris, x0, y0, t, c)
}

func (d *DrawPass1Data) SyncImagesToTexture() {
  //someAdded := false

  for k, info := range d.imageInfos {
    // reset the used
    d.imageInfos[k] = ImageInfo{info.X, info.Y, false}
  }

  for i, tri := range d.imageTris {
    img := d.images[i]

    // verify that the type is actually zero
    t := d.Type.Get(tri, 0, 0)
    if t != VTYPE_IMAGE {
      continue
    }

    // check that the tri has a positive size (i.e. isn't 'cropped-out')
    x0, y0, x1, y1, x2, y2 := d.GetTri2DPos(tri)

    A := triArea(x0, y0, x1, y1, x2, y2)
    if A <= 1e-8 {
      continue
    }

    info, ok := d.imageInfos[img]
    if !ok {
      d.imageInfos[img] = ImageInfo{-1, -1, true}
    } else {
      d.imageInfos[img] = ImageInfo{info.X, info.Y, true}
    }
  }

  for i, tri := range d.imageTris {
    img := d.images[i]

    info, ok := d.imageInfos[img]
    if ok && info.Used {
      if info.X < 0 || info.Y < 0 {
        info.X, info.Y = d.Skin.AllocImage(img, d.imageInfos) // at this point the unused images can be deallocated if needed

        d.imageInfos[img] = ImageInfo{info.X, info.Y, true}

        //someAdded = true
      }

      // now add origin to the texture coords
      d.translateTCoord(tri, 0, float32(info.X), float32(info.Y))
      d.translateTCoord(tri, 1, float32(info.X), float32(info.Y))
      d.translateTCoord(tri, 2, float32(info.X), float32(info.Y))
    }
  }

  //if someAdded {
    //if err := d.Skin.tb.ToImage("skin_w_images.png"); err != nil {
      //panic(err)
    //}
  //}
}

func (d *DrawPass1Data) translateTCoord(triId uint32, vertexId uint32, dx, dy float32) {
  x, y := d.GetTCoord(triId, vertexId)

  x = x + dx/float32(d.texWidth)
  y = y + dy/float32(d.texHeight)

  d.SetTCoord(triId, vertexId, x, y)
}
