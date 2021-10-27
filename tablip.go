package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

//go:generate ./gen_element tabLip "CalcDepth appendChild On"

const (
  INACTIVE_TABLIP_DELTA = 5
  TABLIP_CAPTION_SIZE   = 10
  TABLIP_CLOSE_INNER_SIZE = 15
  TABLIP_CLOSE_OUTER_SIZE = 20
)

type tabLipCaption struct {
  Text

  tl *tabLip // tabLipCaption isn't focusable if tabLip is active
}

type tabLip struct {
  ElementData

  tabbed  *Tabbed
  tab     *tabPage
  caption *tabLipCaption

  touchesRightSide_ bool
}

func newTabLip(tabbed *Tabbed, tab *tabPage, captionText string, closeable bool) *tabLip {
  root := tabbed.Root

  caption_ := NewSans(root, captionText, TABLIP_CAPTION_SIZE)
  caption := &tabLipCaption{*caption_, nil}

  e := &tabLip{
    NewElementData(root, 9*2, 0),
    tabbed,
    tab,
    caption,
    false,
  }

  caption.tl = e

  e.width = 200
  e.height = 50
  e.closerThan = []Element{tab}

  if closeable {
    closeButton := NewFlatIconButton(root, "close-thick", TABLIP_CLOSE_INNER_SIZE).Size(TABLIP_CLOSE_OUTER_SIZE, TABLIP_CLOSE_OUTER_SIZE)

    closeButton.OnClick(e.onClickCloseButton)

    e.appendChild(NewHor(root, STRETCH, CENTER, 0).Padding(0, 10).A(caption, closeButton))
  } else {
    e.appendChild(NewHor(root, START, CENTER, 0).Padding(0, 10).A(caption))
  }

  e.On("mousedown", e.onMouseDown)
  caption.On("focus", e.onFocusCaption)
  caption.On("blur", e.onBlurCaption)
  caption.On("keyup", e.onCaptionKeyUp)

  return e
}

func (e *tabLipCaption) IsFocusable() bool {
  return e.ElementData.IsFocusable() && !e.tl.isActive()
}

func (e *tabLip) onMouseDown(evt *Event) {
  if !e.isActive() {
    e.tabbed.setActiveLip(e)
  }
}

func (e *tabLip) onClickCloseButton() {
  e.closeTab()
}

func (e *tabLip) onFocusCaption(evt *Event) {
  // TODO: use a wrapper with some padding wrt. the text
  if !e.isActive() && evt.IsKeyboardEvent() {
    e.Root.FocusRect.Show(e.caption)
  }
}

func (e *tabLip) onBlurCaption(evt *Event) {
  e.Root.FocusRect.Hide()
}

func (e *tabLip) onCaptionKeyUp(evt *Event) {
  if evt.IsReturnOrSpace() {
    e.tabbed.setActiveLip(e)

    e.Root.FocusRect.Hide()
  }
}

func (e *tabLip) closeTab() {
  e.tabbed.closeTab(e)
}

func (e *tabLip) isActive() bool {
  return e.tabbed.isActiveLip(e)
}

func (e *tabLip) IsFocusable() bool {
  return e.ElementData.IsFocusable() && !e.isActive()
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
        if e.touchesRightSide_ {
          setQuadSkinCoords(e.Root, tri0, tri1, 2, 1, xCornerTex, yCornerTex)
        } else {
          setQuadSkinCoords(e.Root, tri0, tri1, 2, 0, xCornerTex, yCornerTex)
        }
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

func (e *tabLip) touchesRightSide(b bool) {
  if e.touchesRightSide_ != b {

    e.touchesRightSide_ = b
    if e.Visible() {
      e.Show()
    }
  }
}

func (e *tabLip) Select() {
  e.Show()

  e.tab.Show()
}

func (e *tabLip) Unselect() {
  e.Show()

  e.tab.Hide()
}

// the tabPage is moved up by ButtonBorderThickness in order to overlap with the bottom border of the tabLip
func (e *tabLip) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  e.InitRect(e.width, e.height)

  innerHeight := 50 - INACTIVE_TABLIP_DELTA
  if e.isActive() {
    e.height = 50
  } else {
    e.height = innerHeight
  }

  e.SetButtonPos(maxWidth, maxHeight, maxZIndex)

  e.CalcPosChildren(e.width, innerHeight, maxZIndex)

  if !e.isActive() {
    e.Translate(0, INACTIVE_TABLIP_DELTA)
  } else {
    for _, child := range e.children {
      child.Translate(0, INACTIVE_TABLIP_DELTA)
    }
  }

  return e.width, e.height
}
