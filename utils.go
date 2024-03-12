package graphics

import (
	"golang.org/x/exp/constraints"
)

func getFlag[T constraints.Integer](flags T, bit T) bool {
	return flags&bit != 0
}

func clearFlag[T constraints.Integer](flags *T, bit T) {
	*flags &^= bit
}

func setFlag[T constraints.Integer](flags *T, bit T, value bool) {
	if value {
		*flags |= bit
	} else {
		clearFlag(flags, bit)
	}
}
