package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/generator/variable/variabletypes"
	"github.com/EliCDavis/polyform/math"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func importTestInstance(t testing.TB, name string) *graph.Instance {
	t.Helper()
	factory := &refutil.TypeFactory{}
	factory.RegisterBuilder(subgraph.InputNodeTypeKey, func() any { return subgraph.NewInputNode("", "") })
	factory.RegisterBuilder(subgraph.OutputNodeTypeKey, func() any { return subgraph.NewOutputNode("", "") })
	factory.RegisterBuilder("Float64", func() any { return &parameter.Float64{} })
	factory.RegisterBuilder("Sum", func() any { return &nodes.Struct[math.AddNode[float64]]{} })
	nodes.DiscoverPortTypes(factory)
	return graph.New(graph.Config{Name: name, TypeFactory: factory, VariableFactory: variabletypes.New})
}

func floatVariable(t testing.TB, inst *graph.Instance, path string, value float64) {
	t.Helper()
	v := &variable.TypeVariable[float64]{}
	v.SetValue(value)
	require.NoError(t, errOf(inst.NewVariable(path, v)))
}

// readsVariable builds a sub-graph whose "Out" is the variable at path,
// read once for every entry in reads and summed.
func readsVariable(t testing.TB, inst *graph.Instance, id string, reads ...string) *graph.Graph {
	t.Helper()
	def, _ := readsVariableSummed(t, inst, id, reads...)
	return def
}

func readsVariableSummed(t testing.TB, inst *graph.Instance, id string, reads ...string) (*graph.Graph, string) {
	t.Helper()
	require.NoError(t, inst.CreateSubGraph(id, id, ""))
	def, err := inst.SubGraphInstance(id)
	require.NoError(t, err)
	_, outID, err := def.CreateBoundaryNode(subgraph.OutputNodeTypeKey, "float64")
	require.NoError(t, err)
	require.NoError(t, def.SetBoundaryNodeInfo(outID, "Out"))
	_, sumID, err := def.CreateNode("Sum")
	require.NoError(t, err)
	for _, path := range reads {
		_, refID, err := def.CreateNode(path)
		require.NoError(t, err)
		require.NoError(t, def.ConnectNodes(refID, "Value", sumID, "Values"))
	}
	require.NoError(t, def.ConnectNodes(sumID, "Float", outID, subgraph.ValuePortName))
	return def, sumID
}

func place(t testing.TB, scope interface {
	CreateNode(string) (nodes.Node, string, error)
}, id string) (nodes.Node, string) {
	t.Helper()
	node, nodeID, err := scope.CreateNode(subgraph.RuntimeTypePath(id))
	require.NoError(t, err)
	return node, nodeID
}

func encode(t testing.TB, inst *graph.Instance) []byte {
	t.Helper()
	payload, err := inst.EncodeToAppSchema()
	require.NoError(t, err)
	return payload
}

func TestImportedVariableBecomesAnInputDefaultingToItsValue(t *testing.T) {
	src := importTestInstance(t, "Source")
	floatVariable(t, src, "Dims/Size", 7)
	readsVariable(t, src, "scaled", "Dims/Size")

	dest := importTestInstance(t, "Dest")
	floatVariable(t, dest, "Dims/Size", 100)
	result, err := dest.ImportSubGraphDefinitions(encode(t, src), "")
	require.NoError(t, err)

	require.Len(t, result.Imported, 1)
	assert.Equal(t, []graph.VariableInput{{Variable: "Dims/Size", Port: "Size", HasDefault: true}}, result.Imported[0].Inputs)

	placement, placementID := place(t, dest, "scaled")
	assert.Equal(t, 7.0, nodes.GetNodeOutputPort[float64](placement, "Out").Value())

	param, paramID, err := dest.CreateNode("Float64")
	require.NoError(t, err)
	param.(*parameter.Value[float64]).CurrentValue = 3
	require.NoError(t, dest.ConnectNodes(paramID, "Value", placementID, "Size"))
	assert.Equal(t, 3.0, nodes.GetNodeOutputPort[float64](placement, "Out").Value())
}

