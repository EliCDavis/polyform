package marching_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyColorFieldNode(t *testing.T) {
	red := coloring.Color{R: 1, G: 0, B: 0, A: 1}
	blue := coloring.Color{R: 0, G: 0, B: 1, A: 1}

	a := sdf.ColoredField{Distance: sdf.Sphere(vector3.New(0., 0., 0.), 1), Color: sdf.ConstantColor(red)}
	b := sdf.ColoredField{Distance: sdf.Sphere(vector3.New(1.5, 0., 0.), 1), Color: sdf.ConstantColor(blue)}
	field := sdf.SmoothUnionColored(0.5, a, b)

	// 3 vertices: one deep in A, one deep in B, one on the seam.
	mesh := modeling.NewPointCloud(nil, map[string][]vector3.Float64{
		modeling.PositionAttribute: {
			vector3.New(0., 0., 0.),
			vector3.New(1.5, 0., 0.),
			vector3.New(0.75, 0., 0.),
		},
	}, nil, nil)

	node := &nodes.Struct[marching.ApplyColorFieldNode]{
		Data: marching.ApplyColorFieldNode{
			Mesh:  nodes.ConstOutput[modeling.Mesh]{Val: mesh},
			Field: nodes.ConstOutput[sdf.ColoredField]{Val: field},
		},
	}

	got := nodes.GetNodeOutputPort[modeling.Mesh](node, "Out").Value()
	require.True(t, got.HasFloat3Attribute(modeling.ColorAttribute))

	colors := got.Float3Attribute(modeling.ColorAttribute)
	require.Equal(t, 3, colors.Len())

	assert.InDelta(t, 1.0, colors.At(0).X(), 1e-9, "vertex in A should be pure red")
	assert.InDelta(t, 0.0, colors.At(0).Z(), 1e-9)

	assert.InDelta(t, 0.0, colors.At(1).X(), 1e-9, "vertex in B should be pure blue")
	assert.InDelta(t, 1.0, colors.At(1).Z(), 1e-9)

	assert.InDelta(t, 0.5, colors.At(2).X(), 1e-9, "vertex on the seam should be an even blend")
	assert.InDelta(t, 0.5, colors.At(2).Z(), 1e-9)
}
