package variabletypes_test

import (
	"encoding/json"
	"testing"

	"github.com/EliCDavis/polyform/generator/variable/variabletypes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeysRoundTripThroughTheVariablesOwnType(t *testing.T) {
	resolver := refutil.TypeResolution{IncludePackage: false}
	for _, key := range variabletypes.Keys() {
		v, err := variabletypes.New(key)
		require.NoError(t, err, key)

		data, err := json.Marshal(v)
		require.NoError(t, err, key)
		var schema struct {
			Type string `json:"type"`
		}
		require.NoError(t, json.Unmarshal(data, &schema), key)

		if schema.Type == "" {
			continue // image and file carry no type field
		}
		assert.Equal(t, key, schema.Type, "key should match what %s resolves to", resolver.Resolve(v))
	}
}

func TestNewIsCaseInsensitive(t *testing.T) {
	_, err := variabletypes.New("TRS.trs")
	assert.NoError(t, err)

	_, err = variabletypes.New("no.such.type")
	assert.ErrorContains(t, err, "no.such.type")
}
