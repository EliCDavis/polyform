package mcp_test

import (
	"fmt"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/require"
)

func TestCreateFlushPositionSubgraph(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var created polyformmcp.CreateFlushPositionSubgraphOutput
	callTool(t, session, "create_flush_position_subgraph", map[string]any{
		"id": "flush",
	}, &created)
	require.Equal(t, "flush", created.SubgraphId)
	require.Equal(t, []string{"A Position", "A Half Size", "B Half Size", "Direction"}, created.Inputs)
	require.Equal(t, "Position", created.Output)

	wire := func(direction float64) float64 {
		var instantiated polyformmcp.InstantiateSubgraphOutput
		callTool(t, session, "instantiate_subgraph", map[string]any{
			"subgraphId": "flush",
			"inputs": map[string]any{
				"A Position":  map[string]any{"value": "5"},
				"A Half Size": map[string]any{"value": "2"},
				"B Half Size": map[string]any{"value": "1"},
				"Direction":   map[string]any{"value": fmt.Sprintf("%v", direction)},
			},
		}, &instantiated)
		return evalFloat64Output(t, inst, instantiated.NodeId, "Position")
	}

	require.InDelta(t, 8.0, wire(1), 1e-9, "positive side: A.pos + (A.half + B.half)")
	require.InDelta(t, 2.0, wire(-1), 1e-9, "negative side: A.pos - (A.half + B.half)")
}
