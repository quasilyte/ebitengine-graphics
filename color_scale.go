package graphics

import (
	"image/color"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"
)

// ColorScale is like ebiten.ColorScale, but its values don't need to be premultiplied.
// In a sense, it's like color.NRGBA in comparison with color.RGBA.
//
// For a normal color, we use {1, 1, 1, 1}, for a transparent color it's {0, 0, 0, 0}.
// To double the amount of red, you can use {2, 1, 1, 1} or {1, 0.5, 0.5, 1}.
//
// Use ColorScaleFromRGBA and ColorScaleFromColor if you want to construct
// the color scale object easily.
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

// RGB returns a ColorScale created from the bits of rgb value.
// RGB(0xAABBCC) is identical to R=0xAA, G=0xBB, B=0xCC, A=0xFF.
//
// Note: alpha value is fixed to 0xFF, if you need some other value,
// use RGBA() function instead.
//
// Hint: use RGB(v).Color() to get a color.NRGBA object.
func RGB(rgb uint64) ColorScale {
	r := uint8((rgb & (0xFF << (8 * 2))) >> (8 * 2))
	g := uint8((rgb & (0xFF << (8 * 1))) >> (8 * 1))
	b := uint8((rgb & (0xFF << (8 * 0))) >> (8 * 0))
	return ColorScaleFromRGBA(r, g, b, 0xff)
}

// RGBA returns a ColorScale created from the bits of rgba value.
// RGBA(0xAABBCCEE) is identical to R=0xAA, G=0xBB, B=0xCC, A=0xEE.
//
// Hint: use RGBA(v).Color() to get a color.NRGBA object.
func RGBA(rgb uint64) ColorScale {
	r := uint8((rgb & (0xFF << (8 * 3))) >> (8 * 3))
	g := uint8((rgb & (0xFF << (8 * 2))) >> (8 * 2))
	b := uint8((rgb & (0xFF << (8 * 1))) >> (8 * 1))
	a := uint8((rgb & (0xFF << (8 * 0))) >> (8 * 0))
	return ColorScaleFromRGBA(r, g, b, a)
}

// ColorScaleFromRGBA constructs a ColorScale using ColorScale.SetRGBA method.
func ColorScaleFromRGBA(r, g, b, a uint8) ColorScale {
	var cs ColorScale
	cs.SetRGBA(r, g, b, a)
	return cs
}

// ColorScaleFromRGBA constructs a ColorScale using ColorScale.SetColor method.
func ColorScaleFromColor(c color.NRGBA) ColorScale {
	var cs ColorScale
	cs.SetColor(c)
	return cs
}

// SetColor assigns the color.NRGBA equivalent to a color scale.
func (c *ColorScale) SetColor(rgba color.NRGBA) {
	c.SetRGBA(rgba.R, rgba.G, rgba.B, rgba.A)
}

// SetRGBA is like SetColor, but accepts every color part separately.
func (c *ColorScale) SetRGBA(r, g, b, a uint8) {
	c.R = float32(r) / 255
	c.G = float32(g) / 255
	c.B = float32(b) / 255
	c.A = float32(a) / 255
}

// Color returns the color.NRGBA representation of a color scale.
//
// It will only work correctly for color scales those values are in [0, 1] range.
// If some color value overflows (or underflows) this range, the result
// of this operation is truncated garbage.
//
// This function is mostly useful in combination with RGB() function
// to construct a color.NRGBA easily: RGB(0xAABBCCEE).Color().
func (c ColorScale) Color() color.NRGBA {
	return color.NRGBA{
		R: uint8(c.R * 255),
		G: uint8(c.G * 255),
		B: uint8(c.B * 255),
		A: uint8(c.A * 255),
	}
}

// ScaleAlpha returns the color scale with alpha multiplied by x.
// It doesn't affect R/G/B channels.
func (c ColorScale) ScaleAlpha(x float32) ColorScale {
	c2 := c
	c2.A *= x
	return c2
}

// ScaleRGB multiplies R, G and B color scale components by x.
// It doesn't affect the alpha channel.
func (c ColorScale) ScaleRGB(x float32) ColorScale {
	c2 := c
	c2.R *= x
	c2.G *= x
	c2.B *= x
	return c2
}

// RotateHue returns a color scale with its hue rotated.
// The argument specifies the number of degrees to rotate.
func (c ColorScale) RotateHue(deg float32) ColorScale {
	return hueRotate(c, deg)
}

func (c ColorScale) ToHSL() (h, s, l float32) {
	return rgb2hsl(c)
}

// AsVec3 returns a color scale as RGB slice (vec3 for shaders).
func (c *ColorScale) AsVec3() []float32 {
	return unsafe.Slice(&c.R, 3)
}

// AsVec4returns a color scale as RGBA slice (vec4 for shaders).
func (c *ColorScale) AsVec4() []float32 {
	return unsafe.Slice(&c.R, 4)
}

func (c *ColorScale) undoPremultiply() ColorScale {
	return ColorScale{
		R: c.R / c.A,
		G: c.G / c.A,
		B: c.B / c.A,
		A: c.A,
	}
}

func (c *ColorScale) premultiplyAlpha() ColorScale {
	return ColorScale{
		R: c.R * c.A,
		G: c.G * c.A,
		B: c.B * c.A,
		A: c.A,
	}
}

func (c *ColorScale) toEbitenColorScale() ebiten.ColorScale {
	// This basically turns a NRGBA-style color scale into
	// RGBA-style color scale (alpha-premultiplied).
	var ec ebiten.ColorScale
	ec.SetR(c.R * c.A)
	ec.SetG(c.G * c.A)
	ec.SetB(c.B * c.A)
	ec.SetA(c.A)
	return ec
}
