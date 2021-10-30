package glui
func (e *Select) appendChild(children ...Element) Element {
  for _, child := range children {
    e.children = append(e.children, child)
    child.RegisterParent(e)
  }
  return e
}

func (e *Select) CalcDepth(stack *ElementStack) {
  e.zIndex = stack.Add(e, e.closerThan)
  for _, child := range e.Children() {
    child.CalcDepth(stack)
  }
}

func (e *Select) On(name string, fn EventListener) *Select {
  old := e.evtListeners[name]
  if old == nil {
    e.evtListeners[name] = fn
  } else {
    e.evtListeners[name] = func(evt *Event) {fn(evt); if !evt.stopPropagation {old(evt)}}
  }
  return e
}

func (e *Select) Size(w, h int) *Select {
  e.width = w
  e.height = h
  e.Root.ForcePosDirty()
  return e
}

func (e *Select) Padding(p ...int) *Select {
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

