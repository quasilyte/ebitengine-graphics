package graphics

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/gmath"
)

// randomized stuff:
// + spawn pos, emission shape (doesn't need a pool)
// + lifetime randomness 0..1 (doesn't need a pool)
// randomized stuff that needs a pool of sorts:
// - speed spread (maybe 256 "steps" are enough?)
// - direction spread (maybe 256 values plus some offset are enough?)
// - rotation speed (?)
// - hue randomization
//
// need configurators for:
// - initial angle
// - angle transformation
// - scale (+ curve)
// - color scaling
// - "explosiveness"
//
// unsure:
// - orbit velocity
// - linear acceleration (or radial, etc.)
// - damping (slowdown)
// - radial angle (?)

// ParticleSystem describes a CPU particle emission template.
// You can create multiple [ParticleEmitter] from this template.
//
// Usually, for each emission effect kind you create a separate system.
// Then you create [ParticleEmitter] per object that needs that system.
// For more complicated cases when high customization of every emitter is required,
// you might end up with almost 1-to-1 system and emitter ratio.
// This should not be a problem.
//
// This system is not designed to be used for thousands of particles.
// It's should be good enough for some indie games and can serve as an example
// CPU particle system implementation.
//
// Experimental: particles are part of the experimental API and are subject to change.
// You're encouraged to give feedback, but it might not be a good idea
// to use it in your serious long-term project just yet.
type ParticleSystem struct {
	img *ebiten.Image

	particleMinAngle  gmath.Rad
	particleMaxAngle  gmath.Rad
	particleAngleStep gmath.Rad

	particleMinSpeed  float32
	particleMaxSpeed  float32
	particleSpeedStep float32
	particleSpeed     float32

	emitInterval          float32
	particleMaxLifetime   float32
	particleMaxLifetimeMS float32
	particleLifetimeStep  float32

	minEmitBurst       uint8
	maxEmitBurst       uint8
	emitBurstRangeSize uint16

	needsRandBits uint8

	emitters []*ParticleEmitter
}

const (
	speedRandBit = 1 << iota
	burstRandBit
	angleRandBit
	lifetimeRandBit
)

func NewParticleSystem() *ParticleSystem {
	psys := &ParticleSystem{}

	psys.SetEmitInterval(0.5)
	psys.SetParticleLifetime(3, 3)
	psys.SetParticleSpeed(32.0, 32.0)
	psys.SetEmitBurst(1, 1)
	psys.SetImage(globalCache.whitePixel)

	return psys
}

func (sys *ParticleSystem) SetImage(img *ebiten.Image) {
	sys.img = img
}

func (sys *ParticleSystem) SetEmitBurst(minAmount, maxAmount int) {
	amount := gmath.MakeRange(minAmount, maxAmount)
	if !amount.InBounds(0, math.MaxUint8) {
		panic("amount is not in [0, 255] bounds")
	}
	if amount.Min != 1 || amount.Max != 1 {
		sys.emitBurstRangeSize = uint16(amount.Max) - uint16(amount.Min) + 1
		sys.needsRandBits |= burstRandBit
	} else {
		sys.emitBurstRangeSize = 0
		sys.needsRandBits &^= burstRandBit
	}
	sys.minEmitBurst = uint8(amount.Min)
	sys.maxEmitBurst = uint8(amount.Max)
}

func (sys *ParticleSystem) SetParticleLifetime(minLifetime, maxLifetime float64) {
	lifetime := gmath.MakeRange(minLifetime, maxLifetime)
	if !lifetime.IsValid() {
		panic("invalid lifetime range")
	}
	// A sanity check to avoid unexpected behaviors.
	if int(maxLifetime*1000) > math.MaxUint16 {
		panic("specified maxLifetime oveflows the supported lifetime limit of ~65s")
	}

	if lifetime.Min != lifetime.Max {
		sys.needsRandBits |= lifetimeRandBit
	} else {
		sys.needsRandBits &^= lifetimeRandBit
	}

	sys.particleMaxLifetime = float32(lifetime.Max)
	sys.particleMaxLifetimeMS = sys.particleMaxLifetime * 1000
	sys.particleLifetimeStep = float32((lifetime.Max - lifetime.Min) / 255)

	// To avoid call-order dependency, update the particle speed
	// and velocity multiplier when changing the lifetime.
	sys.adjustParticleSpeed(sys.particleMinSpeed, sys.particleMaxSpeed)
}

func (sys *ParticleSystem) SetParticleDirection(dir, spread gmath.Rad) {
	angle := gmath.Range[gmath.Rad]{
		Min: dir - spread*0.5,
		Max: dir + spread*0.5,
	}

	sys.particleMinAngle = angle.Min
	sys.particleMaxAngle = angle.Max

	if angle.Min != angle.Max {
		sys.particleAngleStep = (angle.Max - angle.Min) / 255
		sys.needsRandBits |= angleRandBit
	} else {
		sys.particleAngleStep = 0
		sys.needsRandBits &^= angleRandBit
	}
}

