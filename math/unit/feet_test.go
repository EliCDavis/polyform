package unit_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/unit"
	"github.com/stretchr/testify/assert"
)

func TestParseFeet(t *testing.T) {

	tests := map[string]struct {
		In    string
		Out   float64
		Error string
	}{
		"1": {
			In:  "1",
			Out: 1,
		},
		"1?": {
			In:    "1?",
			Out:   0,
			Error: "no feet or inches markings, can't parse 1?: strconv.ParseFloat: parsing \"1?\": invalid syntax",
		},
		"1'": {
			In:  "1'",
			Out: 1,
		},
		"uhh'": {
			In:    "uhhh'",
			Out:   0,
			Error: "unable to parse feet \"uhhh\": strconv.ParseFloat: parsing \"uhhh\": invalid syntax",
		},
		"6\"": {
			In:  "6\"",
			Out: 0.5,
		},
		"1' 6\"": {
			In:  "1' 6\"",
			Out: 1.5,
		},
		"   1'        6\"     ": {
			In:  "   1'        6\"     ",
			Out: 1.5,
		},
		"uhh\"": {
			In:    "uhh\"",
			Out:   0,
			Error: "unable to parse inches \"uhh\": strconv.ParseFloat: parsing \"uhh\": invalid syntax",
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			feet, err := unit.ParseFeet(testCase.In)
			assert.Equal(t, testCase.Out, feet)
			if testCase.Error == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, testCase.Error)
			}
		})
	}

}
