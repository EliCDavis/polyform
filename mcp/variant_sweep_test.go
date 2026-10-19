package mcp_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cubeExportWithWidthVariable(t *testing.T) sessionT {
	t.Helper()
	session := testSession(t)
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{
			{"path": "Width", "type": "float64", "value": "1"},
			{"path": "Height", "type": "float64", "value": "1"},
		},
	}, &polyformmcp.CreateVariablesOutput{})
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "cube", "type": cubeNodeType, "inputs": map[string]any{
			"Width": map[string]any{"variable": "Width"}, "Height": map[string]any{"variable": "Height"},
		}},
		map[string]any{"alias": "model", "type": gltfModelType, "inputs": map[string]any{
			"Mesh": map[string]any{"nodeId": "cube", "port": "Out"},
		}},
		map[string]any{"alias": "manifest", "type": gltfManifestType, "inputs": map[string]any{
			"Models": map[string]any{"nodeId": "model", "port": "Out"},
		}},
	}}, &created)
	require.Empty(t, created.Errors)
	callTool(t, session, "set_producer", map[string]any{"nodeId": created.Nodes["manifest"], "port": "Out", "name": "cube.glb"}, &polyformmcp.SetProducerOutput{})
	return session
}

func TestRunVariantSweepRestoresTheVariablesItChanged(t *testing.T) {
	session := cubeExportWithWidthVariable(t)
	callTool(t, session, "create_variant_set", map[string]any{
		"name":       "Widths",
		"dimensions": []map[string]any{{"path": "Width", "type": "discrete", "data": `{"values":[2,3]}`}},
	}, &polyformmcp.CreateVariantSetOutput{})

	var run polyformmcp.RunVariantSweepOutput
	callTool(t, session, "run_variant_sweep", map[string]any{"name": "Widths", "outputDir": t.TempDir()}, &run)
	assert.True(t, run.Restored)
	assert.EqualValues(t, 1, variableValue(t, session, "Width"))

	callTool(t, session, "run_variant_sweep", map[string]any{"name": "Widths", "outputDir": t.TempDir(), "keepLast": true}, &run)
	assert.False(t, run.Restored)
	assert.EqualValues(t, 3, variableValue(t, session, "Width"))
}

func TestRunVariantSweepReportsEachVariantAndWritesVariantsJson(t *testing.T) {
	session := cubeExportWithWidthVariable(t)
	callTool(t, session, "create_variant_set", map[string]any{
		"name":       "Widths",
		"dimensions": []map[string]any{{"path": "Width", "type": "discrete", "data": `{"values":[2,3]}`}},
	}, &polyformmcp.CreateVariantSetOutput{})

	outDir := t.TempDir()
	var run polyformmcp.RunVariantSweepOutput
	callTool(t, session, "run_variant_sweep", map[string]any{"name": "Widths", "outputDir": outDir}, &run)
	assert.Equal(t, 2, run.TotalCombinations)
	assert.Equal(t, 2, run.Ran)
	require.Len(t, run.Variants, 2)
	assert.Equal(t, 0, run.Variants[0].Index)
	assert.Equal(t, 1, run.Variants[1].Combination)
	assert.Equal(t, filepath.Join(outDir, "variant-0001"), run.Variants[1].Folder)
	assert.Equal(t, map[string]string{"Width": "3"}, run.Variants[1].Variables)
	assert.Equal(t, 12, run.Variants[1].Triangles)
	assert.Empty(t, run.Variants[1].NodeErrors)
	assert.Equal(t, run.Folders, []string{run.Variants[0].Folder, run.Variants[1].Folder})

	_, err := os.Stat(filepath.Join(outDir, "variant-0001", "cube.glb"))
	require.NoError(t, err)

	data, err := os.ReadFile(run.Manifest)
	require.NoError(t, err)
	var table struct {
		Set      string                      `json:"set"`
		Variants []polyformmcp.VariantResult `json:"variants"`
	}
	require.NoError(t, json.Unmarshal(data, &table))
	assert.Equal(t, "Widths", table.Set)
	assert.Equal(t, run.Variants, table.Variants)
}

