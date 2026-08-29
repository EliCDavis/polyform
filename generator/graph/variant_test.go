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

	err := instance.SaveVariantSet("sweep", map[string]variant.Dimension{
		"Scale": variant.NewNumericRange("Scale", 0, 10, 5),
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

	require.NoError(t, instance.SaveVariantSet("zebra", map[string]variant.Dimension{}))
	require.NoError(t, instance.SaveVariantSet("apple", map[string]variant.Dimension{}))

	assert.Equal(t, []string{"apple", "zebra"}, instance.VariantSets())
}

func TestInstance_RenameVariantSet(t *testing.T) {
	instance := graph.New(graph.Config{TypeFactory: &refutil.TypeFactory{}})
	require.NoError(t, instance.SaveVariantSet("old-name", map[string]variant.Dimension{}))

	require.NoError(t, instance.RenameVariantSet("old-name", "new-name"))
	assert.Equal(t, []string{"new-name"}, instance.VariantSets())

	assert.Error(t, instance.RenameVariantSet("old-name", "another-name"), "renaming a set that no longer exists should fail")
}

func TestInstance_RenameVariantSet_RejectsExistingTarget(t *testing.T) {
	instance := graph.New(graph.Config{TypeFactory: &refutil.TypeFactory{}})
	require.NoError(t, instance.SaveVariantSet("a", map[string]variant.Dimension{}))
	require.NoError(t, instance.SaveVariantSet("b", map[string]variant.Dimension{}))

	assert.Error(t, instance.RenameVariantSet("a", "b"))
}

func TestInstance_DeleteVariantSet(t *testing.T) {
	instance := graph.New(graph.Config{TypeFactory: &refutil.TypeFactory{}})
	require.NoError(t, instance.SaveVariantSet("temp", map[string]variant.Dimension{}))

	require.NoError(t, instance.DeleteVariantSet("temp"))
	assert.Empty(t, instance.VariantSets())

	assert.Error(t, instance.DeleteVariantSet("temp"), "deleting an already-deleted set should fail")
}

func TestInstance_VariantSet_SurvivesEncodeAndApplyAppSchema(t *testing.T) {
	source := graph.New(graph.Config{TypeFactory: &refutil.TypeFactory{}})
	require.NoError(t, source.SaveVariantSet("sweep", map[string]variant.Dimension{
		"Scale":    variant.NewNumericRange("Scale", 0, 10, 5),
		"Position": variant.NewVector3Range("Position", 0, 1, 2, 0, 1, 2, 0, 1, 2),
		"Fur":      variant.NewDiscrete("Fur", rawFloatForTest(t, 1), rawFloatForTest(t, 2)),
	}))

	payload, err := source.EncodeToAppSchema()
	require.NoError(t, err)

	restored := graph.New(graph.Config{TypeFactory: &refutil.TypeFactory{}})
	require.NoError(t, restored.ApplyAppSchema(payload))

	assert.Equal(t, []string{"sweep"}, restored.VariantSets())

	set, err := restored.VariantSet("sweep")
	require.NoError(t, err)
	require.Len(t, set.Dimensions, 3)
	assert.Equal(t, 5*8*2, set.TotalCombinations())
}
