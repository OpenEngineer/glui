package main

import (
  "encoding/base64"
  "errors"
  "fmt"
  "io/ioutil"
  "os"
  "path/filepath"
  "regexp"
  "strings"
)

func main() {
  if err := mainInternal(); err != nil {
    fmt.Fprintf(os.Stderr, err.Error())
  }
}

func mainInternal() error {
  args := os.Args[1:]

  if len(args) != 2 {
    return errors.New("expected 2 args")
  }

  name := args[0]

  imgFile := args[1]

  ext := filepath.Ext(imgFile)

  imgBytes, err := ioutil.ReadFile(imgFile)
  if err != nil {
    return err
  }

  imgData := base64.StdEncoding.EncodeToString(imgBytes)

  var b strings.Builder

  // detect package current
  pkgName, err := detectPackage()
  if err != nil {
    return err
  }

  b.WriteString("package ")
  b.WriteString(pkgName)
  b.WriteString("\n")

  b.WriteString("import (\n")
  b.WriteString("  \"bytes\"\n")
  b.WriteString("  \"encoding/base64\"\n")
  b.WriteString("  \"github.com/computeportal/glui\"\n")
  b.WriteString(")\n")

  b.WriteString("const (\n")

  dataName := name + "_data"
  b.WriteString("  ")
  b.WriteString(dataName)
  b.WriteString(" =\"")
  b.WriteString(imgData)
  b.WriteString("\"\n")
  b.WriteString(")\n")

  b.WriteString("var ")
  b.WriteString(name)
  b.WriteString(" = &glui.ImageData{nil, 0, 0}\n")

  b.WriteString("func load_")
  b.WriteString(name)
  b.WriteString("() bool {\n")

  b.WriteString("  b, err := base64.StdEncoding.DecodeString(")
  b.WriteString(dataName)
  b.WriteString(")\n")
  b.WriteString("  if err != nil {panic(err)}\n")

  b.WriteString("  buf := bytes.NewBuffer(b)\n")

  switch strings.ToLower(ext) {
  case ".jpg", ".jpeg":
    b.WriteString("  data, err := glui.DecodeJPG(buf)\n")
  case ".png":
    b.WriteString("  data, err := glui.DecodePNG(buf)\n")
  default:
    return errors.New("image extension " + ext + " not recognized")
  }
  b.WriteString("  if err != nil {panic(err)}\n")

  b.WriteString("  ")
  b.WriteString(name)
  b.WriteString(".Pix = data.Pix\n")

  b.WriteString("  ")
  b.WriteString(name)
  b.WriteString(".W = data.W\n")

  b.WriteString("  ")
  b.WriteString(name)
  b.WriteString(".H = data.H\n")

  b.WriteString("  return true\n")
  b.WriteString("}\n")

  b.WriteString("var ")
  b.WriteString(name)
  b.WriteString("_loaded = load_")
  b.WriteString(name)
  b.WriteString("()")

  return ioutil.WriteFile("image_" + name + ".go", []byte(b.String()), 0644)
}

func detectPackage() (string, error) {
  pwd, err := os.Getwd()
  if err != nil {
    return "", err
  }

  files, err := ioutil.ReadDir(pwd)
  if err != nil {
    return "", err
  }

  //re := regexp.MustCompile(`^package\s*([a-z]*)`)
  re := regexp.MustCompile(`package\s([a-z]+)`)

  for _, f := range files {
    if strings.HasSuffix(f.Name(), ".go") {
      b, err := ioutil.ReadFile(f.Name())
      if err != nil {
        return "", err
      }

      str := string(b)

      res := re.FindAllStringSubmatch(str, -1)

      if len(res) == 0 || res[0][1] == "" {
        return "", errors.New("package in " + f.Name() + " not defined")
      }

      return res[0][1], nil
    }
  }

  return "", errors.New("go package not found in directory " + pwd)
}
