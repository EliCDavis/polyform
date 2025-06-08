package variable_test

import (
	"encoding/json"
	"testing"

	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToFromSerialize(t *testing.T) {
	// ARRANGE ================================================================
	floatVariable := &variable.TypeVariable[float64]{}
	floatVariable.SetValue(3.)
	container := &variable.JsonContainer{}

	// ACT ====================================================================
	jsonOutput, marshallErr := json.MarshalIndent(floatVariable, "", "\t")
	unmarshalErr := json.Unmarshal(jsonOutput, container)
	backAgain, backErr := json.MarshalIndent(container, "", "\t")

	// ASSERT =================================================================
	require.NoError(t, marshallErr)
	require.NoError(t, unmarshalErr)
	require.NoError(t, backErr)
	assert.Equal(t, `{
	"type": "float64",
	"value": 3
}`, string(jsonOutput))

	_, ok := container.Variable.(*variable.TypeVariable[float64])
	require.True(t, ok)
	assert.Equal(t, string(jsonOutput), string(backAgain))
}
