package geometry_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector"
	"github.com/stretchr/testify/assert"
)

func TestLine3D_ClosestPoint(t *testing.T) {
	p1 := vector.NewVector3(-1, 0, 0)
	p2 := vector.NewVector3(1, 0, 0)
	line := geometry.NewLine3D(p1, p2)

	tests := map[string]struct {
		input vector.Vector3
		want  vector.Vector3
	}{
		"mid point":           {input: vector.NewVector3(0, 0, 0), want: vector.NewVector3(0, 0, 0)},
		"right point":         {input: vector.NewVector3(1, 0, 0), want: vector.NewVector3(1, 0, 0)},
		"left point":          {input: vector.NewVector3(-1, 0, 0), want: vector.NewVector3(-1, 0, 0)},
		"clamped right point": {input: vector.NewVector3(5, 0, 0), want: vector.NewVector3(1, 0, 0)},
		"clamped left point":  {input: vector.NewVector3(-5, 0, 0), want: vector.NewVector3(-1, 0, 0)},

		"+1Y mid point":           {input: vector.NewVector3(0, 1, 0), want: vector.NewVector3(0, 0, 0)},
		"+1Y right point":         {input: vector.NewVector3(1, 1, 0), want: vector.NewVector3(1, 0, 0)},
		"+1Y left point":          {input: vector.NewVector3(-1, 1, 0), want: vector.NewVector3(-1, 0, 0)},
		"+1Y clamped right point": {input: vector.NewVector3(5, 1, 0), want: vector.NewVector3(1, 0, 0)},
		"+1Y clamped left point":  {input: vector.NewVector3(-5, 1, 0), want: vector.NewVector3(-1, 0, 0)},

		"-1Y mid point":           {input: vector.NewVector3(0, -1, 0), want: vector.NewVector3(0, 0, 0)},
		"-1Y right point":         {input: vector.NewVector3(1, -1, 0), want: vector.NewVector3(1, 0, 0)},
		"-1Y left point":          {input: vector.NewVector3(-1, -1, 0), want: vector.NewVector3(-1, 0, 0)},
		"-1Y clamped right point": {input: vector.NewVector3(5, -1, 0), want: vector.NewVector3(1, 0, 0)},
		"-1Y clamped left point":  {input: vector.NewVector3(-5, -1, 0), want: vector.NewVector3(-1, 0, 0)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := line.ClosestPointOnLine(tc.input)
			assert.InDelta(t, tc.want.X(), got.X(), 0.0001)
			assert.InDelta(t, tc.want.Y(), got.Y(), 0.0001)
			assert.InDelta(t, tc.want.Z(), got.Z(), 0.0001)
		})
	}
}
