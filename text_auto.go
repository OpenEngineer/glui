package glui
func (e *Text) CalcDepth(stack *ElementStack) {
  e.zIndex = stack.Add(e, e.closerThan)
  for _, child := range e.Children() {
    child.CalcDepth(stack)
  }
}
func (e *Text) On(name string, fn EventListener) *Text {
  e.evtListeners[name] = fn
  return e
}
