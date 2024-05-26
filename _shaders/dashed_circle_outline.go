//kage:unit pixels

//go:build ignore

package main

var Radius float
var Rotation float
var OutlineWidth float
var DashLength float
var DashGap float
var OutlineColor vec4

func Fragment(_ vec4, pos vec2, _ vec4) vec4 {
	origin := imageSrc0Origin()
	zpos := pos - origin
	r := Radius

	center := vec2(r, r)
	dist := distance(zpos, center)
	if dist > r {
		return vec4(0)
	}

	direction := zpos - center
	angle := atan2(direction.y, direction.x) - Rotation
	arcLength := angle * Radius
	totalLength := DashLength + DashGap
	isDash := mod(arcLength, totalLength) < DashLength

	if !isDash {
		return vec4(0)
	}
	if dist >= r-OutlineWidth {
		return OutlineColor
	}
	return vec4(0)
}
