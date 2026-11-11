package chance_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/math/chance"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func at(d chance.Distribution[float64], u float64) float64 {
	return d(chance.Fixed(u))
}

func TestUnitIsRepeatableAndInRange(t *testing.T) {
	seen := map[float64]bool{}
	for key := range 1000 {
		u := chance.Unit(7, key)
		assert.Equal(t, u, chance.Unit(7, key))
		assert.GreaterOrEqual(t, u, 0.0)
		assert.Less(t, u, 1.0)
		seen[u] = true
	}
	assert.Greater(t, len(seen), 990, "neighbouring keys must not collide")
}

func TestUnitDependsOnSeedAndKeyIndependently(t *testing.T) {
	assert.NotEqual(t, chance.Unit(1, 5), chance.Unit(2, 5))
	assert.NotEqual(t, chance.Unit(1, 5), chance.Unit(1, 6))
	assert.NotEqual(t, chance.Unit(1, 5), chance.Unit(5, 1), "seed and key are not interchangeable")
}

func TestUnitIsEvenlySpread(t *testing.T) {
	buckets := make([]int, 10)
	for key := range 100000 {
		buckets[int(chance.Unit(3, key)*10)]++
	}
	for i, count := range buckets {
		assert.InDelta(t, 10000, count, 400, "bucket %d", i)
	}
}

func TestRollStartsAtUnitAndKeepsGoing(t *testing.T) {
	next := chance.Roll(3, 11)
	first, second, third := next(), next(), next()
	assert.Equal(t, chance.Unit(3, 11), first)
	assert.NotEqual(t, first, second)
	assert.NotEqual(t, second, third)

	again := chance.Roll(3, 11)
	assert.Equal(t, first, again())
	assert.Equal(t, second, again())
}

func TestRollLaterValuesAreEvenlySpread(t *testing.T) {
	buckets := make([]int, 10)
	for key := range 50000 {
		next := chance.Roll(3, key)
		next()
		buckets[int(next()*10)]++
	}
	for i, count := range buckets {
		assert.InDelta(t, 5000, count, 300, "bucket %d", i)
	}
}

func TestDeriveGivesEachLabelItsOwnSeed(t *testing.T) {
	a, b := chance.Derive(4, "stools"), chance.Derive(4, "bottles")
	assert.NotEqual(t, a, b)
	assert.Equal(t, a, chance.Derive(4, "stools"))
	assert.NotEqual(t, a, chance.Derive(5, "stools"))
}

func TestUniformMapsTheUnitRangeOntoMinMax(t *testing.T) {
	d := chance.Uniform(2, 6)
	assert.Equal(t, 2.0, at(d, 0))
	assert.Equal(t, 4.0, at(d, 0.5))
	assert.Equal(t, 6.0, at(d, 1))
}

func TestNormalIsCentredAndClamped(t *testing.T) {
	d := chance.Normal(10, 2, 3)
	assert.InDelta(t, 10, at(d, 0.5), 1e-9)
	assert.InDelta(t, 4, at(d, 0), 1e-9, "clamped at mean - 3 deviations")
	assert.InDelta(t, 16, at(d, 1), 1e-9)
	assert.InDelta(t, 12, at(d, 0.8413), 0.01, "one deviation covers 84.13%")

	unclamped := chance.Normal(0, 1, 0)
	assert.Less(t, at(unclamped, 0.0001), -3.0)
	assert.False(t, math.IsInf(at(unclamped, 0), 0))
}

func TestLogNormalIsPositiveWithTheGivenMedian(t *testing.T) {
	d := chance.LogNormal(3, 0.5)
	assert.InDelta(t, 3, at(d, 0.5), 1e-9)
	assert.Greater(t, at(d, 0.001), 0.0)
	assert.Greater(t, at(d, 0.9), at(d, 0.1))
}

