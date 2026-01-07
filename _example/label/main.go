package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/quasilyte/bitsweetfont"
	graphics "github.com/quasilyte/ebitengine-graphics"
)

func main() {
	g := &game{}

	ebiten.SetWindowSize(1920/4, 1080/4)

	fontface := text.NewGoXFace(bitsweetfont.Scale(bitsweetfont.New1_3(), 2))
	g.l = graphics.NewLabel(fontface)
	g.l.SetSize(1920/4, 1080/4)
	g.l.SetAlignHorizontal(graphics.AlignHorizontalCenter)
	g.l.SetAlignVertical(graphics.AlignVertical(graphics.AlignHorizontalCenter))
	g.l.SetText("Hello, World!")

	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}
}

type game struct {
	l *graphics.Label
}

func (g *game) Draw(screen *ebiten.Image) {
	g.l.Draw(screen)
}

func (g *game) Layout(int, int) (int, int) {
	return 1920 / 4, 1080 / 4
}

func (g *game) Update() error {
	return nil
}
