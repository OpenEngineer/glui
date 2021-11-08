package glui

import (
  "image"
  "image/jpeg"
  "image/png"
  "io"
  "sort"
)

type ImageData struct {
  Pix []byte
  W   int
  H   int
}

type ImageInfo struct {
  X    int
  Y    int
  Used bool
}

type imageDataSorter struct {
  images []*ImageData
}

func (s *imageDataSorter) Len() int {
  return len(s.images)
}

func (s *imageDataSorter) Less(i, j int) bool {
  imgI := s.images[i]
  imgJ := s.images[j]

  return imgI.W*imgI.H > imgJ.W*imgJ.H
}

func (s *imageDataSorter) Swap(i, j int) {
  s.images[i], s.images[j] = s.images[j], s.images[i]
}

func sortImageData(images []*ImageData) []*ImageData {
  s := &imageDataSorter{images}

  sort.Sort(s)

  return s.images
}

func DecodeJPG(buf io.Reader) (*ImageData, error) {
  img, err := jpeg.Decode(buf)
  if err != nil {
    return nil, err
  }

  return DecodeImage(img)
}

func DecodePNG(buf io.Reader) (*ImageData, error) {
  img, err := png.Decode(buf)
  if err != nil {
    return nil, err
  }

  return DecodeImage(img)
}

func DecodeImage(img image.Image) (*ImageData, error) {
  wh := img.Bounds().Max
  w := wh.X
  h := wh.Y

  data := make([]byte, 4*w*h)

  for i := 0; i < w; i++ {
    for j := 0; j < h; j++ {
      c := img.At(i, j)

      r, g, b, a := c.RGBA()

      // convert components from 16bit to 8bit
      r = r >> 8
      g = g >> 8
      b = b >> 8
      a = a >> 8

      k := i*h + j

      data[k*4+0] = byte(r)
      data[k*4+1] = byte(g)
      data[k*4+2] = byte(b)
      data[k*4+3] = byte(a)
    }
  }

  return &ImageData{data, w, h}, nil
}
