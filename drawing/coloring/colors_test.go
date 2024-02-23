package coloring_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/stretchr/testify/assert"
)

func TestColor(t *testing.T) {
	tests := map[string]struct {
		input coloring.WebColor
		wantR byte
		wantG byte
		wantB byte
		wantA byte
	}{
		"black": {input: coloring.Black(), wantR: 0, wantG: 0, wantB: 0, wantA: 255},
		"white": {input: coloring.White(), wantR: 255, wantG: 255, wantB: 255, wantA: 255},
		"red":   {input: coloring.Red(), wantR: 255, wantG: 0, wantB: 0, wantA: 255},
		"green": {input: coloring.Green(), wantR: 0, wantG: 255, wantB: 0, wantA: 255},
		"blue":  {input: coloring.Blue(), wantR: 0, wantG: 0, wantB: 255, wantA: 255},
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
