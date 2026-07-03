package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/stretchr/testify/assert"
)

func TestCollectAllPortTypes_Empty(t *testing.T) {
	assert.Empty(t, graph.CollectAllPortTypes(nil))
}

func TestCollectAllPortTypes_DedupesAndSorts(t *testing.T) {
	nodeTypes := []schema.NodeType{
		{
			Inputs: map[string]schema.NodeTypeInput{
				"B": {Type: "float64"},
				"A": {Type: "int"},
			},
			Outputs: map[string]schema.NodeTypeOutput{
				"Out": {Type: "float64"},
			},
		},
		{
			Inputs: map[string]schema.NodeTypeInput{
				"In": {Type: "string"},
			},
		},
	}

	types := graph.CollectAllPortTypes(nodeTypes)
	assert.Equal(t, []string{"float64", "int", "string"}, types)
}

func TestCollectAllPortTypes_SkipsAnyAndEmpty(t *testing.T) {
	nodeTypes := []schema.NodeType{
		{
			Inputs: map[string]schema.NodeTypeInput{
				"A": {Type: "any"},
				"B": {Type: ""},
				"C": {Type: "bool"},
			},
			Outputs: map[string]schema.NodeTypeOutput{
				"X": {Type: "any"},
			},
		},
	}

	types := graph.CollectAllPortTypes(nodeTypes)
	assert.Equal(t, []string{"bool"}, types)
}

func TestCollectAllPortTypes_MultipleOutputs(t *testing.T) {
	nodeTypes := []schema.NodeType{
		{
			Outputs: map[string]schema.NodeTypeOutput{
				"Z": {Type: "github.com/EliCDavis/polyform/generator/manifest.Manifest"},
				"Y": {Type: "image.Image"},
			},
		},
	}

	types := graph.CollectAllPortTypes(nodeTypes)
	assert.Contains(t, types, "github.com/EliCDavis/polyform/generator/manifest.Manifest")
	assert.Contains(t, types, "image.Image")
	assert.Len(t, types, 2)
}
