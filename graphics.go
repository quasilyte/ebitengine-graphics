package graphics

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/gmath"
)

type DrawOptions struct {
	Offset gmath.Vec

	Rotation gmath.Rad
}

type object interface {
	DrawWithOptions(dst *ebiten.Image, o DrawOptions)

	IsDisposed() bool

	Dispose()
}
