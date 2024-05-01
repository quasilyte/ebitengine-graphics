package graphics

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/gmath"
)

var borderBoxIndices = []uint16{0, 2, 4, 2, 4, 6, 1, 3, 5, 3, 5, 7}

// Rect is a rectangle graphical primitive.
//
// Depending on the configuration, it's one of these:
//   - Fill-only rectangle
//   - Outline-only rectangle
//   - Fill+outline rectangle
//
// If you need a "texture rect", use a sprite instead.
type Rect struct {
	// Pos is a rect location binder.
	// See Pos documentation to learn how it works.
	//
	// When rendering an image, Pos.Resolve() will be used
	// to calculate the final position.
	Pos gmath.Pos

	width  float64
	height float64

	outlineColorScale ColorScale
	fillColorScale    ColorScale

	outlineWidth float64

	outlineVertices *[8]ebiten.Vertex

	centered bool
	visible  bool
	disposed bool
}

// NewRect returns a rectangle of the specified size.
// Use SetSize if you need to resize it afterwards.
//
// By default, a rect has these properties:
// * Centered=true
// * Visible=true
// * The FillColorScale is {1, 1, 1, 1}
// * The OutlineColorScale is {0, 0, 0, 0} (invisible)
// * OutlineWidth is 1 (but the default outline color is invisible)
func NewRect(_ *Cache, width, height float64) *Rect {
	// Cache is not used yet, but it's our way to keep the API stable
	// while keeping the optimization opportunities.
	// It's also consistent with other constructor functions.
	return &Rect{
		visible:           true,
		centered:          true,
		fillColorScale:    defaultColorScale,
		outlineColorScale: transparentColor,
		outlineWidth:      1,
		width:             width,
		height:            height,
	}
}

// BoundsRect returns the properly positioned rectangle.
//
// This is useful when trying to calculate whether this object is contained
// inside some area or not (like a camera view area).
func (rect *Rect) BoundsRect() gmath.Rect {
	pos := rect.Pos.Resolve()
	if rect.centered {
		offset := gmath.Vec{X: rect.width * 0.5, Y: rect.height * 0.5}
		return gmath.Rect{
			Min: pos.Sub(offset),
			Max: pos.Add(offset),
		}
	}
	return gmath.Rect{
		Min: pos,
		Max: pos.Add(gmath.Vec{X: rect.width, Y: rect.height}),
	}
}

// Dispose marks this rect for deletion.
// After calling this method, IsDisposed will report true.
func (rect *Rect) Dispose() {
	rect.disposed = true
}

// IsDisposed reports whether this rect is marked for deletion.
// IsDisposed returns true only after Disposed was called on this rect.
func (rect *Rect) IsDisposed() bool {
	return rect.disposed
}

// IsCentered reports whether Centered flag is set.
// Use SetCentered to change this flag value.
//
// When rect is centered, its image origin will be {w/2, h/2} during rendering.
func (rect *Rect) IsCentered() bool { return rect.centered }

// SetCentered changes the Centered flag value.
// Use IsCentered to get the current flag value.
func (rect *Rect) SetCentered(centered bool) { rect.centered = centered }

// IsVisible reports whether this rect is visible.
// Use SetVisibility to change this flag value.
//
// When rect is invisible (visible=false), it will not be rendered at all.
// This is an efficient way to temporarily hide a rect.
func (rect *Rect) IsVisible() bool { return rect.visible }

// SetVisibility changes the Visible flag value.
// It can be used to show or hide the rect.
// Use IsVisible to get the current flag value.
func (rect *Rect) SetVisibility(visible bool) { rect.visible = visible }

func (rect *Rect) GetWidth() float64 {
	return rect.width
}

func (rect *Rect) SetWidth(w float64) {
	rect.width = w
}

func (rect *Rect) GetHeight() float64 {
	return rect.height
}

func (rect *Rect) SetHeight(h float64) {
	rect.height = h
}

// GetOutlineWidth reports the current outline width.
// Use SetOutlineWidth to change it.
func (rect *Rect) GetOutlineWidth() float64 {
	return rect.outlineWidth
}

// SetOutlineWidth changes the rect outline width.
// Use GetOutlineWidth to retrieve the current outline width value.
func (rect *Rect) SetOutlineWidth(w float64) {
	rect.outlineWidth = w
}

// GetFillColorScale is used to retrieve the current fill color scale value of the rect.
// Use SetFillColorScale to change it.
func (rect *Rect) GetFillColorScale() ColorScale {
	return rect.fillColorScale
}

// SetFillColorScale assigns a new fill ColorScale to this rect.
// Use GetFillColorScale to retrieve the current color scale.
func (rect *Rect) SetFillColorScale(cs ColorScale) {
	rect.fillColorScale = cs
}

// GetOutlineColorScale is used to retrieve the current outline color scale value of the rect.
// Use SetOutlineColorScale to change it.
func (rect *Rect) GetOutlineColorScale() ColorScale {
	return rect.outlineColorScale
}

// SetOutlineColorScale assigns a new outline ColorScale to this rect.
// Use GetOutlineColorScale to retrieve the current color scale.
func (rect *Rect) SetOutlineColorScale(cs ColorScale) {
	rect.outlineColorScale = cs
}

