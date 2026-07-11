package edit_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EliCDavis/polyform/generator/edit"
	"github.com/EliCDavis/polyform/generator/graph"
	_ "github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/math"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func subGraphTestServer(t *testing.T) (http.Handler, *graph.Instance) {
	t.Helper()

	factory := &refutil.TypeFactory{}
	factory.RegisterBuilder(subgraph.InputNodeTypeKey, func() any {
		return subgraph.NewInputNode("", "")
	})
	factory.RegisterBuilder(subgraph.OutputNodeTypeKey, func() any {
		return subgraph.NewOutputNode("", "")
	})
	factory.RegisterBuilder("Float64", func() any {
		return &parameter.Float64{CurrentValue: 1}
	})
	factory.RegisterBuilder("Sum", func() any {
		return &nodes.Struct[math.AddNode[float64]]{}
	})

	inst := graph.New(graph.Config{TypeFactory: factory})
	server := edit.Server{Graph: inst}
	handler, err := server.Handler("./")
	require.NoError(t, err)
	return handler, inst
}

type httpStep struct {
	method string
	url    string
	body   string
}

func serveStep(t *testing.T, handler http.Handler, step httpStep) (int, []byte) {
	t.Helper()
	req := httptest.NewRequest(step.method, step.url, strings.NewReader(step.body))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr.Code, rr.Body.Bytes()
}

func requireOK(t *testing.T, handler http.Handler, step httpStep) []byte {
	t.Helper()
	code, body := serveStep(t, handler, step)
	require.Equal(t, http.StatusOK, code, "request %s %s failed: %s", step.method, step.url, string(body))
	return body
}

func createScopedNode(t *testing.T, handler http.Handler, subGraphID, nodeType, portType string) string {
	t.Helper()
	body := fmt.Sprintf(`{"nodeType":"%s"`, nodeType)
	if portType != "" {
		body += fmt.Sprintf(`,"portType":"%s"`, portType)
	}
	body += "}"
	respBody := requireOK(t, handler, httpStep{
		method: http.MethodPost,
		url:    fmt.Sprintf("/graph/subgraph/%s/node", subGraphID),
		body:   body,
	})
	var resp struct {
		NodeID string `json:"nodeID"`
	}
	require.NoError(t, json.Unmarshal(respBody, &resp))
	require.NotEmpty(t, resp.NodeID)
	return resp.NodeID
}

func createRootNode(t *testing.T, handler http.Handler, nodeType string) string {
	t.Helper()
	body := requireOK(t, handler, httpStep{
		method: http.MethodPost,
		url:    "/node",
		body:   fmt.Sprintf(`{"nodeType":"%s"}`, nodeType),
	})
	var resp struct {
		NodeID string `json:"nodeID"`
	}
	require.NoError(t, json.Unmarshal(body, &resp))
	require.NotEmpty(t, resp.NodeID)
	return resp.NodeID
}

func connectNodes(t *testing.T, handler http.Handler, scopeURL, nodeOutID, outPort, nodeInID, inPort string) {
	t.Helper()
	body := fmt.Sprintf(
		`{"nodeOutId":"%s","outPortName":"%s","nodeInId":"%s","inPortName":"%s"}`,
		nodeOutID, outPort, nodeInID, inPort,
	)
	requireOK(t, handler, httpStep{
		method: http.MethodPost,
		url:    scopeURL,
		body:   body,
	})
}

func setBoundaryInfo(t *testing.T, handler http.Handler, nodeID, subGraphID, portName string) {
	t.Helper()
	body := fmt.Sprintf(
		`{"portName":"%s","scope":"subgraph/%s"}`,
		portName, subGraphID,
	)
	requireOK(t, handler, httpStep{
		method: http.MethodPost,
		url:    "/subgraph/boundary/" + nodeID + "/info",
		body:   body,
	})
}

func setParameterValue(t *testing.T, handler http.Handler, nodeID string, value string) {
	t.Helper()
	requireOK(t, handler, httpStep{
		method: http.MethodPost,
		url:    "/parameter/value/" + nodeID,
		body:   value,
	})
}

