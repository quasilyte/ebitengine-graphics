package graphics

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/gmath"
)

// Sprite is a feature-rich ebiten.Image wrapper.
//
// Sprites make many operations over the image easier.
// It also tries to avoid some performance pitfalls related
// to some edge cases of Ebitengine.
//
// Sprite implements gscene Graphics interface.
type Sprite struct {
	image    *ebiten.Image
	subImage *ebiten.Image

	colorScale       ColorScale
	ebitenColorScale ebiten.ColorScale

	scaleX float64
	scaleY float64

	// Pos is a sprite location binder.
	// See Pos documentation to learn how it works.
	//
	// When rendering an image, Pos.Resolve() will be used
	// to calculate the final position.
	Pos gmath.Pos

	// Rotation is a sprite rotation binder.
	// It's expected that sprite's rotation depends on
	// some other object rotation, hence the pointer.
	Rotation *gmath.Rad

	// Shader is an shader that will be used during rendering of the image.
	// Use NewShader to initialize this field.
	//
	// If nil, no shaders will be used.
	Shader *Shader

	cache *Cache

	frameOffsetX uint16
	frameOffsetY uint16

	frameWidth  uint16
	frameHeight uint16

	flags spriteFlag
}

type spriteFlag uint8

const (
	spriteFlagCentered spriteFlag = 1 << iota
	spriteFlagFlipHorizontal
	spriteFlagFlipVertical
	spriteFlagVisible
	spriteFlagSubImageChanged
	spriteFlagDisposed
)

// NewSprite returns an empty sprite.
// Use SetImage method to assign a texture to it.
//
// By default, a sprite has these properties:
// * Centered=true
// * Visible=true
// * ScaleX and ScaleY are 1
// * The ColorScale is {1, 1, 1, 1}
func NewSprite(cache *Cache) *Sprite {
	return &Sprite{
		cache:            cache,
		colorScale:       defaultColorScale,
		ebitenColorScale: defaultColorScale.toEbitenColorScale(),
		scaleX:           1,
		scaleY:           1,
		flags:            spriteFlagVisible | spriteFlagCentered,
	}
}

// BoundsRect returns the properly positioned image containing rectangle.
//
// This is useful when trying to calculate whether this sprite is contained
// inside some area or not (like a camera view area).
//
// The bounding rectangle can't be used for collisions since it treats
// the frame size as an object size.
func (s *Sprite) BoundsRect() gmath.Rect {
	pos := s.Pos.Resolve()
	if s.IsCentered() {
		offset := gmath.Vec{X: float64(s.frameWidth / 2), Y: float64(s.frameHeight / 2)}
		return gmath.Rect{
			Min: pos.Sub(offset),
			Max: pos.Add(offset),
		}
	}
	return gmath.Rect{
		Min: pos,
		Max: pos.Add(gmath.Vec{X: float64(s.frameWidth), Y: float64(s.frameHeight)}),
	}
}

// ImageWidth returns the bound image width.
// The image size (width/height) can't be changed,
// unless a new image is assigned to the sprite.
func (s *Sprite) ImageWidth() int {
	bounds := s.image.Bounds()
	return bounds.Dx()
}

// ImageHeight returns the bound image height.
// The image size (width/height) can't be changed,
// unless a new image is assigned to the sprite.
func (s *Sprite) ImageHeight() int {
	bounds := s.image.Bounds()
	return bounds.Dy()
}

// GetFrameWidth returns the current frame width.
// Use SetFrameWidth to change it.
func (s *Sprite) GetFrameWidth() int {
	return int(s.frameWidth)
}

// GetFrameHeight returns the current frame height.
// Use SetFrameHeight to change it.
func (s *Sprite) GetFrameHeight() int {
	return int(s.frameHeight)
}

// SetFrameWidth assigns new frame width.
// Use GetFrameWidth to retrieve the current value.
//
// The frame sizes are useful when working with an underlying image
// that contains several logical images ("frames").
// A frame size defines an image rectangle sizes to be used.
// A frame offset defines the rectangle Min value.
func (s *Sprite) SetFrameWidth(w int) {
	uw := uint16(w)
	if s.frameWidth == uw {
		return
	}

	s.frameWidth = uw
	s.flags |= spriteFlagSubImageChanged
}

// SetFrameHeight assigns new frame height.
// Use GetFrameHeight to retrieve the current value.
//
// The frame sizes are useful when working with an underlying image
// that contains several logical images ("frames").
// A frame size defines an image rectangle sizes to be used.
// A frame offset defines the rectangle Min value.
func (s *Sprite) SetFrameHeight(h int) {
	uh := uint16(h)
	if s.frameHeight == uh {
		return
	}

	s.frameHeight = uh
	s.flags |= spriteFlagSubImageChanged
}

// GetColorScale is used to retrieve the current color scale value of the sprite.
// Use SetColorScale to change it.
func (s *Sprite) GetColorScale() ColorScale {
	return s.colorScale
}

