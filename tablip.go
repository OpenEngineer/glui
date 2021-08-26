package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

//go:generate ./gen_element tabLip "CalcDepth appendChild On"

const INACTIVE_LIP_DELTA = 5

type tabLip struct {
  ElementData

  tabbed  *Tabbed
  tab     *tabPage
  caption *Text
}

func newTabLip(tabbed *Tabbed, tab *tabPage, captionText string, closeable bool) *tabLip {
  root := tabbed.Root

  caption := NewSans(root, captionText, 10)

  e := &tabLip{
    NewElementData(root, 9*2, 0),
    tabbed,
    tab,
    caption,
  }

  e.width = 200
  e.height = 50
  e.closerThan = []Element{tab}

  e.appendChild(NewHor(root, START, CENTER, 0).Padding(0, 10).A(caption))
  //e.appendChild(caption)

  e.Show()

  e.On("mousedown", e.onMouseDown)

  return e
}


func (e *tabLip) onMouseDown(evt *Event) {
  if !e.isActive() {
    e.tabbed.setActiveLip(e)
  }
}

func (e *tabLip) isActive() bool {
  return e.tabbed.isActiveLip(e)
}

func (e *tabLip) Cursor() int {
  return e.ButtonCursor(!e.isActive())
}

func (e *tabLip) Show() {
  e.SetButtonStyle()

  bgColor := e.Root.P1.Skin.BGColor()

  xCornerTex, yCornerTex := getCornerSkinCoords(e.Root)

  active := e.isActive()

  j := 2
  for i := 0; i < 3; i++ {
    tri0 := e.p1Tris[(i*3+j)*2 + 0]
    tri1 := e.p1Tris[(i*3+j)*2 + 1]

    if active {
      if i == 0 {
        if e.tabbed.isFirstLip(e) {
          setQuadSkinCoords(e.Root, tri0, tri1, 0, 1, xCornerTex, yCornerTex)
        } else {
          setQuadSkinCoords(e.Root, tri0, tri1, 0, 0, xCornerTex, yCornerTex)
        }
      } else if i == 1 {
        e.Root.P1.Type.Set1Const(tri0, VTYPE_PLAIN)
        e.Root.P1.Type.Set1Const(tri1, VTYPE_PLAIN)

        e.Root.P1.SetColorConst(tri0, bgColor)
        e.Root.P1.SetColorConst(tri1, bgColor)
      } else if i == 2 {
        setQuadSkinCoords(e.Root, tri0, tri1, 2, 0, xCornerTex, yCornerTex)
      }
    } else {
      setQuadSkinCoords(e.Root, tri0, tri1, 1, 0, xCornerTex, yCornerTex)

      j_ := 1

      tri2 := e.p1Tris[(i*3+j_)*2 + 0]
      tri3 := e.p1Tris[(i*3+j_)*2 + 1]

      if i == 1 {
        e.Root.P1.SetQuadColorLinearVGrad(tri2, tri3, bgColor, sdl.Color{0xa0, 0xa0, 0xa0, 0xff})
      } else {
        e.Root.P1.SetQuadColorLinearVGrad(tri2, tri3, 
          sdl.Color{0xff, 0xff, 0xff, 0xff}, sdl.Color{0xd4, 0xd4, 0xd4, 0xff})
      }
    }
  }

  e.ElementData.Show()
}

func (e *tabLip) Select() {
  e.Show()

  e.tab.Show()
}

func (e *tabLip) Unselect() {
  e.Show()

  e.tab.Hide()
}

func (e *tabLip) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  t := e.Root.P1.Skin.ButtonBorderThickness()

  e.InitRect(e.width, e.height)

  if e.isActive() {
    e.height = 50
  } else {
    e.height = 50 - INACTIVE_LIP_DELTA
  }

  e.SetButtonPos(maxWidth, maxHeight, maxZIndex)

  e.CalcPosChildren(e.width, e.height, maxZIndex)

  if !e.isActive() {
    e.Translate(0, INACTIVE_LIP_DELTA + t)
  } else {
    e.Translate(0, t)
  }

  return e.width, e.height
}
