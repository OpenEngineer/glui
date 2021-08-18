package glui

// gives each element a unique depth index, or z-index
type ElementStack struct {
  stack []Element

  offset int

  dirty bool
}

func newElementStack() *ElementStack {
  return &ElementStack{
    make([]Element, 0),
    0,
    true,
  }
}

func (s *ElementStack) contains(dep Element) bool {
  for _, e := range s.stack {
    if e == dep {
      return true
    }
  }

  return false
}

func (s *ElementStack) Add(e Element, deps []Element) int {
  for _, d := range deps {
    if !s.contains(d) {
      s.dirty = true

      return -1
    }
  }

  id := len(s.stack)

  s.stack = append(s.stack, e)

  return id + s.offset
}

func (s *ElementStack) maxZIndex() int {
  return len(s.stack) + s.offset
}
