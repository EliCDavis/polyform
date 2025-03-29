package endpoint_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/stretchr/testify/assert"
)

func TestBodyMethod(t *testing.T) {
	// ARRANGE ================================================================
	type Body struct {
		Blah int `json:"blah"`
	}
	var readBody Body
	handler := endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodGet: endpoint.BodyMethod[Body]{
				Request: endpoint.JsonRequestReader[Body]{},
				Handler: func(request endpoint.Request[Body]) error {
					readBody = request.Body
					return nil
				},
			},
		},
	}
	req := httptest.NewRequest("GET", "http://example.com/foo", bytes.NewReader([]byte(`{"blah":42}`)))
	w := httptest.NewRecorder()

	// ACT ====================================================================
	handler.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	// ASSERT =================================================================

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "", resp.Header.Get("Content-Type"))
	assert.Equal(t, "", string(body))
	assert.Equal(t, 42, readBody.Blah)
}

func TestBodyMethod_ErrorReadingMalformedBody(t *testing.T) {
	// ARRANGE ================================================================
	type Body struct {
		Blah int `json:"blah"`
	}
	var readBody Body
	handler := endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodGet: endpoint.BodyMethod[Body]{
				Request: endpoint.JsonRequestReader[Body]{},
				Handler: func(request endpoint.Request[Body]) error {
					readBody = request.Body
					return nil
				},
			},
		},
	}
	req := httptest.NewRequest("GET", "http://example.com/foo", bytes.NewReader([]byte(`{"blah:42}`)))
	w := httptest.NewRecorder()

	// ACT ====================================================================
	handler.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	// ASSERT =================================================================

	assert.Equal(t, 500, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, `{"error":"unexpected end of JSON input"}`, string(body))
	assert.Equal(t, 0, readBody.Blah)
}

func TestBodyMethod_PanicRecoveryRunningHandler(t *testing.T) {
	// ARRANGE ================================================================
	type Body struct {
		Blah int `json:"blah"`
	}
	var readBody Body
	handler := endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodGet: endpoint.BodyMethod[Body]{
				Request: endpoint.JsonRequestReader[Body]{},
				Handler: func(request endpoint.Request[Body]) error {
					panic("we freakin")
				},
			},
		},
	}
	req := httptest.NewRequest("GET", "http://example.com/foo", bytes.NewReader([]byte(`{"blah":42}`)))
	w := httptest.NewRecorder()

	// ACT ====================================================================
	handler.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	// ASSERT =================================================================

	assert.Equal(t, 500, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, `{"error":"panic recover: we freakin"}`, string(body))
	assert.Equal(t, 0, readBody.Blah)
}
