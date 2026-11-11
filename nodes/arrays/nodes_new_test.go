package arrays_test

import (
	"testing"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/arrays"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func letters() nodes.ConstOutput[[]string] {
	return nodes.ConstOutput[[]string]{Val: []string{"a", "b", "c", "d", "e"}}
}

func TestChunkSplitsIntoRunsOfAFixedSize(t *testing.T) {
	node := &nodes.Struct[arrays.ChunkNode]{}
	wire(t, node, "Array", letters())
	wire(t, node, "Size", nodes.ConstOutput[int]{Val: 2})

	out := node.Outputs()["Out"]
	assert.Equal(t, "[][]string", out.(nodes.Typed).Type())

	typed, ok := out.(nodes.Output[[][]string])
	require.True(t, ok, "got %T", out)
	assert.Equal(t, [][]string{{"a", "b"}, {"c", "d"}, {"e"}}, typed.Value(),
		"the last run is short when the size does not divide evenly")
}

func TestChunkRefusesASizeBelowOne(t *testing.T) {
	node := &nodes.Struct[arrays.ChunkNode]{}
	wire(t, node, "Array", letters())
	wire(t, node, "Size", nodes.ConstOutput[int]{Val: 0})

	assert.NotEmpty(t, errorsOf(node, "Out"), "a zero size would never terminate")
}

func TestGroupSortsElementsIntoNumberedGroups(t *testing.T) {
	node := &nodes.Struct[arrays.GroupNode]{}
	wire(t, node, "Array", letters())
	wire(t, node, "Groups", nodes.ConstOutput[[]int]{Val: []int{2, 0, 2, 0, 2}})
	wire(t, node, "Count", nodes.ConstOutput[int]{Val: 4})

	out := node.Outputs()["Out"]
	assert.Equal(t, "[][]string", out.(nodes.Typed).Type())

	typed, ok := out.(nodes.Output[[][]string])
	require.True(t, ok, "got %T", out)
	assert.Equal(t, [][]string{{"b", "d"}, {}, {"a", "c", "e"}, {}}, typed.Value())
}

func TestGroupDiscardsElementsInGroupMinusOne(t *testing.T) {
	node := &nodes.Struct[arrays.GroupNode]{}
	wire(t, node, "Array", letters())
	wire(t, node, "Groups", nodes.ConstOutput[[]int]{Val: []int{0, -1, 1, -1, 0}})
	wire(t, node, "Count", nodes.ConstOutput[int]{Val: 2})

	assert.Equal(t, [][]string{{"a", "e"}, {"c"}}, nodes.GetNodeOutputPort[[][]string](node, "Out").Value())
}

func TestIndicesCountsTheArray(t *testing.T) {
	node := &nodes.Struct[arrays.IndicesNode]{}
	wire(t, node, "Array", letters())

	assert.Equal(t, []int{0, 1, 2, 3, 4}, nodes.GetNodeOutputPort[[]int](node, "Out").Value())
}

func TestGroupRefusesAGroupOutsideCount(t *testing.T) {
	for name, groups := range map[string][]int{
		"too high": {0, 1, 2, 3, 0},
		"negative": {0, 1, -2, 0, 0},
	} {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[arrays.GroupNode]{}
			wire(t, node, "Array", letters())
			wire(t, node, "Groups", nodes.ConstOutput[[]int]{Val: groups})
			wire(t, node, "Count", nodes.ConstOutput[int]{Val: 3})

			assert.NotEmpty(t, errorsOf(node, "Out"))
		})
	}
}

func TestGroupRefusesMismatchedLengths(t *testing.T) {
	node := &nodes.Struct[arrays.GroupNode]{}
	wire(t, node, "Array", letters())
	wire(t, node, "Groups", nodes.ConstOutput[[]int]{Val: []int{0, 1}})
	wire(t, node, "Count", nodes.ConstOutput[int]{Val: 2})

	assert.NotEmpty(t, errorsOf(node, "Out"))
}

func TestFlattenJoinsInnerArraysInOrder(t *testing.T) {
	node := &nodes.Struct[arrays.FlattenNode]{}
	wire(t, node, "Arrays", nodes.ConstOutput[[][]string]{Val: [][]string{{"a", "b"}, {}, {"c"}, {"d", "e"}}})

	out := node.Outputs()["Out"]
	assert.Equal(t, "[]string", out.(nodes.Typed).Type())

	typed, ok := out.(nodes.Output[[]string])
	require.True(t, ok, "got %T", out)
	assert.Equal(t, []string{"a", "b", "c", "d", "e"}, typed.Value())
}

