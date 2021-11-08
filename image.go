package glui

import (
  "fmt"
  "image"
  "reflect"
)

//go:generate ./gen_element Image "CalcDepth"

type Image struct {
  ElementData

  img image.Image
}

func NewImage(img image.Image) *Image {
  fmt.Println(reflect.TypeOf(img).String())

  e := &Image{
    NewElementData(2, 0),
    img,
  }

  e.Root.P1.SetQuadImage(e.p1Tris[0], e.p1Tris[1], img)

  return e
}

func (e *Image) Show() {
  e.Root.P1.SetTriType(e.p1Tris[0], VTYPE_IMAGE)
  e.Root.P1.SetTriType(e.p1Tris[1], VTYPE_IMAGE)

  e.ElementData.Show()
}

// ignores the maxWidth and maxHeight args
func (e *Image) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  w, h := imgSize(e.img)

  e.Root.P1.SetQuadPos(e.p1Tris[0], e.p1Tris[1], Rect{0, 0, w, h}, e.Z(maxZIndex))
  e.Root.P1.setQuadImageRelTCoords(e.p1Tris[0], e.p1Tris[1])

  return e.InitRect(w, h)
}
