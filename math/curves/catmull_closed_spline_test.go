package curves_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/curves"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func regularPolygon(sides int) []vector3.Float64 {
	points := make([]vector3.Float64, sides)
	for i := range sides {
		angle := 2 * math.Pi * float64(i) / float64(sides)
		points[i] = vector3.New(math.Cos(angle), 0, math.Sin(angle))
	}
	return points
}

var closedAlphas = map[string]float64{"uniform": 0, "centripetal": 0.5, "chordal": 1}

func TestClosedSplineIsSmoothAcrossTheSeam(t *testing.T) {
	for _, sides := range []int{3, 4, 5, 8} {
		for name, alpha := range closedAlphas {
			t.Run(fmt.Sprintf("%d sides %s", sides, name), func(t *testing.T) {
				points := regularPolygon(sides)
				spline := curves.CatmullRomSplineParameters{Points: points, Alpha: alpha, Closed: true}.Spline()
				length := spline.Length()
				require.False(t, math.IsNaN(length))

				assert.InDelta(t, 0, spline.At(0).Distance(points[0]), 1e-9)
				assert.InDelta(t, 0, spline.At(length).Distance(points[0]), 1e-9)

				const step = 1e-4
				leaving := spline.At(step).Sub(spline.At(0)).Normalized()
				arriving := spline.At(length).Sub(spline.At(length - step)).Normalized()
				assert.Greater(t, leaving.Dot(arriving), 0.999, "sides=%d leaving=%v arriving=%v", sides, leaving, arriving)
			})
		}
	}
}

func TestClosedRegularPolygonSplinePassesEveryPointAtEvenSpacing(t *testing.T) {
	for _, sides := range []int{3, 4, 6} {
		for name, alpha := range closedAlphas {
			t.Run(fmt.Sprintf("%d sides %s", sides, name), func(t *testing.T) {
				points := regularPolygon(sides)
				spline := curves.CatmullRomSplineParameters{Points: points, Alpha: alpha, Closed: true}.Spline()
				length := spline.Length()

				for i, p := range points {
					at := spline.At(length * float64(i) / float64(sides))
					assert.InDelta(t, 0, at.Distance(p), 1e-2, "sides=%d point %d: got %v want %v", sides, i, at, p)
				}
			})
		}
	}
}

func TestClosedSplineIgnoresARepeatedClosingPoint(t *testing.T) {
	for name, alpha := range closedAlphas {
		t.Run(name, func(t *testing.T) {
			points := regularPolygon(5)
			repeated := append(append([]vector3.Float64{}, points...), points[0])

			want := curves.CatmullRomSplineParameters{Points: points, Alpha: alpha, Closed: true}.Spline()
			got := curves.CatmullRomSplineParameters{Points: repeated, Alpha: alpha, Closed: true}.Spline()

			require.False(t, math.IsNaN(got.Length()))
			assert.InDelta(t, want.Length(), got.Length(), 1e-9)
		})
	}
}

func TestClosedTwoPointSplineReturnsToTheStart(t *testing.T) {
	points := []vector3.Float64{vector3.New(0., 0., 0.), vector3.New(0., 0., 1.)}
	spline := curves.CatmullRomSplineParameters{Points: points, Alpha: 0.5, Closed: true}.Spline()

	length := spline.Length()
	assert.InDelta(t, 2, length, 1e-3)
	assert.True(t, finite(spline.At(length/4)))
	assert.InDelta(t, 0, spline.At(length/2).Distance(points[1]), 1e-3)
	assert.InDelta(t, 0, spline.At(length).Distance(points[0]), 1e-9)
}
