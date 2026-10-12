package meshops_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlatNormalsOnAWeldedCubeGivesOneNormalPerTriangle(t *testing.T) {
	flat := meshops.FlatNormals(primitives.Cube{Width: 1, Height: 1, Depth: 1}.Welded())

	positions := flat.Float3Attribute(modeling.PositionAttribute)
	normals := flat.Float3Attribute(modeling.NormalAttribute)
	require.Equal(t, 36, positions.Len())
	require.Equal(t, 36, normals.Len())
	require.Equal(t, 12, flat.PrimitiveCount())

	unique := map[vector3.Float64]struct{}{}
	for i := 0; i < flat.PrimitiveCount(); i++ {
		tri := flat.Tri(i)
		face := tri.P2Vec3Attr(modeling.PositionAttribute).
			Sub(tri.P1Vec3Attr(modeling.PositionAttribute)).
			Cross(tri.P3Vec3Attr(modeling.PositionAttribute).Sub(tri.P1Vec3Attr(modeling.PositionAttribute))).
			Normalized()

		for _, n := range []vector3.Float64{
			tri.P1Vec3Attr(modeling.NormalAttribute),
			tri.P2Vec3Attr(modeling.NormalAttribute),
			tri.P3Vec3Attr(modeling.NormalAttribute),
		} {
			assert.InDeltaf(t, 0, n.Sub(face).Length(), 1e-9, "triangle %d", i)
		}
		unique[face.Round()] = struct{}{}
	}
	assert.Len(t, unique, 6)
}

func TestFlatNormalsNodeSplitsSharedVertices(t *testing.T) {
	cube := primitives.Cube{Width: 1, Height: 1, Depth: 1}.Welded()
	node := &nodes.Struct[meshops.FlatNormalsNode]{
		Data: meshops.FlatNormalsNode{Mesh: nodes.ConstOutput[modeling.Mesh]{Val: cube}},
	}

	out := nodes.GetNodeOutputPort[modeling.Mesh](node, "Out").Value()
	assert.Equal(t, 36, out.AttributeLength())
}
