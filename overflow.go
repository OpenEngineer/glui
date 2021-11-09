package glui

//go:generate ./gen_element Overflow "CalcDepth appendChild A"

// element with horizontal and vertical scrolling
type Overflow struct {
  ElementData
}

func NewOverflow() *Overflow {
  e := &Overflow{
    NewElementData(0, 0),
  }

  horSB := NewScrollbar(HOR)
  verSB := NewScrollbar(VER)

  e.appendChild(horSB)
  e.appendChild(verSB)

  return e
}

func (e *Overflow) horScrollbar() *Scrollbar {
  sb, ok := e.children[0].(*Scrollbar)
  if !ok {
    panic("expected scrollbar")
  }

  return sb
}

func (e *Overflow) verScrollbar() *Scrollbar {
  sb, ok := e.children[1].(*Scrollbar)
  if !ok {
    panic("expected scrollbar")
  }

  return sb
}

func (e *Overflow) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  horSB  := e.horScrollbar()
  verSB  := e.verScrollbar()

  childrenBck := e.children[:]
  e.children = e.children[2:]
  innerW, innerH := e.CalcPosChildren(maxWidth, maxHeight, maxZIndex)

  sbTrackSize := e.Root.P1.Skin.ScrollbarTrackSize()

  if innerW > maxWidth - sbTrackSize {
    horSB.Show()
    horSB.SetSliderLength(int(float64(maxWidth - 3*sbTrackSize)*float64(maxWidth - sbTrackSize)/float64(innerW)))
    horSB.CalcPos(maxWidth - sbTrackSize, sbTrackSize, maxZIndex)
    horSB.Translate(0.0, maxHeight - sbTrackSize)
  } else {
    horSB.Hide()
    maxWidth = innerW
  }

  if innerH > maxHeight - sbTrackSize {
    verSB.Show()
    verSB.SetSliderLength(int(float64(maxHeight - 3*sbTrackSize)*float64(maxHeight - sbTrackSize)/float64(innerH)))
    verSB.CalcPos(sbTrackSize, maxHeight - sbTrackSize, maxZIndex)
    verSB.Translate(maxWidth - sbTrackSize, 0.0)
  } else {
    verSB.Hide()
    maxHeight = innerH
  }


  // TODO: crop

  e.children = childrenBck

  return e.InitRect(maxWidth, maxHeight)
}
