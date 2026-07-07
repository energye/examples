// Package pipeline provides the rendering pipeline and batch management
package pipeline

import (
	"fmt"
	"sort"
	"strings"
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

// UniformKind describes a supported shader uniform value type.
type UniformKind int

const (
	UniformFloat UniformKind = iota
	UniformVec2
	UniformVec4
)

// UniformValue stores a typed uniform value for a draw batch.
type UniformValue struct {
	Kind   UniformKind
	Values [4]float32
}

// UniformSet stores per-batch uniforms.
type UniformSet map[string]UniformValue

// FloatUniform creates a float uniform value.
func FloatUniform(v float32) UniformValue {
	return UniformValue{Kind: UniformFloat, Values: [4]float32{v}}
}

// Vec2Uniform creates a vec2 uniform value.
func Vec2Uniform(x, y float32) UniformValue {
	return UniformValue{Kind: UniformVec2, Values: [4]float32{x, y}}
}

// Vec4Uniform creates a vec4 uniform value.
func Vec4Uniform(x, y, z, w float32) UniformValue {
	return UniformValue{Kind: UniformVec4, Values: [4]float32{x, y, z, w}}
}

func (u UniformSet) clone() UniformSet {
	if len(u) == 0 {
		return nil
	}
	out := make(UniformSet, len(u))
	for name, value := range u {
		out[name] = value
	}
	return out
}

func (u UniformSet) key() string {
	if len(u) == 0 {
		return ""
	}
	names := make([]string, 0, len(u))
	for name := range u {
		names = append(names, name)
	}
	sort.Strings(names)

	var b strings.Builder
	for _, name := range names {
		value := u[name]
		fmt.Fprintf(&b, "%s:%d:%g,%g,%g,%g;", name, value.Kind, value.Values[0], value.Values[1], value.Values[2], value.Values[3])
	}
	return b.String()
}

func applyUniforms(shaderMgr *shader.ShaderManager, uniforms UniformSet) {
	if shaderMgr == nil {
		return
	}
	for name, value := range uniforms {
		switch value.Kind {
		case UniformFloat:
			shaderMgr.SetFloat(name, value.Values[0])
		case UniformVec2:
			shaderMgr.SetVec2(name, value.Values[0], value.Values[1])
		case UniformVec4:
			shaderMgr.SetVec4(name, value.Values[0], value.Values[1], value.Values[2], value.Values[3])
		}
	}
}

func cloneRect(rect *math.Rect) *math.Rect {
	if rect == nil {
		return nil
	}
	copied := *rect
	return &copied
}

func rectKey(rect *math.Rect) string {
	if rect == nil {
		return ""
	}
	return fmt.Sprintf("%g,%g,%g,%g", rect.X, rect.Y, rect.W, rect.H)
}

func applyClip(rect *math.Rect, viewportHeight float32) {
	if gl.Disable == nil || gl.Enable == nil || gl.Scissor == nil {
		return
	}
	if rect == nil {
		gl.Disable(gl.GL_SCISSOR_TEST)
		return
	}

	if rect.W <= 0 || rect.H <= 0 {
		gl.Enable(gl.GL_SCISSOR_TEST)
		gl.Scissor(0, 0, 0, 0)
		return
	}

	x := int32(rect.X)
	y := int32(viewportHeight - rect.Y - rect.H)
	w := int32(rect.W)
	h := int32(rect.H)
	if y < 0 {
		h += y
		y = 0
	}
	if x < 0 {
		w += x
		x = 0
	}
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}

	gl.Enable(gl.GL_SCISSOR_TEST)
	gl.Scissor(x, y, w, h)
}

const (
	// VertexSize is the size of a vertex in bytes.
	VertexSize = 8 * 4 // 8 floats * 4 bytes

	defaultMaxVertices = 65536
	defaultMaxIndices  = 65536
)

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
	Shader     *shader.ShaderProgram
	Texture    uint32
	Uniforms   UniformSet
	uniformKey string
	Clip       *math.Rect
	clipKey    string
	Verts      []Vertex
	Indices    []uint32
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
	if bm == nil {
		return
	}
	bm.batches = bm.batches[:0]
	bm.current = nil
}

