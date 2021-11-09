package glui

type Orientation int

const (
  HOR Orientation = iota
  VER
)

func (o Orientation) Rotate() Orientation {
  if o == HOR {
    return VER
  } else {
    return HOR
  }
}
