package meshops_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFlipTriangleWindingNegatesNormals pins that flipping a mesh inside
// out flips its shading too. Reversing winding alone leaves every normal
// pointing at the side that is no longer the front face, so the result
// renders lit from within - which matters most for the one job this node
// has, turning a surface around.
func TestFlipTriangleWindingNegatesNormals(t *testing.T) {
	original := primitives.Cylinder{Sides: 12, Height: 1, Radius: 1}.ToMesh()
	require.True(t, original.HasFloat3Attribute(modeling.NormalAttribute))

	flipped := meshops.FlipTriangleWinding(original)

	before := original.Float3Attribute(modeling.NormalAttribute)
	after := flipped.Float3Attribute(modeling.NormalAttribute)
	require.Equal(t, before.Len(), after.Len())

	for i := 0; i < before.Len(); i++ {
		assert.InDelta(t, 0.0, before.At(i).Add(after.At(i)).Length(), 1e-9,
			"normal %d should be negated", i)
	}
}

// TestFlipTriangleWindingKeepsWindingAndNormalsAgreeing is the property
// that actually matters: after flipping, each triangle's geometric normal
// still agrees with its stored vertex normals.
func TestFlipTriangleWindingKeepsWindingAndNormalsAgreeing(t *testing.T) {
	flipped := meshops.FlipTriangleWinding(primitives.Cylinder{Sides: 12, Height: 1, Radius: 1}.ToMesh())

	positions := flipped.Float3Attribute(modeling.PositionAttribute)
	normals := flipped.Float3Attribute(modeling.NormalAttribute)
	indices := flipped.Indices()

	for i := 0; i < indices.Len(); i += 3 {
		a, b, c := indices.At(i), indices.At(i+1), indices.At(i+2)
		face := positions.At(b).Sub(positions.At(a)).Cross(positions.At(c).Sub(positions.At(a)))
		if face.Length() < 1e-12 {
			continue
		}
		assert.Greaterf(t, face.Normalized().Dot(normals.At(a)), 0.0,
			"triangle %d faces away from its own vertex normal", i/3)
	}
}

// TestFlipTriangleWindingWithoutNormals guards the mesh types that carry
// no normals at all.
func TestFlipTriangleWindingWithoutNormals(t *testing.T) {
	mesh := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(modeling.PositionAttribute, []vector3.Float64{
			vector3.New(0., 0., 0.), vector3.New(1., 0., 0.), vector3.New(0., 1., 0.),
		})

	flipped := meshops.FlipTriangleWinding(mesh)
	assert.False(t, flipped.HasFloat3Attribute(modeling.NormalAttribute))
	assert.Equal(t, 1, flipped.PrimitiveCount())
}
