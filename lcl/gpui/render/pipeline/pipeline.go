// Package pipeline provides the rendering pipeline and batch management
package pipeline

import (
	"unsafe"

	"github.com/energye/examples/lcl/gpui/core/gl"
	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/shader"
)

// Vertex represents a vertex with position, UV, and color
type Vertex struct {
	X, Y       float32 // Position
	U, V       float32 // Texture coordinates
	R, G, B, A float32 // Color
}

// VertexSize is the size of a vertex in bytes
const VertexSize = 8 * 4 // 8 floats * 4 bytes

// QuadVertices creates 4 vertices for a textured/solid rectangle
func QuadVertices(rect math.Rect, uv math.Rect, color math.Color) [4]Vertex {
	return [4]Vertex{
		{X: rect.X, Y: rect.Y, U: uv.X, V: uv.Y, R: color.R, G: color.G, B: color.B, A: color.A},
		{X: rect.X + rect.W, Y: rect.Y, U: uv.X + uv.W, V: uv.Y, R: color.R, G: color.G, B: color.B, A: color.A},
		{X: rect.X + rect.W, Y: rect.Y + rect.H, U: uv.X + uv.W, V: uv.Y + uv.H, R: color.R, G: color.G, B: color.B, A: color.A},
		{X: rect.X, Y: rect.Y + rect.H, U: uv.X, V: uv.Y + uv.H, R: color.R, G: color.G, B: color.B, A: color.A},
	}
}

// Batch represents a draw batch
type Batch struct {
	Shader  *shader.ShaderProgram
	Texture uint32
	Verts   []Vertex
	Indices []uint32
}

// BatchManager manages draw batches
type BatchManager struct {
	batches     []*Batch
	current     *Batch
	maxVertices int
	maxIndices  int
}

// NewBatchManager creates a new batch manager
func NewBatchManager(maxVerts, maxIndices int) *BatchManager {
	return &BatchManager{
		batches:     make([]*Batch, 0, 16),
		maxVertices: maxVerts,
		maxIndices:  maxIndices,
	}
}

// Reset resets the batch manager
func (bm *BatchManager) Reset() {
	bm.batches = bm.batches[:0]
	bm.current = nil
}

// AddQuad adds a quad to the batch
func (bm *BatchManager) AddQuad(shaderProg *shader.ShaderProgram, texture uint32, verts [4]Vertex) {
	// Check if we need a new batch
	if bm.current == nil ||
		bm.current.Shader != shaderProg ||
		bm.current.Texture != texture {

		// Save current batch
		if bm.current != nil && len(bm.current.Verts) > 0 {
			bm.batches = append(bm.batches, bm.current)
		}

		// Create new batch
		bm.current = &Batch{
			Shader:  shaderProg,
			Texture: texture,
		}
	}

	// Add vertices
	offset := uint32(len(bm.current.Verts))
	bm.current.Verts = append(bm.current.Verts, verts[0], verts[1], verts[2], verts[3])
	bm.current.Indices = append(bm.current.Indices, offset, offset+1, offset+2, offset, offset+2, offset+3)
}

// Flush flushes all batches
func (bm *BatchManager) Flush(vao, vbo, ebo uint32, shaderMgr *shader.ShaderManager, projMatrix *[16]float32) {
	// Add current batch
	if bm.current != nil && len(bm.current.Verts) > 0 {
		bm.batches = append(bm.batches, bm.current)
		bm.current = nil
	}

	if len(bm.batches) == 0 {
		return
	}

	// Bind VAO
	gl.BindVertexArray(vao)

	for _, batch := range bm.batches {
		if len(batch.Verts) == 0 {
			continue
		}

		// Use shader
		shaderMgr.UseShader(batch.Shader)
		shaderMgr.SetMat4("uProj", projMatrix)

		// Bind texture if needed
		if batch.Texture > 0 {
			gl.ActiveTexture(gl.GL_TEXTURE0)
			gl.BindTexture(gl.GL_TEXTURE_2D, batch.Texture)
			shaderMgr.SetInt("uTex", 0)
		}

		// Upload vertex data
		gl.BindBuffer(gl.GL_ARRAY_BUFFER, vbo)
		vertSize := len(batch.Verts) * VertexSize
		vertPtr := uintptr(unsafe.Pointer(&batch.Verts[0]))
		gl.BufferSubData(gl.GL_ARRAY_BUFFER, 0, int32(vertSize), vertPtr)

		// Upload index data
		gl.BindBuffer(gl.GL_ELEMENT_ARRAY_BUFFER, ebo)
		idxSize := len(batch.Indices) * 4
		idxPtr := uintptr(unsafe.Pointer(&batch.Indices[0]))
		gl.BufferSubData(gl.GL_ELEMENT_ARRAY_BUFFER, 0, int32(idxSize), idxPtr)

		// Draw
		gl.DrawElements(gl.GL_TRIANGLES, int32(len(batch.Indices)), gl.GL_UNSIGNED_INT, 0)
	}

	// Unbind
	gl.BindVertexArray(0)

	// Reset
	bm.Reset()
}

