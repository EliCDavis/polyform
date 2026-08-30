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

func TestInstance_VariantSet_SaveAndLoad(t *testing.T) {
	instance := graph.New(graph.Config{TypeFactory: &refutil.TypeFactory{}})

	err := instance.SetVariantSet("sweep", variant.Set{
		Dimensions: []variant.Dimension{
			variant.NewNumericRange("Scale", 0, 10, 5),
		},
	})
	require.NoError(t, err)

	set, err := instance.VariantSet("sweep")
	require.NoError(t, err)
	require.Len(t, set.Dimensions, 1)
	assert.Equal(t, "Scale", set.Dimensions[0].Path())
	assert.Equal(t, 5, set.Dimensions[0].Count())
}

func TestInstance_VariantSet_LoadMissingIsAnError(t *testing.T) {
	instance := graph.New(graph.Config{TypeFactory: &refutil.TypeFactory{}})
	_, err := instance.VariantSet("does-not-exist")
	assert.Error(t, err)
}

func TestInstance_VariantSets_ListsSortedNames(t *testing.T) {
	instance := graph.New(graph.Config{TypeFactory: &refutil.TypeFactory{}})

	require.NoError(t, instance.SetVariantSet("zebra", variant.Set{}))
	require.NoError(t, instance.SetVariantSet("apple", variant.Set{}))

	assert.Equal(t, []string{"apple", "zebra"}, instance.VariantSets())
}

func TestInstance_RenameVariantSet(t *testing.T) {
	instance := graph.New(graph.Config{TypeFactory: &refutil.TypeFactory{}})
	require.NoError(t, instance.SetVariantSet("old-name", variant.Set{}))

	require.NoError(t, instance.RenameVariantSet("old-name", "new-name"))
	assert.Equal(t, []string{"new-name"}, instance.VariantSets())

	assert.Error(t, instance.RenameVariantSet("old-name", "another-name"), "renaming a set that no longer exists should fail")
}

func TestInstance_RenameVariantSet_RejectsExistingTarget(t *testing.T) {
	instance := graph.New(graph.Config{TypeFactory: &refutil.TypeFactory{}})
	require.NoError(t, instance.SetVariantSet("a", variant.Set{}))
	require.NoError(t, instance.SetVariantSet("b", variant.Set{}))

	assert.Error(t, instance.RenameVariantSet("a", "b"))
}

func TestInstance_DeleteVariantSet(t *testing.T) {
	instance := graph.New(graph.Config{TypeFactory: &refutil.TypeFactory{}})
	require.NoError(t, instance.SetVariantSet("temp", variant.Set{}))

	require.NoError(t, instance.DeleteVariantSet("temp"))
	assert.Empty(t, instance.VariantSets())

	assert.Error(t, instance.DeleteVariantSet("temp"), "deleting an already-deleted set should fail")
}

func TestInstance_VariantSet_SurvivesEncodeAndApplyAppSchema(t *testing.T) {
	source := graph.New(graph.Config{TypeFactory: &refutil.TypeFactory{}})
	require.NoError(t, source.SetVariantSet("sweep", variant.Set{
		Dimensions: []variant.Dimension{
			variant.NewNumericRange("Scale", 0, 10, 5),
			variant.NewVector3Range("Position", 0, 1, 0, 1, 0, 1, 2),
			variant.NewDiscrete("Fur", rawFloatForTest(t, 1), rawFloatForTest(t, 2)),
		},
	}))

	payload, err := source.EncodeToAppSchema()
	require.NoError(t, err)

	restored := graph.New(graph.Config{TypeFactory: &refutil.TypeFactory{}})
	require.NoError(t, restored.ApplyAppSchema(payload))

	assert.Equal(t, []string{"sweep"}, restored.VariantSets())

	set, err := restored.VariantSet("sweep")
	require.NoError(t, err)
	require.Len(t, set.Dimensions, 3)
	assert.Equal(t, 5*2*2, set.TotalCombinations(), "5 scale samples * 2 position samples (whole-vector lerp, not per-axis) * 2 fur values")
}
