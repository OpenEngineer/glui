package main

import (
  "github.com/computeportal/glui"
)

func main() {
  app := glui.NewApp("test", &glui.ClassicSkin{}, MakeGlyphs())

  dd := app.DrawData()

  rainbow1 := glui.NewRainbow(dd)
  rainbow2 := glui.NewRainbow(dd)

  body := app.Body()
  body.Padding(10)
  body.Spacing(10)

  body.A(rainbow1, rainbow2, dd.Button().A(
      dd.Inline(glui.CENTER, glui.CENTER, 0).A(
        dd.Sans("Submit", 10),
      ),
    ), 
    dd.Input(),
  )

  app.Run()
}
