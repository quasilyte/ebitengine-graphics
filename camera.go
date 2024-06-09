package graphics

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/gmath"
)

type Camera struct {
	offset     gmath.Vec
	drawOffset gmath.Vec // Rounded

	bounds gmath.Rect

	areaRect gmath.Rect
	areaSize gmath.Vec

	layerMask uint64

	visible  bool
	disposed bool
}

// NewCamera creates a new camera for the drawer.
//
// It covers the entire screen by default.
// Use [SetViewportRect] to change that.
//
// The camera has no bounds by default, use [SetBounds] to set panning limits.
//
// It's advised to only call this function after Ebitengine game has already started.
func NewCamera() *Camera {
	w, h := ebiten.WindowSize()
	camera := &Camera{
		visible:   true,
		layerMask: ^uint64(0),
	}
	camera.SetViewportRect(gmath.Rect{
		Max: gmath.Vec{X: float64(w), Y: float64(h)},
	})
	return camera
}

// SetBounds sets the camera panning limits.
//
// The provided rectangle should not be smaller than
// camera's world size (in the simplest case, bounds=worldSize).
// An exception from this rule is zero rect: a zero-value
// rectangle means "unbound" camera that can go anywhere.
func (c *Camera) SetBounds(bounds gmath.Rect) {
	c.bounds = bounds
}

func (c *Camera) SetViewportRect(rect gmath.Rect) {
	c.areaRect = rect
	c.areaSize = rect.Size()
}

func (c *Camera) GetViewportRect() gmath.Rect {
	return c.areaRect
}

func (c *Camera) GetLayerMask() uint64 {
	return c.layerMask
}

func (c *Camera) SetLayerMask(mask uint64) {
	c.layerMask = mask
}

func (c *Camera) Dispose() {
	c.disposed = true
}

func (c *Camera) IsDisposed() bool {
	return c.disposed
}

func (c *Camera) IsVisible() bool { return c.visible }

func (c *Camera) SetVisibility(visible bool) { c.visible = visible }

func (c *Camera) GetOffset() gmath.Vec {
	return c.offset
}

func (c *Camera) SetOffset(offset gmath.Vec) {
	c.setOffset(offset)
}

func (c *Camera) Pan(delta gmath.Vec) {
	if delta.IsZero() {
		return
	}
	c.setOffset(c.offset.Add(delta))
}

func (c *Camera) setOffset(offset gmath.Vec) {
	offset = c.clampOffset(offset)
	if c.offset == offset {
		return
	}
	c.offset = offset
	c.drawOffset = offset.Rounded()
}

func (c *Camera) getDrawOffset() gmath.Vec {
	return gmath.Vec{
		X: -c.drawOffset.X,
		Y: -c.drawOffset.Y,
	}
}

func (c *Camera) clampOffset(offset gmath.Vec) gmath.Vec {
	if c.bounds.IsZero() {
		return offset
	}

	offset.X = gmath.Clamp(offset.X, c.bounds.Min.X, c.bounds.Max.X-c.areaSize.X)
	offset.Y = gmath.Clamp(offset.Y, c.bounds.Min.Y, c.bounds.Max.Y-c.areaSize.Y)
	return offset
}
