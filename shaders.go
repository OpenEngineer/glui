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
  b.WriteString(fmt.Sprintf("const int DUMMY  = %d;\n", VTYPE_DUMMY))
  b.WriteString(fmt.Sprintf("const int IMAGE  = %d;\n", VTYPE_IMAGE))
}

func skinPassVertexShader() string {
  var b strings.Builder
  
  b.WriteString("#version 410\n")

  writeVertexTypes(&b)

  b.WriteString(`
in vec3 aPos;
in float aType;
layout(location = 2) in float aParam;
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

func skinPassFragmentShader() string {
  var b strings.Builder

  b.WriteString("#version 410\n")

  writeVertexTypes(&b)

  b.WriteString(`
precision highp float;

in float vType;
in float vParam;
in vec4  vColor;
in vec2  vTCoord;

uniform sampler2D uTex;

layout(location = 0) out vec4 oColor;

void main() {
  int t = int(vType);

  if (t == t) {
    oColor = vec4(0.5, 1.0, 0.0, 1.0);
  }
  if (t == PLAIN) {
    oColor = vColor;
  } else if (t == SKIN) {
    vec4 sColor = texture(uTex, vTCoord);

    oColor = vec4(
      sColor.x*vColor.x, 
      sColor.y*vColor.y,
      sColor.z*vColor.z,
      sColor.w
    );
  } else if (t == IMAGE) {
    oColor = texture(uTex, vTCoord);
  } else if (t == DUMMY) {
    oColor = vec4(vParam, vParam, vParam, 1.0); // so that aParam isn't optimized out
  }
}
`)

  b.WriteString("\x00")

  return b.String()
}

func glyphPassVertexShader() string {
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

func glyphPassFragmentShader() string {
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

uniform sampler2D uTex;

layout(location = 0) out vec4 oColor;
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
    vec4 gData = texture(uTex, vTCoord);

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

func blurPassVertexShader() string {
  var b strings.Builder
  
  b.WriteString("#version 410\n")

  b.WriteString(`
layout(location = 0) in vec2 aCoord;

out vec2 vCoord;

void main() {
  gl_Position = vec4(aCoord.x*2.0 - 1.0, aCoord.y*2.0 - 1.0, 0.0, 1.0);

  vCoord = aCoord;
}
`)

  b.WriteString("\x00")

  return b.String()
}

func gaussBlurWeight(sigma float64, x float64) float64 {
  return math.Exp(-0.5*x*x/(sigma*sigma))/math.Sqrt(2.0*math.Pi*sigma*sigma)
}

// for a 5 point stencil
func writeGaussBlurWeights(b *strings.Builder) {
  for i := 0; i < 5; i++ {
    b.WriteString(fmt.Sprintf("const float GBW%d = %f;\n", i, gaussBlurWeight(2.0/3.0, float64(i-2))))
  }
}

func blurPassFragmentShader() string {
  var b strings.Builder

  b.WriteString("#version 410\n")

  writeGaussBlurWeights(&b)

  b.WriteString(`
in vec2 vCoord;

layout(location = 0) out vec4 oColor;

uniform vec2 uSize;
uniform vec2 uDir;

uniform sampler2D uTex;

vec2 offset(vec2 c, float d) {
  return vec2((c.x*uSize.x + d*uDir.x)/uSize.x, (c.y*uSize.y + d*uDir.y)/uSize.y);
}

void main() {
`)

  // should be uneven number
  nStencil := 31
  halfStencil := (nStencil - 1)/2
  sigma := float64(halfStencil)/3.0

  for i := 0; i < nStencil; i++ {
    b.WriteString(fmt.Sprintf("vec4 a%d = texture(uTex, offset(vCoord, float(%d)));\n", i, i - halfStencil))
  }

  b.WriteString("oColor = ")

  for i := 0; i < nStencil; i++ {
    b.WriteString(fmt.Sprintf("float(%g)*a%d", gaussBlurWeight(sigma, float64(i - halfStencil)), i))

    if i < nStencil - 1 {
      b.WriteString(" + ")
    } else {
      b.WriteString(";")
    }
  }

  b.WriteString("\n}")

  b.WriteString("\x00")

  return b.String()
}

// copied from https://kylewbanks.com/blog/tutorial-opengl-with-golang-part-1-hello-opengl
func compileShader(source string, shaderType uint32) (uint32, error) {
  checkGLError()
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

    fmt.Println("shader compilation log length:", logLength)
    log := strings.Repeat("\x00", int(logLength+1))
    gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

    return 0, fmt.Errorf("failed to compile %v: %v", source, log)
  }

  checkGLError()
  return shader, nil
}

func compileProgram(vShaderSrc string, fShaderSrc string) (uint32, error) {
  checkGLError()
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

  checkGLError()

  return prog, nil
}

func compileSkinPass() (uint32, error) {
  return compileProgram(skinPassVertexShader(), skinPassFragmentShader())
}

func compileGlyphPass() (uint32, error) {
  return compileProgram(glyphPassVertexShader(), glyphPassFragmentShader())
}

func compileBlurPass() (uint32, error) {
  return compileProgram(blurPassVertexShader(), blurPassFragmentShader())
}
