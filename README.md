# ebitengine-graphics

![Build Status](https://github.com/quasilyte/ebitengine-graphics/workflows/Go/badge.svg)
[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/quasilyte/ebitengine-graphics)](https://pkg.go.dev/mod/github.com/quasilyte/ebitengine-graphics)

## Overview

A package implementing various graphics primitives like Sprite for [Ebitengine](https://github.com/hajimehoshi/ebiten/).

It works the best in combination with [gscene package](https://github.com/quasilyte/gscene), but can be used without it.

Graphical objects list:

* Sprite
* Texture Line
* Line
* Circle (supports dashed style)
* Rect
* Label
* Container
* Canvas

> Missing some graphical object? [Tell us about it](https://github.com/quasilyte/ebitengine-graphics/issues/new).

It also supports a basic camera and layers implementation.

## Installation

```bash
go get github.com/quasilyte/ebitengine-graphics
```

## Quick Start

The most useful type of this package is `Sprite`, but it's tricky to demonstrate its usage without assets. This is a quick start section, it's intended to be as slim as possible.

You can use this library without [gscene](https://github.com/quasilyte/gscene):

```go
package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	graphics "github.com/quasilyte/ebitengine-graphics"
	"github.com/quasilyte/gmath"
)

func main() {
	ebiten.SetWindowSize(640, 480)
	if err := ebiten.RunGame(newExampleGame()); err != nil {
		panic(err)
	}
}

type exampleGame struct {
	pos           gmath.Vec
	initialized   bool
	objects       []drawable
}

type drawable interface {
	Draw(screen *ebiten.Image)
}

func newExampleGame() *exampleGame {
	return &exampleGame{
		pos:           gmath.Vec{X: 32, Y: 32},
	}
}

func (g *exampleGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 640, 480
}

func (g *exampleGame) Draw(screen *ebiten.Image) {
	for _, o := range g.objects {
		o.Draw(screen)
	}
}

func (g *exampleGame) Update() error {
	if !g.initialized {
		g.Init()
		g.initialized = true
	}

	g.pos = g.pos.Add(gmath.Vec{X: 1, Y: 2})
	return nil
}

func (g *exampleGame) Init() {
	{
		from := gmath.Pos{Base: &g.pos}
		to := gmath.Pos{Offset: gmath.Vec{X: 128, Y: 64}}
		l := graphics.NewLine(from, to)
		l.SetWidth(2)
		l.SetColorScale(graphics.ColorScaleFromRGBA(200, 100, 100, 255))
		g.objects = append(g.objects, l)
	}

	{
		r := graphics.NewRect(32, 32)
		r.Pos.Base = &g.pos
		r.SetFillColorScale(graphics.RGB(0xAABB00))
		r.SetOutlineColorScale(graphics.RGB(0x0055ff))
		r.SetOutlineWidth(2)
		g.objects = append(g.objects, r)
	}
}
```

With `gscene` it's even easier (showing only relevant part):

```go
func (c *exampleController) Init(scene *gscene.SimpleRootScene) {
	{
		from := gmath.Pos{Base: &g.pos}
		to := gmath.Pos{Offset: gmath.Vec{X: 128, Y: 64}}
		l := graphics.NewLine(from, to)
		l.SetWidth(2)
		l.SetColorScale(graphics.ColorScaleFromRGBA(200, 100, 100, 255))
		scene.AddGraphics(l)
	}

	{
		r := graphics.NewRect(32, 32)
		r.Pos.Base = &g.pos
		r.SetFillColorScale(graphics.RGB(0xAABB00))
		r.SetOutlineColorScale(graphics.RGB(0x0055ff))
		r.SetOutlineWidth(2)
		scene.AddGraphics(r)
	}
}
```

Note that we can add the graphical object directly to the scene. The scene will manage their `Draw` calls as well as their lifetimes (based on graphical objects being disposed or not).
