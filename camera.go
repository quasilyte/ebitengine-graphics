package graphics

import (
	"github.com/quasilyte/gmath"
)

type Camera struct {
	offset     gmath.Vec
	drawOffset gmath.Vec // Rounded

	bounds gmath.Rect

	viewportSize gmath.Vec

	disposed bool
}

func NewCamera() *Camera {
	return &Camera{}
}

// SetBounds sets the camera display limits.
//
// The provided rectangle should not be smaller than
// camera's world size (in the simplest case, bounds=worldSize).
// An exception from this rule is zero rect: a zero-value
// rectangle means "unbound" camera that can go anywhere.
func (c *Camera) SetBounds(bounds gmath.Rect) {
	c.bounds = bounds
}

func (c *Camera) Dispose() {
	c.disposed = true
}

func (c *Camera) IsDisposed() bool {
	return c.disposed
}

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

	offset.X = gmath.Clamp(offset.X, c.bounds.Min.X, c.bounds.Max.X-c.viewportSize.X)
	offset.Y = gmath.Clamp(offset.Y, c.bounds.Min.Y, c.bounds.Max.Y-c.viewportSize.Y)
	return offset
}
