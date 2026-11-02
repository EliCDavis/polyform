package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func passThroughSubgraph(t *testing.T, inst *graph.Instance, id string) (def *graph.Graph, inputID, outputID string) {
	t.Helper()
	require.NoError(t, inst.CreateSubGraph(id, id, ""))
	def, err := inst.SubGraphInstance(id)
	require.NoError(t, err)
	_, inputID, err = def.CreateBoundaryNode(subgraph.InputNodeTypeKey, "float64")
	require.NoError(t, err)
	require.NoError(t, def.SetBoundaryNodeInfo(inputID, "Size"))
	_, outputID, err = def.CreateBoundaryNode(subgraph.OutputNodeTypeKey, "float64")
	require.NoError(t, err)
	require.NoError(t, def.SetBoundaryNodeInfo(outputID, "Out"))
	require.NoError(t, def.ConnectNodes(inputID, subgraph.ValuePortName, outputID, subgraph.ValuePortName))
	return def, inputID, outputID
}

func placementOut(t *testing.T, inst *graph.Instance, id string) float64 {
	t.Helper()
	return nodes.GetNodeOutputPort[float64](inst.Node(id), "Out").Value()
}

func scopedLiterals(t *testing.T, def *graph.Graph, id string) map[string]string {
	t.Helper()
	shown := map[string]string{}
	for port, raw := range def.Schema().Nodes[id].Literals {
		shown[port] = string(raw)
	}
	return shown
}

func TestABoundaryHoldsItsDefaultAsALiteral(t *testing.T) {
	inst := importTestInstance(t, "literals")
	def, inputID, _ := passThroughSubgraph(t, inst, "pass")

	assert.Equal(t, map[string]string{"default": "0"}, scopedLiterals(t, def, inputID))
	require.NoError(t, def.SetLiteral(inputID, subgraph.DefaultPortName, []byte("7")))

	_, placementID, err := inst.CreateNode(subgraph.RuntimeTypePath("pass"))
	require.NoError(t, err)
	assert.Equal(t, 7., placementOut(t, inst, placementID))
	assert.Equal(t, map[string]string{"Size": "7"}, literalsOf(inst, placementID), "the placement starts at the boundary's default")
}

func TestAPlacementLiteralIsItsOwn(t *testing.T) {
	inst := importTestInstance(t, "literals")
	def, inputID, _ := passThroughSubgraph(t, inst, "pass")
	require.NoError(t, def.SetLiteral(inputID, subgraph.DefaultPortName, []byte("7")))
	_, first, err := inst.CreateNode(subgraph.RuntimeTypePath("pass"))
	require.NoError(t, err)

	require.NoError(t, inst.SetLiteral(first, "Size", []byte("9")))
	assert.Equal(t, 9., placementOut(t, inst, first))
	assert.Equal(t, "7", scopedLiterals(t, def, inputID)["default"], "the definition's default is untouched")

	require.NoError(t, def.SetLiteral(inputID, subgraph.DefaultPortName, []byte("11")))
	_, second, err := inst.CreateNode(subgraph.RuntimeTypePath("pass"))
	require.NoError(t, err)
	assert.Equal(t, 9., placementOut(t, inst, first), "an existing placement keeps its value")
	assert.Equal(t, 11., placementOut(t, inst, second), "a new placement starts at the new default")
}

func TestAWireIntoThePlacementBeatsItsLiteral(t *testing.T) {
	inst := importTestInstance(t, "literals")
	def, inputID, _ := passThroughSubgraph(t, inst, "pass")
	require.NoError(t, def.SetLiteral(inputID, subgraph.DefaultPortName, []byte("7")))
	_, placementID, err := inst.CreateNode(subgraph.RuntimeTypePath("pass"))
	require.NoError(t, err)
	_, paramID, err := inst.CreateNode("Float64")
	require.NoError(t, err)
	require.NoError(t, errOf(inst.UpdateParameter(paramID, []byte("3"))))

	require.NoError(t, inst.ConnectNodes(paramID, "Value", placementID, "Size"))
	assert.Equal(t, 3., placementOut(t, inst, placementID))

	_, err = inst.DeleteNodeInputConnection(placementID, "Size")
	require.NoError(t, err)
	assert.Equal(t, 7., placementOut(t, inst, placementID))
}

func TestABoundaryAddedLaterReachesExistingPlacements(t *testing.T) {
	inst := importTestInstance(t, "literals")
	def, _, _ := passThroughSubgraph(t, inst, "pass")
	_, placementID, err := inst.CreateNode(subgraph.RuntimeTypePath("pass"))
	require.NoError(t, err)

	_, extraID, err := def.CreateBoundaryNode(subgraph.InputNodeTypeKey, "float64")
	require.NoError(t, err)
	require.NoError(t, def.SetLiteral(extraID, subgraph.DefaultPortName, []byte("2")))
	assert.NotContains(t, literalsOf(inst, placementID), "Extra", "an unnamed boundary is not a port yet")

	require.NoError(t, def.SetBoundaryNodeInfo(extraID, "Extra"))
	assert.Equal(t, "2", literalsOf(inst, placementID)["Extra"])
	inputByLabel(t, inst.Node(placementID), "Extra")
}

