package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func subgraphTypeIn(t *testing.T, inst *graph.Instance, name string) string {
	t.Helper()
	require.NoError(t, inst.CreateSubGraph(name, name, ""))
	path, err := inst.RegisterSubGraphNodeType(name)
	require.NoError(t, err)
	return path
}

func knowsType(inst *graph.Instance, typePath string) bool {
	for _, nt := range inst.BuildSchemaForAllNodeTypes() {
		if nt.Type == typePath {
			return true
		}
	}
	return false
}

func TestRegisteringASubgraphTypeDoesNotTouchTheCallersFactory(t *testing.T) {
	shared := &refutil.TypeFactory{}
	before := len(shared.Types())

	instance := graph.New(graph.Config{TypeFactory: shared})
	typePath := subgraphTypeIn(t, instance, "widget")

	assert.NotContains(t, shared.Types(), typePath, "the runtime type leaked into the shared registry")
	assert.Equal(t, before, len(shared.Types()), "the shared registry grew")
}

func TestASecondGraphDoesNotInheritTheFirstsSubgraphTypes(t *testing.T) {
	shared := &refutil.TypeFactory{}

	first := graph.New(graph.Config{TypeFactory: shared})
	typePath := subgraphTypeIn(t, first, "widget")
	assert.True(t, knowsType(first, typePath), "the graph that made it should know it")

	second := graph.New(graph.Config{TypeFactory: shared})
	assert.False(t, knowsType(second, typePath),
		"a fresh graph should not carry a subgraph built in another one")
}
