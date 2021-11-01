package glui

import (
  "errors"
  "fmt"
  "sync"

  "github.com/go-gl/gl/v4.1-core/gl"
  "github.com/veandco/go-sdl2/sdl"
)

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

  checkGLError()

  glVersion := gl.GoStr(gl.GetString(gl.VERSION))
  if glVersion == "" {
    err := errors.New("empty OpenGL version")
    fmt.Fprintf(app.debug, "%s\n", err.Error())
    panic(err)
  }

  app.programs.initPtrs(app.debug)

  app.skinMap.initGL(
    app.programs.skinPass_uTexLoc, 
    app.programs.skinPass_texID, 
    app.programs.skinPass_texUnit,
  )

  app.glyphMap.initGL(
    app.programs.glyphPass_uTexLoc,
    app.programs.glyphPass_texID,
    app.programs.glyphPass_texUnit,
  )

  for _, frame := range app.frames {
    frame.P1.initGL(
      app.programs.skinPass_aPosLoc,
      app.programs.skinPass_aTypeLoc,
      app.programs.skinPass_aParamLoc,
      app.programs.skinPass_aColorLoc,
      app.programs.skinPass_aTCoordLoc,
      app.programs.skinPass_aPosVAO,
      app.programs.skinPass_aTypeVAO,
      app.programs.skinPass_aParamVAO,
      app.programs.skinPass_aColorVAO,
      app.programs.skinPass_aTCoordVAO,
      app.programs.skinPass_aPosVBO,
      app.programs.skinPass_aTypeVBO,
      app.programs.skinPass_aParamVBO,
      app.programs.skinPass_aColorVBO,
      app.programs.skinPass_aTCoordVBO,
    )

    frame.P2.initGL(
      app.programs.glyphPass_aPosLoc,
      app.programs.glyphPass_aTypeLoc,
      app.programs.glyphPass_aParamLoc,
      app.programs.glyphPass_aColorLoc,
      app.programs.glyphPass_aTCoordLoc,
      app.programs.glyphPass_aPosVAO,
      app.programs.glyphPass_aTypeVAO,
      app.programs.glyphPass_aParamVAO,
      app.programs.glyphPass_aColorVAO,
      app.programs.glyphPass_aTCoordVAO,
      app.programs.glyphPass_aPosVBO,
      app.programs.glyphPass_aTypeVBO,
      app.programs.glyphPass_aParamVBO,
      app.programs.glyphPass_aColorVBO,
      app.programs.glyphPass_aTCoordVBO,
    )
  }

  app.syncWindowSize()

  checkGLError()

  x, y := app.window.GetPosition()
  app.x = int(x)
  app.y = int(y)

  if err := app.window.GLMakeCurrent(nil); err != nil {
    fmt.Fprintf(app.debug, "unable to unmake current in render: %s\n", err.Error())
    return
  }

  m.Unlock()
  checkGLError()
}

func (app *App) DrawIfDirty() {
  anyDirty := false
  for i, frame := range app.frames {
    if i > app.activeFrame {
      break
    }

    if frame.posDirty() {
      frame.CalcDepth()

      frame.CalcPos()

      // TODO: hwo does this work for higher frames?
      if app.mouseInWindow() {
        app.updateMouseElement(-1, -1, 0, 0)
      }
    }

    if frame.dirty() {
      anyDirty = true
    }
  }

  if anyDirty {
    app.draw()
  }
}

// can be called from any thread. Doesn't block
func (app *App) Draw() {
  go func() {
    app.drawCh <- true
  }()
}

func (app *App) draw() {
  if err := app.window.GLMakeCurrent(app.ctx); err != nil {
    fmt.Fprintf(app.debug, "unable to make current: %s\n", err.Error())
    return
  }

  a := 0

  for i, frame := range app.frames {
    if i < app.activeFrame {
      b := a + 1
      if b == 2 {
        b = 0
      }

      app.renderToTexture(a)
      if app.activeFrame > 0 {
        frame.ForceAllDirty() // all tris must be uploaded
      }
      app.drawFrame(frame)

      app.renderToTexture(b)
      app.blur(a, b, 1.0, 0.0)

      if (i < app.activeFrame - 1) {
        app.renderToTexture(a)
        app.blur(b, a, 0.0, 1.0)
      } else {
        gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
        app.blur(b, -1, 0.0, 1.0)
      }

      a = a + 1
      if a == 2 {
        a = 0
      }
    } else {
      gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

      if app.activeFrame > 0 {
        frame.ForceAllDirty() // all tris must be uploaded
      }
      app.drawFrame(frame)

      break
    }

  }

  app.window.GLSwap()

  if err := app.window.GLMakeCurrent(nil); err != nil {
    fmt.Fprintf(app.debug, "unable to unmake current: %s\n", err.Error())
    return
  }

  if err := OnAfterDrawOS(app); err != nil {
    fmt.Fprintf(app.debug, "unable to run OnAfterDraw: %s\n", err.Error())
    return
  }

  checkGLError()
}

