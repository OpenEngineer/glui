package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

// wrapper for Body, Menu and FocusRect
type Root struct {
  w         int
  h         int
  maxZIndex int

  P1        *DrawPass1Data
  P2        *DrawPass2Data

  Body      *Body
  Menu      *Menu
  FocusRect *FocusRect
}

// skinmap and glyphmap can be shared across multiple windows/frames/layers
func newRoot(skin *SkinMap, glyphs *GlyphMap) *Root {
  r := &Root{
    0, 0, 0,
    newDrawPass1Data(skin), newDrawPass2Data(glyphs),
    nil, nil, nil,
  }

  r.Body      = newBody(r)
  r.Menu      = newMenu(r)
  r.FocusRect = newFocusRect(r)

  return r
}

func (e *Root) syncSize(window *sdl.Window) {
  w_, h_ := window.GLGetDrawableSize()
  w, h := int(w_), int(h_)

  e.w, e.h = w, h
  e.P1.w, e.P1.h = w, h
  e.P2.w, e.P2.h = w, h
}

func (e *Root) initGL(prog1 uint32, prog2 uint32) {
  e.P1.InitGL(prog1)
  e.P2.InitGL(prog2)
}

func (e *Root) dirty() bool {
  return e.P1.dirty() || e.P2.dirty()
}

func (e *Root) posDirty() bool {
  return e.P1.posDirty() || e.P2.posDirty()
}

func (e *Root) ForcePosDirty() {
  e.P1.Type.dirty = true
}

func (e *Root) GetSize() (int, int) {
  return e.w, e.h
}

func (e *Root) CalcDepth() {
  stack := newElementStack()

  for ; stack.dirty; {
    stack.dirty = false

    e.Body.CalcDepth(stack)
  }

  stack.dirty = true

  // add some offset for menu, so it definitely lies above all body elements
  // XXX: why?
  stack.offset = 2*stack.maxZIndex()

  for ; stack.dirty; {
    stack.dirty = false

    e.Menu.CalcDepth(stack)
  }

  e.maxZIndex = stack.maxZIndex()
}

func (e *Root) CalcPos() {
  if e.maxZIndex == 0 {
    panic("depth not yet calculated")
  }

  e.Body.CalcPos(e.w, e.h, e.maxZIndex)

  e.Menu.CalcPos(e.w, e.h, e.maxZIndex)

  e.FocusRect.CalcPos(e.w, e.h, e.maxZIndex)
}

func (e *Root) Animate(tick uint64) {
  e.Body.Animate(tick)

  e.Menu.Animate(tick)

  e.FocusRect.Animate(tick)
}

func (e *Root) findMouseElement(oldMouseElement Element, x, y int) (Element, bool) {
  if e.Menu.IsHit(x, y) {
    if oldMouseElement == nil {
      oldMouseElement = e.Menu
    } else if !hasAncestor(oldMouseElement, e.Menu) {
      oldMouseElement = e.Menu

      newMouseElement, _ := findHitElement(oldMouseElement, x, y)

      return newMouseElement, false
    }

    return findHitElement(oldMouseElement, x, y)
  } else {
    if oldMouseElement == nil {
      oldMouseElement = e.Body
    } else if !hasAncestor(oldMouseElement, e.Body) {
      oldMouseElement = e.Body

      newMouseElement, _ := findHitElement(oldMouseElement, x, y)

      return newMouseElement, false
    }

    return findHitElement(oldMouseElement, x, y)
  }
}
