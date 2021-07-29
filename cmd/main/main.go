package main

import (
  "github.com/computeportal/glui"
)

func main() {
  app := glui.NewApp("test", &glui.ClassicSkin{}, MakeGlyphs())

  rainbow1 := glui.NewRainbow(app.DrawData())
  app.Body().AppendChild(rainbow1)

  rainbow2 := glui.NewRainbow(app.DrawData())
  app.Body().AppendChild(rainbow2)

  button1 := glui.NewButton(app.DrawData())
  app.Body().AppendChild(button1)

  content := "it is often useful to perform premultiplied alpha blending a slight modification"

  text := glui.NewText(app.DrawData(), content, "dejavusans", 10)
  app.Body().AppendChild(text)

  app.Run()
}
