package nodes_test

import (
	"testing"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	pickType  nodes.DynamicType
	elemType  nodes.DynamicType
	leftType  nodes.DynamicType
	rightType nodes.DynamicType
)

// Go renders a generic's type arguments with their full import path, so
// that is what a type variable is keyed by.
const pkg = "github.com/EliCDavis/polyform/nodes_test."

type pickNode struct {
	Condition nodes.Output[bool]
	A         nodes.DynamicPort[pickType]
	B         nodes.DynamicPort[pickType]
}

func (p pickNode) Out(out *nodes.Dynamic[pickType]) {
	if nodes.TryGetOutputValue(out, p.Condition, false) {
		out.Forward(p.A)
		return
	}
	out.Forward(p.B)
}

type countNode struct {
	Values nodes.DynamicPort[[]elemType]
}

func (c countNode) Out(out *nodes.StructOutput[int]) {
	values, ok := nodes.DynamicArrayValue(c.Values)
	if !ok {
		out.Set(-1)
		return
	}
	out.Set(values.Len())
}

type firstNode struct {
	Values nodes.DynamicPort[[]elemType]
}

func (f firstNode) Out(out *nodes.Dynamic[elemType]) {
	values, ok := nodes.DynamicArrayValue(f.Values)
	if !ok || values.Len() == 0 {
		return
	}
	out.Set(values.At(0))
}

type pairNode struct {
	Left  nodes.DynamicPort[leftType]
	Right nodes.DynamicPort[rightType]
}

func (p pairNode) OutLeft(out *nodes.Dynamic[leftType])   { out.Forward(p.Left) }
func (p pairNode) OutRight(out *nodes.Dynamic[rightType]) { out.Forward(p.Right) }

func connect(t *testing.T, node nodes.Node, port string, source nodes.OutputPort) {
	t.Helper()
	input, ok := node.Inputs()[port].(nodes.SingleValueInputPort)
	require.True(t, ok, "node has no single-value input %q", port)
	require.NoError(t, input.Set(source))
}

func TestDynamicForwardsTheSelectedPortWithoutKnowingItsType(t *testing.T) {
	yes := nodes.ConstOutput[string]{Val: "yes"}
	no := nodes.ConstOutput[string]{Val: "no"}
	condition := nodes.NewValue(true)

	node := &nodes.Struct[pickNode]{}
	connect(t, node, "A", yes)
	connect(t, node, "B", no)
	connect(t, node, "Condition", condition.Outputs()["Value"])

	assert.Equal(t, map[string]string{pkg + "pickType": "string"}, node.DynamicTypes(),
		"the type came from whatever was wired in first, keyed by the variable it was declared with")

	out := node.Outputs()["Out"]
	typed, ok := out.(nodes.Output[string])
	require.True(t, ok, "a bound dynamic output really is an Output[string], got %T", out)
	assert.Equal(t, "string", out.(nodes.Typed).Type())
	assert.Equal(t, "yes", typed.Value())

	condition.Set(false)
	assert.Equal(t, "no", typed.Value(), "flipping the condition re-routes without re-typing anything")
}

func TestDynamicOutputIsUntypedUntilSomethingConnects(t *testing.T) {
	node := &nodes.Struct[pickNode]{}

	out := node.Outputs()["Out"]
	assert.Equal(t, "", out.(nodes.Typed).Type(), "an unbound port reports no type, so anything may connect")
	assert.Equal(t, "", node.Inputs()["A"].(nodes.Typed).Type())

	_, isString := out.(nodes.Output[string])
	assert.False(t, isString)
}

func TestDynamicRefusesASecondTypeOnTheSameVariable(t *testing.T) {
	node := &nodes.Struct[pickNode]{}
	connect(t, node, "A", nodes.ConstOutput[string]{Val: "yes"})

	input := node.Inputs()["B"].(nodes.SingleValueInputPort)
	err := input.Set(nodes.ConstOutput[int]{Val: 3})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already carrying string")
}