func TestImportedVariableReadTwiceBecomesOneInput(t *testing.T) {
	src := importTestInstance(t, "Source")
	floatVariable(t, src, "Size", 7)
	readsVariable(t, src, "doubled", "Size", "Size")

	dest := importTestInstance(t, "Dest")
	result, err := dest.ImportSubGraphDefinitions(encode(t, src), "")
	require.NoError(t, err)

	require.Len(t, result.Imported[0].Inputs, 1)
	placement, _ := place(t, dest, "doubled")
	assert.Equal(t, 14.0, nodes.GetNodeOutputPort[float64](placement, "Out").Value())
}

func TestImportedPlacementsShareTheirPlacersInput(t *testing.T) {
	src := importTestInstance(t, "Source")
	floatVariable(t, src, "Size", 7)
	readsVariable(t, src, "inner", "Size")

	require.NoError(t, src.CreateSubGraph("outer", "Outer", ""))
	outer, err := src.SubGraphInstance("outer")
	require.NoError(t, err)
	_, innerID := place(t, outer, "inner")
	_, outID, err := outer.CreateBoundaryNode(subgraph.OutputNodeTypeKey, "float64")
	require.NoError(t, err)
	require.NoError(t, outer.SetBoundaryNodeInfo(outID, "Out"))
	require.NoError(t, outer.ConnectNodes(innerID, "Out", outID, subgraph.ValuePortName))

	dest := importTestInstance(t, "Dest")
	result, err := dest.ImportSubGraphDefinitions(encode(t, src), "")
	require.NoError(t, err)

	byID := map[string]graph.ImportedSubGraph{}
	for _, imported := range result.Imported {
		byID[imported.ID] = imported
	}
	assert.Equal(t, []graph.VariableInput{{Variable: "Size", Port: "Size", HasDefault: true}}, byID["outer"].Inputs)

	placement, placementID := place(t, dest, "outer")
	assert.Equal(t, 7.0, nodes.GetNodeOutputPort[float64](placement, "Out").Value())

	param, paramID, err := dest.CreateNode("Float64")
	require.NoError(t, err)
	param.(*parameter.Value[float64]).CurrentValue = 3
	require.NoError(t, dest.ConnectNodes(paramID, "Value", placementID, "Size"))
	assert.Equal(t, 3.0, nodes.GetNodeOutputPort[float64](placement, "Out").Value())
}

func TestImportedVariableInputAvoidsTakenPortNames(t *testing.T) {
	src := importTestInstance(t, "Source")
	floatVariable(t, src, "Size", 7)
	def := readsVariable(t, src, "scaled", "Size")
	_, inID, err := def.CreateBoundaryNode(subgraph.InputNodeTypeKey, "float64")
	require.NoError(t, err)
	require.NoError(t, def.SetBoundaryNodeInfo(inID, "Size"))

	dest := importTestInstance(t, "Dest")
	result, err := dest.ImportSubGraphDefinitions(encode(t, src), "")
	require.NoError(t, err)

	assert.Equal(t, "Size_2", result.Imported[0].Inputs[0].Port)
	placement, _ := place(t, dest, "scaled")
	assert.Contains(t, inputLabels(placement), "Size_2")
}

func TestImportedSubGraphsAreGroupedUnderTheirSource(t *testing.T) {
	src := importTestInstance(t, "Donut Shop")
	readsVariable(t, src, "glaze")

	dest := importTestInstance(t, "Dest")
	result, err := dest.ImportSubGraphDefinitions(encode(t, src), "")
	require.NoError(t, err)
	assert.Equal(t, "Donut Shop", result.Group)

	result, err = dest.ImportSubGraphDefinitions(encode(t, src), "Pastries")
	require.NoError(t, err)
	assert.Equal(t, "Pastries", result.Group)

	reloaded := importTestInstance(t, "Reloaded")
	require.NoError(t, reloaded.ApplyAppSchema(encode(t, dest)))
	subGraphs := reloaded.Schema().SubGraphs
	assert.Equal(t, "Donut Shop", subGraphs["glaze"].Group)
	assert.Equal(t, "Pastries", subGraphs["glaze_2"].Group)
}
