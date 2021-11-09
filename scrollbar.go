package glui

import (
)

//go:generate ./gen_element Scrollbar "CalcDepth appendChild On"

// has two children: the up/down or left/right buttons
// the slider and track are managed by the scrollbar itself
// the scrollbar can grab focus
type Scrollbar struct {
  ElementData

  orientation Orientation

  sliderSize  int
  sliderStart int // 0 for slider that is at top/left-most position
  lineHeight  int // move per up/down click

  lastDown    int
  lastDownSS  int

  // TODO: callbacks
}

func NewScrollbar(orientation Orientation) *Scrollbar {
  e := &Scrollbar{
    NewElementData(10*2, 0), // first 9 quads are the slider, last quad is the track
    orientation,
    50,
    0,
    10,
    -1, -1,
  }

  b1 := NewIconButton("arrow-up-drop", 10, e.orientation.Rotate()).Size(e.size(), e.size())
  b2 := NewIconButton("arrow-down-drop", 10, e.orientation.Rotate()).Size(e.size(), e.size())

  e.appendChild(b1, b2)

  e.setStyleAndTCoords()

  b1.OnClick(e.onUpClick)
  b2.OnClick(e.onDownClick)

  b1.On("doubleclick", func(evt *Event) {
    // we already moved by 'e.lineHeight' upon single click
    e.MoveBy(-(e.sliderSize - e.lineHeight))
  })
  
  b2.On("doubleclick", func(evt *Event) {
    // we already moved by 'e.lineHeight' upon single click
    e.MoveBy(e.sliderSize - e.lineHeight)
  })

  e.On("click", e.onMouseClick)
  e.On("mousedown", e.onMouseDown)
  e.On("mousemove", e.onMouseMove)
  e.On("mouseup", e.onMouseUp)
  e.On("focus", e.onFocus)
  e.On("blur", e.onBlur)

  return e
}

func (e *Scrollbar) onFocus(evt *Event) {
  if evt.IsKeyboardEvent() {
    e.Root.FocusRect.Show(e)
  }
}

func (e *Scrollbar) onBlur(evt *Event) {
  e.Root.FocusRect.Hide()
}

func (e *Scrollbar) size() int {
  return e.Root.P1.Skin.ScrollbarTrackSize()
}

func (e *Scrollbar) setStyleAndTCoords() {
  e.SetButtonStyle()

  // track styling
  tri0 := e.p1Tris[9*2]
  tri1 := e.p1Tris[9*2+1]

  x0, y0 := e.Root.P1.Skin.ScrollbarTrackOrigin()

  x := [4]int{x0, x0+1, 0, 0}
  y := [4]int{y0, y0 + e.size(), 0, 0}

  e.Root.P1.SetTriType(tri0, VTYPE_SKIN)
  e.Root.P1.SetTriType(tri1, VTYPE_SKIN)

  e.Root.P1.Color.Set4Const(tri0, 1.0, 1.0, 1.0, 1.0)
  e.Root.P1.Color.Set4Const(tri1, 1.0, 1.0, 1.0, 1.0)

  if e.orientation == HOR {
    e.Root.P1.setQuadSkinCoords(tri0, tri1, 0, 0, x, y)
  } else {
    e.Root.P1.setQuadSkinCoordsT(tri0, tri1, 0, 0, x, y)
  }
}

func (e *Scrollbar) Show() {
  e.setStyleAndTCoords()

  e.ElementData.Show()
}

func (e *Scrollbar) buttons() (*Button, *Button) {
  b1, ok := e.children[0].(*Button)
  if !ok {
    panic("expected button")
  }

  b2, ok := e.children[1].(*Button)
  if !ok {
    panic("expected button")
  }

  return b1, b2
}

func (e *Scrollbar) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  b1, b2 := e.buttons()

  b1.CalcPos(maxWidth, maxHeight, maxZIndex)
  b2.CalcPos(maxWidth, maxHeight, maxZIndex)

  t := e.Root.P1.Skin.ButtonBorderThickness()
  dz := -0.5/float32(maxZIndex) // slider must be closer to viewer than track

  // track tris
  tri0 := e.p1Tris[9*2]
  tri1 := e.p1Tris[9*2+1]

  if e.orientation == HOR {
    e.width = maxWidth
    b2.Translate(maxWidth - e.size(), 0)

    // slider
    w := e.sliderSize
    h := e.size()

    e.limitSliderPos()

    e.SetBorderedElementPos(w, h, t, maxZIndex)
    for _, tri := range e.p1Tris {
      if tri == tri0 {
        break
      }

      e.Root.P1.TranslateTri(tri, e.size() + e.sliderStart, 0.0, dz)
    }

    // track
    e.Root.P1.SetQuadPos(tri0, tri1, Rect{e.size(), 0, maxWidth - 2*e.size(), e.size()}, e.Z(maxZIndex))

    return e.InitRect(maxWidth, e.size())
  } else {
    e.height = maxHeight
    b2.Translate(0, maxHeight - e.size())

    w := e.size()
    h := e.sliderSize

    e.limitSliderPos()

    e.SetBorderedElementPos(w, h, t, maxZIndex)
    for _, tri := range e.p1Tris {
      if tri == tri0 {
        break
      }

      e.Root.P1.TranslateTri(tri, 0.0, e.size() + e.sliderStart, dz)
    }

    // track
    e.Root.P1.SetQuadPos(tri0, tri1, Rect{0, e.size(), e.size(), maxHeight - 2*e.size()}, e.Z(maxZIndex))

    return e.InitRect(e.size(), maxHeight)
  }
}

