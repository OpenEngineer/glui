package glui

type SortState int

const (
  UNSORTED SortState = iota
  ASCENDING 
  DESCENDING
  UNSORTABLE 
)

func NextSortState(ss SortState) SortState {
  switch ss {
  case UNSORTED, DESCENDING:
    return ASCENDING
  case ASCENDING:
    return DESCENDING
  case UNSORTABLE:
    return UNSORTABLE
  default: 
    panic("unhandled")
  }
}
