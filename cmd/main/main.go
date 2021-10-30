package main

import (
  //"time"

  "github.com/computeportal/glui"
)

func main() {
  app := glui.NewApp("test", &glui.ClassicSkin{}, MakeGlyphs())

  root := app.Root() // single (global) root, or multiple roots?

  body := app.Body()
  body.Padding(10)

  tabbed := glui.NewTabbed(root)
  tabPage1 := tabbed.NewTab("Tab one", false)
  tabPage1.Spacing(10)

  tabPage2 := tabbed.NewTab("Tab two", true)

  // the following 2 Submit buttons are equivalent
  button1 := glui.NewCaptionButton(root, "Submit")
  //button1_ := glui.NewButton(root)
  //button1_.A(glui.NewHor(root, glui.CENTER, glui.CENTER, 0).A(glui.NewSans(root, "Submit", 10))) 

  input1 := glui.NewInput(root).Padding(0, 2)

  //icon := glui.NewIcon(root, "floppy", 30)
  button2 := glui.NewFlatIconButton(root, "floppy", 30).Size(40, 40)
  //button2.A(glui.NewHor(root, glui.CENTER, glui.CENTER, 0).A(icon))

  dropdown := glui.NewSelect(root, []string{"Dog", "Cat", "Hamster"})


  tabPage1.A(input1, button1, button2)


  tabPage2.A(dropdown)

  //tabbed2 := glui.NewTabbed(root)
  //otherSidePage := tabbed2.NewTab("other side", false)

  table := glui.NewTable(root).MasterColumn(0)
  column1 := glui.NewTextColumn(root, "Name")
  column2 := glui.NewDateColumn(root, "Date")
  table.A(column1, column2)
  //otherSidePage.A(table.A(

  table.AddRow("Alice", "1865-10-26")
  table.AddRow("Bob", "1945-02-06")
  table.AddRow("Charlie", "1889-04-16")

  vsplit := glui.NewVSplit(root)
  vsplit.MinIntervals([]int{300, 300})
  vsplit.A(table, tabbed)


  menuItems:= []glui.MenuItemConfig {
    glui.MenuItemConfig{
      "Close",
      func(){app.Quit()},
      120,
    },
    glui.MenuItemConfig{
      "Blur",
      func(){app.DrawBlurred()},
      120,
    },
  }

  body.A(glui.NewHor(root, glui.END, glui.CENTER, 0).A(glui.NewIconDropdown(root, "menu", 30, glui.END, menuItems).Size(40, 40)))
  body.A(vsplit)

  app.Run()
}
