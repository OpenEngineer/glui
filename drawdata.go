package glui

import (
  "github.com/veandco/go-sdl2/sdl"
  "github.com/go-gl/gl/v4.1-core/gl"
)

const (
  N_INIT_TRIS     = 2
  GROW_FACTOR     = 2.0

  VTYPE_HIDDEN = 0
  VTYPE_PLAIN  = 1
  VTYPE_SKIN   = 2
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

type DrawData struct {
  free []uint32

  W int
  H int

  Pos    *Float32Buffer
  Type   *Float32Buffer
  Color  *Float32Buffer
  TCoord *Float32Buffer

  Skin   *SkinMap
}

func NewFloat32Buffer(nComp int) *Float32Buffer {
  b := &Float32Buffer{nComp, 0, 0, 0, make([]float32, N_INIT_TRIS*3*nComp), true}

  return b
}

func (b *Float32Buffer) InitGL(location uint32) {
  b.loc = location

  gl.GenBuffers(1, &b.vbo)
  gl.GenVertexArrays(1, &b.vao)
  gl.BindVertexArray(b.vao)
  gl.EnableVertexAttribArray(b.loc)
  gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
  gl.VertexAttribPointer(b.loc, int32(b.nComp), gl.FLOAT, false, 0, nil)
  gl.BindBuffer(gl.ARRAY_BUFFER, 0)
  gl.BindVertexArray(0)

  b.dirty = true
}

func NewUInt8Buffer(nComp int) *UInt8Buffer {
  b := &UInt8Buffer{nComp, 0, 0, 0, make([]uint8, N_INIT_TRIS*3*nComp), true}

  return b
}

func (b *UInt8Buffer) InitGL(location uint32) {
  b.loc = location

  gl.GenBuffers(1, &b.vbo)
  gl.GenVertexArrays(1, &b.vao)
  gl.BindVertexArray(b.vao)
  gl.EnableVertexAttribArray(b.loc)
  gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
  gl.VertexAttribPointer(b.loc, int32(b.nComp), gl.UNSIGNED_BYTE, false, 0, nil)
  gl.BindBuffer(gl.ARRAY_BUFFER, 0)
  gl.BindVertexArray(0)

  b.dirty = true
}

func NewDrawData(s Skin) *DrawData {
  free := make([]uint32, N_INIT_TRIS)
  for i := 0; i < N_INIT_TRIS; i++ {
    free[i] = uint32(i)
  }

  d := &DrawData{
    free,
    0,
    0,
    NewFloat32Buffer(3),
    NewFloat32Buffer(1),
    NewFloat32Buffer(4),
    NewFloat32Buffer(2),
    NewSkinMap(s),
  }

  // set all types to zero
  for i := 0; i < len(d.Type.data); i++ {
    d.Type.data[i] = VTYPE_HIDDEN
  }

  return d
}

func (d *DrawData) InitGL(prog uint32) {
  posLoc := gl.GetAttribLocation(prog, gl.Str("aPos\x00"))
  typeLoc := gl.GetAttribLocation(prog, gl.Str("aType\x00"))
  colorLoc := gl.GetAttribLocation(prog, gl.Str("aColor\x00"))
  tcoordLoc := gl.GetAttribLocation(prog, gl.Str("aTCoord\x00"))
  skinLoc := gl.GetUniformLocation(prog, gl.Str("skin\x00"))

  d.Pos.InitGL(uint32(posLoc))
  d.Type.InitGL(uint32(typeLoc))
  d.Color.InitGL(uint32(colorLoc))
  d.TCoord.InitGL(uint32(tcoordLoc))

  d.Skin.InitGL(uint32(skinLoc))
}

// number of tris
func (d *DrawData) Len() int {
  return len(d.Type.data)/3
}

func (d *DrawData) Grow() {
  nTrisOld := d.Len()
  nTrisNew := int(float64(nTrisOld)*GROW_FACTOR)

  oldNFree := len(d.free)

  d.free = append(d.free, make([]uint32, nTrisNew - nTrisOld)...)

  for i := 0; i < nTrisNew - nTrisOld; i++ {
    d.free[oldNFree + i] = uint32(nTrisOld + i)
  }

  d.Pos.grow(nTrisNew)
  d.Type.grow(nTrisNew)
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
    gl.DeleteBuffers(1, &b.vbo)
    gl.GenBuffers(1, &b.vbo)
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
    gl.DeleteBuffers(1, &b.vbo)
    gl.GenBuffers(1, &b.vbo)
  }

  b.dirty = true
}

