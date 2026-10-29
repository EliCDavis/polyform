package variable_test

import (
	"encoding/json"
	"testing"

	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyingAProfileLeavesUnchangedVariablesAtTheirVersion(t *testing.T) {
	system := variable.NewSystem()
	width := &variable.TypeVariable[float64]{}
	width.SetValue(4)
	height := &variable.TypeVariable[float64]{}
	height.SetValue(2)
	require.NoError(t, system.Add("width", width))
	require.NoError(t, system.Add("height", height))
	widthVersion, heightVersion := width.Version(), height.Version()

	require.NoError(t, system.ApplyProfile(variable.Profile{
		"width":  json.RawMessage("4.0"),
		"height": json.RawMessage("3"),
	}))

	assert.Equal(t, widthVersion, width.Version())
	assert.NotEqual(t, heightVersion, height.Version())
	assert.Equal(t, 3., height.GetValue())
}
