package extrude_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lathedTube(t *testing.T) modeling.Mesh {
	t.Helper()

	return nodes.GetNodeOutputPort[modeling.Mesh](&nodes.Struct[extrude.ScrewNode]{
		Data: extrude.ScrewNode{
			Line: nodes.ConstOutput[[]vector3.Float64]{Val: []vector3.Float64{
				vector3.New(1., 0., 0.),
				vector3.New(1., 1., 0.),
			}},
			Segments:    nodes.ConstOutput[int]{Val: 16},
			Revolutions: nodes.ConstOutput[float64]{Val: 1},
		},
	}, "Out").Value()
}

func TestScrewEmitsNormals(t *testing.T) {
	mesh := lathedTube(t)

	require.True(t, mesh.HasFloat3Attribute(modeling.NormalAttribute), "screw output has no Normal attribute")

	normals := mesh.Float3Attribute(modeling.NormalAttribute)
	positions := mesh.Float3Attribute(modeling.PositionAttribute)
	require.Equal(t, positions.Len(), normals.Len(), "expected one normal per vertex")
	require.Greater(t, normals.Len(), 0)

	for i := 0; i < normals.Len(); i++ {
		assert.InDeltaf(t, 1, normals.At(i).Length(), 1e-9, "normal %d is not unit length", i)
	}
}

func TestScrewNormalsPointOutward(t *testing.T) {
	mesh := lathedTube(t)

	normals := mesh.Float3Attribute(modeling.NormalAttribute)
	positions := mesh.Float3Attribute(modeling.PositionAttribute)

	for i := 0; i < normals.Len(); i++ {
		p := positions.At(i)
		radial := vector3.New(p.X(), 0., p.Z())
		if radial.Length() == 0 {
			continue
		}

		n := normals.At(i)
		assert.InDeltaf(t, 0, n.Y(), 1e-6, "normal %d should be flat along the tube's axis", i)
		assert.Greaterf(t, n.Dot(radial.Normalized()), 0.9,
			"normal %d should point radially outward, got %v at %v", i, n, p)
	}
}