// AddQuad adds a quad to the batch
func (bm *BatchManager) AddQuad(shaderProg *shader.ShaderProgram, texture uint32, verts [4]Vertex) {
	bm.AddQuadWithState(shaderProg, texture, nil, nil, verts)
}

// AddQuadWithUniforms adds a quad to a batch with per-batch uniforms.
func (bm *BatchManager) AddQuadWithUniforms(shaderProg *shader.ShaderProgram, texture uint32, uniforms UniformSet, verts [4]Vertex) {
	bm.AddQuadWithState(shaderProg, texture, uniforms, nil, verts)
}

// AddQuadWithState adds a quad to a batch with per-batch render state.
func (bm *BatchManager) AddQuadWithState(shaderProg *shader.ShaderProgram, texture uint32, uniforms UniformSet, clip *math.Rect, verts [4]Vertex) {
	if bm == nil || shaderProg == nil {
		return
	}
	batch := bm.ensureBatch(shaderProg, texture, uniforms, clip, 4, 6)
	if batch == nil {
		return
	}

	// Add vertices
	offset := uint32(len(batch.Verts))
	batch.Verts = append(batch.Verts, verts[0], verts[1], verts[2], verts[3])
	batch.Indices = append(batch.Indices, offset, offset+1, offset+2, offset, offset+2, offset+3)
}

// AddTriangleWithState adds a triangle to a batch with per-batch render state.
func (bm *BatchManager) AddTriangleWithState(shaderProg *shader.ShaderProgram, texture uint32, uniforms UniformSet, clip *math.Rect, verts [3]Vertex) {
	if bm == nil || shaderProg == nil {
		return
	}
	batch := bm.ensureBatch(shaderProg, texture, uniforms, clip, 3, 3)
	if batch == nil {
		return
	}

	offset := uint32(len(batch.Verts))
	batch.Verts = append(batch.Verts, verts[0], verts[1], verts[2])
	batch.Indices = append(batch.Indices, offset, offset+1, offset+2)
}

func (bm *BatchManager) ensureBatch(shaderProg *shader.ShaderProgram, texture uint32, uniforms UniformSet, clip *math.Rect, addVerts, addIndices int) *Batch {
	if bm == nil || shaderProg == nil {
		return nil
	}
	uniformKey := uniforms.key()
	clipKey := rectKey(clip)

	// Check if we need a new batch
	if bm.current == nil ||
		bm.current.Shader != shaderProg ||
		bm.current.Texture != texture ||
		bm.current.uniformKey != uniformKey ||
		bm.current.clipKey != clipKey {

		// Save current batch
		if bm.current != nil && len(bm.current.Verts) > 0 {
			bm.batches = append(bm.batches, bm.current)
		}

		// Create new batch
		bm.current = &Batch{
			Shader:     shaderProg,
			Texture:    texture,
			Uniforms:   uniforms.clone(),
			uniformKey: uniformKey,
			Clip:       cloneRect(clip),
			clipKey:    clipKey,
		}
	}
	if bm.current != nil && bm.current.wouldOverflow(addVerts, addIndices, bm.maxVertices, bm.maxIndices) {
		if len(bm.current.Verts) > 0 {
			bm.batches = append(bm.batches, bm.current)
		}
		bm.current = &Batch{
			Shader:     shaderProg,
			Texture:    texture,
			Uniforms:   uniforms.clone(),
			uniformKey: uniformKey,
			Clip:       cloneRect(clip),
			clipKey:    clipKey,
		}
	}
	return bm.current
}

func (b *Batch) wouldOverflow(addVerts, addIndices, maxVerts, maxIndices int) bool {
	if b == nil {
		return false
	}
	if maxVerts > 0 && len(b.Verts)+addVerts > maxVerts {
		return true
	}
	if maxIndices > 0 && len(b.Indices)+addIndices > maxIndices {
		return true
	}
	return false
}

