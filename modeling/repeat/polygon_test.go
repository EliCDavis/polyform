package repeat_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolygon(t *testing.T) {
	tests := map[string]struct {
		times  int
		sides  int
		radius float64

		expect []trs.TRS
	}{
		"0 times": {
			times:  0,
			sides:  3,
			radius: 1,

			expect: nil,
		},
		"square": {
			times:  4,
			sides:  4,
			radius: 1,

			// pi / 2
			// 1/2 pi
			// 4/8 pi

			expect: []trs.TRS{
				trs.New(
					vector3.New(.5, 0., .5),
					quaternion.FromTheta(((1./8.)*(2*math.Pi))-(math.Pi/2), vector3.New(0., -1, 0.)),
					vector3.One[float64](),
				),
				trs.New(
					vector3.New(-.5, 0., .5),
					quaternion.FromTheta(((3./8.)*(2*math.Pi))-(math.Pi/2), vector3.New(0., -1, 0.)),
					vector3.One[float64](),
				),
				trs.New(
					vector3.New(-.5, 0., -.5),
					quaternion.FromTheta(((5./8.)*(2*math.Pi))-(math.Pi/2), vector3.New(0., -1, 0.)),
					vector3.One[float64](),
				),
				trs.New(
					vector3.New(.5, 0., -.5),
					quaternion.FromTheta(((7./8.)*(2*math.Pi))-(math.Pi/2), vector3.New(0., -1, 0.)),
					vector3.One[float64](),
				),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actual := repeat.Polygon(tc.times, tc.sides, tc.radius)
			require.Len(t, actual, len(tc.expect))

			for i, v := range actual {
				assert.NoError(t, v.WithinDelta(tc.expect[i], 0.000000001), fmt.Sprintf("item %d", i))
			}
		})
	}

	assert.PanicsWithError(t, "polygon can not have 0 sides", func() {
		repeat.Polygon(0, 0, 0)
	})
}
