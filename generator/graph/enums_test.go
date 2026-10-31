package graph_test

import (
	"strings"
	"testing"

	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func windowStyle(t *testing.T, instance *graph.Instance) string {
	t.Helper()
	id, err := instance.CreateEnum("Window Style", "How a window opens", []enum.Member{
		{Name: "Casement", Description: "Hinged at the side"},
		{Name: "Sash"},
		{Name: "Fixed"},
	})
	require.NoError(t, err)
	return id
}

func enumTestInstance() *graph.Instance {
	tf := &refutil.TypeFactory{}
	tf.RegisterBuilder(subgraph.InputNodeTypeKey, func() any { return subgraph.NewInputNode("", "") })
	tf.RegisterBuilder(subgraph.OutputNodeTypeKey, func() any { return subgraph.NewOutputNode("", "") })
	return graph.New(graph.Config{TypeFactory: tf})
}

func TestCreateEnumDeclaresAVariableTypeAndAParameterNode(t *testing.T) {
	instance := enumTestInstance()
	id := windowStyle(t, instance)
	assert.Equal(t, "WindowStyle", id)

	v, nodeType, err := instance.NewVariableOfType("Facade/Style", "enum/WindowStyle")
	require.NoError(t, err)
	assert.Equal(t, "Facade/Style", nodeType)
	ev := v.(*variable.EnumVariable)
	assert.Equal(t, "Casement", ev.GetValue())
	require.NoError(t, ev.SetValue("sash"))
	assert.Equal(t, enum.Choice{Set: "enum/WindowStyle", Name: "Sash", Index: 1}, ev.Choice())

	param, _, err := instance.CreateNode(graph.EnumParameterTypePath(id))
	require.NoError(t, err)
	port := param.Outputs()["Value"].(nodes.Typed)
	assert.Equal(t, "enum/WindowStyle", port.Type())
	assert.True(t, nodes.IsPortTypeKnown("enum/WindowStyle"))

	nodeType2, err := instance.EnumParameterNodeType(id)
	require.NoError(t, err)
	assert.Equal(t, "enum/WindowStyle", nodeType2.Outputs["Value"].Type)

	schema := instance.Schema()
	assert.Equal(t, "Window Style", schema.Enums[id].Name)
	assert.Equal(t, "enum/WindowStyle", schema.Variables.SubGroups["Facade"].Variables["Style"].Type)
}

func TestGraphEnumsSurviveSaveAndLoad(t *testing.T) {
	instance := enumTestInstance()
	id := windowStyle(t, instance)
	v, _, err := instance.NewVariableOfType("Style", "enum/WindowStyle")
	require.NoError(t, err)
	require.NoError(t, v.(*variable.EnumVariable).SetValue("Fixed"))
	paramNode, paramID, err := instance.CreateNode(graph.EnumParameterTypePath(id))
	require.NoError(t, err)
	require.NoError(t, paramNode.(*graph.EnumParameterNode).SetValue("Sash"))
	instance.SaveProfile("fixed")

	saved, err := instance.EncodeToAppSchema()
	require.NoError(t, err)
	assert.Contains(t, string(saved), `"enums"`)

	loaded := enumTestInstance()
	require.NoError(t, loaded.ApplyAppSchema(saved))

	reloaded, err := loaded.GetVariable("Style")
	require.NoError(t, err)
	assert.Equal(t, "Fixed", reloaded.(*variable.EnumVariable).GetValue())
	assert.Equal(t, "Sash", loaded.Node(paramID).(*graph.EnumParameterNode).Value())
	assert.Equal(t, "Window Style", loaded.Schema().Enums[id].Name)
}

func TestLoadingAnUnknownMemberIsRefused(t *testing.T) {
	instance := enumTestInstance()
	windowStyle(t, instance)
	_, _, err := instance.NewVariableOfType("Style", "enum/WindowStyle")
	require.NoError(t, err)
	saved, err := instance.EncodeToAppSchema()
	require.NoError(t, err)

	require.Contains(t, string(saved), `"value": "Casement"`)
	broken := []byte(strings.Replace(string(saved), `"value": "Casement"`, `"value": "Picture"`, 1))

	err = enumTestInstance().ApplyAppSchema(broken)
	assert.ErrorContains(t, err, `variable Style: "Picture" is not a member; expected Casement, Sash or Fixed`)
}

func TestUpdateEnumRenamesMembersEverywhereAndRefusesRemovingHeldOnes(t *testing.T) {
	instance := enumTestInstance()
	id := windowStyle(t, instance)
	v, _, err := instance.NewVariableOfType("Style", "enum/WindowStyle")
	require.NoError(t, err)
	require.NoError(t, v.(*variable.EnumVariable).SetValue("Sash"))
	instance.SaveProfile("victorian")
	paramNode, _, err := instance.CreateNode(graph.EnumParameterTypePath(id))
	require.NoError(t, err)
	require.NoError(t, paramNode.(*graph.EnumParameterNode).SetValue("Fixed"))

	err = instance.UpdateEnum(id, graph.EnumUpdate{
		Name:    "Window Style",
		Members: []enum.Member{{Name: "Casement"}, {Name: "Fixed"}},
	})
	assert.EqualError(t, err, `"Sash" still used by profile victorian, variable Style`)

	err = instance.UpdateEnum(id, graph.EnumUpdate{
		Name:    "Opening",
		Members: []enum.Member{{Name: "Casement"}, {Name: "Sliding"}, {Name: "Picture"}},
		Renamed: map[string]string{"Sash": "Sliding", "Fixed": "Picture"},
	})
	require.NoError(t, err)

	assert.Equal(t, "Sliding", v.(*variable.EnumVariable).GetValue())
	assert.Equal(t, "Picture", paramNode.(*graph.EnumParameterNode).Value())
	profile, err := instance.Profiles().Get("victorian")
	require.NoError(t, err)
	assert.JSONEq(t, `"Sliding"`, string(profile["Style"]))
	assert.Equal(t, "Opening", instance.Schema().Enums[id].Name)
	table, ok := instance.EnumTable("enum/" + id)
	require.True(t, ok)
	assert.Equal(t, []string{"Casement", "Sliding", "Picture"}, table.Names())
}

func TestDeleteEnumRefusesWhileAnythingHoldsIt(t *testing.T) {
	instance := enumTestInstance()
	id := windowStyle(t, instance)
	_, _, err := instance.NewVariableOfType("Style", "enum/WindowStyle")
	require.NoError(t, err)

	assert.EqualError(t, instance.DeleteEnum(id), `enum "Window Style" is still used by variable Style`)

	require.NoError(t, instance.DeleteVariable("Style"))
	require.NoError(t, instance.DeleteEnum(id))
	assert.Empty(t, instance.Enums())
	_, _, err = instance.NewVariableOfType("Style", "enum/WindowStyle")
	assert.ErrorContains(t, err, `this graph has no enum "WindowStyle"`)
}

func TestEnumVariableFeedsASubgraphPort(t *testing.T) {
	instance := enumTestInstance()
	windowStyle(t, instance)
	v, _, err := instance.NewVariableOfType("Style", "enum/WindowStyle")
	require.NoError(t, err)
	require.NoError(t, v.(*variable.EnumVariable).SetValue("Fixed"))

	require.NoError(t, instance.CreateSubGraph("Window", "Window", ""))
	inside, err := instance.SubGraphInstance("Window")
	require.NoError(t, err)
	_, inID, err := inside.CreateBoundaryNode(subgraph.InputNodeTypeKey, "enum/WindowStyle")
	require.NoError(t, err)
	require.NoError(t, inside.SetBoundaryNodeInfo(inID, "Style"))
	_, outID, err := inside.CreateBoundaryNode(subgraph.OutputNodeTypeKey, "enum/WindowStyle")
	require.NoError(t, err)
	require.NoError(t, inside.SetBoundaryNodeInfo(outID, "Same"))
	require.NoError(t, inside.ConnectNodes(inID, "Value", outID, "Value"))

	placement, placementID, err := instance.CreateNode(subgraph.RuntimeTypePath("Window"))
	require.NoError(t, err)
	_, refID, err := instance.CreateNode("Style")
	require.NoError(t, err)
	require.NoError(t, instance.ConnectNodes(refID, "Value", placementID, inID))

	out := nodes.GetNodeOutputPort[enum.Choice](placement, outID)
	assert.Equal(t, enum.Choice{Set: "enum/WindowStyle", Name: "Fixed", Index: 2}, out.Value())
}
