package main

import (
  "github.com/computeportal/glui"
)

func main() {
  app := glui.NewApp("test", &glui.ClassicSkin{}, MakeGlyphs())

  root := app.Root()

  //rainbow1 := glui.NewRainbow(dd)
  //rainbow2 := glui.NewRainbow(dd)

  body := app.Body()
  body.Padding(10)
  body.Spacing(10)

  button1 := glui.NewButton(root)
  button1.A(glui.NewHor(root, glui.CENTER, glui.CENTER, 0).A(glui.NewSans(root, "Submit", 10)))

  /*input1 := glui.NewInput(root).Padding(0, 2)

  icon := glui.NewIcon(root, "floppy", 30)
  button2 := glui.NewFlatButton(root).Size(40, 40)
  button2.A(glui.NewHor(root, glui.CENTER, glui.CENTER, 0).A(icon))

  dropdown := glui.NewDropdown(root)*/

  body.A(
    button1,
    //input1, glui.NewInput(root), button2, dropdown,
  )

  app.Run()
}
