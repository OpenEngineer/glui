package glui

import (
  "fmt"
  "math"
  "strconv"
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
  VTYPE_DUMMY  = 4 // so that aParam isn't optimized out
)

type Float32Buffer struct {
  nComp int // 1, 2, 3 or 4

  loc uint32
  vbo uint32

  data []float32 // 3 coords per vertex, 3 vertices per tri

  dirty bool // data not in sync with gpu
}

type UInt8Buffer struct {
  nComp int // 1, 2, 3 or 4

  loc uint32
  vbo uint32

  data []uint8

  dirty bool
}

type DrawPassData struct {
  w int
  h int

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
}

type DrawPass2Data struct {
  DrawPassData

  Glyphs *GlyphMap
}

func NewFloat32Buffer(nComp int) *Float32Buffer {
  b := &Float32Buffer{nComp, 0, 0, make([]float32, N_INIT_TRIS*3*nComp), true}

  return b
}

func (b *Float32Buffer) InitGL(location uint32) {
  b.loc = location

  checkGLError()
  gl.GenBuffers(1, &b.vbo)
  checkGLError()
  gl.EnableVertexAttribArray(b.loc)
  checkGLError()
  gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
  checkGLError()
  gl.VertexAttribPointer(b.loc, int32(b.nComp), gl.FLOAT, false, 0, nil)
  checkGLError()
  gl.BindBuffer(gl.ARRAY_BUFFER, 0)
  checkGLError()
  gl.BindVertexArray(0)

  checkGLError()

  b.dirty = true
}

func NewUInt8Buffer(nComp int) *UInt8Buffer {
  b := &UInt8Buffer{nComp, 0, 0, make([]uint8, N_INIT_TRIS*3*nComp), true}

  return b
}

func (b *UInt8Buffer) InitGL(location uint32) {
  b.loc = location

  checkGLError()

  gl.GenBuffers(1, &b.vbo)
  gl.EnableVertexAttribArray(b.loc)
  gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
  gl.VertexAttribPointer(b.loc, int32(b.nComp), gl.UNSIGNED_BYTE, false, 0, nil)
  gl.BindBuffer(gl.ARRAY_BUFFER, 0)
  gl.BindVertexArray(0)

  checkGLError()

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
    free,
    NewFloat32Buffer(3),
    types,
    NewFloat32Buffer(1),
    NewFloat32Buffer(4),
    NewFloat32Buffer(2),
  }
}

func newDrawPass1Data(skin *SkinMap) *DrawPass1Data {
  return &DrawPass1Data{newDrawPassData(), skin}
}

func newDrawPass2Data(glyphs *GlyphMap) *DrawPass2Data {
  return &DrawPass2Data{newDrawPassData(), glyphs}
}

func getGLAttribLocation(prog uint32, name string) uint32 {
  loc := gl.GetAttribLocation(prog, gl.Str(name + "\x00"))

  if loc < 0 {
    panic(name + " bad attrib location: " + strconv.Itoa(int(loc)))
  }

  return uint32(loc)
}

func getGLUniformLocation(prog uint32, name string) uint32 {
  loc := gl.GetUniformLocation(prog, gl.Str(name + "\x00"))

  if loc < 0 {
    panic(name + " bad uniform location: " + strconv.Itoa(int(loc)))
  }

  return uint32(loc)
}

func (d *DrawPassData) InitGL(prog uint32) {
  checkGLError()

  posLoc := getGLAttribLocation(prog, "aPos")
  typeLoc := getGLAttribLocation(prog, "aType")
  paramLoc := getGLAttribLocation(prog, "aParam")
  colorLoc := getGLAttribLocation(prog, "aColor")
  tcoordLoc := getGLAttribLocation(prog, "aTCoord")

  fmt.Println("draw pass param locations", posLoc, typeLoc, paramLoc, colorLoc, tcoordLoc)

  d.Pos.InitGL(uint32(posLoc))
  d.Type.InitGL(uint32(typeLoc))
  d.Param.InitGL(uint32(paramLoc))
  d.Color.InitGL(uint32(colorLoc))
  d.TCoord.InitGL(uint32(tcoordLoc))

  checkGLError()
}

func (d *DrawPass1Data) InitGL(prog uint32) {
  gl.UseProgram(prog)

  d.DrawPassData.InitGL(prog)

  checkGLError()
  skinLoc := gl.GetUniformLocation(prog, gl.Str("skin\x00"))
  d.Skin.InitGL(uint32(skinLoc))
  checkGLError()
}

