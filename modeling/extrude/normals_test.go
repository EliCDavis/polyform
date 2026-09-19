package extrude_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A vertex normal is averaged across the faces that share it, so it drifts
// away from any one of them. Reaching a right angle means it stopped being a
// normal: ProjectFace used to hand back the ring's edge tangent, and Line
// negated both flanks when only one needed it.
func requireNormalsAgreeWithFaces(t *testing.T, m modeling.Mesh) {
	t.Helper()

	require.True(t, m.HasFloat3Attribute(modeling.NormalAttribute), "no normal attribute")
	idx := m.Indices()
	pos := m.Float3Attribute(modeling.PositionAttribute)
	nrm := m.Float3Attribute(modeling.NormalAttribute)
	require.Equal(t, pos.Len(), nrm.Len())

	for i := 0; i < idx.Len(); i += 3 {
		a, b, c := pos.At(idx.At(i)), pos.At(idx.At(i+1)), pos.At(idx.At(i+2))
		face := b.Sub(a).Cross(c.Sub(a))
		if face.Length() == 0 {
			continue
		}
		face = face.Normalized()

		for k := 0; k < 3; k++ {
			vertex := idx.At(i + k)
			normal := nrm.At(vertex)
			require.NotZerof(t, normal.Length(), "vertex %d has a zero normal", vertex)
			assert.Positivef(t, face.Dot(normal.Normalized()),
				"vertex %d carries a normal at or past a right angle to its face", vertex)
		}
	}
}

func TestExtrudersEmitUsableNormals(t *testing.T) {
	square := []vector2.Float64{
		vector2.New(-1., -1.), vector2.New(1., -1.),
		vector2.New(1., 1.), vector2.New(-1., 1.),
	}
	bent := []vector3.Float64{
		vector3.New(0., 0., 0.), vector3.New(0., 2., 0.), vector3.New(1.5, 3.5, 0.),
	}

	outline, err := extrude.Outline{Shape: square, Path: bent}.Extrude()
	require.NoError(t, err)

	for name, mesh := range map[string]modeling.Mesh{
		"shape": extrude.Shape(square, bent),
		"closed shape": extrude.ClosedShape(square, []vector3.Float64{
			vector3.New(2., 0., 0.), vector3.New(-1., 0., 1.7), vector3.New(-1., 0., -1.7),
		}),
		"polygon": extrude.Polygon(8, []extrude.ExtrusionPoint{
			{Point: vector3.New(0., 0., 0.), Thickness: 1},
			{Point: vector3.New(0., 2., 0.), Thickness: 1},
			{Point: vector3.New(1.5, 3.5, 0.), Thickness: .7},
		}),
		"circle": extrude.Circle{Radius: 1, Resolution: 12, Path: bent}.Extrude(),
		"line": extrude.Line([]extrude.LinePoint{
			{Point: vector3.New(0., 0., 0.), Up: vector3.Up[float64](), Width: 1, Height: .2},
			{Point: vector3.New(0., 0., 2.), Up: vector3.Up[float64](), Width: 1, Height: .2},
			{Point: vector3.New(1., 0., 4.), Up: vector3.Up[float64](), Width: 1, Height: .2},
		}),
		"screw":   lathedTube(t),
		"outline": outline,
	} {
		t.Run(name, func(t *testing.T) { requireNormalsAgreeWithFaces(t, mesh) })
	}
}

// The outline decides which way is out, so the same shape wound either way
// has to produce normals pointing the same direction.
func TestProjectFaceFacesOutwardEitherWinding(t *testing.T) {
	square := []vector2.Float64{
		vector2.New(-1., -1.), vector2.New(1., -1.),
		vector2.New(1., 1.), vector2.New(-1., 1.),
	}
	backwards := make([]vector2.Float64, len(square))
	for i, p := range square {
		backwards[len(square)-1-i] = p
	}

	center := vector3.New(0., 0., 0.)
	up := vector3.Up[float64]()
	perpendicular := vector3.Forward[float64]()

	points, normals := extrude.ProjectFace(center, up, perpendicular, square)
	flippedPoints, flippedNormals := extrude.ProjectFace(center, up, perpendicular, backwards)

	for i, p := range points {
		assert.Positivef(t, normals[i].Dot(p.Sub(center).Normalized()), "vertex %d points inward", i)
	}
	for i, p := range flippedPoints {
		assert.Positivef(t, flippedNormals[i].Dot(p.Sub(center).Normalized()),
			"reversed winding, vertex %d points inward", i)
	}
}
