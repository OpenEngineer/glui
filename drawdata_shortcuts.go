package glui

var (
  DEFAULT_SANS = "dejavusans"
)

func (d *DrawData) Button() *Button {
  return NewButton(d)
}

func (d *DrawData) Sans(content string, fontSize int) *Text {
  return NewText(d, content, DEFAULT_SANS, 10)
}

func (d *DrawData) Inline(hAlign, vAlign Align, spacing int) *Inline {
  return NewInline(hAlign, vAlign, spacing)
}

func (d *DrawData) Input() *Input {
  return NewInput(d)
}
