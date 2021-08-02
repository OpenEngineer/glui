package glui

// returns true if new active element is same as old active element, or is child of old active element
func findActive(e Element, x, y int) (Element, bool) {
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
      res, _ := findActive(p, x, y)
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
