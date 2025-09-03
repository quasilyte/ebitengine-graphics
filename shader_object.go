package graphics

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/ebitengine-graphics/internal/cache"
	"github.com/quasilyte/ebitengine-graphics/internal/xmath"
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

	Rotation *gmath.Rad

	colorScale       ColorScale
	ebitenColorScale ebiten.ColorScale

	centered bool
	visible  bool
	disposed bool

	width  uint16
	height uint16
}

func NewShaderObject() *ShaderObject {
	return &ShaderObject{
		visible:          true,
		centered:         true,
		colorScale:       defaultColorScale,
		ebitenColorScale: defaultColorScale.ToEbitenColorScale(),
	}
}

func (o *ShaderObject) IsVisible() bool {
	return o.visible
}

func (o *ShaderObject) SetVisibility(visible bool) {
	o.visible = visible
}

func (o *ShaderObject) SetCentered(centered bool) {
	o.centered = centered
}

func (o *ShaderObject) GetColorScale() ColorScale {
	return o.colorScale
}

func (o *ShaderObject) SetColorScale(cs ColorScale) {
	if o.colorScale == cs {
		return
	}
	o.colorScale = cs
	o.ebitenColorScale = o.colorScale.ToEbitenColorScale()
}

func (o *ShaderObject) GetAlpha() float32 { return o.colorScale.A }

func (o *ShaderObject) SetAlpha(a float32) {
	if o.colorScale.A == a {
		return
	}
	o.colorScale.A = a
	o.ebitenColorScale = o.colorScale.ToEbitenColorScale()
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
	w := float32(o.width)
	h := float32(o.height)

	// It's up to the user to override the default addressing mode.
	// Right now, if src size is different from draw rect,
	// the src texture will be scaled (quite weirdly).
	// There might be a flag in the future to allow different defaults later (how?)
	srcWidth := w
	srcHeight := h

	// Maybe allow the user to provide a custom color scale?
	clrR := o.ebitenColorScale.R()
	clrG := o.ebitenColorScale.G()
	clrB := o.ebitenColorScale.B()
	clrA := o.ebitenColorScale.A()

	angle := gmath.Rad(0)
	if o.Rotation != nil {
		angle = *o.Rotation
	}

	if angle == 0 {
		x := float32(pos.X)
		y := float32(pos.Y)
		if o.centered {
			x -= w * 0.5
			y -= y * 0.5
		}
		vertices = append(vertices,
			ebiten.Vertex{DstX: x, DstY: y, SrcX: 0, SrcY: 0, ColorR: clrR, ColorG: clrG, ColorB: clrB, ColorA: clrA},
			ebiten.Vertex{DstX: x + w, DstY: y, SrcX: srcWidth, SrcY: 0, ColorR: clrR, ColorG: clrG, ColorB: clrB, ColorA: clrA},
			ebiten.Vertex{DstX: x, DstY: y + h, SrcX: 0, SrcY: srcHeight, ColorR: clrR, ColorG: clrG, ColorB: clrB, ColorA: clrA},
			ebiten.Vertex{DstX: x + w, DstY: y + h, SrcX: srcWidth, SrcY: srcHeight, ColorR: clrR, ColorG: clrG, ColorB: clrB, ColorA: clrA},
		)
	} else {
		var geom xmath.Geom32
		halfWidth := w * 0.5
		halfHeight := h * 0.5
		if o.centered {
			geom.Translate(-halfWidth, -halfHeight)
		}
		geom.Rotate(float64(angle))
		geom.Translate(float32(pos.X), float32(pos.Y))
		x := geom.Tx
		y := geom.Ty
		vertices = append(vertices,
			ebiten.Vertex{DstX: x, DstY: y, SrcX: 0, SrcY: 0, ColorR: clrR, ColorG: clrG, ColorB: clrB, ColorA: clrA},
			ebiten.Vertex{DstX: (geom.A1+1)*w + x, DstY: geom.C*w + y, SrcX: srcWidth, SrcY: 0, ColorR: clrR, ColorG: clrG, ColorB: clrB, ColorA: clrA},
			ebiten.Vertex{DstX: geom.B*h + x, DstY: (geom.D1+1)*h + y, SrcX: 0, SrcY: srcHeight, ColorR: clrR, ColorG: clrG, ColorB: clrB, ColorA: clrA},
			ebiten.Vertex{DstX: geom.ApplyX(w, h), DstY: geom.ApplyY(w, h), SrcX: srcWidth, SrcY: srcHeight, ColorR: clrR, ColorG: clrG, ColorB: clrB, ColorA: clrA},
		)
	}

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
