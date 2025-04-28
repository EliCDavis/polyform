package texturing_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/stretchr/testify/assert"
)

func TestTexture(t *testing.T) {
	// ARRANGE ================================================================
	tex := texturing.NewTexture[float64](2, 3)

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
	tex := texturing.NewTexture[int](2, 2)
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
