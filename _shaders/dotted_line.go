//kage:unit pixels

//go:build ignore

package main

var PointA vec2
var PointB vec2
var DotRadius float
var DotSpacing float
var Color vec4

func Fragment(pos vec4, _ vec2, _ vec4) vec4 {
	origin := imageDstOrigin()
	zpos := quantizeToPixel(pos.xy - origin)

	r := DotRadius
	a := PointA
	b := PointB

	ab := b - a
	lineLength := length(ab)

	ap := zpos - a
	projection := (dot(ap, ab) / lineLength)

	dotIndex := round(projection / DotSpacing)
	dotCenter := a + (dotIndex*DotSpacing/lineLength)*ab
	dotCenter = quantizeToPixel(dotCenter)

	distanceToDot := distance(zpos, dotCenter)
	if distanceToDot < r {
		return Color
	}
	return vec4(0)
}

func quantizeToPixel(pos vec2) vec2 {
	return floor(pos) + vec2(0.5, 0.5)
}

func round(x float) float {
	return floor(x + 0.5)
}
