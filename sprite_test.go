package graphics_test

import (
	"testing"
	"unsafe"

	graphics "github.com/quasilyte/gscene-graphics"
)

func TestSpriteSize(t *testing.T) {
	if unsafe.Sizeof(int(0)) != 8 {
		t.Skip("this test is only executed on 64-bit platforms")
	}

	// Sprite objects are heap-allocated in 99.9% of cases.
	// They're also one of the most common used type of graphics in an average game.
	// This means that their size matters.
	wantSize := uintptr(128)
	haveSize := unsafe.Sizeof(graphics.Sprite{})
	if wantSize != haveSize {
		t.Fatalf("sizeof(Sprite):\nhave: %d\nwant: %d", haveSize, wantSize)
	}
}