func TestADeletedBoundaryTakesThePlacementLiteralWithIt(t *testing.T) {
	inst := importTestInstance(t, "literals")
	def, inputID, _ := passThroughSubgraph(t, inst, "pass")
	_, placementID, err := inst.CreateNode(subgraph.RuntimeTypePath("pass"))
	require.NoError(t, err)
	require.NoError(t, inst.SetLiteral(placementID, "Size", []byte("9")))

	_, err = def.DeleteNodeById(inputID)
	require.NoError(t, err)

	assert.Empty(t, literalsOf(inst, placementID))
	saved, err := inst.EncodeToAppSchema()
	require.NoError(t, err)
	assert.NotContains(t, string(saved), `"Size"`)
}

func TestAReplacedPlacementStartsAtTheOtherDefinitionsDefaults(t *testing.T) {
	inst := importTestInstance(t, "literals")
	_, _, _ = passThroughSubgraph(t, inst, "first")
	other, otherInput, _ := passThroughSubgraph(t, inst, "second")
	require.NoError(t, other.SetLiteral(otherInput, subgraph.DefaultPortName, []byte("5")))
	_, placementID, err := inst.CreateNode(subgraph.RuntimeTypePath("first"))
	require.NoError(t, err)
	require.NoError(t, inst.SetLiteral(placementID, "Size", []byte("9")))

	_, err = inst.ReplaceSubgraphInstance(placementID, "second")
	require.NoError(t, err)

	assert.Equal(t, map[string]string{"Size": "5"}, literalsOf(inst, placementID))
	assert.Equal(t, 5., placementOut(t, inst, placementID))
}

func TestPlacementLiteralsSurviveSaveAndLoad(t *testing.T) {
	inst := importTestInstance(t, "literals")
	def, inputID, _ := passThroughSubgraph(t, inst, "pass")
	require.NoError(t, def.SetLiteral(inputID, subgraph.DefaultPortName, []byte("7")))
	_, placementID, err := inst.CreateNode(subgraph.RuntimeTypePath("pass"))
	require.NoError(t, err)
	require.NoError(t, inst.SetLiteral(placementID, "Size", []byte("9")))

	saved, err := inst.EncodeToAppSchema()
	require.NoError(t, err)
	loaded := importTestInstance(t, "loaded")
	require.NoError(t, loaded.ApplyAppSchema(saved))

	assert.Equal(t, 9., placementOut(t, loaded, placementID))
	loadedDef, err := loaded.SubGraphInstance("pass")
	require.NoError(t, err)
	assert.Equal(t, "7", scopedLiterals(t, loadedDef, inputID)["default"])
}

func TestAForEachPlacementRunsOnceOnItsLiteral(t *testing.T) {
	inst := importTestInstance(t, "literals")
	def, inputID, _ := passThroughSubgraph(t, inst, "pass")
	require.NoError(t, def.SetLiteral(inputID, subgraph.DefaultPortName, []byte("7")))
	_, placementID, err := inst.CreateNode(subgraph.RuntimeTypePath("pass"))
	require.NoError(t, err)

	_, err = inst.SetForEach(placementID, true)
	require.NoError(t, err)

	assert.Equal(t, []float64{7}, nodes.GetNodeOutputPort[[]float64](inst.Node(placementID), "Out").Value())
}

func TestAnImportedVariableBecomesABoundaryLiteral(t *testing.T) {
	src := importTestInstance(t, "Source")
	floatVariable(t, src, "Size", 7)
	readsVariable(t, src, "scaled", "Size")

	dest := importTestInstance(t, "Dest")
	result, err := dest.ImportSubGraphDefinitions(encode(t, src), "")
	require.NoError(t, err)
	require.Len(t, result.Imported, 1)
	assert.Equal(t, []graph.VariableInput{{Variable: "Size", Port: "Size", HasDefault: true}}, result.Imported[0].Inputs)

	def, err := dest.SubGraphInstance("scaled")
	require.NoError(t, err)
	var boundaryID string
	for id, node := range def.Schema().Nodes {
		if node.SubGraphInputBoundary != nil {
			boundaryID = id
		}
	}
	require.NotEmpty(t, boundaryID)
	assert.Equal(t, "7", scopedLiterals(t, def, boundaryID)["default"])
	assert.Empty(t, def.Schema().Nodes[boundaryID].AssignedInput, "no constant node is wired in")
	assert.Len(t, def.Schema().Nodes, 3, "boundary, sum and output only")

	placement, _ := place(t, dest, "scaled")
	assert.Equal(t, 7.0, nodes.GetNodeOutputPort[float64](placement, "Out").Value())
}