func (d *DrawPass2Data) InitGL(prog uint32) {
  gl.UseProgram(prog)

  d.DrawPassData.InitGL(prog)

  checkGLError()
  glyphLoc := gl.GetUniformLocation(prog, gl.Str("glyphs\x00"))
  d.Glyphs.InitGL(uint32(glyphLoc))
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
    d.Type.Set1Const(tri, VTYPE_HIDDEN)
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
  checkGLError()
    gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
    gl.BufferData(gl.ARRAY_BUFFER, 4*len(b.data), gl.Ptr(b.data), gl.STATIC_DRAW)
    gl.BindBuffer(gl.ARRAY_BUFFER, 0)
  checkGLError()

    b.dirty = false
  }
}

func (b *UInt8Buffer) sync() {
  if b.dirty {
  checkGLError()
    gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
    gl.BufferData(gl.ARRAY_BUFFER, len(b.data), gl.Ptr(b.data), gl.STATIC_DRAW)
    gl.BindBuffer(gl.ARRAY_BUFFER, 0)
  checkGLError()

    b.dirty = false
  }
}

func (b *Float32Buffer) bind() {
  checkGLError()
  gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
  gl.VertexAttribPointer(b.loc, int32(b.nComp), gl.FLOAT, false, 0, nil)
  gl.EnableVertexAttribArray(b.loc)
  checkGLError()
}

func (b *UInt8Buffer) bind() {
  checkGLError()
  gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
  gl.VertexAttribPointer(b.loc, int32(b.nComp), gl.UNSIGNED_BYTE, false, 0, nil)
  gl.EnableVertexAttribArray(b.loc)
  checkGLError()
}

func (d *DrawPassData) SetPos(triId uint32, vertexId uint32, x_ int, y_ int, z float32) {
  x := 2.0*float32(x_)/float32(d.w) - 1.0
  y := 1.0 - 2.0*float32(y_)/float32(d.h)

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
  dx := 2.0*float32(dx_)/float32(d.w)
  dy := -2.0*float32(dy_)/float32(d.h)

  d.translatePos(triId, vertexId, dx, dy, dz)
}

func (d *DrawPassData) TranslateTri(triId uint32, dx_ int, dy_ int, dz float32) {
  dx := 2.0*float32(dx_)/float32(d.w)
  dy := -2.0*float32(dy_)/float32(d.h)

  d.translatePos(triId, 0, dx, dy, dz)
  d.translatePos(triId, 1, dx, dy, dz)
  d.translatePos(triId, 2, dx, dy, dz)
}

func (d *DrawPassData) SetPosF(triId uint32, vertexId uint32, x_ float64, y_ float64, z float32) {
  x := 2.0*float32(x_)/float32(d.w) - 1.0
  y := 1.0 - 2.0*float32(y_)/float32(d.h)

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
  x := float32(x_)/float32(d.Skin.width)
  y := float32(y_)/float32(d.Skin.height)

  // XXX: transpose for some reason
  d.TCoord.Set2(triId, vertexId, y, x)
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

func (d *DrawPass2Data) SetGlyphCoords(triId0, triId1 uint32, name string) {
  g := d.Glyphs.GetGlyph(name)

  k := g.TexId
  
  r := d.Glyphs.GetRect(k)

  d.SetGlyphCoord(triId0, 0, r.X, r.Y)
  d.SetGlyphCoord(triId0, 1, r.Right(), r.Y)
  d.SetGlyphCoord(triId0, 2, r.X, r.Bottom())

  d.SetGlyphCoord(triId1, 0, r.Right(), r.Bottom())
  d.SetGlyphCoord(triId1, 1, r.Right(), r.Y)
  d.SetGlyphCoord(triId1, 2, r.X, r.Bottom())
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
  d.DrawPassData.SyncAndBind()

  d.Skin.bind()
}

func (d *DrawPass2Data) SyncAndBind() {
  d.DrawPassData.SyncAndBind()

  d.Glyphs.bind()
}

func (d *DrawPassData) dirty() bool {
  return d.Pos.dirty || d.Type.dirty || d.Param.dirty || d.TCoord.dirty || d.Color.dirty
}

func (d *DrawPassData) posDirty() bool {
  return d.Type.dirty // a change of type indicates that some elements became visible/hidden, and thus affect positioning of siblings etc.
}

func (d *DrawPassData) clearPosDirty() {
  d.Type.dirty = false
}
