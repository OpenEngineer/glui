package glui

type Rainbow struct {
  ElementData

  tri0 uint32
  tri1 uint32

  dd   *DrawData
}

func NewRainbow(dd *DrawData) *Rainbow {
  tris := dd.P1.Alloc(2)
  tri0 := tris[0]
  tri1 := tris[1]

  e := &Rainbow{newElementData(), tri0, tri1, dd}
  e.setTypesAndColors()

  return e
}

func (e *Rainbow) AppendChild(child Element) {
  e.ElementData.appendChild(child)

  child.RegisterParent(e)
}

func (e *Rainbow) setTypesAndColors() {
  e.dd.P1.Type.Set1Const(e.tri0, VTYPE_PLAIN)

  e.dd.P1.Color.Set4(e.tri0, 0, 1.0, 0, 0, 1.0);
  e.dd.P1.Color.Set4(e.tri0, 1, 0, 1.0, 0, 1.0);
  e.dd.P1.Color.Set4(e.tri0, 2, 0, 0, 1.0, 1.0);

  e.dd.P1.TCoord.Set2Const(e.tri0, 0.0, 0.0);

  e.dd.P1.Type.Set1Const(e.tri1, VTYPE_PLAIN)

  e.dd.P1.Color.Set4(e.tri1, 0, 1.0, 1.0, 0, 1.0);
  e.dd.P1.Color.Set4(e.tri1, 1, 0, 1.0, 0, 1.0);
  e.dd.P1.Color.Set4(e.tri1, 2, 0, 0, 1.0, 1.0);

  e.dd.P1.TCoord.Set2Const(e.tri1, 0.0, 0.0);
}

func (e *Rainbow) OnResize(rect Rect) {
  margin := 10

  l := rect.X + margin
  r := rect.Right() - margin

  t := rect.Y + margin
  b := rect.Bottom() - margin

  e.bb = Rect{l, t, r - l, b - t}

  e.dd.P1.SetPos(e.tri0, 0, l, b, 0.5)
  e.dd.P1.SetPos(e.tri0, 1, r, b, 0.5)
  e.dd.P1.SetPos(e.tri0, 2, l, t, 0.5)

  e.dd.P1.SetPos(e.tri1, 0, r, t, 0.5)
  e.dd.P1.SetPos(e.tri1, 1, r, b, 0.5)
  e.dd.P1.SetPos(e.tri1, 2, l, t, 0.5)
}
