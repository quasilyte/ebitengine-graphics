package graphics

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/gmath"
)

type Viewport struct {
	camera *Camera
	layers []SceneLayerDrawer
	rect   gmath.Rect
	buf    *ebiten.Image
}

type SceneLayerDrawer interface {
	Update(delta float64)

	DrawWithOptions(dst *ebiten.Image, opts DrawOptions)

	AddChild(o object)
}

func NewViewport(rect gmath.Rect, layers []SceneLayerDrawer) *Viewport {
	if len(layers) == 0 {
		panic("can't create a viewport with 0 layers")
	}

	vp := &Viewport{
		rect:   rect,
		layers: layers,
	}

	width, height := ebiten.WindowSize()
	if rect.Width() != float64(width) || rect.Height() != float64(height) {
		vp.buf = ebiten.NewImage(int(rect.Width()), int(rect.Height()))
	}

	return vp
}

func (vp *Viewport) SetCamera(camera *Camera) {
	vp.camera = camera
	camera.viewportSize = gmath.Vec{
		X: vp.rect.Width(),
		Y: vp.rect.Height(),
	}
}

func (vp *Viewport) AddGraphics(o gsceneGraphics, layer int) {
	l := vp.layers[layer]
	l.AddChild(o.(object))
}

func (vp *Viewport) Update(delta float64) {
	for _, l := range vp.layers {
		l.Update(delta)
	}
}

func (vp *Viewport) Draw(dst *ebiten.Image) {
	// If vp.tmp is nil, it means we can render to dst directly
	// without any temporary buffers.
	// It usually means that the viewport rect is identical to the window size.
	drawDst := dst
	if vp.buf != nil {
		drawDst = vp.buf
		drawDst.Clear()
	}

	var drawOffset gmath.Vec
	if vp.camera != nil {
		drawOffset = vp.camera.getDrawOffset()
	}
	vp.draw(drawDst, drawOffset)

	if drawDst == dst {
		// Nothing else to do.
		// Every draw() call already targeted the original dst.
		return
	}

	// Copy the result to the destination.
	var options ebiten.DrawImageOptions
	options.GeoM.Translate(vp.rect.Min.X, vp.rect.Min.Y)
	dst.DrawImage(vp.buf, &options)
}

func (vp *Viewport) draw(dst *ebiten.Image, offset gmath.Vec) {
	options := DrawOptions{
		Offset: offset,
	}

	for _, l := range vp.layers {
		l.DrawWithOptions(dst, options)
	}
}
