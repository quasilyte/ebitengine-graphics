package particle

import (
	"math"
)

type geom32 struct {
	a1 float32 // The actual 'a' value minus 1
	b  float32
	c  float32
	d1 float32 // The actual 'd' value minus 1
	tx float32
	ty float32
}

func (g *geom32) Translate(tx, ty float32) {
	g.tx += tx
	g.ty += ty
}

func (g *geom32) ApplyX(x, y float32) float32 {
	return (g.a1+1)*x + g.b*y + g.tx
}

func (g *geom32) ApplyY(x, y float32) float32 {
	return g.c*x + (g.d1+1)*y + g.ty
}

func (g *geom32) Scale(x, y float32) {
	a := (g.a1 + 1) * x
	b := g.b * x
	tx := g.tx * x
	c := g.c * y
	d := (g.d1 + 1) * y
	ty := g.ty * y

	g.a1 = a - 1
	g.b = b
	g.c = c
	g.d1 = d - 1
	g.tx = tx
	g.ty = ty
}

func (g *geom32) Rotate(theta float64) {
	sin64, cos64 := math.Sincos(theta)
	sin := float32(sin64)
	cos := float32(cos64)

	a := cos*(g.a1+1) - sin*g.c
	b := cos*g.b - sin*(g.d1+1)
	tx := cos*g.tx - sin*g.ty
	c := sin*(g.a1+1) + cos*g.c
	d := sin*g.b + cos*(g.d1+1)
	ty := sin*g.tx + cos*g.ty

	g.a1 = a - 1
	g.b = b
	g.c = c
	g.d1 = d - 1
	g.tx = tx
	g.ty = ty
}
