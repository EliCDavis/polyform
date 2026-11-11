package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func defineShape(t *testing.T, session *mcpsdk.ClientSession) polyformmcp.EnumSummary {
	t.Helper()
	var shape polyformmcp.EnumSummary
	callTool(t, session, "define_enum", map[string]any{
		"name":    "Shape",
		"members": []map[string]any{{"name": "Round"}, {"name": "Square"}},
	}, &shape)
	return shape
}

func TestDefinedEnumWorksAsAVariableType(t *testing.T) {
	session := testSession(t)
	shape := defineShape(t, session)
	assert.Equal(t, []string{"Round", "Square"}, shape.Members)

	callTool(t, session, "create_variable", map[string]any{"path": "Style", "type": shape.Type, "value": `"Square"`}, &polyformmcp.CreateVariableOutput{})

	var listed polyformmcp.ListVariablesOutput
	callTool(t, session, "list_variables", map[string]any{}, &listed)
	require.Len(t, listed.Variables, 1)
	assert.Equal(t, "Square", listed.Variables[0].Value)
	assert.Equal(t, []string{"Round", "Square"}, listed.Variables[0].Members)
}

func TestEnumVariableRefusesAnUnknownMember(t *testing.T) {
	session := testSession(t)
	shape := defineShape(t, session)
	callTool(t, session, "create_variable", map[string]any{"path": "Style", "type": shape.Type}, &polyformmcp.CreateVariableOutput{})

	msg := callToolExpectingError(t, session, "update_variable", map[string]any{"path": "Style", "value": `"Triangle"`})
	assert.Contains(t, msg, "Triangle")
}

func TestEditEnumCarriesHoldersAcrossARename(t *testing.T) {
	session := testSession(t)
	shape := defineShape(t, session)
	callTool(t, session, "create_variable", map[string]any{"path": "Style", "type": shape.Type, "value": `"Square"`}, &polyformmcp.CreateVariableOutput{})

	var edited polyformmcp.EnumSummary
	callTool(t, session, "edit_enum", map[string]any{
		"id":      shape.Id,
		"members": []map[string]any{{"name": "Round"}, {"name": "Boxy"}},
		"renamed": map[string]string{"Square": "Boxy"},
	}, &edited)
	assert.Equal(t, "Shape", edited.Name)
	assert.Equal(t, []string{"Round", "Boxy"}, edited.Members)

	var listed polyformmcp.ListVariablesOutput
	callTool(t, session, "list_variables", map[string]any{}, &listed)
	assert.Equal(t, "Boxy", listed.Variables[0].Value)
}

func TestDeleteEnumIsRefusedWhileAVariableUsesIt(t *testing.T) {
	session := testSession(t)
	shape := defineShape(t, session)
	callTool(t, session, "create_variable", map[string]any{"path": "Style", "type": shape.Type}, &polyformmcp.CreateVariableOutput{})

	callToolExpectingError(t, session, "delete_enum", map[string]any{"id": shape.Id})

	callTool(t, session, "delete_variable", map[string]any{"path": "Style"}, &polyformmcp.DeleteVariableOutput{})
	var deleted polyformmcp.DeleteEnumOutput
	callTool(t, session, "delete_enum", map[string]any{"id": shape.Id}, &deleted)
	assert.True(t, deleted.Deleted)
}

func TestEnumParameterNodeFeedsASubgraphBoundary(t *testing.T) {
	session := testSession(t)
	shape := defineShape(t, session)

	callTool(t, session, "create_subgraph", map[string]any{
		"id": "pick", "name": "Pick",
		"inputs":  []map[string]any{{"name": "Style", "type": shape.Type}},
		"outputs": []map[string]any{{"name": "Chosen", "type": shape.Type, "nodeId": "Style", "port": "Value"}},
	}, &polyformmcp.CreateSubgraphOutput{})

	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": "pick",
		"inputs":     map[string]any{"Style": map[string]any{"value": `"Square"`}},
	}, &placed)

	var read polyformmcp.EvaluateNodeOutput
	callTool(t, session, "evaluate_node", map[string]any{"nodeId": placed.NodeId, "port": "Chosen"}, &read)
	assert.Equal(t, "Square", read.Value)
}

func TestListEnumsIncludesTheOnesBuiltIntoNodeTypes(t *testing.T) {
	session := testSession(t)
	shape := defineShape(t, session)

	var listed polyformmcp.ListEnumsOutput
	callTool(t, session, "list_enums", map[string]any{}, &listed)
	require.NotEmpty(t, listed.Enums)
	assert.Equal(t, shape.Id, listed.Enums[0].Id)
	assert.Greater(t, len(listed.Enums), 1)
}

func TestGetNodeTypesListsDefaultsAndEnumMembers(t *testing.T) {
	session := testSession(t)

	var details polyformmcp.GetNodeTypesOutput
	callTool(t, session, "get_node_types", map[string]any{"types": []string{materialNodeType}}, &details)
	require.Len(t, details.Results, 1)

	ports := map[string]polyformmcp.PortSummary{}
	for _, port := range details.Results[0].Inputs {
		ports[port.Name] = port
	}
	assert.True(t, ports["Alpha Cutoff"].Literal)
	assert.Equal(t, 0.5, ports["Alpha Cutoff"].Default)
	assert.Contains(t, ports["Alpha Mode"].Members, "MASK")
}

func TestGetNodeTypesMarksLiftedInputs(t *testing.T) {
	session := testSession(t)

	lineType := "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/repeat.LineNode]"
	translateType := "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/meshops.TranslateAttribute3DNode]"

	var details polyformmcp.GetNodeTypesOutput
	callTool(t, session, "get_node_types", map[string]any{"types": []string{lineType, translateType}}, &details)
	require.Len(t, details.Results, 2)

	ports := map[string]map[string]polyformmcp.PortSummary{}
	for _, result := range details.Results {
		ports[result.Type] = map[string]polyformmcp.PortSummary{}
		for _, port := range result.Inputs {
			ports[result.Type][port.Name] = port
		}
	}
	assert.True(t, ports[lineType]["Start"].Lifts)
	assert.True(t, ports[lineType]["Samples"].Lifts)
	assert.False(t, ports[translateType]["Amount"].Lifts)
}
