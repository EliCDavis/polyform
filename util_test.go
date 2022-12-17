package mesh_test

import (
	"testing"

	"github.com/EliCDavis/mesh"
	"github.com/stretchr/testify/assert"
)

func TestClamp(t *testing.T) {
	tests := map[string]struct {
		min  float64
		max  float64
		v    float64
		want float64
	}{
		"no clamping": {
			min:  0,
			max:  1,
			v:    0.5,
			want: 0.5,
		},
		"clamps lower bound": {
			min:  0,
			max:  1,
			v:    -.0001,
			want: 0,
		},
		"clamps upper bound": {
			min:  0,
			max:  1,
			v:    1.0001,
			want: 1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, mesh.Clamp(tc.v, tc.min, tc.max))
		})
	}
}
