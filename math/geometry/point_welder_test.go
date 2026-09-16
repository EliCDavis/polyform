package geometry_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestPointWelder3D_ZeroToleranceMatchesExactly(t *testing.T) {
	w := geometry.NewPointWelder3D(0)

	assert.Equal(t, 0, w.Index(vector3.New(1., 2., 3.)))
	assert.Equal(t, 1, w.Index(vector3.New(1., 2., 3.0000001)))
	assert.Equal(t, 0, w.Index(vector3.New(1., 2., 3.)))
	assert.Len(t, w.Points(), 2)
}

func TestPointWelder3D_FirstPointAddedWins(t *testing.T) {
	w := geometry.NewPointWelder3D(0.1)

	assert.Equal(t, 0, w.Index(vector3.New(0., 0., 0.)))
	assert.Equal(t, 0, w.Index(vector3.New(0.05, 0., 0.)))
	assert.Equal(t, 1, w.Index(vector3.New(0.5, 0., 0.)))
	assert.Equal(t, 0, w.Index(vector3.New(0., 0.09, 0.)))
	assert.Equal(t, []vector3.Float64{
		vector3.New(0., 0., 0.),
		vector3.New(0.5, 0., 0.),
	}, w.Points())
}

func TestPointWelder3D_MergesAcrossCellBoundaries(t *testing.T) {
	const tolerance = 0.001

	// Cells are 8 tolerances wide, so these all sit on a cell boundary.
	for _, boundary := range []float64{0, 0.008, 1, 1000, -0.016} {
		w := geometry.NewPointWelder3D(tolerance)
		below := vector3.New(boundary-tolerance/4, boundary-tolerance/4, boundary-tolerance/4)
		above := vector3.New(boundary+tolerance/4, boundary+tolerance/4, boundary+tolerance/4)
		assert.Equalf(t, w.Index(below), w.Index(above), "around %g", boundary)
	}
}

func TestPointWelder3D_AddNeverMerges(t *testing.T) {
	w := geometry.NewPointWelder3D(0.1)
	p := vector3.New(1., 1., 1.)

	assert.Equal(t, 0, w.Add(p))
	assert.Equal(t, 1, w.Add(p))
	assert.Equal(t, 0, w.Index(p))

	i, ok := w.Find(vector3.New(1.05, 1., 1.))
	assert.True(t, ok)
	assert.Equal(t, 0, i)

	_, ok = w.Find(vector3.New(2., 1., 1.))
	assert.False(t, ok)
}

func TestPointWelder2D_IgnoresTheThirdAxis(t *testing.T) {
	w := geometry.NewPointWelder2D(0.01)

	assert.Equal(t, 0, w.Index(vector2.New(3., 4.)))
	assert.Equal(t, 0, w.Index(vector2.New(3.005, 4.)))
	assert.Equal(t, 1, w.Index(vector2.New(3.02, 4.)))
	assert.Len(t, w.Points(), 2)
}