func subgraphOutputResult(t *testing.T, inst *graph.Instance, subGraphID, outputPortName string) float64 {
	t.Helper()
	child, err := inst.SubGraphInstance(subGraphID)
	require.NoError(t, err)

	for nodeID, nodeInst := range child.Schema().Nodes {
		if nodeInst.SubGraphOutputBoundary == nil {
			continue
		}
		if nodeInst.SubGraphOutputBoundary.PortName != outputPortName {
			continue
		}
		outNode, ok := child.Node(nodeID).(*subgraph.OutputNode)
		require.True(t, ok)
		source := outNode.ConnectedSource()
		require.NotNil(t, source)
		typed, ok := source.(nodes.Output[float64])
		require.True(t, ok)
		return typed.Value()
	}
	t.Fatalf("no output boundary named %q in sub-graph %q", outputPortName, subGraphID)
	return 0
}

func TestSubGraphEditServerEndToEnd(t *testing.T) {
	handler, inst := subGraphTestServer(t)
	const subGraphID = "adder"

	requireOK(t, handler, httpStep{
		method: http.MethodPost,
		url:    "/subgraph/definition/" + subGraphID,
		body:   `{"name":"Adder","description":"Adds two inputs"}`,
	})

	inputID := createScopedNode(t, handler, subGraphID, subgraph.InputNodeTypeKey, "float64")
	inputBID := createScopedNode(t, handler, subGraphID, subgraph.InputNodeTypeKey, "float64")
	outputID := createScopedNode(t, handler, subGraphID, subgraph.OutputNodeTypeKey, "float64")
	sumID := createScopedNode(t, handler, subGraphID, "Sum", "")

	setBoundaryInfo(t, handler, inputID, subGraphID, "A")
	setBoundaryInfo(t, handler, inputBID, subGraphID, "B")
	setBoundaryInfo(t, handler, outputID, subGraphID, "Result")

	scopedConnectionURL := fmt.Sprintf("/graph/subgraph/%s/connection", subGraphID)
	connectNodes(t, handler, scopedConnectionURL, inputID, subgraph.ValuePortName, sumID, "Values")
	connectNodes(t, handler, scopedConnectionURL, inputBID, subgraph.ValuePortName, sumID, "Values")
	connectNodes(t, handler, scopedConnectionURL, sumID, "Float", outputID, subgraph.ValuePortName)

	runtimeNodeID := createRootNode(t, handler, subgraph.RuntimeTypePath(subGraphID))
	paramAID := createRootNode(t, handler, "Float64")
	paramBID := createRootNode(t, handler, "Float64")

	setParameterValue(t, handler, paramAID, "5")
	setParameterValue(t, handler, paramBID, "3")

	connectNodes(t, handler, "/node/connection", paramAID, "Value", runtimeNodeID, "A")
	connectNodes(t, handler, "/node/connection", paramBID, "Value", runtimeNodeID, "B")

	assert.Equal(t, 8.0, subgraphOutputResult(t, inst, subGraphID, "Result"))

	schemaBody := requireOK(t, handler, httpStep{
		method: http.MethodGet,
		url:    "/schema",
		body:   "",
	})
	var schema struct {
		SubGraphs map[string]struct {
			Nodes map[string]struct {
				SubGraphInputBoundary  *struct{ PortName string } `json:"subGraphInputBoundary"`
				SubGraphOutputBoundary *struct{ PortName string } `json:"subGraphOutputBoundary"`
			} `json:"nodes"`
		} `json:"subGraphs"`
		Nodes map[string]struct {
			SubGraphId string `json:"subGraphId"`
		} `json:"nodes"`
	}
	require.NoError(t, json.Unmarshal(schemaBody, &schema))
	require.Contains(t, schema.SubGraphs, subGraphID)
	assert.Len(t, schema.SubGraphs[subGraphID].Nodes, 4)
	assert.Equal(t, subGraphID, schema.Nodes[runtimeNodeID].SubGraphId)

	ports, err := inst.CollectBoundaryPorts(subGraphID)
	require.NoError(t, err)
	require.Len(t, ports, 3)

	runtimeNode := inst.Node(runtimeNodeID)
	assert.Contains(t, runtimeNode.Inputs(), "A")
	assert.Contains(t, runtimeNode.Inputs(), "B")
	assert.Contains(t, runtimeNode.Outputs(), "Result")
}

