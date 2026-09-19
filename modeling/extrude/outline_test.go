package extrude_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	unitSquare = []vector2.Float64{
		vector2.New(-1., -1.), vector2.New(1., -1.),
		vector2.New(1., 1.), vector2.New(-1., 1.),
	}
	straightUp = []vector3.Float64{
		vector3.New(0., 0., 0.), vector3.New(0., 3., 0.),
	}
)

func signedVolume(m modeling.Mesh) float64 {
	idx := m.Indices()
	pos := m.Float3Attribute(modeling.PositionAttribute)
	total := 0.
	for i := 0; i < idx.Len(); i += 3 {
		a, b, c := pos.At(idx.At(i)), pos.At(idx.At(i+1)), pos.At(idx.At(i+2))
		total += a.Dot(b.Cross(c)) / 6
	}
	return total
}

// Welded first because caps carry their own vertices to keep their normals
// flat, so a seam between cap and shell only shows up by position. Zero
// unpaired edges after that means every edge has a face on each side, wound
// opposite ways.
func requireWatertight(t *testing.T, m modeling.Mesh) {
	t.Helper()
	require.Empty(t, meshops.Weld(m, modeling.PositionAttribute, 1e-9).BoundaryEdges())
}

func TestOutlineExtrudesAWatertightSolid(t *testing.T) {
	mesh, err := extrude.Outline{Shape: unitSquare, Path: straightUp}.Extrude()
	require.NoError(t, err)

	requireWatertight(t, mesh)
	assert.InDelta(t, 12., signedVolume(mesh), 1e-9)
}

// A triangle fan over the outline would spill outside a concave shape, and
// the volume is what catches it.
func TestOutlineCapsAConcaveShape(t *testing.T) {
	l := []vector2.Float64{
		vector2.New(0., 0.), vector2.New(2., 0.), vector2.New(2., 1.),
		vector2.New(1., 1.), vector2.New(1., 2.), vector2.New(0., 2.),
	}

	mesh, err := extrude.Outline{Shape: l, Path: straightUp}.Extrude()
	require.NoError(t, err)

	requireWatertight(t, mesh)
	assert.InDelta(t, 9., signedVolume(mesh), 1e-9)
}

func TestOutlineIgnoresTheWindingItWasGiven(t *testing.T) {
	backwards := make([]vector2.Float64, len(unitSquare))
	for i, p := range unitSquare {
		backwards[len(unitSquare)-1-i] = p
	}

	mesh, err := extrude.Outline{Shape: backwards, Path: straightUp}.Extrude()
	require.NoError(t, err)

	requireWatertight(t, mesh)
	assert.InDelta(t, 12., signedVolume(mesh), 1e-9)
}

func TestOutlineFollowsABentPath(t *testing.T) {
	mesh, err := extrude.Outline{
		Shape: unitSquare,
		Path: []vector3.Float64{
			vector3.New(0., 0., 0.),
			vector3.New(0., 2., 0.),
			vector3.New(1.5, 3.5, 0.),
		},
	}.Extrude()
	require.NoError(t, err)

	requireWatertight(t, mesh)
	assert.Positive(t, signedVolume(mesh))
}

func TestOutlineLeavesAClosedPathUncapped(t *testing.T) {
	small := []vector2.Float64{
		vector2.New(-.3, -.3), vector2.New(.3, -.3),
		vector2.New(.3, .3), vector2.New(-.3, .3),
	}

	path := []vector3.Float64{
		vector3.New(2., 0., 0.),
		vector3.New(-1., 0., 1.7),
		vector3.New(-1., 0., -1.7),
	}

	mesh, err := extrude.Outline{Shape: small, Path: path, Closed: true}.Extrude()
	require.NoError(t, err)

	requireWatertight(t, mesh)
	// Two vertices per outline edge per ring, and nothing beyond that: a cap
	// would carry its own.
	assert.Equal(t, len(path)*len(small)*2,
		mesh.Float3Attribute(modeling.PositionAttribute).Len())
}

