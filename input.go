package glui

import (
  "fmt"
  "math"
  "os"
  "strings"

  "github.com/veandco/go-sdl2/sdl"
)

// overflow not (yet) allowed
type Input struct {
  ElementData

  tris []uint32
  dd   *DrawData

  width       int
  height      int
  borderT     int
  barHeight   int

  // state
  text        *Text 
  selText     *Text
  value       string // not necessarily the same as text (in case of overflow)
  col0        int // defaults to end of string
  col1        int // end of selection, same as col0 for no selection
  focused     bool
  mouseDown   bool
  currentVBar bool
  lastTick    uint64
  vBarTick    uint64
  menuVisible bool
}

func NewInput(dd *DrawData) *Input {
  tris := dd.P1.Alloc(10*2) // so first 18 tris are used for border, tri 18 and 19 are used for vertical bar

  e := &Input{
    newElementData(), 
    tris, 
    dd, 
    200,
    50,
    dd.P1.Skin.InputBorderThickness(),
    25,
    NewText(dd, "", "dejavumono", 10), 
    NewText(dd, "", "dejavumono", 10),
    "", 
    0, 
    0,
    false, 
    false,
    false, 
    0, 
    0,
    false,
  }

  e.selText.SetColor(sdl.Color{0xff, 0xff, 0xff, 0xff})

  // default is 1px of right padding (to accomodate the vBar)
  e.padding[1] = 1

  e.setTypesAndTCoords()

  e.SetEventListener("keypress",  e.onKeyPress)
  e.SetEventListener("textinput", e.onTextInput)
  e.SetEventListener("focus",     e.onFocus)
  e.SetEventListener("blur",      e.onBlur)
  e.SetEventListener("mousedown", e.onMouseDown)
  e.SetEventListener("mousemove", e.onMouseMove)
  e.SetEventListener("mouseup",   e.onMouseUp)
  e.SetEventListener("doubleclick", e.onDoubleClick)
  e.SetEventListener("tripleclick", e.onTripleClick)
  e.SetEventListener("rightclick", e.onRightClick)

  return e
}

func (e *Input) onKeyPress(evt *Event) {
  switch {
  case evt.Key == "backspace":
    n := len(e.value)
    if n > 0 {
      if e.hasSel() {
        e.delSel()
      } else if e.atEnd() {
        col := moveInputCol(e.value, e.col0, false, evt.Ctrl)
        e.value = e.value[0:col]
        e.col0 = col
        e.col1 = col
      } else if e.col0 > 0 {
        col := moveInputCol(e.value, e.col0, false, evt.Ctrl)
        e.value = e.value[0:col] + e.value[e.col0:]
        e.col0 = col
        e.col1 = col
      }
    }
    e.refreshVBar()
    break
  case evt.Key == "v" && evt.Ctrl:
    e.insertClipboard()
    break
  case evt.Key == "x" && evt.Ctrl && e.hasSel():
    e.cutSel()
    break
  case evt.Key == "c" && evt.Ctrl && e.hasSel():
    e.copySel()
    break
  case evt.Key == "a" && evt.Ctrl:
    e.selAll()
    e.refreshVBar()
    break
  case evt.Key == "delete":
    n := len(e.value) 
    if n > 0 && !e.atEnd() {
      if e.hasSel() {
        e.delSel()
      } else {
        col := moveInputCol(e.value, e.col0, true, evt.Ctrl)
        e.value = e.value[0:e.col0] + e.value[col:]
      }
    }
    e.refreshVBar()
    break
  case evt.Key == "left":
    col := moveInputCol(e.value, e.col1, false, evt.Ctrl)
    if evt.Shift {
      e.col1 = col
    } else {
      if e.hasSel() {
        col := e.selStart()
        e.col0 = col
        e.col1 = col
      } else {
        e.col0 = col
        e.col1 = e.col0
      }
    }
    e.refreshVBar()
    break
  case evt.Key == "right":
    col := moveInputCol(e.value, e.col1, true, evt.Ctrl)
    if evt.Shift {
      e.col1 = col
    } else {
      if e.hasSel() {
        col := e.selEnd()
        e.col0 = col
        e.col1 = col
      } else {
        e.col0 = col
        e.col1 = e.col0
      }
    }
    e.refreshVBar()
    break
  case evt.Key == "home":
    if evt.Shift {
      e.col1 = 0
    } else {
      e.col0 = 0
      e.col1 = 0
    }
    e.refreshVBar()
    break
  case evt.Key == "end":
    if evt.Shift {
      e.col1 = len(e.value)
    } else {
      e.col0 = len(e.value)
      e.col1 = e.col0
    }
    e.refreshVBar()
    break
  }

  e.sync()
}