func TestSubGraphDefinitionCreate(t *testing.T) {
	handler, inst := subGraphTestServer(t)

	body := `{"name":"My SubGraph","description":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/subgraph/definition/my-sub", strings.NewReader(body))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		NodeType struct {
			Type        string `json:"type"`
			DisplayName string `json:"displayName"`
		} `json:"nodeType"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "subgraph/my-sub", resp.NodeType.Type)
	assert.Equal(t, "My SubGraph", resp.NodeType.DisplayName)

	_, err := inst.SubGraphInstance("my-sub")
	assert.NoError(t, err)
}

func TestSubGraphDefinitionCreateDuplicate(t *testing.T) {
	handler, _ := subGraphTestServer(t)

	create := httptest.NewRequest(http.MethodPost, "/subgraph/definition/dup", strings.NewReader(`{"name":"One"}`))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, create)
	assert.Equal(t, http.StatusOK, rr.Code)

	dup := httptest.NewRequest(http.MethodPost, "/subgraph/definition/dup", strings.NewReader(`{"name":"Two"}`))
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, dup)
	assert.NotEqual(t, http.StatusOK, rr2.Code)
}

func TestSubGraphDefinitionUpdateInfo(t *testing.T) {
	handler, inst := subGraphTestServer(t)

	create := httptest.NewRequest(http.MethodPost, "/subgraph/definition/info-sub", strings.NewReader(`{"name":"Old"}`))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, create)
	require.Equal(t, http.StatusOK, rr.Code)

	update := httptest.NewRequest(http.MethodPut, "/subgraph/definition/info-sub", strings.NewReader(`{"name":"New","description":"updated"}`))
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, update)
	assert.Equal(t, http.StatusOK, rr2.Code)

	runtime := graph.NewRuntimeNode(inst, "info-sub")
	assert.Equal(t, "New", runtime.Name())
	assert.Equal(t, "updated", runtime.Description())
}

func TestSubGraphDefinitionDelete(t *testing.T) {
	handler, inst := subGraphTestServer(t)

	create := httptest.NewRequest(http.MethodPost, "/subgraph/definition/to-delete", strings.NewReader(`{"name":"Delete Me"}`))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, create)
	require.Equal(t, http.StatusOK, rr.Code)

	del := httptest.NewRequest(http.MethodDelete, "/subgraph/definition/to-delete", nil)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, del)
	assert.Equal(t, http.StatusOK, rr2.Code)

	_, err := inst.SubGraphInstance("to-delete")
	assert.Error(t, err)
}

func TestScopedSubGraphCreateNode(t *testing.T) {
	handler, inst := subGraphTestServer(t)

	create := httptest.NewRequest(http.MethodPost, "/subgraph/definition/scoped", strings.NewReader(`{"name":"Scoped"}`))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, create)
	require.Equal(t, http.StatusOK, rr.Code)

	nodeReq := httptest.NewRequest(http.MethodPost, "/graph/subgraph/scoped/node", strings.NewReader(`{"nodeType":"`+subgraph.InputNodeTypeKey+`","portType":"float64"}`))
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, nodeReq)
	assert.Equal(t, http.StatusOK, rr2.Code)

	var resp struct {
		NodeID string `json:"nodeID"`
	}
	require.NoError(t, json.Unmarshal(rr2.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp.NodeID)

	child, err := inst.SubGraphInstance("scoped")
	require.NoError(t, err)
	assert.True(t, child.HasNodeWithId(resp.NodeID))
}

