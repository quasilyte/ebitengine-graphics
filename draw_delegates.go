package graphics

import (
	"github.com/hajimehoshi/ebiten/v2"
)

func ToObject(g gsceneGraphics) Object {
	return &promotedGraphics{g: g}
}

type promotedGraphics struct {
	g gsceneGraphics
}

func (g *promotedGraphics) IsDisposed() bool {
	return g.g.IsDisposed()
}

func (g *promotedGraphics) Draw(dst *ebiten.Image) {
	g.g.Draw(dst)
}

func (g *promotedGraphics) DrawWithOptions(dst *ebiten.Image, opts DrawOptions) {
	g.g.Draw(dst)
}

func BindDrawDst(o Object, dst *ebiten.Image) *dstBinder {
	return &dstBinder{
		drawer: o,
		dst:    dst,
	}
}

type dstBinder struct {
	drawer Object
	dst    *ebiten.Image
}

func (b *dstBinder) IsDisposed() bool {
	return b.drawer.IsDisposed()
}

func (b *dstBinder) Draw(dst *ebiten.Image) {
	b.DrawWithOptions(dst, DrawOptions{})
}

func (b *dstBinder) DrawWithOptions(dst *ebiten.Image, opts DrawOptions) {
	b.drawer.DrawWithOptions(b.dst, opts)
}
