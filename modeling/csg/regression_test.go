package csg_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/csg"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckClosedRejectsAFaceWoundAgainstItsNeighbours(t *testing.T) {
	indices := iter.ReadFull(cube(vector3.Zero[float64](), 2).Indices())
	indices[1], indices[2] = indices[2], indices[1]
	flipped := cube(vector3.Zero[float64](), 2).SetIndices(indices)

	err := csg.CheckClosed(flipped)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "wound the same way")

	_, err = csg.Union(cube(vector3.New(1., 1., 1.), 2), flipped)
	require.Error(t, err)
}

func TestCheckClosedCountsTheZeroAreaTrianglesItDropped(t *testing.T) {
	a := vector3.New(0., 0., 0.)
	b := vector3.New(1., 0., 0.)
	c := vector3.New(0., 1., 0.)
	d := vector3.New(0., 0., 1.)
	m := vector3.New(0.5, 0., 0.)

	// A tetrahedron with one face split at m and the T-junction sealed by a
	// zero-area filler, so every edge is shared by exactly two triangles.
	sealed := modeling.NewTriangleMesh([]int{
		0, 2, 1,
		0, 4, 3,
		4, 1, 3,
		0, 1, 4,
		0, 3, 2,
		1, 2, 3,
	}).SetFloat3Data(map[string][]vector3.Float64{
		modeling.PositionAttribute: {a, b, c, d, m},
	})

	err := csg.CheckClosed(sealed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "1 zero-area triangles were dropped")
	assert.Contains(t, err.Error(), "border a hole")
}

func TestToleranceIgnoresHowFarApartTheMeshesSit(t *testing.T) {
	near := cube(vector3.Zero[float64](), 1)
	far := cube(vector3.New(1e9, 0., 0.), 1)
	require.NoError(t, csg.CheckClosed(near))
	require.NoError(t, csg.CheckClosed(far))

	union, err := csg.Union(near, far)
	require.NoError(t, err)
	assert.Equal(t, 24, union.PrimitiveCount())
}

// A cube turned 45 degrees, placed so one vertical edge runs two nanometres
// inside the other cube's face: the intersection curve then has corners
// within tolerance of that face's edges without any faces being coplanar.
func TestCornersLandingJustOffAnEdgeLeaveNoHole(t *testing.T) {
	a := cube(vector3.Zero[float64](), 2)
	b := cube(vector3.Zero[float64](), 2).
		Rotate(quaternion.FromTheta(math.Pi/4, vector3.Up[float64]())).
		Translate(vector3.New(math.Sqrt2-1+2e-9, 0.3, 0.))

	union, err := csg.Union(a, b)
	require.NoError(t, err)
	require.NoError(t, csg.CheckClosed(union))

	intersection, err := csg.Intersect(a, b)
	require.NoError(t, err)
	require.NoError(t, csg.CheckClosed(intersection))

	difference, err := csg.Subtract(a, b)
	require.NoError(t, err)
	require.NoError(t, csg.CheckClosed(difference))

	assert.InDelta(t, 16., volume(union)+volume(intersection), 1e-6)
	assert.InDelta(t, 8., volume(difference)+volume(intersection), 1e-6)
}

func TestFacesWithinToleranceOfCoplanarCountAsCoplanar(t *testing.T) {
	a := cube(vector3.Zero[float64](), 2)

	for _, tilt := range []float64{0, 1e-10, 1e-9} {
		b := cube(vector3.Zero[float64](), 2).
			Rotate(quaternion.FromTheta(tilt, vector3.Right[float64]())).
			Translate(vector3.New(0., 2., 0.))

		union, err := csg.Union(a, b)
		require.NoErrorf(t, err, "tilt %g", tilt)
		require.NoErrorf(t, csg.CheckClosed(union), "tilt %g", tilt)
		assert.InDeltaf(t, 16., volume(union), 1e-6, "tilt %g", tilt)

		difference, err := csg.Subtract(a, b)
		require.NoErrorf(t, err, "tilt %g", tilt)
		require.NoErrorf(t, csg.CheckClosed(difference), "tilt %g", tilt)
		assert.InDeltaf(t, 8., volume(difference), 1e-6, "tilt %g", tilt)
	}
}

func TestNodesNameThePortAMeshArrivedOn(t *testing.T) {
	open := openMeshes()["a lidless cylinder"]

	errs := errorsFromNode(t, csg.UnionNode{
		Meshes: []nodes.Output[modeling.Mesh]{nil, nodes.ConstOutput[modeling.Mesh]{Val: open}},
	})
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0], "Meshes.1")

	errs = errorsFromNode(t, csg.SubtractNode{
		Base:   nodes.ConstOutput[modeling.Mesh]{Val: cube(vector3.Zero[float64](), 2)},
		Remove: []nodes.Output[modeling.Mesh]{nil, nodes.ConstOutput[modeling.Mesh]{Val: open}},
	})
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0], "Remove.1")
}