func TestScopedSubGraphDeleteNode(t *testing.T) {
	handler, inst := subGraphTestServer(t)

	create := httptest.NewRequest(http.MethodPost, "/subgraph/definition/del-node", strings.NewReader(`{"name":"Del"}`))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, create)
	require.Equal(t, http.StatusOK, rr.Code)

	nodeReq := httptest.NewRequest(http.MethodPost, "/graph/subgraph/del-node/node", strings.NewReader(`{"nodeType":"`+subgraph.InputNodeTypeKey+`","portType":"float64"}`))
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, nodeReq)
	require.Equal(t, http.StatusOK, rr2.Code)

	var createResp struct {
		NodeID string `json:"nodeID"`
	}
	require.NoError(t, json.Unmarshal(rr2.Body.Bytes(), &createResp))

	delReq := httptest.NewRequest(http.MethodDelete, "/graph/subgraph/del-node/node", strings.NewReader(`{"nodeID":"`+createResp.NodeID+`"}`))
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, delReq)
	assert.Equal(t, http.StatusOK, rr3.Code)

	child, err := inst.SubGraphInstance("del-node")
	require.NoError(t, err)
	assert.False(t, child.HasNodeWithId(createResp.NodeID))
}

func TestSubGraphBoundaryInfoUpdate(t *testing.T) {
	handler, inst := subGraphTestServer(t)

	create := httptest.NewRequest(http.MethodPost, "/subgraph/definition/boundary", strings.NewReader(`{"name":"Boundary"}`))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, create)
	require.Equal(t, http.StatusOK, rr.Code)

	nodeReq := httptest.NewRequest(http.MethodPost, "/graph/subgraph/boundary/node", strings.NewReader(`{"nodeType":"`+subgraph.InputNodeTypeKey+`","portType":"float64"}`))
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, nodeReq)
	require.Equal(t, http.StatusOK, rr2.Code)

	var createResp struct {
		NodeID string `json:"nodeID"`
	}
	require.NoError(t, json.Unmarshal(rr2.Body.Bytes(), &createResp))

	boundaryBody := `{"portName":"Scale","scope":"subgraph/boundary"}`
	boundaryReq := httptest.NewRequest(http.MethodPost, "/subgraph/boundary/"+createResp.NodeID+"/info", strings.NewReader(boundaryBody))
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, boundaryReq)
	assert.Equal(t, http.StatusOK, rr3.Code)

	child, err := inst.SubGraphInstance("boundary")
	require.NoError(t, err)

	node := child.Node(createResp.NodeID)
	input, ok := node.(*subgraph.InputNode)
	require.True(t, ok)
	assert.Equal(t, "Scale", input.BoundaryPortName())
	assert.Equal(t, "float64", input.BoundaryPortType())
}

func TestScopedSubGraphNoteMetadata(t *testing.T) {
	handler, inst := subGraphTestServer(t)

	create := httptest.NewRequest(http.MethodPost, "/subgraph/definition/notes", strings.NewReader(`{"name":"Notes"}`))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, create)
	require.Equal(t, http.StatusOK, rr.Code)

	noteBody := `{"text":"hello","width":200,"position":{"x":10,"y":20}}`
	noteReq := httptest.NewRequest(http.MethodPost, "/graph/subgraph/notes/metadata/notes/sticky1", strings.NewReader(noteBody))
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, noteReq)
	require.Equal(t, http.StatusOK, rr2.Code)

	child, err := inst.SubGraphInstance("notes")
	require.NoError(t, err)
	notes := child.Schema().Notes
	require.NotNil(t, notes)
	assert.Equal(t, "hello", notes["sticky1"].(map[string]any)["text"])

	schemaReq := httptest.NewRequest(http.MethodGet, "/schema", nil)
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, schemaReq)
	require.Equal(t, http.StatusOK, rr3.Code)

	var schema struct {
		SubGraphs map[string]struct {
			Notes map[string]any `json:"notes"`
		} `json:"subGraphs"`
	}
	require.NoError(t, json.Unmarshal(rr3.Body.Bytes(), &schema))
	require.NotNil(t, schema.SubGraphs["notes"].Notes)
}

