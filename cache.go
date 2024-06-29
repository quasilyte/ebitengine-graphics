package graphics

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/gmath"
	"golang.org/x/image/font"
)

var globalCache = &cache{}

func init() {
	globalCache.whitePixel = ebiten.NewImage(1, 1)
	globalCache.whitePixel.Fill(color.White)

	// Just some pseudo-random fixed seed to make graphics
	// somewhat random (it's used in particle systems).
	// We can add a function to override the default seed
	// if there is a feature request for it.
	globalCache.rand.SetSeed(271828)

	globalCache.scratchVertices = make([]ebiten.Vertex, 0, 24*4)
	globalCache.scratchIndices = make([]uint16, 0, 24*6)
	globalCache.scratchEmitterBatch = make([]*ParticleEmitter, 0, 8)
}

// cache is a storage that is shared between all
// graphical elements.
type cache struct {
	fontInfoList []fontInfo
	fontInfoMap  map[font.Face]uint16

	shadersCompiled           bool
	circleOutlineShader       *ebiten.Shader
	dashedCircleOutlineShader *ebiten.Shader

	rand                gmath.Rand
	whitePixel          *ebiten.Image
	scratchEmitterBatch []*ParticleEmitter
	scratchVertices     []ebiten.Vertex
	scratchIndices      []uint16
}

type fontInfo struct {
	ff         font.Face
	capHeight  float64
	lineHeight float64
}

func (c *cache) internFontFace(ff font.Face) uint16 {
	if c.fontInfoMap == nil {
		c.fontInfoMap = make(map[font.Face]uint16, 8)
	}

	if id, ok := c.fontInfoMap[ff]; ok {
		return id
	}

	id := uint16(len(c.fontInfoList))
	c.fontInfoMap[ff] = id

	m := ff.Metrics()
	capHeight := math.Abs(float64(m.CapHeight.Floor()))
	lineHeight := float64(m.Height.Floor())
	c.fontInfoList = append(c.fontInfoList, fontInfo{
		ff:         ff,
		capHeight:  capHeight,
		lineHeight: lineHeight,
	})

	return id
}
