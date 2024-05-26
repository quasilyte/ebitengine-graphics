package graphics

import (
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

// Shader is an ebiten.Shader wrapper compatible with types provided
// by the graphics package.
//
// Shader allocates its uniforms map lazily.
type Shader struct {
	compiled *ebiten.Shader

	shaderData map[string]any

	// Enabled reports whether this shader will be used during the rendering.
	Enabled bool

	// Textures are used as ebiten.DrawRectShaderOptions.Images[i] values,
	// where i is the suffix of the field (Texture1 => 1, Texture2 => 2).
	// The graphics object that owns this Shader will be assigned to Images[0].

	Texture1 *ebiten.Image
	Texture2 *ebiten.Image
	Texture3 *ebiten.Image
}

// NewShader returns a shader wrapper.
// The returned shader has Enabled flag set to true.
//
// Use Shader object fields to configure it.
func NewShader(compiled *ebiten.Shader) *Shader {
	return &Shader{
		compiled: compiled,
		Enabled:  true,
	}
}

// Clone returns a cloned shader object.
// It's main purpose is to create a new shader handle
// that has the identical uniform values those
// can be modified independently.
func (s *Shader) Clone() *Shader {
	cloned := *s
	cloned.shaderData = make(map[string]any, len(s.shaderData))
	for k, v := range s.shaderData {
		cloned.shaderData[k] = v
	}
	return &cloned
}

// GetValue returns the current uniform value stored under the key.
func (s *Shader) GetValue(key string) any {
	return s.shaderData[key]
}

// SetVec2Value assigns vec2 Kage uniform variable.
// v must be a 2-element slice, otherwise this method will panic.
func (s *Shader) SetVec2Value(key string, v []float32) {
	if len(v) != 2 {
		panic("vec2 values require exactly 2 elements")
	}
	s.setFloat32SliceValue(key, v)
}

// SetVec3Value assigns vec3 Kage uniform variable.
// v must be a 3-element slice, otherwise this method will panic.
func (s *Shader) SetVec3Value(key string, v []float32) {
	if len(v) != 3 {
		panic("vec3 values require exactly 3 elements")
	}
	s.setFloat32SliceValue(key, v)
}

// SetVec4Value assigns vec4 Kage uniform variable.
// v must be a 4-element slice, otherwise this method will panic.
func (s *Shader) SetVec4Value(key string, v []float32) {
	if len(v) != 4 {
		panic("vec4 values require exactly 4 elements")
	}
	s.setFloat32SliceValue(key, v)
}

// SetIntValue assigns int Kage uniform variable.
// Since Kage uniforms are 32-bits, we use int32 here to avoid overflows.
func (s *Shader) SetIntValue(key string, v int32) {
	// TODO: Kage supports int-typed uniform values.
	// Maybe we should use them instead of converting to float32?
	s.setFloat32Value(key, float32(v))
}

// SetFloatValue assigns float Kage uniform variable.
func (s *Shader) SetFloatValue(key string, v float32) {
	s.setFloat32Value(key, v)
}

func (s *Shader) setFloat32SliceValue(key string, v []float32) {
	if oldValue, ok := s.shaderData[key].([]float32); ok && slices.Equal(oldValue, v) {
		return
	}
	if s.shaderData == nil {
		s.shaderData = make(map[string]any, 2)
	}
	s.shaderData[key] = v
}

func (s *Shader) setFloat32Value(key string, v float32) {
	if oldValue, ok := s.shaderData[key].(float32); ok && oldValue == v {
		return
	}
	if s.shaderData == nil {
		s.shaderData = make(map[string]any, 2)
	}
	s.shaderData[key] = v
}
