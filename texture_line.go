package graphics

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/gmath"
)

type TextureLine struct {
	BeginPos gmath.Pos
	EndPos   gmath.Pos

	// Shader is an shader that will be used during rendering of the texture.
	// Use NewShader to initialize this field.
	//
	// If nil, no shaders will be used.
	Shader *Shader

	colorScale       ColorScale
	ebitenColorScale ebiten.ColorScale

	texture         *ebiten.Image
	textureSubImage *ebiten.Image

	prevLength int

	visible  bool
	disposed bool
}

func NewTextureLine(begin, end gmath.Pos) *TextureLine {
	return &TextureLine{
		BeginPos:         begin,
		EndPos:           end,
		colorScale:       defaultColorScale,
		ebitenColorScale: defaultColorScale.ToEbitenColorScale(),
		visible:          true,
	}
}

func (l *TextureLine) BoundsRect() gmath.Rect {
	return lineBoundsRect(l.BeginPos, l.EndPos)
}

func (l *TextureLine) Dispose() {
	l.disposed = true
}

func (l *TextureLine) IsDisposed() bool {
	return l.disposed
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

func (l *TextureLine) SetTexture(texture *ebiten.Image) {
	l.texture = texture
}

func (l *TextureLine) GetTexture() *ebiten.Image {
	return l.texture
}

// GetColorScale is used to retrieve the current color scale value of the texture line.
// Use SetColorScale to change it.
func (l *TextureLine) GetColorScale() ColorScale {
	return l.colorScale
}

// SetColorScale assigns a new ColorScale to this texture line.
// Use GetColorScale to retrieve the current color scale.
func (l *TextureLine) SetColorScale(cs ColorScale) {
	if l.colorScale == cs {
		return
	}
	l.colorScale = cs
	l.ebitenColorScale = l.colorScale.ToEbitenColorScale()
}

// GetAlpha is a shorthand for GetColorScale().A expression.
// It's mostly provided for a symmetry with SetAlpha.
func (l *TextureLine) GetAlpha() float32 { return l.colorScale.A }

// SetAlpha is a convenient way to change the alpha value of the ColorScale.
func (l *TextureLine) SetAlpha(a float32) {
	if l.colorScale.A == a {
		return
	}
	l.colorScale.A = a
	l.ebitenColorScale = l.colorScale.ToEbitenColorScale()
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
	if !l.visible {
		return
	}
	if l.colorScale.A == 0 {
		return
	}

	pos1 := l.BeginPos.Resolve().Add(opts.Offset)
	pos2 := l.EndPos.Resolve().Add(opts.Offset)

	textureSize := l.texture.Bounds().Size()
	textureWidth := float64(textureSize.X)

	length := gmath.ClampMax(math.Round(pos1.DistanceTo(pos2)), textureWidth)
	if length < gmath.Epsilon {
		return
	}

	// In many cases, the line width should not change too often (?)
	// Doing a subimage is not free, therefore we try to minimize it.
	// See https://github.com/hajimehoshi/ebiten/issues/2902
	if ilength := int(length); ilength != l.prevLength {
		// Use integers here to get some rounding logic.
		// The lengths of 10.4 and 10.1 are not different enough to care.
		// And we need int length below for the bounds anyway.
		l.prevLength = ilength
		bounds := image.Rectangle{
			Max: image.Point{X: ilength, Y: textureSize.Y},
		}
		l.textureSubImage = l.texture.SubImage(bounds).(*ebiten.Image)
		fmt.Printf("%p: re-slice image (len=%d)\n", l, ilength)
	}

	textureHeight := float64(textureSize.Y)
	angle := pos1.AngleToPoint(pos2)
	origin := gmath.Vec{Y: textureHeight * 0.5}

	var drawOptions ebiten.DrawImageOptions
	drawOptions.ColorScale = l.ebitenColorScale

	drawOptions.GeoM.Translate(-origin.X, -origin.Y)
	drawOptions.GeoM.Rotate(float64(angle))
	drawOptions.GeoM.Translate(origin.X, origin.Y)

	drawOptions.GeoM.Translate(pos1.X, pos1.Y)

	if l.Shader == nil || !l.Shader.Enabled {
		dst.DrawImage(l.textureSubImage, &drawOptions)
		return
	}

	srcImageBounds := l.textureSubImage.Bounds()
	var options ebiten.DrawRectShaderOptions
	if opts.Blend != nil {
		options.Blend = *opts.Blend
	}
	options.GeoM = drawOptions.GeoM
	options.ColorScale = drawOptions.ColorScale
	options.Images[0] = l.textureSubImage
	options.Images[1] = l.Shader.Texture1
	options.Images[2] = l.Shader.Texture2
	options.Images[3] = l.Shader.Texture3
	options.Uniforms = l.Shader.shaderData
	dst.DrawRectShader(srcImageBounds.Dx(), srcImageBounds.Dy(), l.Shader.compiled, &options)
}
