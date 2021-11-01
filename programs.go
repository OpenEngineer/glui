package glui

import (
  "fmt"
  "os"

  "github.com/go-gl/gl/v4.1-core/gl"
)

// a collection of gl ptrs
type Programs struct {
  fbos         [2]uint32 // framebuffers 
  fbo_texUnits [2]uint32 // {gl.TEXTURE0, gl.TEXTURE1}
  fbo_drawBufs [2]uint32 // {gl.COLOR_ATTACHMENT0, gl.COLOR_ATTACHMENT1}

  skinPass  uint32
  glyphPass uint32
  blurPass  uint32

  skinPass_aPosLoc    uint32
  skinPass_aTypeLoc   uint32
  skinPass_aParamLoc  uint32
  skinPass_aColorLoc  uint32
  skinPass_aTCoordLoc uint32
  skinPass_uTexLoc    uint32
  skinPass_texID      uint32
  skinPass_texUnit    uint32
  skinPass_aPosVAO     uint32
  skinPass_aTypeVAO    uint32
  skinPass_aParamVAO   uint32
  skinPass_aColorVAO   uint32
  skinPass_aTCoordVAO  uint32
  skinPass_aPosVBO     uint32
  skinPass_aTypeVBO    uint32
  skinPass_aParamVBO   uint32
  skinPass_aColorVBO   uint32
  skinPass_aTCoordVBO  uint32

  glyphPass_aPosLoc    uint32
  glyphPass_aTypeLoc   uint32
  glyphPass_aParamLoc  uint32
  glyphPass_aColorLoc  uint32
  glyphPass_aTCoordLoc uint32
  glyphPass_uTexLoc    uint32
  glyphPass_texID      uint32
  glyphPass_texUnit    uint32
  glyphPass_aPosVAO     uint32
  glyphPass_aTypeVAO    uint32
  glyphPass_aParamVAO   uint32
  glyphPass_aColorVAO   uint32
  glyphPass_aTCoordVAO  uint32
  glyphPass_aPosVBO     uint32
  glyphPass_aTypeVBO    uint32
  glyphPass_aParamVBO   uint32
  glyphPass_aColorVBO   uint32
  glyphPass_aTCoordVBO  uint32

  blurPass_aCoordLoc   uint32
  blurPass_uSizeLoc    uint32
  blurPass_uDirLoc     uint32
  blurPass_uTexLoc     uint32
  blurPass_aCoordVAO   uint32
  blurPass_aCoordVBO   uint32
}