// SetColorScale assigns a new ColorScale to this sprite.
// Use GetColorScale to retrieve the current color scale.
func (s *Sprite) SetColorScale(cs ColorScale) {
	if s.colorScale == cs {
		return
	}
	s.colorScale = cs
	s.ebitenColorScale = s.colorScale.toEbitenColorScale()
}

// GetAlpha is a shorthand for GetColorScale().A expression.
// It's mostly provided for a symmetry with SetAlpha.
func (s *Sprite) GetAlpha() float32 { return s.colorScale.A }

// SetAlpha is a convenient way to change the alpha value of the ColorScale.
func (s *Sprite) SetAlpha(a float32) {
	if s.colorScale.A == a {
		return
	}
	s.colorScale.A = a
	s.ebitenColorScale = s.colorScale.toEbitenColorScale()
}

// Dispose marks this sprite for deletion.
// After calling this method, IsDisposed will report true.
//
// Note that it's up to the scene to actually detach this sprite.
// This method only sets a flag but doesn't delete anything.
func (s *Sprite) Dispose() { s.flags |= spriteFlagDisposed }

// IsDisposed reports whether this sprite is marked for deletion.
// IsDisposed returns true only after Disposed was called on this sprite.
func (s *Sprite) IsDisposed() bool { return s.getFlag(spriteFlagDisposed) }

// IsCentered reports whether Centered flag is set.
// Use SetCentered to change this flag value.
//
// When sprite is centered, its image origin will be {w/2, h/2} during rendering.
// It also makes the sprite properly rotate around that origin point.
func (s *Sprite) IsCentered() bool { return s.getFlag(spriteFlagCentered) }

// SetCentered changes the Centered flag value.
// Use IsCentered to get the current flag value.
func (s *Sprite) SetCentered(centered bool) { s.setFlag(spriteFlagCentered, centered) }

// IsVisible reports whether this sprite is visible.
// Use SetVisibility to change this flag value.
//
// When sprite is invisible (visible=false), its image will not be rendered at all.
// This is an efficient way to temporarily hide a sprite.
func (s *Sprite) IsVisible() bool { return s.getFlag(spriteFlagVisible) }

// SetVisibility changes the Visible flag value.
// It can be used to show or hide the sprite.
// Use IsVisible to get the current flag value.
func (s *Sprite) SetVisibility(visible bool) { s.setFlag(spriteFlagVisible, visible) }

// IsHorizontallyFlipped reports whether HorizontalFlip flag is set.
// Use SetHorizontalFlip to change this flag value.
//
// When sprite is horizontally flipped, it's image will be mirrored horizontally.
func (s *Sprite) IsHorizontallyFlipped() bool { return s.getFlag(spriteFlagFlipHorizontal) }

// SetHorizontalFlip changes the HorizontalFlip flag value.
// Use IsHorizontallyFlipped to get the current flag value.
func (s *Sprite) SetHorizontalFlip(hflip bool) { s.setFlag(spriteFlagFlipHorizontal, hflip) }

// IsVerticallyFlipped reports whether VerticalFlip flag is set.
// Use SetVecricalFlip to change this flag value.
//
// When sprite is vertically flipped, it's image will be mirrored vertically.
func (s *Sprite) IsVerticallyFlipped() bool { return s.getFlag(spriteFlagFlipVertical) }

// SetVecricalFlip changes the VerticalFlip flag value.
// Use IsVerticallyFlipped to get the current flag value.
func (s *Sprite) SetVecricalFlip(vflip bool) { s.setFlag(spriteFlagFlipVertical, vflip) }

// GetFrameOffsetX returns the currently configured frame offset X.
// Use SetFrameOffsetX to change it.
func (s *Sprite) GetFrameOffsetX() int {
	return int(s.frameOffsetX)
}

// GetFrameOffsetY returns the currently configured frame offset Y.
// Use SetFrameOffsetY to change it.
func (s *Sprite) GetFrameOffsetY() int {
	return int(s.frameOffsetY)
}

// SetFrameOffsetX assigns new frame X offset.
// Use GetFrameOffsetX to retrieve the current offset values.
//
// The frame offsets are useful when working with an underlying image
// that contains several logical images ("frames").
// A frame offset defines the rectangle Min value.
// A frame size defines an image rectangle sizes to be used.
func (s *Sprite) SetFrameOffsetX(x int) {
	ux := uint16(x)
	if s.frameOffsetX == ux {
		return
	}

	s.frameOffsetX = ux
	s.flags |= spriteFlagSubImageChanged
}

// SetFrameOffsetY assigns new frame Y offset.
// Use GetFrameOffsetY to retrieve the current offset values.
//
// The frame offsets are useful when working with an underlying image
// that contains several logical images ("frames").
// A frame offset defines the rectangle Min value.
// A frame size defines an image rectangle sizes to be used.
func (s *Sprite) SetFrameOffsetY(y int) {
	uy := uint16(y)
	if s.frameOffsetY == uy {
		return
	}

	s.frameOffsetY = uy
	s.flags |= spriteFlagSubImageChanged
}

