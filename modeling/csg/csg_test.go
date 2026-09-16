package csg_test

import (
	"crypto/sha256"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/csg"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cube(center vector3.Float64, side float64) modeling.Mesh {
	return primitives.Cube{Height: side, Width: side, Depth: side}.
		UnweldedQuads().
		Translate(center)
}

// The divergence theorem only gives the enclosed volume for a closed surface,
// so this and requireWatertight are checked together or neither means much.
func volume(m modeling.Mesh) float64 {
	indices := m.Indices()
	positions := m.Float3Attribute(modeling.PositionAttribute)

	total := 0.
	for i := 0; i+2 < indices.Len(); i += 3 {
		a := positions.At(indices.At(i))
		b := positions.At(indices.At(i + 1))
		c := positions.At(indices.At(i + 2))
		total += a.Dot(b.Cross(c)) / 6
	}
	return total
}

type positionKey [3]int64

func quantize(v vector3.Float64) positionKey {
	round := func(f float64) int64 { return int64(math.Round(f * 1e6)) }
	return positionKey{round(v.X()), round(v.Y()), round(v.Z())}
}

// Faces are emitted unwelded, so a seam only shows up by position. An edge
// used once is a crack; used twice, the two sides agree on where they meet.
func requireWatertight(t *testing.T, m modeling.Mesh) {
	t.Helper()

	indices := m.Indices()
	positions := m.Float3Attribute(modeling.PositionAttribute)
	require.Positive(t, indices.Len(), "mesh is empty")

	edges := map[[2]positionKey]int{}
	for i := 0; i+2 < indices.Len(); i += 3 {
		corners := [3]vector3.Float64{
			positions.At(indices.At(i)),
			positions.At(indices.At(i + 1)),
			positions.At(indices.At(i + 2)),
		}
		for k := 0; k < 3; k++ {
			e := [2]positionKey{quantize(corners[k]), quantize(corners[(k+1)%3])}
			if e[1][0] < e[0][0] ||
				(e[1][0] == e[0][0] && e[1][1] < e[0][1]) ||
				(e[1][0] == e[0][0] && e[1][1] == e[0][1] && e[1][2] < e[0][2]) {
				e[0], e[1] = e[1], e[0]
			}
			edges[e]++
		}
	}

	histogram := map[int]int{}
	for _, count := range edges {
		histogram[count]++
	}
	require.Equalf(t, map[int]int{2: len(edges)}, histogram,
		"every edge should be shared by exactly two triangles")
}

func TestOverlappingCubes(t *testing.T) {
	a := cube(vector3.Zero[float64](), 2)
	b := cube(vector3.New(1., 1., 1.), 2)

	for name, tc := range map[string]struct {
		op   func(x, y modeling.Mesh) (modeling.Mesh, error)
		want float64
	}{
		"union":        {csg.Union, 15},
		"intersection": {csg.Intersect, 1},
		"subtraction":  {csg.Subtract, 7},
	} {
		t.Run(name, func(t *testing.T) {
			mesh, err := tc.op(a, b)
			require.NoError(t, err)
			requireWatertight(t, mesh)
			assert.InDelta(t, tc.want, volume(mesh), 1e-9)
		})
	}
}

func TestDisjointSolidsAreLeftAlone(t *testing.T) {
	a := cube(vector3.Zero[float64](), 2)
	b := cube(vector3.New(10., 0., 0.), 2)

	joined, err := csg.Union(a, b)
	require.NoError(t, err)
	requireWatertight(t, joined)
	assert.InDelta(t, 16., volume(joined), 1e-9)

	carved, err := csg.Subtract(a, b)
	require.NoError(t, err)
	requireWatertight(t, carved)
	assert.InDelta(t, 8., volume(carved), 1e-9)

	shared, err := csg.Intersect(a, b)
	require.NoError(t, err)
	assert.Zero(t, shared.Indices().Len(), "nothing is shared")
}

func TestFullyNestedSolids(t *testing.T) {
	outer := cube(vector3.Zero[float64](), 2)
	inner := cube(vector3.Zero[float64](), 1)

	joined, err := csg.Union(outer, inner)
	require.NoError(t, err)
	requireWatertight(t, joined)
	assert.InDelta(t, 8., volume(joined), 1e-9)

	shared, err := csg.Intersect(outer, inner)
	require.NoError(t, err)
	requireWatertight(t, shared)
	assert.InDelta(t, 1., volume(shared), 1e-9)

	// A shell with a sealed cavity: still closed, and the cavity's negative
	// contribution is what leaves 7 rather than 8.
	hollow, err := csg.Subtract(outer, inner)
	require.NoError(t, err)
	requireWatertight(t, hollow)
	assert.InDelta(t, 7., volume(hollow), 1e-9)
}

