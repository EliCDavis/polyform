package extrude_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lathedTube(t *testing.T) modeling.Mesh {
	t.Helper()

	return nodes.GetNodeOutputPort[modeling.Mesh](&nodes.Struct[extrude.ScrewNode]{
		Data: extrude.ScrewNode{
			Line: nodes.ConstOutput[[]vector3.Float64]{Val: []vector3.Float64{
				vector3.New(1., 0., 0.),
				vector3.New(1., 1., 0.),
			}},
			Segments:    nodes.ConstOutput[int]{Val: 16},
			Revolutions: nodes.ConstOutput[float64]{Val: 1},
		},
	}, "Out").Value()
}

func TestScrewEmitsNormals(t *testing.T) {
	mesh := lathedTube(t)

	require.True(t, mesh.HasFloat3Attribute(modeling.NormalAttribute), "screw output has no Normal attribute")

	normals := mesh.Float3Attribute(modeling.NormalAttribute)
	positions := mesh.Float3Attribute(modeling.PositionAttribute)
	require.Equal(t, positions.Len(), normals.Len(), "expected one normal per vertex")
	require.Greater(t, normals.Len(), 0)

	for i := 0; i < normals.Len(); i++ {
		assert.InDeltaf(t, 1, normals.At(i).Length(), 1e-9, "normal %d is not unit length", i)
	}
}

func lathe(line []vector3.Float64) modeling.Mesh {
	return nodes.GetNodeOutputPort[modeling.Mesh](&nodes.Struct[extrude.ScrewNode]{
		Data: extrude.ScrewNode{
			Line:        nodes.ConstOutput[[]vector3.Float64]{Val: line},
			Segments:    nodes.ConstOutput[int]{Val: 16},
			Revolutions: nodes.ConstOutput[float64]{Val: 1},
		},
	}, "Out").Value()
}

func meanOutwardFacing(mesh modeling.Mesh) float64 {
	positions := mesh.Float3Attribute(modeling.PositionAttribute)
	total := 0.
	for i := 0; i < mesh.PrimitiveCount(); i++ {
		tri := mesh.Tri(i)
		a, b, c := positions.At(tri.P1()), positions.At(tri.P2()), positions.At(tri.P3())
		centroid := a.Add(b).Add(c).Scale(1. / 3)
		total += b.Sub(a).Cross(c.Sub(a)).Normalized().Dot(vector3.New(centroid.X(), 0., centroid.Z()).Normalized())
	}
	return total / float64(mesh.PrimitiveCount())
}

func TestScrewFacesTheWayItsProfileRuns(t *testing.T) {
	up := []vector3.Float64{vector3.New(1., 0., 0.), vector3.New(1., 1., 0.)}
	down := []vector3.Float64{vector3.New(1., 1., 0.), vector3.New(1., 0., 0.)}

	assert.Greater(t, meanOutwardFacing(lathe(up)), 0.9)
	assert.Less(t, meanOutwardFacing(lathe(down)), -0.9)
}

func TestScrewNormalsPointOutward(t *testing.T) {
	mesh := lathedTube(t)

	normals := mesh.Float3Attribute(modeling.NormalAttribute)
	positions := mesh.Float3Attribute(modeling.PositionAttribute)

	for i := 0; i < normals.Len(); i++ {
		p := positions.At(i)
		radial := vector3.New(p.X(), 0., p.Z())
		if radial.Length() == 0 {
			continue
		}

		n := normals.At(i)
		assert.InDeltaf(t, 0, n.Y(), 1e-6, "normal %d should be flat along the tube's axis", i)
		assert.Greaterf(t, n.Dot(radial.Normalized()), 0.9,
			"normal %d should point radially outward, got %v at %v", i, n, p)
	}
}

func screwed(line []vector3.Float64, segments int, revolutions, distance float64) modeling.Mesh {
	return nodes.GetNodeOutputPort[modeling.Mesh](&nodes.Struct[extrude.ScrewNode]{
		Data: extrude.ScrewNode{
			Line:        nodes.ConstOutput[[]vector3.Float64]{Val: line},
			Segments:    nodes.ConstOutput[int]{Val: segments},
			Revolutions: nodes.ConstOutput[float64]{Val: revolutions},
			Distance:    nodes.ConstOutput[float64]{Val: distance},
		},
	}, "Out").Value()
}

func TestScrewDefaultUVsStayInTheUnitSquare(t *testing.T) {
	uvs := lathedTube(t).Float2Attribute(modeling.TexCoordAttribute)
	require.Positive(t, uvs.Len())
	for i := 0; i < uvs.Len(); i++ {
		uv := uvs.At(i)
		assert.Truef(t, uv.X() >= 0 && uv.X() <= 1 && uv.Y() >= 0 && uv.Y() <= 1, "uv %d is %v", i, uv)
	}
}

func TestScrewShadesAFullTurnWithoutASeam(t *testing.T) {
	mesh := lathedTube(t)
	normals := mesh.Float3Attribute(modeling.NormalAttribute)
	last := normals.Len() - 2

	assert.InDelta(t, 0, normals.At(0).Distance(normals.At(last)), 1e-9)
	assert.InDelta(t, 0, normals.At(0).Distance(vector3.Right[float64]()), 1e-9)
}

func TestScrewSegmentsCountTheStepsAround(t *testing.T) {
	mesh := screwed([]vector3.Float64{vector3.New(1., 0., 0.), vector3.New(1., 1., 0.)}, 8, 1, 0)
	assert.Equal(t, 16, mesh.PrimitiveCount())
}

func TestScrewGivesAPointOnTheAxisTheNormalOfItsCone(t *testing.T) {
	mesh := screwed([]vector3.Float64{
		vector3.New(0., 0., 0.), vector3.New(1., 1., 0.), vector3.New(0., 2., 0.),
	}, 16, 1, 0)

	normals := mesh.Float3Attribute(modeling.NormalAttribute)
	assert.InDelta(t, 0, normals.At(0).Distance(vector3.New(1., -1., 0.).Normalized()), 1e-9)
	assert.InDelta(t, 0, normals.At(2).Distance(vector3.New(1., 1., 0.).Normalized()), 1e-9)
}

func TestScrewNormalsSurviveAwkwardProfiles(t *testing.T) {
	for name, mesh := range map[string]modeling.Mesh{
		"repeated point": screwed([]vector3.Float64{
			vector3.New(1., 0., 0.), vector3.New(1., 0., 0.), vector3.New(1., 1., 0.),
		}, 16, 1, 0),
		"turning backwards": screwed([]vector3.Float64{vector3.New(1., 0., 0.), vector3.New(1., 1., 0.)}, 16, -1, 0),
		"helix":             screwed([]vector3.Float64{vector3.New(1., 0., 0.), vector3.New(1.5, 0.2, 0.)}, 32, 3, 4),
	} {
		t.Run(name, func(t *testing.T) {
			require.Positive(t, mesh.PrimitiveCount())
			normals := mesh.Float3Attribute(modeling.NormalAttribute)
			for i := 0; i < normals.Len(); i++ {
				require.InDeltaf(t, 1, normals.At(i).Length(), 1e-9, "normal %d", i)
			}
			requireNormalsAgreeWithFaces(t, mesh)
		})
	}
}
