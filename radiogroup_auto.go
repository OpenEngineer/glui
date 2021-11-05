package glui
func (e *RadioGroup) On(name string, fn EventListener) *RadioGroup {
  old := e.evtListeners[name]
  if old == nil {
    e.evtListeners[name] = fn
  } else {
    e.evtListeners[name] = func(evt *Event) {fn(evt); if !evt.stopPropagation {old(evt)}}
  }
  return e
}

func (e *RadioGroup) CalcDepth(stack *ElementStack) {
  e.zIndex = stack.Add(e, e.closerThan)
  for _, child := range e.Children() {
    child.CalcDepth(stack)
  }
}

func (e *RadioGroup) appendChild(children ...Element) Element {
  for _, child := range children {
    e.children = append(e.children, child)
    child.RegisterParent(e)
  }
  return e
}

