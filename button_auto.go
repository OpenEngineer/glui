package glui
// must return Element in order to implement Container interface
func (e *Button) A(children ...Element) Element {
  for _, child := range children {
    e.children = append(e.children, child)
    child.RegisterParent(e)
  }
  return e
}

func (e *Button) CalcDepth(stack *ElementStack) {
  e.zIndex = stack.Add(e, e.closerThan)
  for _, child := range e.Children() {
    child.CalcDepth(stack)
  }
}

func (e *Button) On(name string, fn EventListener) *Button {
  old := e.evtListeners[name]
  if old == nil {
    e.evtListeners[name] = fn
  } else {
    e.evtListeners[name] = func(evt *Event) {fn(evt); if !evt.stopPropagation {old(evt)}}
  }
  return e
}

func (e *Button) Size(w, h int) *Button {
  e.width = w
  e.height = h
  e.Root.ForcePosDirty()
  return e
}

func (e *Button) Padding(p ...int) *Button {
  switch len(p) {
  case 1:
    e.padding = [4]int{p[0], p[0], p[0], p[0]}
    break
  case 2:
    e.padding = [4]int{p[0], p[1], p[0], p[1]}
    break
  case 3:
    e.padding = [4]int{p[0], p[1], p[0], p[2]}
    break
  case 4:
    e.padding = [4]int{p[0], p[1], p[2], p[3]}
    break
  default:
    panic("unexpected number of padding elements")
  }
  e.Root.ForcePosDirty()
  return e
}

func (e *Button) H(h int) *Button {
  e.height = h
  e.Root.ForcePosDirty()
  return e
}

func (e *Button) W(w int) *Button {
  e.width = w
  e.Root.ForcePosDirty()
  return e
}

