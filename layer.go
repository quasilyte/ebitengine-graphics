package graphics

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Layer is a simple layer that renders objects in the order they were added.
//
// The objects inside this layer are displayed with respect to the camera transformation.
//
// It expects graphics to implement [Object] interface.
// If something implements only a simple gscene Graphics interface,
// use [StaticLayer].
type Layer struct {
	objects    []Object
	needFilter bool
}

func NewLayer() *Layer {
	return &Layer{objects: make([]Object, 0, 16)}
}

func (l *Layer) AddChild(g gsceneGraphics) {
	l.objects = append(l.objects, g.(Object))
	l.needFilter = true
}

func (l *Layer) Update(_ float64) {
	l.needFilter = true
}

func (l *Layer) filter() {
	liveObjects := l.objects[:0]
	for _, o := range l.objects {
		if o.IsDisposed() {
			continue
		}
		liveObjects = append(liveObjects, o)
	}
	l.objects = liveObjects
}

func (l *Layer) DrawWithOptions(dst *ebiten.Image, opts DrawOptions) {
	if l.needFilter {
		l.filter()
	}
	l.needFilter = false

	for _, o := range l.objects {
		o.DrawWithOptions(dst, opts)
	}
}
