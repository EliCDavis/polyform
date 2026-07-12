package subgraph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/stretchr/testify/assert"
)

func TestIsRuntimeNodeType(t *testing.T) {
	assert.True(t, subgraph.IsRuntimeNodeType("subgraph/my-graph"))
	assert.False(t, subgraph.IsRuntimeNodeType("subgraph"))
	assert.False(t, subgraph.IsRuntimeNodeType("github.com/foo/Bar"))
	assert.False(t, subgraph.IsRuntimeNodeType(""))
}

func TestRuntimeTypeID(t *testing.T) {
	assert.Equal(t, "my-graph", subgraph.RuntimeTypeID("subgraph/my-graph"))
	assert.Equal(t, "nested/id", subgraph.RuntimeTypeID("subgraph/nested/id"))
	assert.Equal(t, "", subgraph.RuntimeTypeID("not-a-runtime"))
}

func TestRuntimeTypePath(t *testing.T) {
	assert.Equal(t, "subgraph/abc", subgraph.RuntimeTypePath("abc"))
}

func TestRuntimeTypePathAndIDRoundtrip(t *testing.T) {
	id := "my-sub-graph"
	path := subgraph.RuntimeTypePath(id)
	assert.True(t, subgraph.IsRuntimeNodeType(path))
	assert.Equal(t, id, subgraph.RuntimeTypeID(path))
}
