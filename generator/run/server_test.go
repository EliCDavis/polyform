package run_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/manifest/basics"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/run"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	tf := &refutil.TypeFactory{}
	// tf.RegisterBuilder("Float64", func() any {
	// 	return &parameter.Float64{
	// 		CurrentValue: 1,
	// 	}
	// })

	// tf.RegisterBuilder("Sum", func() any {
	// 	return &nodes.Struct[math.SumNodeData[float64]]{}
	// })

	tf.RegisterBuilder("Text", func() any {
		return &basics.TextNode{
			Data: basics.TextNodeData{
				In: nodes.GetNodeOutputPort[string](&parameter.String{
					CurrentValue: "Yee haw",
				}, "Value"),
			},
		}
	})

	graph := graph.New(graph.Config{
		TypeFactory: tf,
	})
	_, _, err := graph.CreateNode("Text")
	assert.NoError(t, err)

	server := run.Server{
		Graph:     graph,
		CacheSize: 1,
	}
	handler, err := server.Handler()
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/manifest/Node-1/Out", nil)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, 200, rr.Code)

	var createResponse run.CreateManifestResponse
	rawResponse, err := io.ReadAll(rr.Body)
	assert.NoError(t, err)
	require.NoError(t, json.Unmarshal(rawResponse, &createResponse))

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/manifest/%s/text.txt", createResponse.Id), nil)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, 200, rr.Code)

	rawResponse, err = io.ReadAll(rr.Body)
	require.NoError(t, err)
	require.Equal(t, "Yee haw", string(rawResponse))
}