func (d *DrawData) Alloc(nTris int) []uint32 {
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

func (b *DrawData) Dealloc(offsets []uint32) {
  oldFree := b.free

  b.free = make([]uint32, len(oldFree) + len(offsets))

  i := 0
  j := 0
  k := 0

  for ;i < len(oldFree) && j < len(offsets); {
    if i == len(oldFree) {
      b.free[k] = offsets[j]
      j++
      k++
    } else if j == len(offsets) {
      b.free[k] = oldFree[i]
      i++
      k++
    } else if oldFree[i] < offsets[j] {
      b.free[k] = oldFree[i]
      i++
      k++
    } else {
      b.free[k] = offsets[j]
      j++
      k++
    }
  }

  if k != len(b.free) {
    panic("k should be == len(b.free)")
  }
}

func (b *Float32Buffer) Set(triId uint32, vertexId uint32, compId uint32, value float32) {
  offset := (triId*3 + vertexId)*uint32(b.nComp)

  b.data[offset + compId] = value

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
  offset := triId*3

  b.data[(offset + 0)*uint32(b.nComp) + 0] = value0
  b.data[(offset + 0)*uint32(b.nComp) + 1] = value1
  b.data[(offset + 0)*uint32(b.nComp) + 2] = value2
  b.data[(offset + 0)*uint32(b.nComp) + 3] = value3

  b.data[(offset + 1)*uint32(b.nComp) + 0] = value0
  b.data[(offset + 1)*uint32(b.nComp) + 1] = value1
  b.data[(offset + 1)*uint32(b.nComp) + 2] = value2
  b.data[(offset + 1)*uint32(b.nComp) + 3] = value3

  b.data[(offset + 2)*uint32(b.nComp) + 0] = value0
  b.data[(offset + 2)*uint32(b.nComp) + 1] = value1
  b.data[(offset + 2)*uint32(b.nComp) + 2] = value2
  b.data[(offset + 2)*uint32(b.nComp) + 3] = value3

  b.dirty = true
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
    gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
    gl.BufferData(gl.ARRAY_BUFFER, 4*len(b.data), gl.Ptr(b.data), gl.STATIC_DRAW)
    gl.BindBuffer(gl.ARRAY_BUFFER, 0)

    b.dirty = false
  }
}

func (b *UInt8Buffer) sync() {
  if b.dirty {
    gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
    gl.BufferData(gl.ARRAY_BUFFER, len(b.data), gl.Ptr(b.data), gl.STATIC_DRAW)
    gl.BindBuffer(gl.ARRAY_BUFFER, 0)

    b.dirty = false
  }
}

func (b *Float32Buffer) bind() {
  //gl.BindVertexArray(b.vao)
  gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
  gl.VertexAttribPointer(b.loc, int32(b.nComp), gl.FLOAT, false, 0, nil)
  gl.EnableVertexAttribArray(b.loc)
}

func (b *UInt8Buffer) bind() {
  //gl.BindVertexArray(b.vao)
  gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
  gl.VertexAttribPointer(b.loc, int32(b.nComp), gl.UNSIGNED_BYTE, false, 0, nil)
  gl.EnableVertexAttribArray(b.loc)
}

func (d *DrawData) SetPos(triId uint32, vertexId uint32, x_ int, y_ int, z float32) {
  x := 2.0*float32(x_)/float32(d.W) - 1.0
  y := 1.0 - 2.0*float32(y_)/float32(d.H)

  d.Pos.Set3(triId, vertexId, x, y, z)
}

func (d *DrawData) SetSkinCoord(triId uint32, vertexId uint32, x_ int, y_ int) {
  x := float32(x_)/float32(d.Skin.width)
  y := float32(y_)/float32(d.Skin.height)

  // XXX: transpose for some reason
  d.TCoord.Set2(triId, vertexId, y, x)
}

func (d *DrawData) SetColorConst(triId uint32, c sdl.Color) {
  r := float32(c.R)/float32(256)
  g := float32(c.G)/float32(256)
  b := float32(c.B)/float32(256)
  a := float32(c.A)/float32(256)

  d.Color.Set4Const(triId, r, g, b, a)
}

func (d *DrawData) SyncAndBind() {
  d.Pos.sync()
  d.Type.sync()
  d.Color.sync()
  d.TCoord.sync()
  // skin is synced upon init

  d.Pos.bind()
  d.Type.bind()
  d.Color.bind()
  d.TCoord.bind()
  /*gl.BindVertexArray(d.Pos.vao)
  gl.BindVertexArray(d.Type.vao)
  gl.BindVertexArray(d.Color.vao)
  gl.BindVertexArray(d.TCoord.vao)*/

  d.Skin.bind()
}

func (d *DrawData) GetSize() (int, int) {
  return d.W, d.H
}

func (d *DrawData) SyncSize(window *sdl.Window) {
  w, h := window.GetSize()
  d.W, d.H = int(w), int(h)
}

func (d *DrawData) Dirty() bool {
  return d.Pos.dirty || d.Type.dirty || d.TCoord.dirty || d.Color.dirty
}
