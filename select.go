package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

//go:generate ./gen_element Select "appendChild CalcDepth On Size Padding"

//go:generate ./gen_element SelectWrapper "CalcDepth"

type SelectWrapper struct { // given to focusrect instead of actual wrapper
  ElementData

  sel *Select
}

type Select struct {
  ElementData

  options []string
  text   *Text
  arrow  *Icon
  wrapper *SelectWrapper

  value string
}


func NewSelect(options []string) *Select {
  e := &Select{
    NewElementData(9*2, 0),
    options, 
    NewSans("Choose animal", 10), 
    NewIcon("arrow-down-drop", 10),
    nil,
    "",
  }
  
  e.wrapper = &SelectWrapper{NewElementData(0, 0), e}

  e.Size(200, 50)
  e.Padding(e.Root.P1.Skin.ButtonBorderThickness())

  e.appendChild(NewHor(STRETCH, CENTER, 0).H(-1).Padding(0, 10).A(e.text, e.arrow))
  e.Show()

  e.On("mousedown", e.onMouseDown)
  e.On("mousebuttonoutsidemenu", e.onMouseButtonOutsideMenu)
  e.On("focus", e.onFocus)
  e.On("blur", e.onBlur)
  e.On("keydown", e.onKeyDown)
  e.On("keypress", e.onKeyPress)

  return e
}

func (e *Select) Cursor() int {
  if e.enabled {
    return sdl.SYSTEM_CURSOR_HAND
  } else {
    return -1
  }
}

func (e *Select) onMouseButtonOutsideMenu(evt *Event) {
  if !e.IsHit(evt.X, evt.Y) {
    e.Root.Menu.Hide()
  }
}

func (e *Select) menuVisible() bool {
  return e.Root.Menu.IsOwnedBy(e)
}

func (e *Select) onMouseDown(evt *Event) {
  if e.menuVisible() {
    e.Root.Menu.Hide()
  } else {
    e.onShowMenu()
  }
}

func (e *Select) focused() bool {
  return e.Root.FocusRect.IsOwnedBy(e.wrapper)
}

func (e *Select) onFocus(evt *Event) {
  if evt.IsKeyboardEvent() {
    e.Root.FocusRect.Show(e.wrapper)
  }
}

func (e *Select) onBlur(evt *Event) {
  if e.focused() {
    e.Root.FocusRect.Hide()
  }
}

func (e *Select) onKeyDown(evt *Event) {
  if evt.IsReturnOrSpace() {
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

func (e *Select) onKeyPress(evt *Event) {
  // TODO
}

func (e *Select) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  return e.SetButtonPos(maxWidth, maxHeight, maxZIndex)
}

func (e *SelectWrapper) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  panic("shouldn't be called")
}

func (e *SelectWrapper) ZIndex() int {
  return e.sel.ZIndex()
}

func (e *SelectWrapper) Visible() bool {
  return e.sel.Visible()
}

func (e *SelectWrapper) Deleted() bool {
  return e.sel.Deleted()
}

func (e *SelectWrapper) Rect() Rect {
  ddr := e.sel.Rect()

  if e.sel.menuVisible() {
    return Rect{
      ddr.X, 
      ddr.Y, 
      ddr.W, 
      ddr.H*(1 + len(e.sel.options)) + 2*e.Root.P1.Skin.ButtonBorderThickness(),
    }
  } else {
    return ddr
  }
}

func (e *Select) Hide() {
  if e.menuVisible() {
    e.Root.Menu.Hide()
  }

  e.ElementData.Hide()
}

func (e *Select) Show() {
  e.SetButtonStyle()

  e.ElementData.Show()
}

func (e *Select) SetValue(v string) {
  if v == "" {
    e.text.SetContent("Choose animal")
  } else {
    e.text.SetContent(v)
  }

  e.value = v
}

func (e *Select) Value() string {
  return e.value
}

func (e *Select) valueIndex() int {
  for i, opt := range e.options {
    if opt == e.value {
      return i
    }
  }
  
  return -1
}

func (e *Select) onShowMenu() {
  menu := e.Root.Menu

  menu.ClearChildren()

  for _, option := range e.options {
    option_ := option

    item := NewMenuItem(option_, func() {
      e.SetValue(option_)
    }).H(e.height)

    menu.AddItem(item, true, option_ == e.value)
  }

  e.Root.Menu.ShowAt(
    e,
    0.0,
    1.0,
    e.rect.W,
  )
}
