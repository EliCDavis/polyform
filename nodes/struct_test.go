package nodes_test

import (
	"testing"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type SimpleTestStructNode = nodes.Struct[SimpleAddTestStruct]

type SimpleAddTestStruct struct {
	A nodes.Output[float64]
	B nodes.Output[float64]
}

func (and SimpleAddTestStruct) Sum() nodes.StructOutput[float64] {
	result := nodes.NewStructOutput(0.)
	result.Set(and.A.Value() + and.B.Value())
	return result
}

func TestStruct_SimpleAddTest(t *testing.T) {
	var n nodes.Node = &SimpleTestStructNode{
		Data: SimpleAddTestStruct{
			A: nodes.NewValue(1.).Outputs()["Value"].(nodes.Output[float64]),
			B: nodes.NewValue(2.).Outputs()["Value"].(nodes.Output[float64]),
		},
	}

	output := n.Outputs()
	assert.Len(t, output, 1)
	// assert.Contains(t, "Sum", output)
	require.Contains(t, output, "Sum")

	sumOutput := output["Sum"].(nodes.Output[float64])
	assert.Equal(t, 3., sumOutput.Value())
	assert.Equal(t, 0, sumOutput.Version())
	assert.Equal(t, "Sum", sumOutput.Name())
	assert.Equal(t, n, sumOutput.Node())
}
