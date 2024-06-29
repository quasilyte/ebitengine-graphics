package graphics

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/ebitengine-graphics/internal/cache"
	"github.com/quasilyte/gmath"
)

type Circle struct {
	Pos gmath.Pos

	cachedRotation gmath.Rad
	Rotation       *gmath.Rad

	outlineColorScale ColorScale
	fillColorScale    ColorScale

	dashLength float32
	dashGap    float32
	radius     float32
	shaderData map[string]any

	centered bool
	visible  bool
	disposed bool
}

// NewCircle returns a circle of the specified radius.
// Use SetRadius if you need to resize it afterwards.
//
// By default, a circle has these properties:
// * Centered=true
// * Visible=true
// * The FillColorScale is {0, 0, 0, 0} (invisible)
// * FillOffset is 0 (no gaps)
// * The OutlineColorScale is {1, 1, 1, 1}
// * OutlineWidth is 1
//
// The outline can have a dashed style.
//
// You need to call [CompileShaders] before using circles.
func NewCircle(r float64) *Circle {
	c := &Circle{
		centered:          true,
		visible:           true,
		radius:            float32(r),
		outlineColorScale: defaultColorScale,
		fillColorScale:    transparentColor,
	}

	c.shaderData = map[string]any{
		"Radius":       float32(r),
		"OutlineWidth": float32(1),
		"FillColor":    c.fillColorScale.AsVec4(),
		"FillOffset":   float32(0),
		"OutlineColor": c.outlineColorScale.AsVec4(),
		"DashLength":   float32(0),
		"DashGap":      float32(0),
		"Rotation":     float32(0),
	}

	requireShaders()

	return c
}

// BoundsRect returns the properly positioned circle containing rectangle.
//
// This is useful when trying to calculate whether this object is contained
// inside some area or not (like a camera view area).
func (c *Circle) BoundsRect() gmath.Rect {
	pos := c.Pos.Resolve()
	if c.centered {
		offset := gmath.Vec{X: float64(c.radius), Y: float64(c.radius)}
		return gmath.Rect{
			Min: pos.Sub(offset),
			Max: pos.Add(offset),
		}
	}
	return gmath.Rect{
		Min: pos,
		Max: pos.Add(gmath.Vec{X: 2 * float64(c.radius), Y: 2 * float64(c.radius)}),
	}
}

// Dispose marks this circle for deletion.
// After calling this method, IsDisposed will report true.
func (c *Circle) Dispose() {
	c.disposed = true
}

// IsDisposed reports whether this circle is marked for deletion.
// IsDisposed returns true only after Disposed was called on this circle.
func (c *Circle) IsDisposed() bool {
	return c.disposed
}

// IsCentered reports whether Centered flag is set.
// Use SetCentered to change this flag value.
func (c *Circle) IsCentered() bool { return c.centered }

// SetCentered changes the Centered flag value.
// Use IsCentered to get the current flag value.
func (c *Circle) SetCentered(centered bool) { c.centered = centered }

// IsVisible reports whether this circle is visible.
// Use SetVisibility to change this flag value.
//
// When circle is invisible (visible=false), it will not be rendered at all.
// This is an efficient way to temporarily hide a circle.
func (c *Circle) IsVisible() bool { return c.visible }

// SetVisibility changes the Visible flag value.
// It can be used to show or hide the circle.
// Use IsVisible to get the current flag value.
func (c *Circle) SetVisibility(visible bool) { c.visible = visible }

func (c *Circle) GetRadius() float64 {
	return float64(c.radius)
}

func (c *Circle) SetRadius(r float64) {
	c.radius = float32(r)
	c.shaderData["Radius"] = float32(r)
}

func (c *Circle) GetOutlineDash() (length, gap float64) {
	return float64(c.dashLength), float64(c.dashGap)
}

func (c *Circle) SetOutlineDash(length, gap float64) {
	c.dashLength = float32(length)
	c.dashGap = float32(gap)
	c.shaderData["DashLength"] = c.dashLength
	c.shaderData["DashGap"] = c.dashGap
}

func (c *Circle) GetOutlineWidth() float64 {
	return float64(c.shaderData["OutlineWidth"].(float32))
}

func (c *Circle) SetOutlineWidth(w float64) {
	c.shaderData["OutlineWidth"] = float32(w)
}

func (c *Circle) GetFillOffset() float64 {
	return float64(c.shaderData["FillOffset"].(float32))
}

func (c *Circle) SetFillOffset(offset float64) {
	c.shaderData["FillOffset"] = float32(offset)
}

// GetFillColorScale is used to retrieve the current fill color scale value of the circle.
// Use SetFillColorScale to change it.
func (c *Circle) GetFillColorScale() ColorScale {
	return c.fillColorScale.undoPremultiply()
}

// SetFillColorScale assigns a new fill ColorScale to this circle.
// Use GetFillColorScale to retrieve the current color scale.
func (c *Circle) SetFillColorScale(cs ColorScale) {
	c.fillColorScale = cs.premultiplyAlpha()
}

// GetOutlineColorScale is used to retrieve the current outline color scale value of the circle.
// Use SetOutlineColorScale to change it.
func (c *Circle) GetOutlineColorScale() ColorScale {
	return c.outlineColorScale.undoPremultiply()
}

// SetOutlineColorScale assigns a new outline ColorScale to this circle.
// Use GetOutlineColorScale to retrieve the current color scale.
func (c *Circle) SetOutlineColorScale(cs ColorScale) {
	c.outlineColorScale = cs.premultiplyAlpha()
}

// Draw renders the circle onto the provided dst image.
//
// This method is a shorthand to DrawWithOptions(dst, {})
// which also implements the gscene.Graphics interface.
//
// See DrawWithOptions for more info.
func (c *Circle) Draw(dst *ebiten.Image) {
	c.DrawWithOptions(dst, DrawOptions{})
}

func (c *Circle) DrawWithOptions(dst *ebiten.Image, opts DrawOptions) {
	if !c.visible {
		return
	}
	if c.outlineColorScale.A == 0 && c.fillColorScale.A == 0 {
		return
	}

	if c.Rotation == nil {
		if c.cachedRotation != 0 {
			// It's very likely that the rotation got unset.
			// We should handle it correctly too.
			c.cachedRotation = 0
			c.shaderData["Rotation"] = float32(0)
		}
	} else {
		// Only do a map write operation if previously stored value differs.
		if c.cachedRotation != *c.Rotation {
			c.cachedRotation = *c.Rotation
			c.shaderData["Rotation"] = c.cachedRotation
		}
	}

	r := float64(c.radius)
	width := 2 * r
	pos := opts.Offset.Add(c.Pos.Resolve())
	if c.centered {
		pos = pos.Sub(gmath.Vec{X: r, Y: r})
	}

	var drawOptions ebiten.DrawRectShaderOptions
	drawOptions.Uniforms = c.shaderData
	if opts.Blend != nil {
		drawOptions.Blend = *opts.Blend
	}
	drawOptions.GeoM.Translate(pos.X, pos.Y)
	if c.dashLength == 0 {
		dst.DrawRectShader(int(width), int(width), cache.Global.CircleOutlineShader, &drawOptions)
	} else {
		dst.DrawRectShader(int(width), int(width), cache.Global.DashedCircleOutlineShader, &drawOptions)
	}
}
