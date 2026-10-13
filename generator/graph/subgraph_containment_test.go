package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func definitions(t *testing.T, ids ...string) *graph.Instance {
	t.Helper()
	inst := importTestInstance(t, "Defs")
	for _, id := range ids {
		readsVariable(t, inst, id)
	}
	return inst
}

func definition(t *testing.T, inst *graph.Instance, id string) *graph.Graph {
	t.Helper()
	def, err := inst.SubGraphInstance(id)
	require.NoError(t, err)
	return def
}

func refusesPlacing(t *testing.T, inst *graph.Instance, scope, placed string) {
	t.Helper()
	def := definition(t, inst, scope)
	before := len(def.Schema().Nodes)
	_, _, err := def.CreateNode(subgraph.RuntimeTypePath(placed))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "contain itself")
	assert.Len(t, def.Schema().Nodes, before)
}

func TestASubGraphCannotPlaceItself(t *testing.T) {
	inst := definitions(t, "a")
	refusesPlacing(t, inst, "a", "a")
}

func TestTwoSubGraphsCannotPlaceEachOther(t *testing.T) {
	inst := definitions(t, "a", "b")
	place(t, definition(t, inst, "a"), "b")
	refusesPlacing(t, inst, "b", "a")
}

func TestAPlacementCycleThroughThreeSubGraphsIsRefused(t *testing.T) {
	inst := definitions(t, "a", "b", "c")
	place(t, definition(t, inst, "a"), "b")
	place(t, definition(t, inst, "b"), "c")
	refusesPlacing(t, inst, "c", "a")
}

func TestReplacingAPlacementCannotCloseACycle(t *testing.T) {
	inst := definitions(t, "a", "b", "x")
	place(t, definition(t, inst, "a"), "b")
	_, xInB := place(t, definition(t, inst, "b"), "x")

	_, err := definition(t, inst, "b").ReplaceSubgraphInstance(xInB, "a")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "contain itself")
}

func TestSharedPlacementsThatDoNotLoopAreAllowed(t *testing.T) {
	inst := definitions(t, "leaf", "left", "right", "top")
	place(t, definition(t, inst, "left"), "leaf")
	place(t, definition(t, inst, "right"), "leaf")
	place(t, definition(t, inst, "top"), "left")
	place(t, definition(t, inst, "top"), "right")
	place(t, definition(t, inst, "top"), "leaf")
	place(t, inst, "top")
	place(t, inst, "top")
}

func TestARemovedPlacementNoLongerBlocksTheReverse(t *testing.T) {
	inst := definitions(t, "a", "b")
	_, bInA := place(t, definition(t, inst, "a"), "b")
	refusesPlacing(t, inst, "b", "a")

	require.NoError(t, errOf(definition(t, inst, "a").DeleteNodeById(bInA)))
	place(t, definition(t, inst, "b"), "a")
}
