package glui
func (e *Overflow) CalcDepth(stack *ElementStack) {
  e.zIndex = stack.Add(e, e.closerThan)
  for _, child := range e.Children() {
    child.CalcDepth(stack)
  }
}

func (e *Overflow) appendChild(children ...Element) Element {
  for _, child := range children {
    e.children = append(e.children, child)
    child.RegisterParent(e)
  }
  return e
}

// must return Element in order to implement Container interface
func (e *Overflow) A(children ...Element) Element {
  for _, child := range children {
    e.children = append(e.children, child)
    child.RegisterParent(e)
  }
  return e
}

func (e *Overflow) Size(w, h int) *Overflow {
  e.width = w
  e.height = h
  e.Root.ForcePosDirty()
  return e
}

func (e *Overflow) H(h int) *Overflow {
  e.height = h
  e.Root.ForcePosDirty()
  return e
}

func (e *Overflow) W(w int) *Overflow {
  e.width = w
  e.Root.ForcePosDirty()
  return e
}

func (e *Overflow) On(name string, fn EventListener) *Overflow {
  old := e.evtListeners[name]
  if old == nil {
    e.evtListeners[name] = fn
  } else {
    e.evtListeners[name] = func(evt *Event) {fn(evt); if !evt.stopPropagation {old(evt)}}
  }
  return e
}

