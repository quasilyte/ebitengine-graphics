package particle

import (
	"github.com/quasilyte/gmath"
)

func rotatedVec(v gmath.Vec, angle *gmath.Rad) gmath.Vec {
	if angle == nil {
		return v
	}
	return v.Rotated(*angle)
}
