package graphics

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type ColorScale struct {
	R float32
	G float32
	B float32
	A float32
}

var (
	defaultColorScale = ColorScale{1, 1, 1, 1}
	transparentColor  = ColorScale{0, 0, 0, 0}
)

// ColorScaleFromRGBA constructs a ColorScale using ColorScale.SetRGBA method.
func ColorScaleFromRGBA(r, g, b, a uint8) ColorScale {
	var cs ColorScale
	cs.SetRGBA(r, g, b, a)
	return cs
}

// ColorScaleFromRGBA constructs a ColorScale using ColorScale.SetColor method.
func ColorScaleFromColor(c color.RGBA) ColorScale {
	var cs ColorScale
	cs.SetColor(c)
	return cs
}

func (c *ColorScale) SetColor(rgba color.RGBA) {
	c.SetRGBA(rgba.R, rgba.G, rgba.B, rgba.A)
}

func (c *ColorScale) SetRGBA(r, g, b, a uint8) {
	c.R = float32(r) / 255
	c.G = float32(g) / 255
	c.B = float32(b) / 255
	c.A = float32(a) / 255
}

func (c *ColorScale) toEbitenColorScale() ebiten.ColorScale {
	var ec ebiten.ColorScale
	ec.SetR(c.R * c.A)
	ec.SetG(c.G * c.A)
	ec.SetB(c.B * c.A)
	ec.SetA(c.A)
	return ec
}
