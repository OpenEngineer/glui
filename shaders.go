package glui

import (
  "fmt"
  "strings"

  "github.com/go-gl/gl/v4.1-core/gl"
)

func writeVertexTypes(b *strings.Builder) {
  b.WriteString(fmt.Sprintf("\nconst int HIDDEN = %d;\n", VTYPE_HIDDEN))
  b.WriteString(fmt.Sprintf("const int PLAIN  = %d;\n", VTYPE_PLAIN))
  b.WriteString(fmt.Sprintf("const int SKIN  = %d;\n", VTYPE_SKIN))
}

func vertexShader() string {
  var b strings.Builder
  
  b.WriteString("#version 410\n")

  writeVertexTypes(&b)

  b.WriteString(`
in vec3 aPos;
in float aType;
in vec4 aColor;
in vec2 aTCoord;

out float vType;
out vec4  vColor;
out vec2  vTCoord;

void main() {
  float z = aPos.z;
  if (aType == HIDDEN) {
    z = -10.0;
  }

  gl_Position = vec4(aPos.xy, z, 1.0);
  vType = float(aType);
  vColor = vec4(
    aColor.x, 
    aColor.y,
    aColor.z,
    aColor.w
  );
  vTCoord = aTCoord;
}
`)

  b.WriteString("\x00")

  return b.String()
}

func fragmentShader() string {
  var b strings.Builder

  b.WriteString("#version 410\n")

  writeVertexTypes(&b)

  b.WriteString(`
precision highp float;

in float vType;
in vec4  vColor;
in vec2  vTCoord;

uniform sampler2D skin;

out vec4 oColor;

void main() {
  int t = int(vType);

  if (t == PLAIN) {
    oColor = vColor;
  } else if (t == SKIN) {
    oColor = texture(skin, vTCoord);
    //oColor = vec4(float(t) - 1.0, 1.0, 1.0, 0.0);//vColor;
  }
}
`)

  b.WriteString("\x00")

  return b.String()
}
// copied from https://kylewbanks.com/blog/tutorial-opengl-with-golang-part-1-hello-opengl
func compileShader(source string, shaderType uint32) (uint32, error) {
  shader := gl.CreateShader(shaderType)

  csources, free := gl.Strs(source)
  gl.ShaderSource(shader, 1, csources, nil)
  free()
  gl.CompileShader(shader)

  var status int32
  gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
  if status == gl.FALSE {
    var logLength int32
    gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

    log := strings.Repeat("\x00", int(logLength+1))
    gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

    return 0, fmt.Errorf("failed to compile %v: %v", source, log)
  }

  return shader, nil
}

func compileProgram() (uint32, error) {
  vShader, err := compileShader(vertexShader(), gl.VERTEX_SHADER)
  if err != nil {
    return 0, err
  }

  fShader, err := compileShader(fragmentShader(), gl.FRAGMENT_SHADER)
  if err != nil {
    return 0, err
  }

  prog := gl.CreateProgram()
  gl.AttachShader(prog, vShader)
  gl.AttachShader(prog, fShader)

  /*gl.BindAttribLocation(prog, 0, gl.Str("aPos"))
  gl.BindAttribLocation(prog, 1, gl.Str("aType"))
  gl.BindAttribLocation(prog, 2, gl.Str("aColor"))
  gl.BindAttribLocation(prog, 3, gl.Str("aTCoord"))*/

  gl.LinkProgram(prog)

  var status int32
  gl.GetProgramiv(prog, gl.LINK_STATUS, &status)
  if status == gl.FALSE {
    var logLength int32
    gl.GetProgramiv(prog, gl.INFO_LOG_LENGTH, &logLength)

    log := strings.Repeat("\x00", int(logLength+1))
    gl.GetProgramInfoLog(prog, logLength, nil, gl.Str(log))

    return 0, fmt.Errorf("failed to link %v", log)
  }

  return prog, nil
}
