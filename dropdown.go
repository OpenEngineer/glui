package glui

// wrap button for most Element interface methods
// TODO: be its own element
type Dropdown struct {
  options []string
  text   *Text
  button *Button
}

func NewDropdown(root *Root, options []string) *Dropdown {
  button := NewButton(root)

  text := NewSans(root, "Choose unit", 10)

  arrow := NewIcon(root, "arrow-down-drop", 10)

  button.A(NewHor(root, STRETCH, CENTER, 0).Padding(0, 10).A(text, arrow))

  e := &Dropdown{options, text, button}

  button.OnClick(e.onShowMenu)

  return e
}

func (e *Dropdown) RegisterParent(parent Element) {
  e.button.RegisterParent(parent)
}

func (e *Dropdown) Parent() Element {
  return e.button.Parent()
}

func (e *Dropdown) Children() []Element {
  return e.button.Children()
}

func (e *Dropdown) Cursor() int {
  return e.button.Cursor()
}

func (e *Dropdown) Tooltip() string {
  return e.button.Tooltip()
}

func (e *Dropdown) CalcDepth(stack *ElementStack) {
  e.button.CalcDepth(stack)
}

func (e *Dropdown) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  return e.button.CalcPos(maxWidth, maxHeight, maxZIndex)
}

func (e *Dropdown) Animate(tick uint64) {
  e.button.Animate(tick)
}

func (e *Dropdown) Rect() Rect {
  return e.button.Rect()
}

func (e *Dropdown) ZIndex() int {
  return e.button.ZIndex()
}

func (e *Dropdown) Hit(x, y int) int {
  return e.button.Hit(x, y)
}

func (e *Dropdown) Translate(dx, dy int) {
  e.button.Translate(dx, dy)
}

func (e *Dropdown) Visible() bool {
  return e.button.Visible()
}

func (e *Dropdown) Hide() {
  e.button.Hide()
}

func (e *Dropdown) Show() {
  e.button.Show()
}

func (e *Dropdown) Enable() {
  e.button.Enable()
}

func (e *Dropdown) Disable() {
  e.button.Disable()
}

func (e *Dropdown) GetEventListener(name string) EventListener {
  return e.button.GetEventListener(name)
}

func (e *Dropdown) Delete() {
  e.button.Delete()
}

func (e *Dropdown) onShowMenu() {
  menu := e.button.Root.Menu

  menu.ClearChildren()

  for _, option := range e.options {
    option_ := option
    menu.AddButton(option_, true, e.button.height, func() {
      e.text.SetContent(option_)
    })
  }

  e.button.Root.Menu.ShowAt(
    e,
    0.0,
    1.0,
    e.button.rect.W,
  )
}
