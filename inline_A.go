package glui
func (e *Inline) A(children ...Element) Element {
  for _, child := range children {
    e.ElementData.appendChild(child)
    child.RegisterParent(e)
  }
  return e
}
