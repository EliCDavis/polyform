package arrays_test

import (
	"testing"

	"github.com/EliCDavis/polyform/arrays"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func wire(t *testing.T, node nodes.Node, port string, source nodes.OutputPort) {
	t.Helper()
	input, ok := node.Inputs()[port].(nodes.SingleValueInputPort)
	require.True(t, ok, "node has no single-value input %q", port)
	require.NoError(t, input.Set(source))
}

func wireAll(t *testing.T, node nodes.Node, port string, sources ...nodes.OutputPort) {
	t.Helper()
	input, ok := node.Inputs()[port].(nodes.ArrayValueInputPort)
	require.True(t, ok, "node has no array input %q", port)
	for _, source := range sources {
		require.NoError(t, input.Add(source))
	}
}

func vectors(values ...vector3.Float64) nodes.ConstOutput[[]vector3.Float64] {
	return nodes.ConstOutput[[]vector3.Float64]{Val: values}
}

func TestIndexReadsAnElementOfAnyType(t *testing.T) {
	a := vector3.New(1., 2., 3.)
	b := vector3.New(4., 5., 6.)
	c := vector3.New(7., 8., 9.)

	node := &nodes.Struct[arrays.IndexNode]{}
	wire(t, node, "Array", vectors(a, b, c))
	wire(t, node, "Index", nodes.ConstOutput[int]{Val: 1})

	out := node.Outputs()["Out"]
	assert.Equal(t, "github.com/EliCDavis/vector/vector3.Vector[float64]", out.(nodes.Typed).Type())

	typed, ok := out.(nodes.Output[vector3.Float64])
	require.True(t, ok, "the element port really is an Output[vector3.Float64], got %T", out)
	assert.Equal(t, b, typed.Value())
}

func TestIndexCountsBackFromTheEnd(t *testing.T) {
	node := &nodes.Struct[arrays.IndexNode]{}
	wire(t, node, "Array", vectors(vector3.New(1., 0., 0.), vector3.New(2., 0., 0.)))
	wire(t, node, "Index", nodes.ConstOutput[int]{Val: -1})

	typed := node.Outputs()["Out"].(nodes.Output[vector3.Float64])
	assert.Equal(t, vector3.New(2., 0., 0.), typed.Value())
}

func TestFirstAndLast(t *testing.T) {
	values := vectors(vector3.New(1., 0., 0.), vector3.New(2., 0., 0.), vector3.New(3., 0., 0.))

	first := &nodes.Struct[arrays.FirstNode]{}
	wire(t, first, "Array", values)
	assert.Equal(t, vector3.New(1., 0., 0.), first.Outputs()["Out"].(nodes.Output[vector3.Float64]).Value())

	last := &nodes.Struct[arrays.LastNode]{}
	wire(t, last, "Array", values)
	assert.Equal(t, vector3.New(3., 0., 0.), last.Outputs()["Out"].(nodes.Output[vector3.Float64]).Value())
}

func TestLengthOutputsAFixedIntRegardlessOfElementType(t *testing.T) {
	node := &nodes.Struct[arrays.LengthNode]{}
	wire(t, node, "Array", vectors(vector3.Zero[float64](), vector3.One[float64]()))

	assert.Equal(t, 2, nodes.GetNodeOutputPort[int](node, "Out").Value())
	assert.Equal(t, "int", node.Outputs()["Out"].(nodes.Typed).Type(),
		"a fixed output beside dynamic inputs keeps its own type")
}

func TestSliceAndReverseKeepTheArrayType(t *testing.T) {
	values := vectors(
		vector3.New(1., 0., 0.),
		vector3.New(2., 0., 0.),
		vector3.New(3., 0., 0.),
		vector3.New(4., 0., 0.),
	)

	slice := &nodes.Struct[arrays.SliceNode]{}
	wire(t, slice, "Array", values)
	wire(t, slice, "Start", nodes.ConstOutput[int]{Val: 1})
	wire(t, slice, "End", nodes.ConstOutput[int]{Val: 3})

	out := slice.Outputs()["Out"]
	assert.Equal(t, "[]github.com/EliCDavis/vector/vector3.Vector[float64]", out.(nodes.Typed).Type())
	typed, ok := out.(nodes.Output[[]vector3.Float64])
	require.True(t, ok, "got %T", out)
	assert.Equal(t, []vector3.Float64{vector3.New(2., 0., 0.), vector3.New(3., 0., 0.)}, typed.Value())

	reverse := &nodes.Struct[arrays.ReverseNode]{}
	wire(t, reverse, "Array", values)
	assert.Equal(t, []vector3.Float64{
		vector3.New(4., 0., 0.),
		vector3.New(3., 0., 0.),
		vector3.New(2., 0., 0.),
		vector3.New(1., 0., 0.),
	}, reverse.Outputs()["Out"].(nodes.Output[[]vector3.Float64]).Value())
}

func TestSliceClampsOutOfRangeBounds(t *testing.T) {
	node := &nodes.Struct[arrays.SliceNode]{}
	wire(t, node, "Array", vectors(vector3.Zero[float64](), vector3.One[float64]()))
	wire(t, node, "Start", nodes.ConstOutput[int]{Val: -5})
	wire(t, node, "End", nodes.ConstOutput[int]{Val: 99})

	assert.Len(t, node.Outputs()["Out"].(nodes.Output[[]vector3.Float64]).Value(), 2)
}

func TestFromElementsCollectsIndividualPortsIntoAnArray(t *testing.T) {
	node := &nodes.Struct[arrays.FromElementsNode]{}
	wireAll(t, node, "Elements",
		nodes.ConstOutput[vector3.Float64]{Val: vector3.New(1., 0., 0.)},
		nodes.ConstOutput[vector3.Float64]{Val: vector3.New(0., 1., 0.)},
	)

	out := node.Outputs()["Out"]
	assert.Equal(t, "[]github.com/EliCDavis/vector/vector3.Vector[float64]", out.(nodes.Typed).Type())
	typed, ok := out.(nodes.Output[[]vector3.Float64])
	require.True(t, ok, "got %T", out)
	assert.Equal(t, []vector3.Float64{vector3.New(1., 0., 0.), vector3.New(0., 1., 0.)}, typed.Value())
}

func TestAppendAndConcat(t *testing.T) {
	appendNode := &nodes.Struct[arrays.AppendNode]{}
	wire(t, appendNode, "Array", vectors(vector3.New(1., 0., 0.)))
	wireAll(t, appendNode, "Elements", nodes.ConstOutput[vector3.Float64]{Val: vector3.New(2., 0., 0.)})
	assert.Equal(t, []vector3.Float64{vector3.New(1., 0., 0.), vector3.New(2., 0., 0.)},
		appendNode.Outputs()["Out"].(nodes.Output[[]vector3.Float64]).Value())

	concat := &nodes.Struct[arrays.ConcatNode]{}
	wireAll(t, concat, "Arrays",
		vectors(vector3.New(1., 0., 0.)),
		vectors(vector3.New(2., 0., 0.), vector3.New(3., 0., 0.)),
	)
	assert.Equal(t, []vector3.Float64{
		vector3.New(1., 0., 0.),
		vector3.New(2., 0., 0.),
		vector3.New(3., 0., 0.),
	}, concat.Outputs()["Out"].(nodes.Output[[]vector3.Float64]).Value())
}

func TestRepeatBuildsAnArrayOfAnyType(t *testing.T) {
	node := &nodes.Struct[arrays.RepeatNode]{}
	wire(t, node, "Value", nodes.ConstOutput[vector3.Float64]{Val: vector3.New(1., 2., 3.)})
	wire(t, node, "Times", nodes.ConstOutput[int]{Val: 3})

	typed, ok := node.Outputs()["Out"].(nodes.Output[[]vector3.Float64])
	require.True(t, ok)
	assert.Equal(t, []vector3.Float64{
		vector3.New(1., 2., 3.),
		vector3.New(1., 2., 3.),
		vector3.New(1., 2., 3.),
	}, typed.Value())
}

// A mesh is the case a per-type generic would have been worst for: it is
// expensive to copy and there was no reason to register an array node for it.
func TestArrayNodesCarryMeshes(t *testing.T) {
	cube := nodes.GetNodeOutputPort[modeling.Mesh](
		&nodes.Struct[primitives.CubeNode]{}, "Out")

	repeat := &nodes.Struct[arrays.RepeatNode]{}
	wire(t, repeat, "Value", cube)
	wire(t, repeat, "Times", nodes.ConstOutput[int]{Val: 2})

	meshes, ok := repeat.Outputs()["Out"].(nodes.Output[[]modeling.Mesh])
	require.True(t, ok, "got %T", repeat.Outputs()["Out"])
	require.Len(t, meshes.Value(), 2)

	first := &nodes.Struct[arrays.FirstNode]{}
	wire(t, first, "Array", repeat.Outputs()["Out"])
	out, ok := first.Outputs()["Out"].(nodes.Output[modeling.Mesh])
	require.True(t, ok, "got %T", first.Outputs()["Out"])
	assert.Equal(t, cube.Value().PrimitiveCount(), out.Value().PrimitiveCount())
}
