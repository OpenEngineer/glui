package glui

// wrap button for most Element interface methods
type Dropdown struct {
  button *Button
}

func NewDropdown(dd *DrawData) *Dropdown {
  button := NewButton(dd)

  text := dd.Sans("Choose unit", 10)

  arrow := NewIcon(dd, "arrow-down-drop", 10)

  button.A(dd.Hor(STRETCH, CENTER, 0).Padding(0, 10).A(text, arrow))

  return &Dropdown{button}
}

func (e *Dropdown) RegisterParent(parent Element) {
  e.button.RegisterParent(parent)
}

func (e *Dropdown) Cursor() int {
  return e.button.Cursor()
}

func (e *Dropdown) Parent() Element {
  return e.button.Parent()
}

func (e *Dropdown) Children() []Element {
  return e.button.Children()
}

func (e *Dropdown) OnResize(maxWidth, maxHeight int) (int, int) {
  return e.button.OnResize(maxWidth, maxHeight)
}

func (e *Dropdown) OnTick(tick uint64) {
  e.button.OnTick(tick)
}

func (e *Dropdown) Hit(x, y int) bool {
  return e.button.Hit(x, y)
}

func (e *Dropdown) Translate(dx, dy int, dz float32) {
  e.button.Translate(dx, dy, dz)
}

func (e *Dropdown) SetZ(z float32) {
  e.button.SetZ(z)
}

func (e *Dropdown) GetEventListener(name string) EventListener {
  return e.button.GetEventListener(name)
}
