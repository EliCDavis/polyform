package mcp_test

import (
	"path/filepath"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportSubgraphsBringsInAPartFromAnotherSavedGraph(t *testing.T) {
	path := filepath.Join(t.TempDir(), "part.json")
	source := testSession(t)
	scaledSubgraph(t, source, "tail")
	callTool(t, source, "save_graph", map[string]any{"path": path}, &polyformmcp.SaveGraphOutput{})

	session, inst := testSessionWithInstance(t)
	scaledSubgraph(t, session, "tail")

	var out polyformmcp.ImportSubgraphsOutput
	callTool(t, session, "import_subgraphs", map[string]any{"path": path, "group": "Parts"}, &out)
	assert.Equal(t, "Parts", out.Group)
	require.Len(t, out.Subgraphs, 1)
	imported := out.Subgraphs[0]
	assert.Equal(t, "tail_2", imported.Id)
	assert.Equal(t, "tail", imported.OriginalId)
	assert.Contains(t, imported.Inputs, "Base")
	assert.Contains(t, imported.Inputs, "Scale")
	assert.Contains(t, imported.Outputs, "Result")

	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": imported.Id,
		"inputs":     map[string]any{"Base": map[string]any{"value": "2"}, "Scale": map[string]any{"value": "5"}},
	}, &placed)
	assert.InDelta(t, 10, evalFloat64Output(t, inst, placed.NodeId, "Result"), 1e-9)
}

func TestImportSubgraphsListsPlacedSubgraphsAsNested(t *testing.T) {
	path := filepath.Join(t.TempDir(), "part.json")
	source := testSession(t)
	scaledSubgraph(t, source, "inner")
	callTool(t, source, "create_subgraph", map[string]any{"id": "outer", "name": "Outer"}, &polyformmcp.CreateSubgraphOutput{})
	callTool(t, source, "instantiate_subgraph", map[string]any{"subgraphId": "inner", "scope": "outer"}, &polyformmcp.InstantiateSubgraphOutput{})
	callTool(t, source, "save_graph", map[string]any{"path": path}, &polyformmcp.SaveGraphOutput{})

	session := testSession(t)
	var out polyformmcp.ImportSubgraphsOutput
	callTool(t, session, "import_subgraphs", map[string]any{"path": path}, &out)
	require.Len(t, out.Subgraphs, 1)
	assert.Equal(t, "outer", out.Subgraphs[0].Id)
	assert.Equal(t, []string{"inner"}, out.Nested)
}

func TestImportSubgraphsRejectsAMissingFile(t *testing.T) {
	session := testSession(t)
	callToolExpectingError(t, session, "import_subgraphs", map[string]any{"path": filepath.Join(t.TempDir(), "nope.json")})
}
