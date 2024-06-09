package graphics

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type gsceneGraphics = interface {
	Draw(dst *ebiten.Image)
	IsDisposed() bool
}

type gsceneViewport = interface {
	AddGraphics(g gsceneGraphics, layer int)
}
