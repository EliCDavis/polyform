package graph_test

import (
	"encoding/json"
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/variant"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func rawFloatForTest(t *testing.T, v float64) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	return data
}

func TestInstance_VariantSet_SurvivesEncodeAndApplyAppSchema(t *testing.T) {
	source := graph.New(graph.Config{TypeFactory: &refutil.TypeFactory{}})
	source.VariantSets().Set("sweep", variant.Set{
		Dimensions: []variant.Dimension{
			variant.NewNumericRange("Scale", 0, 10, 5),
			variant.NewVector3Range("Position", 0, 1, 0, 1, 0, 1, 2),
			variant.NewDiscrete("Fur", rawFloatForTest(t, 1), rawFloatForTest(t, 2)),
		},
	})

	payload, err := source.EncodeToAppSchema()
	require.NoError(t, err)

	restored := graph.New(graph.Config{TypeFactory: &refutil.TypeFactory{}})
	require.NoError(t, restored.ApplyAppSchema(payload))

	assert.Equal(t, []string{"sweep"}, restored.VariantSets().Names())

	set, err := restored.VariantSets().Get("sweep")
	require.NoError(t, err)
	require.Len(t, set.Dimensions, 3)
	assert.Equal(t, 5*2*2, set.TotalCombinations(), "5 scale samples * 2 position samples (whole-vector lerp, not per-axis) * 2 fur values")
}
