package trs_test

import (
	"encoding/json"
	"testing"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTRSJSONRoundTrip(t *testing.T) {
	original := trs.New(
		vector3.New(1., 2., 3.),
		quaternion.FromTheta(0.5, vector3.New(1., 0., 0.)),
		vector3.New(2., 2., 2.),
	)

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var back trs.TRS
	require.NoError(t, json.Unmarshal(data, &back))
	assert.Equal(t, original, back)
}

func TestTRSJSONMissingFieldsIsIdentity(t *testing.T) {
	var parsed trs.TRS
	require.NoError(t, json.Unmarshal([]byte(`{}`), &parsed))
	assert.Equal(t, trs.Identity(), parsed)

	require.NoError(t, json.Unmarshal([]byte(`{"position":{"x":5}}`), &parsed))
	assert.Equal(t, trs.Position(vector3.New(5., 0., 0.)), parsed)
}

func TestTRSArrayJSON(t *testing.T) {
	data, err := json.Marshal([]trs.TRS{trs.Identity(), trs.Scale(vector3.New(3., 3., 3.))})
	require.NoError(t, err)

	var back []trs.TRS
	require.NoError(t, json.Unmarshal(data, &back))
	require.Len(t, back, 2)
	assert.Equal(t, vector3.New(3., 3., 3.), back[1].Scale())
}
