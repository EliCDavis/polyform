package sdf_test

import (
	"math"
	"math/rand"
	"testing"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func scaledAt(position vector3.Float64, scale float64) trs.TRS {
	return trs.New(position, quaternion.Identity(), vector3.Fill(scale))
}

func TestRepeatSingleCopyReportsWorldDistanceUnderScale(t *testing.T) {
	field := sdf.Repeat(sdf.Sphere(vector3.Zero[float64](), 1), []trs.TRS{scaledAt(vector3.Zero[float64](), 0.5)}, 0)
	assert.InDelta(t, 1.5, field(vector3.New(2., 0., 0.)), 1e-9)
}

func TestRepeatManyCopiesReportWorldDistanceUnderScale(t *testing.T) {
	field := sdf.Repeat(sdf.Sphere(vector3.Zero[float64](), 1), []trs.TRS{
		scaledAt(vector3.Zero[float64](), 0.5),
		scaledAt(vector3.New(10., 0., 0.), 0.5),
	}, 0)
	assert.InDelta(t, 1.5, field(vector3.New(2., 0., 0.)), 1e-9)
}

func TestRepeatSmoothManyCopiesReportWorldDistanceUnderScale(t *testing.T) {
	field := sdf.Repeat(sdf.Sphere(vector3.Zero[float64](), 1), []trs.TRS{
		scaledAt(vector3.Zero[float64](), 0.5),
		scaledAt(vector3.New(10., 0., 0.), 0.5),
	}, 0.1)
	assert.InDelta(t, 1.5, field(vector3.New(2., 0., 0.)), 1e-9)
}

func TestRepeatNonUniformScaleStaysConservative(t *testing.T) {
	transform := trs.New(vector3.Zero[float64](), quaternion.Identity(), vector3.New(4., 0.25, 1.))
	field := sdf.Repeat(sdf.Sphere(vector3.Zero[float64](), 1), []trs.TRS{transform}, 0)
	assert.LessOrEqual(t, field(vector3.New(0., 2., 0.)), 1.75+1e-9)
	assert.LessOrEqual(t, field(vector3.New(6., 0., 0.)), 2.+1e-9)
}

func randomTransforms(rng *rand.Rand, count int) []trs.TRS {
	transforms := make([]trs.TRS, count)
	for i := range transforms {
		transforms[i] = trs.New(
			vector3.New(rng.Float64()*10-5, rng.Float64()*10-5, rng.Float64()*10-5),
			quaternion.FromEulerAngle(vector3.New(rng.Float64()*6, rng.Float64()*6, rng.Float64()*6)),
			vector3.New(0.2+rng.Float64(), 0.2+rng.Float64(), 0.2+rng.Float64()),
		)
	}
	return transforms
}

func bruteForceRepeat(field func(vector3.Float64) float64, transforms []trs.TRS, radius float64) func(vector3.Float64) float64 {
	copies := make([]sample.Vec3ToFloat, len(transforms))
	for i, t := range transforms {
		copies[i] = sdf.Transform(field, t)
	}
	return sdf.SmoothUnion(radius, copies...)
}

func TestRepeatMatchesBruteForceUnion(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	shape := sdf.Box(vector3.Zero[float64](), vector3.New(0.6, 0.3, 0.9))
	for _, radius := range []float64{0, 0.2} {
		transforms := randomTransforms(rng, 300)
		fast := sdf.Repeat(shape, transforms, radius)
		slow := bruteForceRepeat(shape, transforms, radius)
		for i := 0; i < 2000; i++ {
			p := vector3.New(rng.Float64()*16-8, rng.Float64()*16-8, rng.Float64()*16-8)
			assert.InDelta(t, slow(p), fast(p), 1e-9, "radius %v at %v", radius, p)
		}
	}
}

func TestRepeatUnboundedFieldMatchesBruteForceUnion(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	plane := func(v vector3.Float64) float64 { return v.Y() }
	transforms := randomTransforms(rng, 40)
	fast := sdf.Repeat(plane, transforms, 0)
	slow := bruteForceRepeat(plane, transforms, 0)
	for i := 0; i < 500; i++ {
		p := vector3.New(rng.Float64()*16-8, rng.Float64()*16-8, rng.Float64()*16-8)
		assert.InDelta(t, slow(p), fast(p), 1e-9)
	}
}

func TestRepeatEvaluatesFewCopiesNearOne(t *testing.T) {
	calls := 0
	sphere := sdf.Sphere(vector3.Zero[float64](), 0.1)
	counting := func(v vector3.Float64) float64 {
		calls++
		return sphere(v)
	}

	transforms := make([]trs.TRS, 0, 1000)
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			for z := 0; z < 10; z++ {
				transforms = append(transforms, trs.Position(vector3.New(float64(x), float64(y), float64(z))))
			}
		}
	}

	field := sdf.Repeat(counting, transforms, 0.05)
	calls = 0
	assert.InDelta(t, 0.1, field(vector3.New(5.2, 5., 5.)), 1e-9)
	assert.Less(t, calls, 50)
	assert.False(t, math.IsInf(field(vector3.New(50., 50., 50.)), 0))
}
