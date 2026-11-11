package variable_test

import (
	"encoding/json"
	"testing"

	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyingTheCurrentValueIsNotAChange(t *testing.T) {
	v := &variable.TypeVariable[float64]{}
	v.SetValue(4)
	version := v.Version()

	changed, err := v.ApplyMessage([]byte("4.0"))
	require.NoError(t, err)
	assert.False(t, changed)
	assert.Equal(t, version, v.Version())

	changed, err = v.ApplyMessage([]byte("5"))
	require.NoError(t, err)
	assert.True(t, changed)
	assert.NotEqual(t, version, v.Version())
	assert.Equal(t, 5., v.GetValue())
}

func TestApplyingTheCurrentFileIsNotAChange(t *testing.T) {
	v := &variable.FileVariable{}
	v.SetValue([]byte("ply"))
	version := v.Version()

	changed, err := v.ApplyMessage([]byte("ply"))
	require.NoError(t, err)
	assert.False(t, changed)
	assert.Equal(t, version, v.Version())

	changed, err = v.ApplyMessage([]byte("obj"))
	require.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, []byte("obj"), v.GetValue())
}

func TestSystemMatchesAProfileAlreadyInEffect(t *testing.T) {
	system := variable.NewSystem()
	width := &variable.TypeVariable[float64]{}
	width.SetValue(4)
	require.NoError(t, system.Add("width", width))

	assert.True(t, system.Matches(variable.Profile{"width": json.RawMessage("4.0")}))
	assert.False(t, system.Matches(variable.Profile{"width": json.RawMessage("5")}))
	assert.False(t, system.Matches(variable.Profile{"height": json.RawMessage("4")}))
}
