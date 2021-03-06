package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

// base type for all elements
type ElementData struct {
  parent       Element
  children     []Element

  Root       *Frame
  p1Tris     []uint32
  p2Tris     []uint32
  closerThan []Element // these elements must get a smaller z-index than self (i.e. be further away from viewer)

  evtListeners map[string]EventListener // only one eventlistener per event type

  // basic positioning settings
  width   int
  height  int
  padding [4]int
  spacing int

  // state
  rect    Rect
  zIndex  int // returned by succesful Hit test, must be normalized before using in Pos
  visible bool
  enabled bool
  deleted bool
}

func newElementData(frame *Frame, nInitTris1 int, nInitTris2 int) ElementData {
  p1Tris := make([]uint32, 0)
  if nInitTris1 > 0 {
    p1Tris = frame.P1.Alloc(nInitTris1)
  }

  p2Tris := make([]uint32, 0)
  if nInitTris2 > 0 {
    p2Tris = frame.P2.Alloc(nInitTris2)
  }

  return ElementData{
    nil,
    make([]Element, 0),
    frame, p1Tris, p2Tris,
    make([]Element, 0),
    make(map[string]EventListener),
    0, 0, [4]int{0, 0, 0, 0}, 0,
    Rect{0, 0, 0, 0}, -1, true, true, false,
  }
}
func NewElementData(nInitTris1 int, nInitTris2 int) ElementData {
  frame := ActiveFrame()

  return newElementData(frame, nInitTris1, nInitTris2)
}

func (e *ElementData) ZIndex() int {
  return e.zIndex
}

func (e *ElementData) Visible() bool {
  return e.visible
}

func (e *ElementData) nChildren() int {
  return len(e.children)
}

func (e *ElementData) nVisibleChildren() int {
  c := 0

  for _, child := range e.children {
    if child.Visible() {
      c += 1
    }
  }

  return c
}

func (e *ElementData) visibleChildren() []Element {
  els := make([]Element, 0)

  for _, child := range e.children {
    if child.Visible() {
      els = append(els, child)
    }
  }

  return els
}

func (e *ElementData) GetSize() (int, int) {
  return e.width, e.height
}

func (e *ElementData) Hide() {
  for _, tri := range e.p1Tris {
    e.Root.P1.SetTriType(tri, VTYPE_HIDDEN)
  }

  for _, tri := range e.p2Tris {
    e.Root.P2.SetTriType(tri, VTYPE_HIDDEN)
  }

  for _, child := range e.children {
    child.Hide()
  }

  e.visible = false
}

func (e *ElementData) Show() {
  for _, child := range e.children {
    child.Show()
  }

  e.visible = true
}

func (e *ElementData) Enable() {
  for _, child := range e.children {
    child.Enable()
  }

  e.enabled = true
}

func (e *ElementData) Disable() {
  for _, child := range e.children {
    child.Disable()
  }

  e.enabled = false
}

func (e *ElementData) ButtonCursor(x, y int, enabled bool) int {
  if enabled {
    return sdl.SYSTEM_CURSOR_HAND
  } else {
    return e.Cursor(x, y)
  }
}

func (e *ElementData) Cursor(x, y int) int {
  return -1
}

func (e *ElementData) Tooltip() string {
  return ""
}

func (e *ElementData) RegisterParent(parent Element) {
  if parent == nil {
    e.parent = nil
  } else if e.parent != nil {
    panic("parent already registered")
  } else {
    e.parent = parent
  }
}

func (e *ElementData) GetEventListener(name string) EventListener {
  l, ok := e.evtListeners[name]
  if !ok {
    return nil
  } else {
    return l
  }
}

func (e *ElementData) Children() []Element {
  return e.children
}

func (e *ElementData) Parent() Element {
  return e.parent
}

func (e *ElementData) IsHit(x, y int) bool {
  return e.visible && e.rect.Hit(x, y) 
}

// -1 -> no hit
func (e *ElementData) Hit(x, y int) int {
  if e.IsHit(x, y) {
    return e.zIndex
  } else {
    return -1
  }
}

func (e *ElementData) InitRect(w, h int) (int, int) {
  e.rect = Rect{0, 0, w, h}

  return w, h
}

func (e *ElementData) Rect() Rect {
  return e.rect
}

func (e *ElementData) Translate(dx, dy int) {
  for _, tri := range e.p1Tris {
    e.Root.P1.TranslateTri(tri, dx, dy, 0.0)
  }

  for _, tri := range e.p2Tris {
    e.Root.P2.TranslateTri(tri, dx, dy, 0.0)
  }

  for _, child := range e.children {
    child.Translate(dx, dy)
  }

  e.rect = e.rect.Translate(dx, dy)
}

// default positioning of children
// placement elements like Hor can provide better control
func (e *ElementData) CalcPosChildren(maxWidth, maxHeight, maxZIndex int) (int, int) {
  y := e.padding[0]

  maxW := 0
  for _, child := range e.children {
    if child.Visible() {
      w, dy := child.CalcPos(maxWidth - e.padding[1] - e.padding[3], maxHeight - y - e.padding[2], maxZIndex)

      child.Translate(e.padding[3], y)

      y += dy + e.spacing

      if w > maxW {
        maxW = w
      }
    }
  }

  return maxW + e.padding[1] + e.padding[3], y + e.padding[2]
}

func (e *ElementData) Animate(tick uint64) {
  for _, child := range e.children {
    child.Animate(tick)
  }
}

func (e *ElementData) Delete() {
  for _, child := range e.children {
    child.Delete()
  }

  e.Root.P1.Dealloc(e.p1Tris)
  e.Root.P2.Dealloc(e.p2Tris)

  e.deleted = true
  e.p1Tris = nil
  e.p2Tris = nil
  e.parent = nil

  e.Root.ForcePosDirty()

  e.Root = nil // so error is thrown if deletion is tried again
}

func (e *ElementData) Deleted() bool {
  return e.deleted
}

func (e *ElementData) ClearChildren() {
  for _, child := range e.children {
    child.Delete()
  }

  e.children = []Element{}
}

func normalizeZIndex(idx int, maxZIndex int) float32 {
  return float32(maxZIndex - idx)/float32(maxZIndex)
}

func (e *ElementData) Z(maxZIndex int) float32 {
  return normalizeZIndex(e.zIndex, maxZIndex)
}

func (e *ElementData) SetButtonStyle() {
  e.Root.P1.setButtonStyle(e.p1Tris)
}

func (e *ElementData) SetBorderedElementPos(w, h, t, maxZIndex int) {
  e.Root.P1.setBorderedElementPos(e.p1Tris, w, h, t, e.Z(maxZIndex))
}

func (e *ElementData) SetButtonPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  t := e.Root.P1.Skin.ButtonBorderThickness()

  w, h := e.GetSize()

  if w > maxWidth {
    w = maxWidth
  }

  if h > maxHeight {
    h = maxHeight
  }

  e.SetBorderedElementPos(w, h, t, maxZIndex)

  e.CalcPosChildren(w, h, maxZIndex)

  return e.InitRect(w, h)
}

func (e *ElementData) IsFocusable() bool {
  return e.GetEventListener("focus") != nil && e.Visible()
}

func (e *ElementData) Crop(r Rect) {
  e.rect = e.rect.Common(r)

  for _, tri := range e.p1Tris {
    e.Root.P1.CropTri(tri, r)
  }

  for _, tri := range e.p2Tris {
    e.Root.P2.CropTri(tri, r)
  }

  for _, child := range e.children {
    child.Crop(r)
  }
}
