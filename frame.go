package glui

import (
)

// wrapper for Body, Menu and FocusRect
type Frame struct {
  winW      int
  winH      int
  maxW      int
  maxH      int
  maxZIndex int

  P1        *DrawPass1Data
  P2        *DrawPass2Data

  Body      *Body
  Menu      *Menu
  FocusRect *FocusRect

  state     *FrameState
}

// skinmap and glyphmap can be shared across multiple windows/frames/layers
func newFrame(isFirst bool, skin *SkinMap, glyphs *GlyphMap) *Frame {
  frame := &Frame{
    0, 0, 0, 0, 0,
    newDrawPass1Data(skin), newDrawPass2Data(glyphs),
    nil, nil, nil, newFrameState(),
  }

  frame.Body      = newBody(frame, isFirst)
  frame.Menu      = newMenu(frame)
  frame.FocusRect = newFocusRect(frame)

  return frame
}

func (e *Frame) syncWindowSize(winW, winH int) {
  e.winW, e.winH = winW, winH

  e.P1.winW, e.P1.winH = winW, winH
  e.P2.winW, e.P2.winH = winW, winH

  e.ForcePosDirty()
}

func (e *Frame) dirty() bool {
  return e.P1.dirty() || e.P2.dirty()
}

func (e *Frame) posDirty() bool {
  return e.P1.posDirty() || e.P2.posDirty()
}

func (e *Frame) clearPosDirty() {
  e.P1.clearPosDirty()
  e.P2.clearPosDirty()
}

func (e *Frame) ForcePosDirty() {
  e.P1.forcePosDirty()
  e.P2.forcePosDirty()
}

// needed when switching frames
func (e *Frame) ForceAllDirty() {
  e.P1.ForceAllDirty()
  e.P2.ForceAllDirty()
}

func (e *Frame) GetPos() (int, int) {
  x, y := 0, 0

  if e.maxW < e.winW {
    x = (e.winW - e.maxW)/2
  }

  if e.maxH < e.winH {
    y = (e.winH - e.maxH)/2
  }

  return x, y
}

func (e *Frame) GetSize() (int, int) {
  w, h := e.winW, e.winH

  if w > e.maxW {
    w = e.maxW
  }

  if h > e.maxH {
    h = e.maxH
  }

  return w, h
}

func (e *Frame) show() {
  e.Body.Show()
}

func (e *Frame) CalcDepth() {
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

func (e *Frame) CalcPos() {
  if e.maxZIndex == 0 {
    panic("depth not yet calculated")
  }

  x, y := e.GetPos()
  w, h := e.GetSize()

  e.Body.CalcPos(w, h, e.maxZIndex)

  e.Menu.CalcPos(w, h, e.maxZIndex)

  e.FocusRect.CalcPos(w, h, e.maxZIndex)

  if x != 0 || y != 0 {
    e.Body.Translate(x, y)
    e.Menu.Translate(x, y)
    e.FocusRect.Translate(x, y)
  }

  e.P1.SyncImagesToTexture()
}

func (e *Frame) Animate(tick uint64) {
  e.Body.Animate(tick)

  e.Menu.Animate(tick)

  e.FocusRect.Animate(tick)
}

// if this function returns `false` incorrectly, then oldMouseElement probably doesnt correctly have Body as ancestor
func (e *Frame) findMouseElement(oldMouseElement Element, x, y int) (Element, bool) {
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

func (e *Frame) Clear() {
  e.Body.ClearChildren()
  e.Menu.ClearChildren()
}

func (e *Frame) CurrentTick() uint64 {
  return e.state.lastTick
}
