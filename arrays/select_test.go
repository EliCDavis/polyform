package arrays_test

import (
	"testing"

	"github.com/EliCDavis/polyform/arrays"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
)

func TestGatherReadsIndicesInOrderWithRepeatsAndNegatives(t *testing.T) {
	node := &nodes.Struct[arrays.GatherNode]{}
	wire(t, node, "Array", letters())
	wire(t, node, "Indices", nodes.ConstOutput[[]int]{Val: []int{4, 0, 0, -2}})

	assert.Equal(t, []string{"e", "a", "a", "d"}, nodes.GetNodeOutputPort[[]string](node, "Out").Value())
}

func TestGatherNamesAnIndexOutsideTheArray(t *testing.T) {
	node := &nodes.Struct[arrays.GatherNode]{}
	wire(t, node, "Array", letters())
	wire(t, node, "Indices", nodes.ConstOutput[[]int]{Val: []int{1, 5}})

	assert.Contains(t, errorsOf(node, "Out"), "index 5 is outside an array of 5 element(s)")
}

func TestFilterKeepsFlaggedElementsAndDropsTheRest(t *testing.T) {
	node := &nodes.Struct[arrays.FilterNode]{}
	wire(t, node, "Array", letters())
	wire(t, node, "Keep", nodes.ConstOutput[[]bool]{Val: []bool{true, false, true, false, false}})

	assert.Equal(t, []string{"a", "c"}, nodes.GetNodeOutputPort[[]string](node, "Kept").Value())
	assert.Equal(t, []string{"b", "d", "e"}, nodes.GetNodeOutputPort[[]string](node, "Dropped").Value())
}

func TestFilterRefusesFlagsOfTheWrongLength(t *testing.T) {
	node := &nodes.Struct[arrays.FilterNode]{}
	wire(t, node, "Array", letters())
	wire(t, node, "Keep", nodes.ConstOutput[[]bool]{Val: []bool{true}})

	assert.Contains(t, errorsOf(node, "Kept"), "1 flag(s) for 5 element(s)")
}

func TestSplitCutsConsecutiveRunsOfTheGivenLengths(t *testing.T) {
	node := &nodes.Struct[arrays.SplitNode]{}
	wire(t, node, "Array", letters())
	wire(t, node, "Lengths", nodes.ConstOutput[[]int]{Val: []int{2, 0, 3}})

	assert.Equal(t, [][]string{{"a", "b"}, {}, {"c", "d", "e"}}, nodes.GetNodeOutputPort[[][]string](node, "Out").Value())
}

func TestSplitRefusesLengthsThatDoNotCoverTheArray(t *testing.T) {
	node := &nodes.Struct[arrays.SplitNode]{}
	wire(t, node, "Array", letters())
	wire(t, node, "Lengths", nodes.ConstOutput[[]int]{Val: []int{2, 2}})

	assert.Contains(t, errorsOf(node, "Out"), "lengths add up to 4, not 5")
}

func TestZipTurnsParallelPointArraysIntoOnePathEach(t *testing.T) {
	starts := nodes.ConstOutput[[]vector2.Float64]{Val: []vector2.Float64{vector2.New(0., 0.), vector2.New(0., 5.)}}
	bends := nodes.ConstOutput[[]vector2.Float64]{Val: []vector2.Float64{vector2.New(1., 0.), vector2.New(1., 5.)}}
	ends := nodes.ConstOutput[[]vector2.Float64]{Val: []vector2.Float64{vector2.New(1., 1.), vector2.New(1., 6.)}}

	node := &nodes.Struct[arrays.ZipNode]{}
	wireAll(t, node, "Arrays", starts, bends, ends)

	assert.Equal(t, [][]vector2.Float64{
		{vector2.New(0., 0.), vector2.New(1., 0.), vector2.New(1., 1.)},
		{vector2.New(0., 5.), vector2.New(1., 5.), vector2.New(1., 6.)},
	}, nodes.GetNodeOutputPort[[][]vector2.Float64](node, "Out").Value())
}

func TestZipRefusesArraysOfDifferentLengths(t *testing.T) {
	node := &nodes.Struct[arrays.ZipNode]{}
	wireAll(t, node, "Arrays", letters(), nodes.ConstOutput[[]string]{Val: []string{"x"}})

	assert.Contains(t, errorsOf(node, "Out"), "arrays of 5 and 1 element(s) cannot be zipped")
}