func TestOutlineGivesEveryTriangleANormal(t *testing.T) {
	mesh, err := extrude.Outline{Shape: unitSquare, Path: straightUp}.Extrude()
	require.NoError(t, err)

	pos := mesh.Float3Attribute(modeling.PositionAttribute)
	normals := mesh.Float3Attribute(modeling.NormalAttribute)
	require.Equal(t, pos.Len(), normals.Len())

	idx := mesh.Indices()
	for i := 0; i < idx.Len(); i += 3 {
		a, b, c := pos.At(idx.At(i)), pos.At(idx.At(i+1)), pos.At(idx.At(i+2))
		face := b.Sub(a).Cross(c.Sub(a)).Normalized()
		for _, v := range []int{idx.At(i), idx.At(i + 1), idx.At(i + 2)} {
			assert.Positivef(t, face.Dot(normals.At(v)),
				"vertex %d carries a normal facing away from its triangle", v)
		}
	}
}

func TestOutlineRejectsWhatItCannotSweep(t *testing.T) {
	tests := map[string]extrude.Outline{
		"too few outline points": {Shape: unitSquare[:2], Path: straightUp},
		"too few path points":    {Shape: unitSquare, Path: straightUp[:1]},
		"outline with no area": {Shape: []vector2.Float64{
			vector2.New(0., 0.), vector2.New(1., 0.), vector2.New(2., 0.),
		}, Path: straightUp},
		"path that repeats a point": {Shape: unitSquare, Path: []vector3.Float64{
			vector3.New(0., 0., 0.), vector3.New(0., 0., 0.),
		}},
	}

	for name, outline := range tests {
		t.Run(name, func(t *testing.T) {
			mesh, err := outline.Extrude()
			require.Error(t, err)
			assert.Zero(t, mesh.Indices().Len())
		})
	}
}

// A straight path makes the cross product of neighbouring segments vanish,
// which is what used to leave the frame undefined.
func TestOutlineSweepsAStraightPath(t *testing.T) {
	collinear := []vector3.Float64{
		vector3.New(0., 0., 0.),
		vector3.New(0., 1.3001528736574441, -0.37084847944351573),
		vector3.New(0., 2.6003057473148883, -0.7416969588870315),
	}

	mesh, err := extrude.Outline{Shape: unitSquare, Path: collinear}.Extrude()
	require.NoError(t, err)

	requireWatertight(t, mesh)
	length := collinear[2].Sub(collinear[0]).Length()
	assert.InDelta(t, 4*length, signedVolume(mesh), 1e-9)
}

// Carrying the frame forward is what keeps a straight run untwisted. Picking
// a fresh perpendicular at each point would rotate the shell around its own
// axis between rings.
func TestOutlineDoesNotTwistAlongAStraightRun(t *testing.T) {
	path := []vector3.Float64{}
	for i := 0; i <= 6; i++ {
		path = append(path, vector3.New(0., float64(i)*0.5, 0.))
	}

	mesh, err := extrude.Outline{Shape: unitSquare, Path: path}.Extrude()
	require.NoError(t, err)
	requireWatertight(t, mesh)
	assert.InDelta(t, 4*3., signedVolume(mesh), 1e-9)

	pos := mesh.Float3Attribute(modeling.PositionAttribute)
	perRing := len(unitSquare) * 2
	for ring := 1; ring < len(path); ring++ {
		for vertex := 0; vertex < perRing; vertex++ {
			moved := pos.At(ring*perRing + vertex).Sub(pos.At(vertex))
			expected := path[ring].Sub(path[0])
			assert.InDeltaf(t, 0, moved.Sub(expected).Length(), 1e-9,
				"ring %d vertex %d drifted off the sweep axis", ring, vertex)
		}
	}
}

