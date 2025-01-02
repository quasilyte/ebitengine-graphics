package graphics

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/gmath"
)

// Camera implements a 2D camera for the [SceneDrawer].
//
// It's pretty barebones, you might want to wrap it into
// your game's camera type and add functionality like lerping between A and B.
// Add [Camera] itself into the [SceneDrawer] and put it inside your
// own camera implementation. Then call appropriate things like [SetOffset]
// and [Pan] on the wrapped camera to implement something fancier.
//
// Some terminology:
// * screen coordinates - min=(0,0) max=(windowWidth,windowHeight)
// * world coordinates - arbitrary values that spefify the object's location inside the game world
//
// Let's assume there is a world coordinate {32, 32};
// if the camera's offset is {0, 0}, then screen coordinate is {32, 32},
// if the camera's offset is {32, 32}, then screen coordinate is {0, 0}.
//
// Converting coordinates:
// * screen to world: screenPos.Add(camera.GetOffset())
// * world to screen: worldPos.Sub(camera.GetOffset())
//
// Pay attention to the docs, they should tell you which kind of a position
// is expected for an argument and/or method's return value.
type Camera struct {
	offset     gmath.Vec
	drawOffset gmath.Vec // Rounded

	bounds gmath.Rect

	areaRect gmath.Rect
	areaSize gmath.Vec

	layerMask uint64

	pp PostProcessor
}

// NewCamera creates a new camera for the [SceneDrawer].
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
		layerMask: ^uint64(0),
	}
	camera.SetViewportRect(gmath.Rect{
		Max: gmath.Vec{X: float64(w), Y: float64(h)},
	})
	return camera
}

func (c *Camera) SetPostProcessor(pp PostProcessor) {
	c.pp = pp
}

func (c *Camera) GetBounds() gmath.Rect {
	return c.bounds
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

// GetViewportRect returns the camera's rendering rectangle.
//
// This rectangle is in screen coordinates, meaning it is unaffected by the camera pos.
func (c *Camera) GetViewportRect() gmath.Rect {
	return c.areaRect
}

// SetViewportRect changes the camera's viewport area to render to.
//
// The rect is in screen coordinates.
func (c *Camera) SetViewportRect(rect gmath.Rect) {
	c.areaRect = rect
	c.areaSize = rect.Size()
}

// GetLayerMask returns the current camera's layer bitmask.
// See [SetLayerMask] doc comment to learn more about the bitmask.
func (c *Camera) GetLayerMask() uint64 {
	return c.layerMask
}

// SetLayerMask updates the camera's layer bitmask.
//
// The nth bit of the mask controls whether nth layer should be
// rendered onto the camera.
// By default, the camera layermask is "all ones", meaning all layers are rendered.
//
// This mask is useful when you want to draw some layers only to particular cameras.
// For the simplest kinds of games this feature is not needed and you may leave it as is.
//
// This bitmask can affect up to first 64 layers, any other layer is always enabled.
func (c *Camera) SetLayerMask(mask uint64) {
	c.layerMask = mask
}

// GetCenterOffset returns the camera current offset translated
// to the viewport rect's center.
// To get the untranslated position, use [GetOffset].
//
// The returned pos is in world coordinates.
func (c *Camera) GetCenterOffset() gmath.Vec {
	return c.offset.Add(c.areaSize.Mulf(0.5)).Rounded()
}

// SetCenterOffset centers the camera around given position.
// After the clamping rules apply, the pos may end up not being in the perfect
// center of the camera's viewport rect.
//
// The return value reports whether the position was actually updated.
//
// The pos parameter should be in world coordinates.
func (c *Camera) SetCenterOffset(pos gmath.Vec) bool {
	return c.setOffset(pos.Sub(c.areaSize.Mulf(0.5)))
}

// GetOffset returns the camera current offset.
// The pos is given for the top-left corner of the camera's viewport rect.
// To get the center position, use [GetCenterOffset].
//
// The returned pos is in world coordinates.
func (c *Camera) GetOffset() gmath.Vec {
	return c.offset
}

// SetOffset assigns a new offset to the camera.
// It will be clamped to fit the camera bounds.
//
// The return value reports whether the position was actually updated.
//
// The pos parameter should be in world coordinates.
func (c *Camera) SetOffset(pos gmath.Vec) bool {
	return c.setOffset(pos)
}

// Pan adds the specified camera position delta to the camera's current offset.
// It's a shorthand to c.SetOffset(c.GetOffset().Add(delta)).
// The same clamping rules apply as in [SetOffset].
//
// The return value reports whether the position was actually updated.
func (c *Camera) Pan(delta gmath.Vec) bool {
	if delta.IsZero() {
		return false
	}
	return c.setOffset(c.offset.Add(delta))
}

func (c *Camera) setOffset(offset gmath.Vec) bool {
	offset = c.clampOffset(offset)
	if c.offset == offset {
		return false
	}
	c.offset = offset
	c.drawOffset = offset.Rounded()
	return true
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
