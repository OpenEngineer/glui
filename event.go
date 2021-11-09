package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

type EventListener func(evt *Event)

type Event struct {
  X int // mouse X pos (left edge of window is 0)
  Y int // mouse Y pos (top edge of window is 0)

  // mouse movement since last event, 0 if irrelevant
  XRel int 
  YRel int

  Key   string // empty if not a key event
  Ctrl  bool
  Shift bool
  Alt   bool

  Value string // for text Input
  AppMsg string // for quit

  stopBubblingElement Element // exclusive
  stopBubbling bool

  stopPropagation bool // multiple functions can be tied to a single eventlisteners, this stops older functions from being called
  callback  func(args ...interface{}) // for async quit
}

func currentMousePos() (int, int) {
  x_, y_, _ := sdl.GetMouseState()

  x := int(x_)
  y := int(y_)

  return x, y
}

func NewMouseEvent(x, y int) *Event {
  if x < 0 {
    x, y = currentMousePos()
  }

  ks := sdl.GetKeyboardState()
  ctrl := ks[sdl.SCANCODE_LCTRL] > 0 || ks[sdl.SCANCODE_RCTRL] > 0
  shift := ks[sdl.SCANCODE_LSHIFT] > 0 || ks[sdl.SCANCODE_RSHIFT] > 0
  alt := ks[sdl.SCANCODE_LALT] > 0 || ks[sdl.SCANCODE_RALT] > 0

  return &Event{x, y, 0, 0, "", ctrl, shift, alt, "", "", nil, false, false, nil}
}

func NewMouseMoveEvent(x, y int, dx, dy int) *Event {
  e := NewMouseEvent(x, y)
  e.XRel = dx
  e.YRel = dy

  return e
}

func NewKeyboardEvent(keyName string, ctrl bool, shift bool, alt bool) *Event {
  return &Event{0, 0, 0, 0, keyName, ctrl, shift, alt, "", "", nil, false, false, nil}
}

func NewTextInputEvent(str string) *Event {
  return &Event{0, 0, 0, 0, "", false, false, false, str, "", nil, false, false, nil}
}

func NewAppEvent(msg string, fn func(args ...interface{})) *Event {
  return &Event{0, 0, 0, 0, "", false, false, false, "", msg, nil, false, false, fn}
}

func (e *Event) StopBubbling() {
  e.stopBubbling = true
}

func (e *Event) StopPropagation() {
  e.stopPropagation = true
}

func (e *Event) Callback(args ...interface{}) {
  e.callback(args...)
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

func (e *Event) IsReturnOrSpace() bool {
  return e.Key == "space" || e.Key == "return"
}

func (e *Event) IsTab() bool {
  return e.Key == "tab"
}

func (e *Event) IsEscape() bool {
  return e.Key == "escape"
}

func (e *Event) RelPos(r Rect) (int, int) {
  return e.X - r.X, e.Y - r.Y
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
  case sdl.K_ESCAPE:
    kType = "escape"
    break
  case sdl.K_PAGEDOWN:
    kType = "pagedown"
    break
  case sdl.K_PAGEUP:
    kType = "pageup"
    break
  }

  shift := (event.Keysym.Mod & sdl.KMOD_SHIFT > 0)
  ctrl := (event.Keysym.Mod & sdl.KMOD_CTRL > 0)
  alt := (event.Keysym.Mod & sdl.KMOD_ALT > 0)

  return eType, kType, ctrl, shift, alt
}
