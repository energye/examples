// Package shader provides shader management with caching
package shader

import (
	"fmt"
	"unsafe"

	"github.com/energye/examples/lcl/gpui/core/gl"
)

// ShaderProgram represents a compiled shader program
type ShaderProgram struct {
	ID          uint32
	Name        string
	uniformLocs map[string]int32
}

// ShaderManager manages shader programs with uniform caching
type ShaderManager struct {
	shaders map[string]*ShaderProgram
	current *ShaderProgram
}

// NewShaderManager creates a new shader manager
func NewShaderManager() *ShaderManager {
	return &ShaderManager{
		shaders: make(map[string]*ShaderProgram),
	}
}

// LoadShader loads and compiles a shader program
func (sm *ShaderManager) LoadShader(name, vertSrc, fragSrc string) (*ShaderProgram, error) {
	if sm == nil {
		return nil, fmt.Errorf("shader manager is nil")
	}
	if sm.shaders == nil {
		sm.shaders = make(map[string]*ShaderProgram)
	}
	if !shaderGLReady() {
		return nil, fmt.Errorf("shader load requires initialized OpenGL shader functions")
	}

	// Compile vertex shader
	vs := compileShader(vertSrc, gl.GL_VERTEX_SHADER)
	if vs == 0 {
		return nil, fmt.Errorf("failed to compile vertex shader: %s", name)
	}
	defer gl.DeleteShader(vs)

	// Compile fragment shader
	fs := compileShader(fragSrc, gl.GL_FRAGMENT_SHADER)
	if fs == 0 {
		return nil, fmt.Errorf("failed to compile fragment shader: %s", name)
	}
	defer gl.DeleteShader(fs)

	// Link program
	prog := gl.CreateProgram()
	gl.AttachShader(prog, vs)
	gl.AttachShader(prog, fs)

	// Bind attribute locations
	gl.BindAttribLocation(prog, 0, strPtr("aPos\x00"))
	gl.BindAttribLocation(prog, 1, strPtr("aUV\x00"))
	gl.BindAttribLocation(prog, 2, strPtr("aColor\x00"))

	gl.LinkProgram(prog)

	// Check link status
	var status int32
	gl.GetProgramiv(prog, gl.GL_LINK_STATUS, &status)
	if status == gl.GL_FALSE {
		var logLen int32
		gl.GetProgramiv(prog, gl.GL_INFO_LOG_LENGTH, &logLen)
		log := make([]byte, logLen+1)
		gl.GetProgramInfoLog(prog, logLen, nil, &log[0])
		gl.DeleteProgram(prog)
		return nil, fmt.Errorf("failed to link shader %s: %s", name, string(log))
	}

	shader := &ShaderProgram{
		ID:          prog,
		Name:        name,
		uniformLocs: make(map[string]int32),
	}

	if existing := sm.shaders[name]; existing != nil && existing.ID != 0 && gl.DeleteProgram != nil {
		gl.DeleteProgram(existing.ID)
		if sm.current == existing {
			sm.current = nil
		}
	}
	sm.shaders[name] = shader
	return shader, nil
}

// GetShader returns a shader by name
func (sm *ShaderManager) GetShader(name string) *ShaderProgram {
	if sm == nil {
		return nil
	}
	return sm.shaders[name]
}

// UseShader activates a shader program
func (sm *ShaderManager) UseShader(shader *ShaderProgram) {
	if sm == nil || shader == nil || shader.ID == 0 || gl.UseProgram == nil {
		return
	}
	if sm.current != shader {
		gl.UseProgram(shader.ID)
		sm.current = shader
	}
}

// CurrentShader returns the currently active shader
func (sm *ShaderManager) CurrentShader() *ShaderProgram {
	if sm == nil {
		return nil
	}
	return sm.current
}

// GetUniformLocation returns the cached uniform location
func (sm *ShaderManager) GetUniformLocation(name string) int32 {
	if sm == nil || sm.current == nil || gl.GetUniformLocation == nil {
		return -1
	}

	if loc, ok := sm.current.uniformLocs[name]; ok {
		return loc
	}

	loc := gl.GetUniformLocation(sm.current.ID, strPtr(name+"\x00"))
	sm.current.uniformLocs[name] = loc
	return loc
}

