package glui

// special element that is just used for positioning of children

type Align int

const (
  START  Align = iota
  CENTER
  END
)

type Inline struct {
  ElementData

  hAlign  Align
  vAlign  Align
  spacing int
}

func NewInline(hAlign, vAlign Align, spacing int) *Inline {
  return &Inline{newElementData(), hAlign, vAlign, spacing}
}

//go:generate ./A Inline

func (e *Inline) OnResize(maxWidth, maxHeight int) (int, int) {
  // first space the children inline

  x := 0
  maxChildH := 0

  childHs := make([]int, len(e.children))

  for i, child := range e.children {
    if i > 0 {
      x += e.spacing
    }

    childW, childH := child.OnResize(maxWidth - x, maxHeight)
    childHs[i] = childH

    child.Translate(x, 0)

    x += childW

    if childH > maxChildH {
      maxChildH = childH
    }
  }

  dx := 0
  if x < maxWidth {
    switch e.hAlign {
    case CENTER:
      dx = (maxWidth - x)/2
      break
    case END:
      dx = maxWidth - x
      break
    }
  }

  if dx != 0 || e.vAlign != START {
    for i, child := range e.children {
      dy := 0

      switch e.vAlign {
      case CENTER:
        dy = (maxHeight - childHs[i])/2
        break
      case END:
        dy = (maxHeight - childHs[i])
        break
      }

      child.Translate(dx, dy)
    }
  } 

  return e.InitBB(x, maxHeight)
}
