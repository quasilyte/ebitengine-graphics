package graphics

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/gmath"
)

// Line is a simple 2-point line graphical primitive.
// Its color and width can be configured.
type Line struct {
	BeginPos gmath.Pos
	EndPos   gmath.Pos

	width float64

	colorScale       ColorScale
	ebitenColorScale ebiten.ColorScale

	visible  bool
	disposed bool
}

// NewLine returns a line that is drawn from begin pos to end pos.
// BeginPos and EndPos fields can be changed directly.
//
// By default, a line has these properties:
// * Visible=true
// * The ColorScale is {1, 1, 1, 1}
// * Width is 1
func NewLine(begin, end gmath.Pos) *Line {
	return &Line{
		BeginPos:         begin,
		EndPos:           end,
		colorScale:       defaultColorScale,
		ebitenColorScale: defaultColorScale.toEbitenColorScale(),
		width:            1,
		visible:          true,
	}
}

// BoundsRect returns a rectangle that fully contains the line.
//
// This is useful when trying to calculate whether this object is contained
// inside some area or not (like a camera view area).
func (l *Line) BoundsRect() gmath.Rect {
	return lineBoundsRect(l.BeginPos, l.EndPos)
}

// Dispose marks this line for deletion.
// After calling this method, IsDisposed will report true.
func (l *Line) Dispose() {
	l.disposed = true
}

// IsDisposed reports whether this line is marked for deletion.
// IsDisposed returns true only after Disposed was called on this line.
func (l *Line) IsDisposed() bool {
	return l.disposed
}

// IsVisible reports whether this line is visible.
// Use SetVisibility to change this flag value.
//
// When line is invisible (visible=false), it will not be rendered at all.
// This is an efficient way to temporarily hide a line.
func (l *Line) IsVisible() bool { return l.visible }

// SetVisibility changes the Visible flag value.
// It can be used to show or hide the line.
// Use IsVisible to get the current flag value.
func (l *Line) SetVisibility(visible bool) { l.visible = visible }

// GetWidth reports the current line width.
// Use SetWidth to change it.
func (l *Line) GetWidth() float64 {
	return l.width
}

// SetWidth changes the line width.
// Use GetWidth to retrieve the current line width value.
func (l *Line) SetWidth(w float64) {
	l.width = w
}

// GetColorScale is used to retrieve the current color scale value of the line.
// Use SetColorScale to change it.
func (l *Line) GetColorScale() ColorScale {
	return l.colorScale
}

// SetColorScale assigns a new ColorScale to this line.
// Use GetColorScale to retrieve the current color scale.
func (l *Line) SetColorScale(cs ColorScale) {
	if l.colorScale == cs {
		return
	}
	l.colorScale = cs
	l.ebitenColorScale = l.colorScale.toEbitenColorScale()
}

// GetAlpha is a shorthand for GetColorScale().A expression.
// It's mostly provided for a symmetry with SetAlpha.
func (l *Line) GetAlpha() float32 { return l.colorScale.A }

// SetAlpha is a convenient way to change the alpha value of the ColorScale.
func (l *Line) SetAlpha(a float32) {
	if l.colorScale.A == a {
		return
	}
	l.colorScale.A = a
	l.ebitenColorScale = l.colorScale.toEbitenColorScale()
}

// Draw renders the line onto the provided dst image.
//
// This method is a shorthand to DrawWithOptions(dst, {})
// which also implements the gscene.Graphics interface.
//
// See DrawWithOptions for more info.
func (l *Line) Draw(dst *ebiten.Image) {
	l.DrawWithOptions(dst, DrawOptions{})
}

// DrawWithOptions renders the line onto the provided dst image
// while also using the extra provided offset and other options.
//
// The offset is applied to both begin and end positions.
func (l *Line) DrawWithOptions(dst *ebiten.Image, opts DrawOptions) {
	if !l.visible {
		return
	}
	if l.colorScale.A == 0 || l.width <= gmath.Epsilon {
		return
	}

	pos1 := l.BeginPos.Resolve().Add(opts.Offset)
	pos2 := l.EndPos.Resolve().Add(opts.Offset)
	drawLine(dst, opts.Blend, pos1, pos2, l.width, l.ebitenColorScale)
}
