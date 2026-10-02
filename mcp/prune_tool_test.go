package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPruneOrphansListsDeadChainsAndSparesProducersAndBoundaries(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	// A live part: sphere -> model -> manifest (producer).
	body := makeSphereModel(t, session, "1", "0")
	var manifest polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   gltfManifestType,
		"inputs": map[string]any{"Models": map[string]any{"elements": []any{map[string]any{"nodeId": body, "port": "Out"}}}},
	}, &manifest)
	callTool(t, session, "set_producer", map[string]any{"nodeId": manifest.NodeId, "port": "Out", "name": "out.glb"}, &polyformmcp.SetProducerOutput{})

	// A dead chain: literal -> sphere that nothing reads.
	var dead polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   sphereNodeType,
		"inputs": map[string]any{"Radius": map[string]any{"value": "0.3"}},
	}, &dead)

	// A subgraph whose output boundary must survive even though nothing
	// instantiates it yet, plus an orphan inside it.
	callTool(t, session, "create_subgraph", map[string]any{
		"id": "prunable", "name": "Prunable",
		"nodes": []map[string]any{
			{"alias": "s", "type": sphereNodeType},
			{"alias": "stray", "type": sphereNodeType},
		},
		"outputs": []map[string]any{{"name": "Mesh", "type": "github.com/EliCDavis/polyform/modeling.Mesh", "nodeId": "s", "port": "Out"}},
	}, &polyformmcp.CreateSubgraphOutput{})

	before := len(inst.NodeIds())
	var preview polyformmcp.PruneOrphansOutput
	callTool(t, session, "prune_orphans", map[string]any{"scope": "all"}, &preview)
	assert.False(t, preview.Applied)
	assert.Equal(t, before, len(inst.NodeIds()), "a preview deletes nothing")

	byScope := map[string][]string{}
	for _, o := range preview.Orphans {
		byScope[o.Scope] = append(byScope[o.Scope], o.Id)
	}
	assert.Len(t, byScope["root"], 2, "the dead sphere and its literal: %v", preview.Orphans)
	assert.Contains(t, byScope["root"], dead.NodeId)
	assert.Len(t, byScope["prunable"], 1, "only the stray sphere; the output boundary and what feeds it stay: %v", preview.Orphans)

	var applied polyformmcp.PruneOrphansOutput
	callTool(t, session, "prune_orphans", map[string]any{"scope": "all", "apply": true}, &applied)
	require.True(t, applied.Applied)
	assert.Equal(t, before-2, len(inst.NodeIds()))
	assert.False(t, inst.HasNodeWithId(dead.NodeId))
	assert.True(t, inst.HasNodeWithId(manifest.NodeId), "the producer's source survives")
	assert.True(t, inst.HasNodeWithId(body))

	var again polyformmcp.PruneOrphansOutput
	callTool(t, session, "prune_orphans", map[string]any{"scope": "all"}, &again)
	assert.Empty(t, again.Orphans, "a second pass finds nothing")
}

// An input port the interior has not wired up yet is still a declared part
// of the subgraph's interface. Pruning it drops every instance's
// connection into that port and desyncs the outline.
func TestPruneOrphansSparesAnUnusedInputBoundary(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var made polyformmcp.CreateSubgraphOutput
	callTool(t, session, "create_subgraph", map[string]any{
		"id": "ported", "name": "Ported",
		"inputs": []map[string]any{
			{"name": "Used", "type": "float64"},
			{"name": "Spare", "type": "float64"},
		},
		"nodes": []map[string]any{
			{"alias": "sphere", "type": sphereNodeType, "inputs": map[string]any{
				"Radius": map[string]any{"nodeId": "Used", "port": "Value"},
			}},
		},
		"outputs": []map[string]any{{"name": "Mesh", "nodeId": "sphere", "port": "Out"}},
	}, &made)
	require.Empty(t, made.Errors)
	spare := made.Inputs["Spare"]
	require.NotEmpty(t, spare)

	var preview polyformmcp.PruneOrphansOutput
	callTool(t, session, "prune_orphans", map[string]any{"scope": "all"}, &preview)
	for _, o := range preview.Orphans {
		assert.NotEqual(t, spare, o.Id, "an unwired input boundary is interface, not an orphan")
	}

	callTool(t, session, "prune_orphans", map[string]any{"scope": "all", "apply": true}, &polyformmcp.PruneOrphansOutput{})

	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "ported"}, &placed)
	assert.Contains(t, inst.Node(placed.NodeId).Inputs(), "Spare", "the port survives pruning")
}