func TestOutlineHandlesAStraightRunMeetingABend(t *testing.T) {
	mesh, err := extrude.Outline{
		Shape: unitSquare,
		Path: []vector3.Float64{
			vector3.New(0., 0., 0.), vector3.New(0., 1., 0.), vector3.New(0., 2., 0.),
			vector3.New(1., 3., 0.), vector3.New(2.5, 3.5, 0.),
		},
	}.Extrude()
	require.NoError(t, err)

	requireWatertight(t, mesh)
	assert.Positive(t, signedVolume(mesh))
}

func TestShapeSweepsAStraightPathWithoutNaN(t *testing.T) {
	mesh := extrude.Shape(unitSquare, []vector3.Float64{
		vector3.New(0., 0., 0.), vector3.New(0., 1., 0.), vector3.New(0., 2., 0.),
	})

	pos := mesh.Float3Attribute(modeling.PositionAttribute)
	require.Positive(t, pos.Len())
	for i := 0; i < pos.Len(); i++ {
		v := pos.At(i)
		require.Falsef(t, math.IsNaN(v.X()) || math.IsNaN(v.Y()) || math.IsNaN(v.Z()),
			"vertex %d is NaN", i)
	}
}

// An outline swept straight up has to keep the x and z it was handed. The
// triangulator lays 2D out as (x, 0, y), and anything else leaves the solid
// mirrored against the points the caller actually placed.
func TestOutlineKeepsTheFootprintItWasGiven(t *testing.T) {
	outline := []vector2.Float64{
		vector2.New(-1.2606695213856098, 0.6803643854296824),
		vector2.New(1.8964066195723626, -1.040857292644666),
		vector2.New(3.562420954138031, 0.25855444574592834),
		vector2.New(1.2491260267524762, 1.0696235172856896),
		vector2.New(0.8549510226488257, 2.74853141173689),
		vector2.New(3.073117641129813, 4.181240878601283),
		vector2.New(1.408708789567994, 5.02682553215862),
		vector2.New(-0.22301784448859663, 4.248879645678277),
	}

	mesh, err := extrude.Outline{
		Shape: outline,
		Path:  []vector3.Float64{vector3.New(0., 0., 0.), vector3.New(0., 2., 0.)},
	}.Extrude()
	require.NoError(t, err)

	pos := mesh.Float3Attribute(modeling.PositionAttribute)
	for _, p := range outline {
		want := vector3.New(p.X(), 0, p.Y())
		found := false
		for i := 0; i < pos.Len(); i++ {
			if pos.At(i).Sub(want).Length() < 1e-9 {
				found = true
				break
			}
		}
		assert.Truef(t, found, "outline point %v does not appear at %v", p, want)
	}
}

// The cap is a triangulation of the outline, so it has to sit exactly where
// triangulating that outline directly would put it.
func TestOutlineCapMatchesTheTriangulator(t *testing.T) {
	outline := []vector2.Float64{
		vector2.New(0., 0.), vector2.New(3., -1.), vector2.New(4., 1.5),
		vector2.New(1.5, 1.), vector2.New(1., 3.),
	}

	flat, err := triangulation.ConstrainedDelaunay(outline,
		[]triangulation.Constraint{triangulation.NewConstraint(outline)})
	require.NoError(t, err)

	solid, err := extrude.Outline{
		Shape: outline,
		Path:  []vector3.Float64{vector3.New(0., 0., 0.), vector3.New(0., 2., 0.)},
	}.Extrude()
	require.NoError(t, err)

	footprint := func(m modeling.Mesh) (lo, hi vector2.Float64) {
		p := m.Float3Attribute(modeling.PositionAttribute)
		lo = vector2.New(math.Inf(1), math.Inf(1))
		hi = vector2.New(math.Inf(-1), math.Inf(-1))
		for i := 0; i < p.Len(); i++ {
			v := p.At(i)
			if math.Abs(v.Y()) > 1e-9 {
				continue
			}
			lo = vector2.New(math.Min(lo.X(), v.X()), math.Min(lo.Y(), v.Z()))
			hi = vector2.New(math.Max(hi.X(), v.X()), math.Max(hi.Y(), v.Z()))
		}
		return
	}

	flatLo, flatHi := footprint(flat)
	solidLo, solidHi := footprint(solid)
	assert.InDelta(t, 0, solidLo.Sub(flatLo).Length(), 1e-9, "cap starts somewhere else")
	assert.InDelta(t, 0, solidHi.Sub(flatHi).Length(), 1e-9, "cap ends somewhere else")
}

