package extrude_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/curves"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/sketch"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEveryExtrusionHasUVsWhenItsUVsAreUnwired(t *testing.T) {
	path := []vector3.Float64{vector3.New(0., 0., 0.), vector3.New(1., 0., 0.), vector3.New(2., 1., 0.), vector3.New(3., 1., 0.)}
	square := []vector2.Float64{vector2.New(0., 0.), vector2.New(1., 0.), vector2.New(1., 1.), vector2.New(0., 1.)}
	spline := curves.CatmullRomSplineParameters{Points: path, Alpha: 0.5}.Spline()
	area, err := sketch.Fill([]geometry.Shape{square}, nil)
	require.NoError(t, err)

	extrusions := map[string]nodes.Node{
		"circle": &nodes.Struct[extrude.CircleNode]{Data: extrude.CircleNode{
			Path: nodes.ConstOutput[[]vector3.Float64]{Val: path},
		}},
		"circle along spline": &nodes.Struct[extrude.CircleAlongSplineNode]{Data: extrude.CircleAlongSplineNode{
			Spline: nodes.ConstOutput[curves.Spline]{Val: &spline},
		}},
		"outline": &nodes.Struct[extrude.OutlineNode]{Data: extrude.OutlineNode{
			Outline: nodes.ConstOutput[[]vector2.Float64]{Val: square},
			Path:    nodes.ConstOutput[[]vector3.Float64]{Val: path},
		}},
		"screw": &nodes.Struct[extrude.ScrewNode]{Data: extrude.ScrewNode{
			Line: nodes.ConstOutput[[]vector3.Float64]{Val: []vector3.Float64{vector3.New(1., 0., 0.), vector3.New(1., 1., 0.)}},
		}},
		"text": &nodes.Struct[extrude.TextNode]{Data: extrude.TextNode{
			Text:  nodes.ConstOutput[string]{Val: "Hi"},
			Depth: nodes.ConstOutput[float64]{Val: 0.2},
		}},
		"area": &nodes.Struct[extrude.AreaNode]{Data: extrude.AreaNode{
			Area:  nodes.ConstOutput[sketch.Area]{Val: area},
			Depth: nodes.ConstOutput[float64]{Val: 0.2},
		}},
	}

	for name, node := range extrusions {
		t.Run(name, func(t *testing.T) {
			found := false
			for port, out := range node.Outputs() {
				mesh, ok := out.(nodes.Output[modeling.Mesh])
				if !ok {
					continue
				}
				m := mesh.Value()
				require.NotZero(t, m.PrimitiveCount(), "%s port %q is empty", name, port)
				found = true
				if assert.True(t, m.HasFloat2Attribute(modeling.TexCoordAttribute), "%s port %q has no UVs", name, port) {
					assert.Equal(t, m.Float3Attribute(modeling.PositionAttribute).Len(), m.Float2Attribute(modeling.TexCoordAttribute).Len())
				}
			}
			require.True(t, found, "%s has no mesh output", name)
		})
	}
}
