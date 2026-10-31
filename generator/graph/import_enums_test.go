package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A Window sub-graph with a WindowStyle input port, whose output is the
// style read from the root variable Facade/Style (set to Fixed).
func windowGraphPayload(t *testing.T) []byte {
	t.Helper()
	src := enumTestInstance()
	id := windowStyle(t, src)
	v, _, err := src.NewVariableOfType("Facade/Style", "enum/WindowStyle")
	require.NoError(t, err)
	require.NoError(t, v.(*variable.EnumVariable).SetValue("Fixed"))

	require.NoError(t, src.CreateSubGraph("Window", "Window", ""))
	inside, err := src.SubGraphInstance("Window")
	require.NoError(t, err)
	_, inID, err := inside.CreateBoundaryNode(subgraph.InputNodeTypeKey, "enum/WindowStyle")
	require.NoError(t, err)
	require.NoError(t, inside.SetBoundaryNodeInfo(inID, "Style"))
	_, outID, err := inside.CreateBoundaryNode(subgraph.OutputNodeTypeKey, "enum/WindowStyle")
	require.NoError(t, err)
	require.NoError(t, inside.SetBoundaryNodeInfo(outID, "From Variable"))
	_, refID, err := inside.CreateNode("Facade/Style")
	require.NoError(t, err)
	require.NoError(t, inside.ConnectNodes(refID, "Value", outID, "Value"))
	_, _, err = inside.CreateNode(graph.EnumParameterTypePath(id))
	require.NoError(t, err)

	payload, err := src.EncodeToAppSchema()
	require.NoError(t, err)
	return payload
}

func TestImportBringsTheEnumsASubGraphUses(t *testing.T) {
	dest := enumTestInstance()
	result, err := dest.ImportSubGraphDefinitions(windowGraphPayload(t), "")
	require.NoError(t, err)

	require.Len(t, result.Enums, 1)
	assert.Equal(t, graph.ImportedEnum{ID: "WindowStyle", Name: "Window Style"}, result.Enums[0])
	assert.Equal(t, "Window Style", dest.Schema().Enums["WindowStyle"].Name)

	require.Len(t, result.Imported, 1)
	imported := result.Imported[0]
	require.Len(t, imported.Inputs, 1)
	assert.Equal(t, "Facade/Style", imported.Inputs[0].Variable)
	assert.True(t, imported.Inputs[0].HasDefault)
	assert.Equal(t, "enum/WindowStyle", imported.NodeType.Inputs["Style"].Type)

	placement, placementID, err := dest.CreateNode(subgraph.RuntimeTypePath("Window"))
	require.NoError(t, err)
	_ = placementID
	var outPort string
	for name := range placement.Outputs() {
		outPort = name
	}
	out := nodes.GetNodeOutputPort[enum.Choice](placement, outPort)
	assert.Equal(t, "Fixed", out.Value().Name, "the lifted variable's value is the input's default")
}

func TestImportReusesAMatchingEnum(t *testing.T) {
	dest := enumTestInstance()
	existing, err := dest.CreateEnum("window style", "", []enum.Member{{Name: "fixed"}, {Name: "casement"}, {Name: "sash"}})
	require.NoError(t, err)

	result, err := dest.ImportSubGraphDefinitions(windowGraphPayload(t), "")
	require.NoError(t, err)
	require.Len(t, result.Enums, 1)
	assert.Equal(t, graph.ImportedEnum{ID: existing, OriginalID: "WindowStyle", Name: "window style", Reused: true}, result.Enums[0])
	assert.Len(t, dest.Schema().Enums, 1)
	assert.Equal(t, "enum/"+existing, result.Imported[0].NodeType.Inputs["Style"].Type)
}

func TestImportKeepsAnEnumWithTheSameNameButOtherMembersApart(t *testing.T) {
	dest := enumTestInstance()
	_, err := dest.CreateEnum("Window Style", "", []enum.Member{{Name: "Round"}, {Name: "Square"}})
	require.NoError(t, err)

	result, err := dest.ImportSubGraphDefinitions(windowGraphPayload(t), "")
	require.NoError(t, err)
	require.Len(t, result.Enums, 1)
	assert.Equal(t, graph.ImportedEnum{ID: "WindowStyle_2", OriginalID: "WindowStyle", Name: "Window Style_2"}, result.Enums[0])
	assert.Equal(t, "enum/WindowStyle_2", result.Imported[0].NodeType.Inputs["Style"].Type)
	table, ok := dest.EnumTable("enum/WindowStyle_2")
	require.True(t, ok)
	assert.Equal(t, []string{"Casement", "Sash", "Fixed"}, table.Names())
}
