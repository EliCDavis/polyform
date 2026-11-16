package logic_test

import (
	"testing"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/logic"
	"github.com/stretchr/testify/assert"
)

func or[T any](values ...nodes.LiftedPort[bool]) T {
	return nodes.GetNodeOutputPort[T](&nodes.Struct[logic.OrNode]{Data: logic.OrNode{Values: values}}, "Out").Value()
}

func TestOrOfSingleValues(t *testing.T) {
	yes := nodes.ConstOutput[bool]{Val: true}
	no := nodes.ConstOutput[bool]{Val: false}

	assert.True(t, or[bool](no, yes))
	assert.False(t, or[bool](no, no))
	assert.False(t, or[bool](), "nothing holds")
}

func TestOrMatchesArraysElementByElement(t *testing.T) {
	assert.Equal(t, []bool{true, true, false},
		or[[]bool](
			nodes.ConstOutput[[]bool]{Val: []bool{true, false, false}},
			nodes.ConstOutput[[]bool]{Val: []bool{false, true, false}},
		))
}
