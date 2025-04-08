package utils_test

import (
	"testing"

	"github.com/EliCDavis/polyform/utils"
	"github.com/stretchr/testify/assert"
)

func TestCamelCaseToSpaceCase(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		In  string
		Out string
	}{
		"empty": {
			In:  "",
			Out: "",
		},
		"Basic": {
			In:  "Basic",
			Out: "Basic",
		},
		"MultipleWords": {
			In:  "MultipleWords",
			Out: "Multiple Words",
		},
		"ABC": {
			In:  "ABC",
			Out: "ABC",
		},
		"UVsTest": {
			In:  "UVsTest",
			Out: "UVs Test",
		},
		"Number": {
			In:  "My3D",
			Out: "My 3D",
		},
		"Multi-digit number": {
			In:  "float64",
			Out: "float 64",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel() // marks each test case as capable of running in parallel with each other

			assert.Equal(t, test.Out, utils.CamelCaseToSpaceCase(test.In))
		})
	}
}
