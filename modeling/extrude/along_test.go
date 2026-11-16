package extrude_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/sketch"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func areaOf(t *testing.T, outlines []geometry.Shape, holes []geometry.Shape) sketch.Area {
	t.Helper()
	area, err := sketch.Fill(outlines, holes)
	require.NoError(t, err)
	return area
}

func framesAlong(path []vector3.Float64, closed bool) []trs.TRS {
	return geometry.FrameTRSs(geometry.PathFrames(path, closed))
}

var rise = []vector3.Float64{vector3.New(0., 0., 0.), vector3.New(0., 3., 0.)}

func TestAreaAlongPathSweepsAWatertightSolid(t *testing.T) {
	mesh, err := extrude.AreaAlongPath{
		Profile: areaOf(t, []geometry.Shape{square(2, vector2.Zero[float64]())}, nil),
		Frames:  framesAlong(rise, false),
	}.Extrude()
	require.NoError(t, err)

	requireWatertight(t, mesh)
	assert.InDelta(t, 12., signedVolume(mesh), 1e-9)
}

func TestAreaAlongPathKeepsTheProfilesHoles(t *testing.T) {
	mesh, err := extrude.AreaAlongPath{
		Profile: areaOf(t, []geometry.Shape{square(2, vector2.Zero[float64]())}, []geometry.Shape{square(1, vector2.Zero[float64]())}),
		Frames:  framesAlong(rise, false),
	}.Extrude()
	require.NoError(t, err)

	requireWatertight(t, mesh)
	assert.InDelta(t, 9., signedVolume(mesh), 1e-9)
}

func TestAreaAlongPathTakesEachFramesScale(t *testing.T) {
	frames := framesAlong(rise, false)
	frames[1] = frames[1].ScaleLocal(vector3.New(0.5, 0.5, 1))

	mesh, err := extrude.AreaAlongPath{
		Profile: areaOf(t, []geometry.Shape{square(2, vector2.Zero[float64]())}, nil),
		Frames:  frames,
	}.Extrude()
	require.NoError(t, err)

	requireWatertight(t, mesh)
	frustum := 3. / 3 * (4 + 1 + math.Sqrt(4*1))
	assert.InDelta(t, frustum, signedVolume(mesh), 1e-9)
}

func TestAreaAlongPathShadesAProfileStretchedByItsFrames(t *testing.T) {
	frames := framesAlong(rise, false)
	for i := range frames {
		frames[i] = frames[i].ScaleLocal(vector3.New(4., 1., 1.))
	}

	mesh, err := extrude.AreaAlongPath{
		Profile: areaOf(t, []geometry.Shape{circleOutline(64, 1)}, nil),
		Frames:  frames,
	}.Extrude()
	require.NoError(t, err)

	idx := mesh.Indices()
	pos := mesh.Float3Attribute(modeling.PositionAttribute)
	nrm := mesh.Float3Attribute(modeling.NormalAttribute)
	for i := 0; i < idx.Len(); i += 3 {
		a, b, c := pos.At(idx.At(i)), pos.At(idx.At(i+1)), pos.At(idx.At(i+2))
		face := b.Sub(a).Cross(c.Sub(a)).Normalized()
		for k := 0; k < 3; k++ {
			require.LessOrEqual(t, face.Angle(nrm.At(idx.At(i+k)))*180/math.Pi, 12.)
		}
	}
}

func TestAreaAlongPathShadesAFrameScaledToAPoint(t *testing.T) {
	frames := framesAlong(rise, false)
	frames[1] = frames[1].ScaleLocal(vector3.New(0., 0., 1.))

	mesh, err := extrude.AreaAlongPath{
		Profile: areaOf(t, []geometry.Shape{square(2, vector2.Zero[float64]())}, nil),
		Frames:  frames,
	}.Extrude()
	require.NoError(t, err)

	nrm := mesh.Float3Attribute(modeling.NormalAttribute)
	for i := 0; i < nrm.Len(); i++ {
		require.InDelta(t, 1, nrm.At(i).Length(), 1e-9)
	}
}

func requireNormalsWithin(t *testing.T, mesh modeling.Mesh, degrees float64) {
	t.Helper()

	idx := mesh.Indices()
	pos := mesh.Float3Attribute(modeling.PositionAttribute)
	nrm := mesh.Float3Attribute(modeling.NormalAttribute)
	for i := 0; i < idx.Len(); i += 3 {
		a, b, c := pos.At(idx.At(i)), pos.At(idx.At(i+1)), pos.At(idx.At(i+2))
		face := b.Sub(a).Cross(c.Sub(a))
		if face.Length() < 1e-12 {
			continue
		}
		for k := 0; k < 3; k++ {
			off := face.Normalized().Angle(nrm.At(idx.At(i+k))) * 180 / math.Pi
			require.LessOrEqualf(t, off, degrees, "vertex %d is shaded %.1f degrees off its face", idx.At(i+k), off)
		}
	}
}

func TestCircleShadesACone(t *testing.T) {
	mesh, err := extrude.Circle{Resolution: 32, Radii: []float64{1, 0}, Path: rise}.Extrude()
	require.NoError(t, err)
	requireNormalsWithin(t, mesh, 8)
}

func TestAreaAlongPathShadesATaper(t *testing.T) {
	frames := framesAlong(rise, false)
	frames[1] = frames[1].ScaleLocal(vector3.New(0.2, 0.2, 1))

	mesh, err := extrude.AreaAlongPath{
		Profile: areaOf(t, []geometry.Shape{circleOutline(64, 1)}, nil),
		Frames:  frames,
	}.Extrude()
	require.NoError(t, err)
	requireNormalsWithin(t, mesh, 8)
}

