package modeling_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetAttribute1DNode(t *testing.T) {
	t.Run("nil attribute or data returns mesh or empty point mesh", func(t *testing.T) {
		existing := modeling.NewPointCloud(nil, nil, nil, map[string][]float64{
			"keep": {1.},
		})

		suite := nodetest.NewSuite(
			nodetest.NewTestCase(
				"all nil => empty point mesh",
				nodetest.NewNode(modeling.SetAttribute1DNode{}),
				nodetest.AssertOutput("Out", modeling.EmptyMesh(modeling.PointTopology)),
			),
			nodetest.NewTestCase(
				"nil attr => passthrough mesh",
				nodetest.NewNode(modeling.SetAttribute1DNode{
					Mesh: nodetest.NewPortValue(existing),
					Data: nodetest.NewPortValue([]float64{9.}),
				}),
				nodetest.AssertOutput("Out", existing),
			),
			nodetest.NewTestCase(
				"nil data => passthrough mesh",
				nodetest.NewNode(modeling.SetAttribute1DNode{
					Mesh:      nodetest.NewPortValue(existing),
					Attribute: nodetest.NewPortValue("a"),
				}),
				nodetest.AssertOutput("Out", existing),
			),
		)
		suite.Run(t)
	})

	t.Run("nil mesh creates point cloud with attribute", func(t *testing.T) {
		node := nodetest.NewNode(modeling.SetAttribute1DNode{
			Attribute: nodetest.NewPortValue("a"),
			Data:      nodetest.NewPortValue([]float64{1., 2.}),
		})
		out := nodes.GetNodeOutputPort[modeling.Mesh](node, "Out").Value()
		require.True(t, out.HasFloat1Attribute("a"))
		assert.Equal(t, 1., out.Float1Attribute("a").At(0))
		assert.Equal(t, 2., out.Float1Attribute("a").At(1))
		assert.Equal(t, modeling.PointTopology, out.Topology())
	})

	t.Run("sets attribute on existing mesh", func(t *testing.T) {
		existing := modeling.NewPointCloud(nil, nil, nil, map[string][]float64{
			"old": {0.},
		})
		node := nodetest.NewNode(modeling.SetAttribute1DNode{
			Mesh:      nodetest.NewPortValue(existing),
			Attribute: nodetest.NewPortValue("a"),
			Data:      nodetest.NewPortValue([]float64{3.}),
		})
		out := nodes.GetNodeOutputPort[modeling.Mesh](node, "Out").Value()
		require.True(t, out.HasFloat1Attribute("a"))
		assert.Equal(t, 3., out.Float1Attribute("a").At(0))
		require.True(t, out.HasFloat1Attribute("old"))
	})
}

