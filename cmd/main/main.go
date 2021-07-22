package main

import (
  "github.com/computeportal/glui"
)

func main() {
  app := glui.NewApp("test", &glui.ClassicSkin{})

  rainbow1 := glui.NewRainbow(app.DrawData())
  app.Body().AppendChild(rainbow1)

  rainbow2 := glui.NewRainbow(app.DrawData())
  app.Body().AppendChild(rainbow2)

  button1 := glui.NewButton(app.DrawData())
  app.Body().AppendChild(button1)

  app.Run()
}
