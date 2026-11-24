package sdf

import (
	"math"
	"math/rand"
	"testing"

	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestBoundProbeCoverageHoldsForEveryDirection(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	for i := 0; i < 100000; i++ {
		dir := vector3.New(rng.NormFloat64(), rng.NormFloat64(), rng.NormFloat64()).Normalized()
		best := math.Inf(-1)
		for _, probe := range boundProbeDirections {
			best = math.Max(best, dir.Dot(probe))
		}
		assert.GreaterOrEqual(t, best, boundProbeCoverage)
	}
}

func TestBoundingRadiusContainsOffsetSphere(t *testing.T) {
	radius, ok := boundingRadius(Sphere(vector3.New(2., -1., 0.5), 0.75))
	assert.True(t, ok)
	assert.GreaterOrEqual(t, radius, vector3.New(2., -1., 0.5).Length()+0.75)
	assert.Less(t, radius, 5.)
}

func TestBoundingRadiusRejectsPlane(t *testing.T) {
	_, ok := boundingRadius(func(v vector3.Float64) float64 { return v.Y() })
	assert.False(t, ok)
}
