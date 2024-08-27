package particle

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	graphics "github.com/quasilyte/ebitengine-graphics"
	"github.com/quasilyte/ebitengine-graphics/internal/cache"
	"github.com/quasilyte/gmath"
)

const (
	speedRandBit = 1 << iota
	burstRandBit
	angleRandBit
	lifetimeRandBit
	scalingRandBit
)

var defaultPalette = []graphics.ColorScale{
	{R: 1, G: 1, B: 1, A: 1},
}

type SpawnContext struct {
	id       uint32
	userData uint8
	emitter  *Emitter
}

func (ctx *SpawnContext) ParticleUserData() uint8 {
	return ctx.userData
}

func (ctx *SpawnContext) ParticleID() int {
	return int(ctx.emitter.idSeq)
}

func (ctx *SpawnContext) Generation() int {
	return int(ctx.emitter.generation)
}

// Rand returns a pseudo-random value in [0, 1) range.
//
// It's usually faster than stdlib random, but its
// randomness may be inferior. It should be good enough
// for the purposes of generating varying particles.
func (ctx *SpawnContext) Rand() float64 {
	return fastrandFloat(randseed1, uint64(ctx.id))
}

func (ctx *SpawnContext) RandUint() uint64 {
	return fastrand(randseed1, uint64(ctx.id))
}

type UpdateContext struct {
	emitter  *Emitter
	t        float32
	userData uint8
}

func (ctx *UpdateContext) ParticleUserData() uint8 {
	return ctx.userData
}

func (ctx *UpdateContext) Time() float32 {
	return ctx.t
}

type Template struct {
	img *ebiten.Image

	particleMinScaling  gmath.Vec32
	particleMaxScaling  gmath.Vec32
	particleScalingStep gmath.Vec32

	particleMinAngle  float64
	particleMaxAngle  float64
	particleAngleStep float64

	particleMinSpeed  float32
	particleMaxSpeed  float32
	particleSpeedStep float32

	emitInterval        float32
	particleMinLifetime float32
	particleMaxLifetime float32

	minEmitBurst       uint8
	maxEmitBurst       uint8
	emitBurstRangeSize uint16

	needsRandBits uint8

	palette []graphics.ColorScale

	spawnUserDataFunc func(ctx SpawnContext) uint8
	spawnOffsetFunc   func(ctx SpawnContext) gmath.Vec
	spawnColorFunc    func(ctx SpawnContext) uint

	updateColorScaleFunc func(ctx UpdateContext) graphics.ColorScale
	updateScalingFunc    func(ctx UpdateContext) gmath.Vec32
}

func NewTemplate() *Template {
	tmpl := &Template{
		palette: defaultPalette,
	}

	tmpl.SetParticleScaling(gmath.Vec{X: 1, Y: 1})
	tmpl.SetEmitInterval(0.5)
	tmpl.SetParticleLifetime(3)
	tmpl.SetParticleSpeed(32.0)
	tmpl.SetEmitBurst(1, 1)
	tmpl.SetImage(cache.Global.WhitePixel)

	return tmpl
}

func (tmpl *Template) Clone() *Template {
	cloned := *tmpl
	if len(tmpl.palette) != 0 {
		cloned.palette = make([]graphics.ColorScale, len(tmpl.palette))
		copy(cloned.palette, tmpl.palette)
	}
	return &cloned
}

func (tmpl *Template) SetPalette(colors []graphics.ColorScale) {
	tmpl.palette = colors
}

func (tmpl *Template) SetSpawnUserDataFunc(fn func(ctx SpawnContext) uint8) {
	tmpl.spawnUserDataFunc = fn
}

func (tmpl *Template) SetSpawnOffsetFunc(fn func(ctx SpawnContext) gmath.Vec) {
	tmpl.spawnOffsetFunc = fn
}

func (tmpl *Template) SetSpawnColorFunc(fn func(ctx SpawnContext) uint) {
	tmpl.spawnColorFunc = fn
}

func (tmpl *Template) SetUpdateColorScaleFunc(fn func(ctx UpdateContext) graphics.ColorScale) {
	tmpl.updateColorScaleFunc = fn
}

func (tmpl *Template) SetUpdateScalingFunc(fn func(ctx UpdateContext) gmath.Vec32) {
	tmpl.updateScalingFunc = fn
}

func (tmpl *Template) SetImage(img *ebiten.Image) {
	tmpl.img = img
}

