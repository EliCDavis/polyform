package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/graph"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func dereferenceNothingInATool() int {
	var nothing *int
	return *nothing
}

func TestCallLogRecordsTheStackOfAPanic(t *testing.T) {
	s := NewServer(graph.New(graph.Config{TypeFactory: generator.Types(), VariableFactory: NewTypedVariable}))
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{Name: "probe_panic"}, func(ctx context.Context, req *mcpsdk.CallToolRequest, in struct{}) (*mcpsdk.CallToolResult, struct{}, error) {
		var err error
		s.atomic(&err, func() error {
			dereferenceNothingInATool()
			return nil
		})
		return nil, struct{}{}, err
	})
	logPath := filepath.Join(t.TempDir(), "calls.jsonl")
	closeLog, err := s.EnableCallLog(logPath)
	require.NoError(t, err)
	defer closeLog()

	clientTransport, serverTransport := mcpsdk.NewInMemoryTransports()
	ctx := context.Background()
	_, err = s.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	session, err := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test", Version: "0"}, nil).Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer session.Close()

	res, err := session.CallTool(ctx, &mcpsdk.CallToolParams{Name: "probe_panic", Arguments: map[string]any{}})
	require.NoError(t, err)
	require.True(t, res.IsError)

	data, err := os.ReadFile(logPath)
	require.NoError(t, err)
	var entry callLogEntry
	require.NoError(t, json.Unmarshal(data, &entry))
	assert.Equal(t, "runtime error: invalid memory address or nil pointer dereference", entry.Error)
	assert.Contains(t, entry.PanicStack, "dereferenceNothingInATool")
}
