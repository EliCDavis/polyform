package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetParameterMissingScopeNamesWhereTheNodeLives(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_subgraph", map[string]any{"id": "body", "name": "Body"}, &polyformmcp.CreateSubgraphOutput{})
	var created polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": floatParamType, "scope": "body"}, &created)

	msg := callToolExpectingError(t, session, "set_parameter", map[string]any{
		"nodeId": created.NodeId, "value": "2",
	})
	assert.Contains(t, msg, `subgraph "body"`)
	assert.Contains(t, msg, "scope")
}

func TestDeleteNodeMissingScopeNamesWhereTheNodeLives(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_subgraph", map[string]any{"id": "body", "name": "Body"}, &polyformmcp.CreateSubgraphOutput{})
	var created polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": floatParamType, "scope": "body"}, &created)

	msg := callToolExpectingError(t, session, "delete_node", map[string]any{"nodeId": created.NodeId})
	assert.Contains(t, msg, `subgraph "body"`)
}

func TestConnectNodesMissingScopeNamesWhereTheNodeLives(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_subgraph", map[string]any{"id": "body", "name": "Body"}, &polyformmcp.CreateSubgraphOutput{})
	var lit, sub polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": floatParamType, "scope": "body"}, &lit)
	callTool(t, session, "create_node", map[string]any{"type": subtractNodeType, "scope": "body"}, &sub)

	msg := callToolExpectingError(t, session, "connect_nodes", map[string]any{
		"outNodeId": lit.NodeId, "outPort": "Value",
		"inNodeId": sub.NodeId, "inPort": "A",
	})
	assert.Contains(t, msg, `subgraph "body"`)
}

func TestMissingNodeThatExistsNowhereGetsNoHint(t *testing.T) {
	session := testSession(t)
	msg := callToolExpectingError(t, session, "delete_node", map[string]any{"nodeId": "Node-999"})
	require.Contains(t, msg, "Node-999")
	assert.NotContains(t, msg, "scope")
}
