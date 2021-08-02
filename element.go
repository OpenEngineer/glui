package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

type Element interface {
  RegisterParent(parent Element)
  //A(e ...Element) Element // short for "AppendChildren", returns self

  Parent() Element
  Children() []Element

  Cursor() int

  // returns actual width and height used
  OnResize(maxWidth, maxHeight int) (int, int)

  Hit(x, y int) bool
  Translate(dx, dy int)

  GetEventListener(name string) EventListener // returns nil if no EventListener specified
}

type ElementData struct {
  parent       Element
  children     []Element

  bb           Rect
  evtListeners map[string]EventListener // only one eventlistener per event type

  // basic positioning settings
  padding [4]int
  spacing int
}

func newElementData() ElementData {
  return ElementData{
    nil,
    make([]Element, 0),
    Rect{0, 0, 0, 0},
    make(map[string]EventListener),
    [4]int{0, 0, 0, 0},
    0,
  }
}

func (e *ElementData) Cursor() int {
  return sdl.SYSTEM_CURSOR_ARROW
}

func (e *ElementData) appendChild(child Element) {
  e.children = append(e.children, child)
}

func (e *ElementData) RegisterParent(parent Element) {
  if e.parent != nil {
    panic("parent already registered")
  }

  e.parent = parent
}

func (e *ElementData) Padding(p ...int) {
  // TODO: trigger "dirty" or equivalent
  // TODO: return element
  switch len(p) {
  case 1:
    e.padding = [4]int{p[0], p[0], p[0], p[0]}
    break
  case 2:
    e.padding = [4]int{p[0], p[1], p[0], p[1]}
    break
  case 3:
    e.padding = [4]int{p[0], p[1], p[0], p[2]}
    break
  case 4:
    e.padding = [4]int{p[0], p[1], p[2], p[3]}
    break
  default:
    panic("unexpected number of padding elements")
  }
}

func (e *ElementData) Spacing(s int) {
  e.spacing = s
}

func (e *ElementData) GetEventListener(name string) EventListener {
  l, ok := e.evtListeners[name]
  if !ok {
    return nil
  } else {
    return l
  }
}

func (e *ElementData) SetEventListener(name string, evtListener EventListener) {
  e.evtListeners[name] = evtListener
}

func (e *ElementData) Children() []Element {
  return e.children
}

func (e *ElementData) Parent() Element {
  return e.parent
}

func (e *ElementData) Hit(x, y int) bool {
  return e.bb.Hit(x, y)
}

func (e *ElementData) InitBB(w, h int) (int, int) {
  e.bb = Rect{0, 0, w, h}

  return w, h
}

func (e *ElementData) Translate(dx, dy int) {
  for _, child := range e.children {
    child.Translate(dx, dy)
  }

  e.bb = e.bb.Translate(dx, dy)
}

// default positioning of children
// placement elements like Grid can provide better control
func (e *ElementData) resizeChildren(maxWidth, maxHeight int) {
  y := e.padding[0]

  for _, child := range e.children {
    _, dy := child.OnResize(maxWidth - e.padding[1] - e.padding[3], maxHeight - y - e.padding[2])

    child.Translate(e.padding[3], y)

    y += dy + e.spacing
  }
}

// returns true if new active element is same as old active element, or is child of old active element
func findActive(e Element, x, y int) (Element, bool) {
  if e.Hit(x, y) {
    for {
      childHit := false
      for _, c := range e.Children() {
        if c.Hit(x, y) {
          e = c
          childHit = true
          break
        }
      }

      if !childHit {
        return e, true // resulting element is still child of old active element, or same as old active element
      }
    }
  } else {
    p := e.Parent()
    if p == nil {
      return e, true
    } else {
      res, _ := findActive(p, x, y)
      return res, false
    }
  }
}

func collectAncestors(a Element) []Element {
  res := make([]Element, 0)

  for {
    a = a.Parent()

    if a != nil {
      res = append([]Element{a}, res...)
    } else {
      break
    }
  }

  return res
}

// should at least resolve to *Body
func commonAncestor(a Element, b Element) Element {
  if a == b {
    return a
  }

  if _, aIsBody := a.(*Body); aIsBody {
    return a
  } else if _, bIsBody := b.(*Body); bIsBody {
    return b
  }

  aps := collectAncestors(a)
  bps := collectAncestors(b)

  for i := 1; i < len(aps) && i < len(bps); i++ {
    if aps[i] != bps[i] {
      return aps[i-1]
    }
  }

  if len(aps) < len(bps) {
    return aps[len(aps)-1]
  } else {
    return bps[len(bps)-1]
  }
}