// Renderer is the main renderer
type Renderer struct {
	vao        uint32
	vbo        uint32
	ebo        uint32
	shaderMgr  *shader.ShaderManager
	batch      *BatchManager
	projMatrix [16]float32
	width      float32
	height     float32
}

// NewRenderer creates a new renderer
func NewRenderer() *Renderer {
	return &Renderer{
		shaderMgr: shader.NewShaderManager(),
		batch:     NewBatchManager(65536, 65536),
	}
}

// Init initializes the renderer
func (r *Renderer) Init() error {
	// Load GL functions
	if err := gl.Init(); err != nil {
		return err
	}

	// Create VAO
	gl.GenVertexArrays(1, &r.vao)
	gl.BindVertexArray(r.vao)

	// Create VBO
	gl.GenBuffers(1, &r.vbo)
	gl.BindBuffer(gl.GL_ARRAY_BUFFER, r.vbo)
	gl.BufferData(gl.GL_ARRAY_BUFFER, 65536*VertexSize, 0, gl.GL_DYNAMIC_DRAW)

	// Create EBO
	gl.GenBuffers(1, &r.ebo)
	gl.BindBuffer(gl.GL_ELEMENT_ARRAY_BUFFER, r.ebo)
	gl.BufferData(gl.GL_ELEMENT_ARRAY_BUFFER, 65536*4, 0, gl.GL_DYNAMIC_DRAW)

	// Setup vertex attributes
	stride := int32(VertexSize)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.GL_FLOAT, false, stride, 0) // Position
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 2, gl.GL_FLOAT, false, stride, 8) // UV
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 4, gl.GL_FLOAT, false, stride, 16) // Color

	// Unbind
	gl.BindVertexArray(0)

	// Load built-in shaders
	for name, sources := range shader.BuiltinShaderSources {
		_, err := r.shaderMgr.LoadShader(name, sources[0], sources[1])
		if err != nil {
			return err
		}
	}

	// Enable blending
	gl.Enable(gl.GL_BLEND)
	gl.BlendFunc(gl.GL_SRC_ALPHA, gl.GL_ONE_MINUS_SRC_ALPHA)
	gl.Disable(gl.GL_DEPTH_TEST)

	return nil
}

// BeginFrame begins a new frame
func (r *Renderer) BeginFrame(width, height float32) {
	r.width = width
	r.height = height

	// Set viewport
	gl.Viewport(0, 0, int32(width), int32(height))

	// Set projection matrix
	r.projMatrix = math.OrthoMatrix(0, width, height, 0, -1, 1)

	// Clear
	gl.ClearColor(0.15, 0.15, 0.17, 1.0)
	gl.Clear(gl.GL_COLOR_BUFFER_BIT)

	// Reset batch
	r.batch.Reset()
}

// EndFrame ends the frame
func (r *Renderer) EndFrame() {
	r.Flush()
}

// Flush flushes all pending draw calls
func (r *Renderer) Flush() {
	r.batch.Flush(r.vao, r.vbo, r.ebo, r.shaderMgr, &r.projMatrix)
}

// FillRect draws a filled rectangle
func (r *Renderer) FillRect(rect math.Rect, color math.Color) {
	shaderProg := r.shaderMgr.GetShader("color")
	uv := math.NewRect(0, 0, 1, 1)
	verts := QuadVertices(rect, uv, color)
	r.batch.AddQuad(shaderProg, 0, verts)
}

// FillRoundRect draws a filled rounded rectangle
func (r *Renderer) FillRoundRect(rect math.Rect, radius float32, color math.Color) {
	// This shader uses uniforms that vary per primitive. Flush around it so
	// queued rounded rects with different sizes/radii cannot share stale state.
	r.Flush()

	shaderProg := r.shaderMgr.GetShader("rounded_rect")
	r.shaderMgr.UseShader(shaderProg)
	r.shaderMgr.SetFloat("uRadius", radius)
	r.shaderMgr.SetVec2("uSize", rect.W, rect.H)

	uv := math.NewRect(0, 0, 1, 1)
	verts := QuadVertices(rect, uv, color)
	r.batch.AddQuad(shaderProg, 0, verts)
	r.Flush()
}

// DrawTexture draws a textured rectangle
func (r *Renderer) DrawTexture(texture uint32, src, dst math.Rect, color math.Color) {
	shaderProg := r.shaderMgr.GetShader("texture")
	verts := QuadVertices(dst, src, color)
	r.batch.AddQuad(shaderProg, texture, verts)
}

// Delete deletes all resources
func (r *Renderer) Delete() {
	r.shaderMgr.Delete()

	if r.vao != 0 {
		gl.DeleteVertexArrays(1, &r.vao)
		r.vao = 0
	}
	if r.vbo != 0 {
		gl.DeleteBuffers(1, &r.vbo)
		r.vbo = 0
	}
	if r.ebo != 0 {
		gl.DeleteBuffers(1, &r.ebo)
		r.ebo = 0
	}
}

// ShaderManager returns the shader manager
func (r *Renderer) ShaderManager() *shader.ShaderManager {
	return r.shaderMgr
}

// ProjectionMatrix returns the projection matrix
func (r *Renderer) ProjectionMatrix() *[16]float32 {
	return &r.projMatrix
}
