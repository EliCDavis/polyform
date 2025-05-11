package repeat_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGrid(t *testing.T) {
	var tests = map[string]struct {
		rows   int
		column int
		width  float64
		height float64

		expect []vector2.Float64
	}{
		"zero": {
			rows: 1, column: 1, width: 1, height: 1,
			expect: []vector2.Float64{{}},
		},
		"0 columns": {
			rows: 1, column: 0, width: 1, height: 1,
			expect: []vector2.Float64{},
		},
		"0 rows": {
			rows: 0, column: 1, width: 1, height: 1,
			expect: []vector2.Float64{},
		},
		"vertical line": {
			rows: 3, column: 1, width: 1, height: 1,
			expect: []vector2.Float64{
				vector2.New(0., -0.5),
				vector2.New(0., 0.0),
				vector2.New(0., 0.5),
			},
		},
		"horizontal line": {
			rows: 1, column: 3, width: 1, height: 1,
			expect: []vector2.Float64{
				vector2.New(-0.5, 0.),
				vector2.New(0.0, 0.),
				vector2.New(0.5, 0.),
			},
		},
		"4 corners": {
			rows: 2, column: 2, width: 1, height: 1,
			expect: []vector2.Float64{
				vector2.New(-0.5, -0.5),
				vector2.New(0.5, -0.5),
				vector2.New(-0.5, 0.5),
				vector2.New(0.5, 0.5),
			},
		},
	}

	for testName, tc := range tests {
		t.Run(testName, func(t *testing.T) {
			gn := &nodes.Struct[repeat.GridNode]{
				Data: repeat.GridNode{
					Rows:    nodes.ConstOutput[int]{Val: tc.rows},
					Columns: nodes.ConstOutput[int]{Val: tc.column},
					Width:   nodes.ConstOutput[float64]{Val: tc.width},
					Height:  nodes.ConstOutput[float64]{Val: tc.height},
				},
			}

			outputs := gn.Outputs()
			out := outputs["Vector 2"].(nodes.Output[[]vector2.Float64]).Value()

			require.Len(t, out, len(tc.expect))
			delta := 0.0000001
			for i, v := range out {
				assert.InDelta(t, tc.expect[i].X(), v.X(), delta, "[%d]X", i)
				assert.InDelta(t, tc.expect[i].Y(), v.Y(), delta, "[%d]Y", i)
			}
		})
	}
}
