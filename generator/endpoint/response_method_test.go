package endpoint_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/stretchr/testify/assert"
)

func TestResponseMethod(t *testing.T) {
	// ARRANGE ================================================================
	type Response struct {
		Result int `json:"result"`
	}
	handler := endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodGet: endpoint.ResponseMethod[Response]{
				ResponseWriter: endpoint.JsonResponseWriter[Response]{},
				Handler: func(r *http.Request) (Response, error) {
					return Response{Result: 2}, nil
				},
			},
		},
	}
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()

	// ACT ====================================================================
	handler.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	// ASSERT =================================================================

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, `{"result":2}`, string(body))
}

func TestResponseMethod_ErrorPanicRecovery(t *testing.T) {
	// ARRANGE ================================================================
	type Response struct {
		Result int `json:"result"`
	}
	handler := endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodGet: endpoint.ResponseMethod[Response]{
				ResponseWriter: endpoint.JsonResponseWriter[Response]{},
				Handler: func(r *http.Request) (Response, error) {
					panic("yee haw")
				},
			},
		},
	}
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
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