// Section 4 never splits coplanar pairs, and section 9 always takes a shared
// face from the first solid. Two cubes meeting flush exercise both.
func TestSolidsMeetingOnASharedFace(t *testing.T) {
	a := cube(vector3.Zero[float64](), 2)
	b := cube(vector3.New(2., 0., 0.), 2)

	joined, err := csg.Union(a, b)
	require.NoError(t, err)
	requireWatertight(t, joined)
	assert.InDelta(t, 16., volume(joined), 1e-9)
}

// Whatever the operations do to a curved surface, the part kept and the part
// carved away have to add back up to what was there.
func TestCarvedAndKeptPartsAccountForTheWhole(t *testing.T) {
	box := cube(vector3.Zero[float64](), 2)
	ball := primitives.UVSphere(1.3, 16, 24)

	carved, err := csg.Subtract(box, ball)
	require.NoError(t, err)
	requireWatertight(t, carved)

	shared, err := csg.Intersect(box, ball)
	require.NoError(t, err)
	requireWatertight(t, shared)

	assert.InDelta(t, volume(box), volume(carved)+volume(shared), 1e-9)
}

func TestUnionAndIntersectionAccountForBothSolids(t *testing.T) {
	a := primitives.UVSphere(1, 16, 24)
	b := primitives.UVSphere(1, 16, 24).Translate(vector3.New(1., 0., 0.))

	joined, err := csg.Union(a, b)
	require.NoError(t, err)
	requireWatertight(t, joined)

	shared, err := csg.Intersect(a, b)
	require.NoError(t, err)
	requireWatertight(t, shared)

	assert.InDelta(t, volume(a)+volume(b), volume(joined)+volume(shared), 1e-9)
}

func TestRejectsWhatItCannotOperateOn(t *testing.T) {
	solid := cube(vector3.Zero[float64](), 2)
	empty := modeling.EmptyMesh(modeling.TriangleTopology)

	_, err := csg.Union(solid, empty)
	require.Error(t, err)

	_, err = csg.Union(empty, solid)
	require.Error(t, err)

	_, err = csg.Union(solid, modeling.EmptyMesh(modeling.PointTopology))
	require.Error(t, err)
}

func meshFromNode[T any](t *testing.T, data T) modeling.Mesh {
	t.Helper()
	return nodes.GetNodeOutputPort[modeling.Mesh](&nodes.Struct[T]{Data: data}, "Out").Value()
}

func constMeshes(meshes ...modeling.Mesh) []nodes.Output[modeling.Mesh] {
	out := make([]nodes.Output[modeling.Mesh], len(meshes))
	for i, m := range meshes {
		out[i] = nodes.ConstOutput[modeling.Mesh]{Val: m}
	}
	return out
}

func TestNodesFoldOverEveryInput(t *testing.T) {
	a := cube(vector3.Zero[float64](), 2)
	b := cube(vector3.New(1., 1., 1.), 2)
	c := cube(vector3.New(-1., -1., -1.), 2)

	t.Run("union", func(t *testing.T) {
		mesh := meshFromNode(t, csg.UnionNode{Meshes: constMeshes(a, b)})
		requireWatertight(t, mesh)
		assert.InDelta(t, 15., volume(mesh), 1e-9)
	})

	t.Run("intersection", func(t *testing.T) {
		mesh := meshFromNode(t, csg.IntersectionNode{Meshes: constMeshes(a, b)})
		requireWatertight(t, mesh)
		assert.InDelta(t, 1., volume(mesh), 1e-9)
	})

	t.Run("subtraction removes each in turn", func(t *testing.T) {
		mesh := meshFromNode(t, csg.SubtractNode{
			Base:   nodes.ConstOutput[modeling.Mesh]{Val: a},
			Remove: constMeshes(b, c),
		})
		requireWatertight(t, mesh)
		assert.InDelta(t, 6., volume(mesh), 1e-9)
	})

	t.Run("a single mesh passes through", func(t *testing.T) {
		mesh := meshFromNode(t, csg.UnionNode{Meshes: constMeshes(a)})
		assert.InDelta(t, 8., volume(mesh), 1e-9)
	})

	t.Run("no meshes gives nothing", func(t *testing.T) {
		mesh := meshFromNode(t, csg.UnionNode{})
		assert.Zero(t, mesh.Indices().Len())
	})
}

