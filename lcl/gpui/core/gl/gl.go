// Package gl provides OpenGL bindings via purego (non-CGo)
package gl

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/ebitengine/purego"
)

// OpenGL constants
const (
	GL_FALSE = 0
	GL_TRUE  = 1

	GL_ZERO                = 0
	GL_ONE                 = 1
	GL_SRC_ALPHA           = 0x0302
	GL_ONE_MINUS_SRC_ALPHA = 0x0303

	GL_BLEND        = 0x0BE2
	GL_SCISSOR_TEST = 0x0C11
	GL_DEPTH_TEST   = 0x0B71
	GL_CULL_FACE    = 0x0B44
	GL_STENCIL_TEST = 0x0B90

	GL_ALWAYS    = 0x0207
	GL_EQUAL     = 0x0202
	GL_NOTEQUAL  = 0x0205
	GL_KEEP      = 0x1E00
	GL_INVERT    = 0x150A
	GL_INCR_WRAP = 0x8507
	GL_DECR_WRAP = 0x8508

	GL_TEXTURE_2D           = 0x0DE1
	GL_RGBA                 = 0x1908
	GL_UNSIGNED_BYTE        = 0x1401
	GL_LINEAR               = 0x2601
	GL_NEAREST              = 0x2600
	GL_LINEAR_MIPMAP_LINEAR = 0x2703
	GL_CLAMP_TO_EDGE        = 0x812F
	GL_REPEAT               = 0x2901
	GL_TEXTURE_MIN_FILTER   = 0x2801
	GL_TEXTURE_MAG_FILTER   = 0x2800
	GL_TEXTURE_WRAP_S       = 0x2802
	GL_TEXTURE_WRAP_T       = 0x2803

	GL_ARRAY_BUFFER         = 0x8892
	GL_ELEMENT_ARRAY_BUFFER = 0x8893
	GL_STATIC_DRAW          = 0x88E4
	GL_DYNAMIC_DRAW         = 0x88E8
	GL_STREAM_DRAW          = 0x88E0
	GL_FLOAT                = 0x1406
	GL_UNSIGNED_INT         = 0x1405
	GL_TRIANGLES            = 0x0004
	GL_TRIANGLE_STRIP       = 0x0005
	GL_TRIANGLE_FAN         = 0x0006

	GL_VERTEX_SHADER   = 0x8B31
	GL_FRAGMENT_SHADER = 0x8B30
	GL_COMPILE_STATUS  = 0x8B81
	GL_LINK_STATUS     = 0x8B82
	GL_INFO_LOG_LENGTH = 0x8B84

	GL_COLOR_BUFFER_BIT   = 0x00004000
	GL_DEPTH_BUFFER_BIT   = 0x00000100
	GL_STENCIL_BUFFER_BIT = 0x00000400

	GL_TEXTURE0 = 0x84C0
)

// GL function pointers
var (
	// State management
	Viewport     func(x, y, width, height int32)
	ClearColor   func(r, g, b, a float32)
	Clear        func(mask uint32)
	Enable       func(cap uint32)
	Disable      func(cap uint32)
	BlendFunc    func(sfactor, dfactor uint32)
	Scissor      func(x, y, width, height int32)
	ColorMask    func(red, green, blue, alpha bool)
	ClearStencil func(s int32)
	StencilFunc  func(fn uint32, ref int32, mask uint32)
	StencilOp    func(sfail, dpfail, dppass uint32)
	GetError     func() uint32

	// Shader functions
	CreateShader      func(shaderType uint32) uint32
	ShaderSource      func(shader, count uint32, str *uintptr, length *int32)
	CompileShader     func(shader uint32)
	GetShaderiv       func(shader, pname uint32, params *int32)
	GetShaderInfoLog  func(shader uint32, bufSize int32, length *int32, infoLog *byte)
	CreateProgram     func() uint32
	AttachShader      func(program, shader uint32)
	LinkProgram       func(program uint32)
	GetProgramiv      func(program uint32, pname uint32, params *int32)
	GetProgramInfoLog func(program uint32, bufSize int32, length *int32, infoLog *byte)
	UseProgram        func(program uint32)
	DeleteShader      func(shader uint32)
	DeleteProgram     func(program uint32)

	// Uniform functions
	GetUniformLocation       func(program uint32, name *byte) int32
	Uniform1i                func(location int32, v0 int32)
	Uniform1f                func(location int32, v0 float32)
	Uniform2f                func(location int32, v0, v1 float32)
	Uniform4f                func(location int32, v0, v1, v2, v3 float32)
	UniformMatrix4fv         func(location int32, count int32, transpose bool, value *float32)
	GetAttribLocation        func(program uint32, name *byte) int32
	BindAttribLocation       func(program uint32, index uint32, name *byte)
	EnableVertexAttribArray  func(index uint32)
	DisableVertexAttribArray func(index uint32)
	VertexAttribPointer      func(index uint32, size int32, xtype uint32, normalized bool, stride int32, pointer uintptr)

	// Buffer functions
	GenBuffers    func(n int32, buffers *uint32)
	BindBuffer    func(target uint32, buffer uint32)
	BufferData    func(target uint32, size int32, data uintptr, usage uint32)
	BufferSubData func(target uint32, offset int32, size int32, data uintptr)
	DeleteBuffers func(n int32, buffers *uint32)

	// VAO functions
	GenVertexArrays    func(n int32, arrays *uint32)
	BindVertexArray    func(array uint32)
	DeleteVertexArrays func(n int32, arrays *uint32)

	// Draw functions
	DrawArrays   func(mode uint32, first int32, count int32)
	DrawElements func(mode uint32, count int32, xtype uint32, indices uintptr)
	ReadPixels   func(x, y, width, height int32, format uint32, xtype uint32, pixels uintptr)

	// Texture functions
	GenTextures    func(n int32, textures *uint32)
	DeleteTextures func(n int32, textures *uint32)
	BindTexture    func(target uint32, texture uint32)
	TexImage2D     func(target uint32, level int32, internalformat int32, width int32, height int32, border int32, format uint32, xtype uint32, pixels uintptr)
	TexSubImage2D  func(target uint32, level int32, xoffset int32, yoffset int32, width int32, height int32, format uint32, xtype uint32, pixels uintptr)
	TexParameteri  func(target uint32, pname uint32, param int32)
	ActiveTexture  func(texture uint32)
)

