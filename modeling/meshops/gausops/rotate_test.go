package gausops_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops/gausops"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
	"github.com/stretchr/testify/assert"
)

var degreeOneDirections = []vector3.Float64{
	vector3.New(0.3, -1.2, 0.8),
	vector3.New(-2., 0.5, 0.1),
}

func splat() modeling.Mesh {
	v3 := map[string][]vector3.Float64{
		modeling.PositionAttribute: {vector3.New(1., 0., 0.), vector3.New(0., 2., 0.)},
		modeling.FDCAttribute:      {vector3.New(0.1, 0.2, 0.3), vector3.New(0.4, 0.5, 0.6)},
	}
	for j := 0; j < 15; j++ {
		v3[modeling.SHAttribute(j)] = make([]vector3.Float64, len(degreeOneDirections))
	}
	for i, v := range degreeOneDirections {
		v3[modeling.SHAttribute(0)][i] = vector3.Fill(-v.Y())
		v3[modeling.SHAttribute(1)][i] = vector3.Fill(v.Z())
		v3[modeling.SHAttribute(2)][i] = vector3.Fill(-v.X())
		for j := 3; j < 15; j++ {
			f := float64(i*15 + j)
			v3[modeling.SHAttribute(j)][i] = vector3.New(f*0.1, -f*0.2, f*f*0.01)
		}
	}
	return modeling.NewPointCloud(
		map[string][]vector4.Float64{
			modeling.RotationAttribute: {vector4.New(1., 0., 0., 0.), vector4.New(1., 0., 0., 0.)},
		},
		v3,
		nil,
		map[string][]float64{modeling.OpacityAttribute: {0.5, 1}},
	)
}

func assertSHEqual(t *testing.T, want, got modeling.Mesh) {
	assert.Equal(t, modeling.SHCoefficientCount(want), modeling.SHCoefficientCount(got))
	for j := 0; j < modeling.SHCoefficientCount(want); j++ {
		w := want.Float3Attribute(modeling.SHAttribute(j))
		g := got.Float3Attribute(modeling.SHAttribute(j))
		for i := 0; i < w.Len(); i++ {
			assert.InDelta(t, w.At(i).X(), g.At(i).X(), 1e-9, "coefficient %d point %d", j, i)
			assert.InDelta(t, w.At(i).Y(), g.At(i).Y(), 1e-9, "coefficient %d point %d", j, i)
			assert.InDelta(t, w.At(i).Z(), g.At(i).Z(), 1e-9, "coefficient %d point %d", j, i)
		}
	}
}

func TestRotateSHByIdentityLeavesSHUnchanged(t *testing.T) {
	assertSHEqual(t, splat(), gausops.RotateSH(splat(), quaternion.Identity()))
}

func TestRotateSHDegreeOneMatchesRotatedVector(t *testing.T) {
	q := quaternion.FromTheta(2.1, vector3.New(0.5, -1., 0.25).Normalized())
	rotated := gausops.RotateSH(splat(), q)

	for i, v := range degreeOneDirections {
		want := q.Rotate(v)
		got := vector3.New(
			-rotated.Float3Attribute(modeling.SHAttribute(2)).At(i).X(),
			-rotated.Float3Attribute(modeling.SHAttribute(0)).At(i).X(),
			rotated.Float3Attribute(modeling.SHAttribute(1)).At(i).X(),
		)
		assert.InDelta(t, want.X(), got.X(), 1e-9)
		assert.InDelta(t, want.Y(), got.Y(), 1e-9)
		assert.InDelta(t, want.Z(), got.Z(), 1e-9)
	}
}

func TestRotateSHTwiceEqualsRotatingByTheProduct(t *testing.T) {
	a := quaternion.FromEulerAngle(vector3.New(0.3, 1.2, -0.4))
	b := quaternion.FromTheta(-1.7, vector3.New(1., 1., 0.).Normalized())

	twice := gausops.RotateSH(gausops.RotateSH(splat(), a), b)
	once := gausops.RotateSH(splat(), b.Multiply(a))
	assertSHEqual(t, once, twice)
}

func TestRotateSHWithoutSHIsANoop(t *testing.T) {
	plain := modeling.NewPointCloud(nil, map[string][]vector3.Float64{
		modeling.PositionAttribute: {vector3.New(1., 2., 3.)},
	}, nil, nil)
	rotated := gausops.RotateSH(plain, quaternion.FromTheta(1, vector3.Up[float64]()))
	assert.Equal(t, vector3.New(1., 2., 3.), rotated.Float3Attribute(modeling.PositionAttribute).At(0))
}

func TestRotateMovesPositionsOrientationsAndSH(t *testing.T) {
	q := quaternion.FromTheta(1.1, vector3.New(0., 0., 1.))
	rotated := gausops.Rotate(splat(), q)

	position := rotated.Float3Attribute(modeling.PositionAttribute).At(0)
	wantPosition := q.Rotate(vector3.New(1., 0., 0.))
	assert.InDelta(t, wantPosition.X(), position.X(), 1e-9)
	assert.InDelta(t, wantPosition.Y(), position.Y(), 1e-9)
	assert.InDelta(t, wantPosition.Z(), position.Z(), 1e-9)

	orientation := rotated.Float4Attribute(modeling.RotationAttribute).At(0)
	assert.InDelta(t, q.W(), orientation.X(), 1e-9)
	assert.InDelta(t, q.Dir().Z(), orientation.W(), 1e-9)

	assertSHEqual(t, gausops.RotateSH(splat(), q), rotated)
	assert.Equal(t, 0.5, rotated.Float1Attribute(modeling.OpacityAttribute).At(0))
}