func TestNodeTypesIncludesPortTypes(t *testing.T) {
	handler, _ := subGraphTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/node-types", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		PortTypes []string `json:"portTypes"`
		NodeTypes []struct {
			Type string `json:"type"`
		} `json:"nodeTypes"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp.PortTypes)
	assert.NotEmpty(t, resp.NodeTypes)
}

func TestSchemaIncludesSubGraphs(t *testing.T) {
	handler, _ := subGraphTestServer(t)

	create := httptest.NewRequest(http.MethodPost, "/subgraph/definition/in-schema", strings.NewReader(`{"name":"In Schema"}`))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, create)
	require.Equal(t, http.StatusOK, rr.Code)

	schemaReq := httptest.NewRequest(http.MethodGet, "/schema", nil)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, schemaReq)
	require.Equal(t, http.StatusOK, rr2.Code)

	var schema struct {
		SubGraphs map[string]struct{} `json:"subGraphs"`
	}
	require.NoError(t, json.Unmarshal(rr2.Body.Bytes(), &schema))
	require.Contains(t, schema.SubGraphs, "in-schema")
}

func TestScopedSubGraphConnection(t *testing.T) {
	handler, inst := subGraphTestServer(t)

	create := httptest.NewRequest(http.MethodPost, "/subgraph/definition/conn", strings.NewReader(`{"name":"Conn"}`))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, create)
	require.Equal(t, http.StatusOK, rr.Code)

	createNode := func(nodeType, portType string) string {
		body := `{"nodeType":"` + nodeType + `"`
		if portType != "" {
			body += `,"portType":"` + portType + `"`
		}
		body += `}`
		req := httptest.NewRequest(http.MethodPost, "/graph/subgraph/conn/node", strings.NewReader(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		var resp struct {
			NodeID string `json:"nodeID"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		return resp.NodeID
	}

	inputID := createNode(subgraph.InputNodeTypeKey, "float64")
	outputID := createNode(subgraph.OutputNodeTypeKey, "float64")

	connBody := `{"nodeOutId":"` + inputID + `","outPortName":"Value","nodeInId":"` + outputID + `","inPortName":"Value"}`
	connReq := httptest.NewRequest(http.MethodPost, "/graph/subgraph/conn/connection", strings.NewReader(connBody))
	rrConn := httptest.NewRecorder()
	handler.ServeHTTP(rrConn, connReq)
	assert.Equal(t, http.StatusOK, rrConn.Code)

	child, err := inst.SubGraphInstance("conn")
	require.NoError(t, err)

	outNode := child.Node(outputID).(*subgraph.OutputNode)
	assert.NotNil(t, outNode.ConnectedSource())
}

func TestScopedSubGraphNamedMetadataCreatesNode(t *testing.T) {
	handler, inst := subGraphTestServer(t)

	create := httptest.NewRequest(http.MethodPost, "/subgraph/definition/metadata", strings.NewReader(`{"name":"Metadata"}`))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, create)
	require.Equal(t, http.StatusOK, rr.Code)

	nodeReq := httptest.NewRequest(http.MethodPost, "/graph/subgraph/metadata/node", strings.NewReader(`{"nodeType":"`+subgraph.InputNodeTypeKey+`","portType":"float64"}`))
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, nodeReq)
	require.Equal(t, http.StatusOK, rr2.Code)

	child, err := inst.SubGraphInstance("metadata")
	require.NoError(t, err)
	assert.Len(t, child.Schema().Nodes, 1)
}

func TestScopedSubGraphSlugWithSlash(t *testing.T) {
	handler, inst := subGraphTestServer(t)

	create := httptest.NewRequest(http.MethodPost, "/subgraph/definition/nested/id", strings.NewReader(`{"name":"Nested"}`))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, create)
	require.Equal(t, http.StatusOK, rr.Code)

	nodeReq := httptest.NewRequest(http.MethodPost, "/graph/subgraph/nested/id/node", strings.NewReader(`{"nodeType":"`+subgraph.InputNodeTypeKey+`","portType":"float64"}`))
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, nodeReq)
	require.Equal(t, http.StatusOK, rr2.Code)

	child, err := inst.SubGraphInstance("nested/id")
	require.NoError(t, err)
	assert.Len(t, child.Schema().Nodes, 1)
}
