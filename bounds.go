package graphics

import (
	"github.com/quasilyte/gmath"
)

func lineBoundsRect(from, to gmath.Pos) gmath.Rect {
	pos1 := from.Resolve()
	pos2 := to.Resolve()
	x0 := pos1.X
	x1 := pos2.X
	y0 := pos1.Y
	y1 := pos2.Y
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	return gmath.Rect{
		Min: gmath.Vec{X: x0, Y: y0},
		Max: gmath.Vec{X: x1, Y: y1},
	}
}