func (p *Programs) initPtrs(debug *os.File) {
  gl.GenFramebuffers(2, &(p.fbos[0]))
  for _, fbo := range p.fbos {
    if fbo <= 0 {
      panic("bad framebuffer from GenFramebuffers")
    }
  }

  p.fbo_texUnits = [2]uint32{gl.TEXTURE0, gl.TEXTURE1}
  p.fbo_drawBufs = [2]uint32{gl.COLOR_ATTACHMENT0, gl.COLOR_ATTACHMENT1}

  var err error 

  p.skinPass, err = compileSkinPass()
  if err != nil {
    fmt.Fprintf(debug, "failed to compile OpenGL skinPass program: %s\n", err.Error())
    panic(err)
  }

  p.glyphPass, err = compileGlyphPass()
  if err != nil {
    fmt.Fprintf(debug, "failed to compile OpenGL glyphPass program: %s\n", err.Error())
    panic(err)
  }

  p.blurPass, err = compileBlurPass()
  if err != nil {
    fmt.Fprintf(debug, "failed to compile OpenGL blurPass program: %s\n", err.Error())
    panic(err)
  }

  texIDs := [2]uint32{0, 0}
  gl.GenTextures(2, &(texIDs[0]))

  for _, texID := range texIDs {
    if texID < 0 {
      panic("negative texID")
    }
  }

  checkGLError()

  vaos := make([]uint32, 11)
  gl.GenVertexArrays(int32(len(vaos)), &(vaos[0]))
  for _, vao := range vaos {
    if vao < 0 {
      panic("negative vao")
    }
  }

  checkGLError()

  vbos := make([]uint32, 11)
  gl.GenBuffers(int32(len(vbos)), &(vbos[0]))
  for _, vbo := range vbos {
    if vbo <= 0 {
      panic("negative vbo")
    }
  }

  p.skinPass_aPosLoc    = getGLAttribLocation(p.skinPass, "aPos")
  p.skinPass_aTypeLoc   = getGLAttribLocation(p.skinPass, "aType")
  p.skinPass_aParamLoc  = getGLAttribLocation(p.skinPass, "aParam")
  p.skinPass_aColorLoc  = getGLAttribLocation(p.skinPass, "aColor")
  p.skinPass_aTCoordLoc = getGLAttribLocation(p.skinPass, "aTCoord")
  p.skinPass_uTexLoc    = getGLUniformLocation(p.skinPass, "uTex")
  p.skinPass_texID      = texIDs[0]
  p.skinPass_texUnit    = gl.TEXTURE0
  p.skinPass_aPosVAO    = vaos[0]
  p.skinPass_aTypeVAO   = vaos[1]
  p.skinPass_aParamVAO  = vaos[2]
  p.skinPass_aColorVAO  = vaos[3]
  p.skinPass_aTCoordVAO = vaos[4]
  p.skinPass_aPosVBO    = vbos[0]
  p.skinPass_aTypeVBO   = vbos[1]
  p.skinPass_aParamVBO  = vbos[2]
  p.skinPass_aColorVBO  = vbos[3]
  p.skinPass_aTCoordVBO = vbos[4]

  p.glyphPass_aPosLoc    = getGLAttribLocation(p.glyphPass, "aPos")
  p.glyphPass_aTypeLoc   = getGLAttribLocation(p.glyphPass, "aType")
  p.glyphPass_aParamLoc  = getGLAttribLocation(p.glyphPass, "aParam")
  p.glyphPass_aColorLoc  = getGLAttribLocation(p.glyphPass, "aColor")
  p.glyphPass_aTCoordLoc = getGLAttribLocation(p.glyphPass, "aTCoord")
  p.glyphPass_uTexLoc    = getGLUniformLocation(p.glyphPass, "uTex")
  p.glyphPass_texID      = texIDs[1]
  p.glyphPass_texUnit    = gl.TEXTURE0
  p.glyphPass_aPosVAO    = vaos[5]
  p.glyphPass_aTypeVAO   = vaos[6]
  p.glyphPass_aParamVAO  = vaos[7]
  p.glyphPass_aColorVAO  = vaos[8]
  p.glyphPass_aTCoordVAO = vaos[9]
  p.glyphPass_aPosVBO    = vbos[5]
  p.glyphPass_aTypeVBO   = vbos[6]
  p.glyphPass_aParamVBO  = vbos[7]
  p.glyphPass_aColorVBO  = vbos[8]
  p.glyphPass_aTCoordVBO = vbos[9]

  p.blurPass_aCoordLoc = getGLAttribLocation(p.blurPass, "aCoord")
  p.blurPass_uSizeLoc  = getGLUniformLocation(p.blurPass, "uSize")
  p.blurPass_uDirLoc   = getGLUniformLocation(p.blurPass, "uDir")
  p.blurPass_uTexLoc   = getGLUniformLocation(p.blurPass, "uTex")
  p.blurPass_aCoordVAO = vaos[10]
  p.blurPass_aCoordVBO = vbos[10]

  checkGLError()

  // set up the attrib structures and the texture locations

  gl.UseProgram(p.skinPass)
  setupFloatVAO(p.skinPass_aPosLoc,    p.skinPass_aPosVAO, 3)
  setupFloatVAO(p.skinPass_aTypeLoc,   p.skinPass_aTypeVAO, 1)
  setupFloatVAO(p.skinPass_aParamLoc,  p.skinPass_aParamVAO, 1)
  setupFloatVAO(p.skinPass_aColorLoc,  p.skinPass_aColorVAO, 4)
  setupFloatVAO(p.skinPass_aTCoordLoc, p.skinPass_aTCoordVAO, 2)
  gl.Uniform1i(int32(p.skinPass_uTexLoc), int32(p.skinPass_texUnit - gl.TEXTURE0))

  checkGLError()

  gl.UseProgram(p.glyphPass)
  setupFloatVAO(p.glyphPass_aPosLoc,    p.glyphPass_aPosVAO, 3)
  setupFloatVAO(p.glyphPass_aTypeLoc,   p.glyphPass_aTypeVAO, 1)
  setupFloatVAO(p.glyphPass_aParamLoc,  p.glyphPass_aParamVAO, 1)
  setupFloatVAO(p.glyphPass_aColorLoc,  p.glyphPass_aColorVAO, 4)
  setupFloatVAO(p.glyphPass_aTCoordLoc, p.glyphPass_aTCoordVAO, 2)
  gl.Uniform1i(int32(p.glyphPass_uTexLoc), int32(p.glyphPass_texUnit - gl.TEXTURE0))

  checkGLError()

  // the uTexLoc of the blurPass is different every draw call, so can't be set here
  gl.UseProgram(p.blurPass)
  setupFloatVAO(p.blurPass_aCoordLoc, p.blurPass_aCoordVAO, 2)
}