// SetFloat sets a float uniform
func (sm *ShaderManager) SetFloat(name string, value float32) {
	loc := sm.GetUniformLocation(name)
	if loc >= 0 && gl.Uniform1f != nil {
		gl.Uniform1f(loc, value)
	}
}

// SetVec2 sets a vec2 uniform
func (sm *ShaderManager) SetVec2(name string, x, y float32) {
	loc := sm.GetUniformLocation(name)
	if loc >= 0 && gl.Uniform2f != nil {
		gl.Uniform2f(loc, x, y)
	}
}

// SetVec4 sets a vec4 uniform
func (sm *ShaderManager) SetVec4(name string, x, y, z, w float32) {
	loc := sm.GetUniformLocation(name)
	if loc >= 0 && gl.Uniform4f != nil {
		gl.Uniform4f(loc, x, y, z, w)
	}
}

// SetInt sets an int uniform
func (sm *ShaderManager) SetInt(name string, value int32) {
	loc := sm.GetUniformLocation(name)
	if loc >= 0 && gl.Uniform1i != nil {
		gl.Uniform1i(loc, value)
	}
}

// SetMat4 sets a mat4 uniform
func (sm *ShaderManager) SetMat4(name string, mat *[16]float32) {
	if mat == nil {
		return
	}
	loc := sm.GetUniformLocation(name)
	if loc >= 0 && gl.UniformMatrix4fv != nil {
		gl.UniformMatrix4fv(loc, 1, false, &mat[0])
	}
}

// Delete deletes all shaders
func (sm *ShaderManager) Delete() {
	if sm == nil {
		return
	}
	for _, shader := range sm.shaders {
		if shader != nil && shader.ID != 0 && gl.DeleteProgram != nil {
			gl.DeleteProgram(shader.ID)
		}
	}
	sm.shaders = make(map[string]*ShaderProgram)
	sm.current = nil
}

// compileShader compiles a shader
func compileShader(source string, shaderType uint32) uint32 {
	if source == "" || !compileShaderGLReady() {
		return 0
	}
	shader := gl.CreateShader(shaderType)
	cs := cStringPtr(source)
	gl.ShaderSource(shader, 1, &cs, nil)
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.GL_COMPILE_STATUS, &status)
	if status == gl.GL_FALSE {
		var logLen int32
		gl.GetShaderiv(shader, gl.GL_INFO_LOG_LENGTH, &logLen)
		log := make([]byte, logLen+1)
		gl.GetShaderInfoLog(shader, logLen, nil, &log[0])
		fmt.Printf("shader compile error: %s\n", string(log))
		gl.DeleteShader(shader)
		return 0
	}

	return shader
}

func strPtr(s string) *byte {
	if s == "" {
		return nil
	}
	return &([]byte(s))[0]
}

func cStringPtr(s string) uintptr {
	if s == "" {
		return 0
	}
	return uintptr(unsafe.Pointer(&([]byte(s))[0]))
}

func shaderGLReady() bool {
	return compileShaderGLReady() &&
		gl.CreateProgram != nil &&
		gl.AttachShader != nil &&
		gl.BindAttribLocation != nil &&
		gl.LinkProgram != nil &&
		gl.GetProgramiv != nil &&
		gl.GetProgramInfoLog != nil &&
		gl.DeleteProgram != nil
}

func compileShaderGLReady() bool {
	return gl.CreateShader != nil &&
		gl.ShaderSource != nil &&
		gl.CompileShader != nil &&
		gl.GetShaderiv != nil &&
		gl.GetShaderInfoLog != nil &&
		gl.DeleteShader != nil
}

