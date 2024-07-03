package particle

import (
	"github.com/quasilyte/gmath"
)

func vec32(v gmath.Vec) gmath.Vec32 {
	return gmath.Vec32{
		X: float32(v.X),
		Y: float32(v.Y),
	}
}

func rotatedVec(v gmath.Vec, angle *gmath.Rad) gmath.Vec {
	if angle == nil {
		return v
	}
	return v.Rotated(*angle)
}
