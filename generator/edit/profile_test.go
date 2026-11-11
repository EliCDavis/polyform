package edit_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyingTheProfileInEffectIsNotAStep(t *testing.T) {
	handler, instance := historyServer(t)
	call(t, handler, http.MethodPost, "/variable/instance/Height", `{"type": "float64"}`)
	call(t, handler, http.MethodPost, "/variable/value/Height", `5`)
	call(t, handler, http.MethodPost, "/profile", `{"name": "tall"}`)
	before := readHistory(t, call(t, handler, http.MethodGet, "/graph/history", "")).Undo

	rr := call(t, handler, http.MethodPost, "/profile/apply", `{"name": "tall"}`)
	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	assert.Equal(t, before, readHistory(t, call(t, handler, http.MethodGet, "/graph/history", "")).Undo)

	call(t, handler, http.MethodPost, "/variable/value/Height", `7`)
	rr = call(t, handler, http.MethodPost, "/profile/apply", `{"name": "tall"}`)
	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	after := readHistory(t, call(t, handler, http.MethodGet, "/graph/history", "")).Undo
	require.NotEmpty(t, after)
	assert.Equal(t, "Update applied profile", after[0])

	value, err := instance.VariableData("Height")
	require.NoError(t, err)
	assert.JSONEq(t, `5`, string(value))
}
