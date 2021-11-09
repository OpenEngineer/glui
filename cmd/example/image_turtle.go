package main
import (
  "bytes"
  "encoding/base64"
  "github.com/computeportal/glui"
)
const (
)
var turtle = &glui.ImageData{nil, 0, 0}
func load_turtle() bool {
  b, err := base64.StdEncoding.DecodeString(turtle_data)
  if err != nil {panic(err)}
  buf := bytes.NewBuffer(b)
  data, err := glui.DecodePNG(buf)
  if err != nil {panic(err)}
  turtle.Pix = data.Pix
  turtle.W = data.W
  turtle.H = data.H
  return true
}
var turtle_loaded = load_turtle()