var initMu sync.Mutex
var initialized bool

// Init loads OpenGL functions using purego
func Init() error {
	initMu.Lock()
	defer initMu.Unlock()
	if initialized {
		return nil
	}

	lib, err := openGLLibrary()
	if err != nil {
		return fmt.Errorf("gl: %v", err)
	}

	// If Init fails partway, clear all partially-bound function pointers.
	initOK := false
	defer func() {
		if !initOK {
			resetGLFuncs()
		}
	}()

	bind := func(fn any, name string) {
		purego.RegisterLibFunc(fn, lib, name)
	}

	// Core OpenGL 1.x functions
	bind(&Viewport, "glViewport")
	bind(&ClearColor, "glClearColor")
	bind(&Clear, "glClear")
	bind(&Enable, "glEnable")
	bind(&Disable, "glDisable")
	bind(&BlendFunc, "glBlendFunc")
	bind(&Scissor, "glScissor")
	bind(&ColorMask, "glColorMask")
	bind(&ClearStencil, "glClearStencil")
	bind(&StencilFunc, "glStencilFunc")
	bind(&StencilOp, "glStencilOp")
	bind(&GetError, "glGetError")

	// Shader functions (GL 2.0+)
	bind(&CreateShader, "glCreateShader")
	bind(&ShaderSource, "glShaderSource")
	bind(&CompileShader, "glCompileShader")
	bind(&GetShaderiv, "glGetShaderiv")
	bind(&GetShaderInfoLog, "glGetShaderInfoLog")
	bind(&CreateProgram, "glCreateProgram")
	bind(&AttachShader, "glAttachShader")
	bind(&LinkProgram, "glLinkProgram")
	bind(&GetProgramiv, "glGetProgramiv")
	bind(&GetProgramInfoLog, "glGetProgramInfoLog")
	bind(&UseProgram, "glUseProgram")
	bind(&DeleteShader, "glDeleteShader")
	bind(&DeleteProgram, "glDeleteProgram")

	// Uniform functions
	bind(&GetUniformLocation, "glGetUniformLocation")
	bind(&Uniform1i, "glUniform1i")
	bind(&Uniform1f, "glUniform1f")
	bind(&Uniform2f, "glUniform2f")
	bind(&Uniform4f, "glUniform4f")
	bind(&UniformMatrix4fv, "glUniformMatrix4fv")
	bind(&GetAttribLocation, "glGetAttribLocation")
	bind(&BindAttribLocation, "glBindAttribLocation")
	bind(&EnableVertexAttribArray, "glEnableVertexAttribArray")
	bind(&DisableVertexAttribArray, "glDisableVertexAttribArray")
	bind(&VertexAttribPointer, "glVertexAttribPointer")

	// Buffer functions (GL 1.5+)
	if fn := getGLFunc(lib, "glGenBuffers"); fn != 0 {
		bind(&GenBuffers, "glGenBuffers")
		bind(&BindBuffer, "glBindBuffer")
		bind(&BufferData, "glBufferData")
		bind(&BufferSubData, "glBufferSubData")
		bind(&DeleteBuffers, "glDeleteBuffers")
	} else {
		return fmt.Errorf("GL 1.5+ not supported (glGenBuffers missing)")
	}

	// VAO (GL 3.0+ core or extension)
	if fn := getGLFunc(lib, "glGenVertexArrays"); fn != 0 {
		bind(&GenVertexArrays, "glGenVertexArrays")
		bind(&BindVertexArray, "glBindVertexArray")
		bind(&DeleteVertexArrays, "glDeleteVertexArrays")
	} else if fn := getGLFunc(lib, "glGenVertexArraysAPPLE"); fn != 0 {
		purego.RegisterLibFunc(&GenVertexArrays, lib, "glGenVertexArraysAPPLE")
		purego.RegisterLibFunc(&BindVertexArray, lib, "glBindVertexArrayAPPLE")
		purego.RegisterLibFunc(&DeleteVertexArrays, lib, "glDeleteVertexArraysAPPLE")
	} else {
		return fmt.Errorf("VAO not supported (needs GL 3.0+ or ARB_vertex_array_object)")
	}

	// Draw functions
	bind(&DrawArrays, "glDrawArrays")
	bind(&DrawElements, "glDrawElements")
	bind(&ReadPixels, "glReadPixels")

	// Texture functions
	bind(&GenTextures, "glGenTextures")
	bind(&DeleteTextures, "glDeleteTextures")
	bind(&BindTexture, "glBindTexture")
	bind(&TexImage2D, "glTexImage2D")
	bind(&TexSubImage2D, "glTexSubImage2D")
	bind(&TexParameteri, "glTexParameteri")
	bind(&ActiveTexture, "glActiveTexture")

	initOK = true
	initialized = true
	return nil
}