// The segment that closes a loop must not be geometrically special. Wringing
// the outline round at the seam drags every vertex much further across that
// one segment than any other, which a wide profile makes plain.
func TestOutlineClosesALoopWithoutWringingTheSeam(t *testing.T) {
	ribbon := []vector2.Float64{
		vector2.New(-1.2, -0.12), vector2.New(1.2, -0.12),
		vector2.New(1.2, 0.12), vector2.New(-1.2, 0.12),
	}
	trefoil := []vector3.Float64{}
	for i := 0; i < 24; i++ {
		a := math.Pi * 2 * float64(i) / 24
		trefoil = append(trefoil, vector3.New(
			math.Sin(a)+2*math.Sin(2*a), -math.Sin(3*a), math.Cos(a)-2*math.Cos(2*a)))
	}

	mesh, err := extrude.Outline{Shape: ribbon, Path: trefoil, Closed: true}.Extrude()
	require.NoError(t, err)
	requireWatertight(t, mesh)

	pos := mesh.Float3Attribute(modeling.PositionAttribute)
	perRing := len(ribbon) * 2

	travel := make([]float64, len(trefoil))
	total := 0.
	for ring := range trefoil {
		next := (ring + 1) % len(trefoil)
		for v := 0; v < perRing; v++ {
			moved := pos.At(next*perRing + v).Sub(pos.At(ring*perRing + v)).Length()
			travel[ring] = math.Max(travel[ring], moved)
		}
		total += travel[ring]
	}

	mean := total / float64(len(travel))
	for ring, moved := range travel {
		assert.Lessf(t, moved, 1.3*mean,
			"segment %d drags vertices %.3f against a mean of %.3f, so it absorbed the loop's twist",
			ring, moved, mean)
	}
}

func circleOutline(points int, radius float64) []vector2.Float64 {
	out := make([]vector2.Float64, points)
	for i := range out {
		a := math.Pi * 2 * float64(i) / float64(points)
		out[i] = vector2.New(math.Cos(a)*radius, math.Sin(a)*radius)
	}
	return out
}

func ringPath(points int, radius float64) []vector3.Float64 {
	out := make([]vector3.Float64, points)
	for i := range out {
		a := math.Pi * 2 * float64(i) / float64(points)
		out[i] = vector3.New(math.Cos(a)*radius, 0, math.Sin(a)*radius)
	}
	return out
}

// Normals that disagree where two vertices share a position are a facet edge.
func worstSplitNormal(t *testing.T, m modeling.Mesh) float64 {
	t.Helper()
	pos := m.Float3Attribute(modeling.PositionAttribute)
	nrm := m.Float3Attribute(modeling.NormalAttribute)

	worst := 0.
	for i := 0; i < pos.Len(); i++ {
		for j := i + 1; j < pos.Len(); j++ {
			if pos.At(i).Sub(pos.At(j)).Length() > 1e-9 {
				continue
			}
			worst = math.Max(worst, nrm.At(i).Angle(nrm.At(j)))
		}
	}
	return worst
}

// A spline sampled outline turns a fraction of a degree per sample. Splitting
// every one of those doubles the vertices and facets what should be a smooth
// sweep.
func TestOutlineMergesVerticesAlongASmoothOutline(t *testing.T) {
	outline := circleOutline(100, 0.6)
	path := ringPath(8, 3)

	mesh, err := extrude.Outline{Shape: outline, Path: path, Closed: true}.Extrude()
	require.NoError(t, err)
	requireWatertight(t, mesh)

	assert.Equal(t, len(path)*len(outline),
		mesh.Float3Attribute(modeling.PositionAttribute).Len(),
		"a smooth outline should contribute one vertex per ring per point")
	assert.InDelta(t, 0, worstSplitNormal(t, mesh), 1e-9,
		"a smooth outline should carry no split normals")
}

