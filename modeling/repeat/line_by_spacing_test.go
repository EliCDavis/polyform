package repeat_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lineBySpacing(start, end vector3.Float64, maxSpacing float64) []trs.TRS {
	return repeat.LineBySpacing{Line: geometry.NewLine3D(start, end), MaxSpacing: maxSpacing}.TRS()
}

func zPositions(transforms []trs.TRS) []float64 {
	out := make([]float64, len(transforms))
	for i, t := range transforms {
		out[i] = t.Position().Z()
	}
	return out
}

func TestLineBySpacingUsesTheFewestTransformsThatFitTheSpacing(t *testing.T) {
	origin := vector3.Zero[float64]()
	assert.InDeltaSlice(t, []float64{3, 6}, zPositions(lineBySpacing(origin, vector3.New(0., 0., 9.), 3)), 1e-9)
	assert.InDeltaSlice(t, []float64{2.5, 5, 7.5}, zPositions(lineBySpacing(origin, vector3.New(0., 0., 10.), 3)), 1e-9)
	assert.Len(t, lineBySpacing(origin, vector3.New(0., 0., 0.9), 0.3), 2, "0.9/0.3 rounding up must not add a point")
}

func TestLineBySpacingLeavesShortLinesEmpty(t *testing.T) {
	origin := vector3.Zero[float64]()
	assert.Empty(t, lineBySpacing(origin, vector3.New(0., 0., 3.), 3))
	assert.Empty(t, lineBySpacing(origin, vector3.New(0., 0., 2.), 3))
	assert.Empty(t, lineBySpacing(origin, origin, 3))
}

func TestLineBySpacingFacesAlongTheLineIncludingItsSlope(t *testing.T) {
	for name, end := range map[string]vector3.Float64{
		"sloped":   vector3.New(3., 4., 0.),
		"vertical": vector3.New(0., 5., 0.),
		"backward": vector3.New(0., 0., -5.),
	} {
		t.Run(name, func(t *testing.T) {
			transforms := lineBySpacing(vector3.Zero[float64](), end, 1)
			require.NotEmpty(t, transforms)
			facing := transforms[0].Rotation().Rotate(vector3.Forward[float64]())
			assert.InDelta(t, 0, facing.Distance(end.Normalized()), 1e-9)
		})
	}
}

func TestLineBySpacingNode(t *testing.T) {
	line := geometry.NewLine3D(vector3.Zero[float64](), vector3.New(0., 0., 9.))
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"places transforms",
			nodetest.NewNode(repeat.LineBySpacingNode{
				Line:       nodes.ConstOutput[geometry.Line3D]{Val: line},
				MaxSpacing: nodes.ConstOutput[float64]{Val: 3},
			}),
			nodetest.AssertOutput("Out", lineBySpacing(vector3.Zero[float64](), vector3.New(0., 0., 9.), 3)),
		),
		nodetest.NewTestCase(
			"non-positive spacing is reported",
			nodetest.NewNode(repeat.LineBySpacingNode{
				Line:       nodes.ConstOutput[geometry.Line3D]{Val: line},
				MaxSpacing: nodes.ConstOutput[float64]{Val: 0},
			}),
			nodetest.AssertOutputError("Out", "not positive"),
		),
	).Run(t)
}
