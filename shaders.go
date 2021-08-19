package glui

import (
  "fmt"
  "math"
  "strings"

  "github.com/go-gl/gl/v4.1-core/gl"
)

func writeVertexTypes(b *strings.Builder) {
  b.WriteString(fmt.Sprintf("\nconst int HIDDEN = %d;\n", VTYPE_HIDDEN))
  b.WriteString(fmt.Sprintf("const int PLAIN  = %d;\n", VTYPE_PLAIN))
  b.WriteString(fmt.Sprintf("const int SKIN  = %d;\n", VTYPE_SKIN))
  b.WriteString(fmt.Sprintf("const int GLYPH  = %d;\n", VTYPE_GLYPH))
}

func vertexShader1() string {
  var b strings.Builder
  
  b.WriteString("#version 410\n")

  writeVertexTypes(&b)

  b.WriteString(`
in vec3 aPos;
in float aType;
in float aParam;
in vec4 aColor;
in vec2 aTCoord;

out float vType;
out float vParam;
out vec4  vColor;
out vec2  vTCoord;

void main() {
  float z = aPos.z;
  if (aType == HIDDEN) {
    z = -10.0;
  }

  gl_Position = vec4(aPos.xy, z, 1.0);
  vType = aType;
  vParam = aParam;
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

func fragmentShader1() string {
  var b strings.Builder

  b.WriteString("#version 410\n")

  writeVertexTypes(&b)

  b.WriteString(`
precision highp float;

in float vType;
in float vParam;
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
  }
}
`)

  b.WriteString("\x00")

  return b.String()
}

func vertexShader2() string {
  var b strings.Builder
  
  b.WriteString("#version 410\n")

  writeVertexTypes(&b)

  b.WriteString(`
in vec3 aPos;
in float aType;
in float aParam;
in vec4 aColor;
in vec2 aTCoord;

out float vDepth;
out float vType;
out float vParam;
out vec4  vColor;
out vec2  vTCoord;

void main() {
  float z = aPos.z;
  if (aType == HIDDEN) {
    z = -10.0;
  }

  gl_Position = vec4(aPos.xy, z, 1.0);
  vType = aType;
  vParam = aParam;
  vColor = vec4(
    aColor.x, 
    aColor.y,
    aColor.z,
    aColor.w
  );
  vTCoord = aTCoord;
  vDepth = z;
}
`)

  b.WriteString("\x00")

  return b.String()
}

func fragmentShader2() string {
  var b strings.Builder

  b.WriteString("#version 410\n")

  writeVertexTypes(&b)
  b.WriteString(fmt.Sprintf("const float D_PER_PX = %g;\n", float64(GlyphDPerPx)))
  b.WriteString(fmt.Sprintf("const float PI = %g;\n", math.Pi))
  b.WriteString(fmt.Sprintf("const float HALF_SQRT2 = %g;\n", math.Sqrt(2.0)*0.5))

  b.WriteString(`
precision highp float;

in float vDepth;
in float vType;
in float vParam;
in vec4  vColor;
in vec2  vTCoord;

uniform sampler2D glyphs;

out vec4 oColor;
//out float gl_FragDepth;

float calcPixelCoverage(float d, float a) {
  float tana = tan(a);
  float d_over_cosa = d/cos(a);

  float h1 = 0.5 + 0.5*tana - d_over_cosa;

  if (h1 < 1e-5) {
    return 0.0;
  } else if (tana >= h1) {
    float w = h1/tana;

    return 0.5*w*h1;
  } else {
    float h2 = 0.5 - 0.5*tana - d_over_cosa;

    return 0.5*(h1 + h2);
  }
}

float pixelCoverageToAlpha(float A) {
  if (A <= 0.0) {
    return 0.0;
  } else if (A >= 1.0) {
    return 1.0;
  } else {
    return (A - 0.0)/(1.0 - 0.0);
  }
  //return smoothstep(0.0, 1.0, A);
}

void main() {
  int t = int(vType);

  if (t == GLYPH) {
    vec4 gData = texture(glyphs, vTCoord);

    float d = (gData.x - 0.5)*(vParam*255.0/D_PER_PX);
    float a = gData.y*(0.25*PI);

    bool outside = d < 0.5;
    if (outside) {
      d *= -1.0;
    }

    float A = 0.0;
    if (d < HALF_SQRT2) {
      A = calcPixelCoverage(d, a);
    } 

    if (!outside) {
      A = 1.0 - A;
    }

    float alpha = pixelCoverageToAlpha(A);
    //oColor = vec4(gData.y, 0.0, 0.0, A);

    oColor = vec4(vColor.xyz, alpha);
    //oColor = vec4(0.0, 0.0, 0.0, A);

    gl_FragDepth = (alpha > 0.1) ? vDepth : 2.0;
    //gl_FragDepth = gl_FragDepth*alpha + 0.55*(1.0-alpha);
  } else {
    oColor = vec4(0.0, 0.0, 0.0, 0.0);
    gl_FragDepth = 2.0;//vDepth;
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

func compileProgram(vShaderSrc string, fShaderSrc string) (uint32, error) {
  vShader, err := compileShader(vShaderSrc, gl.VERTEX_SHADER)
  if err != nil {
    return 0, err
  }

  fShader, err := compileShader(fShaderSrc, gl.FRAGMENT_SHADER)
  if err != nil {
    return 0, err
  }

  prog := gl.CreateProgram()
  gl.AttachShader(prog, vShader)
  gl.AttachShader(prog, fShader)

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

func compileProgram1() (uint32, error) {
  return compileProgram(vertexShader1(), fragmentShader1())
}

func compileProgram2() (uint32, error) {
  return compileProgram(vertexShader2(), fragmentShader2())
}
