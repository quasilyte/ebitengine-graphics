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

func (g *promotedGraphics) DrawWithOptions(dst *ebiten.Image, _ DrawOptions) {
	g.g.Draw(dst)
}

func DecorateDraw(o Object, f func(dst *ebiten.Image, opts DrawOptions)) Object {
	return &decoratedObject{
		underlying: o,
		f:          f,
	}
}

type decoratedObject struct {
	underlying Object
	f          func(dst *ebiten.Image, opts DrawOptions)
}

func (o *decoratedObject) IsDisposed() bool {
	return o.underlying.IsDisposed()
}

func (o *decoratedObject) Draw(dst *ebiten.Image) {
	o.f(dst, DrawOptions{})
}

func (o *decoratedObject) DrawWithOptions(dst *ebiten.Image, opts DrawOptions) {
	o.f(dst, opts)
}

func BindDrawDst(o Object, dst *ebiten.Image) Object {
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

func (b *dstBinder) DrawWithOptions(_ *ebiten.Image, opts DrawOptions) {
	b.drawer.DrawWithOptions(b.dst, opts)
}
