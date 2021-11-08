package glui

//go:generate ./gen_element Checkbox "On CalcDepth"

type Checkbox struct {
  ElementData

  value bool
}

func NewCheckbox() *Checkbox {
  e := &Checkbox{
    NewElementData(9*2, 0),
    false,
  }

  outerSize := 2*e.borderT() + e.Root.P1.Skin.TickSize()
  e.width, e.height = outerSize, outerSize

  e.setTypesAndTCoords()

  e.On("keypress", e.onKeyPress)
  e.On("focus", e.onFocus)
  e.On("blur", e.onBlur)
  e.On("click", e.onMouseClick)

  return e
}

func (e *Checkbox) onFocus(evt *Event) {
  if evt.IsKeyboardEvent() {
    e.Root.FocusRect.Show(e)
  }
}

func (e *Checkbox) onBlur(evt *Event) {
  e.Root.FocusRect.Hide()
}

func (e *Checkbox) onKeyPress(evt *Event) {
  if evt.IsReturnOrSpace() {
    e.toggleAndUpdate()
  }
}

func (e *Checkbox) borderT() int {
  return e.Root.P1.Skin.InputBorderThickness()
}

func (e *Checkbox) getSkinCoords() ([4]int, [4]int) {
  var (
    x [4]int
    y [4]int
  )

  x0, y0 := e.Root.P1.Skin.TickOrigin()
  innerSize := e.Root.P1.Skin.TickSize()

  x[0] = x0
  x[1] = x0 + innerSize

  y[0] = y0
  y[1] = y0 + innerSize

  return x, y
}

func (e *Checkbox) setTypesAndTCoords() {
  e.Root.P1.setInputLikeElementTypesAndTCoords(e.p1Tris)

  if e.value {
    tri0 := e.p1Tris[8]
    tri1 := e.p1Tris[9]

    e.Root.P1.SetTriType(tri0, VTYPE_SKIN)
    e.Root.P1.SetTriType(tri1, VTYPE_SKIN)

    e.Root.P1.Color.Set4Const(tri0, 1.0, 1.0, 1.0, 1.0)
    e.Root.P1.Color.Set4Const(tri1, 1.0, 1.0, 1.0, 1.0)

    x, y := e.getSkinCoords()

    e.Root.P1.setQuadSkinCoords(tri0, tri1, 0, 0, x, y)
  }
}

func (e *Checkbox) Show() {
  e.setTypesAndTCoords()

  e.ElementData.Show()
}

func (e *Checkbox) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  w, h := e.GetSize()

  e.SetBorderedElementPos(w, h, e.borderT(), maxZIndex)

  return e.InitRect(w, h)
}

func (e *Checkbox) onMouseClick(evt *Event) {
  e.toggleAndUpdate()
}

func (e *Checkbox) toggleAndUpdate() {
  e.value = !e.value

  e.setTypesAndTCoords()
}

func (e *Checkbox) Cursor() int {
  return e.ButtonCursor(e.enabled)
}

func (e *Checkbox) Value() bool {
  return e.value
}
