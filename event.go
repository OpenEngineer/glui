package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

type EventListener func(evt *Event)

type Event struct {
  X int // mouse X pos (left edge of window is 0)
  Y int // mouse Y pos (top edge of window is 0)

  Key   string // empty if not a key event
  Ctrl  bool
  Shift bool
  Alt   bool

  Value string // for text Input

  stopBubblingElement Element // exclusive
  stopBubbling bool
}

func NewMouseEvent(x, y int) *Event {
  if x < 0 {
    x_, y_, _ := sdl.GetMouseState()

    x = int(x_)
    y = int(y_)
  }

  return &Event{x, y, "", false, false, false, "", nil, false}
}

func NewKeyboardEvent(keyName string, ctrl bool, shift bool, alt bool) *Event {
  return &Event{0, 0, keyName, ctrl, shift, alt, "", nil, false}
}

func NewTextInputEvent(str string) *Event {
  return &Event{0, 0, "", false, false, false, str, nil, false}
}

func (e *Event) StopBubbling() {
  e.stopBubbling = true
}

func (e *Event) stopBubblingWhenElementReached(stopEl Element) {
  e.stopBubblingElement = stopEl
}

func (e *Event) IsMouseEvent() bool {
  return !e.IsKeyboardEvent() && !e.IsTextInputEvent()
}

func (e *Event) IsTextInputEvent() bool {
  return e.Value != "" 
}

func (e *Event) IsKeyboardEvent() bool {
  return e.Key != ""
}

func extractKeyboardEventDetails(event *sdl.KeyboardEvent) (string, string, bool, bool, bool) {
  eType := ""
  if event.State == sdl.RELEASED {
    eType = "keyup"
  } else if event.State == sdl.PRESSED {
    if event.Repeat != 0 {
      eType = "keypress"
    } else {
      eType = "keydown"
    }
  }

  kType := ""
  switch event.Keysym.Sym {
  case sdl.K_a:
    kType = "a"
    break
  case sdl.K_c:
    kType = "c"
    break
  case sdl.K_v:
    kType = "v"
    break
  case sdl.K_x:
    kType = "x"
    break
  case sdl.K_SPACE:
    kType = "space"
    break
  case sdl.K_BACKSPACE:
    kType = "backspace"
    break
  case sdl.K_DELETE:
    kType = "delete"
    break
  case sdl.K_LEFT:
    kType = "left"
    break
  case sdl.K_RIGHT:
    kType = "right"
    break
  case sdl.K_DOWN:
    kType = "down"
    break
  case sdl.K_UP:
    kType = "up"
    break
  case sdl.K_HOME:
    kType = "home"
    break
  case sdl.K_END:
    kType = "end"
    break
  case sdl.K_RETURN, sdl.K_RETURN2:
    kType = "return"
    break
  }

  shift := (event.Keysym.Mod & sdl.KMOD_SHIFT > 0)
  ctrl := (event.Keysym.Mod & sdl.KMOD_CTRL > 0)
  alt := (event.Keysym.Mod & sdl.KMOD_ALT > 0)

  return eType, kType, ctrl, shift, alt
}
