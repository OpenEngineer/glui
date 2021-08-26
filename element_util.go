package glui

func elementNotNil(e Element) bool {
  return !(e == nil || e.Deleted())
}

// returns true if new active element is same as old active element, or is child of old active element
func findHitElement(e Element, x, y int) (Element, bool) {
  if z := e.Hit(x, y); z > -1 {
    for {
      childHit := false
      for _, c := range e.Children() {
        if zc := c.Hit(x, y); zc > z {
          e = c
          z = zc
          childHit = true
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

  for ; elementNotNil(a); {
    a = a.Parent()

    if elementNotNil(a) {
      res = append([]Element{a}, res...)
    } else {
      break
    }
  }

  return res
}

func commonAncestor(a Element, b Element) Element {
  if a == b {
    return a
  }

  aps := collectAncestors(a)
  bps := collectAncestors(b)

  for i := 1; i < len(aps) && i < len(bps); i++ {
    if aps[i] != bps[i] {
      return aps[i-1]
    }
  }

  if len(aps) < len(bps) {
    if len(aps) == 0 { 
      return nil
    } else {
      return aps[len(aps)-1]
    }
  } else {
    if len(bps) == 0 {
      return nil
    } else {
      return bps[len(bps)-1]
    }
  }
}

func hasAncestor(a Element, anc Element) bool {
  if a == anc {
    return true
  }

  aps := collectAncestors(a)

  for _, test := range aps {
    if test == anc {
      return true
    }
  }

  return false
}

func hasEvent(e Element, name string) bool {
  return e.GetEventListener(name) != nil
}

func focusable(e Element) bool {
  return elementNotNil(e) && hasEvent(e, "focus") && e.Visible()
}

func findFocusable(e_ Element) Element {
  e := e_

  for elementNotNil(e) {
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
    } else if inner := findFirstFocusable(eChild); elementNotNil(inner) {
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
    } else if inner := findLastFocusable(eChild); elementNotNil(inner) {
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

  for elementNotNil(p) {
    thisChildFound := false
    for _, pChild := range p.Children() {
      if pChild == e {
        thisChildFound = true
      } else if thisChildFound || !b {
        next := findFirstFocusable(pChild)
        if elementNotNil(next) {
          return next
        }
      }
    }

    e = p
    p = p.Parent()
  }

  // we are the end, start from the beginning until we encounter initial element
  next := findFirstFocusable(e)
  if !elementNotNil(next) {
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

  for elementNotNil(p) {
    thisChildFound := false

    children := p.Children()
    for i := len(children) - 1; i >= 0; i-- {
      pChild := children[i]

      if pChild == e {
        thisChildFound = true
      } else if thisChildFound || !b {
        prev := findLastFocusable(pChild)
        if elementNotNil(prev) {
          return prev
        }
      }
    }

    e = p
    p = p.Parent()
  }

  // we are the beginning , start from the end until we encounter the initial element
  prev := findLastFocusable(e)
  if !elementNotNil(prev) {
    return nil
  } else if prev == e_ {
    return nil
  } else {
    return prev
  }
}
