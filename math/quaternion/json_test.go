package quaternion_test

import (
	"encoding/json"
	"testing"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuaternionJSONRoundTrip(t *testing.T) {
	original := quaternion.FromTheta(1.2, vector3.New(0., 1., 0.))

	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.JSONEq(t, `{"x":0,"y":0.5646424733950354,"z":0,"w":0.8253356149096783}`, string(data))

	var back quaternion.Quaternion
	require.NoError(t, json.Unmarshal(data, &back))
	assert.Equal(t, original, back)
}

func TestQuaternionJSONMissingFieldsIsIdentity(t *testing.T) {
	var q quaternion.Quaternion
	require.NoError(t, json.Unmarshal([]byte(`{}`), &q))
	assert.Equal(t, quaternion.Identity(), q)
}