func TestSetAttribute2DNode(t *testing.T) {
	t.Run("nil attribute or data returns mesh or empty point mesh", func(t *testing.T) {
		existing := modeling.NewPointCloud(nil, nil, map[string][]vector2.Float64{
			"keep": {vector2.New(1., 0.)},
		}, nil)

		suite := nodetest.NewSuite(
			nodetest.NewTestCase(
				"all nil => empty point mesh",
				nodetest.NewNode(modeling.SetAttribute2DNode{}),
				nodetest.AssertOutput("Out", modeling.EmptyMesh(modeling.PointTopology)),
			),
			nodetest.NewTestCase(
				"nil attr => passthrough mesh",
				nodetest.NewNode(modeling.SetAttribute2DNode{
					Mesh: nodetest.NewPortValue(existing),
					Data: nodetest.NewPortValue([]vector2.Float64{vector2.New(9., 9.)}),
				}),
				nodetest.AssertOutput("Out", existing),
			),
			nodetest.NewTestCase(
				"nil data => passthrough mesh",
				nodetest.NewNode(modeling.SetAttribute2DNode{
					Mesh:      nodetest.NewPortValue(existing),
					Attribute: nodetest.NewPortValue("a"),
				}),
				nodetest.AssertOutput("Out", existing),
			),
		)
		suite.Run(t)
	})

	t.Run("nil mesh creates point cloud with attribute", func(t *testing.T) {
		node := nodetest.NewNode(modeling.SetAttribute2DNode{
			Attribute: nodetest.NewPortValue("uv"),
			Data: nodetest.NewPortValue([]vector2.Float64{
				vector2.New(0., 1.),
				vector2.New(1., 0.),
			}),
		})
		out := nodes.GetNodeOutputPort[modeling.Mesh](node, "Out").Value()
		require.True(t, out.HasFloat2Attribute("uv"))
		assert.Equal(t, vector2.New(0., 1.), out.Float2Attribute("uv").At(0))
		assert.Equal(t, vector2.New(1., 0.), out.Float2Attribute("uv").At(1))
		assert.Equal(t, modeling.PointTopology, out.Topology())
	})

	t.Run("sets attribute on existing mesh", func(t *testing.T) {
		existing := modeling.NewPointCloud(nil, nil, map[string][]vector2.Float64{
			"old": {vector2.Zero[float64]()},
		}, nil)
		node := nodetest.NewNode(modeling.SetAttribute2DNode{
			Mesh:      nodetest.NewPortValue(existing),
			Attribute: nodetest.NewPortValue("uv"),
			Data:      nodetest.NewPortValue([]vector2.Float64{vector2.New(0.5, 0.5)}),
		})
		out := nodes.GetNodeOutputPort[modeling.Mesh](node, "Out").Value()
		require.True(t, out.HasFloat2Attribute("uv"))
		assert.Equal(t, vector2.New(0.5, 0.5), out.Float2Attribute("uv").At(0))
		require.True(t, out.HasFloat2Attribute("old"))
	})
}

func TestSetAttribute3DNode(t *testing.T) {
	t.Run("nil attribute or data returns mesh or empty point mesh", func(t *testing.T) {
		existing := modeling.NewPointCloud(nil, map[string][]vector3.Float64{
			"keep": {vector3.New(1., 0., 0.)},
		}, nil, nil)

		suite := nodetest.NewSuite(
			nodetest.NewTestCase(
				"all nil => empty point mesh",
				nodetest.NewNode(modeling.SetAttribute3DNode{}),
				nodetest.AssertOutput("Out", modeling.EmptyMesh(modeling.PointTopology)),
			),
			nodetest.NewTestCase(
				"nil attr => passthrough mesh",
				nodetest.NewNode(modeling.SetAttribute3DNode{
					Mesh: nodetest.NewPortValue(existing),
					Data: nodetest.NewPortValue([]vector3.Float64{vector3.New(9., 9., 9.)}),
				}),
				nodetest.AssertOutput("Out", existing),
			),
			nodetest.NewTestCase(
				"nil data => passthrough mesh",
				nodetest.NewNode(modeling.SetAttribute3DNode{
					Mesh:      nodetest.NewPortValue(existing),
					Attribute: nodetest.NewPortValue("a"),
				}),
				nodetest.AssertOutput("Out", existing),
			),
		)
		suite.Run(t)
	})

	t.Run("nil mesh creates point cloud with attribute", func(t *testing.T) {
		node := nodetest.NewNode(modeling.SetAttribute3DNode{
			Attribute: nodetest.NewPortValue(modeling.PositionAttribute),
			Data: nodetest.NewPortValue([]vector3.Float64{
				vector3.New(0., 1., 0.),
				vector3.New(1., 0., 0.),
			}),
		})
		out := nodes.GetNodeOutputPort[modeling.Mesh](node, "Out").Value()
		require.True(t, out.HasFloat3Attribute(modeling.PositionAttribute))
		assert.Equal(t, vector3.New(0., 1., 0.), out.Float3Attribute(modeling.PositionAttribute).At(0))
		assert.Equal(t, vector3.New(1., 0., 0.), out.Float3Attribute(modeling.PositionAttribute).At(1))
		assert.Equal(t, modeling.PointTopology, out.Topology())
	})

	t.Run("sets attribute on existing mesh", func(t *testing.T) {
		existing := modeling.NewPointCloud(nil, map[string][]vector3.Float64{
			"old": {vector3.Zero[float64]()},
		}, nil, nil)
		node := nodetest.NewNode(modeling.SetAttribute3DNode{
			Mesh:      nodetest.NewPortValue(existing),
			Attribute: nodetest.NewPortValue(modeling.NormalAttribute),
			Data:      nodetest.NewPortValue([]vector3.Float64{vector3.Up[float64]()}),
		})
		out := nodes.GetNodeOutputPort[modeling.Mesh](node, "Out").Value()
		require.True(t, out.HasFloat3Attribute(modeling.NormalAttribute))
		assert.Equal(t, vector3.Up[float64](), out.Float3Attribute(modeling.NormalAttribute).At(0))
		require.True(t, out.HasFloat3Attribute("old"))
	})
}

