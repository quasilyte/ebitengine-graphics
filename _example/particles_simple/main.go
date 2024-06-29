package main

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	graphics "github.com/quasilyte/ebitengine-graphics"
	"github.com/quasilyte/gmath"
)

func main() {
	g := &game{}
	g.createSystems()

	// ebiten.SetFullscreen(true)
	ebiten.SetWindowSize(1920/2, 1080/2)

	// {
	// 	f, err := os.Create("cpu.out")
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	pprof.StartCPUProfile(f)
	// 	defer pprof.StopCPUProfile()
	// }
	// runtime.MemProfileRate = 1
	// defer func() {
	// 	f, err := os.Create("mem.out")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	runtime.GC()
	// 	pprof.WriteHeapProfile(f)
	// 	f.Close()
	// }()

	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}
}

type game struct {
	psys     *graphics.ParticleSystem
	emitters []*graphics.ParticleEmitter
}

func (g *game) createSystems() {
	g.psys = graphics.NewParticleSystem()
	g.psys.SetParticleSpeed(128.0, 256.0)
	g.psys.SetEmitInterval(1)
	// g.psys.SetEmitBurst(100, 200)
	g.psys.SetEmitBurst(140, 150)
	g.psys.SetParticleLifetime(2, 2)
	g.psys.SetParticleDirection(0, 2*math.Pi)
}

func (g *game) createEmitterAt(x, y int) {
	e := g.psys.NewEmitter()
	e.Pos.Offset = gmath.Vec{
		X: float64(x),
		Y: float64(y),
	}
	g.emitters = append(g.emitters, e)
}

func (g *game) Draw(screen *ebiten.Image) {
	numParticles := 0
	for _, e := range g.emitters {
		numParticles += e.NumParticles()
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS: %v, Particles: %d Emitters: %d", math.Round(ebiten.ActualFPS()), numParticles, len(g.emitters)))

	g.psys.Draw(screen)
	// for _, e := range g.emitters {
	// 	ebitenutil.DebugPrintAt(screen, "x", int(e.Pos.Offset.X)-2, int(e.Pos.Offset.Y)-8)
	// 	e.Draw(screen)
	// }
}

func (g *game) Layout(int, int) (int, int) {
	return 1920 / 2, 1080 / 2
}

func (g *game) Update() error {
	numParticles := 0
	for _, e := range g.emitters {
		e.UpdateWithDelta(1.0 / 60.0)
		numParticles += e.NumParticles()
		// e.Pos.Offset.Y += 8 * (1.0 / 60.0)
	}
	// if numParticles >= 40000 {
	// 	return ebiten.Termination
	// }

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		g.createEmitterAt(x, y)
	}

	// Handle the movement.
	var offset gmath.Vec
	speed := 96.0
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		offset.X += speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		offset.X -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		offset.Y += speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		offset.Y -= speed
	}
	if !offset.IsZero() {
		for _, e := range g.emitters {
			e.Pos.Offset = e.Pos.Offset.Add(offset.Mulf(1.0 / 60.0))
		}
	}

	return nil
}
