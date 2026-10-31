package edit_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EliCDavis/polyform/generator/edit"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/variable/variabletypes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnumEndpointDeclaresAVariableTypeAndAParameterNode(t *testing.T) {
	server := edit.Server{
		Graph:           graph.New(graph.Config{TypeFactory: &refutil.TypeFactory{}, VariableFactory: variabletypes.New}),
		VariableFactory: variabletypes.New,
	}
	handler, err := server.Handler("./")
	require.NoError(t, err)

	do := func(method, url, body string) *httptest.ResponseRecorder {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(method, url, strings.NewReader(body)))
		return rr
	}

	rr := do(http.MethodPost, "/enum/instance/", `{"name":"Window Style","members":[{"name":"Casement"},{"name":"Sash"},{"name":"Fixed"}]}`)
	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	var created struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		NodeType struct {
			Type      string `json:"type"`
			Parameter struct {
				Type string `json:"type"`
			} `json:"parameter"`
		} `json:"nodeType"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &created))
	assert.Equal(t, "WindowStyle", created.ID)
	assert.Equal(t, "enum/WindowStyle", created.Type)
	assert.Equal(t, "enum-parameter/WindowStyle", created.NodeType.Type)
	assert.Equal(t, "enum/WindowStyle", created.NodeType.Parameter.Type)

	rr = do(http.MethodPost, "/variable/instance/Style", `{"type":"enum/WindowStyle"}`)
	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	assert.Contains(t, rr.Body.String(), `"Value":{"type":"enum/WindowStyle"}`)

	rr = do(http.MethodPost, "/variable/value/Style", `"Picture"`)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), `Style: \"Picture\" is not a member; expected Casement, Sash or Fixed`)

	rr = do(http.MethodGet, "/schema", "")
	require.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"enums":{"WindowStyle":{"name":"Window Style","members":[{"name":"Casement"},{"name":"Sash"},{"name":"Fixed"}]}}`)

	rr = do(http.MethodPut, "/enum/instance/WindowStyle", `{"name":"Window Style","members":[{"name":"Casement"},{"name":"Sliding"},{"name":"Fixed"}],"renamed":{"Sash":"Sliding"}}`)
	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())

	rr = do(http.MethodDelete, "/enum/instance/WindowStyle", "")
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "still used by variable Style")
}
