package subgraph_test

import (
	"encoding/json"
	"testing"

	"github.com/EliCDavis/jbtf"
	_ "github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubOutputPort struct {
	version int
	name    string
}

func (s *stubOutputPort) Node() nodes.Node { return nil }
func (s *stubOutputPort) Name() string     { return s.name }
func (s *stubOutputPort) Version() int     { return s.version }
func (s *stubOutputPort) Type() string     { return "float64" }

func TestInputNodeDefaults(t *testing.T) {
	n := subgraph.NewInputNode("", "")
	assert.Equal(t, "Input", n.PortName())
	assert.Equal(t, "", n.PortType())
	assert.Equal(t, "Input", n.Name())
	assert.Len(t, n.Inputs(), 1)
	assert.Contains(t, n.Inputs(), subgraph.DefaultPortName)
	assert.Len(t, n.Outputs(), 1)
	assert.Contains(t, n.Outputs(), subgraph.ValuePortName)
}

func TestInputNodeCustomPort(t *testing.T) {
	n := subgraph.NewInputNode("Scale", "float64")
	assert.Equal(t, "Scale", n.PortName())
	assert.Equal(t, "float64", n.PortType())

	out := n.Outputs()[subgraph.ValuePortName]
	assert.Equal(t, "float64", out.(interface{ Type() string }).Type())
	assert.Equal(t, 0, out.Version())

	typedOut, ok := out.(nodes.Output[float64])
	assert.True(t, ok)
	assert.Equal(t, float64(0), typedOut.Value())
}

func TestInputNodeDomainTypeImplementsTypedOutput(t *testing.T) {
	portType := refutil.TypeResolution{
		IncludePackage: true,
		IncludePointer: false,
	}.Resolve(vector3.New(0., 0., 0.))

	n := subgraph.NewInputNode("Position", portType)
	out := n.Outputs()[subgraph.ValuePortName]

	typedOut, ok := out.(nodes.Output[vector3.Float64])
	require.True(t, ok, "InputNode.Value must implement nodes.Output[vector3.Float64] for port type %q", portType)

	want := vector3.New(1., 2., 3.)
	n.ReadFrom(fed{nodes.ConstOutput[vector3.Float64]{Val: want}})

	assert.Equal(t, want, typedOut.Value())
}

func TestInputNodeExternalSourceVersion(t *testing.T) {
	n := subgraph.NewInputNode("X", "float64")
	ext := &stubOutputPort{version: 42, name: "Value"}

	n.ReadFrom(fed{ext})

	out := n.Outputs()[subgraph.ValuePortName]
	assert.Equal(t, 42, out.Version())

	ext.version = 99
	assert.Equal(t, 99, out.Version())
}

func TestInputNodeDefaultFallback(t *testing.T) {
	n := subgraph.NewInputNode("Scale", "float64")
	out := n.Outputs()[subgraph.ValuePortName].(nodes.Output[float64])

	// Both unset → typed zero, no error.
	assert.Equal(t, float64(0), out.Value())

	defaultSrc := nodes.ConstOutput[float64]{Val: 3.5}
	defIn := n.Inputs()[subgraph.DefaultPortName].(nodes.SingleValueInputPort)
	require.NoError(t, defIn.Set(defaultSrc))
	assert.Equal(t, 3.5, out.Value())

	ext := nodes.ConstOutput[float64]{Val: 9}
	n.ReadFrom(fed{ext})
	assert.Equal(t, 9.0, out.Value())

	n.ReadFrom(fed{nil})
	assert.Equal(t, 3.5, out.Value())

	defIn.Clear()
	assert.Equal(t, float64(0), out.Value())
}

func TestInputNodeDefaultVersion(t *testing.T) {
	n := subgraph.NewInputNode("X", "float64")
	out := n.Outputs()[subgraph.ValuePortName]
	assert.Equal(t, 0, out.Version())

	def := &stubOutputPort{version: 7, name: "Value"}
	defIn := n.Inputs()[subgraph.DefaultPortName].(nodes.SingleValueInputPort)
	require.NoError(t, defIn.Set(def))
	assert.Equal(t, 7, out.Version())

	def.version = 11
	assert.Equal(t, 11, out.Version())

	ext := &stubOutputPort{version: 42, name: "Value"}
	n.ReadFrom(fed{ext})
	assert.Equal(t, 42, out.Version())

	n.ReadFrom(fed{nil})
	assert.Equal(t, 11, out.Version())
}

func TestInputNodeJSONRoundtrip(t *testing.T) {
	n := subgraph.NewInputNode("Position", "github.com/EliCDavis/vector/vector3.Vector[float64]")
	data, err := n.ToJSON(&jbtf.Encoder{})
	require.NoError(t, err)

	decoder, err := jbtf.NewDecoder([]byte(`{}`))
	require.NoError(t, err)

	restored := subgraph.NewInputNode("", "")
	err = restored.FromJSON(decoder, data)
	require.NoError(t, err)

	assert.Equal(t, "Position", restored.PortName())
	assert.Equal(t, "github.com/EliCDavis/vector/vector3.Vector[float64]", restored.PortType())
}

