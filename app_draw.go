package glui

import (
  "errors"
  "fmt"
  "sync"
  "unsafe"

  "github.com/go-gl/gl/v4.1-core/gl"
  "github.com/veandco/go-sdl2/sdl"
)

func (app *App) DrawIfDirty() {
  if app.root.posDirty() {
    fmt.Println("recalculating...")

    app.root.CalcDepth()

    app.root.CalcPos()

    if app.mouseInWindow() {
      app.updateMouseElement(-1, -1, 0, 0)
    }
  }

  if app.root.dirty() {
    fmt.Println("redrawing...", app.root.P1.nTris(), "&", app.root.P2.nTris())

    app.draw()
  }
}

// can be called from any thread
func (app *App) Draw() {
  app.drawCh <- true
}

func (app *App) draw() {
  if err := app.window.GLMakeCurrent(app.ctx); err != nil {
    fmt.Fprintf(app.debug, "unable to make current: %s\n", err.Error())
    return
  }

  w, h := app.root.GetSize()

  gl.Viewport(0, 0, int32(w), int32(h))

  app.drawInner()

  app.window.GLSwap()

  if err := app.window.GLMakeCurrent(nil); err != nil {
    fmt.Fprintf(app.debug, "unable to unmake current: %s\n", err.Error())
    return
  }

  if err := OnAfterDrawOS(app); err != nil {
    fmt.Fprintf(app.debug, "unable to run OnAfterDraw: %s\n", err.Error())
    return
  }
}

func (app *App) drawInner() {
  color := app.root.P1.Skin.BGColor()

  gl.ClearColor(
    float32(color.R)/float32(256),
    float32(color.G)/float32(256),
    float32(color.B)/float32(256),
    float32(color.A)/float32(256),
  )

  gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

  gl.Enable(gl.BLEND)
  gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA)

  gl.Enable(gl.DEPTH_TEST)
  gl.DepthFunc(gl.LESS)

  gl.UseProgram(app.program1)

  app.root.P1.SyncAndBind()

  gl.DrawArrays(gl.TRIANGLES, 0, int32(app.root.P1.Len())*3)

  gl.UseProgram(app.program2)

  app.root.P2.SyncAndBind()

  gl.DrawArrays(gl.TRIANGLES, 0, int32(app.root.P2.Len())*3)
}

func (app *App) drawThumbnail(w int, h int, dst unsafe.Pointer) {
  if err := app.window.GLMakeCurrent(app.ctx); err != nil {
    fmt.Fprintf(app.debug, "unable to make current: %s\n", err.Error())
    return
  }
  gl.BindFramebuffer(gl.FRAMEBUFFER, app.framebuffers[0])

  wWin, hWin := app.window.GetSize()

  gl.Viewport(0, 0, wWin, hWin)

  app.drawInner()

  // create the thumbnail
  gl.BindFramebuffer(gl.READ_FRAMEBUFFER, app.framebuffers[0])
  gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, app.framebuffers[1])
  gl.BlitFramebuffer(0, 0, wWin, hWin, 0, 0, int32(w), int32(h), gl.COLOR_BUFFER_BIT, gl.LINEAR)
  gl.BindFramebuffer(gl.READ_FRAMEBUFFER, app.framebuffers[1])
  gl.ReadPixels(0, 0, int32(w), int32(h), gl.RGBA, gl.UNSIGNED_BYTE, dst)

  // take this opportunity to also write to the screen
  gl.BindFramebuffer(gl.READ_FRAMEBUFFER, app.framebuffers[0])
  gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
  gl.BlitFramebuffer(0, 0, wWin, hWin, 0, 0, wWin, hWin, gl.COLOR_BUFFER_BIT, gl.NEAREST)

  // make sure default framebuffer is also the read framebuffer
  gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

  app.window.GLSwap()

  if err := app.window.GLMakeCurrent(nil); err != nil {
    fmt.Fprintf(app.debug, "unable to unmake current: %s\n", err.Error())
    return
  }
}

