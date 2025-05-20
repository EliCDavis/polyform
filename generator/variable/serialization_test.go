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
	floatVariable.SetName("Test Name")
	floatVariable.SetDescription("Test Description")
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
	"name": "Test Name",
	"type": "float64",
	"description": "Test Description",
	"value": 3
}`, string(jsonOutput))

	casted, ok := container.Variable.(*variable.TypeVariable[float64])
	require.True(t, ok)
	assert.Equal(t, floatVariable.Name(), casted.Name())
	assert.Equal(t, string(jsonOutput), string(backAgain))
}
