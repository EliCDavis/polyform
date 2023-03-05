package obj_test

import (
	"image/color"
	"strings"
	"testing"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/stretchr/testify/assert"
)

func assertColor(t *testing.T, expected color.RGBA, actual color.Color) {
	if assert.NotNil(t, actual) == false {
		return
	}

	r, g, b, a := actual.RGBA()

	assert.Equal(t, expected.R, uint8(r))
	assert.Equal(t, expected.G, uint8(g))
	assert.Equal(t, expected.B, uint8(b))
	assert.Equal(t, expected.A, uint8(a))
}

func Test_ReadMaterial_Simple(t *testing.T) {
	// ARRANGE ================================================================
	matString := `
	
# Example Material
# Pulled from:
# https://people.sc.fsu.edu/~jburkardt/data/mtl/mtl.html

newmtl shinyred
Ka  0.1986  0.0000  0.0000
Kd  0.5922  0.0166  0.0000
Ks  0.5974  0.2084  0.2084
illum 2
Ns 100.2237
map_Kd something/other.jpg
`

	// ACT ====================================================================
	mats, err := obj.ReadMaterials(strings.NewReader(matString))

	// ASSERT =================================================================
	assert.NoError(t, err)
	if assert.Len(t, mats, 1) {
		assert.Equal(t, "shinyred", mats[0].Name)
		if assert.NotNil(t, mats[0].ColorTextureURI) {
			assert.Equal(t, "something/other.jpg", *mats[0].ColorTextureURI)
			assert.InDelta(t, 100.2237, mats[0].SpecularHighlight, 0.0001)
			assertColor(t, color.RGBA{151, 4, 0, 255}, mats[0].DiffuseColor)
		}
	}
}
