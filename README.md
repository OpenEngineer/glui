# glui
OpenGL UI Framework built on top of SDL.

Works on both Windows and Linux.

Mac isn't supported because Mac doesn't support OpenGL anymore.

# Element types

* Button
* Hor
* Icon
* Input
* Tabbed
* Table
* Text
* Ver
* VSplit

See the example source `./cmd/example/main.go` and example executable `./build/example`.

# Handled by framework

## (Shift-)tab
Tab to prev/next element with a `"focus"` event listener.

## Focusrect
Some elements act as the anchor for a focusrect when focused. These elements grab keyboard input.

## Dialogs
A dialog can be created with the `PushFrame(maxWidth, maxHeight)` function. Then elements can be added to the new `ActiveBody()`. 

Gaussian blur is applied to the lower lying frames.

The dialog can be removed by calling the `PopFrame()` function.

## Fonts/icons
Fonts/icons can be included as a texture. There is no font hinting, but this is hardly noticeable on modern computer screens.

# TODO
* Images
* Scrollbars
* Panel with overflow scrollbar

# Notes
## Mixing a color with a skin
* if the skin color is white -> show the mixing color
* if the skin color is gray -> apply that grayness to the mixing color (i.e. multiply normalized colors)
