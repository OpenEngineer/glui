package glui
func (e *MenuItem) appendChild(children ...Element) Element {
  for _, child := range children {
    e.children = append(e.children, child)
    child.RegisterParent(e)
  }
  return e
}

func (e *MenuItem) CalcDepth(stack *ElementStack) {
  e.zIndex = stack.Add(e, e.closerThan)
  for _, child := range e.Children() {
    child.CalcDepth(stack)
  }
}

func (e *MenuItem) On(name string, fn EventListener) *MenuItem {
  old := e.evtListeners[name]
  if old == nil {
    e.evtListeners[name] = fn
  } else {
    e.evtListeners[name] = func(evt *Event) {fn(evt); if !evt.stopPropagation {old(evt)}}
  }
  return e
}

func (e *MenuItem) Size(w, h int) *MenuItem {
  e.width = w
  e.height = h
  e.Root.ForcePosDirty()
  return e
}

func (e *MenuItem) Padding(p ...int) *MenuItem {
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

func (e *MenuItem) H(h int) *MenuItem {
  e.height = h
  e.Root.ForcePosDirty()
  return e
}

func (e *MenuItem) W(w int) *MenuItem {
  e.width = w
  e.Root.ForcePosDirty()
  return e
}