func (app *App) renderToTexture(fboI int) {
  winW, winH := app.getWindowSize()

  setupRenderToTextureFBO(
    winW, 
    winH, 
    app.programs.fbos[fboI],
    app.programs.fbo_texIDs[fboI],
    app.programs.fbo_drawBufs[fboI],
  )
}

func (app *App) drawFrame(frame *Frame) {
  winW, winH := app.getWindowSize()

  gl.Viewport(0, 0, int32(winW), int32(winH))

  x, y := frame.GetPos()
  w, h := frame.GetSize()

  applyScissor := w < winW || h < winH || x > 0 || y > 0

  if applyScissor {
    gl.Enable(gl.SCISSOR_TEST)

    gl.Scissor(int32(x), int32(y), int32(w), int32(h))
  }

  color := frame.P1.Skin.BGColor()

  checkGLError()

  gl.ClearColor(
    float32(color.R)/float32(256),
    float32(color.G)/float32(256),
    float32(color.B)/float32(256),
    float32(color.A)/float32(256),
  )

  gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

  gl.Enable(gl.BLEND)
  gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA)

  checkGLError()

  gl.Enable(gl.DEPTH_TEST)
  gl.DepthFunc(gl.LESS)

  checkGLError()
  gl.UseProgram(app.programs.skinPass)

  checkGLError()

  frame.P1.SyncAndBind()

  checkGLError()
  gl.DrawArrays(gl.TRIANGLES, 0, int32(frame.P1.Len())*3)

  checkGLError()
  gl.UseProgram(app.programs.glyphPass)

  checkGLError()
  frame.P2.SyncAndBind()

  checkGLError()
  gl.DrawArrays(gl.TRIANGLES, 0, int32(frame.P2.Len())*3)

  checkGLError()

  if applyScissor {
    gl.Disable(gl.SCISSOR_TEST)
  }
}

func (app *App) blur(srcI int, dstI int, dirX, dirY float64) {
  var fbo uint32 = 0
  if dstI > -1 {
    fbo = app.programs.fbos[dstI]
  }

  texID := app.programs.fbo_texIDs[srcI]
  texUnit := app.programs.fbo_texUnits[srcI]

  w, h := app.getWindowSize()

  gl.UseProgram(app.programs.blurPass)

  gl.Uniform2f(int32(app.programs.blurPass_uSizeLoc), float32(w), float32(h))

  gl.Uniform2f(int32(app.programs.blurPass_uDirLoc), float32(dirX), float32(dirY))

  gl.ActiveTexture(texUnit)

  gl.BindTexture(gl.TEXTURE_2D, texID)

  gl.Uniform1i(int32(app.programs.blurPass_uTexLoc), int32(texUnit - gl.TEXTURE0))

  gl.EnableVertexAttribArray(app.programs.blurPass_aCoordLoc)
  gl.BindBuffer(gl.ARRAY_BUFFER, app.programs.blurPass_aCoordVBO)
  gl.VertexAttribPointer(app.programs.blurPass_aCoordLoc, 2, gl.FLOAT, false, 0, nil)

  data := []float32{
    0.0, 0.0,
    1.0, 0.0,
    0.0, 1.0,
    1.0, 0.0,
    1.0, 1.0,
    0.0, 1.0,
  }

  gl.BufferData(gl.ARRAY_BUFFER, 4*len(data), gl.Ptr(data), gl.STATIC_DRAW)

  gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

  gl.Viewport(0, 0, int32(w), int32(h))

  gl.Clear(gl.DEPTH_BUFFER_BIT)

  gl.DrawArrays(gl.TRIANGLES, 0, 6)

  checkGLError()
}

/*func (app *App) drawThumbnail(w int, h int, dst unsafe.Pointer) {
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

  checkGLError()
}*/

