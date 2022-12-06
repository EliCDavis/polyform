package obj_test

import (
	"strings"
	"testing"

	"github.com/EliCDavis/mesh/obj"
	"github.com/stretchr/testify/assert"
)

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
		}
	}
}
