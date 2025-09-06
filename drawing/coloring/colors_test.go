package coloring_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/stretchr/testify/assert"
)

func TestColor(t *testing.T) {
	tests := map[string]struct {
		input coloring.Color
		wantR float64
		wantG float64
		wantB float64
		wantA float64
	}{
		"black": {input: coloring.Black(), wantR: 0, wantG: 0, wantB: 0, wantA: 1},
		"white": {input: coloring.White(), wantR: 1, wantG: 1, wantB: 1, wantA: 1},
		// "red":   {input: coloring.Red(), wantR: 1, wantG: 0, wantB: 0, wantA: 1},
		// "green": {input: coloring.Green(), wantR: 0, wantG: 1, wantB: 0, wantA: 1},
		// "blue":  {input: coloring.Blue(), wantR: 0, wantG: 0, wantB: 1, wantA: 1},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.wantR, tc.input.R)
			assert.Equal(t, tc.wantG, tc.input.G)
			assert.Equal(t, tc.wantB, tc.input.B)
			assert.Equal(t, tc.wantA, tc.input.A)
		})
	}
}
