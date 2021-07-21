package main

import (
  "github.com/computeportal/glui"
)

func main() {
  app := glui.NewApp("test")

  rainbow1 := glui.NewRainbow(app.DrawData())
  app.Body().AppendChild(rainbow1)

  rainbow2 := glui.NewRainbow(app.DrawData())
  app.Body().AppendChild(rainbow2)

  app.Run()
}
