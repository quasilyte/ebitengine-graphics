package graphics

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// StaticLayer is like [Layer], but objects are rendered in a camera-independent way.
// It renders the objects in the order they were added to the layer.
//
// The objects inside this layer ignore the camera transformation,
// this is why having just Draw() method on objects is enough.
//
// This layer is well-suited for HUDs and overlays.
type StaticLayer struct {
	objects []gsceneGraphics
}

func NewStaticLayer() *StaticLayer {
	return &StaticLayer{objects: make([]gsceneGraphics, 0, 16)}
}

func (l *StaticLayer) AddChild(g gsceneGraphics) {
	l.objects = append(l.objects, g)
}

func (l *StaticLayer) Update(_ float64) {
	liveObjects := l.objects[:0]
	for _, o := range l.objects {
		if o.IsDisposed() {
			continue
		}
		liveObjects = append(liveObjects, o)
	}
	l.objects = liveObjects
}

func (l *StaticLayer) DrawWithOptions(dst *ebiten.Image, _ DrawOptions) {
	for _, o := range l.objects {
		o.Draw(dst)
	}
}
