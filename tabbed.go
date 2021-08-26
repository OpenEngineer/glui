package glui

import (
  "fmt"
)

//go:generate ./gen_element Tabbed "CalcDepth appendChild"

type Tabbed struct {
  ElementData

  active int

  lips []*tabLip
  tabs []*tabPage
}

func NewTabbed(root *Root) *Tabbed {
  return &Tabbed{
    NewElementData(root, 0, 0),
    -1,
    []*tabLip{},
    []*tabPage{},
  }
}

// returns a handle
func (e *Tabbed) NewTab(caption string, closeable bool) Container {
  tab := newTabPage(e)

  lip := newTabLip(e, tab, caption, closeable)

  e.lips = append(e.lips, lip)
  e.tabs = append(e.tabs, tab)

  e.appendChild(tab)
  e.appendChild(lip)

  if len(e.lips) == 1 {
    e.setActive(0)
  } else {
    e.setActive(e.active)
  }

  return tab
}

func (e *Tabbed) lipIndex(l *tabLip) int {
  for i, lip := range e.lips {
    if lip == l {
      return i
    }
  }

  return -1
}

func (e *Tabbed) tabIndex(t *tabPage) int {
  for i, tab := range e.tabs {
    if tab == t {
      return i
    }
  }

  return -1
}

func (e *Tabbed) setActive(idx int) {
  if idx > len(e.lips) - 1 {
    idx = len(e.lips) - 1
  }

  e.active = idx

  e.Show()
}

func (e *Tabbed) setActiveLip(l *tabLip) {
  e.setActive(e.lipIndex(l))
}

func (e *Tabbed) isActiveLip(l *tabLip) bool {
  return e.lipIndex(l) == e.active
}

func (e *Tabbed) isFirstLip(l *tabLip) bool {
  if len(e.lips) < 1 {
    return false
  } else {
    return e.lips[0] == l
  }
}

func (e *Tabbed) isActiveTab(t *tabPage) bool {
  return e.tabIndex(t) == e.active
}

func (e *Tabbed) closeTab(l *tabLip) {
  i := e.lipIndex(l)

  fmt.Println("deleting tab: ", i)
  e.lips[i].Delete()
  e.tabs[i].Delete()

  e.lips = append(e.lips[0:i], e.lips[i+1:]...)
  e.tabs = append(e.tabs[0:i], e.tabs[i+1:]...)

  if i > 0 {
    i -= 1
  } else if len(e.lips) == 0 {
    i = -1
  } else {
    i += 1
  }

  if i > len(e.lips) - 1 {
    i = len(e.lips) - 1
  }

  fmt.Println("new active tab: ", i)

  e.setActive(i)
}

func (e *Tabbed) Show() {
  for i, lip := range e.lips {
    if i == e.active {
      lip.Select()
    } else {
      lip.Unselect()
    }
  }
}

func (e *Tabbed) CalcPos(maxWidth, maxHeight, maxZIndex int) (int, int) {
  lipH := 0

  x := e.padding[3]
  y := e.padding[0]

  for _, lip := range e.lips {
    w, h := lip.CalcPos(maxWidth - x - e.padding[1], maxHeight - y - e.padding[2], maxZIndex)

    if h > lipH {
      lipH = h
    }

    lip.Translate(x, y)

    x += w
  }

  x = e.padding[3]
  y += lipH

  for i, tab := range e.tabs {
    if i == e.active {
      tab.CalcPos(maxWidth - x - e.padding[1], maxHeight - y - e.padding[2], maxZIndex)

      tab.Translate(x, y)
    }
  }

  return e.InitRect(maxWidth, maxHeight)
}
