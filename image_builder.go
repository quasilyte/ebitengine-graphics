package graphics

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type ImageBuilder struct {
	dst   *ebiten.Image
	cache *Cache
	s     *Sprite
}

func NewImageBuilder(cache *Cache, w, h int) *ImageBuilder {
	return &ImageBuilder{
		dst:   ebiten.NewImage(w, h),
		cache: cache,
	}
}

func (b *ImageBuilder) Image() *ebiten.Image {
	return b.dst
}

func (d *ImageBuilder) DrawSprite(s *Sprite, opts DrawOptions) {
	s.DrawWithOptions(d.dst, opts)
}

func (d *ImageBuilder) DrawImage(img *ebiten.Image, opts DrawOptions) {
	if d.s == nil {
		d.s = NewSprite(d.cache)
	}
	d.s.SetImage(img)
	d.DrawSprite(d.s, opts)
}
