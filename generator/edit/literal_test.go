package edit_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EliCDavis/polyform/generator/edit"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/variable/variabletypes"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type literalEndpointNode struct {
	Radius nodes.Output[float64] `default:"0.5"`
	Label  nodes.Output[string]
}

func (n literalEndpointNode) Out(out *nodes.StructOutput[float64]) {
	out.Set(nodes.TryGetOutputValue(out, n.Radius, 0.5) * 2)
}

func TestLiteralEndpoint(t *testing.T) {
	factory := &refutil.TypeFactory{}
	factory.RegisterBuilder("Sink", func() any { return &nodes.Struct[literalEndpointNode]{} })
	server := edit.Server{
		Graph:           graph.New(graph.Config{TypeFactory: factory, VariableFactory: variabletypes.New}),
		VariableFactory: variabletypes.New,
	}
	handler, err := server.Handler("./")
	require.NoError(t, err)

	do := func(method, url, body string) *httptest.ResponseRecorder {
		t.Helper()
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(method, url, strings.NewReader(body)))
		return rr
	}

	rr := do(http.MethodPost, "/node", `{"nodeType": "Sink"}`)
	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	assert.Contains(t, rr.Body.String(), `"literals":{"Label":"","Radius":0.5}`)

	rr = do(http.MethodPost, "/node/literal/Node-0/Radius", `2.5`)
	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())

	rr = do(http.MethodGet, "/node/literal/Node-0/Radius", "")
	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "2.5", rr.Body.String())

	rr = do(http.MethodGet, "/schema", "")
	assert.Contains(t, rr.Body.String(), `"Radius":2.5`)

	rr = do(http.MethodGet, "/node-types", "")
	assert.Contains(t, rr.Body.String(), `"literal":true,"default":0.5`)

	rr = do(http.MethodPost, "/node/literal/Node-0/Radius", `"wide"`)
	assert.GreaterOrEqual(t, rr.Code, http.StatusBadRequest)
	rr = do(http.MethodPost, "/node/literal/Node-0/Nope", `1`)
	assert.GreaterOrEqual(t, rr.Code, http.StatusBadRequest)
	rr = do(http.MethodPost, "/node/literal/Node-0", `1`)
	assert.GreaterOrEqual(t, rr.Code, http.StatusBadRequest)

	rr = do(http.MethodPost, "/graph/history/undo", "")
	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	rr = do(http.MethodGet, "/node/literal/Node-0/Radius", "")
	assert.Equal(t, "0.5", rr.Body.String(), "the refused edits left no steps, so the undo reverts the 2.5")
}