func (e *Input) selAll() {
  e.col0 = 0
  e.col1 = len(e.value)
}

func (e *Input) onDoubleClick(evt *Event) {
  // the first mouse up will have correctly set col0 and col1
  // now expand to next word boundary

  col0 := moveInputCol(e.value, e.col0, false, true)
  col1 := moveInputCol(e.value, e.col1, true, true)

  e.col0 = col0
  e.col1 = col1

  for ; e.col1 > 0 && isDelimiter(rune(e.value[e.col1-1])); {
    e.col1-=1
  }


  e.refreshVBar()
  e.sync()
}

func (e *Input) onTripleClick(evt *Event) {
  e.selAll()
  e.refreshVBar()
  e.sync()
}

func (e *Input) onMouseUp(evt *Event) {
  if e.mouseDown {
    e.mouseDown = false
  }
}

func (e *Input) onMouseMove(evt *Event) {
  if e.mouseDown {
    col := e.mousePosToCol(evt)

    if col != e.col1 {
      e.col1 = col
      e.refreshVBar()
      e.sync()
    }
  }
}

func (e *Input) onRightClick(evt *Event) {
  e.fillRightClickMenu()

  e.dd.Dialog.Show(Rect{evt.X, evt.Y, 70, 100})

  e.menuVisible = true
}

func (e *Input) mousePosToCol(evt *Event) int {
  relX := evt.X - e.bb.X

  // RIGHT ALIGN

  // from right
  relX = e.width - relX - e.padding[1] - e.borderT

  colFromRight := math.Floor((float64(relX))/e.text.RefAdvance() + 0.5)
  if colFromRight < 0.0 {
    colFromRight = 0.0
  }

  col := len(e.value) - int(colFromRight)
  if col < 0 {
    col = 0
  } 

  return col
}

func (e *Input) onMouseDown(evt *Event) {
  col := e.mousePosToCol(evt)

  e.col0 = col
  e.col1 = col

  e.mouseDown = true

  e.refreshVBar()
  e.sync()
}

func moveInputCol(value string, col int, moveRight bool, word bool) int {
  if word {
    // move by word
    if moveRight {
      if col >= len(value) {
        return len(value)
      } else {
        for i, c := range value[col:] {
          if i > 0 {
            prev := value[col+i-1]
            if isDelimiter(rune(prev)) && !isDelimiter(rune(c)) {
              return col+i
            }
          }
        }
      }

      return len(value)
    } else {
      if col <= 0 {
        return 0
      } 

      for i := col - 1; i > 0; i-- {
        c := value[i]
        prev := value[i-1]

        if isDelimiter(rune(prev)) && !isDelimiter(rune(c)) {
          return i
        }
      }

      return 0
    }
  } else {
    if moveRight {
      if col+1 >= len(value) {
        return len(value)
      } else {
        return col + 1
      }
    } else {
      if col - 1 <= 0 {
        return 0
      } else {
        return col -1
      }
    }
  }
}

func (e *Input) atEnd() bool {
  return e.col0 == len(e.value)
}

func (e *Input) maxLen() int {
  return (e.width - e.padding[1] - e.padding[3] - 2*e.borderT)/int(e.text.RefAdvance())
}

func (e *Input) delSel() {
  col := e.selStart()
  if e.selEnd() == len(e.value) {
    e.value = e.value[0:col]
  } else {
    e.value = e.value[0:col] + e.value[e.selEnd():]
  }
  
  e.col0 = col
  e.col1 = col
}

func (e *Input) cutSel() {
  txt := e.getSelText()
  e.delSel()

  if err := sdl.SetClipboardText(txt); err != nil {
    fmt.Fprintf(os.Stderr, "failed to set clipboard text: %s\n", err.Error())
  }
}

func (e *Input) copySel() {
  txt := e.getSelText()

  if err := sdl.SetClipboardText(txt); err != nil {
    fmt.Fprintf(os.Stderr, "failed to set clipboard text: %s\n", err.Error())
  }
}

