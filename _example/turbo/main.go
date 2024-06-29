package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	graphics "github.com/quasilyte/ebitengine-graphics"
	"github.com/quasilyte/gmath"
)

func main() {
	g := &game{}
	g.init()

	ebiten.SetFullscreen(true)
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
	psys   *graphics.ParticleSystem
	player *player
}

type player struct {
	img      *ebiten.Image
	rotation gmath.Rad
	pos      gmath.Vec
	emitter  *graphics.ParticleEmitter
}

func (g *game) init() {
	g.psys = graphics.NewParticleSystem()
	g.psys.SetParticleSpeed(64.0, 96.0)
	g.psys.SetEmitInterval(0.05)
	// g.psys.SetEmitBurst(100, 200)
	g.psys.SetEmitBurst(1, 4)
	g.psys.SetParticleLifetime(1, 1)
	g.psys.SetParticleDirection(math.Pi, 0.2)

	triangle := ebiten.NewImage(32, 32)
	{
		clr := color.NRGBA{R: 40, G: 200, B: 150, A: 255}
		vector.StrokeLine(triangle, 1, 1, 32, 16, 2, clr, false)
		vector.StrokeLine(triangle, 32, 16, 1, 32, 2, clr, false)
		vector.StrokeLine(triangle, 1, 32, 1, 1, 2, clr, false)
	}
	emitter := g.psys.NewEmitter()
	g.player = &player{
		img:     triangle,
		emitter: emitter,
		pos:     gmath.Vec{X: 256, Y: 256},
	}
	emitter.Pos.Base = &g.player.pos
	emitter.PivotOffset.X = -12
	emitter.Rotation = &g.player.rotation
}

func (g *game) Draw(screen *ebiten.Image) {
	g.psys.Draw(screen)

	var opts ebiten.DrawImageOptions
	{
		opts.GeoM.Translate(-16, -16)
		opts.GeoM.Rotate(float64(g.player.rotation))
	}
	opts.GeoM.Translate(g.player.pos.X, g.player.pos.Y)
	screen.DrawImage(g.player.img, &opts)
	// for _, e := range g.emitters {
	// 	ebitenutil.DebugPrintAt(screen, "x", int(e.Pos.Offset.X)-2, int(e.Pos.Offset.Y)-8)
	// 	e.Draw(screen)
	// }
}

func (g *game) Layout(int, int) (int, int) {
	return 1920 / 2, 1080 / 2
}

func (g *game) Update() error {
	g.player.emitter.Update()

	// Handle the movement.
	moving := false
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.rotation += 2.0 * (1.0 / 60.0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.rotation -= 2.0 * (1.0 / 60.0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		moving = true
		g.player.pos = g.player.pos.MoveInDirection(96*(1.0/60.0), g.player.rotation)
	}
	g.player.emitter.SetEmitting(moving)

	// if !offset.IsZero() {
	// 	for _, e := range g.emitters {
	// 		e.Pos.Offset = e.Pos.Offset.Add(offset.Mulf(1.0 / 60.0))
	// 	}
	// }

	return nil
}