// Results have to be usable as inputs, which is what makes the operations
// worth having. Four in a row, each one cutting into the last.
func TestOperationsChain(t *testing.T) {
	box := cube(vector3.Zero[float64](), 2)
	ball := primitives.UVSphere(1.28, 20, 32)

	result, err := csg.Intersect(box, ball)
	require.NoError(t, err)
	requireWatertight(t, result)

	for _, turn := range []quaternion.Quaternion{
		quaternion.New(vector3.Zero[float64](), 1),
		quaternion.FromTheta(math.Pi/2, vector3.Forward[float64]()),
		quaternion.FromTheta(math.Pi/2, vector3.Right[float64]()),
	} {
		rod := primitives.Cylinder{Sides: 28, Height: 4, Radius: 0.62}.ToMesh().Rotate(turn)

		result, err = csg.Subtract(result, rod)
		require.NoError(t, err)
		requireWatertight(t, result)
	}

	assert.Positive(t, volume(result))
	assert.Less(t, volume(result), 8.)
}

// Two meshes with the same geometry but a different face order do not stay
// the same downstream: face order decides which of two coincident points a
// cut snaps to, and how the octree splits. So the order has to be settled,
// not just the geometry, or chaining these gives a different mesh each run.
func fingerprint(m modeling.Mesh) string {
	indices := m.Indices()
	positions := m.Float3Attribute(modeling.PositionAttribute)

	var out strings.Builder
	for i := 0; i < indices.Len(); i++ {
		v := positions.At(indices.At(i))
		fmt.Fprintf(&out, "%.12f,%.12f,%.12f;", v.X(), v.Y(), v.Z())
	}
	sum := sha256.Sum256([]byte(out.String()))
	return fmt.Sprintf("%x", sum[:8])
}

func TestTheSameInputGivesTheSameMesh(t *testing.T) {
	box := cube(vector3.Zero[float64](), 2)
	ball := primitives.UVSphere(1.28, 20, 32)

	first := ""
	for run := 0; run < 8; run++ {
		result, err := csg.Intersect(box, ball)
		require.NoError(t, err)

		for _, turn := range []quaternion.Quaternion{
			quaternion.New(vector3.Zero[float64](), 1),
			quaternion.FromTheta(math.Pi/2, vector3.Forward[float64]()),
		} {
			rod := primitives.Cylinder{Sides: 28, Height: 4, Radius: 0.62}.ToMesh().Rotate(turn)
			result, err = csg.Subtract(result, rod)
			require.NoError(t, err)
		}

		got := fingerprint(result)
		if run == 0 {
			first = got
			continue
		}
		require.Equalf(t, first, got, "run %d produced a different mesh", run)
	}
}

// A cylinder without its lid, and a single loose triangle: both enclose
// nothing, and section 7 would happily classify against them and return
// nonsense rather than complain.
func openMeshes() map[string]modeling.Mesh {
	loose := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: {
				vector3.New(0., 0., 0.),
				vector3.New(1., 0., 0.),
				vector3.New(0., 1., 0.),
			},
		})

	return map[string]modeling.Mesh{
		"a lidless cylinder": primitives.Cylinder{
			Sides: 12, Height: 2, Radius: 1, NoTop: true,
		}.ToMesh(),
		"a single triangle": loose,
	}
}

func TestCheckClosedAcceptsSolids(t *testing.T) {
	for name, solid := range map[string]modeling.Mesh{
		"cube":     cube(vector3.Zero[float64](), 2),
		"sphere":   primitives.UVSphere(1, 16, 24),
		"cylinder": primitives.Cylinder{Sides: 12, Height: 2, Radius: 1}.ToMesh(),
	} {
		t.Run(name, func(t *testing.T) {
			assert.NoError(t, csg.CheckClosed(solid))
		})
	}
}

func TestCheckClosedRejectsOpenMeshes(t *testing.T) {
	for name, open := range openMeshes() {
		t.Run(name, func(t *testing.T) {
			err := csg.CheckClosed(open)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not a closed solid")
		})
	}
}

func TestOperationsRefuseOpenMeshes(t *testing.T) {
	solid := cube(vector3.Zero[float64](), 2)

	for name, open := range openMeshes() {
		t.Run(name, func(t *testing.T) {
			_, err := csg.Union(solid, open)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "second mesh is not a closed solid")

			_, err = csg.Subtract(open, solid)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "first mesh is not a closed solid")
		})
	}
}

func errorsFromNode[T any](t *testing.T, data T) []string {
	t.Helper()
	port := nodes.GetNodeOutputPort[modeling.Mesh](&nodes.Struct[T]{Data: data}, "Out")
	port.Value()

	observable, ok := port.(nodes.ObservableExecution)
	require.True(t, ok, "output port does not report execution")
	return observable.ExecutionReport().Errors
}