// Draw renders the rect onto the provided dst image.
//
// This method is a shorthand to DrawWithOptions(dst, {})
// which also implements the gscene.Graphics interface.
//
// See DrawWithOptions for more info.
func (rect *Rect) Draw(dst *ebiten.Image) {
	rect.DrawWithOptions(dst, DrawOptions{})
}

// DrawWithOptions renders the rect onto the provided dst image
// while also using the extra provided offset.
func (rect *Rect) DrawWithOptions(dst *ebiten.Image, opts DrawOptions) {
	if !rect.visible {
		return
	}
	if rect.outlineColorScale.A == 0 && rect.fillColorScale.A == 0 {
		return
	}

	// TODO: compare the peformance of this method with vector package.
	// TODO: implement the rotation.
	// TODO: implement the scaling.
	// TODO: maybe add a special case for opaque rectangles.

	finalOffset := rect.calculateFinalOffset(opts.Offset)

	if rect.outlineColorScale.A == 0 || rect.outlineWidth < 1 {
		// Fill-only mode.
		var drawOptions ebiten.DrawImageOptions
		drawOptions.GeoM = rect.calculateGeom(rect.width, rect.height, finalOffset)
		drawOptions.ColorScale = rect.fillColorScale.toEbitenColorScale()
		dst.DrawImage(whitePixel, &drawOptions)
		return
	}

	if rect.fillColorScale.A == 0 && rect.outlineWidth >= 1 {
		// Outline-only mode.
		rect.drawOutline(dst, finalOffset)
		return
	}

	rect.drawOutline(dst, finalOffset)

	var drawOptions ebiten.DrawImageOptions
	drawOptions.GeoM.Scale(rect.width-rect.outlineWidth*2, rect.height-rect.outlineWidth*2)
	drawOptions.GeoM.Translate(rect.outlineWidth+finalOffset.X, rect.outlineWidth+finalOffset.Y)
	drawOptions.ColorScale = rect.fillColorScale.toEbitenColorScale()
	dst.DrawImage(whitePixel, &drawOptions)
}

func (rect *Rect) drawOutline(dst *ebiten.Image, offset gmath.Vec) {
	if rect.outlineVertices == nil {
		// Allocate these vertices lazily when we need them and then re-use them.
		rect.outlineVertices = new([8]ebiten.Vertex)
	}

	borderWidth := float32(rect.outlineWidth)
	x := float32(offset.X)
	y := float32(offset.Y)
	r := rect.outlineColorScale.R
	g := rect.outlineColorScale.G
	b := rect.outlineColorScale.B
	a := rect.outlineColorScale.A
	width := float32(rect.width)
	height := float32(rect.height)

	rect.outlineVertices[0] = ebiten.Vertex{
		DstX:   x,
		DstY:   y,
		SrcX:   0,
		SrcY:   0,
		ColorR: r,
		ColorG: g,
		ColorB: b,
		ColorA: a,
	}
	rect.outlineVertices[1] = ebiten.Vertex{
		DstX: x + borderWidth,
		DstY: y + borderWidth,
		SrcX: 0,
		SrcY: 0,
	}
	rect.outlineVertices[2] = ebiten.Vertex{
		DstX:   x + width,
		DstY:   y,
		SrcX:   1,
		SrcY:   0,
		ColorR: r,
		ColorG: g,
		ColorB: b,
		ColorA: a,
	}
	rect.outlineVertices[3] = ebiten.Vertex{
		DstX: x + width - borderWidth,
		DstY: y + borderWidth,
		SrcX: 1,
		SrcY: 0,
	}
	rect.outlineVertices[4] = ebiten.Vertex{
		DstX:   x,
		DstY:   y + height,
		SrcX:   0,
		SrcY:   1,
		ColorR: r,
		ColorG: g,
		ColorB: b,
		ColorA: a,
	}
	rect.outlineVertices[5] = ebiten.Vertex{
		DstX: x + borderWidth,
		DstY: y + height - borderWidth,
		SrcX: 0,
		SrcY: 1,
	}
	rect.outlineVertices[6] = ebiten.Vertex{
		DstX:   x + width,
		DstY:   y + height,
		SrcX:   1,
		SrcY:   1,
		ColorR: r,
		ColorG: g,
		ColorB: b,
		ColorA: a,
	}
	rect.outlineVertices[7] = ebiten.Vertex{
		DstX: x + width - borderWidth,
		DstY: y + height - borderWidth,
		SrcX: 1,
		SrcY: 1,
	}

	options := ebiten.DrawTrianglesOptions{
		FillRule: ebiten.EvenOdd,
	}
	dst.DrawTriangles(rect.outlineVertices[:], borderBoxIndices, whitePixel, &options)
}

func (rect *Rect) calculateFinalOffset(offset gmath.Vec) gmath.Vec {
	var origin gmath.Vec
	if rect.centered {
		origin = gmath.Vec{X: rect.width * 0.5, Y: rect.height * 0.5}
	}
	origin = origin.Sub(rect.Pos.Offset)

	var pos gmath.Vec
	if rect.Pos.Base != nil {
		pos = rect.Pos.Base.Sub(origin)
	} else if !origin.IsZero() {
		pos = gmath.Vec{X: -origin.X, Y: -origin.Y}
	}
	return pos.Add(offset)
}

func (rect *Rect) calculateGeom(w, h float64, pos gmath.Vec) ebiten.GeoM {
	var geom ebiten.GeoM
	geom.Scale(w, h)
	geom.Translate(pos.X, pos.Y)
	return geom
}
