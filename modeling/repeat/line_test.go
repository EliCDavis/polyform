package repeat_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestLine(t *testing.T) {

	tests := map[string]struct {
		line repeat.Line
		want []trs.TRS
	}{
		"empty line/no samples/non-exclusive/no trs": {
			line: repeat.Line{
				Exclusive: false,
			},
			want: []trs.TRS{},
		},
		"empty line/no samples/exclusive/no trs": {
			line: repeat.Line{
				Exclusive: true,
			},
			want: []trs.TRS{},
		},
		"1 sample/exclusive/middle of line": {
			line: repeat.Line{
				Start:     vector3.Zero[float64](),
				End:       vector3.Up[float64](),
				Exclusive: true,
				Samples:   1,
			},
			want: []trs.TRS{
				trs.Position(vector3.New(0., 0.5, 0.)),
			},
		},
		"1 sample/non-exclusive/middle of line": {
			line: repeat.Line{
				Start:     vector3.Zero[float64](),
				End:       vector3.Up[float64](),
				Exclusive: false,
				Samples:   1,
			},
			want: []trs.TRS{
				trs.Position(vector3.New(0., 0.5, 0.)),
			},
		},
		"2 sample/non-exclusive/start and end": {
			line: repeat.Line{
				Start:     vector3.Zero[float64](),
				End:       vector3.Up[float64](),
				Exclusive: false,
				Samples:   2,
			},
			want: []trs.TRS{
				trs.Position(vector3.New(0., 0., 0.)),
				trs.Position(vector3.New(0., 1., 0.)),
			},
		},
		"2 sample/exclusive/33 and 66": {
			line: repeat.Line{
				Start:     vector3.Zero[float64](),
				End:       vector3.Up[float64](),
				Exclusive: true,
				Samples:   2,
			},
			want: []trs.TRS{
				trs.Position(vector3.New(0., 1./3., 0.)),
				trs.Position(vector3.New(0., 2./3., 0.)),
			},
		},
		"3 sample/nonexclusive/start middle end": {
			line: repeat.Line{
				Start:     vector3.Zero[float64](),
				End:       vector3.Up[float64](),
				Exclusive: false,
				Samples:   3,
			},
			want: []trs.TRS{
				trs.Position(vector3.New(0., 0., 0.)),
				trs.Position(vector3.New(0., 0.5, 0.)),
				trs.Position(vector3.New(0., 1., 0.)),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.line.TRS())
		})
	}

}
