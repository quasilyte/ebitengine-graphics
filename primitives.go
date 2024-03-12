package graphics

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/gmath"
)

var emptyImage = ebiten.NewImage(3, 3)

var whitePixel = emptyImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)

func init() {
	emptyImage.Fill(color.White)
}

func drawLine(dst *ebiten.Image, pos1, pos2 gmath.Vec, width float64, cs ebiten.ColorScale) {
	// TODO: compare with vector API.

	x1 := pos1.X
	y1 := pos1.Y
	x2 := pos2.X
	y2 := pos2.Y

	length := math.Hypot(x2-x1, y2-y1)

	var drawOptions ebiten.DrawImageOptions
	drawOptions.GeoM.Scale(length, width)
	drawOptions.GeoM.Rotate(math.Atan2(y2-y1, x2-x1))
	drawOptions.GeoM.Translate(x1, y1)

	drawOptions.ColorScale = cs

	dst.DrawImage(whitePixel, &drawOptions)
}
