package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/arrays"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnumMembersNodeListsTheMembersAndFollowsEdits(t *testing.T) {
	instance := enumTestInstance()
	id := windowStyle(t, instance)

	node, nodeID, err := instance.CreateNode(graph.EnumMembersTypePath(id))
	require.NoError(t, err)
	members := node.(*graph.EnumMembersNode)
	assert.Equal(t, "Window Style Members", members.Name())

	port := nodes.GetNodeOutputPort[[]enum.Choice](members, "Members")
	assert.Equal(t, "[]enum/WindowStyle", port.(nodes.Typed).Type())
	assert.Equal(t, []enum.Choice{
		{Set: "enum/WindowStyle", Name: "Casement", Index: 0},
		{Set: "enum/WindowStyle", Name: "Sash", Index: 1},
		{Set: "enum/WindowStyle", Name: "Fixed", Index: 2},
	}, port.Value())

	nodeType, err := instance.EnumMembersNodeType(id)
	require.NoError(t, err)
	assert.Equal(t, "[]enum/WindowStyle", nodeType.Outputs["Members"].Type)

	err = instance.DeleteEnum(id)
	assert.EqualError(t, err, `enum "Window Style" is still used by members `+nodeID)

	require.NoError(t, instance.UpdateEnum(id, graph.EnumUpdate{
		Name:    "Window Style",
		Members: []enum.Member{{Name: "Fixed"}, {Name: "Sash"}},
	}))
	assert.Equal(t, []string{"Fixed", "Sash"}, []string{port.Value()[0].Name, port.Value()[1].Name})

	saved, err := instance.EncodeToAppSchema()
	require.NoError(t, err)
	loaded := enumTestInstance()
	require.NoError(t, loaded.ApplyAppSchema(saved))
	assert.Len(t, loaded.Node(nodeID).(*graph.EnumMembersNode).Members(), 2)
}

func TestArrayNodesCarryGraphEnumValues(t *testing.T) {
	tf := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[arrays.ReverseNode]](tf)
	refutil.RegisterType[nodes.Struct[enum.MemberIndexNode]](tf)
	instance := graph.New(graph.Config{TypeFactory: tf})
	id := windowStyle(t, instance)

	_, members, err := instance.CreateNode(graph.EnumMembersTypePath(id))
	require.NoError(t, err)
	_, reverse, err := instance.CreateNode("github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/nodes/arrays.ReverseNode]")
	require.NoError(t, err)
	index, indexID, err := instance.CreateNode("github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/enum.MemberIndexNode]")
	require.NoError(t, err)
	require.NoError(t, instance.ConnectNodes(members, "Members", reverse, "Array"))
	require.NoError(t, instance.ConnectNodes(reverse, "Out", indexID, "Value"))

	instance.Evaluate(func() {
		assert.Equal(t, []int{2, 1, 0}, nodes.GetNodeOutputPort[[]int](index, "Out").Value())
	})
}

func TestForEachCarriesGraphEnumValuesFromAnArrayNode(t *testing.T) {
	tf := &refutil.TypeFactory{}
	tf.RegisterBuilder(subgraph.InputNodeTypeKey, func() any { return subgraph.NewInputNode("", "") })
	tf.RegisterBuilder(subgraph.OutputNodeTypeKey, func() any { return subgraph.NewOutputNode("", "") })
	refutil.RegisterType[nodes.Struct[arrays.ReverseNode]](tf)
	refutil.RegisterType[nodes.Struct[enum.MemberIndexNode]](tf)
	instance := graph.New(graph.Config{TypeFactory: tf})
	id := windowStyle(t, instance)

	require.NoError(t, instance.CreateSubGraph("position", "Position", ""))
	definition, err := instance.SubGraphInstance("position")
	require.NoError(t, err)
	_, styleID, err := definition.CreateBoundaryNode(subgraph.InputNodeTypeKey, "enum/"+id)
	require.NoError(t, err)
	_, indexID, err := definition.CreateNode("github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/enum.MemberIndexNode]")
	require.NoError(t, err)
	_, outID, err := definition.CreateBoundaryNode(subgraph.OutputNodeTypeKey, "int")
	require.NoError(t, err)
	require.NoError(t, definition.SetBoundaryNodeInfo(styleID, "Style"))
	require.NoError(t, definition.SetBoundaryNodeInfo(outID, "Index"))
	require.NoError(t, definition.ConnectNodes(styleID, subgraph.ValuePortName, indexID, "Value"))
	require.NoError(t, definition.ConnectNodes(indexID, "Out", outID, subgraph.ValuePortName))

	_, members, err := instance.CreateNode(graph.EnumMembersTypePath(id))
	require.NoError(t, err)
	_, reverse, err := instance.CreateNode("github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/nodes/arrays.ReverseNode]")
	require.NoError(t, err)
	placement, placementID, err := instance.CreateNode(subgraph.RuntimeTypePath("position"))
	require.NoError(t, err)
	_, err = instance.SetForEach(placementID, true)
	require.NoError(t, err)
	require.NoError(t, instance.ConnectNodes(members, "Members", reverse, "Array"))
	require.NoError(t, instance.ConnectNodes(reverse, "Out", placementID, "Style"))

	instance.Evaluate(func() {
		assert.Equal(t, []int{2, 1, 0}, nodes.GetNodeOutputPort[[]int](placement, "Index").Value())
	})
}

func TestEnumMembersNodeSchemaNamesItsRegisteredType(t *testing.T) {
	instance := enumTestInstance()
	id := windowStyle(t, instance)

	node, _, err := instance.CreateNode(graph.EnumMembersTypePath(id))
	require.NoError(t, err)

	assert.Equal(t, graph.EnumMembersTypePath(id), instance.NodeInstanceSchema(node).Type)
}
