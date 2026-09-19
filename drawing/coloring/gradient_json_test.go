package coloring_test

import (
	"encoding/json"
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGradientColorJSONRoundTrip(t *testing.T) {
	original := coloring.NewGradientColor(
		coloring.GradientKey[coloring.Color]{Time: 0, Value: coloring.Color{R: 1, A: 1}},
		coloring.GradientKey[coloring.Color]{Time: 10, Value: coloring.Color{B: 1, A: 1}},
	)

	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.JSONEq(t, `{"keys":[{"time":0,"value":"#ff0000"},{"time":1,"value":"#0000ff"}]}`, string(data))

	var back coloring.Gradient[coloring.Color]
	require.NoError(t, json.Unmarshal(data, &back))
	assert.Equal(t, original.Sample(0.25), back.Sample(0.25))
	assert.Equal(t, original.Keys(), back.Keys())
}

func TestGradient3DJSONRoundTrip(t *testing.T) {
	original := coloring.NewGradient3D(
		coloring.GradientKey[vector3.Float64]{Time: 0, Value: vector3.Zero[float64]()},
		coloring.GradientKey[vector3.Float64]{Time: 1, Value: vector3.One[float64]()},
	)

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var back coloring.Gradient[vector3.Float64]
	require.NoError(t, json.Unmarshal(data, &back))
	assert.Equal(t, vector3.Fill(0.5), back.Sample(0.5))
}

func TestGradientJSONEmptyKeys(t *testing.T) {
	var g coloring.Gradient[coloring.Color]
	require.NoError(t, json.Unmarshal([]byte(`{"keys":[]}`), &g))
	assert.Empty(t, g.Keys())
	assert.Equal(t, coloring.Color{}, g.Sample(0.5))
}