// BuiltinShaderSources contains the source code for built-in shaders
var BuiltinShaderSources = map[string][2]string{
	"color": {
		// Vertex shader
		`#version 120
attribute vec2 aPos;
attribute vec2 aUV;
attribute vec4 aColor;
varying vec2 vPos;
varying vec4 vColor;
uniform mat4 uProj;

void main() {
    vPos = aPos;
    vColor = aColor;
    gl_Position = uProj * vec4(aPos, 0.0, 1.0);
}
` + "\x00",
		// Fragment shader with rounded clip support
		`#version 120
varying vec2 vPos;
varying vec4 vColor;
uniform vec4 uClipRect;
uniform float uClipRadius;

void main() {
    if (uClipRadius > 0.0) {
        vec2 center = uClipRect.xy + uClipRect.zw * 0.5;
        vec2 q = abs(vPos - center) - (uClipRect.zw * 0.5 - vec2(uClipRadius));
        float d = length(max(q, 0.0)) - uClipRadius;
        if (d > 0.5) discard;
    }
    gl_FragColor = vColor;
}
` + "\x00",
	},
	"texture": {
		// Vertex shader
		`#version 120
attribute vec2 aPos;
attribute vec2 aUV;
attribute vec4 aColor;
varying vec2 vPos;
varying vec2 vUV;
varying vec4 vColor;
uniform mat4 uProj;

void main() {
    vPos = aPos;
    vUV = aUV;
    vColor = aColor;
    gl_Position = uProj * vec4(aPos, 0.0, 1.0);
}
` + "\x00",
		// Fragment shader with rounded clip support
		`#version 120
varying vec2 vPos;
varying vec2 vUV;
varying vec4 vColor;
uniform sampler2D uTex;
uniform vec4 uClipRect;
uniform float uClipRadius;

void main() {
    if (uClipRadius > 0.0) {
        vec2 center = uClipRect.xy + uClipRect.zw * 0.5;
        vec2 q = abs(vPos - center) - (uClipRect.zw * 0.5 - vec2(uClipRadius));
        float d = length(max(q, 0.0)) - uClipRadius;
        if (d > 0.5) discard;
    }
    gl_FragColor = texture2D(uTex, vUV) * vColor;
}
` + "\x00",
	},
	"rounded_rect": {
		// Vertex shader
		`#version 120
attribute vec2 aPos;
attribute vec2 aUV;
attribute vec4 aColor;
varying vec2 vUV;
varying vec4 vColor;
uniform mat4 uProj;

void main() {
    vUV = aUV;
    vColor = aColor;
    gl_Position = uProj * vec4(aPos, 0.0, 1.0);
}
` + "\x00",
		// Fragment shader (anti-aliased)
		`#version 120
varying vec2 vUV;
varying vec4 vColor;
uniform vec2 uSize;
uniform float uRadius;

void main() {
    vec2 pos = vUV * uSize;
    vec2 center = uSize * 0.5;
    float maxRadius = min(center.x, center.y);
    float radius = min(uRadius, maxRadius);
    vec2 q = abs(pos - center) - (center - vec2(radius));
    float d = length(max(q, 0.0)) - radius;

    // Screen-space SDF antialiasing. Keep at least one pixel of transition so
    // small controls do not look jagged on low DPI displays.
    float aa = max(length(vec2(dFdx(d), dFdy(d))), 0.75);
    float alpha = 1.0 - smoothstep(-aa, aa, d);

    gl_FragColor = vec4(vColor.rgb, vColor.a * alpha);
}
` + "\x00",
	},
	"rounded_rect_stroke": {
		// Vertex shader
		`#version 120
attribute vec2 aPos;
attribute vec2 aUV;
attribute vec4 aColor;
varying vec2 vUV;
varying vec4 vColor;
uniform mat4 uProj;

void main() {
    vUV = aUV;
    vColor = aColor;
    gl_Position = uProj * vec4(aPos, 0.0, 1.0);
}
` + "\x00",
		// Fragment shader for transparent rounded-rect stroke.
		`#version 120
varying vec2 vUV;
varying vec4 vColor;
uniform vec2 uSize;
uniform float uRadius;
uniform float uWidth;

void main() {
    vec2 pos = vUV * uSize;
    vec2 center = uSize * 0.5;
    float maxRadius = min(center.x, center.y);
    float radius = min(uRadius, maxRadius);
    vec2 q = abs(pos - center) - (center - vec2(radius));
    float d = length(max(q, 0.0)) - radius;

    float strokeWidth = max(uWidth, 0.0);
    float aa = max(length(vec2(dFdx(d), dFdy(d))), 0.75);
    float halfWidth = strokeWidth * 0.5;
    float strokeCenter = -halfWidth;
    float alpha = 1.0 - smoothstep(halfWidth - aa, halfWidth + aa, abs(d - strokeCenter));

    gl_FragColor = vec4(vColor.rgb, vColor.a * alpha);
}
` + "\x00",
	},
	"gradient": {
		// Vertex shader for gradients
		`#version 120
attribute vec2 aPos;
attribute vec2 aUV;
attribute vec4 aColor;
varying vec2 vUV;
varying vec4 vColor;
varying vec2 vPos;
uniform mat4 uProj;

void main() {
    vUV = aUV;
    vColor = aColor;
    vPos = aPos;
    gl_Position = uProj * vec4(aPos, 0.0, 1.0);
}
` + "\x00",
		// Fragment shader for linear gradient
		`#version 120
varying vec2 vUV;
varying vec4 vColor;
varying vec2 vPos;
uniform vec4 uColorStart;
uniform vec4 uColorEnd;
uniform vec2 uStart;
uniform vec2 uEnd;
uniform vec2 uSize;
uniform float uRadius;
uniform float uUseRadius;

void main() {
    // Use UV coordinates (0-1) for gradient calculation to avoid transform issues
    vec2 gradDir = uEnd - uStart;
    float gradLen = max(length(gradDir), 0.001);
    vec2 gradNorm = gradDir / gradLen;
    float t = dot(vUV - uStart, gradNorm) / gradLen;
    t = clamp(t, 0.0, 1.0);

    // Interpolate colors
    vec4 color = mix(uColorStart, uColorEnd, t);
    if (uUseRadius > 0.5) {
        vec2 pos = vUV * uSize;
        vec2 center = uSize * 0.5;
        vec2 q = abs(pos - center) - (center - vec2(uRadius));
        float d = length(max(q, 0.0)) - uRadius;
        float pixelLength = length(vec2(dFdx(d), dFdy(d)));
        float aa = max(pixelLength * 1.5, 0.001);
        float alpha = 1.0 - smoothstep(-aa, aa, d);
        color.a *= alpha;
    }
    gl_FragColor = color;
}
` + "\x00",
	},
	"circle": {
		// Vertex shader
		`#version 120
attribute vec2 aPos;
attribute vec2 aUV;
attribute vec4 aColor;
varying vec2 vUV;
varying vec4 vColor;
varying vec2 vPos;
uniform mat4 uProj;

void main() {
    vUV = aUV;
    vColor = aColor;
    vPos = aPos;
    gl_Position = uProj * vec4(aPos, 0.0, 1.0);
}
` + "\x00",
		// Fragment shader for circle with anti-aliasing
		`#version 120
varying vec2 vUV;
varying vec4 vColor;
varying vec2 vPos;
uniform vec2 uCenter;
uniform float uRadius;
uniform float uWidth; // 0 = filled, >0 = stroke

void main() {
    float dist = length(vPos - uCenter);

    if (uWidth <= 0.0) {
        // Filled circle
        float aa = 1.5;
        float alpha = 1.0 - smoothstep(uRadius - aa, uRadius, dist);
        gl_FragColor = vec4(vColor.rgb, vColor.a * alpha);
    } else {
        // Stroke circle
        float innerRadius = uRadius - uWidth;
        float aa = 1.5;
        float alphaOuter = 1.0 - smoothstep(uRadius - aa, uRadius, dist);
        float alphaInner = smoothstep(innerRadius - aa, innerRadius, dist);
        float alpha = alphaOuter * alphaInner;
        gl_FragColor = vec4(vColor.rgb, vColor.a * alpha);
    }
}
` + "\x00",
	},
	"shadow": {
		// Vertex shader
		`#version 120
attribute vec2 aPos;
attribute vec2 aUV;
attribute vec4 aColor;
varying vec2 vUV;
varying vec4 vColor;
uniform mat4 uProj;

void main() {
    vUV = aUV;
    vColor = aColor;
    gl_Position = uProj * vec4(aPos, 0.0, 1.0);
}
` + "\x00",
		// Fragment shader for shadow with SDF-based blur
		`#version 120
varying vec2 vUV;
varying vec4 vColor;
uniform vec2 uSize;
uniform float uRadius;
uniform float uBlur;

float roundRectSDF(vec2 pos, vec2 size, float radius) {
    vec2 center = size * 0.5;
    vec2 q = abs(pos - center) - (center - vec2(radius));
    return length(max(q, 0.0)) - radius;
}

void main() {
    vec2 pos = vUV * uSize;
    float d = roundRectSDF(pos, uSize, uRadius);

    // Smooth falloff based on blur radius
    float blur = max(uBlur, 1.0);
    float alpha = 1.0 - smoothstep(-blur, blur * 0.5, d);

    // Additional soft edge falloff
    float softEdge = 1.0 - smoothstep(0.0, blur * 2.0, d);
    alpha *= softEdge;

    gl_FragColor = vec4(vColor.rgb, vColor.a * alpha);
}
` + "\x00",
	},
	"line": {
		// Vertex shader for line
		`#version 120
attribute vec2 aPos;
attribute vec2 aUV;
attribute vec4 aColor;
varying vec2 vUV;
varying vec4 vColor;
varying vec2 vPos;
uniform mat4 uProj;

void main() {
    vUV = aUV;
    vColor = aColor;
    vPos = aPos;
    gl_Position = uProj * vec4(aPos, 0.0, 1.0);
}
` + "\x00",
		// Fragment shader for line with SDF anti-aliasing
		`#version 120
varying vec2 vUV;
varying vec4 vColor;
varying vec2 vPos;
uniform vec2 uLineStart;
uniform vec2 uLineEnd;
uniform float uLineWidth;

void main() {
    vec2 lineDir = uLineEnd - uLineStart;
    float lineLen = length(lineDir);
    if (lineLen < 0.001) {
        discard;
    }
    vec2 lineNorm = lineDir / lineLen;

    // Project point onto line
    vec2 toPoint = vPos - uLineStart;
    float proj = dot(toPoint, lineNorm);
    proj = clamp(proj, 0.0, lineLen);

    // Calculate distance to closest point on line
    vec2 closest = uLineStart + lineNorm * proj;
    float dist = length(vPos - closest);

    // Anti-aliasing using screen-space derivatives
    float halfWidth = uLineWidth * 0.5;
    float aa = max(length(vec2(dFdx(dist), dFdy(dist))), 0.75);
    float alpha = 1.0 - smoothstep(halfWidth - aa, halfWidth + aa, dist);

    gl_FragColor = vec4(vColor.rgb, vColor.a * alpha);
}
` + "\x00",
	},
	"triangle": {
		// Vertex shader for triangle
		`#version 120
attribute vec2 aPos;
attribute vec2 aUV;
attribute vec4 aColor;
varying vec2 vUV;
varying vec4 vColor;
varying vec2 vPos;
uniform mat4 uProj;

void main() {
    vUV = aUV;
    vColor = aColor;
    vPos = aPos;
    gl_Position = uProj * vec4(aPos, 0.0, 1.0);
}
` + "\x00",
		// Fragment shader for triangle with SDF anti-aliasing
		`#version 120
varying vec2 vUV;
varying vec4 vColor;
varying vec2 vPos;
uniform vec2 uV0;
uniform vec2 uV1;
uniform vec2 uV2;

float cross2d(vec2 a, vec2 b) {
    return a.x * b.y - a.y * b.x;
}

float distToEdge(vec2 p, vec2 a, vec2 b) {
    vec2 ab = b - a;
    vec2 ap = p - a;
    float t = clamp(dot(ap, ab) / dot(ab, ab), 0.0, 1.0);
    vec2 closest = a + t * ab;
    return length(p - closest);
}

void main() {
    // Calculate signed distances to each edge
    float d0 = cross2d(uV1 - uV0, vPos - uV0);
    float d1 = cross2d(uV2 - uV1, vPos - uV1);
    float d2 = cross2d(uV0 - uV2, vPos - uV2);

    // Check if point is inside triangle
    bool inside = (d0 >= 0.0 && d1 >= 0.0 && d2 >= 0.0) || (d0 <= 0.0 && d1 <= 0.0 && d2 <= 0.0);

    // Calculate distance to nearest edge
    float dist0 = distToEdge(vPos, uV0, uV1);
    float dist1 = distToEdge(vPos, uV1, uV2);
    float dist2 = distToEdge(vPos, uV2, uV0);
    float dist = min(dist0, min(dist1, dist2));

    // Anti-aliasing
    float aa = max(length(vec2(dFdx(dist), dFdy(dist))), 0.75);
    float alpha = inside ? (1.0 - smoothstep(0.0, aa, dist)) : 0.0;

    gl_FragColor = vec4(vColor.rgb, vColor.a * alpha);
}
` + "\x00",
	},
}
