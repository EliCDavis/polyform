package nodes_test

import (
	"testing"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

var negateRuns, forwardRuns int

type countedNegateNode struct {
	In nodes.LiftedPort[float64]
}

func (n countedNegateNode) Out(out *nodes.Lifted[float64]) {
	negateRuns++
	nodes.Zip1(out, n.In, func(v float64) float64 { return -v })
}

type countedForwardNode struct {
	In nodes.DynamicPort[pickType]
}

func (n countedForwardNode) Out(out *nodes.Dynamic[pickType]) {
	forwardRuns++
	out.Forward(n.In)
}

func TestALiftedReportDoesNotRunTheNode(t *testing.T) {
	negateRuns = 0
	node := &nodes.Struct[countedNegateNode]{}
	scalars(t, node, "In", 2)
	out := node.Outputs()["Out"]

	assert.Empty(t, out.(nodes.ObservableExecution).ExecutionReport().Steps)
	assert.Equal(t, 0, negateRuns)

	assert.Equal(t, -2., out.(nodes.Output[float64]).Value())
	assert.NotNil(t, out.(nodes.ObservableExecution).ExecutionReport().SelfTime)
	assert.Equal(t, 1, negateRuns)
}

func TestADynamicReportDoesNotRunTheNode(t *testing.T) {
	forwardRuns = 0
	node := &nodes.Struct[countedForwardNode]{}
	connect(t, node, "In", nodes.ConstOutput[string]{Val: "abc"})
	out := node.Outputs()["Out"]

	out.(nodes.ObservableExecution).ExecutionReport()
	assert.Equal(t, 0, forwardRuns)

	assert.Equal(t, "abc", out.(nodes.Output[string]).Value())
	assert.Equal(t, 1, forwardRuns)
}
