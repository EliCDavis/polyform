package logic_test

import (
	"testing"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/logic"
	"github.com/stretchr/testify/assert"
)

func and[T any](values ...nodes.LiftedPort[bool]) T {
	return nodes.GetNodeOutputPort[T](&nodes.Struct[logic.AndNode]{Data: logic.AndNode{Values: values}}, "Out").Value()
}

func TestAndOfSingleValues(t *testing.T) {
	yes := nodes.ConstOutput[bool]{Val: true}
	no := nodes.ConstOutput[bool]{Val: false}

	assert.True(t, and[bool](yes, yes))
	assert.False(t, and[bool](yes, no))
	assert.True(t, and[bool](), "nothing to contradict")
}

func TestAndMatchesArraysElementByElement(t *testing.T) {
	assert.Equal(t, []bool{true, false, false},
		and[[]bool](
			nodes.ConstOutput[[]bool]{Val: []bool{true, true, false}},
			nodes.ConstOutput[[]bool]{Val: []bool{true, false, true}},
		))
	assert.Equal(t, []bool{true, false},
		and[[]bool](
			nodes.ConstOutput[[]bool]{Val: []bool{true, false}},
			nodes.ConstOutput[bool]{Val: true},
		), "a single value applies to every element")
}