func TestDynamicVariablesOnOneNodeBindIndependently(t *testing.T) {
	node := &nodes.Struct[pairNode]{}
	connect(t, node, "Left", nodes.ConstOutput[string]{Val: "a"})
	connect(t, node, "Right", nodes.ConstOutput[float64]{Val: 2})

	assert.Equal(t, map[string]string{
		pkg + "leftType":  "string",
		pkg + "rightType": "float64",
	}, node.DynamicTypes())

	assert.Equal(t, "string", node.Outputs()["Out Left"].(nodes.Typed).Type())
	assert.Equal(t, "float64", node.Outputs()["Out Right"].(nodes.Typed).Type())
}

func TestDynamicReleasingOneVariableLeavesTheOther(t *testing.T) {
	node := &nodes.Struct[pairNode]{}
	connect(t, node, "Left", nodes.ConstOutput[string]{Val: "a"})
	connect(t, node, "Right", nodes.ConstOutput[float64]{Val: 2})

	node.Inputs()["Left"].Clear()
	node.Inputs()["Left"].(nodes.DynamicallyTypedPort).ReleaseType()

	assert.Equal(t, map[string]string{pkg + "rightType": "float64"}, node.DynamicTypes())

	connect(t, node, "Left", nodes.ConstOutput[int]{Val: 1})
	assert.Equal(t, "int", node.Outputs()["Out Left"].(nodes.Typed).Type())
}

func TestDynamicTypeCanBeClearedToCarrySomethingElse(t *testing.T) {
	node := &nodes.Struct[pickNode]{}
	connect(t, node, "A", nodes.ConstOutput[string]{Val: "yes"})
	require.Equal(t, map[string]string{pkg + "pickType": "string"}, node.DynamicTypes())

	node.Inputs()["A"].Clear()
	node.Inputs()["A"].(nodes.DynamicallyTypedPort).ReleaseType()
	require.Empty(t, node.DynamicTypes())

	connect(t, node, "A", nodes.ConstOutput[float64]{Val: 1})
	assert.Equal(t, map[string]string{pkg + "pickType": "float64"}, node.DynamicTypes())
}

func TestDynamicBindsFromTheConsumerSide(t *testing.T) {
	// Nothing is connected to learn the type from, so this path needs the
	// type index a real graph builds at startup.
	nodes.DiscoverNodePortTypes(&nodes.Struct[floatValueNode]{})

	node := &nodes.Struct[pickNode]{}
	bindable, ok := node.Outputs()["Out"].(nodes.DynamicallyTypedPort)
	require.True(t, ok)
	require.NoError(t, bindable.BindType("float64"))

	assert.Equal(t, map[string]string{pkg + "pickType": "float64"}, node.DynamicTypes())
	_, isFloat := node.Outputs()["Out"].(nodes.Output[float64])
	assert.True(t, isFloat, "binding from the consumer end types the port the same way")
}

func TestDynamicPatternRelatesAnArrayInputToAScalarOutput(t *testing.T) {
	values := nodes.ConstOutput[[]string]{Val: []string{"a", "b", "c"}}

	counter := &nodes.Struct[countNode]{}
	connect(t, counter, "Values", values)
	assert.Equal(t, map[string]string{pkg + "elemType": "string"}, counter.DynamicTypes(),
		"[]string solved against []Element binds Element to string")
	assert.Equal(t, "[]string", counter.Inputs()["Values"].(nodes.Typed).Type())
	assert.Equal(t, 3, nodes.GetNodeOutputPort[int](counter, "Out").Value())

	first := &nodes.Struct[firstNode]{}
	connect(t, first, "Values", values)
	out := first.Outputs()["Out"]
	assert.Equal(t, "string", out.(nodes.Typed).Type(), "the output is Element, not []Element")
	typed, ok := out.(nodes.Output[string])
	require.True(t, ok, "got %T", out)
	assert.Equal(t, "a", typed.Value())
}

func TestDynamicSetOfTheWrongTypeReadsAsZero(t *testing.T) {
	node := &nodes.Struct[firstNode]{}
	connect(t, node, "Values", nodes.ConstOutput[[]string]{Val: []string{}})

	typed, ok := node.Outputs()["Out"].(nodes.Output[string])
	require.True(t, ok)
	assert.Equal(t, "", typed.Value(), "a node that set nothing reads as the zero value")
}

type floatValueNode struct {
	Val nodes.Output[float64]
}

func (f floatValueNode) Out(out *nodes.StructOutput[float64]) {
	out.Set(nodes.TryGetOutputValue(out, f.Val, 0))
}
