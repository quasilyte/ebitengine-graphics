package graphics

import (
	_ "embed"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/ebitengine-graphics/internal/cache"
)

var (
	//go:embed _shaders/circle.go
	shaderCircleOutline []byte

	//go:embed _shaders/dashed_circle.go
	shaderDashedCircleOutline []byte
)

// CompileShaders prepares shaders bundled with this package.
// It should be called once if any of the shader-dependend objects
// are being used.
//
// Objects that require shaders so far:
// * Circle
func CompileShaders() {
	if cache.Global.ShadersCompiled {
		return
	}
	cache.Global.ShadersCompiled = true

	mustCompileShader := func(src []byte) *ebiten.Shader {
		s, err := ebiten.NewShader(src)
		if err != nil {
			panic(err)
		}
		return s
	}

	cache.Global.CircleOutlineShader = mustCompileShader(shaderCircleOutline)
	cache.Global.DashedCircleOutlineShader = mustCompileShader(shaderDashedCircleOutline)
}

func requireShaders() {
	if cache.Global.ShadersCompiled {
		return
	}
	panic("call graphics.CompileShaders once to use this function")
}
