package modeling_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestVectorIntRoundsProperly(t *testing.T) {
	tests := map[string]struct {
		input vector3.Float64
		want  modeling.VectorInt
		pow   int
	}{
		"0,1,2=>0,1,2":         {input: vector3.New(0., 1., 2.), want: modeling.VectorInt{0, 1, 2}, pow: 0},
		"0.1,1.1,2.1=>0,1,2":   {input: vector3.New(0.1, 1.1, 2.1), want: modeling.VectorInt{0, 1, 2}, pow: 0},
		"0.1,1.1,2.1=>1,11,21": {input: vector3.New(0.1, 1.1, 2.1), want: modeling.VectorInt{1, 11, 21}, pow: 1},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, modeling.Vector3ToInt(tc.input, tc.pow))
		})
	}
}