// eg. apply gaussian blur
func (app *App) drawFiltered(program uint32) {
  if err := app.window.GLMakeCurrent(app.ctx); err != nil {
    fmt.Fprintf(app.debug, "unable to make current: %s\n", err.Error())
    return
  }


  w, h := app.root.GetSize()

  gl.Viewport(0, 0, int32(w), int32(h))

  // do an inner draw to framebuffer 0
  fb := app.framebuffers[0]
  gl.BindFramebuffer(gl.FRAMEBUFFER, fb)

  // create the texture
  var texture uint32
  gl.GenTextures(1, &texture) // TODO: save this texture
  gl.BindTexture(gl.TEXTURE_2D, texture)
  gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, int32(w), int32(h), 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
  gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
  gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

  // attach the framebuffer to the texture (in order to avoid blitting this must be done before
  //gl.BindFramebuffer(gl.READ_FRAMEBUFFER, fb)
  var colorAttach uint32 = gl.COLOR_ATTACHMENT0
  gl.BindFramebuffer(gl.FRAMEBUFFER, fb)
  gl.FramebufferTexture(gl.FRAMEBUFFER, colorAttach, texture, 0)
  gl.DrawBuffers(1, &colorAttach)
  //gl.DrawBuffer(colorAttach)

  if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
    fmt.Println(gl.CheckFramebufferStatus(gl.FRAMEBUFFER))
    panic("something went wrong")
  }

  // do the inner draw
  gl.BindFramebuffer(gl.FRAMEBUFFER, fb)
  gl.Viewport(0, 0, int32(w), int32(h))
  //gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
  app.drawInner()


  gl.UseProgram(program)
  //gl.DrawBuffer(gl.FRONT)
  gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
  gl.Viewport(0, 0, int32(w), int32(h))
  gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

  // buffer the tri data for the filter program
  var vbo uint32
  loc := uint32(gl.GetAttribLocation(program, gl.Str("aCoord\x00")))
  gl.GenBuffers(1, &vbo)
  gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
  gl.VertexAttribPointer(loc, 2, gl.FLOAT, false, 0, nil)
  gl.EnableVertexAttribArray(loc)
  data := []float32{
    0.0, 0.0,
    1.0, 0.0,
    0.0, 1.0,
    1.0, 0.0,
    1.0, 1.0,
    0.0, 1.0,
  }
  gl.BufferData(gl.ARRAY_BUFFER, 4*len(data), gl.Ptr(data), gl.STATIC_DRAW)

  // bind texture to program location
  gl.ActiveTexture(gl.TEXTURE0)
  gl.BindTexture(gl.TEXTURE_2D, texture)
  texLoc := gl.GetUniformLocation(program, gl.Str("frame\x00"))
  gl.Uniform1i(int32(texLoc), 0) // 1 -> gl.TEXTURE1
  
  gl.DrawArrays(gl.TRIANGLES, 0, 6) // 2 triangles with 3 vertices each

  // clean up
  gl.BindBuffer(gl.ARRAY_BUFFER, 0)
  gl.DisableVertexAttribArray(loc)
  gl.BindTexture(gl.TEXTURE_2D, 0)
  gl.DeleteTextures(1, &texture)
  gl.DeleteBuffers(1, &vbo)
  gl.ActiveTexture(gl.TEXTURE0)

  // make sure the default framebuffer is also the read framebuffer
  gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

  app.window.GLSwap()

  // now bind the framebuffer to a texture so we can apply the filtering
  if err := app.window.GLMakeCurrent(nil); err != nil {
    fmt.Fprintf(app.debug, "unable to unmake current: %s\n", err.Error())
    return
  }
}

func (app *App) DrawBlurred() {
  app.drawFiltered(app.programGaussBlur)

  sdl.Delay(1000)
}

func (app *App) drawAndCopyToBitmap(w int, h int, dst unsafe.Pointer) {
  if err := app.window.GLMakeCurrent(app.ctx); err != nil {
    fmt.Fprintf(app.debug, "unable to make current: %s\n", err.Error())
    return
  }
  gl.BindFramebuffer(gl.FRAMEBUFFER, app.framebuffers[0])

  gl.Viewport(0, 0, int32(w), int32(h))

  app.drawInner()

  // copy the pixels into a bitmap
  gl.BindFramebuffer(gl.READ_FRAMEBUFFER, app.framebuffers[0])
  gl.ReadPixels(0, 0, int32(w), int32(h), gl.RGBA, gl.UNSIGNED_BYTE, dst)

  // take this opportunity to also write to the screen
  gl.BindFramebuffer(gl.READ_FRAMEBUFFER, app.framebuffers[0])
  gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
  gl.BlitFramebuffer(0, 0, int32(w), int32(h), 0, 0, int32(w), int32(h), gl.COLOR_BUFFER_BIT, gl.NEAREST)

  // make sure default framebuffer is also the read framebuffer
  gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

  app.window.GLSwap()

  if err := app.window.GLMakeCurrent(nil); err != nil {
    fmt.Fprintf(app.debug, "unable to unmake current: %s\n", err.Error())
    return
  }
}

func (app *App) initDrawLoop(m *sync.Mutex) {
  m.Lock()

  ctx, err := app.window.GLCreateContext()
  if err != nil {
    fmt.Fprintf(app.debug, "unable to create context: %s\n", err.Error())
    panic(err)
  }

  app.ctx = ctx

  if err := app.window.GLMakeCurrent(ctx); err != nil {
    fmt.Fprintf(app.debug, "unable to make current in render: %s\n", err.Error())
    panic(err)
  }

  if err := gl.Init(); err != nil {
    fmt.Fprintf(app.debug, "render gl.Init error: %s\n", err.Error())
    panic(err)
  }

  glVersion := gl.GoStr(gl.GetString(gl.VERSION))
  if glVersion == "" {
    err := errors.New("empty OpenGL version")
    fmt.Fprintf(app.debug, "%s\n", err.Error())
    panic(err)
  }

  app.program1, err = compileProgram1()
  if err != nil {
    fmt.Fprintf(app.debug, "failed to compile OpenGL program1: %s\n", err.Error())
    panic(err)
  }

  app.program2, err = compileProgram2()
  if err != nil {
    fmt.Fprintf(app.debug, "failed to compile OpenGL program2: %s\n", err.Error())
    panic(err)
  }

  app.programGaussBlur, err = compileProgramGaussBlur()
  if err != nil {
    fmt.Fprintf(app.debug, "failed to compile OpenGL program2: %s\n", err.Error())
    panic(err)
  }

  app.root.syncSize(app.window)
  app.root.initGL(app.program1, app.program2)

  //gl.CreateFramebuffers(1, &(app.framebuffers[0]))
  //gl.CreateFramebuffers(1, &(app.framebuffers[1]))

  gl.GenFramebuffers(1, &(app.framebuffers[0]))
  gl.GenFramebuffers(1, &(app.framebuffers[1]))

  x, y := app.window.GetPosition()
  app.x = int(x)
  app.y = int(y)

  if err := app.window.GLMakeCurrent(nil); err != nil {
    fmt.Fprintf(app.debug, "unable to unmake current in render: %s\n", err.Error())
    return
  }

  m.Unlock()
}

func (app *App) endDrawLoop() {
  sdl.GLDeleteContext(app.ctx)
}
