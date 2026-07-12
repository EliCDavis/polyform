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

func (s *stubOutputPort) Node() nodes.Node  { return nil }
func (s *stubOutputPort) Name() string      { return s.name }
func (s *stubOutputPort) Version() int      { return s.version }
func (s *stubOutputPort) Type() string      { return "float64" }

func TestInputNodeDefaults(t *testing.T) {
	n := subgraph.NewInputNode("", "")
	assert.Equal(t, "Input", n.BoundaryPortName())
	assert.Equal(t, "", n.BoundaryPortType())
	assert.Equal(t, "Input", n.Name())
	assert.Nil(t, n.Inputs())
	assert.Len(t, n.Outputs(), 1)
	assert.Contains(t, n.Outputs(), subgraph.ValuePortName)
}

func TestInputNodeCustomPort(t *testing.T) {
	n := subgraph.NewInputNode("Scale", "float64")
	assert.Equal(t, "Scale", n.BoundaryPortName())
	assert.Equal(t, "float64", n.BoundaryPortType())

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
	n.SetExternalSource(nodes.ConstOutput[vector3.Float64]{Val: want})

	assert.Equal(t, want, typedOut.Value())
}

func TestInputNodeExternalSourceVersion(t *testing.T) {
	n := subgraph.NewInputNode("X", "float64")
	ext := &stubOutputPort{version: 42, name: "Value"}

	n.SetExternalSource(ext)
	assert.Equal(t, ext, n.ExternalSource())

	out := n.Outputs()[subgraph.ValuePortName]
	assert.Equal(t, 42, out.Version())

	ext.version = 99
	assert.Equal(t, 99, out.Version())
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

	assert.Equal(t, "Position", restored.PortName)
	assert.Equal(t, "github.com/EliCDavis/vector/vector3.Vector[float64]", restored.PortType)
	assert.Equal(t, "Position", restored.BoundaryPortName())
}

func TestOutputNodeDefaults(t *testing.T) {
	n := subgraph.NewOutputNode("", "")
	assert.Equal(t, "Output", n.BoundaryPortName())
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

	assert.Equal(t, "Mesh", restored.BoundaryPortName())
	assert.Equal(t, "github.com/EliCDavis/polyform/generator/manifest.Manifest", restored.BoundaryPortType())
}

func TestIsBoundaryNode(t *testing.T) {
	input := subgraph.NewInputNode("In", "float64")
	output := subgraph.NewOutputNode("Out", "float64")

	b, ok := subgraph.IsBoundaryNode(input)
	assert.True(t, ok)
	assert.Equal(t, "In", b.BoundaryPortName())

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
	assert.Equal(t, "Foo", n.BoundaryPortName())
	assert.Equal(t, "int", n.BoundaryPortType())
}
