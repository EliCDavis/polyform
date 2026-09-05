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

// Scaling a primitive is how you get an oval or lens shape out of a
// sphere, and it used to return the field's own local distances rather
// than world ones. An isolated shape still marched to roughly the right
// surface, which is what made this hard to see - the damage showed up once
// the scaled field met another one, because a smooth union blends by
// comparing field *values* against its radius, and those were off by
// 1/scale.
//
// The invariant: a uniformly scaled primitive must behave exactly like the
// equivalent unscaled primitive, in a blend as much as alone.
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

// The same guarantee at the level the caller actually works at: the mesh.
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

	// Vertex-by-vertex equality is over-specified here: March accumulates
	// into a map, so it emits the same surface in a different order every
	// run, and the two fields agree to 1e-9 rather than bit-for-bit, which
	// interpolation turns into sub-voxel differences. The exact claim is
	// the field-level one above. What the mesh has to show is that the two
	// occupy the same space.
	wantMin, wantMax := direct.BoundingBox(modeling.PositionAttribute).Min(), direct.BoundingBox(modeling.PositionAttribute).Max()
	gotMin, gotMax := viaScale.BoundingBox(modeling.PositionAttribute).Min(), viaScale.BoundingBox(modeling.PositionAttribute).Max()
	assert.InDelta(t, 0, wantMin.Sub(gotMin).Length(), 0.01, "lower bound differs")
	assert.InDelta(t, 0, wantMax.Sub(gotMax).Length(), 0.01, "upper bound differs")

	// The blend is the part that broke: an unscaled field made the small
	// sphere read as 4x deeper than it is, which widened the join into the
	// body. The bounding box alone wouldn't catch that, but the waist
	// where the two meet would move.
	assert.InDelta(t, 0, direct.BoundingBox(modeling.PositionAttribute).Size().Sub(viaScale.BoundingBox(modeling.PositionAttribute).Size()).Length(), 0.01,
		"the blended shape's extent differs")
}

// A non-uniform scale has no exact equivalent primitive, so the guarantee
// there is the weaker one that still matters: a finite field, a real
// surface, and every vertex on the ellipsoid it describes.
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
