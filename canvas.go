package graphics

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/gmath"
)

type Canvas struct {
	Pos gmath.Pos

	Rotation *gmath.Rad

	spr       *Sprite
	container *Container

	offscreen bool
}

func NewCanvas(cache *Cache) *Canvas {
	c := &Canvas{
		spr:       NewSprite(cache),
		container: NewContainer(),
	}
	c.spr.SetCentered(false)
	return c
}

func (c *Canvas) SetDstImage(img *ebiten.Image) {
	c.spr.SetImage(img)
}

func (c *Canvas) IsDisposed() bool {
	return c.container.IsDisposed()
}

func (c *Canvas) Dispose() {
	c.container.Dispose()
}

func (c *Canvas) IsVisible() bool {
	return c.container.IsVisible()
}

func (c *Canvas) SetVisibility(visible bool) {
	c.container.SetVisibility(visible)
}

func (c *Canvas) IsOffscreen() bool {
	return c.offscreen
}

func (c *Canvas) SetOffscreen(offscreen bool) {
	c.offscreen = offscreen
}

func (c *Canvas) Draw(dst *ebiten.Image) {
	c.DrawWithOptions(dst, DrawOptions{})
}

func (c *Canvas) AddChild(o object) {
	c.container.AddChild(o)
}

func (c *Canvas) DrawWithOptions(dst *ebiten.Image, opts DrawOptions) {
	if !c.container.IsVisible() {
		return
	}

	c.spr.GetImage().Clear()
	c.container.Draw(c.spr.GetImage())

	if !c.offscreen {
		opts.Offset = opts.Offset.Add(c.Pos.Resolve())
		if c.Rotation != nil {
			opts.Rotation += *c.Rotation
		}
		c.spr.DrawWithOptions(dst, opts)
	}
}
