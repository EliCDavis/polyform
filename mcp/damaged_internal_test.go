package mcp

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/graph/history"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAFailedRollbackFreezesTheAutosaveUntilAGraphIsLoaded(t *testing.T) {
	dir := t.TempDir()
	s := NewServer(graph.New(graph.Config{TypeFactory: generator.Types(), VariableFactory: NewTypedVariable}))
	s.projectDir = dir
	autosave := filepath.Join(dir, autosaveFileName)

	call := func(tool string, f func() error) error {
		s.currentTool = tool
		var err error
		s.atomic(&err, f)
		return err
	}

	require.NoError(t, call("create_variable", func() error { return nil }))
	good, err := os.ReadFile(autosave)
	require.NoError(t, err)

	err = call("create_nodes", func() error { return fmt.Errorf("rebuild: %w", history.ErrRollbackFailed) })
	require.ErrorIs(t, err, history.ErrRollbackFailed)
	assert.Contains(t, err.Error(), "load_graph")

	ran := false
	err = call("connect_nodes", func() error { ran = true; return nil })
	require.Error(t, err)
	assert.False(t, ran, "nothing may run against a graph left half restored")

	err = call("save_graph", func() error { ran = true; return nil })
	require.Error(t, err)
	assert.False(t, ran, "a save would overwrite a good file with the damaged graph")

	after, err := os.ReadFile(autosave)
	require.NoError(t, err)
	assert.Equal(t, good, after)

	require.NoError(t, call("load_graph", func() error { return nil }))
	require.NoError(t, call("connect_nodes", func() error { ran = true; return nil }))
	assert.True(t, ran)
}
