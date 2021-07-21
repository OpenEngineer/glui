package glui

type Element interface {
  AppendChild(e Element)

  OnResize(this Rect)
}

type ElementData struct {
  children []Element
}

func newElementData() ElementData {
  return ElementData{
    make([]Element, 0),
  }
}

func (e *ElementData) AppendChild(child Element) {
  e.children = append(e.children, child)
}
