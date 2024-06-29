package particle

import (
	"math"

	"github.com/quasilyte/gmath"
)

type SpawnContext struct {
	generation uint16
}

func (ctx *SpawnContext) Generation() int {
	return int(ctx.generation)
}

func (ctx *SpawnContext) Rand() float64 {
	v := fastrand(randseed1, 10000*uint64(ctx.generation))
	return float64(v) / math.MaxUint64
}

type Template struct {
	spawnPosFunc func(ctx *SpawnContext) gmath.Vec
}

func NewTemplate() *Template {
	return &Template{}
}

func (tmpl *Template) SetSpawnPosFunc(fn func(ctx *SpawnContext) gmath.Vec) {
	tmpl.spawnPosFunc = fn
}
