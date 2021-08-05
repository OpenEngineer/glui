package main

import (
  "github.com/computeportal/glui"
)

func main() {
  app := glui.NewApp("test", &glui.ClassicSkin{}, MakeGlyphs())

  dd := app.DrawData()

  //rainbow1 := glui.NewRainbow(dd)
  //rainbow2 := glui.NewRainbow(dd)

  body := app.Body()
  body.Padding(10)
  body.Spacing(10)

  button1 := glui.NewFlatButton(dd)
  button1.A(dd.Hor(glui.CENTER, glui.CENTER, 0).A(dd.Sans("Submit", 10)))

  input1 := dd.Input()
  input1.Padding(0, 2)

  icon := glui.NewIcon(dd, "floppy", 30)
  button2 := glui.NewFlatButton(dd)
  button2.SetSize(40, 40)
  button2.A(dd.Hor(glui.CENTER, glui.CENTER, 0).A(icon))

  dropdown := glui.NewDropdown(dd)

  body.A(
    button1,
    input1,
    dd.Input(),
    button2,
    dropdown,
  )

  app.Run()
}
