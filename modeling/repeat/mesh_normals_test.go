package repeat_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// singleTriangleFacingUp is a minimal mesh whose one normal points +Y, so
// any change in orientation is unambiguous.
func singleTriangleFacingUp() modeling.Mesh {
	return modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(modeling.PositionAttribute, []vector3.Float64{
			vector3.New(0., 0., 0.),
			vector3.New(1., 0., 0.),
			vector3.New(0., 0., 1.),
		}).
		SetFloat3Attribute(modeling.NormalAttribute, []vector3.Float64{
			vector3.Up[float64](),
			vector3.Up[float64](),
			vector3.Up[float64](),
		})
}

// TestMeshRotatesNormalsWithCopies covers a silent correctness bug: every
// copy was placed by transforming positions only, so a rotated copy kept
// the original's normals and was lit as though it had never turned.
func TestMeshRotatesNormalsWithCopies(t *testing.T) {
	// Second copy is rolled a quarter turn about Z, taking +Y to -X.
	quarterTurnZ := quaternion.FromEulerAngle(vector3.New(0., 0., math.Pi/2))

	result := repeat.Mesh(singleTriangleFacingUp(), []trs.TRS{
		trs.Identity(),
		trs.New(vector3.New(5., 0., 0.), quarterTurnZ, vector3.One[float64]()),
	})

	normals := result.Float3Attribute(modeling.NormalAttribute)
	require.Equal(t, 6, normals.Len(), "expected two 3-vertex copies")

	for i := 0; i < 3; i++ {
		assert.InDeltaf(t, 1, normals.At(i).Y(), 1e-9, "untransformed copy's normal %d should still point up", i)
	}
	for i := 3; i < 6; i++ {
		n := normals.At(i)
		assert.InDeltaf(t, -1, n.X(), 1e-9, "rotated copy's normal %d should point -X", i)
		assert.InDeltaf(t, 0, n.Y(), 1e-9, "rotated copy's normal %d should no longer point up", i)
	}
}

// TestMeshDoesNotTranslateNormals pins the other half of the rule: a
// normal is a direction, so a copy that is only moved keeps its
// orientation and stays unit length.
func TestMeshDoesNotTranslateNormals(t *testing.T) {
	result := repeat.Mesh(singleTriangleFacingUp(), []trs.TRS{
		trs.Position(vector3.New(100., -50., 25.)),
	})

	normals := result.Float3Attribute(modeling.NormalAttribute)
	require.Equal(t, 3, normals.Len())
	for i := 0; i < normals.Len(); i++ {
		n := normals.At(i)
		assert.InDeltaf(t, 1, n.Y(), 1e-9, "translation must not tilt normal %d", i)
		assert.InDeltaf(t, 1, n.Length(), 1e-9, "translation must not stretch normal %d", i)
	}
}

// TestMeshWithoutNormalsStillRepeats guards the other direction: meshes
// carrying no normal attribute must keep working, not start erroring.
func TestMeshWithoutNormalsStillRepeats(t *testing.T) {
	noNormals := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(modeling.PositionAttribute, []vector3.Float64{
			vector3.New(0., 0., 0.),
			vector3.New(1., 0., 0.),
			vector3.New(0., 0., 1.),
		})

	require.NotPanics(t, func() {
		result := repeat.Mesh(noNormals, []trs.TRS{trs.Identity(), trs.Identity()})
		assert.False(t, result.HasFloat3Attribute(modeling.NormalAttribute))
		assert.Equal(t, 6, result.Float3Attribute(modeling.PositionAttribute).Len())
	})
}