func TestExponentialHasTheGivenMedian(t *testing.T) {
	d := chance.Exponential(2)
	assert.InDelta(t, 2, at(d, 0.5), 1e-9)
	assert.Equal(t, 0.0, at(d, 0))
	assert.False(t, math.IsInf(at(d, 1), 0))
}

func TestTriangularPeaksAtTheMode(t *testing.T) {
	d := chance.Triangular(0, 1, 4)
	assert.Equal(t, 0.0, at(d, 0))
	assert.InDelta(t, 1, at(d, 0.25), 1e-9, "the mode sits at (mode-min)/(max-min) of the unit range")
	assert.Equal(t, 4.0, at(d, 1))
}

func TestPowerBiasesTowardsMin(t *testing.T) {
	d := chance.Power(0, 1, 2)
	assert.InDelta(t, 0.25, at(d, 0.5), 1e-9)
	assert.Equal(t, 1.0, at(d, 1))
}

func TestDiscretePicksByCumulativeWeight(t *testing.T) {
	d, err := chance.Discrete([]float64{10, 20, 30}, []float64{1, 0, 3})
	require.NoError(t, err)
	assert.Equal(t, 10.0, at(d, 0))
	assert.Equal(t, 10.0, at(d, 0.24))
	assert.Equal(t, 30.0, at(d, 0.26), "the zero-weight value is skipped")
	assert.Equal(t, 30.0, at(d, 0.999))
}

func TestDiscreteRefusesBadWeights(t *testing.T) {
	_, err := chance.Discrete([]float64{1, 2}, []float64{1})
	assert.Error(t, err, "lengths differ")
	_, err = chance.Discrete([]float64{1, 2}, []float64{0, 0})
	assert.Error(t, err, "nothing can be picked")
	_, err = chance.Discrete([]float64{1, 2}, []float64{1, -1})
	assert.Error(t, err, "negative")
	_, err = chance.Discrete(nil, nil)
	assert.Error(t, err, "empty")
}

func TestDiscreteWithNoWeightsIsEven(t *testing.T) {
	d, err := chance.Discrete([]float64{10, 20}, nil)
	require.NoError(t, err)
	assert.Equal(t, 10.0, at(d, 0.49))
	assert.Equal(t, 20.0, at(d, 0.51))
}

func TestRampSamplesTheGradient(t *testing.T) {
	d := chance.Ramp(coloring.NewGradient1D(
		coloring.GradientKey[float64]{Time: 0, Value: 5},
		coloring.GradientKey[float64]{Time: 1, Value: 7},
	))
	assert.InDelta(t, 6, at(d, 0.5), 1e-9)
}

func TestComposeDrawsEachAxisFromItsOwnPull(t *testing.T) {
	d := chance.Compose3D(chance.Uniform(0, 1), chance.Uniform(0, 10), chance.Constant(3))
	v := d(chance.Roll(5, 0))
	next := chance.Roll(5, 0)
	assert.InDelta(t, next(), v.X(), 1e-9)
	assert.InDelta(t, 10*next(), v.Y(), 1e-9)
	assert.InDelta(t, 3, v.Z(), 1e-9)

	same := chance.Compose3D(chance.Uniform(0, 1), chance.Uniform(0, 1), chance.Uniform(0, 1))(chance.Roll(5, 0))
	assert.NotEqual(t, same.X(), same.Y(), "the same scalar distribution on two axes still gives independent components")
	assert.NotEqual(t, same.Y(), same.Z())
}

func TestUniform3DStaysInsideItsBounds(t *testing.T) {
	d := chance.Uniform3D(vector3.New(-1.0, 0, 2), vector3.New(1.0, 0, 5))
	for key := range 1000 {
		p := d(chance.Roll(4, key))
		assert.GreaterOrEqual(t, p.X(), -1.0)
		assert.Less(t, p.X(), 1.0)
		assert.Equal(t, 0.0, p.Y())
		assert.GreaterOrEqual(t, p.Z(), 2.0)
		assert.Less(t, p.Z(), 5.0)
	}
	assert.Equal(t, vector3.New(0.0, 0, 3.5), d(chance.Fixed(0.5)))
}

