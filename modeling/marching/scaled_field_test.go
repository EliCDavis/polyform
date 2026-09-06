package marching_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func scaledUnitSphere(center vector3.Float64, scale float64) sample.Vec3ToFloat {
	return sdf.Transform(
		sdf.Sphere(vector3.Zero[float64](), 1),
		trs.New(center, trs.Identity().Rotation(), vector3.Fill(scale)),
	)
}

func TestSmoothUnionTreatsAScaledPrimitiveLikeAnEquivalentOne(t *testing.T) {
	const (
		radius = 0.25
		blend  = 0.15
	)
	offset := vector3.New(0.6, 0., 0.)
	body := sdf.Sphere(vector3.Zero[float64](), 0.5)

	viaScale := sdf.SmoothUnion(blend, body, scaledUnitSphere(offset, radius))
	direct := sdf.SmoothUnion(blend, body, sdf.Sphere(offset, radius))

	for x := -1.0; x <= 1.5; x += 0.02 {
		for y := -0.8; y <= 0.8; y += 0.02 {
			at := vector3.New(x, y, 0.)
			require.InDelta(t, direct(at), viaScale(at), 1e-9,
				"a scaled unit sphere must blend identically to the sphere it equals, at %v", at)
		}
	}
}

func TestMarchingAScaledPrimitiveMatchesAnEquivalentOne(t *testing.T) {
	const (
		radius = 0.3
		blend  = 0.12
	)
	offset := vector3.New(0.55, 0., 0.)
	body := sdf.Sphere(vector3.Zero[float64](), 0.5)
	domain := geometry.NewAABB(vector3.Zero[float64](), vector3.Fill(3.))

	march := func(field sample.Vec3ToFloat) modeling.Mesh {
		return marching.March(field, domain, 0.05, 0)
	}

	viaScale := march(sdf.SmoothUnion(blend, body, scaledUnitSphere(offset, radius)))
	direct := march(sdf.SmoothUnion(blend, body, sdf.Sphere(offset, radius)))

	require.Greater(t, direct.PrimitiveCount(), 100, "the reference mesh should be a real surface")
	assert.Equal(t, direct.PrimitiveCount(), viaScale.PrimitiveCount(),
		"a scaled primitive should march to the same surface as the one it equals")

	wantMin, wantMax := direct.BoundingBox(modeling.PositionAttribute).Min(), direct.BoundingBox(modeling.PositionAttribute).Max()
	gotMin, gotMax := viaScale.BoundingBox(modeling.PositionAttribute).Min(), viaScale.BoundingBox(modeling.PositionAttribute).Max()
	assert.InDelta(t, 0, wantMin.Sub(gotMin).Length(), 0.01, "lower bound differs")
	assert.InDelta(t, 0, wantMax.Sub(gotMax).Length(), 0.01, "upper bound differs")

	assert.InDelta(t, 0, direct.BoundingBox(modeling.PositionAttribute).Size().Sub(viaScale.BoundingBox(modeling.PositionAttribute).Size()).Length(), 0.01,
		"the blended shape's extent differs")
}

func TestMarchingANonUniformlyScaledPrimitive(t *testing.T) {
	for _, tc := range []struct {
		name  string
		scale vector3.Float64
	}{
		{"squashed on one axis", vector3.New(1., 0.3, 1.)},
		{"squashed on two axes", vector3.New(0.25, 1., 0.4)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			scale := tc.scale
			field := sdf.Transform(
				sdf.Sphere(vector3.Zero[float64](), 1),
				trs.New(vector3.Zero[float64](), trs.Identity().Rotation(), scale),
			)
			mesh := marching.March(field, geometry.NewAABB(vector3.Zero[float64](), vector3.Fill(3.)), 0.05, 0)
			require.Greater(t, mesh.PrimitiveCount(), 100)

			positions := mesh.Float3Attribute(modeling.PositionAttribute)
			for i := 0; i < positions.Len(); i++ {
				p := positions.At(i)
				require.False(t,
					math.IsNaN(p.X()) || math.IsNaN(p.Y()) || math.IsNaN(p.Z()),
					"vertex %d is NaN", i)

				q := vector3.New(p.X()/scale.X(), p.Y()/scale.Y(), p.Z()/scale.Z())
				assert.InDelta(t, 1.0, q.Length(), 0.1,
					"vertex %d at %v is not on the ellipsoid", i, p)
			}
		})
	}
}
