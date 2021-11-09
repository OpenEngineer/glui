package main
import (
  "bytes"
  "encoding/base64"
  "github.com/computeportal/glui"
)
const (
)
var parrot = &glui.ImageData{nil, 0, 0}
func load_parrot() bool {
  b, err := base64.StdEncoding.DecodeString(parrot_data)
  if err != nil {panic(err)}
  buf := bytes.NewBuffer(b)
  data, err := glui.DecodeJPG(buf)
  if err != nil {panic(err)}
  parrot.Pix = data.Pix
  parrot.W = data.W
  parrot.H = data.H
  return true
}
var parrot_loaded = load_parrot()