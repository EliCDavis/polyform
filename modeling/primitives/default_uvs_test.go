package primitives_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUVSphereUVsNeverWrapAcrossATriangle(t *testing.T) {
	for name, sphere := range map[string]modeling.Mesh{
		"welded":   primitives.UVSphere(1, 6, 8),
		"unwelded": primitives.UVSphereUnwelded(1, 6, 8),
	} {
		uvs := sphere.Float2Attribute(modeling.TexCoordAttribute)
		indices := sphere.Indices()
		for k := 0; k+2 < indices.Len(); k += 3 {
			a, b, c := uvs.At(indices.At(k)), uvs.At(indices.At(k+1)), uvs.At(indices.At(k+2))
			assert.LessOrEqual(t, max(a.X(), b.X(), c.X())-min(a.X(), b.X(), c.X()), 1.0/8+1e-9, "%s triangle %d", name, k/3)
			assert.LessOrEqual(t, max(a.Y(), b.Y(), c.Y())-min(a.Y(), b.Y(), c.Y()), 1.0/6+1e-9, "%s triangle %d", name, k/3)
		}
	}
}

func TestEveryPrimitiveHasUVsWhenItsUVsAreUnwired(t *testing.T) {
	primitiveNodes := map[string]nodes.Node{
		"cube":        &nodes.Struct[primitives.CubeNode]{},
		"quad":        &nodes.Struct[primitives.QuadNode]{},
		"cylinder":    &nodes.Struct[primitives.CylinderNode]{},
		"hemisphere":  &nodes.Struct[primitives.HemisphereNode]{},
		"uv sphere":   &nodes.Struct[primitives.UvSphereNode]{},
		"quad sphere": &nodes.Struct[primitives.QuadSphereNode]{},
		"circle":      &nodes.Struct[primitives.CircleNode]{},
		"cone":        &nodes.Struct[primitives.ConeNode]{},
		"tube":        &nodes.Struct[primitives.TubeNode]{},
		"torus":       &nodes.Struct[primitives.TorusNode]{},
	}

	for name, node := range primitiveNodes {
		t.Run(name, func(t *testing.T) {
			found := false
			for port, out := range node.Outputs() {
				mesh, ok := out.(nodes.Output[modeling.Mesh])
				if !ok {
					continue
				}
				found = true
				m := mesh.Value()
				if m.PrimitiveCount() == 0 {
					continue
				}
				if assert.True(t, m.HasFloat2Attribute(modeling.TexCoordAttribute), "%s port %q has no UVs", name, port) {
					assert.Equal(t, m.Float3Attribute(modeling.PositionAttribute).Len(), m.Float2Attribute(modeling.TexCoordAttribute).Len())
				}
			}
			require.True(t, found, "%s has no mesh output", name)
		})
	}
}
