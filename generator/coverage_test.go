package generator_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeCoverageCountsTypesInNestedGraphsAndSkipsAutosave(t *testing.T) {
	types := &refutil.TypeFactory{}
	refutil.RegisterType[parameter.Value[float64]](types)
	refutil.RegisterType[parameter.Value[int]](types)
	refutil.RegisterType[parameter.Value[bool]](types)
	keys := types.Types()

	root := t.TempDir()
	write := func(rel, content string) {
		path := filepath.Join(root, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	}
	floatKey, intKey := "", ""
	for _, k := range keys {
		switch {
		case strings.HasSuffix(k, "[float64]"):
			floatKey = k
		case strings.HasSuffix(k, "[int]"):
			intKey = k
		}
	}

	write("a/graph.json", `{"data":{"subGraphs":{"s":{"nodes":{"n":{"type":"`+floatKey+`"}}}}}}`)
	write("a/autosave.json", `{"data":{"nodes":{"n":{"type":"`+intKey+`"}}}}`)
	write("b/notes.json", `{"hello":"world"}`)

	report, err := generator.NodeCoverage(types, root)
	require.NoError(t, err)

	assert.Equal(t, []string{filepath.Join("a", "graph.json")}, report.Graphs)
	assert.Len(t, report.Usage[floatKey], 1)
	assert.Empty(t, report.Usage[intKey])

	var out bytes.Buffer
	require.NoError(t, report.WriteMarkdown(&out))
	assert.Contains(t, out.String(), "1 of 3 node types used across 1 graphs.")
}

func TestRankAdditionsPicksGreedilyAndSkipsCoveredTypes(t *testing.T) {
	report := generator.CoverageReport{
		Graphs: []string{"a", "b", "c"},
		Usage: map[string][]string{
			"x": {"a", "b"},
			"y": {"a"},
			"z": {"b"},
			"w": {"b", "c"},
			"v": {"c"},
		},
	}

	picks := report.RankAdditions(map[string]bool{"v": true}, 10)

	assert.Equal(t, []generator.CoverageAddition{
		{Graph: "b", Adds: 3},
		{Graph: "a", Adds: 1},
	}, picks)
}