func TestOutputNodeDefaults(t *testing.T) {
	n := subgraph.NewOutputNode("", "")
	assert.Equal(t, "Output", n.PortName())
	assert.Nil(t, n.Outputs())
	assert.Len(t, n.Inputs(), 1)
}

func TestOutputNodeConnection(t *testing.T) {
	n := subgraph.NewOutputNode("Result", "float64")
	ext := &stubOutputPort{version: 7, name: "Value"}

	in := n.Inputs()[subgraph.ValuePortName].(nodes.SingleValueInputPort)
	require.NoError(t, in.Set(ext))
	assert.Equal(t, ext, n.ConnectedSource())
	assert.Equal(t, 7, n.ConnectedSource().Version())

	in.Clear()
	assert.Nil(t, n.ConnectedSource())
}

func TestOutputNodeJSONRoundtrip(t *testing.T) {
	n := subgraph.NewOutputNode("Mesh", "github.com/EliCDavis/polyform/generator/manifest.Manifest")
	data, err := n.ToJSON(&jbtf.Encoder{})
	require.NoError(t, err)

	decoder, err := jbtf.NewDecoder([]byte(`{}`))
	require.NoError(t, err)

	restored := subgraph.NewOutputNode("", "")
	require.NoError(t, restored.FromJSON(decoder, data))

	assert.Equal(t, "Mesh", restored.PortName())
	assert.Equal(t, "github.com/EliCDavis/polyform/generator/manifest.Manifest", restored.PortType())
}

func TestIsBoundaryNode(t *testing.T) {
	input := subgraph.NewInputNode("In", "float64")
	output := subgraph.NewOutputNode("Out", "float64")

	b, ok := subgraph.IsBoundaryNode(input)
	assert.True(t, ok)
	assert.Equal(t, "In", b.PortName())

	_, ok = subgraph.IsBoundaryNode(output)
	assert.True(t, ok)

	ib, ok := subgraph.IsInputBoundary(input)
	assert.True(t, ok)
	assert.NotNil(t, ib)

	_, ok = subgraph.IsInputBoundary(output)
	assert.False(t, ok)
}

func TestInputNodeFromJSONInvalid(t *testing.T) {
	n := subgraph.NewInputNode("X", "float64")
	decoder, err := jbtf.NewDecoder([]byte(`{}`))
	require.NoError(t, err)
	err = n.FromJSON(decoder, []byte(`{not json}`))
	assert.Error(t, err)
}

func TestBoundaryDataJSONShape(t *testing.T) {
	data, err := json.Marshal(map[string]string{
		"portName": "Foo",
		"portType": "int",
	})
	require.NoError(t, err)

	decoder, err := jbtf.NewDecoder([]byte(`{}`))
	require.NoError(t, err)

	n := subgraph.NewInputNode("", "")
	require.NoError(t, n.FromJSON(decoder, data))
	assert.Equal(t, "Foo", n.PortName())
	assert.Equal(t, "int", n.PortType())
}

func TestCopyKeepsPortNameTypeAndDescription(t *testing.T) {
	in := subgraph.NewInputNode("Scale", "float64")
	in.SetPortDescription("Uniform scale")
	out := subgraph.NewOutputNode("Mesh", "int")
	out.SetPortDescription("The result")

	for original, copied := range map[subgraph.Boundary]subgraph.Boundary{in: in.Copy(), out: out.Copy()} {
		assert.Equal(t, original.PortName(), copied.PortName())
		assert.Equal(t, original.PortType(), copied.PortType())
		assert.Equal(t, original.PortDescription(), copied.PortDescription())
	}
}

func TestCopyLeavesAnUnnamedBoundaryUnnamed(t *testing.T) {
	assert.False(t, subgraph.BoundaryPortNameConfigured(subgraph.NewInputNode("", "float64").Copy()))
	assert.False(t, subgraph.BoundaryPortNameConfigured(subgraph.NewOutputNode("", "float64").Copy()))
}

func TestCopyDoesNotShareConnections(t *testing.T) {
	in := subgraph.NewInputNode("Scale", "float64")
	defaultIn := in.Inputs()[subgraph.DefaultPortName].(nodes.SingleValueInputPort)
	require.NoError(t, defaultIn.Set(nodes.ConstOutput[float64]{Val: 2}))
	in.ReadFrom(fed{nodes.ConstOutput[float64]{Val: 3}})

	copied := in.Copy()
	assert.Nil(t, copied.Inputs()[subgraph.DefaultPortName].(nodes.SingleValueInputPort).Value())
	assert.Equal(t, 0., copied.Outputs()[subgraph.ValuePortName].(nodes.Output[float64]).Value())

	out := subgraph.NewOutputNode("Result", "float64")
	require.NoError(t, out.Inputs()[subgraph.ValuePortName].(nodes.SingleValueInputPort).Set(nodes.ConstOutput[float64]{Val: 1}))
	assert.Nil(t, out.Copy().ConnectedSource())
}

type fed struct{ port nodes.OutputPort }

func (f fed) Value() nodes.OutputPort { return f.port }
