package xmath

import (
	"math"
)

// TODO: move it to gmath?

type Geom32 struct {
	A1 float32 // The actual 'a' value minus 1
	B  float32
	C  float32
	D1 float32 // The actual 'd' value minus 1
	Tx float32
	Ty float32
}

func (g *Geom32) Translate(tx, ty float32) {
	g.Tx += tx
	g.Ty += ty
}

func (g *Geom32) ApplyX(x, y float32) float32 {
	return (g.A1+1)*x + g.B*y + g.Tx
}

func (g *Geom32) ApplyY(x, y float32) float32 {
	return g.C*x + (g.D1+1)*y + g.Ty
}

func (g *Geom32) Scale(x, y float32) {
	a := (g.A1 + 1) * x
	b := g.B * x
	tx := g.Tx * x
	c := g.C * y
	d := (g.D1 + 1) * y
	ty := g.Ty * y

	g.A1 = a - 1
	g.B = b
	g.C = c
	g.D1 = d - 1
	g.Tx = tx
	g.Ty = ty
}

func (g *Geom32) Rotate(theta float64) {
	sin64, cos64 := math.Sincos(theta)
	sin := float32(sin64)
	cos := float32(cos64)

	a := cos*(g.A1+1) - sin*g.C
	b := cos*g.B - sin*(g.D1+1)
	tx := cos*g.Tx - sin*g.Ty
	c := sin*(g.A1+1) + cos*g.C
	d := sin*g.B + cos*(g.D1+1)
	ty := sin*g.Tx + cos*g.Ty

	g.A1 = a - 1
	g.B = b
	g.C = c
	g.D1 = d - 1
	g.Tx = tx
	g.Ty = ty
}
