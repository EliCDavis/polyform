package curves_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/curves"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func finite(v vector3.Float64) bool {
	for _, c := range []float64{v.X(), v.Y(), v.Z()} {
		if math.IsNaN(c) || math.IsInf(c, 0) {
			return false
		}
	}
	return true
}

func threePoints() []vector3.Float64 {
	return []vector3.Float64{
		vector3.New(0., 0., 0.),
		vector3.New(1., 1., 0.),
		vector3.New(2., 0., 0.),
	}
}

func TestCentripetalSplineWithThreePointsIsFinite(t *testing.T) {
	for name, alpha := range map[string]float64{"centripetal": 0.5, "chordal": 1} {
		t.Run(name, func(t *testing.T) {
			spline := curves.CatmullRomSplineParameters{Points: threePoints(), Alpha: alpha}.Spline()

			length := spline.Length()
			require.False(t, math.IsNaN(length))
			assert.Greater(t, length, 2*math.Sqrt2-1e-9, "shorter than the straight segments")

			mid := spline.At(length / 2)
			assert.True(t, finite(mid), "%v", mid)
			assert.InDelta(t, 0, mid.Sub(vector3.New(1., 1., 0.)).Length(), 1e-3)
			assert.InDelta(t, 0, spline.At(0).Sub(threePoints()[0]).Length(), 1e-9)
			assert.InDelta(t, 0, spline.At(length).Sub(threePoints()[2]).Length(), 1e-9)
		})
	}
}

func TestCentripetalSplineWithTwoPointsIsAFiniteLine(t *testing.T) {
	points := []vector3.Float64{vector3.New(0., 0., 0.), vector3.New(0., 3., 4.)}
	spline := curves.CatmullRomSplineParameters{Points: points, Alpha: 0.5}.Spline()

	assert.InDelta(t, 5, spline.Length(), 1e-3)
	assert.InDelta(t, 0, spline.At(2.5).Sub(vector3.New(0., 1.5, 2.)).Length(), 1e-3)
}

func TestCatmullRomSplineNodeThreePointsCentripetalLength(t *testing.T) {
	splineNode := &nodes.Struct[curves.CatmullRomSplineNode]{
		Data: curves.CatmullRomSplineNode{
			Points: nodes.ConstOutput[[]vector3.Float64]{Val: threePoints()},
			Alpha:  nodes.ConstOutput[float64]{Val: 0.5},
		},
	}
	lengthNode := &nodes.Struct[curves.LengthNode]{
		Data: curves.LengthNode{Spline: nodes.GetNodeOutputPort[curves.Spline](splineNode, "Out")},
	}

	length := nodes.GetNodeOutputPort[float64](lengthNode, "Out").Value()
	assert.False(t, math.IsNaN(length))
	assert.Greater(t, length, 0.)
}
