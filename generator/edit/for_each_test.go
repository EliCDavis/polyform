package edit_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestForEachIsSetThroughTheEditServer(t *testing.T) {
	handler, inst := subGraphTestServer(t)
	const subGraphID = "adder"
	requireOK(t, handler, httpStep{method: http.MethodPost, url: "/subgraph/definition/" + subGraphID, body: `{"name":"Adder"}`})
	inputID := createScopedNode(t, handler, subGraphID, subgraph.InputNodeTypeKey, "float64")
	outputID := createScopedNode(t, handler, subGraphID, subgraph.OutputNodeTypeKey, "float64")
	setBoundaryInfo(t, handler, inputID, subGraphID, "A")
	setBoundaryInfo(t, handler, outputID, subGraphID, "Result")
	placement := createRootNode(t, handler, subgraph.RuntimeTypePath(subGraphID))

	requireOK(t, handler, httpStep{method: http.MethodPost, url: "/subgraph/instance/" + placement + "/for-each", body: `{"each":true,"scope":null}`})

	var schema struct {
		Nodes map[string]struct {
			ForEach bool                             `json:"forEach"`
			Output  map[string]struct{ Type string } `json:"output"`
		} `json:"nodes"`
	}
	require.NoError(t, json.Unmarshal(requireOK(t, handler, httpStep{method: http.MethodGet, url: "/schema"}), &schema))
	assert.True(t, schema.Nodes[placement].ForEach)
	assert.Equal(t, "[]float64", schema.Nodes[placement].Output["Result"].Type)
	assert.True(t, inst.Node(placement).(*graph.SubgraphInstanceNode).ForEach())

	requireOK(t, handler, httpStep{method: http.MethodPost, url: "/subgraph/instance/" + placement + "/for-each", body: `{"each":false,"scope":null}`})
	assert.False(t, inst.Node(placement).(*graph.SubgraphInstanceNode).ForEach())

	code, _ := serveStep(t, handler, httpStep{method: http.MethodPost, url: "/subgraph/instance/" + inputID + "/for-each", body: `{"each":true,"scope":"subgraph/adder"}`})
	assert.GreaterOrEqual(t, code, http.StatusBadRequest, "a boundary is not a placement")
}

func TestDeletingANodeReportsTheEdgesItTookWithIt(t *testing.T) {
	handler, _ := subGraphTestServer(t)
	const subGraphID = "adder"
	requireOK(t, handler, httpStep{method: http.MethodPost, url: "/subgraph/definition/" + subGraphID, body: `{"name":"Adder"}`})
	inputID := createScopedNode(t, handler, subGraphID, subgraph.InputNodeTypeKey, "float64")
	setBoundaryInfo(t, handler, inputID, subGraphID, "A")
	placement := createRootNode(t, handler, subgraph.RuntimeTypePath(subGraphID))
	param := createRootNode(t, handler, "Float64")
	connectNodes(t, handler, "/node/connection", param, "Value", placement, "A")

	body := requireOK(t, handler, httpStep{method: http.MethodDelete, url: "/graph/subgraph/" + subGraphID + "/node", body: `{"nodeID":"` + inputID + `"}`})

	var response struct {
		Dropped []graph.DroppedEdge `json:"dropped"`
	}
	require.NoError(t, json.Unmarshal(body, &response))
	require.Len(t, response.Dropped, 1, "%s", body)
	assert.Equal(t, param, response.Dropped[0].From)
	assert.Equal(t, placement, response.Dropped[0].To)
}
