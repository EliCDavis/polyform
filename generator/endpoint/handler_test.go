package endpoint_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_NodefinedMethod_404s(t *testing.T) {
	handler := endpoint.Handler{
		Methods: map[string]endpoint.Method{},
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/bad", nil)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, 404, rr.Code)

	rawResponse, err := io.ReadAll(rr.Body)
	require.NoError(t, err)
	require.Equal(t, "", string(rawResponse))
}
