package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/stretchr/testify/require"
)

func benchmarkSnapshot(b *testing.B, nodeCount int) {
	instance := testInstanceWithLiftedNodesB(b)
	for range nodeCount {
		if _, _, err := instance.CreateNode(floatSourceType); err != nil {
			b.Fatal(err)
		}
	}

	first, err := instance.EncodeToAppSchema()
	require.NoError(b, err)
	b.ReportMetric(float64(len(first))/1024, "KB/graph")

	b.ResetTimer()
	for range b.N {
		if _, err := instance.EncodeToAppSchema(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSnapshot100Nodes(b *testing.B)  { benchmarkSnapshot(b, 100) }
func BenchmarkSnapshot1000Nodes(b *testing.B) { benchmarkSnapshot(b, 1000) }
func BenchmarkSnapshot4000Nodes(b *testing.B) { benchmarkSnapshot(b, 4000) }

func testInstanceWithLiftedNodesB(b *testing.B) *graph.Instance {
	b.Helper()
	t := &testing.T{}
	return testInstanceWithLiftedNodes(t)
}
