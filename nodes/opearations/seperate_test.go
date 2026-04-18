package opearations_test

import (
	"testing"

	"github.com/EliCDavis/polyform/nodes/opearations"
	"github.com/stretchr/testify/assert"
)

func TestSeperate(t *testing.T) {

	tests := map[string]struct {
		in            []int
		keep          []bool
		keptResult    []int
		removedResult []int
	}{
		"basic": {
			in:            []int{1, 2, 3, 4},
			keep:          []bool{true, false, true, false},
			keptResult:    []int{1, 3},
			removedResult: []int{4, 2},
		},
		"lacking keep": {
			in:            []int{1, 2, 3, 4},
			keep:          []bool{true, false},
			keptResult:    []int{1},
			removedResult: []int{4, 3, 2},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			kept, removed := opearations.Seperate(tc.in, tc.keep)
			if assert.Len(t, kept, len(tc.keptResult)) {
				for i, v := range kept {
					assert.Equal(t, tc.keptResult[i], v)
				}
			}

			if assert.Len(t, removed, len(tc.removedResult)) {
				for i, v := range removed {
					assert.Equal(t, tc.removedResult[i], v)
				}
			}
		})
	}

}
