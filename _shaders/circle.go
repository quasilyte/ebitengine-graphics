//kage:unit pixels

//go:build ignore

package main

var Radius float
var OutlineWidth float
var OutlineColor vec4
var FillColor vec4

func Fragment(_ vec4, pos vec2, _ vec4) vec4 {
	origin := imageSrc0Origin()
	zpos := pos - origin
	r := Radius

	center := vec2(r, r)
	dist := distance(zpos, center)
	if dist > r {
		return vec4(0)
	}
	if dist >= r-OutlineWidth {
		return OutlineColor
	}
	return FillColor
}
