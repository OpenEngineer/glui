package main

import (
  //"time"

  "github.com/computeportal/glui"
)

func main() {
  glui.NewApp("test", &glui.ClassicSkin{}, MakeGlyphs(), 2)

  body := glui.ActiveBody()
  body.Padding(10)

  tabbed := glui.NewTabbed()
  tabPage1 := tabbed.NewTab("Tab one", false)
  tabPage1.Spacing(10)

  tabPage2 := tabbed.NewTab("Tab two", true)

  // the following 2 Submit buttons are equivalent
  button1 := glui.NewCaptionButton("Submit")
  //button1_ := glui.NewButton()
  //button1_.A(glui.NewHor(glui.CENTER, glui.CENTER, 0).A(glui.NewSans( "Submit", 10))) 

  input1 := glui.NewInput().Padding(0, 2)

  //icon := glui.NewIcon("floppy", 30)
  button2 := glui.NewFlatIconButton("floppy", 30).Size(40, 40)
  //button2.A(glui.NewHor(glui.CENTER, glui.CENTER, 0).A(icon))

  cb := glui.NewCheckbox()

  dropdown := glui.NewSelect([]string{"Dog", "Cat", "Hamster"})

  tabPage1.A(input1, button1, button2, cb)

  tabPage2.A(dropdown)

  //tabbed2 := glui.NewTabbed()
  //otherSidePage := tabbed2.NewTab("other side", false)

  table := glui.NewTable().MasterColumn(0)
  column1 := glui.NewTextColumn("Name")
  column2 := glui.NewDateColumn("Date")
  table.A(column1, column2)
  //otherSidePage.A(table.A(

  table.AddRow("Alice", "1865-10-26")
  table.AddRow("Bob", "1945-02-06")
  table.AddRow("Charlie", "1889-04-16")

  vsplit := glui.NewVSplit()
  vsplit.MinIntervals([]int{300, 300})
  vsplit.A(table, tabbed)

  body.On("quit", func(evt *glui.Event) {
    glui.PushFrame(400, 200)

    glui.ActiveBody().A(glui.NewVer(glui.CENTER, glui.CENTER, 40).W(-1).A(glui.NewSans("Are you sure you want to quit?", 12), 
      glui.NewHor(glui.START, glui.CENTER, 20).A(
      glui.NewCaptionButton("No").W(100).OnClick(func() {
        evt.StopPropagation()
        glui.PopFrame()
      }), glui.NewCaptionButton("Yes").W(100).OnClick(func() {
        evt.Callback()
      }))))
  })

  menuItems:= []glui.MenuItemConfig {
    glui.MenuItemConfig{
      "Close",
      func(){
        glui.Quit()
      },
      120,
    },
  }

  body.A(glui.NewHor(glui.END, glui.CENTER, 0).A(glui.NewIconDropdown("menu", 30, glui.END, menuItems).Size(40, 40)))
  body.A(vsplit)

  glui.Run()

}
