package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

type EventListener func(evt *Event)

type Event struct {
  X int // mouse X pos (left edge of window is 0)
  Y int // mouse Y pos (top edge of window is 0)

  stopBubblingElement Element // exclusive
  stopBubbling bool
}

func NewMouseEvent(x, y int) *Event {
  if x < 0 {
    x_, y_, _ := sdl.GetMouseState()

    x = int(x_)
    y = int(y_)
  }

  return &Event{x, y, nil, false}
}

func (e *Event) StopBubbling() {
  e.stopBubbling = true
}

func (e *Event) stopBubblingWhenElementReached(stopEl Element) {
  e.stopBubblingElement = stopEl
}
