package glui

// returns true if new active element is same as old active element, or is child of old active element
func findHitElement(e Element, x, y int) (Element, bool) {
  if e.Hit(x, y) {
    for {
      childHit := false
      for _, c := range e.Children() {
        if c.Hit(x, y) {
          e = c
          childHit = true
          break
        }
      }

      if !childHit {
        return e, true // resulting element is still child of old active element, or same as old active element
      }
    }
  } else {
    p := e.Parent()
    if p == nil {
      return e, true
    } else {
      res, _ := findHitElement(p, x, y)
      return res, false
    }
  }
}

func collectAncestors(a Element) []Element {
  res := make([]Element, 0)

  for {
    a = a.Parent()

    if a != nil {
      res = append([]Element{a}, res...)
    } else {
      break
    }
  }

  return res
}

// should at least resolve to *Body
func commonAncestor(a Element, b Element) Element {
  if a == b {
    return a
  }

  if _, aIsBody := a.(*Body); aIsBody {
    return a
  } else if _, bIsBody := b.(*Body); bIsBody {
    return b
  }

  aps := collectAncestors(a)
  bps := collectAncestors(b)

  for i := 1; i < len(aps) && i < len(bps); i++ {
    if aps[i] != bps[i] {
      return aps[i-1]
    }
  }

  if len(aps) < len(bps) {
    return aps[len(aps)-1]
  } else {
    return bps[len(bps)-1]
  }
}

func focusable(e Element) bool {
  return e.GetEventListener("focus") != nil
}

func findFocusable(e_ Element) Element {
  e := e_

  for e != nil {
    if focusable(e) {
      return e
    }

    e = e.Parent()
  }

  return nil
}

func findFirstFocusable(e Element) Element {
  if focusable(e) {
    return e
  }

  for _, eChild := range e.Children() {
    if focusable(eChild) {
      return eChild
    } else if inner := findFirstFocusable(eChild); inner != nil {
      return inner
    }
  }

  return nil
}

func findLastFocusable(e Element) Element {
  if focusable(e) {
    return e
  }

  children := e.Children()
  for i := len(children) - 1; i >= 0; i-- {
    eChild := children[i]

    if focusable(eChild) {
      return eChild
    } else if inner := findLastFocusable(eChild); inner != nil {
      return inner
    }
  }

  return nil
}

func findNextFocusable(e_ Element) Element {
  b := focusable(e_)
  e := e_
  p := e_
  if b {
    p = p.Parent()
  } 

  for p != nil {
    thisChildFound := false
    for _, pChild := range p.Children() {
      if pChild == e {
        thisChildFound = true
      } else if thisChildFound || !b {
        next := findFirstFocusable(pChild)
        if next != nil {
          return next
        }
      }
    }

    e = p
    p = p.Parent()
  }

  // we are the end, start from the beginning until we encounter initial element
  next := findFirstFocusable(e)
  if next == nil {
    return nil
  } else if next == e_ {
    return nil
  } else {
    return next
  }
}

func findPrevFocusable(e_ Element) Element {
  b := focusable(e_)
  e := e_
  p := e_
  if b {
    p = p.Parent()
  } 

  for p != nil {
    thisChildFound := false

    children := p.Children()
    for i := len(children) - 1; i >= 0; i-- {
      pChild := children[i]

      if pChild == e {
        thisChildFound = true
      } else if thisChildFound || !b {
        prev := findLastFocusable(pChild)
        if prev != nil {
          return prev
        }
      }
    }

    e = p
    p = p.Parent()
  }

  // we are the beginning , start from the end until we encounter the initial element
  prev := findLastFocusable(e)
  if prev == nil {
    return nil
  } else if prev == e_ {
    return nil
  } else {
    return prev
  }
}