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

  tabbed := glui.NewTabbed(root)
  tabPage1 := tabbed.NewTab("Tab one", false)
  tabPage1.Spacing(10)

  tabPage2 := tabbed.NewTab("Tab two", true)

  tabbed.NewTab("Tab three", true)
  tabbed.NewTab("Tab four", true)

  button1 := glui.NewButton(root)
  button1.A(glui.NewHor(root, glui.CENTER, glui.CENTER, 0).A(glui.NewSans(root, "Submit", 10)))

  input1 := glui.NewInput(root).Padding(0, 2)

  //icon := glui.NewIcon(root, "floppy", 30)
  button2 := glui.NewFlatIconButton(root, "floppy", 30).Size(40, 40)
  //button2.A(glui.NewHor(root, glui.CENTER, glui.CENTER, 0).A(icon))

  dropdown := glui.NewDropdown(root, []string{"Dog", "Cat", "Hamster"})

  tabPage1.A(button1, input1, button2)
  tabPage2.A(dropdown)

  body.A(tabbed)

  app.Run()
}
