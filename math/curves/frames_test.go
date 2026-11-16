package curves_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/curves"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func splineFramesNode(spline curves.Spline, spacing float64) *nodes.Struct[curves.SplineFramesNode] {
	return &nodes.Struct[curves.SplineFramesNode]{Data: curves.SplineFramesNode{
		Spline:  nodes.ConstOutput[curves.Spline]{Val: spline},
		Spacing: nodes.ConstOutput[float64]{Val: spacing},
	}}
}

func TestSplineFramesSpaceFramesEvenlyAndFaceAlongTheSpline(t *testing.T) {
	spline := curves.CatmullRomSplineParameters{
		Points: []vector3.Float64{vector3.New(0., 0., 0.), vector3.New(1., 0., 0.), vector3.New(2., 0., 0.), vector3.New(3., 0., 0.)},
		Alpha:  0.5,
	}.Spline()
	node := splineFramesNode(&spline, 0.25)

	frames := nodes.GetNodeOutputPort[[]trs.TRS](node, "Frames").Value()
	along := nodes.GetNodeOutputPort[[]float64](node, "Along").Value()
	require.Len(t, frames, len(along))
	require.Greater(t, len(frames), 2)

	for i, f := range frames {
		assert.InDelta(t, 0, f.RotateDirection(vector3.Forward[float64]()).Distance(vector3.Right[float64]()), 1e-6, "frame %d travels along +X", i)
		assert.InDelta(t, 0, f.RotateDirection(vector3.Up[float64]()).Distance(vector3.Up[float64]()), 1e-6, "frame %d stays upright", i)
	}
	assert.InDelta(t, 0, along[0], 1e-9)
	assert.InDelta(t, 1, along[len(along)-1], 1e-9)
	gap := frames[1].Position().Distance(frames[0].Position())
	for i := 2; i < len(frames); i++ {
		assert.InDelta(t, gap, frames[i].Position().Distance(frames[i-1].Position()), 0.02)
	}
	assert.False(t, nodes.GetNodeOutputPort[bool](node, "Closed").Value())
}

func TestSplineFramesNoticeALoop(t *testing.T) {
	points := make([]vector3.Float64, 8)
	for i := range points {
		a := 2 * math.Pi * float64(i) / 8
		points[i] = vector3.New(math.Cos(a), 0, math.Sin(a))
	}
	spline := curves.CatmullRomSplineParameters{Points: points, Alpha: 0.5, Closed: true}.Spline()
	node := splineFramesNode(&spline, 0.1)

	assert.True(t, nodes.GetNodeOutputPort[bool](node, "Closed").Value())
	frames := nodes.GetNodeOutputPort[[]trs.TRS](node, "Frames").Value()
	assert.Greater(t, frames[len(frames)-1].Position().Distance(frames[0].Position()), 0.01, "the first point is not repeated")
}