func (e *Input) insertClipboard() {
  txt, err := sdl.GetClipboardText()
  if err == nil {
    if e.hasSel() {
      e.delSel()
    }
    e.insertText(txt)
  } else {
    fmt.Fprintf(os.Stderr, "failed to get clipboard text: %s\n", err.Error())
  }
}

func (e *Input) onTextInput(evt *Event) {
  if e.hasSel() {
    e.delSel()
  }

  e.insertText(evt.Value)
}

func (e *Input) insertText(text string) {
  if len(e.value) != e.maxLen() {
    v := text
    if len(e.value) + len(v) > e.maxLen() {
      v = v[0:e.maxLen() - len(e.value)]
    }

    if e.atEnd() {
      e.value += v
    } else {
      e.value = e.value[0:e.col0] + v + e.value[e.col0:]
    }

    e.col0 += len(v)
    e.col1 = e.col0
  }

  e.refreshVBar()

  e.sync()
}

func (e *Input) onFocus(evt *Event) {
  e.focused = true

  e.refreshVBar()
  e.sync()

  e.dd.FocusBox.Show(e.bb)
}

func (e *Input) onBlur(evt *Event) {
  e.focused = false

  e.sync()
  e.hideVBar()

  e.dd.FocusBox.Hide()
}

func (e *Input) Cursor() int {
  return sdl.SYSTEM_CURSOR_IBEAM
}

func (e *Input) setTypesAndTCoords() {
  x0, y0 := e.dd.P1.Skin.InputOrigin()

  setBorderElementTypesAndTCoords(e.dd, e.tris, x0, y0, e.borderT, e.dd.P1.Skin.InputBGColor())
  e.setVBarTypeAndColor()
}

func (e *Input) sync() {
  if e.hasSel() && e.focused {
    e.text.SetContent(e.value[0:e.selStart()] + strings.Repeat(" ", e.selWidth()) + e.value[e.selEnd():])
    e.selText.SetContent(strings.Repeat(" ", e.selStart()) + e.getSelText() + strings.Repeat(" ", len(e.value) - e.selEnd()))
  } else {
    e.text.SetContent(e.value)
    e.selText.SetContent(strings.Repeat(" ", len(e.value)))
  }

  for _, textElem := range []*Text{e.text, e.selText} {
    textWidth, textHeight := textElem.OnResize(e.width - e.padding[1] - e.padding[3] - 2*e.borderT, 0)

    // RIGHT ALIGN
    textElem.Translate(
      e.bb.X + e.width - e.borderT - textWidth - e.padding[1], 
      e.bb.Y + (e.height - textHeight)/2, 0.0) 
  }

  e.syncVBarPos()
}

func (e *Input) setVBarTypeAndColor() {
  tri0 := e.tris[18]
  tri1 := e.tris[19]

  e.dd.P1.Color.Set4Const(tri0, 0.0, 0.0, 0.0, 1.0)
  e.dd.P1.Color.Set4Const(tri1, 0.0, 0.0, 0.0, 1.0)

  e.hideVBar()
}

func (e *Input) refreshVBar() {
  if !e.currentVBar {
    e.showVBar()
  }

  e.vBarTick = e.lastTick
}

func (e *Input) fillRightClickMenu() {
  e.dd.Dialog.Clear()
  e.dd.Dialog.Padding(5)
  e.dd.Dialog.Spacing(0)

  if len(e.dd.Dialog.Children()) == 0 {
    cutButton := NewFlatButton(e.dd)
    cutButton.A(NewHor(START, CENTER, 0).A(e.dd.Sans("Cut", 10)))
    cutButton.SetSize(60, 30)
    cutButton.Padding(0, 10)
    cutButton.SetZ(-0.5)
    cutButton.OnClick(func() {
      if e.hasSel() {
        e.cutSel()
      }
      e.dd.Dialog.Hide()
    })

    copyButton := NewFlatButton(e.dd)
    copyButton.A(NewHor(START, CENTER, 0).A(e.dd.Sans("Copy", 10)))
    copyButton.SetSize(60, 30)
    copyButton.Padding(0, 10)
    copyButton.SetZ(-0.5)
    copyButton.OnClick(func() {
      if e.hasSel() {
        e.copySel()
      }
      e.dd.Dialog.Hide()
    })

    pasteButton := NewFlatButton(e.dd)
    pasteButton.A(NewHor(START, CENTER, 0).A(e.dd.Sans("Paste", 10)))
    pasteButton.SetSize(60, 30)
    pasteButton.Padding(0, 10)
    pasteButton.SetZ(-0.5)
    pasteButton.OnClick(func() {
      e.insertClipboard()
      e.dd.Dialog.Hide()
    })

    e.dd.Dialog.A(cutButton, copyButton, pasteButton)
  }
}

