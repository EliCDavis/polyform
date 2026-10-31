package logic_test

import (
	"testing"

	"github.com/EliCDavis/polyform/logic"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

func TestEquals(t *testing.T) {
	equal := func(a, b nodes.OutputPort) bool {
		return nodes.GetNodeOutputPort[bool](&nodes.Struct[logic.EqualsNode]{Data: logic.EqualsNode{A: a, B: b}}, "Out").Value()
	}
	assert.True(t, equal(nodes.ConstOutput[string]{Val: "a"}, nodes.ConstOutput[string]{Val: "a"}))
	assert.False(t, equal(nodes.ConstOutput[string]{Val: "a"}, nodes.ConstOutput[string]{Val: "b"}))
	assert.True(t, equal(nodes.ConstOutput[int]{Val: 2}, nodes.ConstOutput[int]{Val: 2}))
	assert.False(t, equal(nodes.ConstOutput[float64]{Val: 1}, nil))
}
