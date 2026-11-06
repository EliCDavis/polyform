package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/registry"
	"github.com/EliCDavis/polyform/generator/variable/variabletypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBundledExamplesLoad(t *testing.T) {
	files, err := filepath.Glob("../../generator/edit/examples/*.json")
	require.NoError(t, err)
	require.NotEmpty(t, files)

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			data, err := os.ReadFile(file)
			require.NoError(t, err)
			g := graph.New(graph.Config{TypeFactory: registry.Types(), VariableFactory: variabletypes.New, Casts: casts()})
			assert.NoError(t, g.ApplyAppSchema(data))
		})
	}
}
