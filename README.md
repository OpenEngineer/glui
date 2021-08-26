# glui
OpenGL UI Framework built on SDL

* cross-platform windows/linux (no mac 

# Handled by framework

## (Shift-)tab
Tab to prev/next element with a `"focus"`

# Types of elements

* Body
* Hor
* Text
* Icon
* Button
* Input
* Tabbed

# DWMPeekPreview

Windows only

Use `DwmSetIconicLivePreviewBitmap()` when receiving `WM_DWMSENDICONICLIVEPREVIEWBITMAP` message.

Declared in dwmapi.h

# Mixing a color with a skin

* if the skin color is white -> show the mixing color
* if the skin color is gray -> apply that grayness to the mixing color (i.e. multiply normalized colors)
