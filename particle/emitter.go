package particle

import (
	"math"

	"github.com/quasilyte/ebitengine-graphics/internal/cache"
	"github.com/quasilyte/gmath"
)

type Emitter struct {
	sys *Renderer

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

	emitting bool
	visible  bool
}

type particle struct {
	// This is a packed countdown value that specifies the particle progression
	// and its remaining lifetime.
	// We're not using float32 here to save 16 bits.
	// Instead, every countdown value is equal to 0.001 seconds (millisecond);
	// therefore, the max value representable is ~65 seconds.
	// It should be enough for 99.(9)% use cases.
	countdown uint16

	speedSeed uint8
	angleSeed uint8
	origAngle uint8 // 256 possible values: 2*PI/256

	// Would use {uint16, uint16} here to save 4 bytes,
	// but it can be desirable to support negative coords
	// and values larget than MaxUint16.
	// The 32-bit float will grant us a precision area of about
	// 8_388_608 pixels (as opposed to 65_536 of uint16).
	// A world larger than that would make particles less accurate.
	origPos gmath.Vec32
}

func newParticleEmitter(sys *Renderer) *Emitter {
	return &Emitter{
		sys:       sys,
		particles: make([]particle, 0, 8),
		visible:   true,
	}
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
			t += e.sys.emitInterval
			e.emitDelay += e.sys.emitInterval
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
	for _, p := range e.particles {
		if dt > p.countdown {
			// An underflow would occur.
			// This particle is expired.
			continue
		}
		p.countdown -= dt
		live = append(live, p)
	}
	e.particles = live
}

func (e *Emitter) emit(t float32) {
	pos := e.Pos.Resolve()
	if !e.PivotOffset.IsZero() {
		offset := e.PivotOffset
		if e.Rotation != nil {
			offset = offset.Rotated(*e.Rotation)
		}
		pos = pos.Add(offset)
	}

	// Compute up to 64 random bits only once.
	// Then use the fastrand to generate more.
	// If system doesn't need any rand bits, don't bother generating it.
	randBits := uint64(0)
	randSeq := uint64(0)
	if e.sys.needsRandBits != 0 {
		randBits = cache.Global.Rand.Uint64()
	}

	numParticles := 1
	if e.sys.emitBurstRangeSize != 0 {
		// Calculate the range using the 8 rand bits.
		x := uint16(fastrand(randBits, randSeq) & 0xff)
		randSeq++
		numParticles = int(e.sys.minEmitBurst + uint8(x*e.sys.emitBurstRangeSize/256))
	}

	maxCountdown := uint16((e.sys.particleMaxLifetime - t) * 1000)
	for i := 0; i < numParticles; i++ {
		countdown := maxCountdown
		if e.sys.needsRandBits&lifetimeRandBit != 0 {
			x := uint8(fastrand(randBits, randSeq))
			randSeq++
			countdown -= uint16(e.sys.particleLifetimeStep * float32(x) * 1000)
		}

		origAngle := uint8(0)
		if e.Rotation != nil {
			origAngle = uint8((e.Rotation.Normalized() / (2 * math.Pi)) * 256)
		}

		p := particle{
			countdown: countdown,
			origAngle: origAngle,
			origPos: gmath.Vec32{
				X: float32(pos.X),
				Y: float32(pos.Y),
			},
		}

		if e.sys.needsRandBits&speedRandBit != 0 {
			x := uint8(fastrand(randBits, randSeq))
			randSeq++
			p.speedSeed = x
		}

		if e.sys.needsRandBits&angleRandBit != 0 {
			x := uint8(fastrand(randBits, randSeq))
			randSeq++
			p.angleSeed = x
		}

		e.particles = append(e.particles, p)
	}
}