// resetGLFuncs clears all partially-bound function pointers so that
// callers cannot use a half-initialized GL state after Init fails.
func resetGLFuncs() {
	Viewport = nil
	ClearColor = nil
	Clear = nil
	Enable = nil
	Disable = nil
	BlendFunc = nil
	Scissor = nil
	ColorMask = nil
	ClearStencil = nil
	StencilFunc = nil
	StencilOp = nil
	GetError = nil

	CreateShader = nil
	ShaderSource = nil
	CompileShader = nil
	GetShaderiv = nil
	GetShaderInfoLog = nil
	CreateProgram = nil
	AttachShader = nil
	LinkProgram = nil
	GetProgramiv = nil
	GetProgramInfoLog = nil
	UseProgram = nil
	DeleteShader = nil
	DeleteProgram = nil

	GetUniformLocation = nil
	Uniform1i = nil
	Uniform1f = nil
	Uniform2f = nil
	Uniform4f = nil
	UniformMatrix4fv = nil
	GetAttribLocation = nil
	BindAttribLocation = nil
	EnableVertexAttribArray = nil
	DisableVertexAttribArray = nil
	VertexAttribPointer = nil

	GenBuffers = nil
	BindBuffer = nil
	BufferData = nil
	BufferSubData = nil
	DeleteBuffers = nil

	GenVertexArrays = nil
	BindVertexArray = nil
	DeleteVertexArrays = nil

	DrawArrays = nil
	DrawElements = nil
	ReadPixels = nil

	GenTextures = nil
	DeleteTextures = nil
	BindTexture = nil
	TexImage2D = nil
	TexSubImage2D = nil
	TexParameteri = nil
	ActiveTexture = nil
}

func openGLLibrary() (uintptr, error) {
	var flags int = purego.RTLD_LAZY | purego.RTLD_GLOBAL
	switch runtime.GOOS {
	case "linux":
		lib, err := purego.Dlopen("libGL.so.1", flags)
		if err != nil {
			lib, err = purego.Dlopen("libGL.so", flags)
			if err != nil {
				return 0, fmt.Errorf("cannot open libGL.so: %w", err)
			}
		}
		return lib, nil
	case "darwin":
		lib, err := purego.Dlopen("/System/Library/Frameworks/OpenGL.framework/OpenGL", flags)
		if err != nil {
			return 0, fmt.Errorf("cannot open OpenGL.framework: %w", err)
		}
		return lib, nil
	case "windows":
		lib, err := purego.Dlopen("opengl32.dll", flags)
		if err != nil {
			return 0, fmt.Errorf("cannot open opengl32.dll: %w", err)
		}
		return lib, nil
	default:
		return 0, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func getGLFunc(lib uintptr, name string) uintptr {
	fn, err := purego.Dlsym(lib, name)
	if err != nil {
		return 0
	}
	return fn
}
