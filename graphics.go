package graphics

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/gmath"
)

type DrawOptions struct {
	Offset gmath.Vec

	Rotation gmath.Rad

	// Blend is an optional blend mode override.
	// You usually want to use a predefined blend from Ebitengine and
	// assign it like DrawOptopns{Blend: &ebiten.BlendCopy}.
	// We're using a pointer here mostly to decrease the [DrawOptions]
	// object size as most of the time this field is going to be nil.
	Blend *ebiten.Blend
}

// TODO: make it exported?
type object interface {
	gsceneGraphics

	DrawWithOptions(dst *ebiten.Image, o DrawOptions)
	Dispose()
}