func TestFlattenOfNoArraysIsEmpty(t *testing.T) {
	node := &nodes.Struct[arrays.FlattenNode]{}
	wire(t, node, "Arrays", nodes.ConstOutput[[][]string]{Val: [][]string{}})

	value := nodes.GetNodeOutputPort[[]string](node, "Out").Value()
	assert.NotNil(t, value)
	assert.Empty(t, value)
}

func TestFlattenUndoesChunk(t *testing.T) {
	chunk := &nodes.Struct[arrays.ChunkNode]{}
	wire(t, chunk, "Array", letters())
	wire(t, chunk, "Size", nodes.ConstOutput[int]{Val: 2})

	flatten := &nodes.Struct[arrays.FlattenNode]{}
	wire(t, flatten, "Arrays", chunk.Outputs()["Out"])

	assert.Equal(t, letters().Val, nodes.GetNodeOutputPort[[]string](flatten, "Out").Value())
}

func TestPartitionSplitsInTwo(t *testing.T) {
	node := &nodes.Struct[arrays.PartitionNode]{}
	wire(t, node, "Array", letters())
	wire(t, node, "At", nodes.ConstOutput[int]{Val: 2})

	assert.Equal(t, []string{"a", "b"}, nodes.GetNodeOutputPort[[]string](node, "Before").Value())
	assert.Equal(t, []string{"c", "d", "e"}, nodes.GetNodeOutputPort[[]string](node, "After").Value())
}

func TestPartitionCountsBackFromTheEnd(t *testing.T) {
	node := &nodes.Struct[arrays.PartitionNode]{}
	wire(t, node, "Array", letters())
	wire(t, node, "At", nodes.ConstOutput[int]{Val: -1})

	assert.Equal(t, []string{"a", "b", "c", "d"}, nodes.GetNodeOutputPort[[]string](node, "Before").Value())
	assert.Equal(t, []string{"e"}, nodes.GetNodeOutputPort[[]string](node, "After").Value())
}

func TestPartitionClampsAnOutOfRangeSplit(t *testing.T) {
	node := &nodes.Struct[arrays.PartitionNode]{}
	wire(t, node, "Array", letters())
	wire(t, node, "At", nodes.ConstOutput[int]{Val: 99})

	assert.Len(t, nodes.GetNodeOutputPort[[]string](node, "Before").Value(), 5)
	assert.Empty(t, nodes.GetNodeOutputPort[[]string](node, "After").Value())
}

func TestIndexOfFindsAValue(t *testing.T) {
	node := &nodes.Struct[arrays.IndexOfNode]{}
	wire(t, node, "Array", letters())
	wire(t, node, "Value", nodes.ConstOutput[string]{Val: "c"})

	assert.Equal(t, 2, nodes.GetNodeOutputPort[int](node, "Index").Value())
	assert.True(t, nodes.GetNodeOutputPort[bool](node, "Found").Value())
}

func TestIndexOfReportsAMissingValue(t *testing.T) {
	node := &nodes.Struct[arrays.IndexOfNode]{}
	wire(t, node, "Array", letters())
	wire(t, node, "Value", nodes.ConstOutput[string]{Val: "z"})

	assert.Equal(t, -1, nodes.GetNodeOutputPort[int](node, "Index").Value())
	assert.False(t, nodes.GetNodeOutputPort[bool](node, "Found").Value(),
		"Found is why -1 does not have to be read as an index")
}

func TestShuffleIsDeterministicForASeed(t *testing.T) {
	shuffled := func(seed int) []string {
		node := &nodes.Struct[arrays.ShuffleNode]{}
		wire(t, node, "Array", letters())
		wire(t, node, "Seed", nodes.ConstOutput[int]{Val: seed})
		return nodes.GetNodeOutputPort[[]string](node, "Out").Value()
	}

	assert.Equal(t, shuffled(7), shuffled(7), "the same seed has to rebuild the same graph")
	assert.ElementsMatch(t, letters().Val, shuffled(7), "every element survives")
	assert.NotEqual(t, shuffled(7), shuffled(8))
}

