package extrude_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The outline's frame follows the path direction; the OutlineNode port
// description states each mapping, and this pins them.
func TestOutlineFrameByPathDirection(t *testing.T) {
	// A triangle whose only far point is at outline x = +2.
	pointer := []vector2.Float64{vector2.New(0., -1.), vector2.New(2., 0.), vector2.New(0., 1.)}

	cases := map[string]struct {
		dir   vector3.Float64
		xAxis vector3.Float64
		yAxis vector3.Float64
	}{
		"up":            {vector3.Up[float64](), vector3.Right[float64](), vector3.Forward[float64]()},
		"toward viewer": {vector3.Backwards[float64](), vector3.Right[float64](), vector3.Up[float64]()},
		"away":          {vector3.Forward[float64](), vector3.Left[float64](), vector3.Up[float64]()},
		"right":         {vector3.Right[float64](), vector3.Down[float64](), vector3.Forward[float64]()},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			mesh, err := extrude.Outline{
				Shape: pointer,
				Path:  []vector3.Float64{vector3.Zero[float64](), c.dir.Scale(3)},
			}.Extrude()
			require.NoError(t, err)

			positions := mesh.Float3Attribute(modeling.PositionAttribute)
			var far, top vector3.Float64
			for i := 0; i < positions.Len(); i++ {
				p := positions.At(i)
				if p.Dot(c.xAxis) > far.Dot(c.xAxis) {
					far = p
				}
				if p.Dot(c.yAxis) > top.Dot(c.yAxis) {
					top = p
				}
			}
			assert.InDelta(t, 2, far.Dot(c.xAxis), 1e-9, "outline +x should land on %v", c.xAxis)
			assert.InDelta(t, 1, top.Dot(c.yAxis), 1e-9, "outline +y should land on %v", c.yAxis)
		})
	}
}
