package glui
func (e *tabLip) CalcDepth(stack *ElementStack) {
  e.zIndex = stack.Add(e, e.closerThan)
  for _, child := range e.Children() {
    child.CalcDepth(stack)
  }
}
func (e *tabLip) appendChild(children ...Element) Element {
  for _, child := range children {
    e.children = append(e.children, child)
    child.RegisterParent(e)
  }
  return e
}
func (e *tabLip) On(name string, fn EventListener) *tabLip {
  e.evtListeners[name] = fn
  return e
}
