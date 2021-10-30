package glui

import (
)

// dropdown is a flat button, with a menu showing below it when clicked,
// the button stays depressed while the menu is showing

func NewDropdown(root *Root, align Align, items []MenuItemConfig) *Button {
  e := NewStickyFlatButton(root)

  menuItemMaker := func() []*MenuItem {
    items_ := make([]*MenuItem, len(items))
    for i, item := range items {
      items_[i] = newMenuItemFromConfig(root, item)
    }

    return items_
  }

  e.OnClick(func() {
    if root.Menu.IsOwnedBy(e) {
      root.Menu.Hide()

      e.Unstick()
    } else {
      bh := 30

      root.Menu.ClearChildren()

      items := menuItemMaker()

      menuW := 0
      for _, item := range items {
        item.On("click", func(evt *Event) {
          e.Unstick()
        })

        root.Menu.AddItem(item.H(bh), true, false)

        itemW, _ := item.GetSize()

        if itemW > menuW {
          menuW = itemW
        }
      }

      switch align {
      case START:
        root.Menu.ShowAt(
          e, 0.0, 1.0, menuW,
        )
      case CENTER:
        root.Menu.ShowAt(
          e, 0.5, 1.0, menuW,
        )
      case END:
        root.Menu.ShowAt(
          e, 1.0, 1.0, menuW,
        )
      default:
        panic("unhandled align for Dropdown")
      }
    }
  })

  e.On("mousebuttonoutsidemenu", func(evt *Event) {
    root.Menu.Hide()

    e.Unstick()
    
    if e.IsHit(evt.X, evt.Y) {
      evt.stopPropagation = true
    }
  })

  // keyup event is triggered before keypress, stop propagation so events are actually handled by keypress instead
  e.On("keyup", func(evt *Event) {
    if root.Menu.IsOwnedBy(e) {
      evt.stopPropagation = true
    }
  })

  e.On("keypress", func(evt *Event) {
    if root.Menu.IsOwnedBy(e) {
      evt.stopPropagation = true

      if evt.IsEscape() {
        root.Menu.Hide()

        e.Unstick()
      } else if evt.IsReturnOrSpace() {
        if root.Menu.SelectedIndex() == -1 {
          root.Menu.Hide()

          e.Unstick()
        } else {
          root.Menu.ClickSelected()

          e.Unstick()
        }
      } else if evt.Key == "down" {
        root.Menu.SelectNext()
      } else if evt.Key == "up" {
        root.Menu.SelectPrev()
      }
    }
  })

  return e
}

func NewIconDropdown(root *Root, iconName string, iconSize int, align Align, items []MenuItemConfig) *Button {
  e := NewDropdown(root, align, items)

  e.A(NewHor(root, CENTER, CENTER, 0).H(-1).A(NewIcon(root, iconName, iconSize)))

  return e
}
