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

func TestBodyResponseMethod(t *testing.T) {
	// ARRANGE ================================================================
	type Body struct {
		Blah int `json:"blah"`
	}
	type Response struct {
		Result int `json:"result"`
	}
	handler := endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodGet: endpoint.BodyResponseMethod[Body, Response]{
				Request:        endpoint.JsonRequestReader[Body]{},
				ResponseWriter: endpoint.JsonResponseWriter[Response]{},
				Handler: func(request endpoint.Request[Body]) (Response, error) {
					return Response{Result: request.Body.Blah * 2}, nil
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
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, `{"result":84}`, string(body))
}

func TestBodyResponseMethod_ErrorReadingMalformedBody(t *testing.T) {
	// ARRANGE ================================================================
	type Body struct {
		Blah int `json:"blah"`
	}
	type Response struct {
		Result int `json:"result"`
	}
	handler := endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodGet: endpoint.BodyResponseMethod[Body, Response]{
				Request:        endpoint.JsonRequestReader[Body]{},
				ResponseWriter: endpoint.JsonResponseWriter[Response]{},
				Handler: func(request endpoint.Request[Body]) (Response, error) {
					return Response{Result: request.Body.Blah * 2}, nil
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
}

func TestBodyResponseMethod_ErrorPanicRecovery(t *testing.T) {
	// ARRANGE ================================================================
	type Body struct {
		Blah int `json:"blah"`
	}
	type Response struct {
		Result int `json:"result"`
	}
	handler := endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodGet: endpoint.BodyResponseMethod[Body, Response]{
				Request:        endpoint.JsonRequestReader[Body]{},
				ResponseWriter: endpoint.JsonResponseWriter[Response]{},
				Handler: func(request endpoint.Request[Body]) (Response, error) {
					panic("yee haw")
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
	assert.Equal(t, `{"error":"panic recover: yee haw"}`, string(body))
}