// The editor only shows what the node reports, so the check is worth little
// unless it comes back out through the port naming the input at fault.
func TestNodesReportAnOpenMesh(t *testing.T) {
	solid := cube(vector3.Zero[float64](), 2)
	open := primitives.Cylinder{Sides: 12, Height: 2, Radius: 1, NoTop: true}.ToMesh()

	t.Run("union names the offending input", func(t *testing.T) {
		errs := errorsFromNode(t, csg.UnionNode{
			Meshes: constMeshes(solid, solid, open, solid),
		})
		require.Len(t, errs, 1)
		assert.Contains(t, errs[0], "Meshes.2")
		assert.Contains(t, errs[0], "not a closed solid")
	})

	t.Run("every offending input is named", func(t *testing.T) {
		errs := errorsFromNode(t, csg.IntersectionNode{
			Meshes: constMeshes(open, solid, open),
		})
		require.Len(t, errs, 2)
		assert.Contains(t, errs[0], "Meshes.0")
		assert.Contains(t, errs[1], "Meshes.2")
	})

	t.Run("subtraction names base and removals apart", func(t *testing.T) {
		errs := errorsFromNode(t, csg.SubtractNode{
			Base:   nodes.ConstOutput[modeling.Mesh]{Val: open},
			Remove: constMeshes(solid, open),
		})
		require.Len(t, errs, 2)
		assert.Contains(t, errs[0], "Base")
		assert.Contains(t, errs[1], "Remove.1")
	})

	t.Run("solids report nothing", func(t *testing.T) {
		errs := errorsFromNode(t, csg.UnionNode{Meshes: constMeshes(solid, solid)})
		assert.Empty(t, errs)
	})
}

// Meshes arrive from all sorts of generators, and an empty triangle among
// them should change nothing: it covers no surface, and its normal is
// meaningless to steer a classification ray by.
func TestEmptyTrianglesAreIgnored(t *testing.T) {
	clean := cube(vector3.Zero[float64](), 2)

	indices := clean.Indices()
	positions := clean.Float3Attribute(modeling.PositionAttribute)

	verts := make([]vector3.Float64, 0, positions.Len()+3)
	for i := 0; i < positions.Len(); i++ {
		verts = append(verts, positions.At(i))
	}
	verts = append(verts,
		vector3.New(0.4, 0.4, 0.4),
		vector3.New(-0.2, -0.2, -0.2),
		vector3.New(0.9, 0.9, 0.9), // all three on one line
	)

	tris := make([]int, 0, indices.Len()+12)
	for i := 0; i < indices.Len(); i++ {
		tris = append(tris, indices.At(i))
	}
	collapsed := positions.Len()
	tris = append(tris,
		collapsed, collapsed+1, collapsed+2, // collinear, no area
		collapsed, collapsed, collapsed+1, // two corners the same
	)

	littered := modeling.NewTriangleMesh(tris).
		SetFloat3Data(map[string][]vector3.Float64{modeling.PositionAttribute: verts})

	require.NoError(t, csg.CheckClosed(littered),
		"empty triangles should not make a closed mesh look open")

	other := cube(vector3.New(1., 1., 1.), 2)

	fromClean, err := csg.Union(clean, other)
	require.NoError(t, err)

	fromLittered, err := csg.Union(littered, other)
	require.NoError(t, err)

	requireWatertight(t, fromLittered)
	assert.InDelta(t, volume(fromClean), volume(fromLittered), 1e-9)
}

// Two copies of one solid offset along an axis stay exactly symmetric, which
// aims a face's classification ray straight down an edge two triangles share.
// The ray then sits a rounding error outside both of them, gets rejected by
// each in turn, and the crossing disappears: the face reads as outside and
// leaves a hole where it should have been kept. Offsetting diagonally breaks
// the symmetry and hides the whole thing, so the axis is the test.
func TestSubtractingASolidFromItselfOffset(t *testing.T) {
	sphere := primitives.QuadSphere(0.5, primitives.Cube{
		Width: 1, Height: 1, Depth: 1, Dimensions: 10,
	}, false, true)

	for _, offset := range []float64{0.025, 0.1, 0.25, 0.5546622379772047, 0.75} {
		t.Run(fmt.Sprintf("offset %.4f", offset), func(t *testing.T) {
			carved, err := csg.Subtract(sphere, sphere.Translate(vector3.New(0., offset, 0.)))
			require.NoError(t, err)
			requireWatertight(t, carved)
			assert.Positive(t, volume(carved))
		})
	}
}