func TestUniform2DAnd4DStayInsideTheirBounds(t *testing.T) {
	flat := chance.Uniform2D(vector2.New(-1.0, 2), vector2.New(1.0, 5))
	wide := chance.Uniform4D(vector4.New(0.0, 0, 0, 10), vector4.New(1.0, 1, 1, 20))
	for key := range 500 {
		p := flat(chance.Roll(6, key))
		assert.GreaterOrEqual(t, p.X(), -1.0)
		assert.Less(t, p.X(), 1.0)
		assert.GreaterOrEqual(t, p.Y(), 2.0)
		assert.Less(t, p.Y(), 5.0)

		q := wide(chance.Roll(6, key))
		assert.GreaterOrEqual(t, q.W(), 10.0)
		assert.Less(t, q.W(), 20.0)
		assert.NotEqual(t, q.X(), q.Y(), "components draw their own values")
	}
}

func TestInsideCircleStaysInsideAndFillsTheArea(t *testing.T) {
	d := chance.InsideCircle(2)
	inner := 0
	for key := range 20000 {
		p := d(chance.Roll(1, key))
		assert.LessOrEqual(t, p.Length(), 2.0+1e-9)
		if p.Length() < 2*math.Sqrt(0.5) {
			inner++
		}
	}
	assert.InDelta(t, 10000, inner, 400, "half the points sit inside half the area")
}

func TestDirection2DStaysInsideTheArc(t *testing.T) {
	axis := vector2.New(0.0, 1)
	d := chance.Direction2D(axis, 30)
	for key := range 2000 {
		v := d(chance.Roll(2, key))
		assert.InDelta(t, 1, v.Length(), 1e-9)
		angle := math.Acos(v.Dot(axis)) * 180 / math.Pi
		assert.LessOrEqual(t, angle, 30+1e-6)
	}
	exact := chance.Direction2D(vector2.New(3.0, 0), 0)(chance.Roll(2, 5))
	assert.InDelta(t, 1, exact.X(), 1e-9)
}

func rotationAngle(q quaternion.Quaternion) float64 {
	return 2 * math.Acos(math.Min(1, math.Abs(q.W()))) * 180 / math.Pi
}

func TestRotationStaysWithinMaxAngleAndIsEvenOverRotations(t *testing.T) {
	within := chance.Rotation(30)
	for key := range 2000 {
		assert.LessOrEqual(t, rotationAngle(within(chance.Roll(3, key))), 30+1e-6)
	}

	any := chance.Rotation(180)
	under90 := 0
	for key := range 20000 {
		if rotationAngle(any(chance.Roll(4, key))) < 90 {
			under90++
		}
	}
	assert.InDelta(t, 20000*(math.Pi/2-1)/math.Pi, under90, 400, "rotations are spread evenly over rotation space, not over angles")
}

func TestInsideSphereStaysInsideAndFillsTheVolume(t *testing.T) {
	d := chance.InsideSphere(2)
	inner := 0
	for key := range 20000 {
		p := d(chance.Roll(1, key))
		assert.LessOrEqual(t, p.Length(), 2.0+1e-9)
		if p.Length() < 2*math.Cbrt(0.5) {
			inner++
		}
	}
	assert.InDelta(t, 10000, inner, 400, "half the points sit inside half the volume")
}

func TestDirectionStaysInsideTheCone(t *testing.T) {
	axis := vector3.Up[float64]()
	d := chance.Direction(axis, 30)
	for key := range 2000 {
		v := d(chance.Roll(2, key))
		assert.InDelta(t, 1, v.Length(), 1e-9)
		angle := math.Acos(v.Dot(axis)) * 180 / math.Pi
		assert.LessOrEqual(t, angle, 30+1e-6)
	}
	exact := chance.Direction(vector3.New(1.0, 0, 0), 0)(chance.Roll(2, 5))
	assert.InDelta(t, 1, exact.X(), 1e-9)
}
