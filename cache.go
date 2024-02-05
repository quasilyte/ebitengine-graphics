package graphics

import (
	"math"

	"golang.org/x/image/font"
)

// Cache is a storage that is shared between all
// graphical elements.
//
// Usually, there should be only 1 graphical cache per app.
type Cache struct {
	fontInfoList []fontInfo
	fontInfoMap  map[font.Face]uint16
}

type fontInfo struct {
	ff         font.Face
	capHeight  float64
	lineHeight float64
}

func NewCache() *Cache {
	return &Cache{}
}

func (c *Cache) internFontFace(ff font.Face) uint16 {
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