func TestRunVariantSweepLimitAndSamplePickASubset(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_variable", map[string]any{"path": "Count", "type": "int", "value": "0"}, &polyformmcp.CreateVariableOutput{})
	callTool(t, session, "create_variant_set", map[string]any{
		"name":       "Counts",
		"dimensions": []map[string]any{{"path": "Count", "type": "intRange", "data": `{"min":0,"max":9,"samples":10}`}},
	}, &polyformmcp.CreateVariantSetOutput{})

	combinations := func(args map[string]any) ([]int, polyformmcp.RunVariantSweepOutput) {
		args["name"], args["outputDir"] = "Counts", t.TempDir()
		var run polyformmcp.RunVariantSweepOutput
		callTool(t, session, "run_variant_sweep", args, &run)
		picked := make([]int, len(run.Variants))
		for i, v := range run.Variants {
			picked[i] = v.Combination
		}
		return picked, run
	}

	first, run := combinations(map[string]any{"limit": 4})
	assert.Equal(t, []int{0, 1, 2, 3}, first)
	assert.Equal(t, 10, run.TotalCombinations)
	assert.Equal(t, 4, run.Ran)

	spaced, _ := combinations(map[string]any{"limit": 4, "sample": "spaced"})
	assert.Equal(t, []int{0, 3, 6, 9}, spaced)

	random, seeded := combinations(map[string]any{"limit": 4, "sample": "random", "seed": 7})
	again, _ := combinations(map[string]any{"limit": 4, "sample": "random", "seed": 7})
	assert.Len(t, random, 4)
	assert.Equal(t, random, again)
	assert.EqualValues(t, 7, seeded.Seed)

	_, unseeded := combinations(map[string]any{"limit": 2, "sample": "random"})
	assert.NotZero(t, unseeded.Seed, "the seed picked is reported so the run can be repeated")

	msg := callToolExpectingError(t, session, "run_variant_sweep", map[string]any{"name": "Counts", "outputDir": t.TempDir(), "sample": "stratified"})
	assert.Contains(t, msg, "stratified")

	msg = callToolExpectingError(t, session, "run_variant_sweep", map[string]any{"name": "Counts", "outputDir": t.TempDir(), "warnThreshold": 3})
	assert.Contains(t, msg, "limit")
	all, _ := combinations(map[string]any{"limit": 3, "warnThreshold": 3})
	assert.Len(t, all, 3, "the threshold is on what runs, not on the set's size")
}

func TestOneAtATimeSweepVariesEachControlFromTheBaseline(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{
			{"path": "Width", "type": "float64", "value": "1"},
			{"path": "Count", "type": "int", "value": "4"},
		},
	}, &polyformmcp.CreateVariablesOutput{})

	var created polyformmcp.CreateVariantSetOutput
	callTool(t, session, "create_variant_set", map[string]any{
		"name": "Extremes",
		"dimensions": []map[string]any{
			{"path": "Width", "type": "discrete", "data": `{"values":[0.5,2]}`},
			{"path": "Count", "type": "intRange", "data": `{"min":1,"max":9,"samples":3}`},
		},
	}, &created)
	assert.Equal(t, 6, created.TotalCombinations)
	assert.Equal(t, 5, created.OneAtATime, "counts add instead of multiplying")

	var run polyformmcp.RunVariantSweepOutput
	callTool(t, session, "run_variant_sweep", map[string]any{"name": "Extremes", "mode": "oneAtATime", "outputDir": t.TempDir()}, &run)
	assert.Equal(t, "oneAtATime", run.Mode)
	assert.Equal(t, 5, run.TotalCombinations)
	require.Len(t, run.Variants, 5)
	assert.Equal(t, map[string]string{"Width": "0.5"}, run.Variants[0].Variables)
	assert.Equal(t, map[string]string{"Count": "9"}, run.Variants[4].Variables)
	assert.EqualValues(t, 1, variableValue(t, session, "Width"), "restored after the sweep")

	var cross polyformmcp.RunVariantSweepOutput
	callTool(t, session, "run_variant_sweep", map[string]any{"name": "Extremes", "outputDir": t.TempDir()}, &cross)
	assert.Equal(t, "cross", cross.Mode)
	require.Len(t, cross.Variants, 6, "the same set sweeps every combination by default")
	assert.Equal(t, map[string]string{"Width": "2", "Count": "9"}, cross.Variants[5].Variables)

	var list polyformmcp.ListVariantSetsOutput
	callTool(t, session, "list_variant_sets", map[string]any{}, &list)
	assert.Equal(t, 6, list.VariantSets[0].TotalCombinations)
	assert.Equal(t, 5, list.VariantSets[0].OneAtATime)

	msg := callToolExpectingError(t, session, "run_variant_sweep", map[string]any{"name": "Extremes", "mode": "pairwise", "outputDir": t.TempDir()})
	assert.Contains(t, msg, "pairwise")
}

func TestCreateVariantSetRefusesALargeProductWithoutAllowLarge(t *testing.T) {
	session := testSession(t)
	dimensions := []map[string]any{
		{"path": "A", "type": "numericRange", "data": `{"min":0,"max":1,"samples":30}`},
		{"path": "B", "type": "numericRange", "data": `{"min":0,"max":1,"samples":30}`},
	}

	msg := callToolExpectingError(t, session, "create_variant_set", map[string]any{"name": "Grid", "dimensions": dimensions})
	assert.Contains(t, msg, "900")
	assert.Contains(t, msg, "allowLarge")

	var list polyformmcp.ListVariantSetsOutput
	callTool(t, session, "list_variant_sets", map[string]any{}, &list)
	assert.Empty(t, list.VariantSets)

	var created polyformmcp.CreateVariantSetOutput
	callTool(t, session, "create_variant_set", map[string]any{"name": "Grid", "dimensions": dimensions, "allowLarge": true}, &created)
	assert.Equal(t, 900, created.TotalCombinations)
}
