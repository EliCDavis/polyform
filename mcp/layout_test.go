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

func savedPositions(t *testing.T, path string, scope string) map[string][2]float64 {
	t.Helper()
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	var saved struct {
		Data struct {
			Metadata  map[string]any `json:"metadata"`
			SubGraphs map[string]struct {
				Metadata map[string]any `json:"metadata"`
			} `json:"subGraphs"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(raw, &saved))
	meta := saved.Data.Metadata
	if scope != "" {
		meta = saved.Data.SubGraphs[scope].Metadata
	}
	nodes, _ := meta["nodes"].(map[string]any)
	out := map[string][2]float64{}
	for id, m := range nodes {
		p, _ := m.(map[string]any)["position"].(map[string]any)
		x, xOk := p["x"].(float64)
		y, yOk := p["y"].(float64)
		if xOk && yOk {
			out[id] = [2]float64{x, y}
		}
	}
	return out
}

func TestSaveGraphLaysOutNodesThatWereNeverPlaced(t *testing.T) {
	session := testSession(t)

	callTool(t, session, "create_subgraph", map[string]any{
		"id": "part", "name": "Part",
		"inputs": []map[string]any{{"name": "Mesh", "type": meshPortType}},
		"nodes": []map[string]any{
			{"alias": "model", "type": modelNodeType, "inputs": map[string]any{"Mesh": map[string]any{"nodeId": "Mesh", "port": "Value"}}},
		},
		"outputs": []map[string]any{{"name": "Model", "nodeId": "model", "port": "Out"}},
	}, &polyformmcp.CreateSubgraphOutput{})

	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []map[string]any{
		{"alias": "cube", "type": cubeNodeType},
		{"alias": "part", "subgraphId": "part", "inputs": map[string]any{"Mesh": map[string]any{"nodeId": "cube", "port": "Out"}}},
		{"alias": "manifest", "type": manifestNodeType, "inputs": map[string]any{
			"Models": map[string]any{"elements": []map[string]any{{"nodeId": "part", "port": "Model"}}},
		}},
	}}, &created)

	path := filepath.Join(t.TempDir(), "graph.json")
	callTool(t, session, "save_graph", map[string]any{"path": path}, nil)

	root := savedPositions(t, path, "")
	require.Len(t, root, 3, "every root node gets a position")
	cube, part, manifest := root[created.Nodes["cube"]], root[created.Nodes["part"]], root[created.Nodes["manifest"]]
	assert.Less(t, cube[0], part[0], "a node sits to the right of what feeds it")
	assert.Less(t, part[0], manifest[0])

	inside := savedPositions(t, path, "part")
	assert.Len(t, inside, 3, "boundary nodes and the interior are placed too")
}

func TestSaveGraphKeepsPositionsItAlreadyHas(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []map[string]any{
		{"alias": "cube", "type": cubeNodeType},
	}}, &created)
	inst.SetMetadata("nodes."+created.Nodes["cube"]+".position", map[string]any{"x": 1234., "y": 56.})

	path := filepath.Join(t.TempDir(), "graph.json")
	callTool(t, session, "save_graph", map[string]any{"path": path}, nil)

	assert.Equal(t, [2]float64{1234, 56}, savedPositions(t, path, "")[created.Nodes["cube"]])
}