func batchGLReady() bool {
	return gl.BindVertexArray != nil &&
		gl.BindBuffer != nil &&
		gl.BufferSubData != nil &&
		gl.DrawElements != nil &&
		gl.Disable != nil
}

// Flush flushes all batches
func (bm *BatchManager) Flush(vao, vbo, ebo uint32, shaderMgr *shader.ShaderManager, projMatrix *[16]float32, viewportHeight float32) {
	if bm == nil || shaderMgr == nil || projMatrix == nil || vao == 0 || vbo == 0 || ebo == 0 || !batchGLReady() {
		return
	}
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
		if batch.Shader == nil || len(batch.Verts) == 0 || len(batch.Indices) == 0 {
			continue
		}

		// Use shader
		shaderMgr.UseShader(batch.Shader)
		shaderMgr.SetMat4("uProj", projMatrix)
		applyUniforms(shaderMgr, batch.Uniforms)
		applyClip(batch.Clip, viewportHeight)

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
	gl.Disable(gl.GL_SCISSOR_TEST)

	// Reset
	bm.Reset()
}

// Renderer is the main renderer
type Renderer struct {
	vao            uint32
	vbo            uint32
	ebo            uint32
	shaderMgr      *shader.ShaderManager
	batch          *BatchManager
	projMatrix     [16]float32
	width          float32
	height         float32
	clipStack      []math.Rect
	transformStack []math.Mat4
	initialized    bool
}

// NewRenderer creates a new renderer
func NewRenderer() *Renderer {
	return &Renderer{
		shaderMgr: shader.NewShaderManager(),
		batch:     NewBatchManager(defaultMaxVertices, defaultMaxIndices),
	}
}

// Init initializes the renderer
func (r *Renderer) Init() error {
	if r == nil {
		return fmt.Errorf("renderer is nil")
	}
	if r.initialized {
		return nil
	}
	if r.shaderMgr == nil {
		r.shaderMgr = shader.NewShaderManager()
	}
	if r.batch == nil {
		r.batch = NewBatchManager(defaultMaxVertices, defaultMaxIndices)
	}
	cleanupOnError := true
	defer func() {
		if cleanupOnError {
			r.Delete()
		}
	}()

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
	gl.BufferData(gl.GL_ARRAY_BUFFER, defaultMaxVertices*VertexSize, 0, gl.GL_DYNAMIC_DRAW)

	// Create EBO
	gl.GenBuffers(1, &r.ebo)
	gl.BindBuffer(gl.GL_ELEMENT_ARRAY_BUFFER, r.ebo)
	gl.BufferData(gl.GL_ELEMENT_ARRAY_BUFFER, defaultMaxIndices*4, 0, gl.GL_DYNAMIC_DRAW)

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

	r.initialized = true
	cleanupOnError = false
	return nil
}

// BeginFrame begins a new frame
func (r *Renderer) BeginFrame(width, height float32) {
	if r == nil || r.batch == nil {
		return
	}
	r.width = width
	r.height = height
	r.batch.Reset()
	r.clipStack = r.clipStack[:0]
	r.transformStack = r.transformStack[:0]
	if !r.initialized {
		return
	}

	// Set viewport
	gl.Viewport(0, 0, int32(width), int32(height))
	r.resetFrameState()

	// Set projection matrix
	r.projMatrix = math.OrthoMatrix(0, width, height, 0, -1, 1)

	// Clear
	gl.ClearColor(0.15, 0.15, 0.17, 1.0)
	gl.Clear(gl.GL_COLOR_BUFFER_BIT)

}

// EndFrame ends the frame
func (r *Renderer) EndFrame() {
	if r == nil {
		return
	}
	r.Flush()
}

// Flush flushes all pending draw calls
func (r *Renderer) Flush() {
	if r == nil || r.batch == nil || !r.initialized {
		return
	}
	r.batch.Flush(r.vao, r.vbo, r.ebo, r.shaderMgr, &r.projMatrix, r.height)
}