// SetImage changes the image associated with a sprite.
//
// Assigning an image sets the frame offsets to {0, 0}.
// The default frame width/height are image sizes.
func (s *Sprite) SetImage(img *ebiten.Image) {
	s.image = img

	imageBounds := img.Bounds()
	s.frameWidth = uint16(imageBounds.Dx())
	s.frameHeight = uint16(imageBounds.Dy())
	s.frameOffsetX = 0
	s.frameOffsetY = 0

	s.flags |= spriteFlagSubImageChanged
}

// GetImage returns the sprite's current texture image.
func (s *Sprite) GetImage() *ebiten.Image {
	return s.image
}

// Draw renders the associated image onto the provided dst image.
//
// This method is a shorthand to DrawWithOffset(dst, {})
// which also implements the gscene.Graphics interface.
//
// See DrawWithOptions for more info.
func (s *Sprite) Draw(dst *ebiten.Image) {
	s.DrawWithOptions(dst, DrawOptions{})
}

// DrawWithOffset renders the associated image onto the provided dst image
// while also using the extra provided offset.
func (s *Sprite) DrawWithOptions(dst *ebiten.Image, opts DrawOptions) {
	// Calculations that are expensive to re-calculate on every Draw call
	// should be memorized inside Sprite object.
	// Otherwise we should make compute it here to avoid making Sprite object too big.
	//
	// The order of operations in this function matters.

	// Try to save some processing time if this sprite should not be rendered.
	if !s.IsVisible() || s.image == nil || s.colorScale.A == 0 {
		return
	}

	var drawOptions ebiten.DrawImageOptions
	drawOptions.ColorScale = s.ebitenColorScale

	if s.IsHorizontallyFlipped() {
		drawOptions.GeoM.Scale(-1, 1)
		drawOptions.GeoM.Translate(float64(s.frameWidth), 0)
	}
	if s.IsVerticallyFlipped() {
		drawOptions.GeoM.Scale(1, -1)
		drawOptions.GeoM.Translate(0, float64(s.frameHeight))
	}

	origin := gmath.Vec{}
	if s.IsCentered() {
		origin = gmath.Vec{X: float64(s.frameWidth / 2), Y: float64(s.frameHeight / 2)}
	}

	targetRotation := opts.Rotation
	if s.Rotation != nil {
		targetRotation += *s.Rotation
	}

	// The rotation and scaling should be done around the origin point.
	drawOptions.GeoM.Translate(-origin.X, -origin.Y)
	if targetRotation != 0 {
		drawOptions.GeoM.Rotate(float64(targetRotation))
	}
	if s.scaleX != 1 || s.scaleY != 1 {
		drawOptions.GeoM.Scale(s.scaleX, s.scaleY)
	}

	pos := opts.Offset.Add(s.Pos.Resolve())
	drawOptions.GeoM.Translate(pos.X, pos.Y)

	// Making a sub-image can be more expensive than we would like it
	// to be, therefore we cache the subimage result and update it
	// only when subimage reslicing might be needed.
	// https://github.com/hajimehoshi/ebiten/issues/2902
	if s.getFlag(spriteFlagSubImageChanged) {
		clearFlag(&s.flags, spriteFlagSubImageChanged)
		s.updateSubImage()
	}

	srcImage := s.subImage
	if srcImage == nil {
		srcImage = s.image
	}

	if s.Shader == nil || !s.Shader.Enabled {
		dst.DrawImage(srcImage, &drawOptions)
		return
	}

	srcImageBounds := srcImage.Bounds()
	var options ebiten.DrawRectShaderOptions
	options.GeoM = drawOptions.GeoM
	options.ColorScale = drawOptions.ColorScale
	options.Images[0] = srcImage
	options.Images[1] = s.Shader.Texture1
	options.Images[2] = s.Shader.Texture2
	options.Images[3] = s.Shader.Texture3
	options.Uniforms = s.Shader.shaderData
	dst.DrawRectShader(srcImageBounds.Dx(), srcImageBounds.Dy(), s.Shader.compiled, &options)
}

func (s *Sprite) updateSubImage() {
	imageBounds := s.image.Bounds()

	needSubImage := (s.frameOffsetX != 0 || s.frameOffsetY != 0) ||
		s.frameWidth != uint16(imageBounds.Dx()) ||
		s.frameHeight != uint16(imageBounds.Dy())
	if !needSubImage {
		s.subImage = nil
		return
	}

	subImageBounds := image.Rectangle{
		Min: image.Point{
			X: int(s.frameOffsetX),
			Y: int(s.frameOffsetY),
		},
		Max: image.Point{
			X: int(s.frameOffsetX) + int(s.frameWidth),
			Y: int(s.frameOffsetY) + int(s.frameHeight),
		},
	}
	s.subImage = s.image.SubImage(subImageBounds).(*ebiten.Image)
}

func (s *Sprite) getFlag(f spriteFlag) bool {
	return getFlag(s.flags, f)
}

func (s *Sprite) setFlag(f spriteFlag, v bool) {
	setFlag(&s.flags, f, v)
}
