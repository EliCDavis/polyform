package geometry_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func signedArea(ring []vector2.Float64) float64 {
	return geometry.Shape(ring).SignedArea()
}

func TestNewPlaneFromPolygon_FlatSquare(t *testing.T) {
	plane := geometry.NewPlaneFromPolygon([]vector3.Float64{
		vector3.New(0., 2., 0.),
		vector3.New(0., 2., 1.),
		vector3.New(1., 2., 1.),
		vector3.New(1., 2., 0.),
	})

	assert.InDelta(t, 1., plane.Normal().Length(), 1e-12)
	assert.InDelta(t, 1., plane.Normal().Y(), 1e-12,
		"Z then X is counter-clockwise seen from +Y, so the ring faces +Y")
	assert.InDelta(t, 2., plane.Origin().Y(), 1e-12)
}

func TestNewPlaneFromPolygon_AgreesWithThreePoints(t *testing.T) {
	a := vector3.New(0., 0., 0.)
	b := vector3.New(3., 0., 1.)
	c := vector3.New(1., 2., 0.)

	fromPolygon := geometry.NewPlaneFromPolygon([]vector3.Float64{a, b, c})
	fromPoints := geometry.NewPlaneFromPoints(a, b, c)

	assert.InDelta(t, 0., fromPolygon.Normal().Distance(fromPoints.Normal()), 1e-12)
	assert.InDelta(t, 0., fromPolygon.Origin().Distance(fromPoints.Origin()), 1e-12)
}

func TestNewPlaneFromPolygon_FoldedRingSplitsTheDifference(t *testing.T) {
	plane := geometry.NewPlaneFromPolygon([]vector3.Float64{
		vector3.New(0., 0., 0.),
		vector3.New(1., 0., 0.),
		vector3.New(1., 0., 1.),
		vector3.New(0., 0., 1.),
		vector3.New(0., 1., 1.),
		vector3.New(0., 1., 0.),
	})

	assert.InDelta(t, plane.Normal().X(), plane.Normal().Y(), 1e-12)
	assert.InDelta(t, 0., plane.Normal().Z(), 1e-12)
}

func TestNewPlaneFromPolygon_NoAreaHasNoNormal(t *testing.T) {
	collinear := geometry.NewPlaneFromPolygon([]vector3.Float64{
		vector3.New(0., 0., 0.),
		vector3.New(1., 0., 0.),
		vector3.New(2., 0., 0.),
		vector3.New(3., 0., 0.),
	})
	assert.Zero(t, collinear.Normal().Length())
	assert.Zero(t, geometry.NewPlaneFromPolygon(nil).Normal().Length())
}

func TestPlane_Basis_IsOrthonormalAndRightHanded(t *testing.T) {
	normals := []vector3.Float64{
		vector3.New(0., 0., 1.),
		vector3.New(0., 0., -1.),
		vector3.New(1., 0., 0.),
		vector3.New(-1., 0., 0.),
		vector3.New(0., 1., 0.),
		vector3.New(0., -1., 0.),
		vector3.New(1., 2., 3.).Normalized(),
		vector3.New(-3., 1., -0.001).Normalized(),
		vector3.New(0.6, -0.8, 1e-9).Normalized(),
	}

	for _, n := range normals {
		u, v := geometry.NewPlane(vector3.Zero[float64](), n).Basis()
		assert.InDeltaf(t, 1., u.Length(), 1e-12, "|u| for %v", n)
		assert.InDeltaf(t, 1., v.Length(), 1e-12, "|v| for %v", n)
		assert.InDeltaf(t, 0., u.Dot(v), 1e-12, "u·v for %v", n)
		assert.InDeltaf(t, 0., u.Dot(n), 1e-12, "u·n for %v", n)
		assert.InDeltaf(t, 0., u.Cross(v).Distance(n), 1e-12, "u×v for %v", n)
	}
}

func TestPlane_Project_KeepsTheWindingSeenFromTheNormal(t *testing.T) {
	ring := []vector3.Float64{
		vector3.New(0., 0., 0.),
		vector3.New(0., 2., 1.),
		vector3.New(1., 2., 1.),
		vector3.New(1., 0., 0.),
	}
	plane := geometry.NewPlaneFromPolygon(ring)

	assert.Greater(t, signedArea(plane.Project(ring)), 0.,
		"the polygon's own plane faces the side it is counter-clockwise from")

	flipped := geometry.NewPlane(plane.Origin(), plane.Normal().Scale(-1))
	assert.Less(t, signedArea(flipped.Project(ring)), 0.)
}

func TestPlane_Project_PreservesDistancesInThePlane(t *testing.T) {
	ring := []vector3.Float64{
		vector3.New(1., 1., 1.),
		vector3.New(4., 1., 1.),
		vector3.New(4., 5., 1.),
		vector3.New(1., 5., 1.),
	}
	flat := geometry.NewPlaneFromPolygon(ring).Project(ring)

	for i := range ring {
		j := (i + 1) % len(ring)
		assert.InDeltaf(t, ring[i].Distance(ring[j]), flat[i].Distance(flat[j]), 1e-12, "edge %d", i)
	}
	assert.InDelta(t, 12., math.Abs(signedArea(flat)), 1e-12)
}

func TestPlane_Project_DropsHeightAboveThePlane(t *testing.T) {
	plane := geometry.NewPlane(vector3.Zero[float64](), vector3.Up[float64]())
	flat := plane.Project([]vector3.Float64{
		vector3.New(3., 0., 4.),
		vector3.New(3., 7., 4.),
	})
	assert.Equal(t, flat[0], flat[1])
}
