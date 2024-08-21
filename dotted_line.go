package graphics

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/ebitengine-graphics/internal/cache"
	"github.com/quasilyte/gmath"
)

// DottedLine is a 2-point line graphical primitive
// drawn using small segments.
type DottedLine struct {
	BeginPos gmath.Pos
	EndPos   gmath.Pos

	beginVec   gmath.Vec32
	endVec     gmath.Vec32
	shaderData map[string]any

	colorScale ColorScale

	visible  bool
	disposed bool
}

// NewDottedLine returns a line that is drawn from begin pos to end pos.
// BeginPos and EndPos fields can be changed directly.
//
// By default, a line has these properties:
// * Visible=true
// * The ColorScale is {1, 1, 1, 1}
// * DotRadius=1
// * DotSpacing=3
func NewDottedLine(begin, end gmath.Pos) *DottedLine {
	requireShaders()

	l := &DottedLine{
		BeginPos:   begin,
		EndPos:     end,
		colorScale: defaultColorScale,
		visible:    true,
	}

	l.shaderData = map[string]any{
		"DotRadius":  float32(1),
		"DotSpacing": float32(3),
		"Color":      l.colorScale.AsVec4(),
		"PointA":     l.beginVec.AsSlice(),
		"PointB":     l.endVec.AsSlice(),
	}

	return l
}

// BoundsRect returns a rectangle that fully contains the line.
//
// This is useful when trying to calculate whether this object is contained
// inside some area or not (like a camera view area).
func (l *DottedLine) BoundsRect() gmath.Rect {
	rr := l.GetDotRadius() + 1
	bounds := lineBoundsRect(l.BeginPos, l.EndPos)
	bounds.Min.X -= rr
	bounds.Min.Y -= rr
	bounds.Max.X += rr
	bounds.Max.Y += rr
	return bounds
}

// Dispose marks this line for deletion.
// After calling this method, IsDisposed will report true.
func (l *DottedLine) Dispose() {
	l.disposed = true
}

// IsDisposed reports whether this line is marked for deletion.
// IsDisposed returns true only after Disposed was called on this line.
func (l *DottedLine) IsDisposed() bool {
	return l.disposed
}

// IsVisible reports whether this line is visible.
// Use SetVisibility to change this flag value.
//
// When line is invisible (visible=false), it will not be rendered at all.
// This is an efficient way to temporarily hide a line.
func (l *DottedLine) IsVisible() bool { return l.visible }

// SetVisibility changes the Visible flag value.
// It can be used to show or hide the line.
// Use IsVisible to get the current flag value.
func (l *DottedLine) SetVisibility(visible bool) { l.visible = visible }

func (l *DottedLine) GetDotRadius() float64 {
	return float64(l.shaderData["DotRadius"].(float32))
}

func (l *DottedLine) SetDotRadius(r float64) {
	l.shaderData["DotRadius"] = float32(r)
}

func (l *DottedLine) GetDotSpacing() float64 {
	return float64(l.shaderData["DotSpacing"].(float32))
}

func (l *DottedLine) SetDotSpacing(spacing float64) {
	l.shaderData["DotSpacing"] = float32(spacing)
}

// GetColorScale is used to retrieve the current color scale value of the line.
// Use SetColorScale to change it.
func (l *DottedLine) GetColorScale() ColorScale {
	return l.colorScale.undoPremultiply()
}

// SetColorScale assigns a new ColorScale to this line.
// Use GetColorScale to retrieve the current color scale.
func (l *DottedLine) SetColorScale(cs ColorScale) {
	l.colorScale = cs.premultiplyAlpha()
}

// GetAlpha is a shorthand for GetColorScale().A expression.
// It's mostly provided for a symmetry with SetAlpha.
func (l *DottedLine) GetAlpha() float32 { return l.colorScale.A }

// SetAlpha is a convenient way to change the alpha value of the ColorScale.
func (l *DottedLine) SetAlpha(a float32) {
	l.colorScale.A = a
}

func (l *DottedLine) Draw(dst *ebiten.Image) {
	l.DrawWithOptions(dst, DrawOptions{})
}

func (l *DottedLine) DrawWithOptions(dst *ebiten.Image, opts DrawOptions) {
	if !l.visible {
		return
	}
	if l.colorScale.A == 0 {
		return
	}

	bounds := l.BoundsRect()
	width := bounds.Width()
	height := bounds.Height()
	pos := bounds.Min.Add(opts.Offset)

	l.beginVec = l.BeginPos.Resolve().Add(opts.Offset).AsVec32()
	l.endVec = l.EndPos.Resolve().Add(opts.Offset).AsVec32()

	var drawOptions ebiten.DrawRectShaderOptions
	drawOptions.Uniforms = l.shaderData
	if opts.Blend != nil {
		drawOptions.Blend = *opts.Blend
	}
	drawOptions.GeoM.Translate(pos.X, pos.Y)
	dst.DrawRectShader(int(width), int(height), cache.Global.DottedLineShader, &drawOptions)
}
