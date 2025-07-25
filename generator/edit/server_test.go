package edit_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EliCDavis/polyform/generator/edit"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/math"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	tf := &refutil.TypeFactory{}
	tf.RegisterBuilder("Float64", func() any {
		return &parameter.Float64{
			CurrentValue: 1,
		}
	})

	tf.RegisterBuilder("Sum", func() any {
		return &nodes.Struct[math.SumNodeData[float64]]{}
	})

	server := edit.Server{
		Graph: graph.New(graph.Config{
			TypeFactory: tf,
		}),
	}
	handler, err := server.Handler("./")
	assert.NoError(t, err)

	type Step struct {
		name   string
		req    *http.Request
		assert func(*httptest.ResponseRecorder)
	}

	steps := []Step{
		{
			req: httptest.NewRequest(http.MethodPost, "/new-graph", strings.NewReader(`{
				"name": "HTTP Test",
				"description": "This test takes place in mock http",
				"version": "test",
				"author": "Test Runner"
			}`)),
		},

		{
			name: "Create Variable",
			req: httptest.NewRequest(http.MethodPost, "/variable/instance/MyVariable", strings.NewReader(`{
				"type": "float64",
				"description": "Variable HTTP test"
			}`)),
			assert: func(rr *httptest.ResponseRecorder) {
				assert.Equal(t, `{"nodeType":{"displayName":"MyVariable","info":"Variable HTTP test","type":"MyVariable","path":"generator/variable","outputs":{"Value":{"type":"float64"}}}}`, rr.Body.String())
			},
		},

		{
			name: "Get Variable",
			req:  httptest.NewRequest(http.MethodGet, "/variable/instance/MyVariable", nil),
			assert: func(rr *httptest.ResponseRecorder) {
				assert.Equal(t, `{"type":"float64","value":0}`, rr.Body.String())
			},
		},

		{
			name: "Set Variable Value",
			req:  httptest.NewRequest(http.MethodPost, "/variable/value/MyVariable", strings.NewReader(`12.34`)),
			assert: func(rr *httptest.ResponseRecorder) {
				assert.Equal(t, "", rr.Body.String())
			},
		},

		{
			name: "Get Variable Value",
			req:  httptest.NewRequest(http.MethodGet, "/variable/value/MyVariable", nil),
			assert: func(rr *httptest.ResponseRecorder) {
				assert.Equal(t, "12.34", rr.Body.String())
			},
		},

		{
			name: "Create Profile",
			req:  httptest.NewRequest(http.MethodPost, "/profile", strings.NewReader(`{ "name": "My Test Profile" }`)),
			assert: func(rr *httptest.ResponseRecorder) {
				assert.Equal(t, "", rr.Body.String())
			},
		},

		{
			name: "Rename Profile",
			req:  httptest.NewRequest(http.MethodPost, "/profile/rename", strings.NewReader(`{ "original": "My Test Profile", "new": "Renamed Profile" }`)),
			assert: func(rr *httptest.ResponseRecorder) {
				assert.Equal(t, "", rr.Body.String())
			},
		},

		{
			name: "Get Node Types",
			req:  httptest.NewRequest(http.MethodGet, "/node-types", nil),
			assert: func(rr *httptest.ResponseRecorder) {
				out := make([]json.RawMessage, 0)
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
				assert.Equal(t, `{"displayName":"parameter.Value[float64]","info":"","type":"Float64","path":"generator/parameter","outputs":{"Value":{"type":"float64"}},"parameter":{"name":"","description":"","type":"float64","currentValue":1}}`, string(out[0]))
				assert.Equal(t, `{"displayName":"MyVariable","info":"Variable HTTP test","type":"MyVariable","path":"generator/variable","outputs":{"Value":{"type":"float64"}}}`, string(out[1]))
				assert.Equal(t, `{"displayName":"Sum[float64]","info":"","type":"Sum","path":"math","outputs":{"Out":{"type":"float64"}},"inputs":{"Values":{"type":"float64","isArray":true,"description":"The nodes to sum"}}}`, string(out[2]))
				assert.Len(t, out, 3)
			},
		},

		// Node 1
		{
			req: httptest.NewRequest(http.MethodPost, "/node", strings.NewReader(`{
				"nodeType": "Float64"
			}`)),
			assert: func(rr *httptest.ResponseRecorder) {
				assert.True(t, strings.Contains(rr.Body.String(), `"nodeID":"Node-0"`))
			},
		},

		// Node 2
		{
			req: httptest.NewRequest(http.MethodPost, "/node", strings.NewReader(`{
				"nodeType": "Float64"
			}`)),
			assert: func(rr *httptest.ResponseRecorder) {
				assert.True(t, strings.Contains(rr.Body.String(), `"nodeID":"Node-1"`))
			},
		},

		// Node 3
		{
			req: httptest.NewRequest(http.MethodPost, "/node", strings.NewReader(`{
				"nodeType": "Sum"
			}`)),
			assert: func(rr *httptest.ResponseRecorder) {
				assert.True(t, strings.Contains(rr.Body.String(), `"nodeID":"Node-2"`))
			},
		},

		// Connect 1 => 3
		{
			req: httptest.NewRequest(http.MethodPost, "/node/connection", strings.NewReader(`{
				"nodeOutId": "Node-0",
				"outPortName": "Value",
				"nodeInId": "Node-2",
				"inPortName": "Values"
			}`)),
			assert: func(rr *httptest.ResponseRecorder) {
				assert.Equal(t, "{}", rr.Body.String())
			},
		},

		// Connect 2 => 3
		{
			req: httptest.NewRequest(http.MethodPost, "/node/connection", strings.NewReader(`{
				"nodeOutId": "Node-1",
				"outPortName": "Value",
				"nodeInId": "Node-2",
				"inPortName": "Values"
			}`)),
			assert: func(rr *httptest.ResponseRecorder) {
				assert.Equal(t, "{}", rr.Body.String())
			},
		},

		// Set Parameter Name
		{
			req: httptest.NewRequest(http.MethodPost, "/parameter/name/Node-0", strings.NewReader(`My Parameter Name`)),
			assert: func(rr *httptest.ResponseRecorder) {
				assert.Equal(t, "", rr.Body.String())
			},
		},

		// Set Parameter Description
		{
			req: httptest.NewRequest(http.MethodPost, "/parameter/description/Node-0", strings.NewReader(`My Parameter Description`)),
			assert: func(rr *httptest.ResponseRecorder) {
				assert.Equal(t, "", rr.Body.String())
			},
		},

		// Set Parameter Value
		{
			req: httptest.NewRequest(http.MethodPost, "/parameter/value/Node-0", strings.NewReader(`32.1`)),
			assert: func(rr *httptest.ResponseRecorder) {
				assert.Equal(t, "", rr.Body.String())
			},
		},

		{
			req: httptest.NewRequest(http.MethodGet, "/graph", nil),
			assert: func(rr *httptest.ResponseRecorder) {
				assert.Equal(t, `{
	"buffers": [
		{
			"byteLength": 0,
			"uri": "data:application/octet-stream;base64,"
		}
	],
	"data": {
		"authors": [
			{
				"name": "Test Runner"
			}
		],
		"description": "This test takes place in mock http",
		"name": "HTTP Test",
		"nodes": {
			"Node-0": {
				"type": "github.com/EliCDavis/polyform/generator/parameter.Value[float64]",
				"data": {
					"name": "My Parameter Name",
					"description": "My Parameter Description",
					"currentValue": 32.1
				}
			},
			"Node-1": {
				"type": "github.com/EliCDavis/polyform/generator/parameter.Value[float64]",
				"data": {
					"name": "",
					"description": "",
					"currentValue": 1
				}
			},
			"Node-2": {
				"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.SumNodeData[float64]]",
				"assignedInput": {
					"Values.0": {
						"id": "Node-0",
						"port": "Value"
					},
					"Values.1": {
						"id": "Node-1",
						"port": "Value"
					}
				}
			}
		},
		"producers": {},
		"profiles": {
			"Renamed Profile": {
				"data": {
					"MyVariable": 12.34
				}
			}
		},
		"variables": {
			"subgroups": {},
			"variables": {
				"MyVariable": {
					"description": "Variable HTTP test",
					"data": {
						"type": "float64",
						"value": 12.34
					}
				}
			}
		},
		"version": "test"
	}
}`, rr.Body.String())
			},
		},
	}

	for _, step := range steps {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, step.req)
		assert.Equal(t, 200, rr.Code, step.name)

		if step.assert != nil {
			step.assert(rr)
		}
	}

}
