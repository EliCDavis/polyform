package mcp_test

import (
	"strings"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	sdfLineType       = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sdf.LineNode]"
	multiplyFloatType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.MultiplyNode[float64]]"
	sdfTransformNode  = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sdf.TransformNode]"
	fieldType         = "github.com/EliCDavis/polyform/math/sample.Vec3ToFloat"
	meshType          = "github.com/EliCDavis/polyform/modeling.Mesh"
)

func findingsOfKind(out polyformmcp.CheckOutlineOutput, kind string) []string {
	var msgs []string
	for _, f := range out.Findings {
		if f.Kind == kind {
			msgs = append(msgs, f.Part+"/"+f.Child+": "+f.Message)
		}
	}
	return msgs
}


func TestCheckOutlineKeepsAcceptedFindingsOutOfTheOpenList(t *testing.T) {
	session := testSession(t)
	var set polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": dogOutline + `
accepted:
  - { kind: missing-part, part: tail, reason: the tail comes in the next pass }
`}, &set)
	require.True(t, set.Accepted, issueMessages(set.Issues))

	var out polyformmcp.CheckOutlineOutput
	callTool(t, session, "check_outline", map[string]any{}, &out)
	for _, m := range findingsOfKind(out, "missing-part") {
		assert.False(t, strings.HasPrefix(m, "tail/"), m)
	}
	require.Len(t, out.Accepted, 1)
	assert.Equal(t, "missing-part", out.Accepted[0].Kind)
	assert.Contains(t, out.Accepted[0].Message, "accepted (the tail comes in the next pass)")

	var refused polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": dogOutline + `
accepted:
  - { kind: too-many-nodes, reason: whatever }
`}, &refused)
	assert.False(t, refused.Accepted)
	assert.Contains(t, issueMessages(refused.Issues), "not a check_outline finding kind")
}

func TestPatchOutlineMergesRemovesAndRenames(t *testing.T) {
	session := testSession(t)
	var set polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": dogOutline}, &set)
	require.True(t, set.Accepted, issueMessages(set.Issues))

	var patched polyformmcp.SetOutlineOutput
	var got polyformmcp.GetOutlineOutput
	callTool(t, session, "patch_outline", map[string]any{"yaml": `
parts:
  pivot: { kind: helper }
accepted:
  - { kind: missing-part, part: torso, reason: built next pass }
`}, &patched)
	require.True(t, patched.Accepted, issueMessages(patched.Issues))
	assert.Empty(t, patched.Nodes)
	assert.NotZero(t, patched.NodeCount)
	callTool(t, session, "get_outline", map[string]any{}, &got)
	assert.Contains(t, got.Yaml, "pivot:")
	assert.Contains(t, got.Yaml, "built next pass")
	assert.Contains(t, got.Yaml, "muzzle:", "parts not mentioned are kept")

	var out polyformmcp.CheckOutlineOutput
	callTool(t, session, "check_outline", map[string]any{}, &out)
	for _, m := range findingsOfKind(out, "missing-part") {
		assert.False(t, strings.HasPrefix(m, "torso/"), m)
	}

	callTool(t, session, "patch_outline", map[string]any{"remove": []string{"pivot"}, "rename": map[string]string{"tail": "tail_part"}}, &patched)
	require.True(t, patched.Accepted, issueMessages(patched.Issues))
	callTool(t, session, "get_outline", map[string]any{}, &got)
	assert.NotContains(t, got.Yaml, "pivot")
	assert.Contains(t, got.Yaml, "tail_part:")
	assert.NotContains(t, got.Yaml, "\n  tail:", "the dog's child followed the rename")
}

func TestPatchOutlineNamesTheValueAnUnquotedCommaSplit(t *testing.T) {
	session := testSession(t)
	var set polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": dogOutline}, &set)
	require.True(t, set.Accepted, issueMessages(set.Issues))

	msg := callToolExpectingError(t, session, "patch_outline", map[string]any{"yaml": `
accepted:
  - { kind: missing-part, part: torso, reason: built from materials, color conversions and wrappers, not geometry }
`})

	assert.Contains(t, msg, `unquoted comma split "color conversions and wrappers"`)
	assert.NotContains(t, msg, "mcp.OutlineAcceptance")
}

func TestSetOutlineNamesAnUnknownKeyInSchemaTerms(t *testing.T) {
	session := testSession(t)

	msg := callToolExpectingError(t, session, "set_outline", map[string]any{"yaml": dogOutline + `
accepted:
  - { kind: missing-part, part: torso, because: built next pass }
`})

	assert.Contains(t, msg, `unknown key "because" in accepted entry`)
	assert.NotContains(t, msg, "mcp.")
}

func TestSetOutlineAWholeLeafBelowHalfDrawsNoSizeWarning(t *testing.T) {
	session := testSession(t)
	var out polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": `
version: 1
subject: lamp
style: mechanical
extent: 1
parts:
  lamp:
    material: brass
    children:
      base:  { count: 1, size: 0.4, thickness: 0.3, primitive: primitives.Cylinder, whole: true }
      shade: { count: 1, size: 0.45, thickness: 0.3, primitive: primitives.Cylinder, whole: true }
      stem:  { count: 1, size: 0.2, thickness: 0.1, primitive: primitives.Cylinder }
`}, &out)
	require.True(t, out.Accepted, issueMessages(out.Issues))
	for _, i := range out.Issues {
		assert.NotContains(t, i.Message, "fine for a core mass", "whole: true already answered that question: %s", i.Message)
	}
}

func TestCheckOutlineTreatsTheRootAsTheSubjectsAssembly(t *testing.T) {
	session := testSession(t)
	var set polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": dogOutline}, &set)
	require.True(t, set.Accepted, issueMessages(set.Issues))

	callTool(t, session, "create_subgraph", map[string]any{"id": "head", "name": "Head",
		"inputs":  []map[string]any{{"name": "Head Size", "type": "float64"}},
		"outputs": []map[string]any{{"name": "Field", "type": fieldType, "nodeId": "skull", "port": "Field"}},
		"nodes":   []map[string]any{{"alias": "skull", "type": sdfSphereType}}}, &polyformmcp.CreateSubgraphOutput{})
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "head"}, &polyformmcp.InstantiateSubgraphOutput{})
	// Placement math wired into the instance is plumbing; a sphere fed
	// by a multiply is geometry, and both get named.
	var made polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []map[string]any{
		{"alias": "size", "type": multiplyFloatType, "inputs": map[string]any{"Values": map[string]any{"elements": []map[string]any{{"value": "0.5"}, {"value": "2"}}}}},
		{"alias": "stray", "type": sdfSphereType, "inputs": map[string]any{"Radius": map[string]any{"nodeId": "size", "port": "Float"}}},
	}}, &made)

	var out polyformmcp.CheckOutlineOutput
	callTool(t, session, "check_outline", map[string]any{}, &out)
	for _, m := range findingsOfKind(out, "missing-part") {
		assert.False(t, strings.HasPrefix(m, "dog/"), "the root is the dog's assembly: %s", m)
	}
	assert.Equal(t, 2, out.Matched, "dog and head")
	clutter := findingsOfKind(out, "root-clutter")
	require.Len(t, clutter, 1)
	assert.Contains(t, clutter[0], "2 node(s)")
	assert.Contains(t, clutter[0], made.Nodes["stray"]+" (sdf.Sphere)")
	assert.Contains(t, clutter[0], made.Nodes["size"])
	assert.Equal(t, 2, out.Stats.RootGeometry)
}

func TestCheckOutlineCreditsCopiesMadeByTransformsAndChildScopes(t *testing.T) {
	session := testSession(t)
	var set polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": `
version: 1
subject: rig
style: organic
extent: 1
parts:
  rig:
    material: clay
    children:
      pair:  { count: 1, size: 0.5 }
      twins: { count: 1, size: 0.5 }
  pair:
    children:
      bead: { count: 2, size: 0.3, thickness: 1, layout: free, primitive: sdf.Sphere }
      core: { count: 1, size: 0.5, thickness: 1, primitive: sdf.Sphere }
  twins:
    children:
      bead: { count: 2, size: 0.3, thickness: 1, layout: mirror-x, primitive: sdf.Sphere }
      core: { count: 1, size: 0.5, thickness: 1, primitive: sdf.Sphere }
`}, &set)
	require.True(t, set.Accepted, issueMessages(set.Issues))

	// One sphere placed twice by two transforms.
	callTool(t, session, "create_subgraph", map[string]any{"id": "pair", "name": "Pair",
		"nodes": []map[string]any{
			{"alias": "core", "type": sdfSphereType},
			{"alias": "bead", "type": sdfSphereType},
			{"alias": "left", "type": sdfTransformNode, "inputs": map[string]any{"Field": map[string]any{"nodeId": "bead", "port": "Field"}}},
			{"alias": "right", "type": sdfTransformNode, "inputs": map[string]any{"Field": map[string]any{"nodeId": "bead", "port": "Field"}}},
		}}, &polyformmcp.CreateSubgraphOutput{})
	// One sphere mirrored inside an unplanned helper the part instantiates.
	callTool(t, session, "create_subgraph", map[string]any{"id": "spread", "name": "Spread",
		"outputs": []map[string]any{{"name": "Field", "type": fieldType, "nodeId": "mirror", "port": "X"}},
		"nodes": []map[string]any{
			{"alias": "bead", "type": sdfSphereType},
			{"alias": "mirror", "type": sdfMirrorType, "inputs": map[string]any{"Field": map[string]any{"nodeId": "bead", "port": "Field"}}},
		}}, &polyformmcp.CreateSubgraphOutput{})
	callTool(t, session, "create_subgraph", map[string]any{"id": "twins", "name": "Twins",
		"nodes": []map[string]any{{"alias": "core", "type": sdfSphereType}}}, &polyformmcp.CreateSubgraphOutput{})
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "spread", "scope": "twins"}, &polyformmcp.InstantiateSubgraphOutput{})

	var out polyformmcp.CheckOutlineOutput
	callTool(t, session, "check_outline", map[string]any{}, &out)
	assert.Empty(t, findingsOfKind(out, "count-mismatch"))
	assert.Empty(t, findingsOfKind(out, "missing-leaf"))
}

func TestCheckOutlineBudgetsAUnionByTheLeavesItBlends(t *testing.T) {
	session := testSession(t)
	var set polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": dogOutline}, &set)
	require.True(t, set.Accepted, issueMessages(set.Issues))

	// Two chest-sized spheres blended at 0.05 swallow nothing; the same
	// radius over a tail-tip-sized line swallows the tip. The old rule
	// charged the tip to both.
	var made polyformmcp.CreateSubgraphOutput
	callTool(t, session, "create_subgraph", map[string]any{
		"id": "dog", "name": "Dog",
		"nodes": []map[string]any{
			{"alias": "chest", "type": sdfSphereType, "inputs": map[string]any{"Radius": map[string]any{"value": "0.126"}}},
			{"alias": "hips", "type": sdfSphereType, "inputs": map[string]any{"Radius": map[string]any{"value": "0.1"}}},
			{"alias": "body", "type": sdfSmoothUnionNodeType, "inputs": map[string]any{
				"Fields": map[string]any{"elements": []map[string]any{{"nodeId": "chest", "port": "Field"}, {"nodeId": "hips", "port": "Field"}}},
				"Radius": map[string]any{"value": "0.05"},
			}},
			{"alias": "tip", "type": sdfLineType, "inputs": map[string]any{"Radius": map[string]any{"value": "0.0075"}}},
			{"alias": "withTail", "type": sdfSmoothUnionNodeType, "inputs": map[string]any{
				"Fields": map[string]any{"elements": []map[string]any{{"nodeId": "body", "port": "Union"}, {"nodeId": "tip", "port": "Field"}}},
				"Radius": map[string]any{"value": "0.05"},
			}},
			{"alias": "march", "type": marchNodeType, "inputs": map[string]any{
				"Field":      map[string]any{"nodeId": "withTail", "port": "Union"},
				"Resolution": map[string]any{"value": "300"},
			}},
		},
	}, &made)

	var out polyformmcp.CheckOutlineOutput
	callTool(t, session, "check_outline", map[string]any{}, &out)
	wide := findingsOfKind(out, "blend-too-wide")
	require.Len(t, wide, 1, "%v", out.Findings)
	assert.Contains(t, wide[0], made.Nodes["withTail"])
	assert.Contains(t, wide[0], "dog/tail/tip")
	assert.Empty(t, findingsOfKind(out, "under-resolved"), "300 spans the 0.015 tip")
}

func TestCheckOutlineJudgesAMarchInItsOwnSpace(t *testing.T) {
	session := testSession(t)
	var set polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": dogOutline}, &set)
	require.True(t, set.Accepted, issueMessages(set.Issues))

	// Built ten times too big and scaled down by the model that places
	// it: Resolution 30 here is 300 in the world, which spans the toe.
	callTool(t, session, "create_subgraph", map[string]any{
		"id": "dog", "name": "Dog",
		"outputs": []map[string]any{{"name": "Mesh", "type": meshType, "nodeId": "march", "port": "Mesh"}},
		"nodes": []map[string]any{
			{"alias": "blob", "type": sdfSphereType},
			{"alias": "march", "type": marchNodeType, "inputs": map[string]any{
				"Field":      map[string]any{"nodeId": "blob", "port": "Field"},
				"Resolution": map[string]any{"value": "30"},
			}},
		},
	}, &polyformmcp.CreateSubgraphOutput{})
	var inst polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "dog"}, &inst)
	callTool(t, session, "create_nodes", map[string]any{"nodes": []map[string]any{
		{"alias": "model", "type": gltfModelType, "inputs": map[string]any{
			"Mesh":  map[string]any{"nodeId": inst.NodeId, "port": "Mesh"},
			"Scale": map[string]any{"value": `{"x":0.1,"y":0.1,"z":0.1}`},
		}},
	}}, &polyformmcp.CreateNodesOutput{})

	var out polyformmcp.CheckOutlineOutput
	callTool(t, session, "check_outline", map[string]any{}, &out)
	assert.Empty(t, findingsOfKind(out, "under-resolved"), "%v", out.Findings)
}

func TestCheckOutlineLeavesMathOnlySubgraphsOutOfThePlan(t *testing.T) {
	session := testSession(t)
	var set polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": dogOutline}, &set)
	require.True(t, set.Accepted, issueMessages(set.Issues))

	callTool(t, session, "create_equation_subgraph", map[string]any{"id": "halve", "equation": "h = x / 2"}, &polyformmcp.CreateEquationSubgraphOutput{})

	var out polyformmcp.CheckOutlineOutput
	callTool(t, session, "check_outline", map[string]any{}, &out)
	assert.Empty(t, findingsOfKind(out, "extra-subgraph"))
}

func TestDeleteSubgraphDropsItsPartFromTheOutline(t *testing.T) {
	session := testSession(t)
	var set polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": dogOutline + `
  old_tail: { kind: helper }
`}, &set)
	require.True(t, set.Accepted, issueMessages(set.Issues))
	callTool(t, session, "create_subgraph", map[string]any{"id": "old_tail", "name": "Old tail"}, &polyformmcp.CreateSubgraphOutput{})
	callTool(t, session, "delete_subgraph", map[string]any{"id": "old_tail"}, &polyformmcp.DeleteSubgraphOutput{})

	var got polyformmcp.GetOutlineOutput
	callTool(t, session, "get_outline", map[string]any{}, &got)
	assert.NotContains(t, got.Yaml, "old_tail")
	assert.Contains(t, got.Yaml, "muzzle:")

	var out polyformmcp.CheckOutlineOutput
	callTool(t, session, "check_outline", map[string]any{}, &out)
	for _, m := range findingsOfKind(out, "missing-part") {
		assert.False(t, strings.HasPrefix(m, "old_tail/"), m)
	}
}
