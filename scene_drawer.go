package graphics

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/gmath"
)

type SceneDrawer struct {
	viewports []*Viewport
}

// NewSimpleSceneDrawer is a common case of a single viewport drawer configuration.
// It also uses a full game window size as its viewport rect.
// It creates the requested number of layers ([Layer] typed).
// The [camera] argument can be nil.
//
// If you need multiple viewports or you need
// to use layer types other than [Layer], use [NewSceneDrawer].
// See [SceneDrawer] doc comments for more info.
func NewSimpleSceneDrawer(camera *Camera, numLayers int) *SceneDrawer {
	width, height := ebiten.WindowSize()
	rect := gmath.Rect{
		Max: gmath.Vec{X: float64(width), Y: float64(height)},
	}

	layers := make([]SceneLayerDrawer, numLayers)
	for i := range layers {
		layers[i] = NewLayer()
	}

	vp := NewViewport(rect, layers)
	if camera != nil {
		vp.SetCamera(camera)
	}
	return NewSceneDrawer([]*Viewport{vp})
}

// NewSceneDrawer creates a configured [gscene.Drawer] for the scene.
//
// See [SceneDrawer] doc comments for more info.
func NewSceneDrawer(viewports []*Viewport) *SceneDrawer {
	if len(viewports) == 0 {
		panic("can't create a scene drawer with 0 viewports")
	}
	return &SceneDrawer{
		viewports: viewports,
	}
}

func (d *SceneDrawer) Viewport(index int) gsceneViewport {
	return d.viewports[index]
}

func (d *SceneDrawer) Update(delta float64) {
	for _, vp := range d.viewports {
		vp.Update(delta)
	}
}

func (d *SceneDrawer) Draw(dst *ebiten.Image) {
	for _, vp := range d.viewports {
		vp.Draw(dst)
	}
}
