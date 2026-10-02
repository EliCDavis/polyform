package sdf_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestRoundedCylinderExtentIncludesTheRounding(t *testing.T) {
	const radius, rounding, bodyHeight = 0.5, 0.25, 1.
	field := sdf.RoundedCylinder(vector3.Zero[float64](), radius, rounding, bodyHeight)

	tests := map[string]struct {
		at   vector3.Float64
		want float64
	}{
		"the cap is a whole rounding radius past BodyHeight": {
			at:   vector3.New(0., bodyHeight+rounding, 0.),
			want: 0,
		},
		"BodyHeight alone is still well inside": {
			at:   vector3.New(0., bodyHeight, 0.),
			want: -rounding,
		},
		"the side wall sits exactly at Radius": {
			at:   vector3.New(radius, 0., 0.),
			want: 0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.InDelta(t, tc.want, field(tc.at), 1e-12)
		})
	}
}
