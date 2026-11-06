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
)

func TestGetAttributeNodesReadEverySize(t *testing.T) {
	mesh := modeling.NewPointCloud(
		map[string][]vector4.Float64{"d": {vector4.New(1., 2., 3., 4.)}},
		map[string][]vector3.Float64{modeling.PositionAttribute: {vector3.New(1., 2., 3.)}},
		map[string][]vector2.Float64{"uv": {vector2.New(0.5, 0.25)}},
		map[string][]float64{modeling.OpacityAttribute: {0.75}},
	)

	nodetest.NewSuite(
		nodetest.NewTestCase("1D",
			nodetest.NewNode(modeling.GetAttribute1DNode{Mesh: nodetest.NewPortValue(mesh), Attribute: nodetest.NewPortValue(modeling.OpacityAttribute)}),
			nodetest.AssertOutput("Out", []float64{0.75}),
		),
		nodetest.NewTestCase("2D",
			nodetest.NewNode(modeling.GetAttribute2DNode{Mesh: nodetest.NewPortValue(mesh), Attribute: nodetest.NewPortValue("uv")}),
			nodetest.AssertOutput("Out", []vector2.Float64{vector2.New(0.5, 0.25)}),
		),
		nodetest.NewTestCase("3D",
			nodetest.NewNode(modeling.GetAttribute3DNode{Mesh: nodetest.NewPortValue(mesh), Attribute: nodetest.NewPortValue(modeling.PositionAttribute)}),
			nodetest.AssertOutput("Out", []vector3.Float64{vector3.New(1., 2., 3.)}),
		),
		nodetest.NewTestCase("4D",
			nodetest.NewNode(modeling.GetAttribute4DNode{Mesh: nodetest.NewPortValue(mesh), Attribute: nodetest.NewPortValue("d")}),
			nodetest.AssertOutput("Out", []vector4.Float64{vector4.New(1., 2., 3., 4.)}),
		),
		nodetest.NewTestCase("missing attribute",
			nodetest.NewNode(modeling.GetAttribute3DNode{Mesh: nodetest.NewPortValue(mesh), Attribute: nodetest.NewPortValue("uv")}),
			nodetest.AssertOutput("Out", []vector3.Float64{}),
			nodetest.AssertOutputError("Out", `mesh has no "uv" attribute of this size`),
		),
	).Run(t)
}

func TestGetAttributeNodeReadsEachMeshInAList(t *testing.T) {
	meshes := []modeling.Mesh{
		modeling.NewPointCloud(nil, nil, nil, map[string][]float64{"a": {1.}}),
		modeling.NewPointCloud(nil, nil, nil, map[string][]float64{"a": {2., 3.}}),
	}

	out := nodes.GetNodeOutputPort[[][]float64](nodetest.NewNode(modeling.GetAttribute1DNode{
		Mesh:      nodetest.NewPortValue(meshes),
		Attribute: nodetest.NewPortValue("a"),
	}), "Out").Value()

	assert.Equal(t, [][]float64{{1.}, {2., 3.}}, out)
}
