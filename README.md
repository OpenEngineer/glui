# glui
OpenGL UI Framework built on SDL

Works on both Windows and Linux.

Mac isn't supported because Mac doesn't support OpenGL anymore.

# Handled by framework

## (Shift-)tab
Tab to prev/next element with a `"focus"` event listener.

## Focusrect
Some elements act as the anchor for a focusrect when focused. These elements grab keyboard input.

## Dialogs
A dialog is created with the `PushFrame(maxWidth, maxHeight)` function. Gaussian blur is then applied to the lower lying frames.

A dialog can be removed by calling the `PopFrame()` function.

## Fonts/icons
Fonts/icons can be included as a texture. There is no font hinting, but this is barely noticeable on modern screens.

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


# TODO

## Mixing a color with a skin
* if the skin color is white -> show the mixing color
* if the skin color is gray -> apply that grayness to the mixing color (i.e. multiply normalized colors)
