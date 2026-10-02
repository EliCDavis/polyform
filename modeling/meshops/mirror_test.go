package meshops_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func wedge() modeling.Mesh {
	return modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: {
				vector3.New(1., 0., 0.),
				vector3.New(3., 0., 0.),
				vector3.New(1., 2., 0.),
			},
			modeling.NormalAttribute: {
				vector3.New(1., 0., 0.),
				vector3.New(1., 0., 0.),
				vector3.New(1., 0., 0.),
			},
		})
}

func TestMirrorReflectsPositionsAcrossThePlane(t *testing.T) {
	out := meshops.Mirror(wedge(), vector3.Zero[float64](), vector3.Right[float64]())

	positions := out.Float3Attribute(modeling.PositionAttribute)
	require.Equal(t, 3, positions.Len())
	assert.Equal(t, vector3.New(-1., 0., 0.), positions.At(0))
	assert.Equal(t, vector3.New(-3., 0., 0.), positions.At(1))
	assert.Equal(t, vector3.New(-1., 2., 0.), positions.At(2), "the plane's own axes are untouched")
}

func TestMirrorReversesTheWinding(t *testing.T) {
	out := meshops.Mirror(wedge(), vector3.Zero[float64](), vector3.Right[float64]())

	indices := out.Indices()
	require.Equal(t, 3, indices.Len())
	assert.Equal(t, 1, indices.At(0))
	assert.Equal(t, 0, indices.At(1))
	assert.Equal(t, 2, indices.At(2))
}

func TestMirrorReflectsNormalsWithoutNegatingThem(t *testing.T) {
	out := meshops.Mirror(wedge(), vector3.Zero[float64](), vector3.Right[float64]())

	normals := out.Float3Attribute(modeling.NormalAttribute)
	assert.Equal(t, vector3.New(-1., 0., 0.), normals.At(0),
		"an outward +X normal becomes an outward -X normal")
}

func TestMirrorAboutAnOffsetPlane(t *testing.T) {
	out := meshops.Mirror(wedge(), vector3.New(2., 0., 0.), vector3.Right[float64]())

	positions := out.Float3Attribute(modeling.PositionAttribute)
	assert.Equal(t, vector3.New(3., 0., 0.), positions.At(0))
	assert.Equal(t, vector3.New(1., 0., 0.), positions.At(1))
}

func TestMirrorLeavesAMeshAloneWhenTheNormalIsDegenerate(t *testing.T) {
	out := meshops.Mirror(wedge(), vector3.Zero[float64](), vector3.Zero[float64]())

	positions := out.Float3Attribute(modeling.PositionAttribute)
	assert.Equal(t, vector3.New(1., 0., 0.), positions.At(0), "no plane to reflect across")
}

func TestMirrorNodeKeepsTheOriginalByDefault(t *testing.T) {
	node := &nodes.Struct[meshops.MirrorNode]{
		Data: meshops.MirrorNode{
			Mesh: nodes.ConstOutput[modeling.Mesh]{Val: wedge()},
		},
	}

	out := nodes.GetNodeOutputPort[modeling.Mesh](node, "Out").Value()
	require.Equal(t, 2, out.PrimitiveCount(), "both sides should be present")

	positions := out.Float3Attribute(modeling.PositionAttribute)
	var sawLeft, sawRight bool
	for i := range positions.Len() {
		if positions.At(i).X() > 0 {
			sawRight = true
		}
		if positions.At(i).X() < 0 {
			sawLeft = true
		}
	}
	assert.True(t, sawRight, "the original wedge is at +X")
	assert.True(t, sawLeft, "its reflection is at -X")
}

func TestMirrorNodeUnionFalseDropsTheOriginal(t *testing.T) {
	node := &nodes.Struct[meshops.MirrorNode]{
		Data: meshops.MirrorNode{
			Mesh:  nodes.ConstOutput[modeling.Mesh]{Val: wedge()},
			Union: nodes.ConstOutput[bool]{Val: false},
		},
	}

	out := nodes.GetNodeOutputPort[modeling.Mesh](node, "Out").Value()
	require.Equal(t, 1, out.PrimitiveCount())

	positions := out.Float3Attribute(modeling.PositionAttribute)
	for i := range positions.Len() {
		assert.LessOrEqual(t, positions.At(i).X(), 0., "only the reflection remains")
	}
}
