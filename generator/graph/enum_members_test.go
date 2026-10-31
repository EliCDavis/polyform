package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/nodes"
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