func TestSamplePicksWithoutRepeating(t *testing.T) {
	node := &nodes.Struct[arrays.SampleNode]{}
	wire(t, node, "Array", letters())
	wire(t, node, "Count", nodes.ConstOutput[int]{Val: 3})
	wire(t, node, "Seed", nodes.ConstOutput[int]{Val: 1})

	picked := nodes.GetNodeOutputPort[[]string](node, "Out").Value()
	require.Len(t, picked, 3)

	seen := map[string]bool{}
	for _, value := range picked {
		assert.False(t, seen[value], "%q was picked twice", value)
		seen[value] = true
		assert.Contains(t, letters().Val, value)
	}
}

func TestSampleCapsAtWhatTheArrayHolds(t *testing.T) {
	node := &nodes.Struct[arrays.SampleNode]{}
	wire(t, node, "Array", letters())
	wire(t, node, "Count", nodes.ConstOutput[int]{Val: 99})

	assert.Len(t, nodes.GetNodeOutputPort[[]string](node, "Out").Value(), 5)
}

func TestRandomPickCanPickMoreThanTheArrayHolds(t *testing.T) {
	node := &nodes.Struct[arrays.RandomPickNode]{}
	wire(t, node, "Array", letters())
	wire(t, node, "Count", nodes.ConstOutput[int]{Val: 200})
	wire(t, node, "Seed", nodes.ConstOutput[int]{Val: 1})

	picked := nodes.GetNodeOutputPort[[]string](node, "Out").Value()
	require.Len(t, picked, 200)

	seen := map[string]int{}
	for _, value := range picked {
		require.Contains(t, letters().Val, value)
		seen[value]++
	}
	assert.Len(t, seen, len(letters().Val), "200 picks from 5 letters should land on every letter")
}

func TestRandomPickIsTheSameForTheSameSeed(t *testing.T) {
	pick := func(seed int) []string {
		node := &nodes.Struct[arrays.RandomPickNode]{}
		wire(t, node, "Array", letters())
		wire(t, node, "Count", nodes.ConstOutput[int]{Val: 8})
		wire(t, node, "Seed", nodes.ConstOutput[int]{Val: seed})
		return nodes.GetNodeOutputPort[[]string](node, "Out").Value()
	}

	assert.Equal(t, pick(3), pick(3))
	assert.NotEqual(t, pick(3), pick(4))
}

func TestRandomPickFromAnEmptyArrayIsAnError(t *testing.T) {
	node := &nodes.Struct[arrays.RandomPickNode]{}
	wire(t, node, "Array", nodes.ConstOutput[[]string]{Val: nil})
	wire(t, node, "Count", nodes.ConstOutput[int]{Val: 2})

	port := node.Outputs()["Out"]
	assert.Empty(t, nodes.GetNodeOutputPort[[]string](node, "Out").Value())
	assert.NotEmpty(t, port.(nodes.ObservableExecution).ExecutionReport().Errors)
}

func TestPadExtendsWithAValue(t *testing.T) {
	node := &nodes.Struct[arrays.PadNode]{}
	wire(t, node, "Array", nodes.ConstOutput[[]string]{Val: []string{"a", "b"}})
	wire(t, node, "Length", nodes.ConstOutput[int]{Val: 4})
	wire(t, node, "Value", nodes.ConstOutput[string]{Val: "-"})

	assert.Equal(t, []string{"a", "b", "-", "-"}, nodes.GetNodeOutputPort[[]string](node, "Out").Value())
}

func TestPadWithoutAValueUsesTheTypesZero(t *testing.T) {
	node := &nodes.Struct[arrays.PadNode]{}
	wire(t, node, "Array", nodes.ConstOutput[[]int]{Val: []int{1, 2}})
	wire(t, node, "Length", nodes.ConstOutput[int]{Val: 4})

	assert.Equal(t, []int{1, 2, 0, 0}, nodes.GetNodeOutputPort[[]int](node, "Out").Value())
}

func TestPadLeavesALongEnoughArrayAlone(t *testing.T) {
	node := &nodes.Struct[arrays.PadNode]{}
	wire(t, node, "Array", letters())
	wire(t, node, "Length", nodes.ConstOutput[int]{Val: 2})

	assert.Len(t, nodes.GetNodeOutputPort[[]string](node, "Out").Value(), 5)
}