func (tmpl *Template) SetEmitBurst(minAmount, maxAmount int) {
	amount := gmath.MakeRange(minAmount, maxAmount)
	if !amount.InBounds(0, math.MaxUint8) {
		panic("amount is not in [0, 255] bounds")
	}
	if amount.Min != 1 || amount.Max != 1 {
		tmpl.emitBurstRangeSize = uint16(amount.Max) - uint16(amount.Min) + 1
		tmpl.needsRandBits |= burstRandBit
	} else {
		tmpl.emitBurstRangeSize = 0
		tmpl.needsRandBits &^= burstRandBit
	}
	tmpl.minEmitBurst = uint8(amount.Min)
	tmpl.maxEmitBurst = uint8(amount.Max)
}

func (tmpl *Template) SetParticleScaling(scaling gmath.Vec) {
	tmpl.SetParticleScalingRange(scaling, scaling)
}

func (tmpl *Template) SetParticleScalingRange(minScaling, maxScaling gmath.Vec) {
	isValid := gmath.MakeRange(minScaling.X, maxScaling.X).IsValid() &&
		gmath.MakeRange(minScaling.Y, maxScaling.Y).IsValid()
	if !isValid {
		panic("invalid scaling range")
	}

	if minScaling != maxScaling {
		tmpl.needsRandBits |= scalingRandBit
	} else {
		tmpl.needsRandBits &^= scalingRandBit
	}

	tmpl.particleMinScaling = minScaling.AsVec32()
	tmpl.particleMaxScaling = maxScaling.AsVec32()
	tmpl.particleScalingStep = tmpl.particleMaxScaling.Sub(tmpl.particleMinScaling).Divf(255)
}

func (tmpl *Template) SetParticleLifetime(lifetime float64) {
	tmpl.SetParticleLifetimeRange(lifetime, lifetime)
}

func (tmpl *Template) SetParticleLifetimeRange(minLifetime, maxLifetime float64) {
	lifetime := gmath.MakeRange(minLifetime, maxLifetime)
	if !lifetime.IsValid() {
		panic("invalid lifetime range")
	}
	// A sanity check to avoid unexpected behaviors.
	if int(maxLifetime*1000) > math.MaxUint16 {
		panic("specified maxLifetime oveflows the supported lifetime limit of ~65s")
	}

	if lifetime.Min != lifetime.Max {
		tmpl.needsRandBits |= lifetimeRandBit
	} else {
		tmpl.needsRandBits &^= lifetimeRandBit
	}

	tmpl.particleMinLifetime = float32(lifetime.Min)
	tmpl.particleMaxLifetime = float32(lifetime.Max)

	// To avoid call-order dependency, update the particle speed
	// and velocity multiplier when changing the lifetime.
	tmpl.adjustParticleSpeed(tmpl.particleMinSpeed, tmpl.particleMaxSpeed)
}

func (tmpl *Template) SetParticleDirection(dir, spread gmath.Rad) {
	angle := gmath.Range[float64]{
		Min: float64(dir - spread*0.5),
		Max: float64(dir + spread*0.5),
	}

	tmpl.particleMinAngle = angle.Min
	tmpl.particleMaxAngle = angle.Max

	if angle.Min != angle.Max {
		tmpl.particleAngleStep = (angle.Max - angle.Min) / 255
		tmpl.needsRandBits |= angleRandBit
	} else {
		tmpl.particleAngleStep = 0
		tmpl.needsRandBits &^= angleRandBit
	}
}

func (tmpl *Template) SetParticleSpeed(speed float64) {
	tmpl.SetParticleSpeedRange(speed, speed)
}

func (tmpl *Template) SetParticleSpeedRange(minSpeed, maxSpeed float64) {
	speed := gmath.MakeRange(minSpeed, maxSpeed)
	if speed.Max < speed.Min {
		panic("maxSpeed can't be less than speed.Min")
	}
	tmpl.adjustParticleSpeed(float32(speed.Min), float32(speed.Max))
}

func (tmpl *Template) adjustParticleSpeed(minSpeed, maxSpeed float32) {
	tmpl.particleMinSpeed = minSpeed
	tmpl.particleMaxSpeed = maxSpeed

	if minSpeed != maxSpeed {
		tmpl.particleSpeedStep = (maxSpeed - minSpeed) / 255
		tmpl.needsRandBits |= speedRandBit
	} else {
		tmpl.particleSpeedStep = 0
		tmpl.needsRandBits &^= speedRandBit
	}
}

func (tmpl *Template) SetEmitInterval(t float64) {
	tmpl.emitInterval = float32(t)
}