func (e *Input) showVBar() {
  tri0 := e.tris[18]
  tri1 := e.tris[19]

  e.dd.P1.Type.Set1Const(tri0, VTYPE_PLAIN)
  e.dd.P1.Type.Set1Const(tri1, VTYPE_PLAIN)

  e.currentVBar = true
}

func (e *Input) hideVBar() {
  tri0 := e.tris[18]
  tri1 := e.tris[19]

  e.dd.P1.Type.Set1Const(tri0, VTYPE_HIDDEN)
  e.dd.P1.Type.Set1Const(tri1, VTYPE_HIDDEN)

  e.currentVBar = false
}

func (e *Input) OnResize(maxWidth, maxHeight int) (int, int) {
  setBorderElementPosZ(e.dd, e.tris, e.width, e.height, e.borderT, e.z)

  e.InitBB(e.width, e.height)

  e.sync()

  if e.focused {
    e.dd.FocusBox.Show(Rect{0, 0, e.width, e.height})
  }

  if e.menuVisible {
    // origin is handled by translate
    e.dd.Dialog.Show(Rect{0, 0, 70, 100})
  }

  return e.width, e.height
}

func (e *Input) selStart() int {
  if e.col0 < e.col1 {
    return e.col0
  } else {
    return e.col1
  }
}

func (e *Input) selEnd() int {
  if e.col0 < e.col1 {
    return e.col1
  } else {
    return e.col0
  }
}

func (e *Input) selWidth() int {
  return e.selEnd() - e.selStart()
}

func (e *Input) hasSel() bool {
  return e.col0 != e.col1
}

func (e *Input) getSelText() string {
  return e.value[e.selStart():e.selEnd()]
}

func (e *Input) syncVBarPos() {
  y0 := e.bb.Y + e.height/2 - e.barHeight/2

  // Right Aligned
  x0 := e.bb.X + e.width - e.padding[1] - e.borderT - 
    int(math.Ceil(float64(len(e.value) - e.selStart())*e.text.RefAdvance()))

  tri0 := e.tris[18]
  tri1 := e.tris[19]

  var vBarWidth int

  if e.col0 == e.col1 {
    vBarWidth = 1

    e.dd.P1.SetColorConst(tri0, sdl.Color{0, 0, 0, 255})
    e.dd.P1.SetColorConst(tri1, sdl.Color{0, 0, 0, 255})
    
    e.dd.P1.SetQuadPos(tri0, tri1, Rect{x0, y0, vBarWidth, e.barHeight}, 0.4)
  } else {
    vBarWidth = e.selWidth()*int(e.text.RefAdvance())

    e.dd.P1.SetColorConst(tri0, e.dd.P1.Skin.SelColor())
    e.dd.P1.SetColorConst(tri1, e.dd.P1.Skin.SelColor())
  }

  e.dd.P1.SetQuadPos(tri0, tri1, Rect{x0, y0, vBarWidth, e.barHeight}, 0.4)
}

func (e *Input) Translate(dx, dy int, dz float32) {
  for _, tri := range e.tris {
    e.dd.P1.TranslateTri(tri, dx, dy, dz)
  }

  e.ElementData.Translate(dx, dy, dz)

  e.text.Translate(dx, dy, dz)
  e.selText.Translate(dx, dy, dz)

  if e.focused {
    e.dd.FocusBox.Translate(dx, dy, dz)
  }

  if e.menuVisible {
    e.dd.Dialog.Translate(dx, dy, 0.0)
  }
}

func (e *Input) OnTick(tick uint64) {
  e.lastTick = tick

  if e.focused && !e.hasSel() {
    if (tick - e.vBarTick + 1)%30 == 0 {
      if e.currentVBar {
        e.hideVBar()
      } else {
        e.showVBar()
      }
    }
  }
}
