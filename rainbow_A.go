package glui
func (e *Rainbow) A(children ...Element) Element {
  for _, child := range children {
    e.ElementData.appendChild(child)
    child.RegisterParent(e)
  }
  return e
}