func TestOutlineKeepsCreasesAtRealCorners(t *testing.T) {
	square := []vector2.Float64{
		vector2.New(-.4, -.4), vector2.New(.4, -.4),
		vector2.New(.4, .4), vector2.New(-.4, .4),
	}
	path := ringPath(8, 3)

	mesh, err := extrude.Outline{Shape: square, Path: path, Closed: true}.Extrude()
	require.NoError(t, err)
	requireWatertight(t, mesh)

	assert.Equal(t, len(path)*len(square)*2,
		mesh.Float3Attribute(modeling.PositionAttribute).Len(),
		"every corner of a square should stay split")
	assert.InDelta(t, math.Pi/2, worstSplitNormal(t, mesh), 1e-9)
}

// A regular octagon turns exactly the default 45 degrees at every corner.
// Comparing straight against the threshold let rounding split some and merge
// others on a shape whose corners are all identical.
func TestOutlineTreatsIdenticalCornersIdentically(t *testing.T) {
	mesh, err := extrude.Outline{
		Shape: circleOutline(8, 0.6), Path: ringPath(8, 3), Closed: true,
	}.Extrude()
	require.NoError(t, err)

	perRing := mesh.Float3Attribute(modeling.PositionAttribute).Len() / 8
	assert.Contains(t, []int{8, 16}, perRing,
		"every corner is the same, so they should all split or none should")
}

func TestOutlineHonoursTheSmoothingAngle(t *testing.T) {
	outline := circleOutline(8, 0.6)
	path := ringPath(8, 3)

	for name, tc := range map[string]struct {
		angle   float64
		perRing int
	}{
		"nothing smooths":    {0, 16},
		"everything smooths": {180, 8},
	} {
		t.Run(name, func(t *testing.T) {
			angle := tc.angle
			mesh, err := extrude.Outline{
				Shape: outline, Path: path, Closed: true, SmoothingAngle: &angle,
			}.Extrude()
			require.NoError(t, err)
			requireWatertight(t, mesh)
			assert.Equal(t, len(path)*tc.perRing,
				mesh.Float3Attribute(modeling.PositionAttribute).Len())
		})
	}
}

// Repeating the first point at the end is the natural way to draw a loop, and
// it used to land two rings on the same spot instead of saying so.
func TestOutlineRejectsAPathThatRepeatsAPoint(t *testing.T) {
	path := ringPath(8, 3)

	mesh, err := extrude.Outline{
		Shape:  circleOutline(6, 0.4),
		Path:   append(append([]vector3.Float64{}, path...), path[0]),
		Closed: true,
	}.Extrude()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "same point")
	assert.Zero(t, mesh.Indices().Len())
}

func TestOutlineIgnoresRepeatedPoints(t *testing.T) {
	clean, err := extrude.Outline{Shape: unitSquare, Path: straightUp}.Extrude()
	require.NoError(t, err)

	for name, shape := range map[string][]vector2.Float64{
		"closing point repeated": append(append([]vector2.Float64{}, unitSquare...), unitSquare[0]),
		"corner repeated thrice": {
			unitSquare[0], unitSquare[1], unitSquare[1], unitSquare[1], unitSquare[2], unitSquare[3],
		},
	} {
		t.Run(name, func(t *testing.T) {
			mesh, err := extrude.Outline{Shape: shape, Path: straightUp}.Extrude()
			require.NoError(t, err)

			assert.Equal(t, clean.PrimitiveCount(), mesh.PrimitiveCount())
			requireWatertight(t, mesh)
			requireNormalsAgreeWithFaces(t, mesh)
		})
	}

	_, err = extrude.Outline{
		Shape: []vector2.Float64{unitSquare[0], unitSquare[0], unitSquare[1], unitSquare[1]},
		Path:  straightUp,
	}.Extrude()
	require.Error(t, err, "two distinct points do not make an outline")
}
