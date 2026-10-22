package mcp_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/graph"
	polyformmcp "github.com/EliCDavis/polyform/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newServer() *polyformmcp.Server {
	return polyformmcp.NewServer(graph.New(graph.Config{TypeFactory: generator.Types(), VariableFactory: polyformmcp.NewTypedVariable}))
}

func connect(t *testing.T, server *polyformmcp.Server) *mcpsdk.ClientSession {
	t.Helper()
	clientTransport, serverTransport := mcpsdk.NewInMemoryTransports()
	ctx := context.Background()
	_, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	session, err := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test", Version: "0"}, nil).Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { session.Close() })
	return session
}

func TestReplayRebuildsTheGraphUpToACallAndRunsIt(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "calls.jsonl")
	original := newServer()
	closeLog, err := original.EnableCallLog(logPath)
	require.NoError(t, err)
	session := connect(t, original)

	var first, second polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{map[string]any{"alias": "a", "type": cubeNodeType}}}, &first)
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{map[string]any{"alias": "b", "type": cubeNodeType}}}, &second)
	_, _ = session.CallTool(context.Background(), &mcpsdk.CallToolParams{Name: "render_preview", Arguments: map[string]any{"nodeId": first.Nodes["a"], "outputPath": filepath.Join(dir, "r.png")}})
	callTool(t, session, "delete_node", map[string]any{"nodeId": first.Nodes["a"]}, nil)
	require.NoError(t, closeLog())
	require.NoError(t, os.Remove(filepath.Join(dir, "r.png")))

	before := filepath.Join(dir, "before.json")
	report, err := polyformmcp.Replay(context.Background(), newServer(), logPath, 4, before)
	require.NoError(t, err)

	assert.Equal(t, 3, report.Replayed, "renders replay too, since they evaluate the graph")
	assert.Zero(t, report.Skipped)
	assert.NoFileExists(t, filepath.Join(dir, "r.png"), "a replayed render writes somewhere else")
	assert.Empty(t, report.Divergences)
	require.NotNil(t, report.Target)
	assert.Equal(t, "delete_node", report.Target.Tool)
	assert.Empty(t, report.Target.Error)

	saved, err := os.ReadFile(before)
	require.NoError(t, err)
	var doc map[string]any
	require.NoError(t, json.Unmarshal(saved, &doc))
	assert.Contains(t, string(saved), first.Nodes["a"], "the saved graph is the state before the target ran")
	assert.Contains(t, string(saved), second.Nodes["b"])
}
