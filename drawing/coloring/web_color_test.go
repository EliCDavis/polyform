package coloring_test

import (
	"encoding/json"
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/stretchr/testify/assert"
)

func TestWebColor(t *testing.T) {
	tests := map[string]struct {
		input        string
		remarshalled string
		want         coloring.Color
	}{
		"#000000":   {remarshalled: "\"#000000\"", input: "\"#000000\"", want: coloring.Color{R: 0, G: 0, B: 0, A: 1}},
		"#ffffff":   {remarshalled: "\"#ffffff\"", input: "\"#ffffff\"", want: coloring.Color{R: 1, G: 1, B: 1, A: 1}},
		"#00f100":   {remarshalled: "\"#00f100\"", input: "\"#00f100\"", want: coloring.Color{R: 0, G: 241 / 255., B: 0, A: 1}},
		"#FFFFFF":   {remarshalled: "\"#ffffff\"", input: "\"#FFFFFF\"", want: coloring.Color{R: 1, G: 1, B: 1, A: 1}},
		"#FFFFFFFF": {remarshalled: "\"#ffffff\"", input: "\"#FFFFFFFF\"", want: coloring.Color{R: 1, G: 1, B: 1, A: 1}},
		"#FFFFFFf1": {remarshalled: "\"#fffffff1\"", input: "\"#FFFFFFf1\"", want: coloring.Color{R: 1, G: 1, B: 1, A: 241 / 255.}},

		"#fff":  {remarshalled: "\"#ffffff\"", input: "\"#fff\"", want: coloring.Color{R: 1, G: 1, B: 1, A: 1}},
		"#FFF":  {remarshalled: "\"#ffffff\"", input: "\"#FFF\"", want: coloring.Color{R: 1, G: 1, B: 1, A: 1}},
		"#0F0":  {remarshalled: "\"#00ff00\"", input: "\"#0F0\"", want: coloring.Color{R: 0, G: 1, B: 0, A: 1}},
		"#000":  {remarshalled: "\"#000000\"", input: "\"#000\"", want: coloring.Color{R: 0, G: 0, B: 0, A: 1}},
		"#000F": {remarshalled: "\"#000000\"", input: "\"#000F\"", want: coloring.Color{R: 0, G: 0, B: 0, A: 1}},
		"#0000": {remarshalled: "\"#00000000\"", input: "\"#0000\"", want: coloring.Color{R: 0, G: 0, B: 0, A: 0}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			parsedColor := coloring.Color{}

			// Test Unmarshall
			err := json.Unmarshal([]byte(tc.input), &parsedColor)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, parsedColor)

			// Test Marshall
			backToString, err := json.Marshal(parsedColor)
			assert.NoError(t, err)
			assert.Equal(t, []byte(tc.remarshalled), backToString)

			// Test RGBA values
			wR, wG, wB, wA := parsedColor.RGBA()
			r8, g8, b8, a8 := parsedColor.RGBA8().RGBA()
			assert.Equal(t, r8, wR)
			assert.Equal(t, g8, wG)
			assert.Equal(t, b8, wB)
			assert.Equal(t, a8, wA)
		})
	}
}

func TestUnmarshallWebColorErrors(t *testing.T) {
	tests := map[string]struct {
		input string
		want  string
	}{
		"#ZZ0000":   {input: "\"#ZZ0000\"", want: "unable to parse r component of color '\"#ZZ0000\"': strconv.ParseUint: parsing \"ZZ\": invalid syntax"},
		"#00ZZ00":   {input: "\"#00ZZ00\"", want: "unable to parse g component of color '\"#00ZZ00\"': strconv.ParseUint: parsing \"ZZ\": invalid syntax"},
		"#0000ZZ":   {input: "\"#0000ZZ\"", want: "unable to parse b component of color '\"#0000ZZ\"': strconv.ParseUint: parsing \"ZZ\": invalid syntax"},
		"#000000ZZ": {input: "\"#000000ZZ\"", want: "unable to parse a component of color '\"#000000ZZ\"': strconv.ParseUint: parsing \"ZZ\": invalid syntax"},

		"#Z00":  {input: "\"#Z00\"", want: "unable to parse r component of color '\"#Z00\"': strconv.ParseUint: parsing \"ZZ\": invalid syntax"},
		"#0Z0":  {input: "\"#0Z0\"", want: "unable to parse g component of color '\"#0Z0\"': strconv.ParseUint: parsing \"ZZ\": invalid syntax"},
		"#00Z":  {input: "\"#00Z\"", want: "unable to parse b component of color '\"#00Z\"': strconv.ParseUint: parsing \"ZZ\": invalid syntax"},
		"#000Z": {input: "\"#000Z\"", want: "unable to parse a component of color '\"#000Z\"': strconv.ParseUint: parsing \"ZZ\": invalid syntax"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			parsedColor := coloring.Color{}
			err := json.Unmarshal([]byte(tc.input), &parsedColor)

			assert.EqualError(t, err, tc.want)
		})
	}
}
