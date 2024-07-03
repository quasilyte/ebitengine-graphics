package particle

import (
	"math"

	"github.com/quasilyte/ebitengine-graphics/internal/cache"
	"github.com/quasilyte/gmath"
)

type Emitter struct {
	tmpl *Template

	particles []particle

	Pos gmath.Pos

	PivotOffset gmath.Vec

	Rotation *gmath.Rad

	// dtError is a fractional part that accumulates between the updates.
	// When it becomes 1.0, it will be consumed to correct the s->ms conversion
	// we do (the error happens due to the uint16 truncation).
	// It helps to avoid the infinitely growing error.
	dtError float64

	// emitDelay is a time (in seconds) until the next emission step.
	emitDelay float32

	idSeq      uint32
	generation uint16

	emitting bool
	visible  bool
	disposed bool
}

type particle struct {
	counter uint16

	scalingSeed  uint8
	speedSeed    uint8
	angleSeed    uint8
	lifetimeSeed uint8
	origAngle    uint8
	paletteIndex uint8

	// Would use {uint16, uint16} here to save 4 bytes,
	// but it can be desirable to support negative coords
	// and values larget than MaxUint16.
	// The 32-bit float will grant us a precision area of about
	// 8_388_608 pixels (as opposed to 65_536 of uint16).
	// A world larger than that would make particles less accurate.
	origPos gmath.Vec32
}

func NewEmitter(tmpl *Template) *Emitter {
	e := &Emitter{
		tmpl:      tmpl,
		particles: make([]particle, 0, 8),
		visible:   true,
	}
	return e
}

func (e *Emitter) IsDisposed() bool {
	return e.disposed
}

func (e *Emitter) Dispose() {
	e.disposed = true
}

func (e *Emitter) SetVisibility(visible bool) { e.visible = visible }

func (e *Emitter) SetEmitting(emitting bool) {
	if e.emitting == emitting {
		return
	}

	e.emitDelay = 0
	e.emitting = emitting
}

func (e *Emitter) NumParticles() int {
	return len(e.particles)
}

func (e *Emitter) Update() {
	e.UpdateWithDelta(1.0 / 60.0)
}

func (e *Emitter) UpdateWithDelta(delta float64) {
	if e.emitting {
		// With a very low emit interval, it's possible to have
		// delta > emitInterval.
		// If delta is 0.1 and emitInterval is 0.04, then we need
		// to emit twice and have the delay set to 0.02.
		// We should also create the second particles emitted
		// with t=0.04 instead of t=0.00.
		e.emitDelay -= float32(delta)
		t := float32(0.0)
		for e.emitDelay < 0 {
			e.emit(t)
			t += e.tmpl.emitInterval
			e.emitDelay += e.tmpl.emitInterval
		}
	}

	deltaMS := delta * 1000
	// This loop should execute either 0 or 1 time in most cases,
	// since error gets only a fractional part per tick.
	// But just to be safe, make it a loop.
	// Should still be more efficient than computing another modf.
	for e.dtError >= 1 {
		deltaMS++
		e.dtError--
	}
	_, fract := math.Modf(deltaMS)
	e.dtError += fract

	live := e.particles[:0]
	dt := uint16(deltaMS)
	maxLifetime := uint16(1000 * e.tmpl.particleMaxLifetime)
	lifetimeStep := e.tmpl.particleLifetimeStep
	for _, p := range e.particles {
		p.counter += dt
		lifetime := maxLifetime - (uint16(p.lifetimeSeed) * lifetimeStep)
		if p.counter > lifetime {
			continue
		}
		live = append(live, p)
	}
	e.particles = live
}

func (e *Emitter) emit(t float32) {
	tmpl := e.tmpl
	e.generation++

	w, h := float64(e.tmpl.img.Bounds().Dx()), float64(e.tmpl.img.Bounds().Dy())
	pos := e.Pos.Resolve().Sub(gmath.Vec{X: w * 0.5, Y: h * 0.5})
	if !e.PivotOffset.IsZero() {
		offset := rotatedVec(e.PivotOffset, e.Rotation)
		pos = pos.Add(offset)
	}

	// Compute up to 64 random bits only once.
	// Then use the fastrand to generate more.
	// If system doesn't need any rand bits, don't bother generating it.
	randBits := uint64(0)
	randSeq := uint64(0)
	if e.tmpl.needsRandBits != 0 {
		randBits = cache.Global.Rand.Uint64()
	}

	numParticles := 1
	if e.tmpl.emitBurstRangeSize != 0 {
		// Calculate the range using the 8 rand bits.
		x := uint16(fastrand(randBits, randSeq) & 0xff)
		randSeq++
		numParticles = int(e.tmpl.minEmitBurst + uint8(x*e.tmpl.emitBurstRangeSize/256))
	}

	ctx := SpawnContext{emitter: e}
	for i := 0; i < numParticles; i++ {
		ctx.id = e.idSeq
		e.idSeq++
		particlePos := pos
		if tmpl.spawnOffsetFunc != nil {
			offset := rotatedVec(tmpl.spawnOffsetFunc(ctx), e.Rotation)
			particlePos = particlePos.Add(offset)
		}

		paletteIndex := uint8(0)
		if tmpl.spawnColorFunc != nil {
			paletteIndex = uint8(tmpl.spawnColorFunc(ctx))
		}

		lifetimeSeed := uint8(0)
		if e.tmpl.needsRandBits&lifetimeRandBit != 0 {
			lifetimeSeed = uint8(fastrand(randBits, randSeq))
			randSeq++
		}

		scalingSeed := uint8(0)
		if e.tmpl.needsRandBits&lifetimeRandBit != 0 {
			scalingSeed = uint8(fastrand(randBits, randSeq))
			randSeq++
		}

		origAngle := uint8(0)
		if e.Rotation != nil {
			origAngle = uint8((e.Rotation.Normalized() / (2 * math.Pi)) * 256)
		}

		p := particle{
			counter:      uint16(t * 1000),
			paletteIndex: paletteIndex,
			lifetimeSeed: lifetimeSeed,
			scalingSeed:  scalingSeed,
			origAngle:    origAngle,
			origPos:      vec32(particlePos),
		}

		if e.tmpl.needsRandBits&speedRandBit != 0 {
			x := uint8(fastrand(randBits, randSeq))
			randSeq++
			p.speedSeed = x
		}

		if e.tmpl.needsRandBits&angleRandBit != 0 {
			x := uint8(fastrand(randBits, randSeq))
			randSeq++
			p.angleSeed = x
		}

		e.particles = append(e.particles, p)
	}
}
