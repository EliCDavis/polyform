package texturing_test

import (
	"image/color"
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/stretchr/testify/assert"
)

func TestTexture(t *testing.T) {
	// ARRANGE ================================================================
	tex := texturing.Empty[float64](2, 3)

	// ACT ====================================================================
	tex.Fill(2)
	tex.Set(1, 1, 3)
	copy := tex.Copy()

	// ASSERT =================================================================r
	textures := []texturing.Texture[float64]{tex, copy}
	for _, tex := range textures {
		assert.Equal(t, 2., tex.Get(0, 0))
		assert.Equal(t, 2., tex.Get(1, 0))
		assert.Equal(t, 2., tex.Get(0, 1))
		assert.Equal(t, 3., tex.Get(1, 1))
		assert.Equal(t, 2., tex.Get(0, 2))
		assert.Equal(t, 2., tex.Get(1, 2))

		assert.Equal(t, 2, tex.Width())
		assert.Equal(t, 3, tex.Height())
	}
}

func TestTexture_Convert(t *testing.T) {
	// ARRANGE ================================================================
	tex := texturing.Empty[int](2, 2)
	tex.Set(0, 0, 1)
	tex.Set(0, 1, 2)
	tex.Set(1, 0, 3)
	tex.Set(1, 1, 4)

	// ACT ====================================================================
	out := texturing.Convert(tex, func(x, y, v int) float64 {
		return float64(v) / 2.
	})

	// ASSERT =================================================================r
	assert.Equal(t, 0.5, out.Get(0, 0))
	assert.Equal(t, 1.0, out.Get(0, 1))
	assert.Equal(t, 1.5, out.Get(1, 0))
	assert.Equal(t, 2.0, out.Get(1, 1))

	assert.Equal(t, 2, out.Width())
	assert.Equal(t, 2, out.Height())

}

func TestColorImageMatchesStandardConversion(t *testing.T) {
	colors := []coloring.Color{
		{R: 0.2, G: 0.5, B: 0.8, A: 1},
		{R: -1, G: 2, B: 0.333, A: 0.5},
		{R: 1, G: 0, B: 0.999, A: 0},
	}
	tex := texturing.Empty[coloring.Color](len(colors), 1)
	for x, c := range colors {
		tex.Set(x, 0, c)
	}

	img := texturing.ColorImage(tex)

	for x, c := range colors {
		assert.Equal(t, color.RGBAModel.Convert(c), img.At(x, 0))
	}
}
