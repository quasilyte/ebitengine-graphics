package graphics_test

import (
	"testing"
	"unsafe"

	graphics "github.com/quasilyte/ebitengine-graphics"
)

func TestLabelSize(t *testing.T) {
	if unsafe.Sizeof(int(0)) != 8 {
		t.Skip("this test is only executed on 64-bit platforms")
	}

	wantSize := uintptr(104)
	haveSize := unsafe.Sizeof(graphics.Label{})
	if wantSize != haveSize {
		t.Fatalf("sizeof(Label):\nhave: %d\nwant: %d", haveSize, wantSize)
	}
}
