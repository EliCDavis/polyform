package generator_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

type TestNode = nodes.StructNode[float64, TestNodeData]

type TestNodeData struct {
	A nodes.NodeOutput[float64]
	B nodes.NodeOutput[int]
}

func (bn TestNodeData) Process() (float64, error) {
	return 0, nil
}

func TestBuildNodeTypeSchema(t *testing.T) {
	schema := generator.BuildNodeTypeSchema(&TestNode{})

	assert.Equal(t, "TestNodeData", schema.DisplayName)
	assert.Equal(t, "generator_test", schema.Path)

	assert.Len(t, schema.Inputs, 2)
	assert.Equal(t, "float64", schema.Inputs["A"].Type)
	assert.Equal(t, "int", schema.Inputs["B"].Type)

	assert.Len(t, schema.Outputs, 1)
	assert.Equal(t, "float64", schema.Outputs[0].Type)
	assert.Equal(t, "Out", schema.Outputs[0].Name)
}
