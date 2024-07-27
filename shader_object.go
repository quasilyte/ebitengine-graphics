package graphics

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/ebitengine-graphics/internal/cache"
	"github.com/quasilyte/gmath"
)

// ShaderObject implements shader-based rendering.
//
// Instead of using a dummy sprite/image to draw something
// with a shader, a ShaderObject should be used instead.
//
// The ShaderObject creates a quad shape out of vertices
// based on the provided position and sizes (w, h).
// The associated shader is then executed over that
// quad and the result of that rendering is
// drawn right into the Draw dst image.
//
// ShaderObject implements gscene Graphics interface.
type ShaderObject struct {
	Shader *Shader

	Pos gmath.Pos

	visible  bool
	disposed bool

	width  uint16
	height uint16
}

func NewShaderObject() *ShaderObject {
	return &ShaderObject{
		visible: true,
	}
}

func (o *ShaderObject) IsVisible() bool {
	return o.visible
}

func (o *ShaderObject) SetVisibility(visible bool) {
	o.visible = visible
}

func (o *ShaderObject) Dispose() {
	o.disposed = true
}

func (o *ShaderObject) IsDisposed() bool {
	return o.disposed
}

func (o *ShaderObject) GetWidth() int {
	return int(o.width)
}

func (o *ShaderObject) SetWidth(w int) {
	uw := uint16(w)
	o.width = uw
}

func (o *ShaderObject) GetHeight() int {
	return int(o.height)
}

func (o *ShaderObject) SetHeight(h int) {
	uh := uint16(h)
	o.height = uh
}

func (o *ShaderObject) Draw(dst *ebiten.Image) {
	o.DrawWithOptions(dst, DrawOptions{})
}

func (o *ShaderObject) DrawWithOptions(dst *ebiten.Image, opts DrawOptions) {
	if !o.IsVisible() || o.Shader == nil || !o.Shader.Enabled {
		return
	}
	if int(o.width)+int(o.height) == 0 {
		return
	}

	// Use pre-allocated slices.
	vertices := cache.Global.ScratchVertices[:0]
	indices := cache.Global.ScratchIndices[:0]
	defer func() {
		cache.Global.ScratchVertices = vertices[:0]
		cache.Global.ScratchIndices = indices[:0]
	}()

	pos := opts.Offset.Add(o.Pos.Resolve())
	x := float32(pos.X)
	y := float32(pos.Y)
	w := float32(o.width)
	h := float32(o.height)

	// It's up to the user to override the default addressing mode.
	// Right now, if src size is different from draw rect,
	// the src texture will be scaled (quite weirdly).
	// There might be a flag in the future to allow different defaults later (how?)
	srcWidth := w
	srcHeight := h

	// Maybe allow the user to provide a custom color scale?
	clr := defaultColorScale

	vertices = append(vertices,
		ebiten.Vertex{DstX: x, DstY: y, SrcX: 0, SrcY: 0, ColorR: clr.R, ColorG: clr.G, ColorB: clr.B, ColorA: clr.A},
		ebiten.Vertex{DstX: x + w, DstY: y, SrcX: srcWidth, SrcY: 0, ColorR: clr.R, ColorG: clr.G, ColorB: clr.B, ColorA: clr.A},
		ebiten.Vertex{DstX: x, DstY: y + h, SrcX: 0, SrcY: srcHeight, ColorR: clr.R, ColorG: clr.G, ColorB: clr.B, ColorA: clr.A},
		ebiten.Vertex{DstX: x + w, DstY: y + h, SrcX: srcWidth, SrcY: srcHeight, ColorR: clr.R, ColorG: clr.G, ColorB: clr.B, ColorA: clr.A},
	)
	indices = append(indices, 0, 1, 2, 1, 2, 3)

	var drawOptions ebiten.DrawTrianglesShaderOptions
	if opts.Blend != nil {
		drawOptions.Blend = *opts.Blend
	}
	drawOptions.Uniforms = o.Shader.shaderData
	drawOptions.Images[0] = o.Shader.Texture1 // Unsure if it's a good idea
	drawOptions.Images[1] = o.Shader.Texture1
	drawOptions.Images[2] = o.Shader.Texture2
	drawOptions.Images[3] = o.Shader.Texture3
	dst.DrawTrianglesShader(vertices, indices, o.Shader.compiled, &drawOptions)
}
