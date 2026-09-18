package csg_test

import (
	"sync"
	"testing"

	"github.com/EliCDavis/polyform/modeling/csg"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSolidRejectsWhatCheckClosedRejects(t *testing.T) {
	open := primitives.Cylinder{Sides: 12, Height: 2, Radius: 1, NoTop: true}.ToMesh()

	_, err := csg.NewSolid(open)
	require.Error(t, err)
	assert.EqualError(t, err, csg.CheckClosed(open).Error())
}

func TestASolidIsUnchangedByTheOperationsItTakesPartIn(t *testing.T) {
	box, err := csg.NewSolid(cube(vector3.Zero[float64](), 2))
	require.NoError(t, err)
	left, err := csg.NewSolid(primitives.UVSphere(1.1, 12, 18).Translate(vector3.New(-0.7, 0., 0.)))
	require.NoError(t, err)
	right, err := csg.NewSolid(primitives.UVSphere(1.1, 12, 18).Translate(vector3.New(0.7, 0., 0.)))
	require.NoError(t, err)

	before, err := box.Subtract(left)
	require.NoError(t, err)
	_, err = box.Subtract(right)
	require.NoError(t, err)
	after, err := box.Subtract(left)
	require.NoError(t, err)

	assert.Equal(t, before.Mesh().PrimitiveCount(), after.Mesh().PrimitiveCount())
	assert.InDelta(t, volume(before.Mesh()), volume(after.Mesh()), 1e-9)
}

func TestASolidCanTakePartInOperationsAtOnce(t *testing.T) {
	box, err := csg.NewSolid(cube(vector3.Zero[float64](), 2))
	require.NoError(t, err)
	ball, err := csg.NewSolid(primitives.UVSphere(1.28, 12, 18))
	require.NoError(t, err)

	expected, err := box.Subtract(ball)
	require.NoError(t, err)

	var wg sync.WaitGroup
	results := make([]*csg.Solid, 8)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], _ = box.Subtract(ball)
		}(i)
	}
	wg.Wait()

	for i, result := range results {
		require.NotNilf(t, result, "run %d", i)
		assert.Equalf(t, expected.Mesh().PrimitiveCount(), result.Mesh().PrimitiveCount(), "run %d", i)
	}
}

func TestResultsFeedTheNextOperation(t *testing.T) {
	box, err := csg.NewSolid(cube(vector3.Zero[float64](), 2))
	require.NoError(t, err)
	ball, err := csg.NewSolid(primitives.UVSphere(1.28, 12, 18))
	require.NoError(t, err)
	bar, err := csg.NewSolid(cube(vector3.Zero[float64](), 1).Scale(vector3.New(4., 1., 1.)))
	require.NoError(t, err)

	chained, err := box.Intersect(ball)
	require.NoError(t, err)
	chained, err = chained.Subtract(bar)
	require.NoError(t, err)

	step, err := csg.Intersect(box.Mesh(), ball.Mesh())
	require.NoError(t, err)
	direct, err := csg.Subtract(step, bar.Mesh())
	require.NoError(t, err)

	requireWatertight(t, chained.Mesh())
	assert.Equal(t, direct.PrimitiveCount(), chained.Mesh().PrimitiveCount())
	assert.InDelta(t, volume(direct), volume(chained.Mesh()), 1e-9)
}
