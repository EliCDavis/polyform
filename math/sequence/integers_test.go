package sequence_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/sequence"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

func TestIntegersCountsUpFromStart(t *testing.T) {
	node := nodes.Struct[sequence.IntegersNode]{Data: sequence.IntegersNode{
		Count: nodes.ConstOutput[int]{Val: 3},
		Start: nodes.ConstOutput[int]{Val: 5},
	}}
	assert.Equal(t, []int{5, 6, 7}, nodes.GetNodeOutputPort[[]int](&node, "Out").Value())
}

func TestIntegersWithNoCountIsEmpty(t *testing.T) {
	node := nodes.Struct[sequence.IntegersNode]{Data: sequence.IntegersNode{
		Count: nodes.ConstOutput[int]{Val: -2},
	}}
	values := nodes.GetNodeOutputPort[[]int](&node, "Out").Value()
	assert.NotNil(t, values)
	assert.Empty(t, values)
}