func (r *Renderer) resetFrameState() {
	if gl.ColorMask != nil {
		gl.ColorMask(true, true, true, true)
	}
	if gl.Disable != nil {
		gl.Disable(gl.GL_SCISSOR_TEST)
		gl.Disable(gl.GL_STENCIL_TEST)
		gl.Disable(gl.GL_DEPTH_TEST)
	}
	if gl.Enable != nil {
		gl.Enable(gl.GL_BLEND)
	}
	if gl.BlendFunc != nil {
		gl.BlendFunc(gl.GL_SRC_ALPHA, gl.GL_ONE_MINUS_SRC_ALPHA)
	}
}

func (r *Renderer) addQuad(shaderProg *shader.ShaderProgram, texture uint32, uniforms UniformSet, verts [4]Vertex) {
	if r == nil || r.batch == nil {
		return
	}
	verts = r.transformQuad(verts)

	var clip *math.Rect
	if len(r.clipStack) > 0 {
		top := r.clipStack[len(r.clipStack)-1]
		clip = &top
	}
	r.batch.AddQuadWithState(shaderProg, texture, uniforms, clip, verts)
}

func (r *Renderer) addTriangle(shaderProg *shader.ShaderProgram, texture uint32, uniforms UniformSet, verts [3]Vertex) {
	if r == nil || r.batch == nil {
		return
	}
	verts = r.transformTriangle(verts)

	var clip *math.Rect
	if len(r.clipStack) > 0 {
		top := r.clipStack[len(r.clipStack)-1]
		clip = &top
	}
	r.batch.AddTriangleWithState(shaderProg, texture, uniforms, clip, verts)
}

func (r *Renderer) transformQuad(verts [4]Vertex) [4]Vertex {
	if r == nil {
		return verts
	}
	if len(r.transformStack) == 0 {
		return verts
	}

	mat := r.transformStack[len(r.transformStack)-1]
	for i := range verts {
		verts[i].X, verts[i].Y = transformPoint(mat, verts[i].X, verts[i].Y)
	}
	return verts
}

func (r *Renderer) transformTriangle(verts [3]Vertex) [3]Vertex {
	if r == nil {
		return verts
	}
	if len(r.transformStack) == 0 {
		return verts
	}

	mat := r.transformStack[len(r.transformStack)-1]
	for i := range verts {
		verts[i].X, verts[i].Y = transformPoint(mat, verts[i].X, verts[i].Y)
	}
	return verts
}

func transformPoint(mat math.Mat4, x, y float32) (float32, float32) {
	tx := x*mat[0] + y*mat[4] + mat[12]
	ty := x*mat[1] + y*mat[5] + mat[13]
	return tx, ty
}

func transformRect(mat math.Mat4, rect math.Rect) math.Rect {
	x1, y1 := transformPoint(mat, rect.X, rect.Y)
	x2, y2 := transformPoint(mat, rect.X+rect.W, rect.Y)
	x3, y3 := transformPoint(mat, rect.X+rect.W, rect.Y+rect.H)
	x4, y4 := transformPoint(mat, rect.X, rect.Y+rect.H)

	minX := min4(x1, x2, x3, x4)
	maxX := max4(x1, x2, x3, x4)
	minY := min4(y1, y2, y3, y4)
	maxY := max4(y1, y2, y3, y4)
	return math.NewRect(minX, minY, maxX-minX, maxY-minY)
}

func min4(a, b, c, d float32) float32 {
	min := a
	if b < min {
		min = b
	}
	if c < min {
		min = c
	}
	if d < min {
		min = d
	}
	return min
}

func max4(a, b, c, d float32) float32 {
	max := a
	if b > max {
		max = b
	}
	if c > max {
		max = c
	}
	if d > max {
		max = d
	}
	return max
}

// PushTransform appends a transform to the current transform stack.
func (r *Renderer) PushTransform(mat math.Mat4) {
	if r == nil {
		return
	}
	r.Flush()
	if len(r.transformStack) > 0 {
		mat = r.transformStack[len(r.transformStack)-1].Multiply(mat)
	}
	r.transformStack = append(r.transformStack, mat)
}

