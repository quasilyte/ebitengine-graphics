package graphics

import (
	"image"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/gmath"
)

type SceneDrawer struct {
	cameras []installedCamera

	// This camera is used only when len(cameras) is 0.
	// Stored as a slice for convenience.
	defaultCamera []installedCamera

	viewportRect gmath.Rect

	layers []SceneLayerDrawer
	buf    *ebiten.Image
}

type installedCamera struct {
	c *Camera

	buf *ebiten.Image

	cachedRect gmath.Rect
}

type SceneLayerDrawer interface {
	Update(delta float64)

	DrawWithOptions(dst *ebiten.Image, opts DrawOptions)

	AddChild(o gsceneGraphics)
}

// NewSceneDrawer creates a configured [gscene.Drawer] for the scene.
//
// It will have a default camera that always has its offset at (0, 0)
// and its rendered at a full display.
// This default camera is used if no other cameras are available.
// Use [SceneDrawer.AddCamera] to install custom cameras.
//
// See [SceneDrawer] doc comments for more info.
//
// It's advised to only call this function after Ebitengine game has already started.
func NewSceneDrawer(layers []SceneLayerDrawer) *SceneDrawer {
	if len(layers) == 0 {
		panic("can't create a scene drawer with 0 layers")
	}

	w, h := ebiten.WindowSize()
	viewportRect := gmath.Rect{
		Max: gmath.Vec{X: float64(w), Y: float64(h)},
	}
	d := &SceneDrawer{
		layers:       layers,
		viewportRect: viewportRect,
	}

	d.defaultCamera = []installedCamera{
		{c: NewCamera()},
	}

	return d
}

func (d *SceneDrawer) AddCamera(camera *Camera) {
	d.cameras = append(d.cameras, installedCamera{
		c: camera,
	})
}

func (d *SceneDrawer) RemoveCamera(camera *Camera) {
	index := slices.IndexFunc(d.cameras, func(ic installedCamera) bool {
		return ic.c == camera
	})
	if index < 0 {
		return
	}
	d.cameras = slices.Delete(d.cameras, index, index+1)
}

func (d *SceneDrawer) AddGraphics(o gsceneGraphics, layer int) {
	l := d.layers[layer]
	l.AddChild(o)
}

func (d *SceneDrawer) Update(delta float64) {
	for _, l := range d.layers {
		l.Update(delta)
	}
}

func (d *SceneDrawer) Draw(dst *ebiten.Image) {
	cameras := d.cameras
	if len(cameras) == 0 {
		cameras = d.defaultCamera // Contains a single full-display camera
	}

	for i := range cameras {
		camera := &cameras[i]

		cameraDst := dst
		if d.cameraNeedsTmpBuf(camera) {
			cameraDst = d.cameraAdjustedBuf(camera, d.getBuf())
			cameraDst.Clear()
		}

		options := DrawOptions{
			Offset: camera.c.getDrawOffset(),
		}
		for i, l := range d.layers {
			if i < 64 {
				if uint64(1<<i)&camera.c.layerMask == 0 {
					continue
				}
			}
			l.DrawWithOptions(cameraDst, options)
		}

		if cameraDst != dst {
			// Copy the result to the actual destination.
			if camera.c.pp == nil {
				// A simple drawing without post-processing.
				var options ebiten.DrawImageOptions
				options.GeoM.Translate(camera.c.areaRect.Min.X, camera.c.areaRect.Min.Y)
				dst.DrawImage(cameraDst, &options)
			} else {
				camera.c.pp.PostProcess(dst, cameraDst, DrawOptions{
					Offset: gmath.Vec{
						X: camera.c.areaRect.Min.X,
						Y: camera.c.areaRect.Min.Y,
					},
				})
			}
		}
	}
}

func (d *SceneDrawer) cameraAdjustedBuf(camera *installedCamera, buf *ebiten.Image) *ebiten.Image {
	// Maybe we already have a suitable subimage?
	// If camera viewport sizes are the same, use it.
	// This is a very common case.
	if camera.buf != nil {
		if camera.cachedRect == camera.c.areaRect {
			return camera.buf
		}
	}

	// Calculate a subimage and cache it.
	camera.buf = buf.SubImage(image.Rectangle{
		Max: camera.c.areaSize.ToStd(),
	}).(*ebiten.Image)
	camera.cachedRect = camera.c.areaRect

	return camera.buf
}

func (d *SceneDrawer) cameraNeedsTmpBuf(camera *installedCamera) bool {
	return camera.c.areaRect != d.viewportRect
}

func (d *SceneDrawer) getBuf() *ebiten.Image {
	if d.buf != nil {
		return d.buf
	}

	d.buf = ebiten.NewImage(int(d.viewportRect.Max.X), int(d.viewportRect.Max.Y))
	return d.buf
}
