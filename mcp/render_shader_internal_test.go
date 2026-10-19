package mcp

import (
	"image"
	"image/color"
	"testing"

	"github.com/fogleman/fauxgl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func solidTexture(c color.Color) fauxgl.Texture {
	im := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := range 2 {
		for x := range 2 {
			im.Set(x, y, c)
		}
	}
	return fauxgl.NewImageTexture(im)
}

// unlit isolates the albedo the fragment stage produced from the lighting
// and tone mapping applied on top of it.
func unlitShader(tex fauxgl.Texture) previewShader {
	phong := fauxgl.NewPhongShader(fauxgl.Identity(), fauxgl.V(0, 0, 1), fauxgl.V(0, 0, 1))
	phong.ObjectColor = fauxgl.Discard
	phong.AmbientColor = fauxgl.Color{R: 1, G: 1, B: 1, A: 1}
	phong.DiffuseColor = fauxgl.Color{}
	phong.SpecularPower = 0
	phong.Texture = tex
	return previewShader{PhongShader: phong}
}

func fragment(s previewShader, vertexColor fauxgl.Color) fauxgl.Color {
	return s.Fragment(fauxgl.Vertex{
		Normal:  fauxgl.V(0, 0, 1),
		Color:   vertexColor,
		Texture: fauxgl.V(0.5, 0.5, 0),
	})
}

// glTF multiplies a base color texture by COLOR_0. The preview used to let
// the texture win outright, so dirt or wear painted per vertex over a
// textured part showed in every real viewer and not here.
func TestTexturedSurfaceTakesItsVertexColor(t *testing.T) {
	shader := unlitShader(solidTexture(color.RGBA{R: 255, G: 255, B: 255, A: 255}))

	white := fragment(shader, fauxgl.Color{R: 1, G: 1, B: 1, A: 1})
	darkened := fragment(shader, fauxgl.Color{R: 0.25, G: 0.25, B: 0.25, A: 1})

	// Tone mapping sits on top of the multiply, so the ratio is compressed
	// rather than the 0.25 that went in - what matters is that the paint
	// reaches the surface at all.
	require.Greater(t, white.R, darkened.R, "vertex paint has to darken a textured surface")
	assert.Less(t, darkened.R/white.R, 0.9, "and visibly so")
}

func TestTexturedSurfaceKeepsItsTextureWhenVertexColorIsWhite(t *testing.T) {
	shader := unlitShader(solidTexture(color.RGBA{R: 255, G: 0, B: 0, A: 255}))

	got := fragment(shader, fauxgl.Color{R: 1, G: 1, B: 1, A: 1})

	assert.Greater(t, got.R, got.G, "white is the identity for the multiply")
	assert.Greater(t, got.R, got.B)
}

func TestVertexColorTintsTheTexture(t *testing.T) {
	shader := unlitShader(solidTexture(color.RGBA{R: 255, G: 255, B: 255, A: 255}))

	got := fragment(shader, fauxgl.Color{R: 1, G: 0, B: 0, A: 1})

	assert.Greater(t, got.R, got.G, "a red vertex color over a white texture reads red")
	assert.Greater(t, got.R, got.B)
}
