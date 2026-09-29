package meshops_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestSrgbToLinearNodePassesAnEmptyMeshThrough(t *testing.T) {
	empty := modeling.EmptyMesh(modeling.TriangleTopology)
	got := nodes.GetNodeOutputPort[modeling.Mesh](&nodes.Struct[meshops.SrgbToLinearNode]{
		Data: meshops.SrgbToLinearNode{Mesh: nodes.ConstOutput[modeling.Mesh]{Val: empty}},
	}, "Out").Value()
	assert.Equal(t, 0, got.PrimitiveCount())
}

func TestSrgbToLinearNodeConvertsAColoredMesh(t *testing.T) {
	mesh := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(modeling.PositionAttribute, []vector3.Float64{vector3.Zero[float64](), vector3.Right[float64](), vector3.Up[float64]()}).
		SetFloat3Attribute(modeling.ColorAttribute, []vector3.Float64{vector3.One[float64](), vector3.Zero[float64](), vector3.New(0.5, 0.5, 0.5)})

	got := nodes.GetNodeOutputPort[modeling.Mesh](&nodes.Struct[meshops.SrgbToLinearNode]{
		Data: meshops.SrgbToLinearNode{Mesh: nodes.ConstOutput[modeling.Mesh]{Val: mesh}},
	}, "Out").Value()

	colors := got.Float3Attribute(modeling.ColorAttribute)
	assert.InDelta(t, 1, colors.At(0).X(), 1e-9)
	assert.InDelta(t, 0, colors.At(1).X(), 1e-9)
	assert.InDelta(t, 0.214, colors.At(2).X(), 0.001, "mid grey darkens in linear space")
}
