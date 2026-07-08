// Package pipeline provides FBO (Framebuffer Object) management for offscreen rendering.
package pipeline

import (
	"fmt"

	"github.com/energye/examples/lcl/gpui/core/gl"
)

// Framebuffer wraps an OpenGL framebuffer object for offscreen rendering.
type Framebuffer struct {
	fbo        uint32
	colorTex   uint32
	depthRbo   uint32
	width      int32
	height     int32
	ownsColor  bool
}

// FramebufferConfig configures framebuffer creation.
type FramebufferConfig struct {
	Width      int32
	Height     int32
	ColorTex   uint32 // If 0, a new color texture is created
	DepthBuffer bool  // If true, create a depth renderbuffer
}

// NewFramebuffer creates a new framebuffer object.
func NewFramebuffer(config FramebufferConfig) (*Framebuffer, error) {
	if gl.GenFramebuffers == nil {
		return nil, fmt.Errorf("FBO not supported (GL 3.0+ required)")
	}

	fb := &Framebuffer{
		width:  config.Width,
		height: config.Height,
	}

	// Generate FBO
	gl.GenFramebuffers(1, &fb.fbo)
	if fb.fbo == 0 {
		return nil, fmt.Errorf("failed to create framebuffer")
	}

	gl.BindFramebuffer(gl.GL_FRAMEBUFFER, fb.fbo)

	// Color attachment
	if config.ColorTex > 0 {
		fb.colorTex = config.ColorTex
		fb.ownsColor = false
	} else {
		// Create new color texture
		gl.GenTextures(1, &fb.colorTex)
		if fb.colorTex == 0 {
			gl.DeleteFramebuffers(1, &fb.fbo)
			return nil, fmt.Errorf("failed to create color texture")
		}
		fb.ownsColor = true

		gl.BindTexture(gl.GL_TEXTURE_2D, fb.colorTex)
		gl.TexImage2D(gl.GL_TEXTURE_2D, 0, int32(gl.GL_RGBA), config.Width, config.Height, 0,
			gl.GL_RGBA, gl.GL_UNSIGNED_BYTE, 0)
		gl.TexParameteri(gl.GL_TEXTURE_2D, gl.GL_TEXTURE_MIN_FILTER, gl.GL_LINEAR)
		gl.TexParameteri(gl.GL_TEXTURE_2D, gl.GL_TEXTURE_MAG_FILTER, gl.GL_LINEAR)
		gl.TexParameteri(gl.GL_TEXTURE_2D, gl.GL_TEXTURE_WRAP_S, gl.GL_CLAMP_TO_EDGE)
		gl.TexParameteri(gl.GL_TEXTURE_2D, gl.GL_TEXTURE_WRAP_T, gl.GL_CLAMP_TO_EDGE)
		gl.BindTexture(gl.GL_TEXTURE_2D, 0)

		gl.FramebufferTexture2D(gl.GL_FRAMEBUFFER, gl.GL_COLOR_ATTACHMENT0, gl.GL_TEXTURE_2D, fb.colorTex, 0)
	}

	// Depth buffer
	if config.DepthBuffer && gl.GenRenderbuffers != nil {
		gl.GenRenderbuffers(1, &fb.depthRbo)
		if fb.depthRbo > 0 {
			gl.BindRenderbuffer(gl.GL_RENDERBUFFER, fb.depthRbo)
			gl.RenderbufferStorage(gl.GL_RENDERBUFFER, gl.GL_DEPTH24_STENCIL8, config.Width, config.Height)
			gl.FramebufferRenderbuffer(gl.GL_FRAMEBUFFER, gl.GL_DEPTH_STENCIL_ATTACHMENT, gl.GL_RENDERBUFFER, fb.depthRbo)
			gl.BindRenderbuffer(gl.GL_RENDERBUFFER, 0)
		}
	}

	// Check completeness
	status := gl.CheckFramebufferStatus(gl.GL_FRAMEBUFFER)
	if status != gl.GL_FRAMEBUFFER_COMPLETE {
		fb.Delete()
		return nil, fmt.Errorf("framebuffer incomplete: status=%d", status)
	}

	// Unbind
	gl.BindFramebuffer(gl.GL_FRAMEBUFFER, 0)

	return fb, nil
}

// Bind binds the framebuffer for rendering.
func (fb *Framebuffer) Bind() {
	if fb == nil || fb.fbo == 0 {
		return
	}
	gl.BindFramebuffer(gl.GL_FRAMEBUFFER, fb.fbo)
	gl.Viewport(0, 0, fb.width, fb.height)
}

// Unbind unbinds the framebuffer (renders to default framebuffer).
func (fb *Framebuffer) Unbind() {
	if gl.BindFramebuffer != nil {
		gl.BindFramebuffer(gl.GL_FRAMEBUFFER, 0)
	}
}

// ColorTexture returns the color attachment texture ID.
func (fb *Framebuffer) ColorTexture() uint32 {
	if fb == nil {
		return 0
	}
	return fb.colorTex
}

// Width returns the framebuffer width.
func (fb *Framebuffer) Width() int32 {
	if fb == nil {
		return 0
	}
	return fb.width
}

// Height returns the framebuffer height.
func (fb *Framebuffer) Height() int32 {
	if fb == nil {
		return 0
	}
	return fb.height
}

// Delete releases all framebuffer resources.
func (fb *Framebuffer) Delete() {
	if fb == nil {
		return
	}
	if fb.depthRbo > 0 && gl.DeleteRenderbuffers != nil {
		gl.DeleteRenderbuffers(1, &fb.depthRbo)
		fb.depthRbo = 0
	}
	if fb.ownsColor && fb.colorTex > 0 && gl.DeleteTextures != nil {
		gl.DeleteTextures(1, &fb.colorTex)
		fb.colorTex = 0
	}
	if fb.fbo > 0 && gl.DeleteFramebuffers != nil {
		gl.DeleteFramebuffers(1, &fb.fbo)
		fb.fbo = 0
	}
}

// FBOSupported reports whether FBO functions are available.
func FBOSupported() bool {
	return gl.GenFramebuffers != nil && gl.BindFramebuffer != nil && gl.CheckFramebufferStatus != nil
}
