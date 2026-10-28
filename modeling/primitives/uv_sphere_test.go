package primitives_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnweldedUVSphereNormalsPointOutward(t *testing.T) {
	for name, sphere := range map[string]modeling.Mesh{
		"welded":   primitives.UVSphere(2, 6, 8),
		"unwelded": primitives.UVSphereUnwelded(2, 6, 8),
	} {
		require.True(t, sphere.HasFloat3Attribute(modeling.NormalAttribute), name)
		positions := sphere.Float3Attribute(modeling.PositionAttribute)
		normals := sphere.Float3Attribute(modeling.NormalAttribute)
		for i := range positions.Len() {
			assert.InDelta(t, 1, normals.At(i).Dot(positions.At(i).Normalized()), 1e-9, "%s vertex %d", name, i)
		}
	}
}