func (e *Scrollbar) onUpClick() {
  e.MoveBy(-e.lineHeight)
}

func (e *Scrollbar) onDownClick() {
  e.MoveBy(e.lineHeight)
}

func (e *Scrollbar) SliderRect() Rect {
  thisRect := e.Rect()

  var slRect Rect 

  if e.orientation == HOR {
    slRect = Rect{
      thisRect.X + e.size() + e.sliderStart,
      thisRect.Y,
      e.sliderSize,
      thisRect.H,
    }
  } else {
    slRect = Rect{
      thisRect.X,
      thisRect.Y + e.size() + e.sliderStart,
      thisRect.W,
      e.sliderSize,
    }
  }

  return slRect
}

func (e *Scrollbar) TrackRects() (Rect, Rect) {
  thisRect := e.Rect()

  var (
    rUp      Rect
    rDown    Rect
  )

  if e.orientation == HOR {
    rUp = Rect{
      thisRect.X + e.size(),
      thisRect.Y,
      e.sliderStart,
      thisRect.H,
    }

    rDown = Rect{
      thisRect.X + e.size() + e.sliderStart + e.sliderSize,
      thisRect.Y,
      thisRect.W - 2*e.size() - e.sliderStart - e.sliderSize,
      thisRect.H,
    }
  } else {
    rUp = Rect{
      thisRect.X,
      thisRect.Y + e.size(),
      thisRect.W,
      e.sliderStart,
    }

    rDown = Rect{
      thisRect.X,
      thisRect.Y + e.size() + e.sliderStart + e.sliderSize,
      thisRect.W,
      thisRect.H - 2*e.size() - e.sliderStart - e.sliderSize,
    }
  }

  return rUp, rDown
}

func (e *Scrollbar) trackPos(evt *Event) int {
  var p int

  if e.orientation == HOR {
    p = evt.X - (e.Rect().X + e.size())
  } else {
    p = evt.Y - (e.Rect().Y + e.size())
  }

  return p
}

func (e *Scrollbar) onMouseClick(evt *Event) {
  p := e.trackPos(evt)

  e.MoveTo(p - e.sliderSize/2)
}

func (e *Scrollbar) getTrackLength() int {
  if e.orientation == HOR {
    return e.width - 2*e.size()
  } else {
    return e.height - 2*e.size()
  }
}

func (e *Scrollbar) limitSliderPos() {
  if e.sliderStart + e.sliderSize > e.getTrackLength() {
    e.sliderStart = e.getTrackLength() - e.sliderSize
  }

  if e.sliderStart < 0 {
    e.sliderStart = 0
  }
}

func (e *Scrollbar) MoveTo(ss int) {
  e.sliderStart = ss

  e.limitSliderPos()

  e.Root.ForcePosDirty()
}

func (e *Scrollbar) MoveBy(d int) {
  e.MoveTo(e.sliderStart + d)
}

func (e *Scrollbar) onMouseDown(evt *Event) {
  slRect := e.SliderRect()

  if !slRect.Hit(evt.X, evt.Y) {
    e.onMouseClick(evt)
  }

  e.lastDown = e.trackPos(evt)
  e.lastDownSS = e.sliderStart
}

func (e *Scrollbar) onMouseUp(evt *Event) {
  if e.lastDown < 0 {
    return
  }

  d := e.lastDown - e.trackPos(evt)
  if d > 1 || d < -1 {
    // don't trigger the click
    evt.StopPropagation()
  }

  e.lastDown = -1
  e.lastDownSS = -1
}

func (e *Scrollbar) onMouseMove(evt *Event) {
  if e.lastDown < 0 {
    return
  }

  p := e.trackPos(evt)

  e.MoveTo(e.lastDownSS + p - e.lastDown)
}