// PopTransform restores the previous transform.
func (r *Renderer) PopTransform() {
	if r == nil {
		return
	}
	if len(r.transformStack) == 0 {
		return
	}
	r.Flush()
	r.transformStack = r.transformStack[:len(r.transformStack)-1]
}

// CurrentTransform returns the active transform matrix.
func (r *Renderer) CurrentTransform() (math.Mat4, bool) {
	if r == nil {
		return math.Mat4{}, false
	}
	if len(r.transformStack) == 0 {
		return math.Mat4{}, false
	}
	return r.transformStack[len(r.transformStack)-1], true
}

// PushClip intersects the provided clip rectangle with the current clip state.
func (r *Renderer) PushClip(rect math.Rect) {
	if r == nil {
		return
	}
	if len(r.transformStack) > 0 {
		rect = transformRect(r.transformStack[len(r.transformStack)-1], rect)
	}
	if len(r.clipStack) > 0 {
		rect = rect.Intersect(r.clipStack[len(r.clipStack)-1])
	}
	r.Flush()
	r.clipStack = append(r.clipStack, rect)
}

// PopClip restores the previous clip rectangle.
func (r *Renderer) PopClip() {
	if r == nil {
		return
	}
	if len(r.clipStack) == 0 {
		return
	}
	r.Flush()
	r.clipStack = r.clipStack[:len(r.clipStack)-1]
}

// CurrentClip returns the active clip rectangle.
func (r *Renderer) CurrentClip() (math.Rect, bool) {
	if r == nil {
		return math.Rect{}, false
	}
	if len(r.clipStack) == 0 {
		return math.Rect{}, false
	}
	return r.clipStack[len(r.clipStack)-1], true
}

// FillRect draws a filled rectangle
func (r *Renderer) FillRect(rect math.Rect, color math.Color) {
	if r == nil || r.shaderMgr == nil {
		return
	}
	shaderProg := r.shaderMgr.GetShader("color")
	uv := math.NewRect(0, 0, 1, 1)
	verts := QuadVertices(rect, uv, color)
	r.addQuad(shaderProg, 0, nil, verts)
}

// FillRoundRect draws a filled rounded rectangle
func (r *Renderer) FillRoundRect(rect math.Rect, radius float32, color math.Color) {
	if r == nil || r.shaderMgr == nil {
		return
	}
	shaderProg := r.shaderMgr.GetShader("rounded_rect")
	uniforms := UniformSet{
		"uRadius": FloatUniform(radius),
		"uSize":   Vec2Uniform(rect.W, rect.H),
	}

	uv := math.NewRect(0, 0, 1, 1)
	verts := QuadVertices(rect, uv, color)
	r.addQuad(shaderProg, 0, uniforms, verts)
}

// DrawTexture draws a textured rectangle
func (r *Renderer) DrawTexture(texture uint32, src, dst math.Rect, color math.Color) {
	if r == nil || r.shaderMgr == nil {
		return
	}
	shaderProg := r.shaderMgr.GetShader("texture")
	verts := QuadVertices(dst, src, color)
	r.addQuad(shaderProg, texture, nil, verts)
}

// Delete deletes all resources
func (r *Renderer) Delete() {
	if r == nil {
		return
	}
	if r.shaderMgr != nil {
		r.shaderMgr.Delete()
	}

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
	if r.batch != nil {
		r.batch.Reset()
	}
	r.clipStack = r.clipStack[:0]
	r.transformStack = r.transformStack[:0]
	r.initialized = false
}

// ShaderManager returns the shader manager
func (r *Renderer) ShaderManager() *shader.ShaderManager {
	if r == nil {
		return nil
	}
	return r.shaderMgr
}

// ProjectionMatrix returns the projection matrix
func (r *Renderer) ProjectionMatrix() *[16]float32 {
	if r == nil {
		return nil
	}
	return &r.projMatrix
}
