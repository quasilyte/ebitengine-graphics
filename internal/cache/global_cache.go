package cache

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/quasilyte/gmath"
)

var Global = &cache{}

func init() {
	Global.WhitePixel = ebiten.NewImage(1, 1)
	Global.WhitePixel.Fill(color.White)

	// Just some pseudo-random fixed seed to make graphics
	// somewhat random (it's used in particle systems).
	// We can add a function to override the default seed
	// if there is a feature request for it.
	Global.Rand.SetSeed(271828)

	Global.ScratchVertices = make([]ebiten.Vertex, 0, 24*4)
	Global.ScratchIndices = make([]uint16, 0, 24*6)
}

// cache is a storage that is shared between all
// graphical elements.
type cache struct {
	FontInfoList []FontInfo
	FontInfoMap  map[text.Face]uint16

	ShadersCompiled           bool
	CircleOutlineShader       *ebiten.Shader
	DashedCircleOutlineShader *ebiten.Shader
	DottedLineShader          *ebiten.Shader

	Rand            gmath.Rand
	WhitePixel      *ebiten.Image
	ScratchVertices []ebiten.Vertex
	ScratchIndices  []uint16
}

type FontInfo struct {
	Face       text.Face
	CapHeight  float64
	LineHeight float64
}

func (c *cache) InternFontFace(ff text.Face) uint16 {
	if c.FontInfoMap == nil {
		c.FontInfoMap = make(map[text.Face]uint16, 8)
	}

	if id, ok := c.FontInfoMap[ff]; ok {
		return id
	}

	id := uint16(len(c.FontInfoList))
	c.FontInfoMap[ff] = id

	m := ff.Metrics()
	capHeight := math.Abs(m.CapHeight)

	// > HLineGap = fixed26_6ToFloat64(fm.Height - fm.Ascent - fm.Descent)
	// > HAscent:   fixed26_6ToFloat64(fm.Ascent)
	// > HDescent:  fixed26_6ToFloat64(fm.Descent)
	lineHeight := m.HLineGap + m.HAscent + m.HDescent

	c.FontInfoList = append(c.FontInfoList, FontInfo{
		Face:       ff,
		CapHeight:  capHeight,
		LineHeight: lineHeight,
	})

	return id
}
