package glui

import (
)

//go:generate ./gen_element Image "CalcDepth"

type Image struct {
  ElementData

  img *ImageData
}

func NewImage(img *ImageData) *Image {
  e := &Image{
    NewElementData(2, 0),
    img,
  }

  if img != nil {
    e.width = img.W
    e.height = img.H

    e.Root.P1.SetQuadImage(e.p1Tris[0], e.p1Tris[1], img)
  } else {
    e.Hide()
  }

  return e
}

func (e *Image) Img(img *ImageData) *Image {
  e.img = img

  if img != nil {

    e.width = img.W
    e.height = img.H

    e.Root.P1.SetQuadImage(e.p1Tris[0], e.p1Tris[1], img)
  } else {
    e.Hide()
  }

  e.Root.ForcePosDirty()

  return e
}

func (e *Image) Show() {
  if e.img != nil {
    e.Root.P1.SetTriType(e.p1Tris[0], VTYPE_IMAGE)
    e.Root.P1.SetTriType(e.p1Tris[1], VTYPE_IMAGE)
  }

  e.ElementData.Show()
}

// ignores the maxWidth and maxHeight args
func (e *Image) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  if e.img != nil {
    e.Root.P1.SetQuadPos(e.p1Tris[0], e.p1Tris[1], Rect{0, 0, e.width, e.height}, e.Z(maxZIndex))
    e.Root.P1.setQuadImageRelTCoords(e.p1Tris[0], e.p1Tris[1])

    return e.InitRect(e.width, e.height)
  } else {
    return e.InitRect(0, 0)
  }
}
