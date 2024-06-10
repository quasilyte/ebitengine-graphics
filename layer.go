package graphics

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Layer struct {
	objects []object
}

func NewLayer() *Layer {
	return &Layer{objects: make([]object, 0, 16)}
}

func (l *Layer) AddChild(o object) {
	l.objects = append(l.objects, o)
}

func (l *Layer) Update(_ float64) {
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
	for _, o := range l.objects {
		o.DrawWithOptions(dst, opts)
	}
}
