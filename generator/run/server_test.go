package run_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/manifest/basics"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/run"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func typeFactory() *refutil.TypeFactory {
	tf := &refutil.TypeFactory{}
	tf.RegisterBuilder("Text", func() any {
		return &nodes.Struct[basics.TextNode]{
			Data: basics.TextNode{
				In: nodes.GetNodeOutputPort[string](&parameter.String{
					CurrentValue: "Yee haw",
				}, "Value"),
			},
		}
	})
	return tf
}

func createManifest(t *testing.T, handler http.Handler) run.CreateManifestResponse {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/manifest/Node-1/Out", nil)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, 200, rr.Code)

	var createResponse run.CreateManifestResponse
	rawResponse, err := io.ReadAll(rr.Body)
	assert.NoError(t, err)
	require.NoError(t, json.Unmarshal(rawResponse, &createResponse))
	return createResponse
}

func TestServer_ProducesManifest(t *testing.T) {
	graph := graph.New(graph.Config{
		TypeFactory: typeFactory(),
	})
	_, _, err := graph.CreateNode("Text")
	assert.NoError(t, err)

	server := run.Server{
		Graph:     graph,
		CacheSize: 1,
	}
	handler, err := server.Handler()
	assert.NoError(t, err)

	createResponse := createManifest(t, handler)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/manifest/%s/text.txt", createResponse.Id), nil)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, 200, rr.Code)

	rawResponse, err := io.ReadAll(rr.Body)
	require.NoError(t, err)
	require.Equal(t, "Yee haw", string(rawResponse))
}

func TestServer_FailureCases(t *testing.T) {
	graph := graph.New(graph.Config{
		TypeFactory: typeFactory(),
	})

	_, _, err := graph.CreateNode("Text")
	require.NoError(t, err)

	server := run.Server{
		Graph:     graph,
		CacheSize: 1,
	}
	handler, err := server.Handler()
	assert.NoError(t, err)

	createResponse := createManifest(t, handler)

	tests := map[string]struct {
		req    *http.Request
		assert func(*testing.T, []byte)
		Code   int
	}{
		"POST /manifest/ - 500s": {
			Code: 500,
			req:  httptest.NewRequest(http.MethodPost, "/manifest/", nil),
			assert: func(t *testing.T, body []byte) {
				assert.Equal(t, `{"error":"url is missing manifest node name or id"}`, string(body))
			},
		},

		"POST /manifest/bad-node - 500s": {
			Code: 500,
			req:  httptest.NewRequest(http.MethodPost, "/manifest/bad-node", nil),
			assert: func(t *testing.T, body []byte) {
				assert.Equal(t, `{"error":"url is missing manifest port name"}`, string(body))
			},
		},

		"POST /manifest/bad-node/bad-port - 500s": {
			Code: 500,
			req:  httptest.NewRequest(http.MethodPost, "/manifest/bad-node/bad-port", nil),
			assert: func(t *testing.T, body []byte) {
				assert.Equal(t, `{"error":"bad-node/bad-port does not match any node/port combination that produces a manifest"}`, string(body))
			},
		},

		"POST /manifest/Node-1/bad-port - 500s": {
			Code: 500,
			req:  httptest.NewRequest(http.MethodPost, "/manifest/Node-1/bad-port", nil),
			assert: func(t *testing.T, body []byte) {
				assert.Equal(t, `{"error":"Node-1/bad-port does not match any node/port combination that produces a manifest"}`, string(body))
			},
		},

		"POST /manifest/Node-0/Value - 500s (not a manifest port)": {
			Code: 500,
			req:  httptest.NewRequest(http.MethodPost, "/manifest/Node-0/Value", nil),
			assert: func(t *testing.T, body []byte) {
				assert.Equal(t, `{"error":"Node-0/Value does not match any node/port combination that produces a manifest"}`, string(body))
			},
		},

		"POST /manifest/Node-1/Out - bad json - 500s": {
			Code: 500,
			req:  httptest.NewRequest(http.MethodPost, "/manifest/Node-1/Out", strings.NewReader(`{bad}`)),
			assert: func(t *testing.T, body []byte) {
				assert.Equal(t, `{"error":"unable to interpret variable profile"}`, string(body))
			},
		},

		"POST /manifest/Node-1/Out - bad profile - 500s": {
			Code: 500,
			req:  httptest.NewRequest(http.MethodPost, "/manifest/Node-1/Out", strings.NewReader(`{"bad":"somethin"}`)),
			assert: func(t *testing.T, body []byte) {
				assert.Equal(t, `{"error":"unable to apply profile: unable to apply \"bad\" from profile, variable does not exist"}`, string(body))
			},
		},

		"GET /manifest/ - 500s": {
			Code: 500,
			req:  httptest.NewRequest(http.MethodGet, "/manifest/", nil),
			assert: func(t *testing.T, body []byte) {
				assert.Equal(t, `{"error":"invalid url: \"/manifest/\""}`, string(body))
			},
		},

		"GET /manifest/bad-id - 500s": {
			Code: 500,
			req:  httptest.NewRequest(http.MethodGet, "/manifest/bad-id", nil),
			assert: func(t *testing.T, body []byte) {
				assert.Equal(t, `{"error":"invalid url: \"/manifest/bad-id\""}`, string(body))
			},
		},

		"GET /manifest/bad-id/bad-file - 500s": {
			Code: 500,
			req:  httptest.NewRequest(http.MethodGet, "/manifest/bad-id/bad-file", nil),
			assert: func(t *testing.T, body []byte) {
				assert.Equal(t, `{"error":"no manifest exists with id \"bad-id\""}`, string(body))
			},
		},
		"GET /manifest/good-id/bad-file - 500s": {
			Code: 500,
			req:  httptest.NewRequest(http.MethodGet, fmt.Sprintf("/manifest/%s/bad-file", createResponse.Id), nil),
			assert: func(t *testing.T, body []byte) {
				assert.Equal(
					t,
					fmt.Sprintf(`{"error":"manifest \"%s\" contains no entry \"bad-file\""}`, createResponse.Id),
					string(body),
				)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, tc.req)
			assert.Equal(t, tc.Code, rr.Code)
			body, err := io.ReadAll(rr.Body)
			require.NoError(t, err)
			if tc.assert != nil {
				tc.assert(t, body)
			}
		})
	}
}

