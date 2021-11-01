package glui

import (
)

// dropdown is a flat button, with a menu showing below it when clicked,
// the button stays depressed while the menu is showing

func NewDropdown(align Align, items []MenuItemConfig) *Button {
  e := NewStickyFlatButton()

  menuItemMaker := func() []*MenuItem {
    items_ := make([]*MenuItem, len(items))
    for i, item := range items {
      items_[i] = newMenuItemFromConfig(item)
    }

    return items_
  }

  m := e.Root.Menu

  e.OnClick(func() {
    if m.IsOwnedBy(e) {
      m.Hide()

      e.Unstick()
    } else {
      bh := 30

      m.ClearChildren()

      items := menuItemMaker()

      menuW := 0
      for _, item := range items {
        item.On("click", func(evt *Event) {
          e.Unstick()
        })

        m.AddItem(item.H(bh), true, false)

        itemW, _ := item.GetSize()

        if itemW > menuW {
          menuW = itemW
        }
      }

      switch align {
      case START:
        m.ShowAt(
          e, 0.0, 1.0, menuW,
        )
      case CENTER:
        m.ShowAt(
          e, 0.5, 1.0, menuW,
        )
      case END:
        m.ShowAt(
          e, 1.0, 1.0, menuW,
        )
      default:
        panic("unhandled align for Dropdown")
      }
    }
  })

  e.On("mousebuttonoutsidemenu", func(evt *Event) {
    m.Hide()

    e.Unstick()
    
    if e.IsHit(evt.X, evt.Y) {
      evt.stopPropagation = true
    }
  })

  // keyup event is triggered before keypress, stop propagation so events are actually handled by keypress instead
  e.On("keyup", func(evt *Event) {
    if m.IsOwnedBy(e) {
      evt.stopPropagation = true
    }
  })

  e.On("keypress", func(evt *Event) {
    if m.IsOwnedBy(e) {
      evt.stopPropagation = true

      if evt.IsEscape() {
        m.Hide()

        e.Unstick()
      } else if evt.IsReturnOrSpace() {
        if m.SelectedIndex() == -1 {
          m.Hide()

          e.Unstick()
        } else {
          m.ClickSelected()

          e.Unstick()
        }
      } else if evt.Key == "down" {
        m.SelectNext()
      } else if evt.Key == "up" {
        m.SelectPrev()
      }
    }
  })

  return e
}

func NewIconDropdown(iconName string, iconSize int, align Align, items []MenuItemConfig) *Button {
  e := NewDropdown(align, items)

  e.A(NewHor(CENTER, CENTER, 0).H(-1).A(NewIcon(iconName, iconSize)))

  return e
}