// eg. apply gaussian blur
// everything generated should be done initialy?
/*func (app *App) drawFiltered(program uint32) {
  if err := app.window.GLMakeCurrent(app.ctx); err != nil {
    fmt.Fprintf(app.debug, "unable to make current: %s\n", err.Error())
    return
  }


  w, h := app.root.GetSize()

  checkGLError()

  gl.Viewport(0, 0, int32(w), int32(h))

  checkGLError()
  // do an inner draw to framebuffer 0
  fb := app.framebuffers[0]
  gl.BindFramebuffer(gl.FRAMEBUFFER, fb)

  checkGLError()
  // create the texture
  var texture uint32
  gl.GenTextures(1, &texture)
  if texture < 0 {
    panic("texture is negative")
  }
  checkGLError()
  gl.BindTexture(gl.TEXTURE_2D, texture)
  checkGLError()
  gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, int32(w), int32(h), 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
  checkGLError()
  gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
  checkGLError()
  gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
  gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
  gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

  // attach the framebuffer to the texture (in order to avoid blitting this must be done before
  //gl.BindFramebuffer(gl.READ_FRAMEBUFFER, fb)
  var colorAttach uint32 = gl.COLOR_ATTACHMENT0
  checkGLError()
  gl.NamedFramebufferTexture(fb, gl.COLOR_ATTACHMENT0, texture, 0)
  fmt.Println("attempted to use framebuffer:", fb)
  checkGLError()
  gl.DrawBuffers(1, &colorAttach)

  if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
    //fmt.Println(gl.CheckFramebufferStatus(gl.FRAMEBUFFER))
    panic("something went wrong")
  }

  checkGLError()
  // do the inner draw
  gl.BindFramebuffer(gl.FRAMEBUFFER, fb)
  checkGLError()
  gl.Viewport(0, 0, int32(w), int32(h))
  checkGLError()
  app.drawInner()

  gl.UseProgram(program)
  dirLoc := getGLUniformLocation(program, "dir")
  sizeLoc := getGLUniformLocation(program, "size")
  gl.Uniform2f(int32(sizeLoc), float32(w), float32(h))

  // bind texture to program location
  gl.ActiveTexture(gl.TEXTURE0)
  gl.BindTexture(gl.TEXTURE_2D, texture)
  texLoc := getGLUniformLocation(program, "frame")
  gl.Uniform1i(int32(texLoc), 0)
  //gl.Viewport(0, 0, int32(w), int32(h))
  var vbo uint32
  coordLoc := getGLAttribLocation(program, "aCoord")
  gl.GenBuffers(1, &vbo)
  if vbo < 0 {
    panic("vbo is negative")
  }
  gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
  gl.VertexAttribPointer(coordLoc, 2, gl.FLOAT, false, 0, nil)
  gl.EnableVertexAttribArray(coordLoc)
  data := []float32{
    0.0, 0.0,
    1.0, 0.0,
    0.0, 1.0,
    1.0, 0.0,
    1.0, 1.0,
    0.0, 1.0,
  }
  gl.BufferData(gl.ARRAY_BUFFER, 4*len(data), gl.Ptr(data), gl.STATIC_DRAW)

  //gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
  //gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

  for _, step := range []string{"x", "y"} {
    //if step == "y" {
      gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
    //} else {
      //gl.BindFramebuffer(gl.FRAMEBUFFER, fb)
      //gl.DrawBuffers(1, &colorAttach)
    //}
    gl.Viewport(0, 0, int32(w), int32(h))
    gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

    // buffer the tri data for the filter program

    // update the size

    // update the direction
    if step == "x" {
      gl.Uniform2f(int32(dirLoc), 1.0, 0.0);
    } else {
      gl.Uniform2f(int32(dirLoc), 0.0, 1.0);
    }

    gl.DrawArrays(gl.TRIANGLES, 0, 6) // 2 triangles with 3 vertices each

    if step == "x" {
    }
  }

  checkGLError()

  // clean up
  gl.BindBuffer(gl.ARRAY_BUFFER, 0)
  gl.DisableVertexAttribArray(coordLoc)
  gl.BindTexture(gl.TEXTURE_2D, 0)
  gl.DeleteTextures(1, &texture)
  gl.DeleteBuffers(1, &vbo)

  // make sure the default framebuffer is also the read framebuffer
  gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

  app.window.GLSwap()

  // now bind the framebuffer to a texture so we can apply the filtering
  if err := app.window.GLMakeCurrent(nil); err != nil {
    fmt.Fprintf(app.debug, "unable to unmake current: %s\n", err.Error())
    return
  }

  checkGLError()
}*/

