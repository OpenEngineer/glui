package glui

// XXX: chain parent rects everytime absolute positioning is specified
type Rect struct {
  X int // coordinates of top left
  Y int
  W int
  H int
}

func (r Rect) Right() int {
  return r.X + r.W
}

func (r Rect) Bottom() int {
  return r.Y + r.H
}

func (r Rect) Hit(x, y int) bool {
  return (x >= r.X) && (x < r.Right()) && (y >= r.Y) && (y < r.Bottom())
}
