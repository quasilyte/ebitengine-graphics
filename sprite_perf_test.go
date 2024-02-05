package graphics

import (
	"testing"
)

func BenchmarkSpriteBoundsRectCentered(b *testing.B) {
	s := NewSprite(NewCache())
	s.SetCentered(true)
	for i := 0; i < b.N; i++ {
		_ = s.BoundsRect()
	}
}

func BenchmarkSpriteBoundsRect(b *testing.B) {
	s := NewSprite(NewCache())
	s.SetCentered(false)
	for i := 0; i < b.N; i++ {
		_ = s.BoundsRect()
	}
}
