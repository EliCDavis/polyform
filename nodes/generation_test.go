package nodes_test

import (
	"testing"
	"time"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type countingAdd struct {
	A     nodes.Output[float64]
	B     nodes.Output[float64]
	calls *int
}

func (c countingAdd) Sum(out *nodes.StructOutput[float64]) {
	*c.calls++
	out.Set(nodes.TryGetOutputValue(out, c.A, 0) + nodes.TryGetOutputValue(out, c.B, 0))
}

func sumOf(a, b nodes.Output[float64], calls *int) (nodes.Node, nodes.Output[float64]) {
	n := &nodes.Struct[countingAdd]{Data: countingAdd{A: a, B: b, calls: calls}}
	return n, n.Outputs()["Sum"].(nodes.Output[float64])
}

func TestMemoizedVersionsStillSeeEveryKindOfChange(t *testing.T) {
	leaf := nodes.NewValue(1.)
	leafOut := leaf.Outputs()["Value"].(nodes.Output[float64])
	other := nodes.NewValue(10.)
	otherOut := other.Outputs()["Value"].(nodes.Output[float64])

	calls := 0
	mid, midOut := sumOf(leafOut, leafOut, &calls)
	_, topOut := sumOf(midOut, midOut, &calls)

	require.InDelta(t, 4, topOut.Value(), 1e-9)
	require.Equal(t, 2, calls)
	topOut.Value()
	require.Equal(t, 2, calls, "an unchanged graph is served from cache")

	leaf.Set(2.)
	assert.InDelta(t, 8, topOut.Value(), 1e-9, "a parameter change reaches the top")
	assert.Equal(t, 4, calls)

	require.NoError(t, mid.Inputs()["B"].(nodes.SingleValueInputPort).Set(otherOut))
	assert.InDelta(t, 24, topOut.Value(), 1e-9, "a rewired input reaches the top")

	mid.Inputs()["B"].Clear()
	assert.InDelta(t, 4, topOut.Value(), 1e-9, "a cleared input reaches the top")
}

// A ladder of N nodes where each reads the one below it twice has 2^N
// paths to the leaf; the version check must not walk them all.
func TestVersionCheckIsLinearInDiamondLadders(t *testing.T) {
	leaf := nodes.NewValue(1.)
	out := leaf.Outputs()["Value"].(nodes.Output[float64])
	calls := 0
	for i := 0; i < 40; i++ {
		_, out = sumOf(out, out, &calls)
	}

	start := time.Now()
	out.Value()
	out.Version()
	out.Value()
	assert.Less(t, time.Since(start), 2*time.Second)
	assert.Equal(t, 40, calls)
}
