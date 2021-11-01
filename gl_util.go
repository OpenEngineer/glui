package glui

import (
  "strconv"

  "github.com/go-gl/gl/v4.1-core/gl"
)

func checkGLError() {
  errNum := gl.GetError()
  if errNum != gl.NO_ERROR {
    switch errNum {
    case gl.INVALID_ENUM:
      panic("gl.INVALID_ENUM")
    case gl.INVALID_VALUE:
      panic("gl.INVALID_VALUE")
    case gl.INVALID_OPERATION:
      panic("gl.INVALID_OPERATION")
    case gl.STACK_OVERFLOW:
      panic("gl.STACK_OVERFLOW")
    case gl.STACK_UNDERFLOW:
      panic("gl.STACK_UNDERFLOW")
    case gl.OUT_OF_MEMORY:
      panic("gl.OUT_OF_MEMORY")
    }
  }
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

func setupFloatVAO(loc uint32, vao uint32, nComp int32) {
  //gl.BindVertexArray(vao)

  //gl.EnableVertexArrayAttrib(vao, loc)

  //checkGLError()

  //gl.VertexAttribPointer(loc, nComp, gl.FLOAT, false, 0, nil)

  //checkGLError()

  //gl.DisableVertexAttribArray(loc)

  //checkGLError()
}

func setupRenderToTextureFBO(w, h int, fbo, texID, drawBuf uint32) {
  gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
  gl.Viewport(0, 0, int32(w), int32(h))

  checkGLError()

  gl.BindTexture(gl.TEXTURE_2D, texID)
  gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, int32(w), int32(h), 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
  gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
  gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
  gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
  gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

  gl.NamedFramebufferTexture(fbo, drawBuf, texID, 0)
  gl.DrawBuffers(1, &drawBuf)

  if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
    panic("something went wrong")
  }
  checkGLError()
}
