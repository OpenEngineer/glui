package glui
func (e *tabPage) CalcDepth(stack *ElementStack) {
  e.zIndex = stack.Add(e, e.closerThan)
  for _, child := range e.Children() {
    child.CalcDepth(stack)
  }
}
func (e *tabPage) A(children ...Element) Container {
  for _, child := range children {
    e.children = append(e.children, child)
    child.RegisterParent(e)
  }
  return e
}
func (e *tabPage) Padding(p ...int) Container {
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
func (e *tabPage) Spacing(s int) Container {
  e.spacing = s
  e.Root.ForcePosDirty()
  return e
}
