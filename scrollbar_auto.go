package glui
func (e *Scrollbar) CalcDepth(stack *ElementStack) {
  e.zIndex = stack.Add(e, e.closerThan)
  for _, child := range e.Children() {
    child.CalcDepth(stack)
  }
}

func (e *Scrollbar) appendChild(children ...Element) Element {
  for _, child := range children {
    e.children = append(e.children, child)
    child.RegisterParent(e)
  }
  return e
}

func (e *Scrollbar) On(name string, fn EventListener) *Scrollbar {
  old := e.evtListeners[name]
  if old == nil {
    e.evtListeners[name] = fn
  } else {
    e.evtListeners[name] = func(evt *Event) {fn(evt); if !evt.stopPropagation {old(evt)}}
  }
  return e
}

