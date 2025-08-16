package coloring_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/stretchr/testify/assert"
)

func TestSRGBToLinear(t *testing.T) {
	tests := map[string]struct {
		input float64
		want  float64
	}{
		"0 => 0":      {input: 0, want: 0},
		"1 => 1":      {input: 1, want: 1},
		"0.73 => 0.5": {input: 0.73, want: 0.4919050408},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.InDelta(t, tc.want, coloring.SRGBToLinear(tc.input), 0.0000001)
		})
	}
}

func TestLinearToSRGB(t *testing.T) {
	tests := map[string]struct {
		input float64
		want  float64
	}{
		"0 => 0":      {input: 0, want: 0},
		"1 => 1":      {input: 1, want: 1},
		"0.5 => 0.73": {input: 0.5, want: 0.735360635},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.InDelta(t, tc.want, coloring.LinearToSRGB(tc.input), 0.0000001)
		})
	}
}
