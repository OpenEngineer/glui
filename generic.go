// +build !windows

package glui

import (
  "github.com/veandco/go-sdl2/sdl"
)

func InitOS(window *sdl.Window) error {
  return nil
}

func HandleSysWMEvent(app *App, event *sdl.SysWMEvent) error {
  return nil
}
