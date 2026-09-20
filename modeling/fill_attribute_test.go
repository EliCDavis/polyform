package modeling_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func threePointCloud() modeling.Mesh {
	return modeling.NewPointCloud(nil, map[string][]vector3.Float64{
		modeling.PositionAttribute: {vector3.Zero[float64](), vector3.Right[float64](), vector3.Up[float64]()},
	}, nil, nil)
}

func TestFillAttribute3DNode(t *testing.T) {
	existing := threePointCloud()
	red := vector3.New(1., 0., 0.)

	t.Run("fills every vertex with the value", func(t *testing.T) {
		node := nodetest.NewNode(modeling.FillAttribute3DNode{
			Mesh:      nodetest.NewPortValue(existing),
			Attribute: nodetest.NewPortValue(modeling.ColorAttribute),
			Value:     nodetest.NewPortValue(red),
		})
		out := nodes.GetNodeOutputPort[modeling.Mesh](node, "Out").Value()
		require.True(t, out.HasFloat3Attribute(modeling.ColorAttribute))
		colors := out.Float3Attribute(modeling.ColorAttribute)
		require.Equal(t, 3, colors.Len())
		for i := 0; i < colors.Len(); i++ {
			assert.Equal(t, red, colors.At(i))
		}
		assert.True(t, out.HasFloat3Attribute(modeling.PositionAttribute), "existing attributes are kept")
	})

	t.Run("nil value fills with zero", func(t *testing.T) {
		node := nodetest.NewNode(modeling.FillAttribute3DNode{
			Mesh:      nodetest.NewPortValue(existing),
			Attribute: nodetest.NewPortValue("weights"),
		})
		out := nodes.GetNodeOutputPort[modeling.Mesh](node, "Out").Value()
		require.True(t, out.HasFloat3Attribute("weights"))
		assert.Equal(t, vector3.Zero[float64](), out.Float3Attribute("weights").At(2))
	})

	t.Run("nil attribute passes the mesh through", func(t *testing.T) {
		nodetest.NewSuite(nodetest.NewTestCase(
			"passthrough",
			nodetest.NewNode(modeling.FillAttribute3DNode{
				Mesh:  nodetest.NewPortValue(existing),
				Value: nodetest.NewPortValue(red),
			}),
			nodetest.AssertOutput("Out", existing),
		)).Run(t)
	})
}

func TestFillAttribute1DNode(t *testing.T) {
	node := nodetest.NewNode(modeling.FillAttribute1DNode{
		Mesh:      nodetest.NewPortValue(threePointCloud()),
		Attribute: nodetest.NewPortValue(modeling.OpacityAttribute),
		Value:     nodetest.NewPortValue(0.25),
	})
	out := nodes.GetNodeOutputPort[modeling.Mesh](node, "Out").Value()
	require.True(t, out.HasFloat1Attribute(modeling.OpacityAttribute))
	assert.Equal(t, 3, out.Float1Attribute(modeling.OpacityAttribute).Len())
	assert.Equal(t, 0.25, out.Float1Attribute(modeling.OpacityAttribute).At(1))
}
