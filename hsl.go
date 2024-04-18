package graphics

import (
	"math"
)

func hueRotate(rgb ColorScale, deg float32) ColorScale {
	h, s, l := rgb2hsl(rgb)
	h = float32(math.Mod(float64(h+(deg/360)), 1.0))
	return hsl2rgb(h, s, l, rgb.A)
}

func rgb2hsl(rgb ColorScale) (h, s, l float32) {
	r := rgb.R
	g := rgb.G
	b := rgb.B

	clrMax := max(max(r, g), b)
	clrMin := min(min(r, g), b)

	l = (clrMax + clrMin) / 2

	delta := clrMax - clrMin
	if delta == 0 {
		return 0, 0, l
	}

	if l < 0.5 {
		s = delta / (clrMax + clrMin)
	} else {
		s = delta / (2 - clrMax - clrMin)
	}

	r2 := (((clrMax - r) / 6) + (delta / 2)) / delta
	g2 := (((clrMax - g) / 6) + (delta / 2)) / delta
	b2 := (((clrMax - b) / 6) + (delta / 2)) / delta
	switch {
	case r == clrMax:
		h = b2 - g2
	case g == clrMax:
		h = (1.0 / 3.0) + r2 - b2
	case b == clrMax:
		h = (2.0 / 3.0) + g2 - r2
	}

	switch {
	case h < 0:
		h++
	case h > 1:
		h--
	}

	return h, s, l
}

func hsl2rgb(h, s, l, a float32) ColorScale {
	if s == 0 {
		return ColorScale{R: l, G: l, B: l, A: a}
	}

	var v1, v2 float32
	if l < 0.5 {
		v2 = l * (1 + s)
	} else {
		v2 = (l + s) - (s * l)
	}

	v1 = 2*l - v2

	r := hueToRGB(v1, v2, h+(1.0/3.0))
	g := hueToRGB(v1, v2, h)
	b := hueToRGB(v1, v2, h-(1.0/3.0))

	return ColorScale{R: r, G: g, B: b, A: a}
}

func hueToRGB(v1, v2, h float32) float32 {
	if h < 0 {
		h++
	}
	if h > 1 {
		h--
	}
	switch {
	case 6*h < 1:
		return (v1 + (v2-v1)*6*h)
	case 2*h < 1:
		return v2
	case 3*h < 2:
		return v1 + (v2-v1)*((2.0/3.0)-h)*6
	}
	return v1
}
