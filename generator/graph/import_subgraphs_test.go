package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildSimpleAdderPayload(t *testing.T) []byte {
	t.Helper()
	src := testInstanceWithSubGraphTypesExtended(t)
	require.NoError(t, src.CreateSubGraph("adder", "Adder", "adds"))

	child, err := src.SubGraphInstance("adder")
	require.NoError(t, err)

	_, inputAID, err := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, "float64")
	require.NoError(t, err)
	_, inputBID, err := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, "float64")
	require.NoError(t, err)
	_, outputID, err := child.CreateBoundaryNode(subgraph.OutputNodeTypeKey, "float64")
	require.NoError(t, err)
	_, sumID, err := child.CreateNode("Sum")
	require.NoError(t, err)

	require.NoError(t, child.SetBoundaryNodeInfo(inputAID, "A"))
	require.NoError(t, child.SetBoundaryNodeInfo(inputBID, "B"))
	require.NoError(t, child.SetBoundaryNodeInfo(outputID, "Result"))
	child.ConnectNodes(inputAID, subgraph.ValuePortName, sumID, "Values")
	child.ConnectNodes(inputBID, subgraph.ValuePortName, sumID, "Values")
	child.ConnectNodes(sumID, "Float", outputID, subgraph.ValuePortName)

	// Root noise that must not be imported.
	_, _, err = src.CreateNode("Float64")
	require.NoError(t, err)

	payload, err := src.EncodeToAppSchema()
	require.NoError(t, err)
	return payload
}

func TestImportSubGraphDefinitions_KeepsIDsWhenFree(t *testing.T) {
	payload := buildSimpleAdderPayload(t)
	dest := testInstanceWithSubGraphTypesExtended(t)

	result, err := dest.ImportSubGraphDefinitions(payload)
	require.NoError(t, err)
	require.Len(t, result.Imported, 1)
	assert.Equal(t, "adder", result.Imported[0].ID)
	assert.Empty(t, result.Imported[0].OriginalID)
	assert.Equal(t, "Adder", result.Imported[0].Name)
	assert.Equal(t, "subgraph/adder", result.Imported[0].NodeType.Type)

	_, err = dest.SubGraphInstance("adder")
	require.NoError(t, err)
}

func TestImportSubGraphDefinitions_EmptySubGraphs(t *testing.T) {
	src := testInstanceWithSubGraphTypesExtended(t)
	_, _, err := src.CreateNode("Float64")
	require.NoError(t, err)
	payload, err := src.EncodeToAppSchema()
	require.NoError(t, err)

	dest := testInstanceWithSubGraphTypesExtended(t)
	result, err := dest.ImportSubGraphDefinitions(payload)
	require.NoError(t, err)
	assert.Empty(t, result.Imported)
}

func TestImportSubGraphDefinitions_ConflictRenames(t *testing.T) {
	payload := buildSimpleAdderPayload(t)

	dest := testInstanceWithSubGraphTypesExtended(t)
	require.NoError(t, dest.CreateSubGraph("adder", "Existing", ""))

	result, err := dest.ImportSubGraphDefinitions(payload)
	require.NoError(t, err)
	require.Len(t, result.Imported, 1)
	assert.Equal(t, "adder_2", result.Imported[0].ID)
	assert.Equal(t, "adder", result.Imported[0].OriginalID)
	assert.Equal(t, "subgraph/adder_2", result.Imported[0].NodeType.Type)

	_, err = dest.SubGraphInstance("adder")
	require.NoError(t, err)
	_, err = dest.SubGraphInstance("adder_2")
	require.NoError(t, err)
}

func TestImportSubGraphDefinitions_DoesNotTouchRoot(t *testing.T) {
	payload := buildSimpleAdderPayload(t)

	dest := testInstanceWithSubGraphTypesExtended(t)
	param, paramID, err := dest.CreateNode("Float64")
	require.NoError(t, err)
	param.(*parameter.Float64).CurrentValue = 42

	result, err := dest.ImportSubGraphDefinitions(payload)
	require.NoError(t, err)
	require.Len(t, result.Imported, 1)

	assert.True(t, dest.HasNodeWithId(paramID))
	assert.Equal(t, 42.0, param.(*parameter.Float64).CurrentValue)
	assert.Len(t, dest.Schema().Nodes, 1)
}

func TestImportSubGraphDefinitions_NestedTypeRemap(t *testing.T) {
	src := testInstanceWithSubGraphTypesExtended(t)
	require.NoError(t, src.CreateSubGraph("inner", "Inner", ""))
	require.NoError(t, src.CreateSubGraph("outer", "Outer", ""))

	inner, err := src.SubGraphInstance("inner")
	require.NoError(t, err)
	_, inID, err := inner.CreateBoundaryNode(subgraph.InputNodeTypeKey, "float64")
	require.NoError(t, err)
	_, outID, err := inner.CreateBoundaryNode(subgraph.OutputNodeTypeKey, "float64")
	require.NoError(t, err)
	require.NoError(t, inner.SetBoundaryNodeInfo(inID, "In"))
	require.NoError(t, inner.SetBoundaryNodeInfo(outID, "Out"))
	inner.ConnectNodes(inID, subgraph.ValuePortName, outID, subgraph.ValuePortName)

	outer, err := src.SubGraphInstance("outer")
	require.NoError(t, err)
	_, _, err = outer.CreateNode(subgraph.RuntimeTypePath("inner"))
	require.NoError(t, err)

	payload, err := src.EncodeToAppSchema()
	require.NoError(t, err)

	dest := testInstanceWithSubGraphTypesExtended(t)
	require.NoError(t, dest.CreateSubGraph("inner", "Taken", ""))

	result, err := dest.ImportSubGraphDefinitions(payload)
	require.NoError(t, err)
	require.Len(t, result.Imported, 2)

	byOriginal := map[string]graph.ImportedSubGraph{}
	for _, entry := range result.Imported {
		key := entry.OriginalID
		if key == "" {
			key = entry.ID
		}
		byOriginal[key] = entry
	}
	assert.Equal(t, "inner_2", byOriginal["inner"].ID)
	assert.Equal(t, "outer", byOriginal["outer"].ID)

	destOuter, err := dest.SubGraphInstance("outer")
	require.NoError(t, err)
	foundNested := false
	for _, node := range destOuter.Schema().Nodes {
		if node.Type == subgraph.RuntimeTypePath("inner_2") {
			foundNested = true
			break
		}
		assert.NotEqual(t, subgraph.RuntimeTypePath("inner"), node.Type)
	}
	assert.True(t, foundNested)
}

func TestImportSubGraphDefinitions_EvaluatesAfterImport(t *testing.T) {
	payload := buildSimpleAdderPayload(t)
	dest := testInstanceWithSubGraphTypesExtended(t)

	result, err := dest.ImportSubGraphDefinitions(payload)
	require.NoError(t, err)
	require.Len(t, result.Imported, 1)

	paramA, paramAID, err := dest.CreateNode("Float64")
	require.NoError(t, err)
	paramA.(*parameter.Float64).CurrentValue = 5
	paramB, paramBID, err := dest.CreateNode("Float64")
	require.NoError(t, err)
	paramB.(*parameter.Float64).CurrentValue = 3

	runtime, runtimeID, err := dest.CreateNode(subgraph.RuntimeTypePath("adder"))
	require.NoError(t, err)
	dest.ConnectNodes(paramAID, "Value", runtimeID, "A")
	dest.ConnectNodes(paramBID, "Value", runtimeID, "B")

	assert.Equal(t, 8.0, nodes.GetNodeOutputPort[float64](runtime, "Result").Value())
}
