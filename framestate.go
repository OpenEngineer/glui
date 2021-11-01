package glui

type FrameState struct {
  mouseElement   Element
  focusElement   Element
  cursor         int
  lastDown       Element
  outside        bool
  lastDownX      int
  lastDownY      int
  mouseMoveSumX  int
  mouseMoveSumY  int
  lastUpX        int
  lastUpY        int
  upCount        int // limited to three
  lastTick       uint64
  lastUpTick     uint64
  blockNextMouseButtonEvent bool
}

func newFrameState() *FrameState {
  return &FrameState{
    nil,
    nil,
    -1,
    nil,
    false,
    0,0, 0,0,
    0,0,
    0,
    0,0,
    false,
  }
}
