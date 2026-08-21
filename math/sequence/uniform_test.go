package sequence_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/sequence"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinearNode_SingleSample(t *testing.T) {
	node := nodes.Struct[sequence.LinearNode]{
		Data: sequence.LinearNode{
			Start:   nodes.ConstOutput[float64]{Val: 0.5},
			End:     nodes.ConstOutput[float64]{Val: 9.0},
			Samples: nodes.ConstOutput[int]{Val: 1},
		},
	}

	out := node.Outputs()["Out"]
	require.NotNil(t, out)

	val := out.(nodes.Output[[]float64]).Value()
	require.Len(t, val, 1)
	assert.False(t, math.IsNaN(val[0]), "single-sample LinearNode produced NaN instead of Start")
	assert.Equal(t, 0.5, val[0])
}

func TestLinearNode_MultiSample(t *testing.T) {
	node := nodes.Struct[sequence.LinearNode]{
		Data: sequence.LinearNode{
			Start:   nodes.ConstOutput[float64]{Val: 0},
			End:     nodes.ConstOutput[float64]{Val: 4},
			Samples: nodes.ConstOutput[int]{Val: 5},
		},
	}

	out := node.Outputs()["Out"]
	require.NotNil(t, out)

	val := out.(nodes.Output[[]float64]).Value()
	assert.Equal(t, []float64{0, 1, 2, 3, 4}, val)
}
