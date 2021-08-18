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
    app.root.CalcDepth()

    app.root.CalcPos()

    if app.mouseInWindow() {
      app.updateMouseElement(-1, -1)
    }
  }

  if app.root.dirty() {
    fmt.Println("redrawing...")

    app.Draw()
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

  //app.draw()

  m.Unlock()

  for true {
    //fmt.Println("in render loop")
    //for _ = range app.drawCh {
    //}
    <- app.drawCh

    app.draw()

    sdl.Delay(RENDER_LOOP_INTERVAL)

    /*draining := true
    for draining {
      select {
      case <- app.drawCh:
        continue
      default:
        draining = false
        break
      }
    }*/

    //}
    //default:
    //fmt.Println("nothing done in render loop")
  }

  sdl.GLDeleteContext(ctx)
}
