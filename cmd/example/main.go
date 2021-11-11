package main

import (
  "fmt"
  "os"
  "os/signal"
  "runtime/pprof"

  "github.com/computeportal/glui"
)

func main() {
  prof := false
  profFile := "example.prof"
  for _, arg := range os.Args {
    if arg == "-prof" || arg == "--prof" {
      prof = true
    }
  }

  if prof {
    startProfiling(profFile)
  }

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
  rg := glui.NewRadioGroup([]string{"Jaguar", "Rabbit", "Parrot", "Turtle", "Camel"}, glui.VER)
  img1 := glui.NewImage(nil)
  rg.OnChange(func(i int, _ string) {
    switch i {
    case 0:
      img1.Img(jaguar)
    case 1:
      img1.Img(rabbit)
    case 2:
      img1.Img(parrot)
    case 3:
      img1.Img(turtle)
    case 4:
      img1.Img(camel)
    }
  })

  sel := glui.NewSelect([]string{"Dog", "Cat", "Hamster"})

  overflow := glui.NewOverflow().Size(-1, -1).A(rg, img1)

  tabPage1.A(input1, button1, button2, cb, overflow)

  img2 := glui.NewImage(nil)
  sel.OnChange(func(i int, _ string) {
    switch i {
    case 0:
      img2.Img(dog)
    case 1:
      img2.Img(cat)
    case 2:
      img2.Img(hamster)
    }
  })
  tabPage2.A(sel, img2)

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
        evt.Callback(false)
      }), glui.NewCaptionButton("Yes").W(100).OnClick(func() {
        evt.Callback(true)
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

  if prof {
    stopProfiling(profFile)
  }
}

var fProf *os.File = nil

func printMessageAndExit(msg string) {
	fmt.Fprintf(os.Stderr, "\u001b[1m"+msg+"\u001b[0m\n\n")
  os.Exit(1)
}

func startProfiling(profFile string) {
  var err error
  fProf, err = os.Create(profFile)
  if err != nil {
    printMessageAndExit(err.Error())
  }

  pprof.StartCPUProfile(fProf)

  go func() {
    sigchan := make(chan os.Signal)
    signal.Notify(sigchan, os.Interrupt)
    <-sigchan

    stopProfiling(profFile)

    os.Exit(1)
  }()
}

func stopProfiling(profFile string) {
  if fProf != nil {
		pprof.StopCPUProfile()

    // also write mem profile
		fMem, err := os.Create(profFile + ".mprof")
		if err != nil {
			printMessageAndExit(err.Error())
		}

		pprof.WriteHeapProfile(fMem)
		fMem.Close()

    fProf = nil
  }
}