func TestAreaAlongPathCreasesWhereATaperStartsSharply(t *testing.T) {
	frames := framesAlong([]vector3.Float64{vector3.New(0., 0., 0.), vector3.New(0., 2., 0.), vector3.New(0., 2.5, 0.)}, false)
	frames[2] = frames[2].ScaleLocal(vector3.New(0., 0., 1.))

	mesh, err := extrude.AreaAlongPath{
		Profile: areaOf(t, []geometry.Shape{circleOutline(64, 1)}, nil),
		Frames:  frames,
	}.Extrude()
	require.NoError(t, err)
	requireNormalsWithin(t, mesh, 8)
}

func TestAreaAlongPathShadesFramesTiltedAgainstThePath(t *testing.T) {
	frames := make([]trs.TRS, 3)
	for i := range frames {
		frames[i] = trs.Position(vector3.New(float64(i)*2, 0., float64(i)*2))
	}

	mesh, err := extrude.AreaAlongPath{
		Profile: areaOf(t, []geometry.Shape{circleOutline(64, 1)}, nil),
		Frames:  frames,
	}.Extrude()
	require.NoError(t, err)
	requireNormalsWithin(t, mesh, 8)
}

func TestAreaAlongPathFacesOutwardThroughMirroredFrames(t *testing.T) {
	frames := framesAlong(rise, false)
	for i := range frames {
		frames[i] = frames[i].ScaleLocal(vector3.New(-1., 1., 1.))
	}

	mesh, err := extrude.AreaAlongPath{
		Profile: areaOf(t, []geometry.Shape{square(2, vector2.Zero[float64]())}, nil),
		Frames:  frames,
	}.Extrude()
	require.NoError(t, err)

	requireWatertight(t, mesh)
	assert.InDelta(t, 12., signedVolume(mesh), 1e-9)
	requireNormalsWithin(t, mesh, 1)
}

func TestAreaAlongPathKeepsEachSectionInItsFramesPlane(t *testing.T) {
	stations := []vector3.Float64{vector3.New(0., 0., 0.), vector3.New(0., 0., 2.), vector3.New(2., 0., 4.)}
	frames := make([]trs.TRS, len(stations))
	for i, at := range stations {
		frames[i] = trs.Position(at)
	}

	mesh, err := extrude.AreaAlongPath{
		Profile: areaOf(t, []geometry.Shape{square(1, vector2.Zero[float64]())}, nil),
		Frames:  frames,
	}.Extrude()
	require.NoError(t, err)

	pos := mesh.Float3Attribute(modeling.PositionAttribute)
	for i := 0; i < pos.Len(); i++ {
		z := pos.At(i).Z()
		require.InDeltaf(t, math.Round(z/2)*2, z, 1e-9, "vertex %d left its frame's plane", i)
	}
}

func TestAreaAlongPathMatchesAnOutlineSweptAlongTheSamePath(t *testing.T) {
	path := []vector3.Float64{
		vector3.New(0., 0., 0.), vector3.New(0., 2., 0.), vector3.New(1.5, 3., 0.5), vector3.New(4., 3., 1.), vector3.New(4.2, 3., 4.),
	}
	shape := []vector2.Float64{vector2.New(-0.3, -0.2), vector2.New(0.4, -0.2), vector2.New(0.4, 0.3), vector2.New(-0.3, 0.3)}

	outline, err := extrude.Outline{Shape: shape, Path: path}.Extrude()
	require.NoError(t, err)
	along, err := extrude.AreaAlongPath{
		Profile: areaOf(t, []geometry.Shape{shape}, nil),
		Frames:  framesAlong(path, false),
	}.Extrude()
	require.NoError(t, err)

	requireWatertight(t, along)
	assert.InDelta(t, signedVolume(outline), signedVolume(along), 1e-9)
	assert.InDelta(t, 0, outline.BoundingBox(modeling.PositionAttribute).Min().Distance(along.BoundingBox(modeling.PositionAttribute).Min()), 1e-9)
	assert.InDelta(t, 0, outline.BoundingBox(modeling.PositionAttribute).Max().Distance(along.BoundingBox(modeling.PositionAttribute).Max()), 1e-9)
}

func TestAreaAlongPathClosesALoopWithoutCaps(t *testing.T) {
	loop := []vector3.Float64{vector3.New(0., 0., 0.), vector3.New(4., 0., 0.), vector3.New(4., 0., 4.), vector3.New(0., 0., 4.)}
	mesh, err := extrude.AreaAlongPath{
		Profile: areaOf(t, []geometry.Shape{square(1, vector2.Zero[float64]())}, nil),
		Frames:  framesAlong(loop, true),
		Closed:  true,
	}.Extrude()
	require.NoError(t, err)

	requireWatertight(t, mesh)
	assert.InDelta(t, 16., signedVolume(mesh), 1e-9)
}

func TestAreaAlongPathNodeLaysOutUVsByDefault(t *testing.T) {
	mesh := nodes.GetNodeOutputPort[modeling.Mesh](&nodes.Struct[extrude.AreaAlongPathNode]{Data: extrude.AreaAlongPathNode{
		Profile: nodes.ConstOutput[sketch.Area]{Val: areaOf(t, []geometry.Shape{square(1, vector2.Zero[float64]())}, nil)},
		Frames:  nodes.ConstOutput[[]trs.TRS]{Val: framesAlong(rise, false)},
	}}, "Out").Value()

	require.NotZero(t, mesh.PrimitiveCount())
	assert.True(t, mesh.HasFloat2Attribute(modeling.TexCoordAttribute))
}