/*func (app *App) drawFiltered1D(fbIn uint32, fbOut uint32, dirX float32, dirY float32, colorAttachment uint32, texUnit uint32, fnDraw func()) {
  w, h := app.root.GetSize()

  checkGLError()

  gl.Viewport(0, 0, int32(w), int32(h))

  checkGLError()
  gl.BindFramebuffer(gl.FRAMEBUFFER, fbIn)

  checkGLError()
  // create the texture
  var texture uint32
  gl.GenTextures(1, &texture)
  if texture < 0 {
    panic("texture is negative")
  }
  checkGLError()
  gl.BindTexture(gl.TEXTURE_2D, texture)
  checkGLError()
  gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, int32(w), int32(h), 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
  checkGLError()
  gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
  checkGLError()
  gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
  gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
  gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

  // attach the framebuffer to the texture (in order to avoid blitting this must be done before
  checkGLError()
  fmt.Println("attempting to draw to", fbIn)
  gl.NamedFramebufferTexture(fbIn, colorAttachment, texture, 0)
  checkGLError()
  var colorAttachment_ uint32 = colorAttachment
  gl.DrawBuffers(1, &colorAttachment_)

  if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
    //fmt.Println(gl.CheckFramebufferStatus(gl.FRAMEBUFFER))
    panic("something went wrong")
  }

  checkGLError()
  // do the inner draw
  gl.BindFramebuffer(gl.FRAMEBUFFER, fbIn)
  checkGLError()
  gl.Viewport(0, 0, int32(w), int32(h))
  checkGLError()
  fnDraw()

  gl.UseProgram(app.gaussBlur)
  dirLoc := getGLUniformLocation(app.gaussBlur, "dir")
  sizeLoc := getGLUniformLocation(app.gaussBlur, "size")
  gl.Uniform2f(int32(sizeLoc), float32(w), float32(h))

  // bind texture to program location
  gl.ActiveTexture(texUnit)
  gl.BindTexture(gl.TEXTURE_2D, texture)
  texLoc := getGLUniformLocation(app.gaussBlur, "frame")
  gl.Uniform1i(int32(texLoc), int32(texUnit - gl.TEXTURE0))
  var vbo uint32
  coordLoc := getGLAttribLocation(app.gaussBlur, "aCoord")
  gl.GenBuffers(1, &vbo)
  if vbo < 0 {
    panic("vbo is negative")
  }
  gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
  gl.VertexAttribPointer(coordLoc, 2, gl.FLOAT, false, 0, nil)
  gl.EnableVertexAttribArray(coordLoc)
  data := []float32{
    0.0, 0.0,
    1.0, 0.0,
    0.0, 1.0,
    1.0, 0.0,
    1.0, 1.0,
    0.0, 1.0,
  }
  gl.BufferData(gl.ARRAY_BUFFER, 4*len(data), gl.Ptr(data), gl.STATIC_DRAW)

  gl.BindFramebuffer(gl.FRAMEBUFFER, fbOut)
  gl.Viewport(0, 0, int32(w), int32(h))
  gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

  // buffer the tri data for the filter program

  // update the size

  // update the direction
  gl.Uniform2f(int32(dirLoc), dirX, dirY);

  gl.DrawArrays(gl.TRIANGLES, 0, 6) // 2 triangles with 3 vertices each

  checkGLError()

  // clean up
  gl.BindBuffer(gl.ARRAY_BUFFER, 0)
  gl.DisableVertexAttribArray(coordLoc)
  gl.BindTexture(gl.TEXTURE_2D, 0)
  gl.DeleteTextures(1, &texture)
  gl.DeleteBuffers(1, &vbo)
}

func (app *App) DrawBlurred() {
  // old way
  //app.drawFiltered(app.gaussBlur)


  // new way
  if err := app.window.GLMakeCurrent(app.ctx); err != nil {
    fmt.Fprintf(app.debug, "unable to make current: %s\n", err.Error())
    return
  }

  // one pass
  app.drawFiltered1D(app.framebuffers[0], 0, 1.0, 0.0, gl.COLOR_ATTACHMENT0, gl.TEXTURE0, func() {
    app.drawInner()
  })


  // two pass
  app.drawFiltered1D(app.framebuffers[0], 0, 0.0, 1.0, gl.COLOR_ATTACHMENT1, gl.TEXTURE1, func() {
    app.drawFiltered1D(app.framebuffers[1], app.framebuffers[0], 1.0, 0.0, gl.COLOR_ATTACHMENT0, gl.TEXTURE0, func() {
      app.drawInner()
    })
  })

  gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

  app.window.GLSwap()

  // now bind the framebuffer to a texture so we can apply the filtering
  if err := app.window.GLMakeCurrent(nil); err != nil {
    fmt.Fprintf(app.debug, "unable to unmake current: %s\n", err.Error())
    return
  }

  checkGLError()

  sdl.Delay(10000)
}*/

/*func (app *App) drawAndCopyToBitmap(w int, h int, dst unsafe.Pointer) {
  checkGLError()
  if err := app.window.GLMakeCurrent(app.ctx); err != nil {
    fmt.Fprintf(app.debug, "unable to make current: %s\n", err.Error())
    return
  }

  checkGLError()
  gl.BindFramebuffer(gl.FRAMEBUFFER, app.framebuffers[0])

  checkGLError()
  gl.Viewport(0, 0, int32(w), int32(h))

  checkGLError()
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
  checkGLError()
}*/

func (app *App) endDrawLoop() {
  sdl.GLDeleteContext(app.ctx)
}