func TestSetAttribute4DNode(t *testing.T) {
	t.Run("nil attribute or data returns mesh or empty point mesh", func(t *testing.T) {
		existing := modeling.NewPointCloud(map[string][]vector4.Float64{
			"keep": {vector4.New(1., 0., 0., 0.)},
		}, nil, nil, nil)

		suite := nodetest.NewSuite(
			nodetest.NewTestCase(
				"all nil => empty point mesh",
				nodetest.NewNode(modeling.SetAttribute4DNode{}),
				nodetest.AssertOutput("Out", modeling.EmptyMesh(modeling.PointTopology)),
			),
			nodetest.NewTestCase(
				"nil attr => passthrough mesh",
				nodetest.NewNode(modeling.SetAttribute4DNode{
					Mesh: nodetest.NewPortValue(existing),
					Data: nodetest.NewPortValue([]vector4.Float64{vector4.New(9., 9., 9., 9.)}),
				}),
				nodetest.AssertOutput("Out", existing),
			),
			nodetest.NewTestCase(
				"nil data => passthrough mesh",
				nodetest.NewNode(modeling.SetAttribute4DNode{
					Mesh:      nodetest.NewPortValue(existing),
					Attribute: nodetest.NewPortValue("a"),
				}),
				nodetest.AssertOutput("Out", existing),
			),
		)
		suite.Run(t)
	})

	t.Run("nil mesh creates point cloud with attribute", func(t *testing.T) {
		node := nodetest.NewNode(modeling.SetAttribute4DNode{
			Attribute: nodetest.NewPortValue(modeling.WeightAttribute),
			Data: nodetest.NewPortValue([]vector4.Float64{
				vector4.New(1., 0., 0., 0.),
				vector4.New(0., 1., 0., 0.),
			}),
		})
		out := nodes.GetNodeOutputPort[modeling.Mesh](node, "Out").Value()
		require.True(t, out.HasFloat4Attribute(modeling.WeightAttribute))
		assert.Equal(t, vector4.New(1., 0., 0., 0.), out.Float4Attribute(modeling.WeightAttribute).At(0))
		assert.Equal(t, vector4.New(0., 1., 0., 0.), out.Float4Attribute(modeling.WeightAttribute).At(1))
		assert.Equal(t, modeling.PointTopology, out.Topology())
	})

	t.Run("sets attribute on existing mesh", func(t *testing.T) {
		existing := modeling.NewPointCloud(map[string][]vector4.Float64{
			"old": {vector4.Zero[float64]()},
		}, nil, nil, nil)
		node := nodetest.NewNode(modeling.SetAttribute4DNode{
			Mesh:      nodetest.NewPortValue(existing),
			Attribute: nodetest.NewPortValue(modeling.JointAttribute),
			Data:      nodetest.NewPortValue([]vector4.Float64{vector4.New(0., 1., 2., 3.)}),
		})
		out := nodes.GetNodeOutputPort[modeling.Mesh](node, "Out").Value()
		require.True(t, out.HasFloat4Attribute(modeling.JointAttribute))
		assert.Equal(t, vector4.New(0., 1., 2., 3.), out.Float4Attribute(modeling.JointAttribute).At(0))
		require.True(t, out.HasFloat4Attribute("old"))
	})
}
