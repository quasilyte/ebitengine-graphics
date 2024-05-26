package graphics

import (
	_ "embed"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	//go:embed _shaders/circle_outline.go
	shaderCircleOutline []byte

	//go:embed _shaders/dashed_circle_outline.go
	shaderDashedCircleOutline []byte
)

// CompileShaders prepares shaders bundled with this package.
// It should be called once if any of the shader-dependend objects
// are being used.
//
// Objects that require shaders so far:
// * Circle
func CompileShaders() {
	if globalCache.shadersCompiled {
		return
	}
	globalCache.shadersCompiled = true

	mustCompileShader := func(src []byte) *ebiten.Shader {
		s, err := ebiten.NewShader(src)
		if err != nil {
			panic(err)
		}
		return s
	}

	globalCache.circleOutlineShader = mustCompileShader(shaderCircleOutline)
	globalCache.dashedCircleOutlineShader = mustCompileShader(shaderDashedCircleOutline)
}

func requireShaders() {
	if globalCache.shadersCompiled {
		return
	}
	panic("call graphics.CompileShaders once to use this function")
}
