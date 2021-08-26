package glui

import (
  "fmt"

  "github.com/veandco/go-sdl2/sdl"
)

//go:generate ./gen_element Dropdown "appendChild CalcDepth On Size Padding"

//go:generate ./gen_element DropdownWrapper "CalcDepth"

type DropdownWrapper struct { // given to focusrect instead of actual wrapper
  ElementData

  dropdown *Dropdown
}

type Dropdown struct {
  ElementData

  options []string
  text   *Text
  arrow  *Icon
  wrapper *DropdownWrapper

  value string
}


func NewDropdown(root *Root, options []string) *Dropdown {
  e := &Dropdown{
    NewElementData(root, 9*2, 0),
    options, 
    NewSans(root, "Choose animal", 10), 
    NewIcon(root, "arrow-down-drop", 10),
    nil,
    "",
  }
  
  e.wrapper = &DropdownWrapper{NewElementData(root, 0, 0), e}

  e.Size(200, 50)
  e.Padding(e.Root.P1.Skin.ButtonBorderThickness())

  e.appendChild(NewHor(root, STRETCH, CENTER, 0).Padding(0, 10).A(e.text, e.arrow))
  e.Show()


  e.On("mousedown", e.onMouseDown)
  e.On("mousebuttonoutsidemenu", e.onMouseButtonOutsideMenu)
  e.On("focus", e.onFocus)
  e.On("blur", e.onBlur)
  e.On("keydown", e.onKeyDown)
  e.On("keypress", e.onKeyPress)

  return e
}

func (e *Dropdown) Cursor() int {
  if e.enabled {
    return sdl.SYSTEM_CURSOR_HAND
  } else {
    return -1
  }
}

func (e *Dropdown) onMouseButtonOutsideMenu(evt *Event) {
  if !e.IsHit(evt.X, evt.Y) {
    e.Root.Menu.Hide()
  }
}

func (e *Dropdown) menuVisible() bool {
  return e.Root.Menu.IsOwnedBy(e)
}

func (e *Dropdown) onMouseDown(evt *Event) {
  if e.menuVisible() {
    e.Root.Menu.Hide()
  } else {
    e.onShowMenu()
  }
}

func (e *Dropdown) focused() bool {
  return e.Root.FocusRect.IsOwnedBy(e.wrapper)
}

func (e *Dropdown) onFocus(evt *Event) {
  if evt.IsKeyboardEvent() {
    e.Root.FocusRect.Show(e.wrapper)
  }
}

func (e *Dropdown) onBlur(evt *Event) {
  if e.focused() {
    e.Root.FocusRect.Hide()
  }
}

func (e *Dropdown) onKeyDown(evt *Event) {
  if evt.Key == "space" || evt.Key == "return" {
    if e.menuVisible() {
      e.Root.Menu.ClickSelected()
    } else {
      e.onShowMenu()
    }
  } else if e.menuVisible() {
    if evt.Key == "down" {
      e.Root.Menu.SelectNext()
    } else if evt.Key == "up" {
      e.Root.Menu.SelectPrev()
    } else {
      e.Root.Menu.Hide()
    } 
  } else if evt.Key == "down" {
    i := e.valueIndex()

    if i == len(e.options) - 1 {
      e.SetValue(e.options[0])
    }  else {
      e.SetValue(e.options[i+1])
    }
  } else if evt.Key == "up" {
    i := e.valueIndex()
    
    if i <= 0 {
      e.SetValue(e.options[len(e.options)-1])
    } else {
      e.SetValue(e.options[i-1])
    }
  }
}

func (e *Dropdown) onKeyPress(evt *Event) {
  // TODO
}

func (e *Dropdown) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  return e.SetButtonPos(maxWidth, maxHeight, maxZIndex)
}

func (e *DropdownWrapper) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  panic("shouldn't be called")
}

func (e *DropdownWrapper) ZIndex() int {
  return e.dropdown.ZIndex()
}

func (e *DropdownWrapper) Rect() Rect {
  ddr := e.dropdown.Rect()

  if e.dropdown.menuVisible() {
    return Rect{
      ddr.X, 
      ddr.Y, 
      ddr.W, 
      ddr.H*(1 + len(e.dropdown.options)) + 2*e.Root.P1.Skin.ButtonBorderThickness(),
    }
  } else {
    return ddr
  }
}

func (e *Dropdown) Hide() {
  if e.menuVisible() {
    e.Root.Menu.Hide()
  }

  fmt.Println("hiding dropdown")

  e.ElementData.Hide()
}

func (e *Dropdown) Show() {
  e.SetButtonStyle()

  e.ElementData.Show()
}

func (e *Dropdown) SetValue(v string) {
  if v == "" {
    e.text.SetContent("Choose animal")
  } else {
    e.text.SetContent(v)
  }

  e.value = v
}

func (e *Dropdown) Value() string {
  return e.value
}

func (e *Dropdown) valueIndex() int {
  for i, opt := range e.options {
    if opt == e.value {
      return i
    }
  }
  
  return -1
}

func (e *Dropdown) onShowMenu() {
  menu := e.Root.Menu

  menu.ClearChildren()

  for _, option := range e.options {
    option_ := option
    menu.AddButton(option_, true, option_ == e.value, e.height, func() {
      e.SetValue(option_)
    })
  }

  e.Root.Menu.ShowAt(
    e,
    0.0,
    1.0,
    e.rect.W,
  )
}
