package edit_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/EliCDavis/polyform/generator/edit"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/manifest/basics"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/serialize"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type countedArtifact struct {
	manifest.Artifact
	writes *int
}

func (a countedArtifact) Write(w io.Writer) error {
	*a.writes++
	return a.Artifact.Write(w)
}

func outputServer(t *testing.T) http.Handler {
	t.Helper()
	handler, _ := countingOutputServer(t)
	return handler
}

func countingOutputServer(t *testing.T) (http.Handler, *int) {
	t.Helper()
	writes := new(int)
	tf := &refutil.TypeFactory{}
	tf.RegisterBuilder(subgraph.InputNodeTypeKey, func() any { return subgraph.NewInputNode("", "") })
	tf.RegisterBuilder(subgraph.OutputNodeTypeKey, func() any { return subgraph.NewOutputNode("", "") })
	tf.RegisterBuilder("Float64", func() any { return &parameter.Float64{CurrentValue: 7} })
	tf.RegisterBuilder("Text", func() any { return &nodes.Struct[basics.TextNode]{} })

	serializer := &serialize.TypeSwitch[manifest.Entry]{}
	serialize.Register(serializer, func(f float64) manifest.Entry {
		return manifest.Entry{Artifact: countedArtifact{Artifact: basics.Binary{Data: []byte(fmt.Sprint(f))}, writes: writes}}
	})

	instance := graph.New(graph.Config{TypeFactory: tf})
	server := edit.Server{Graph: instance, NodeOutputSerialization: serializer}
	handler, err := server.Handler("./")
	require.NoError(t, err)
	return handler, writes
}

func TestNodeOutputIsBuiltOncePerVersion(t *testing.T) {
	handler, writes := countingOutputServer(t)
	value := createRootNode(t, handler, "Float64")

	first := call(t, handler, http.MethodGet, "/node/output/"+value+"/Value", "")
	again := call(t, handler, http.MethodGet, "/node/output/"+value+"/Value?v=3", "")
	require.Equal(t, http.StatusOK, again.Code, again.Body.String())
	assert.Equal(t, "7", first.Body.String())
	assert.Equal(t, "7", again.Body.String())
	assert.Equal(t, 1, *writes)

	setParameterValue(t, handler, value, "9")
	changed := call(t, handler, http.MethodGet, "/node/output/"+value+"/Value", "")
	assert.Equal(t, "9", changed.Body.String())
	assert.Equal(t, 2, *writes)
}

func TestNodeOutputIsRebuiltForANewNodeAtTheSameIDAndVersion(t *testing.T) {
	handler := outputServer(t)
	value := createRootNode(t, handler, "Float64")
	setParameterValue(t, handler, value, "9")
	assert.Equal(t, "9", call(t, handler, http.MethodGet, "/node/output/"+value+"/Value", "").Body.String())

	requireOK(t, handler, httpStep{method: http.MethodPost, url: "/new-graph", body: `{}`})
	replacement := createRootNode(t, handler, "Float64")
	require.Equal(t, value, replacement)
	setParameterValue(t, handler, replacement, "8")

	assert.Equal(t, "8", call(t, handler, http.MethodGet, "/node/output/"+replacement+"/Value", "").Body.String())
}

func TestNodeOutputIsServedFromInsideASubGraphWhenScoped(t *testing.T) {
	handler := outputServer(t)

	requireOK(t, handler, httpStep{method: http.MethodPost, url: "/subgraph/definition/part", body: `{"name":"Part"}`})
	inner := createScopedNode(t, handler, "part", "Float64", "")
	outer := createRootNode(t, handler, "Float64")

	scoped := call(t, handler, http.MethodGet, "/node/output/"+inner+"/Value?scope=subgraph/part", "")
	require.Equal(t, http.StatusOK, scoped.Code, scoped.Body.String())
	assert.Equal(t, "7", scoped.Body.String())

	root := call(t, handler, http.MethodGet, "/node/output/"+outer+"/Value", "")
	require.Equal(t, http.StatusOK, root.Code, root.Body.String())
	assert.Equal(t, "7", root.Body.String())

	unknown := call(t, handler, http.MethodGet, "/node/output/"+inner+"/Value?scope=subgraph/nope", "")
	assert.Equal(t, http.StatusInternalServerError, unknown.Code)
}

func TestManifestNamesTheOutputVersionItWasMadeAt(t *testing.T) {
	handler := outputServer(t)
	text := createRootNode(t, handler, "Text")

	served := call(t, handler, http.MethodGet, "/manifest/"+text+"/Out", "")
	require.Equal(t, http.StatusOK, served.Code, served.Body.String())

	var graph struct {
		Nodes map[string]struct {
			Output map[string]struct {
				Version int `json:"version"`
			} `json:"output"`
		} `json:"nodes"`
	}
	require.NoError(t, json.Unmarshal(requireOK(t, handler, httpStep{method: http.MethodGet, url: "/schema"}), &graph))
	assert.Equal(t, strconv.Itoa(graph.Nodes[text].Output["Out"].Version), served.Header().Get("Output-Version"))
}

func TestPlacementOutputIsServedByItsLabel(t *testing.T) {
	handler := outputServer(t)

	requireOK(t, handler, httpStep{method: http.MethodPost, url: "/subgraph/definition/part", body: `{"name":"Part"}`})
	outputID := createScopedNode(t, handler, "part", subgraph.OutputNodeTypeKey, "float64")
	valueID := createScopedNode(t, handler, "part", "Float64", "")
	setBoundaryInfo(t, handler, outputID, "part", "Result")
	connectNodes(t, handler, "/graph/subgraph/part/connection", valueID, "Value", outputID, subgraph.ValuePortName)
	placement := createRootNode(t, handler, subgraph.RuntimeTypePath("part"))

	served := call(t, handler, http.MethodGet, "/node/output/"+placement+"/Result", "")
	require.Equal(t, http.StatusOK, served.Code, served.Body.String())
	assert.Equal(t, "7", served.Body.String())
}