func (sys *ParticleSystem) SetParticleSpeed(minSpeed, maxSpeed float64) {
	speed := gmath.MakeRange(minSpeed, maxSpeed)
	if speed.Max < speed.Min {
		panic("maxSpeed can't be less than speed.Min")
	}
	sys.adjustParticleSpeed(float32(speed.Min), float32(speed.Max))
}

func (sys *ParticleSystem) adjustParticleSpeed(minSpeed, maxSpeed float32) {
	sys.particleMinSpeed = minSpeed
	sys.particleMaxSpeed = maxSpeed
	sys.particleSpeed = sys.particleMaxLifetime * sys.particleMinSpeed

	if minSpeed != maxSpeed {
		sys.particleSpeedStep = (maxSpeed - minSpeed) / 255
		sys.needsRandBits |= speedRandBit
	} else {
		sys.particleSpeedStep = 0
		sys.needsRandBits &^= speedRandBit
	}
}

func (sys *ParticleSystem) SetEmitInterval(t float64) {
	sys.emitInterval = float32(t)
}

func (sys *ParticleSystem) NewEmitter() *ParticleEmitter {
	e := newParticleEmitter(sys)
	sys.emitters = append(sys.emitters, e)
	return e
}

func (sys *ParticleSystem) Draw(dst *ebiten.Image) {
	sys.DrawWithOptions(dst, DrawOptions{})
}

func (sys *ParticleSystem) drawBatch(dst *ebiten.Image, opts DrawOptions, emitters []*ParticleEmitter) {
	// Use pre-allocated slices.
	vertices := globalCache.scratchVertices[:0]
	indices := globalCache.scratchIndices[:0]
	defer func() {
		globalCache.scratchVertices = vertices[:0]
		globalCache.scratchIndices = indices[:0]
	}()

	img := sys.img

	w, h := float32(img.Bounds().Dx()), float32(img.Bounds().Dy())
	idx := uint16(0)
	for _, e := range emitters {
		for _, p := range e.particles {
			var pos ebiten.GeoM
			{
				origPos := p.origPos
				progress := float32(p.countdown) / e.sys.particleMaxLifetimeMS

				dir := gmath.Vec32{X: 1, Y: 0}
				if e.sys.needsRandBits&angleRandBit != 0 {
					angle := e.sys.particleMinAngle + e.sys.particleAngleStep*gmath.Rad(p.angleSeed)
					angle += gmath.Rad(p.origAngle) * ((2 * math.Pi) / 256)
					dir = dir.Rotated(angle)
				}

				// The currentPos might benefit from rounding, but since particle drawing
				// can be considered to be a hot path, I'm not sure we should do it.
				// Rounding thousands vectors can add up.
				speed := e.sys.particleSpeed + (e.sys.particleSpeedStep * float32(p.speedSeed))
				currentPos := origPos.Add(dir.Mulf(speed).Mulf(1 - progress))

				pos.Translate(opts.Offset.X, opts.Offset.Y)
				pos.Translate(float64(currentPos.X), float64(currentPos.Y))
			}

			x := float32(pos.Element(0, 2))
			y := float32(pos.Element(1, 2))

			vertices = append(vertices,
				ebiten.Vertex{DstX: x, DstY: y, SrcX: 0, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
				ebiten.Vertex{DstX: x + w, DstY: y, SrcX: w, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
				ebiten.Vertex{DstX: x + w, DstY: y + h, SrcX: w, SrcY: h, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
				ebiten.Vertex{DstX: x, DstY: y + h, SrcX: 0, SrcY: h, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
			)

			indices = append(indices, idx, idx+1, idx+2, idx, idx+2, idx+3)
			idx += 4
		}
	}

	var drawOptions ebiten.DrawTrianglesOptions
	dst.DrawTriangles(vertices, indices, img, &drawOptions)
}

func (sys *ParticleSystem) DrawWithOptions(dst *ebiten.Image, opts DrawOptions) {
	const batchThreshold = math.MaxUint16 / 24 // Doesn't have to be bigger
	batchParticles := 0

	batch := globalCache.scratchEmitterBatch[:0]
	defer func() {
		globalCache.scratchEmitterBatch = batch[:0]
	}()

	for _, e := range sys.emitters {
		if !e.visible {
			continue
		}
		n := e.NumParticles()
		if n == 0 {
			continue
		}
		if batchParticles+n > batchThreshold {
			sys.drawBatch(dst, opts, batch)
			batch = batch[:0]
			batchParticles = 0
		} else {
			batch = append(batch, e)
			batchParticles += n
		}
	}
	if len(batch) != 0 {
		sys.drawBatch(dst, opts, batch)
	}
}
