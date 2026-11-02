package nodes_test

import (
	"reflect"
	"testing"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type opaqueThing struct{ Weight float64 }

type fillNode struct {
	Radius nodes.Output[float64] `default:"0.5"`
	Count  nodes.Output[int]
	Name   nodes.Output[string]              `default:"shape"`
	Offset nodes.Output[vector3.Float64]     `default:"1,2,3"`
	Scale  nodes.LiftedPort[float64]         `default:"2"`
	Shift  nodes.LiftedPort[vector3.Float64] `default:"10,0,0"`
	Thing  nodes.Output[opaqueThing]
	Items  []nodes.Output[float64]
}

func (n fillNode) Out(out *nodes.StructOutput[float64]) {
	radius := nodes.GetOutputValue(out, n.Radius)
	count := nodes.GetOutputValue(out, n.Count)
	offset := nodes.GetOutputValue(out, n.Offset)
	scale := nodes.GetOutputValue(out, n.Scale.(nodes.Output[float64]))
	shift := nodes.GetOutputValue(out, n.Shift.(nodes.Output[vector3.Float64]))
	out.Set(radius*float64(count) + offset.X()*scale + shift.X())
}

func (n fillNode) Label(out *nodes.StructOutput[string]) {
	out.Set(nodes.GetOutputValue(out, n.Name))
}

func (n fillNode) Held(out *nodes.StructOutput[bool]) {
	out.Set(n.Thing != nil)
}

func TestANilPrimitiveInputReadsItsDefault(t *testing.T) {
	node := &nodes.Struct[fillNode]{}

	assert.Equal(t, 12., nodes.GetNodeOutputPort[float64](node, "Out").Value(), "0.5*0 + 1*2 + 10")
	assert.Equal(t, "shape", nodes.GetNodeOutputPort[string](node, "Label").Value())
	assert.False(t, nodes.GetNodeOutputPort[bool](node, "Held").Value(), "an unregistered type stays nil")

	held := node.Data.Radius
	require.NotNil(t, held)
	assert.Same(t, node, held.Node(), "the filled port belongs to the node itself")
	assert.Equal(t, "Radius", held.Name())
}

func TestAWiredInputIsLeftAlone(t *testing.T) {
	node := &nodes.Struct[fillNode]{Data: fillNode{
		Radius: nodes.ConstOutput[float64]{Val: 2},
		Count:  nodes.ConstOutput[int]{Val: 3},
	}}
	assert.Equal(t, 18., nodes.GetNodeOutputPort[float64](node, "Out").Value(), "2*3 + 1*2 + 10")
}

func TestAClearedInputReadsItsDefaultAgain(t *testing.T) {
	node := &nodes.Struct[fillNode]{Data: fillNode{Count: nodes.NewValue(4).Outputs()["Value"].(nodes.Output[int])}}
	out := nodes.GetNodeOutputPort[float64](node, "Out")
	assert.Equal(t, 14., out.Value())

	require.NoError(t, node.Inputs()["Count"].(nodes.SingleValueInputPort).Set(nodes.NewValue(10).Outputs()["Value"]))
	assert.Equal(t, 17., out.Value())

	node.Inputs()["Count"].Clear()
	assert.Equal(t, 12., out.Value(), "back to the zero default")
}

type badDefaultNode struct {
	Radius nodes.Output[float64] `default:"wide"`
}

func (n badDefaultNode) Out(out *nodes.StructOutput[float64]) {
	out.Set(nodes.GetOutputValue(out, n.Radius))
}

func TestAnUnparsableDefaultTagIsReportedByName(t *testing.T) {
	node := &nodes.Struct[badDefaultNode]{}
	assert.PanicsWithError(t, `nodes_test.badDefaultNode.Radius: default tag: default "wide" is not a number`, func() {
		nodes.GetNodeOutputPort[float64](node, "Out").Value()
	})
}

type Weave int

const (
	Plain Weave = iota
	Twill
)

func parseWeave(text string) (Weave, error) {
	if text == "Twill" {
		return Twill, nil
	}
	return Plain, nil
}

type weaveNode struct {
	Weave nodes.Output[Weave] `default:"Twill"`
	Other nodes.Output[Weave]
}

func (n weaveNode) Out(out *nodes.StructOutput[int]) {
	out.Set(int(nodes.GetOutputValue(out, n.Weave))*10 + int(nodes.GetOutputValue(out, n.Other)))
}

func TestARegisteredParserReadsTheTagByName(t *testing.T) {
	nodes.RegisterLiteralParser(Plain, parseWeave)
	node := &nodes.Struct[weaveNode]{}
	assert.Equal(t, 10, nodes.GetNodeOutputPort[int](node, "Out").Value())

	raw, err := nodes.DefaultJSON("Twill", reflect.TypeOf(Plain))
	require.NoError(t, err)
	assert.Equal(t, `"Twill"`, string(raw))
}

func TestDefaultJSONForms(t *testing.T) {
	cases := []struct {
		text string
		of   any
		json string
	}{
		{"0.5", float64(0), "0.5"},
		{"64", 0, "64"},
		{"true", false, "true"},
		{"hello world", "", `"hello world"`},
		{"1,2.5,3", vector3.Zero[float64](), `{"x":1,"y":2.5,"z":3}`},
		{"1, 2, 3", vector3.Zero[int](), `{"x":1,"y":2,"z":3}`},
		{"max", float64(0), "1.7976931348623157e308"},
		{"MIN", float64(0), "-1.7976931348623157e308"},
		{"max", float32(0), "3.4028234663852886e38"},
		{"min", int32(0), "-2147483648"},
		{"max", int8(0), "127"},
		{"max", uint16(0), "65535"},
		{"min", uint(0), "0"},
		{"min, 0, max", vector3.Zero[float64](), `{"x":-1.7976931348623157e308,"y":0,"z":1.7976931348623157e308}`},
		{"max,min", vector2.Zero[int](), `{"x":9223372036854775807,"y":-9223372036854775808}`},
	}
	for _, c := range cases {
		raw, err := nodes.DefaultJSON(c.text, reflect.TypeOf(c.of))
		require.NoError(t, err, c.text)
		assert.JSONEq(t, c.json, string(raw), c.text)
	}

	for _, bad := range []struct {
		text string
		of   any
	}{
		{"abc", float64(0)},
		{"1.5", 0},
		{"yes", false},
		{"1,2", vector3.Zero[float64]()},
		{"1.5,2,3", vector3.Zero[int]()},
		{"maximum", float64(0)},
		{"max", false},
		{"1", opaqueThing{}},
	} {
		_, err := nodes.DefaultJSON(bad.text, reflect.TypeOf(bad.of))
		assert.Error(t, err, "%v %q", bad.of, bad.text)
	}
}
