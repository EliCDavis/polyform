package nodes_test

import (
	"testing"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type GenericTestStructNode[T any] struct {
}

type SimpleAddTestStructNode struct {
	A nodes.Output[float64] `description:"A Desc"`
	B nodes.Output[float64]
}

func (SimpleAddTestStructNode) Description() string {
	return "Adds A and B"
}

func (and SimpleAddTestStructNode) Sum(out *nodes.StructOutput[float64]) {
	out.Set(and.A.Value() + and.B.Value())
}

func (and SimpleAddTestStructNode) SumDescription() string {
	return "The addition of A and B"
}

func TestStruct_SimpleAdd(t *testing.T) {
	var n nodes.Node = &nodes.Struct[SimpleAddTestStructNode]{
		Data: SimpleAddTestStructNode{
			A: nodes.NewValue(1.).Outputs()["Value"].(nodes.Output[float64]),
			B: nodes.NewValue(2.).Outputs()["Value"].(nodes.Output[float64]),
		},
	}

	output := n.Outputs()
	require.Len(t, output, 1)
	require.Contains(t, output, "Sum")

	inputs := n.Inputs()
	require.Len(t, inputs, 2)
	require.Contains(t, inputs, "A")
	require.Contains(t, inputs, "B")
	assert.Equal(t, inputs["A"].(nodes.Describable).Description(), "A Desc")

	sumOutput := output["Sum"].(nodes.Output[float64])
	assert.Equal(t, 3., sumOutput.Value())
	assert.Equal(t, 0, sumOutput.Version())
	assert.Equal(t, "Sum", sumOutput.Name())
	assert.Equal(t, n, sumOutput.Node())
	assert.Equal(t, "The addition of A and B", sumOutput.(nodes.Describable).Description())
}

// ================================================================================================

type ArrayTestStructNode = nodes.Struct[ArrayTestStruct]

type ArrayTestStruct struct {
	Values []nodes.Output[float64]
}

func (and ArrayTestStruct) Sum(out *nodes.StructOutput[float64]) {
	sum := 0.
	for _, v := range and.Values {
		sum += v.Value()
	}
	out.Set(sum)
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

	// ACT ====================================================================
	output := n.Outputs()
	namedNode, named := n.(nodes.Named)
	pathedNode, pathed := n.(nodes.Pathed)

	// ASSERT =================================================================
	assert.Len(t, output, 1)

	assert.True(t, named)
	assert.Equal(t, "Array Test Struct", namedNode.Name())

	assert.True(t, pathed)
	assert.Equal(t, "nodes_test", pathedNode.Path())

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

func TestNodeInfo(t *testing.T) {

	tests := map[string]struct {
		Node        nodes.Node
		Name        string
		Type        string
		Description string
	}{
		"SimpleTestStruct": {
			Node:        &nodes.Struct[SimpleAddTestStructNode]{},
			Name:        "Simple Add Test Struct",
			Type:        "SimpleAddTestStructNode",
			Description: "Adds A and B",
		},
		"ArrayTestStruct": {
			Node:        &ArrayTestStructNode{},
			Name:        "Array Test Struct",
			Type:        "ArrayTestStruct",
			Description: "",
		},
		"Generic[float64]": {
			Node:        &nodes.Struct[GenericTestStructNode[float64]]{},
			Name:        "Generic Test Struct[float64]",
			Type:        "GenericTestStructNode[float64]",
			Description: "",
		},
		"Generic[vector3.Float64]": {
			Node:        &nodes.Struct[GenericTestStructNode[vector3.Float64]]{},
			Name:        "Generic Test Struct[vector3.Vector[float64]]",
			Type:        "GenericTestStructNode[github.com/EliCDavis/vector/vector3.Vector[float64]]",
			Description: "",
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {

			// ACT ============================================================
			namedNode, named := testCase.Node.(nodes.Named)
			pathedNode, pathed := testCase.Node.(nodes.Pathed)
			describeNode, described := testCase.Node.(nodes.Describable)
			typedNode, typed := testCase.Node.(nodes.Typed)

			// ASSERT =========================================================
			assert.True(t, named, "Should be named")
			assert.Equal(t, testCase.Name, namedNode.Name(), "Incorrect Name")

			assert.True(t, pathed, "Should be pathed")
			assert.Equal(t, "nodes_test", pathedNode.Path(), "Incorrect Path")

			assert.True(t, described, "Should have a description")
			assert.Equal(t, testCase.Description, describeNode.Description(), "Incorrect description")

			assert.True(t, typed, "Should be typed")
			assert.Equal(t, testCase.Type, typedNode.Type(), "Incorrect Type")
		})
	}
}
