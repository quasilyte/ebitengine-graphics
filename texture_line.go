package graphics

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/ebitengine-graphics/internal/cache"
	"github.com/quasilyte/ebitengine-graphics/internal/xmath"
	"github.com/quasilyte/gmath"
)

type TextureLine struct {
	BeginPos gmath.Pos
	EndPos   gmath.Pos

	// Shader is an shader that will be used during rendering of the line.
	// Use NewShader to initialize this field.
	//
	// If nil, no shaders will be used.
	Shader *Shader

	colorScale ColorScale

	texture    *ebiten.Image
	texturePad float64

	visible  bool
	disposed bool
}

func NewTextureLine(begin, end gmath.Pos) *TextureLine {
	l := &TextureLine{
		BeginPos:   begin,
		EndPos:     end,
		visible:    true,
		colorScale: defaultColorScale,
	}

	return l
}

func (l *TextureLine) BoundsRect() gmath.Rect {
	bounds := lineBoundsRect(l.BeginPos, l.EndPos)
	pad := gmath.Vec{X: l.texturePad, Y: l.texturePad}
	return gmath.Rect{
		Min: bounds.Min.Sub(pad),
		Max: bounds.Max.Add(pad),
	}
}

func (l *TextureLine) Dispose() {
	l.disposed = true
}

func (l *TextureLine) IsDisposed() bool {
	return l.disposed
}

// GetColorScale is used to retrieve the current color scale value of texture line.
// Use SetColorScale to change it.
func (l *TextureLine) GetColorScale() ColorScale {
	return l.colorScale
}

// SetColorScale assigns a new ColorScale to this texture line.
// Use GetColorScale to retrieve the current color scale.
func (l *TextureLine) SetColorScale(cs ColorScale) {
	l.colorScale = cs
}

// GetAlpha is a shorthand for GetColorScale().A expression.
// It's mostly provided for a symmetry with SetAlpha.
func (l *TextureLine) GetAlpha() float32 { return l.colorScale.A }

// SetAlpha is a convenient way to change the alpha value of the ColorScale.
func (l *TextureLine) SetAlpha(a float32) {
	l.colorScale.A = a
}

// IsVisible reports whether this texture line is visible.
// Use SetVisibility to change this flag value.
//
// When texture line is invisible (visible=false), it will not be rendered at all.
// This is an efficient way to temporarily hide a texture line.
func (l *TextureLine) IsVisible() bool { return l.visible }

// SetVisibility changes the Visible flag value.
// It can be used to show or hide the texture line.
// Use IsVisible to get the current flag value.
func (l *TextureLine) SetVisibility(visible bool) { l.visible = visible }

// SetTexture assigns the repeatable line texture.
//
// This texture should loop well if the line's length can be higher
// than the texture's width.
func (l *TextureLine) SetTexture(texture *ebiten.Image) {
	l.texture = texture

	bounds := texture.Bounds()
	l.texturePad = 2 + math.Ceil((0.5 * float64(bounds.Dy())))
}

func (l *TextureLine) GetTexture() *ebiten.Image {
	return l.texture
}

// Draw renders the texture line onto the provided dst image.
//
// This method is a shorthand to DrawWithOptions(dst, {})
// which also implements the gscene.Graphics interface.
//
// See DrawWithOptions for more info.
func (l *TextureLine) Draw(dst *ebiten.Image) {
	l.DrawWithOptions(dst, DrawOptions{})
}

// DrawWithOptions renders the texture line onto the provided dst image
// while also using the extra provided offset and other options.
//
// The offset is applied to both begin and end positions.
func (l *TextureLine) DrawWithOptions(dst *ebiten.Image, opts DrawOptions) {
	if !l.visible || l.colorScale.A == 0 {
		return
	}

	beginVec := l.BeginPos.Resolve().Add(opts.Offset).AsVec32()
	endVec := l.EndPos.Resolve().Add(opts.Offset).AsVec32()

	vertices := cache.Global.ScratchVertices[:0]
	indices := cache.Global.ScratchIndices[:0]
	defer func() {
		cache.Global.ScratchVertices = vertices[:0]
		cache.Global.ScratchIndices = indices[:0]
	}()

	textureWidth := float32(l.texture.Bounds().Dx())
	textureHeight := float32(l.texture.Bounds().Dy())

	halfHeight := textureHeight * 0.5

	clr := l.colorScale.premultiplyAlpha()

	step := beginVec.DirectionTo(endVec).Mulf(textureWidth)
	numSteps := int(beginVec.DistanceTo(endVec)/textureWidth) + 1
	lastStep := numSteps - 1

	angle := float64(step.Angle())

	// Since rotation is identical for every quad, precompute
	// the rotation here and copy this data for every quad.
	// Rotation involves operations like Sincos and several
	// multiplications, so copying is faster.
	var geomBase xmath.Geom32
	geomBase.Translate(0, -halfHeight)
	geomBase.Rotate(angle)

	currentPos := beginVec
	idx := uint16(0)
	for i := 0; i < numSteps; i++ {
		geom := geomBase
		geom.Translate(currentPos.X, currentPos.Y)

		x := geom.Tx
		y := geom.Ty
		w := textureWidth
		h := textureHeight
		if i == lastStep {
			w = currentPos.DistanceTo(endVec)
		}

		vertices = append(vertices,
			ebiten.Vertex{DstX: x, DstY: y, SrcX: 0, SrcY: 0, ColorR: clr.R, ColorG: clr.G, ColorB: clr.B, ColorA: clr.A},
			ebiten.Vertex{DstX: (geom.A1+1)*w + x, DstY: geom.C*w + y, SrcX: w, SrcY: 0, ColorR: clr.R, ColorG: clr.G, ColorB: clr.B, ColorA: clr.A},
			ebiten.Vertex{DstX: geom.B*h + x, DstY: (geom.D1+1)*h + y, SrcX: 0, SrcY: h, ColorR: clr.R, ColorG: clr.G, ColorB: clr.B, ColorA: clr.A},
			ebiten.Vertex{DstX: geom.ApplyX(w, h), DstY: geom.ApplyY(w, h), SrcX: w, SrcY: h, ColorR: clr.R, ColorG: clr.G, ColorB: clr.B, ColorA: clr.A},
		)
		indices = append(indices,
			idx+0, idx+1, idx+2,
			idx+1, idx+2, idx+3,
		)

		currentPos = currentPos.Add(step)
		idx += 4
	}

	if l.Shader == nil || !l.Shader.Enabled {
		var drawOptions ebiten.DrawTrianglesOptions
		if opts.Blend != nil {
			drawOptions.Blend = *opts.Blend
		}
		dst.DrawTriangles(vertices, indices, l.texture, &drawOptions)
		return
	}

	var drawOptions ebiten.DrawTrianglesShaderOptions
	if opts.Blend != nil {
		drawOptions.Blend = *opts.Blend
	}
	drawOptions.Images[0] = l.texture
	drawOptions.Images[1] = l.Shader.Texture1
	drawOptions.Images[2] = l.Shader.Texture2
	drawOptions.Images[3] = l.Shader.Texture3
	drawOptions.Uniforms = l.Shader.shaderData
	dst.DrawTrianglesShader(vertices, indices, l.Shader.compiled, &drawOptions)
}
