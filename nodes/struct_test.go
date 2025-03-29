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

func TestStruct_SimpleAdd(t *testing.T) {
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

// ================================================================================================

type ArrayTestStructNode = nodes.Struct[ArrayTestStruct]

type ArrayTestStruct struct {
	Values []nodes.Output[float64]
}

func (and ArrayTestStruct) Sum() nodes.StructOutput[float64] {
	sum := 0.
	for _, v := range and.Values {
		sum += v.Value()
	}
	result := nodes.NewStructOutput(sum)
	return result
}

func TestStruct_ArrayAdd(t *testing.T) {
	var n nodes.Node = &ArrayTestStructNode{
		Data: ArrayTestStruct{
			Values: []nodes.Output[float64]{
				nodes.NewValue(1.).Outputs()["Value"].(nodes.Output[float64]),
				nodes.NewValue(2.).Outputs()["Value"].(nodes.Output[float64]),
				nodes.NewValue(3.).Outputs()["Value"].(nodes.Output[float64]),
			},
		},
	}

	output := n.Outputs()
	assert.Len(t, output, 1)
	// assert.Contains(t, "Sum", output)
	require.Contains(t, output, "Sum")

	sumOutput := output["Sum"].(nodes.Output[float64])
	assert.Equal(t, 6., sumOutput.Value())
	assert.Equal(t, 0, sumOutput.Version())
	assert.Equal(t, "Sum", sumOutput.Name())
	assert.Equal(t, n, sumOutput.Node())
}

func TestStruct_ArrayHierchyAdd(t *testing.T) {

	var a nodes.Node = &ArrayTestStructNode{
		Data: ArrayTestStruct{
			Values: []nodes.Output[float64]{
				nodes.NewValue(1.).Outputs()["Value"].(nodes.Output[float64]),
				nodes.NewValue(2.).Outputs()["Value"].(nodes.Output[float64]),
				nodes.NewValue(3.).Outputs()["Value"].(nodes.Output[float64]),
			},
		},
	}

	var n nodes.Node = &ArrayTestStructNode{
		Data: ArrayTestStruct{
			Values: []nodes.Output[float64]{
				a.Outputs()["Sum"].(nodes.Output[float64]),
				nodes.NewValue(2.).Outputs()["Value"].(nodes.Output[float64]),
				nodes.NewValue(3.).Outputs()["Value"].(nodes.Output[float64]),
			},
		},
	}

	output := n.Outputs()
	assert.Len(t, output, 1)
	// assert.Contains(t, "Sum", output)
	require.Contains(t, output, "Sum")

	sumOutput := output["Sum"].(nodes.Output[float64])
	assert.Equal(t, 11., sumOutput.Value())
	assert.Equal(t, 0, sumOutput.Version())
	assert.Equal(t, "Sum", sumOutput.Name())
	assert.Equal(t, n, sumOutput.Node())
}

func TestStruct_ArrayAdd_AddAndRemoveInputNodes(t *testing.T) {
	var n nodes.Node = &ArrayTestStructNode{
		Data: ArrayTestStruct{
			Values: []nodes.Output[float64]{
				nodes.NewValue(1.).Outputs()["Value"].(nodes.Output[float64]),
				nodes.NewValue(2.).Outputs()["Value"].(nodes.Output[float64]),
				nodes.NewValue(3.).Outputs()["Value"].(nodes.Output[float64]),
			},
		},
	}

	output := n.Outputs()
	require.Contains(t, output, "Sum")
	sumOutput := output["Sum"].(nodes.Output[float64])
	assert.Equal(t, 6., sumOutput.Value())
	assert.Equal(t, 0, sumOutput.Version())

	valueNode := nodes.NewValue(4.)
	inputToAdd := valueNode.Outputs()["Value"].(nodes.Output[float64])

	// Test Adding to the input port
	n.Inputs()["Values"].(nodes.ArrayValueInputPort).Add(inputToAdd)
	assert.Equal(t, 10., sumOutput.Value())
	assert.Equal(t, 1, sumOutput.Version())

	// Change the value thats added to the input port
	valueNode.Set(5)
	assert.Equal(t, 11., sumOutput.Value())
	assert.Equal(t, 2, sumOutput.Version())

	// Test removing from the input port
	n.Inputs()["Values"].(nodes.ArrayValueInputPort).Remove(inputToAdd)
	assert.Equal(t, 6., sumOutput.Value())
	assert.Equal(t, 3, sumOutput.Version())
}
