package glui

// Void is the most basic Element which can serve as eg. special anchors for the FocusRect

type Void struct {
  ElementData
}

func NewVoid() Void {
  return Void{
    NewElementData(0, 0),
  }
}

func (e *Void) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  return 0, 0
}

func (e *Void) CalcDepth(stack *ElementStack) {
  // not part of stack
  return
}
