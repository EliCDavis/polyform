package graph_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strconv"
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func boundary(t *testing.T, definition *graph.Graph, typeKey, portType, name string) string {
	t.Helper()
	_, id, err := definition.CreateBoundaryNode(typeKey, portType)
	require.NoError(t, err)
	require.NoError(t, definition.SetBoundaryNodeInfo(id, name))
	return id
}

// outer places inner; both add a parameter of their own to what they are fed.
func nestedAdders(t *testing.T) (root *graph.Instance, placed nodes.Node) {
	t.Helper()
	root = testInstanceWithSubGraphTypesExtended(t)

	adder := func(id string, own float64, build func(definition *graph.Graph, in string) (from, port string)) {
		require.NoError(t, root.CreateSubGraph(id, id, ""))
		definition, err := root.SubGraphInstance(id)
		require.NoError(t, err)

		in := boundary(t, definition, subgraph.InputNodeTypeKey, "float64", "In")
		out := boundary(t, definition, subgraph.OutputNodeTypeKey, "float64", "Out")
		_, param, err := definition.CreateNode("Float64")
		require.NoError(t, err)
		_, err = definition.UpdateParameter(param, []byte(strconv.FormatFloat(own, 'g', -1, 64)))
		require.NoError(t, err)
		_, sum, err := definition.CreateNode("Sum")
		require.NoError(t, err)

		from, port := build(definition, in)
		require.NoError(t, definition.ConnectNodes(from, port, sum, "Values.0"))
		require.NoError(t, definition.ConnectNodes(param, "Value", sum, "Values.1"))
		require.NoError(t, definition.ConnectNodes(sum, "Float", out, subgraph.ValuePortName))
	}

	adder("inner", 10, func(_ *graph.Graph, in string) (string, string) {
		return in, subgraph.ValuePortName
	})
	adder("outer", 100, func(definition *graph.Graph, in string) (string, string) {
		_, inner, err := definition.CreateNode(subgraph.RuntimeTypePath("inner"))
		require.NoError(t, err)
		require.NoError(t, definition.ConnectNodes(in, subgraph.ValuePortName, inner, "In"))
		return inner, "Out"
	})

	_, source, err := root.CreateNode("Float64")
	require.NoError(t, err)
	_, err = root.UpdateParameter(source, []byte(`1`))
	require.NoError(t, err)
	placed, placedID, err := root.CreateNode(subgraph.RuntimeTypePath("outer"))
	require.NoError(t, err)
	require.NoError(t, root.ConnectNodes(source, "Value", placedID, "In"))
	return root, placed
}

func TestAPlacedSubgraphIsACopyOfItsDefinition(t *testing.T) {
	root, placed := nestedAdders(t)
	require.NoError(t, graph.CopiesMatchDefinitions(root))
	assert.Equal(t, 111., nodes.GetNodeOutputPort[float64](placed, "Out").Value())
}

func TestALoadedSubgraphIsACopyOfItsDefinition(t *testing.T) {
	root, _ := nestedAdders(t)
	saved, err := root.EncodeToAppSchema()
	require.NoError(t, err)

	reloaded := testInstanceWithSubGraphTypesExtended(t)
	require.NoError(t, reloaded.ApplyAppSchema(saved))
	require.NoError(t, graph.CopiesMatchDefinitions(reloaded))

	for _, id := range reloaded.NodeIds() {
		if placed, ok := reloaded.Node(id).(*graph.SubgraphInstanceNode); ok {
			assert.Equal(t, 111., nodes.GetNodeOutputPort[float64](placed, "Out").Value())
		}
	}
}

func TestAParameterChangedInADefinitionReachesItsCopies(t *testing.T) {
	root, placed := nestedAdders(t)
	inner, err := root.SubGraphInstance("inner")
	require.NoError(t, err)

	for _, id := range inner.NodeIds() {
		if _, ok := inner.Node(id).(*parameter.Float64); ok {
			_, err := inner.UpdateParameter(id, []byte(`20`))
			require.NoError(t, err)
		}
	}

	require.NoError(t, graph.CopiesMatchDefinitions(root))
	assert.Equal(t, 121., nodes.GetNodeOutputPort[float64](placed, "Out").Value())
}

func TestANodeCreatedInADefinitionReachesItsCopies(t *testing.T) {
	root, placed := nestedAdders(t)
	inner, err := root.SubGraphInstance("inner")
	require.NoError(t, err)

	_, _, err = inner.CreateNode("Float64")
	require.NoError(t, err)
	_, _, err = inner.CreateBoundaryNode(subgraph.InputNodeTypeKey, "float64")
	require.NoError(t, err)

	require.NoError(t, graph.CopiesMatchDefinitions(root))
	assert.Equal(t, 111., nodes.GetNodeOutputPort[float64](placed, "Out").Value())
}

func TestAnImageInADefinitionReachesItsCopies(t *testing.T) {
	factory := &refutil.TypeFactory{}
	factory.RegisterBuilder(subgraph.InputNodeTypeKey, func() any { return subgraph.NewInputNode("", "") })
	factory.RegisterBuilder(subgraph.OutputNodeTypeKey, func() any { return subgraph.NewOutputNode("", "") })
	factory.RegisterBuilder("Image", func() any { return &parameter.Image{} })
	root := graph.New(graph.Config{TypeFactory: factory})

	require.NoError(t, root.CreateSubGraph("picture", "picture", ""))
	definition, err := root.SubGraphInstance("picture")
	require.NoError(t, err)
	_, param, err := definition.CreateNode("Image")
	require.NoError(t, err)

	pixels := image.NewRGBA(image.Rect(0, 0, 3, 2))
	pixels.Set(1, 1, color.RGBA{R: 200, G: 100, B: 50, A: 255})
	encoded := bytes.Buffer{}
	require.NoError(t, png.Encode(&encoded, pixels))
	_, err = definition.UpdateParameter(param, encoded.Bytes())
	require.NoError(t, err)

	placed, _, err := root.CreateNode(subgraph.RuntimeTypePath("picture"))
	require.NoError(t, err)

	copied := placed.(*graph.SubgraphInstanceNode).LiveGraph().Node(param).(*parameter.Image).Value()
	require.NotNil(t, copied)
	assert.Equal(t, pixels.Bounds(), copied.Bounds())
	assert.Equal(t, pixels.At(1, 1), color.RGBAModel.Convert(copied.At(1, 1)))
}