func TestServer_GetAllManifests(t *testing.T) {
	// ARRANGE ================================================================
	graph := graph.New(graph.Config{
		TypeFactory: typeFactory(),
	})
	_, _, err := graph.CreateNode("Text")
	require.NoError(t, err)

	server := run.Server{
		Graph:     graph,
		CacheSize: 1,
	}
	handler, err := server.Handler()
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/manifests", nil)

	// ACT ====================================================================
	handler.ServeHTTP(rr, req)
	rawResponse, err := io.ReadAll(rr.Body)

	// ASSERT =================================================================
	assert.Equal(t, 200, rr.Code)
	require.NoError(t, err)
	require.Equal(t, `[{"name":"Node-1","port":"Out"}]`, string(rawResponse))
}

func TestServer_NamedManifest(t *testing.T) {
	// ARRANGE ================================================================
	graph := graph.New(graph.Config{
		TypeFactory: typeFactory(),
	})
	_, _, err := graph.CreateNode("Text")
	require.NoError(t, err)

	graph.SetNodeAsProducer("Node-1", "Out", "CoolName")

	server := run.Server{
		Graph:     graph,
		CacheSize: 1,
	}
	handler, err := server.Handler()
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/manifests", nil)

	// ACT ====================================================================
	handler.ServeHTTP(rr, req)
	rawResponse, err := io.ReadAll(rr.Body)

	// ASSERT =================================================================
	assert.Equal(t, 200, rr.Code)
	require.NoError(t, err)
	require.Equal(t, `[{"name":"CoolName","port":"Out"}]`, string(rawResponse))
}

func TestServer_VariableProfile(t *testing.T) {
	// ARRANGE ================================================================
	graph := graph.New(graph.Config{
		TypeFactory: typeFactory(),
	})

	graph.NewVariable("Test Variable", &variable.TypeVariable[float64]{})

	server := run.Server{
		Graph:     graph,
		CacheSize: 1,
	}
	handler, err := server.Handler()
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/profile", nil)

	// ACT ====================================================================
	handler.ServeHTTP(rr, req)
	rawResponse, err := io.ReadAll(rr.Body)

	// ASSERT =================================================================
	assert.Equal(t, 200, rr.Code)
	require.NoError(t, err)
	require.Equal(t, `{"Test Variable":{"type":"number","format":"double"}}`, string(rawResponse))
}
