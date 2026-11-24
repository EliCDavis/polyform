package nodes_test

import (
	"testing"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

var onlyFirstRuns int

type twoReadersNode struct {
	First  nodes.Output[float64]
	Second nodes.Output[float64]
}

func (n twoReadersNode) OnlyFirst(out *nodes.StructOutput[float64]) {
	onlyFirstRuns++
	out.Set(nodes.GetOutputValue(out, n.First))
}

func (n twoReadersNode) Both(out *nodes.StructOutput[float64]) {
	out.Set(nodes.GetOutputValue(out, n.First) + nodes.GetOutputValue(out, n.Second))
}

type twoValueNode struct{}

func (twoValueNode) Out(out *nodes.StructOutput[float64]) { out.Set(2) }

func TestAnInputAnOutputNeverReadDoesNotStaleItWhenFirstEvaluated(t *testing.T) {
	onlyFirstRuns = 0
	node := &nodes.Struct[twoReadersNode]{Data: twoReadersNode{
		First:  nodes.GetNodeOutputPort[float64](&nodes.Struct[twoValueNode]{}, "Out"),
		Second: nodes.GetNodeOutputPort[float64](&nodes.Struct[twoValueNode]{}, "Out"),
	}}
	onlyFirst := nodes.GetNodeOutputPort[float64](node, "Only First")
	assert.Equal(t, 2., onlyFirst.Value())
	version := onlyFirst.Version()

	assert.Equal(t, 4., nodes.GetNodeOutputPort[float64](node, "Both").Value())

	assert.Equal(t, 2., onlyFirst.Value())
	assert.Equal(t, version, onlyFirst.Version())
	assert.Equal(t, 1, onlyFirstRuns)
}
