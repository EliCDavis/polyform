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
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/math"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type historyBody struct {
	Undo    []string `json:"undo"`
	Redo    []string `json:"redo"`
	Applied string   `json:"applied"`
}

func historyServer(t *testing.T) (http.Handler, *graph.Instance) {
	t.Helper()

	tf := &refutil.TypeFactory{}
	tf.RegisterBuilder("Float64", func() any { return &parameter.Float64{CurrentValue: 1} })
	tf.RegisterBuilder("Sum", func() any { return &nodes.Struct[math.AddNode[float64]]{} })

	variableFactory := func(s string) (variable.Variable, error) {
		if s == "float64" {
			return &variable.TypeVariable[float64]{}, nil
		}
		return nil, fmt.Errorf("unrecognized variable: %s", s)
	}

	instance := graph.New(graph.Config{TypeFactory: tf, VariableFactory: variableFactory})
	server := edit.Server{Graph: instance, VariableFactory: variableFactory}
	handler, err := server.Handler("./")
	require.NoError(t, err)
	return handler, instance
}

func call(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(method, path, reader))
	return rr
}

func readHistory(t *testing.T, rr *httptest.ResponseRecorder) historyBody {
	t.Helper()
	var out historyBody
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out), "body: %s", rr.Body.String())
	return out
}

func TestEditorUndoTakesBackANodeCreation(t *testing.T) {
	handler, instance := historyServer(t)

	call(t, handler, http.MethodPost, "/node", `{"nodeType": "Sum"}`)
	require.Len(t, instance.Schema().Nodes, 1)

	listed := readHistory(t, call(t, handler, http.MethodGet, "/graph/history", ""))
	assert.Equal(t, []string{"Update node"}, listed.Undo)

	undone := readHistory(t, call(t, handler, http.MethodPost, "/graph/history/undo", ""))
	assert.Equal(t, "Update node", undone.Applied)
	assert.Empty(t, instance.Schema().Nodes)
	assert.Equal(t, []string{"Update node"}, undone.Redo)

	redone := readHistory(t, call(t, handler, http.MethodPost, "/graph/history/redo", ""))
	assert.Equal(t, "Update node", redone.Applied)
	assert.Len(t, instance.Schema().Nodes, 1)
}

func TestEditorUndoCoversLayoutChanges(t *testing.T) {
	handler, instance := historyServer(t)
	call(t, handler, http.MethodPost, "/node", `{"nodeType": "Sum"}`)

	call(t, handler, http.MethodPost, "/graph/metadata/nodes/Node-0/position", `{"x": 40, "y": 80}`)
	listed := readHistory(t, call(t, handler, http.MethodGet, "/graph/history", ""))
	require.Equal(t, "Update layout", listed.Undo[0], "a drag is its own step")

	call(t, handler, http.MethodPost, "/graph/history/undo", "")
	assert.Nil(t, instance.Metadata("nodes.Node-0.position"))
}

func TestEditorUndoOnAFreshGraphIsRefusedRatherThanSilent(t *testing.T) {
	handler, _ := historyServer(t)

	rr := call(t, handler, http.MethodPost, "/graph/history/undo", "")
	assert.GreaterOrEqual(t, rr.Code, 400)
	assert.Contains(t, rr.Body.String(), "nothing to undo")
}

func TestEditorRollsBackARefusedEdit(t *testing.T) {
	handler, instance := historyServer(t)
	call(t, handler, http.MethodPost, "/node", `{"nodeType": "Sum"}`)
	before := len(instance.Schema().Nodes)

	rr := call(t, handler, http.MethodPost, "/node", `{"nodeType": "NoSuchType"}`)
	require.GreaterOrEqual(t, rr.Code, 400)

	assert.Equal(t, before, len(instance.Schema().Nodes))
	listed := readHistory(t, call(t, handler, http.MethodGet, "/graph/history", ""))
	assert.Len(t, listed.Undo, 1, "the failed request is not a step")
}
