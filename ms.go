// +build windows

package glui

// #cgo CFLAGS: -I/usr/share/mingw-w64/include -D_WIN32_WINNT=0x0601
// #cgo LDFLAGS: /usr/x86_64-w64-mingw32/lib/libdwmapi.a /usr/x86_64-w64-mingw32/lib/libgdi32.a
// #include "ms.h" 
import "C"

import (
  "fmt"
  "image"
  "image/color"
  "image/png"
  "os"
  "unsafe"

  "github.com/veandco/go-sdl2/sdl"
)

func getWindowHandle(window *sdl.Window) (*C.struct_HWND__, error) {
  info_, err := window.GetWMInfo()
  if err != nil {
    return nil, err
  }

  info := info_.GetWindowsInfo()

  return (*C.struct_HWND__)(info.Window), nil
}

func InitOS(window *sdl.Window) error {
  hwnd, err := getWindowHandle(window)
  if err != nil {
    return err
  }

  C.SetWindowAttributes(hwnd)

  sdl.EventState(sdl.SYSWMEVENT, sdl.ENABLE)

  return nil
}

func HandleSysWMEvent(app *App, event *sdl.SysWMEvent) error {
  sysMsg := (*C.SDL_SysWMmsg)(unsafe.Pointer(event.Msg))
  switch C.GetSysWMmsgType(sysMsg) {
  case C.WM_DWMSENDICONICLIVEPREVIEWBITMAP:
    fmt.Println("received SENDICONICLIVEPREVIEW request\n")
    break
  case C.WM_DWMSENDICONICTHUMBNAIL:
    fmt.Println("received SENICONICTHUMBNAIL request\n")

    wMax := uint32(C.GetIconicThumbnailMaxWidth(sysMsg))
    hMax := uint32(C.GetIconicThumbnailMaxHeight(sysMsg))

    b := make([]byte, wMax*hMax*4)

    bUnsafe := unsafe.Pointer(&(b[0]))

    app.drawThumbnail(int(wMax), int(hMax), bUnsafe)

    // also save as image
    bmpRect := image.Rect(0, 0, int(wMax), int(hMax))
    bmp := image.NewRGBA(bmpRect)
    for i := 0; i < int(wMax); i++ {
      for j := 0; j < int(hMax); j++ {
        k := i*int(hMax) + j

        bmp.SetRGBA(i, j, color.RGBA{b[k*4+0], b[k*4+1], b[k*4+2], b[k*4+3]})
      }
    }

    // save for debugging
    f, err := os.Create("thumbnail.png")
    if err != nil {
      fmt.Fprintf(os.Stderr, "failed to create file for thumbnail")
    } else {
      if err := png.Encode(f, bmp); err != nil {
        fmt.Fprintf(os.Stderr, "thumbnail generation failed")
      }
    }

    f.Close()

    hwnd, err := getWindowHandle(app.window)
    if err != nil {
      return err
    }

    C.SetIconicThumbnail(hwnd, (C.uint32_t)(wMax), (C.uint32_t)(hMax), (*C.uchar)(bUnsafe))
  }

  return nil
}
