package mcp_test

import (
	"fmt"
	"strings"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const dogOutline = `
version: 1
subject: dog
style: organic
extent: 1.0
parts:
  dog:
    inputs: { Body Scale: float64 }
    children:
      torso: { count: 1, size: 0.6 }
      head:  { count: 1, size: 0.4 }
      leg:   { count: 4, size: 0.35, layout: mirror-xz }
      tail:  { count: 1, size: 0.3 }
  tail:
    children:
      base: { count: 1, size: 0.6, thickness: 0.2, primitive: sdf.Line }
      tip:  { count: 1, size: 0.5, thickness: 0.1, primitive: sdf.Line }
  torso:
    children:
      chest: { count: 1, size: 0.6, thickness: 0.7, primitive: sdf.Sphere }
      hips:  { count: 1, size: 0.5, thickness: 0.7, primitive: sdf.Sphere }
  head:
    inputs: { Head Size: float64 }
    children:
      skull:  { count: 1, size: 0.8, thickness: 0.9, primitive: sdf.Sphere }
      muzzle: { count: 1, size: 0.4, thickness: 0.6, primitive: sdf.Sphere }
      eye:    { count: 2, size: 0.2, layout: mirror-x }
      ear:    { count: 2, size: 0.5, thickness: 0.3, layout: mirror-x, primitive: sdf.Sphere }
  eye:
    kind: mesh
    children:
      ball: { count: 1, size: 1.0, thickness: 1.0, primitive: primitives.UvSphere }
      iris: { count: 1, size: 0.5, thickness: 0.3, primitive: primitives.UvSphere }
  leg:
    children:
      upper: { count: 1, size: 0.5, thickness: 0.4, primitive: sdf.Sphere }
      paw:   { count: 1, size: 0.3 }
  paw:
    children:
      pad: { count: 1, size: 0.8, thickness: 0.6, primitive: sdf.Sphere }
      toe: { count: 4, size: 0.3, thickness: 0.5, layout: arc, primitive: sdf.Sphere }
`

func issueMessages(issues []polyformmcp.OutlineIssue) string {
	var b strings.Builder
	for _, i := range issues {
		b.WriteString(i.Level + " " + i.Part + "/" + i.Child + ": " + i.Message + "\n")
	}
	return b.String()
}

func TestSetOutlineAcceptsAWellFormedPlan(t *testing.T) {
	session := testSession(t)
	var out polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": dogOutline}, &out)

	require.True(t, out.Accepted, issueMessages(out.Issues))
	for _, i := range out.Issues {
		assert.Equal(t, "warning", i.Level)
	}
	assert.Equal(t, 7, out.Stats.Parts)
	assert.Equal(t, 3, out.Stats.MaxDepth, "dog/leg/paw/toe")
	assert.Greater(t, out.Stats.Repeated, 3)

	var got polyformmcp.GetOutlineOutput
	callTool(t, session, "get_outline", map[string]any{}, &got)
	assert.Equal(t, dogOutline, got.Yaml)

	byPath := map[string]polyformmcp.OutlineNodeStat{}
	for _, n := range out.Nodes {
		byPath[n.Path] = n
	}
	toe := byPath["dog/leg/paw/toe"]
	assert.InDelta(t, 0.35*0.3*0.3, toe.AbsoluteSize, 1e-9)
	assert.Contains(t, toe.Primitive, "sdf.SphereNode", "short primitive spelling resolves to the registered key")
}

func TestSetOutlineRejectsLargeLeaf(t *testing.T) {
	session := testSession(t)
	var out polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": `
version: 1
subject: cat
style: organic
extent: 0.5
parts:
  cat:
    children:
      body: { count: 1, size: 0.7, thickness: 0.5, primitive: sdf.Sphere }
      head: { count: 1, size: 0.4, thickness: 0.8, primitive: sdf.Sphere }
`}, &out)

	require.False(t, out.Accepted)
	msgs := issueMessages(out.Issues)
	assert.Contains(t, msgs, "error cat/body: cat/body is 70%")
	assert.Contains(t, msgs, "warning cat/head: cat/head is 40%")

	var got polyformmcp.GetOutlineOutput
	callTool(t, session, "get_outline", map[string]any{}, &got)
	assert.Empty(t, got.Yaml, "a rejected outline is not stored")
}

func TestSetOutlineRejectsMissingCountSizeLayoutAndPrimitive(t *testing.T) {
	session := testSession(t)
	var out polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": `
version: 1
subject: thing
style: mechanical
parts:
  thing:
    children:
      wheel: { count: 4, size: 0.2 }
      bolt:  { size: 0.02, primitive: primitives.Cylinder }
      ears:  { count: 1, primitive: sdf.Sphere }
      lever: { count: 1, size: 0.1, primitive: nope.Nothing }
  wheel:
    children:
      rim: { count: 1, size: 0.9, primitive: primitives.Torus }
`}, &out)

	require.False(t, out.Accepted)
	msgs := issueMessages(out.Issues)
	assert.Contains(t, msgs, "thing/wheel: count 4 needs a layout")
	assert.Contains(t, msgs, "thing/bolt: count is required")
	assert.Contains(t, msgs, "thing/ears: size is required")
	assert.Contains(t, msgs, "thing/lever: thickness is required")
	assert.Contains(t, msgs, "matches no registered node type")
	assert.Contains(t, msgs, "extent is required")
}

func TestSetOutlineReportsTheThinnestFeatureAndWarnsOnBoxesInOrganicParts(t *testing.T) {
	session := testSession(t)
	var out polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": `
version: 1
subject: cat
style: organic
extent: 0.5
parts:
  cat:
    children:
      torso: { count: 1, size: 0.6 }
      leg:   { count: 4, size: 0.4, layout: mirror-xz }
  torso:
    children:
      chest: { count: 1, size: 0.6, thickness: 0.7, primitive: sdf.RoundCube }
      hips:  { count: 1, size: 0.5, thickness: 0.7, primitive: sdf.Sphere }
  leg:
    children:
      upper: { count: 1, size: 0.6, thickness: 0.25, primitive: sdf.RoundedCone }
      lower: { count: 1, size: 0.5, thickness: 0.1, primitive: sdf.RoundedCone }
`}, &out)

	require.True(t, out.Accepted, issueMessages(out.Issues))
	assert.Contains(t, issueMessages(out.Issues), "warning torso/chest: sdf.RoundCube is a box-family primitive in an organic part")
	assert.Equal(t, "cat/leg/lower", out.Stats.ThinnestFeature)
	assert.InDelta(t, 0.5*0.4*0.5*0.1, out.Stats.ThinnestWorld, 1e-9)
	assert.InDelta(t, 400, out.Stats.MinResolution, 1e-9, "4 voxels across a 0.01 feature")
	assert.InDelta(t, 0.005, out.Stats.MaxBlend, 1e-9)
}

func TestCheckOutlineFlagsAMarchTooCoarseForThePlan(t *testing.T) {
	session := testSession(t)
	var set polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": dogOutline}, &set)
	require.True(t, set.Accepted, issueMessages(set.Issues))
	require.Equal(t, "dog/tail/tip", set.Stats.ThinnestFeature, "the iris is thinner but the eye is a mesh part, outside the march")
	require.InDelta(t, 267, set.Stats.MinResolution, 1e-9)

	callTool(t, session, "create_subgraph", map[string]any{
		"id": "dog", "name": "Dog",
		"nodes": []map[string]any{
			{"alias": "blob", "type": sdfSphereType},
			{"alias": "union", "type": sdfSmoothUnionNodeType, "inputs": map[string]any{
				"Fields": map[string]any{"elements": []map[string]any{{"nodeId": "blob", "port": "Field"}}},
				"Radius": map[string]any{"value": "0.05"},
			}},
			{"alias": "march", "type": marchNodeType, "inputs": map[string]any{
				"Field":      map[string]any{"nodeId": "union", "port": "Union"},
				"Resolution": map[string]any{"value": "120"},
			}},
		},
	}, &polyformmcp.CreateSubgraphOutput{})

	var out polyformmcp.CheckOutlineOutput
	callTool(t, session, "check_outline", map[string]any{}, &out)
	kinds := map[string][]string{}
	for _, f := range out.Findings {
		kinds[f.Kind] = append(kinds[f.Kind], f.Message)
	}
	require.Len(t, kinds["under-resolved"], 1, "%v", out.Findings)
	assert.Contains(t, kinds["under-resolved"][0], "Resolution >= 267")
	require.Len(t, kinds["blend-too-wide"], 1, "%v", out.Findings)
	assert.Contains(t, kinds["blend-too-wide"][0], "Radius 0.05")
}

func TestCheckOutlineScalesABlendRadiusByTheTransformsDownstreamOfIt(t *testing.T) {
	session := testSession(t)
	var set polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": dogOutline}, &set)
	require.True(t, set.Accepted, issueMessages(set.Issues))
	require.InDelta(t, 0.0075, set.Stats.MaxBlend, 1e-9)

	const trsNewType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/trs.NewNode]"
	const sdfTransformType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sdf.TransformNode]"

	// The head is built in a unit frame (Radius 0.5) and scaled to 0.01
	// on its way out, so its blend is 0.005 in the world: under budget.
	callTool(t, session, "create_subgraph", map[string]any{
		"id": "head", "name": "Head",
		"outputs": []map[string]any{{"name": "Field", "type": "github.com/EliCDavis/polyform/math/sample.Vec3ToFloat", "nodeId": "scaled", "port": "Result"}},
		"nodes": []map[string]any{
			{"alias": "skull", "type": sdfSphereType},
			{"alias": "union", "type": sdfSmoothUnionNodeType, "inputs": map[string]any{
				"Fields": map[string]any{"elements": []map[string]any{{"nodeId": "skull", "port": "Field"}}},
				"Radius": map[string]any{"value": "0.5"},
			}},
			{"alias": "trs", "type": trsNewType, "inputs": map[string]any{
				"Scale": map[string]any{"value": `{"x":0.01,"y":0.01,"z":0.01}`},
			}},
			{"alias": "scaled", "type": sdfTransformType, "inputs": map[string]any{
				"Field":     map[string]any{"nodeId": "union", "port": "Union"},
				"Transform": map[string]any{"nodeId": "trs", "port": "Out"},
			}},
		},
	}, &polyformmcp.CreateSubgraphOutput{})
	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "head"}, &placed)
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []map[string]any{
			{"alias": "body", "type": sdfSmoothUnionNodeType, "inputs": map[string]any{
				"Fields": map[string]any{"elements": []map[string]any{{"nodeId": placed.NodeId, "port": "Field"}}},
				"Radius": map[string]any{"value": "0.05"},
			}},
			{"alias": "march", "type": marchNodeType, "inputs": map[string]any{
				"Field":      map[string]any{"nodeId": "body", "port": "Union"},
				"Resolution": map[string]any{"value": "300"},
			}},
		},
	}, &polyformmcp.CreateNodesOutput{})

	var out polyformmcp.CheckOutlineOutput
	callTool(t, session, "check_outline", map[string]any{}, &out)
	var wide []string
	for _, f := range out.Findings {
		if f.Kind == "blend-too-wide" {
			wide = append(wide, f.Message)
		}
	}
	require.Len(t, wide, 1, "only the root union is over budget: %v", wide)
	assert.Contains(t, wide[0], "Radius 0.05")
}

func TestCheckOutlineBudgetsEachPartByItsOwnThinnestLeaf(t *testing.T) {
	session := testSession(t)
	var set polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": dogOutline}, &set)
	require.True(t, set.Accepted, issueMessages(set.Issues))

	for _, part := range []string{"torso", "tail"} {
		callTool(t, session, "create_subgraph", map[string]any{
			"id": part, "name": part,
			"nodes": []map[string]any{
				{"alias": "blob", "type": sdfSphereType},
				{"alias": "union", "type": sdfSmoothUnionNodeType, "inputs": map[string]any{
					"Fields": map[string]any{"elements": []map[string]any{{"nodeId": "blob", "port": "Field"}}},
					"Radius": map[string]any{"value": "0.05"},
				}},
			},
		}, &polyformmcp.CreateSubgraphOutput{})
	}

	var out polyformmcp.CheckOutlineOutput
	callTool(t, session, "check_outline", map[string]any{}, &out)
	var wide []string
	for _, f := range out.Findings {
		if f.Kind == "blend-too-wide" {
			wide = append(wide, f.Part)
		}
	}
	assert.Equal(t, []string{"tail"}, wide, "the torso's thinnest leaf is its 0.25 chest, not the tail tip")
}

func TestCheckOutlineCountsAMeshMirrorAsARepeater(t *testing.T) {
	session := testSession(t)
	var set polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": dogOutline}, &set)
	require.True(t, set.Accepted, issueMessages(set.Issues))

	const meshMirrorType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/meshops.MirrorNode]"
	callTool(t, session, "create_subgraph", map[string]any{
		"id": "paw", "name": "Paw",
		"nodes": []map[string]any{
			{"alias": "pad", "type": sdfSphereType},
			{"alias": "toe", "type": sdfSphereType},
			{"alias": "mirror", "type": meshMirrorType},
		},
	}, &polyformmcp.CreateSubgraphOutput{})

	var out polyformmcp.CheckOutlineOutput
	callTool(t, session, "check_outline", map[string]any{}, &out)
	for _, f := range out.Findings {
		assert.False(t, f.Kind == "count-mismatch" && f.Part == "paw", f.Message)
	}
}

func TestSetOutlineRejectsUndefinedChild(t *testing.T) {
	session := testSession(t)
	var out polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": `
version: 1
subject: robot
style: mechanical
parts:
  robot:
    children:
      arm: { count: 2, size: 0.4, layout: mirror-x }
`}, &out)
	require.False(t, out.Accepted)
	assert.Contains(t, issueMessages(out.Issues), "neither a part defined under parts nor a leaf")
}

func TestSetOutlineRejectsUnknownFields(t *testing.T) {
	session := testSession(t)
	msg := callToolExpectingError(t, session, "set_outline", map[string]any{"yaml": `
version: 1
subject: x
style: organic
objects: {}
`})
	assert.Contains(t, msg, "not valid YAML for this schema")
}

func TestCheckOutlineFindsMissingPartsInputsAndInlineChildren(t *testing.T) {
	session := testSession(t)
	var set polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": dogOutline}, &set)
	require.True(t, set.Accepted, issueMessages(set.Issues))

	// Build only the paw, with its toe repeated properly, and a head
	// missing its declared input and with eyes done inline.
	callTool(t, session, "create_subgraph", map[string]any{
		"id": "paw", "name": "Paw",
		"nodes": []map[string]any{
			{"alias": "pad", "type": sdfSphereType},
			{"alias": "toe", "type": sdfSphereType},
			{"alias": "mirror", "type": sdfMirrorType, "inputs": map[string]any{
				"Field": map[string]any{"nodeId": "toe", "port": "Field"},
			}},
		},
	}, &polyformmcp.CreateSubgraphOutput{})
	callTool(t, session, "create_subgraph", map[string]any{
		"id": "head", "name": "Head",
		"nodes": []map[string]any{
			{"alias": "skull", "type": sdfSphereType},
			{"alias": "muzzle", "type": sdfSphereType},
			{"alias": "ear", "type": sdfSphereType},
			{"alias": "ear2", "type": sdfSphereType},
		},
	}, &polyformmcp.CreateSubgraphOutput{})
	callTool(t, session, "create_subgraph", map[string]any{"id": "eye", "name": "Eye",
		"nodes": []map[string]any{{"type": sphereNodeType}, {"type": sphereNodeType}}}, &polyformmcp.CreateSubgraphOutput{})
	callTool(t, session, "create_subgraph", map[string]any{"id": "scratch", "name": "Scratch"}, &polyformmcp.CreateSubgraphOutput{})
	callTool(t, session, "create_node", map[string]any{"type": sdfSphereType}, &polyformmcp.CreateNodeOutput{})

	var out polyformmcp.CheckOutlineOutput
	callTool(t, session, "check_outline", map[string]any{}, &out)

	kinds := map[string][]string{}
	for _, f := range out.Findings {
		kinds[f.Kind] = append(kinds[f.Kind], f.Part+"/"+f.Child+": "+f.Message)
	}
	assert.Equal(t, 3, out.Matched)
	assert.Equal(t, 7, out.Planned)
	assert.Len(t, kinds["missing-part"], 4, "dog, torso, leg, tail: %v", kinds["missing-part"])
	assert.Len(t, kinds["extra-subgraph"], 1, "%v", kinds["extra-subgraph"])
	assert.Len(t, kinds["missing-input"], 1, "head declares Head Size: %v", kinds["missing-input"])
	assert.Len(t, kinds["inline-part"], 1, "head never instantiates eye: %v", kinds["inline-part"])
	assert.Empty(t, kinds["count-mismatch"], "paw's toe is mirrored, head's ear has 2 spheres: %v", kinds["count-mismatch"])
	assert.Len(t, kinds["root-clutter"], 1, "%v", kinds["root-clutter"])

	assert.Equal(t, 4, out.Stats.Subgraphs)
	assert.Equal(t, 1, out.Stats.RootGeometry)
}

func TestCheckOutlineWithoutOutlineIsError(t *testing.T) {
	session := testSession(t)
	msg := callToolExpectingError(t, session, "check_outline", map[string]any{})
	assert.Contains(t, msg, "set_outline")
}

func TestGraphStatsCountsNestingAndInstances(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_subgraph", map[string]any{"id": "toe", "name": "Toe",
		"nodes": []map[string]any{{"type": sdfSphereType}}}, &polyformmcp.CreateSubgraphOutput{})
	callTool(t, session, "create_subgraph", map[string]any{"id": "paw", "name": "Paw"}, &polyformmcp.CreateSubgraphOutput{})
	for i := 0; i < 4; i++ {
		callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "toe", "scope": "paw"}, &polyformmcp.InstantiateSubgraphOutput{})
	}
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "paw"}, &polyformmcp.InstantiateSubgraphOutput{})

	var stats polyformmcp.GraphStats
	callTool(t, session, "graph_stats", map[string]any{}, &stats)

	assert.Equal(t, 2, stats.Subgraphs)
	assert.Equal(t, 5, stats.Instances)
	assert.Equal(t, 4, stats.InstancesOf["toe"])
	assert.Equal(t, 2, stats.MaxNesting, "root -> paw -> toe")
	assert.Equal(t, 0, stats.RootGeometry)
	assert.Equal(t, 6, stats.TotalNodes)
	assert.InDelta(t, 5.0/6.0, stats.InSubgraphShare, 1e-9)
}

func TestSetOutlineAcceptsALargeLeafDeclaredWhole(t *testing.T) {
	session := testSession(t)
	var out polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": `
version: 1
subject: scene
style: organic
extent: 0.7
parts:
  scene:
    children:
      cushion: { count: 1, size: 0.9 }
      ball:    { count: 1, size: 0.15, thickness: 1, primitive: sdf.Sphere }
  cushion:
    children:
      slab: { count: 1, size: 1.0, thickness: 0.15, primitive: sdf.RoundCube, whole: true }
      rim:  { count: 1, size: 0.95, thickness: 0.1, primitive: sdf.RoundCube, whole: true }
`}, &out)

	require.True(t, out.Accepted, issueMessages(out.Issues))
	msgs := issueMessages(out.Issues)
	assert.Contains(t, msgs, "warning cushion/slab: scene/cushion/slab is 90%", "whole demotes the refusal to the core-mass warning")
}

func TestSetOutlinePaintLeavesNeedNoThicknessAndSetNoBudget(t *testing.T) {
	session := testSession(t)
	var out polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": `
version: 1
subject: cat
style: organic
extent: 0.5
parts:
  cat:
    children:
      body:   { count: 1, size: 0.6 }
      stripe: { count: 5, size: 0.3, layout: free, primitive: sdf.Sphere, paint: true }
      bib:    { count: 1, size: 0.8, primitive: sdf.Sphere, paint: true }
  body:
    children:
      chest: { count: 1, size: 0.6, thickness: 0.7, primitive: sdf.Sphere }
      hips:  { count: 1, size: 0.5, thickness: 0.7, primitive: sdf.Sphere }
`}, &out)

	require.True(t, out.Accepted, issueMessages(out.Issues))
	assert.Equal(t, "cat/body/hips", out.Stats.ThinnestFeature, "paint blobs set no voxel budget")
	assert.NotContains(t, issueMessages(out.Issues), "cat/bib", "an 80% paint blob is not a decomposition problem")
}

func TestCheckOutlineCountsLeavesInsideHelperInstances(t *testing.T) {
	session := testSession(t)
	var set polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": `
version: 1
subject: cat
style: organic
extent: 0.5
parts:
  cat:
    children:
      body: { count: 1, size: 0.45, thickness: 0.7, primitive: sdf.Sphere }
      tail: { count: 1, size: 0.5 }
  tail:
    children:
      curve: { count: 1, size: 1.0, thickness: 0.1, primitive: sdf.VaryingRadiusLines }
      tip:   { count: 1, size: 0.1, thickness: 1.0, primitive: sdf.Sphere }
`}, &set)
	require.True(t, set.Accepted, issueMessages(set.Issues))

	callTool(t, session, "create_tapered_curve_subgraph", map[string]any{"id": "tail_curve"}, &polyformmcp.CreateTaperedCurveSubgraphOutput{})
	callTool(t, session, "create_subgraph", map[string]any{"id": "tail", "name": "Tail",
		"nodes": []map[string]any{{"alias": "tip", "type": sdfSphereType}}}, &polyformmcp.CreateSubgraphOutput{})
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "tail_curve", "scope": "tail"}, &polyformmcp.InstantiateSubgraphOutput{})

	var out polyformmcp.CheckOutlineOutput
	callTool(t, session, "check_outline", map[string]any{}, &out)
	for _, f := range out.Findings {
		assert.NotEqual(t, "missing-leaf", f.Kind, "the curve lives inside the tail_curve helper: %s", f.Message)
	}
}

func TestSetOutlineSeparatelyMarchedLeafSetsNoSharedBudget(t *testing.T) {
	session := testSession(t)
	var out polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": `
version: 1
subject: cat
style: organic
extent: 0.5
parts:
  cat:
    children:
      body:   { count: 1, size: 0.45, thickness: 0.7, primitive: sdf.Sphere }
      head:   { count: 1, size: 0.4 }
  head:
    children:
      skull:  { count: 1, size: 0.8, thickness: 0.9, primitive: sdf.Sphere }
      tongue: { count: 1, size: 0.05, thickness: 0.3, primitive: sdf.Sphere, separate: true }
`}, &out)

	require.True(t, out.Accepted, issueMessages(out.Issues))
	assert.Equal(t, "cat/head/skull", out.Stats.ThinnestFeature, "the tongue has its own march, so the next-thinnest shared leaf sets the budget")
}

func TestSetOutlineResolvesGenericShortNamesToFloat64(t *testing.T) {
	session := testSession(t)
	var out polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": `
version: 1
subject: rig
style: mechanical
extent: 1
parts:
  rig:
    children:
      body:  { count: 1, size: 0.4, thickness: 0.5, primitive: primitives.Cube }
      scale: { count: 1, size: 0.1, thickness: 1, primitive: math.Multiply, paint: true }
      sum:   { count: 1, size: 0.1, thickness: 1, primitive: "math.Add[int]", paint: true }
`}, &out)

	require.True(t, out.Accepted, issueMessages(out.Issues))
	byPath := map[string]string{}
	for _, n := range out.Nodes {
		byPath[n.Path] = n.Primitive
	}
	assert.Contains(t, byPath["rig/scale"], "MultiplyNode[float64]", "bare generic name prefers float64")
	assert.Contains(t, byPath["rig/sum"], "AddNode[int]", "an explicit type parameter still resolves exactly")
}

func TestOutlineHelperPartsAreMatchedByNameAndExemptFromGeometryRules(t *testing.T) {
	session := testSession(t)
	var set polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": `
version: 1
subject: cat
style: organic
extent: 0.5
parts:
  cat:
    children:
      body: { count: 1, size: 0.45, thickness: 0.7, primitive: sdf.Sphere }
      tail: { count: 1, size: 0.5 }
  tail:
    children:
      curve: { count: 1, size: 1.0, thickness: 0.1, primitive: sdf.VaryingRadiusLines }
      tip:   { count: 1, size: 0.1, thickness: 1.0, primitive: sdf.Sphere }
      deg2rad: { count: 1 }
  tail_curve: { kind: helper }
  deg2rad:    { kind: helper }
`}, &set)
	require.True(t, set.Accepted, issueMessages(set.Issues))

	callTool(t, session, "create_tapered_curve_subgraph", map[string]any{"id": "tail_curve"}, &polyformmcp.CreateTaperedCurveSubgraphOutput{})
	callTool(t, session, "create_equation_subgraph", map[string]any{"id": "deg2rad", "equation": "r = radians(d)"}, &polyformmcp.CreateEquationSubgraphOutput{})
	callTool(t, session, "create_subgraph", map[string]any{"id": "tail", "name": "Tail",
		"nodes": []map[string]any{{"alias": "tip", "type": sdfSphereType}}}, &polyformmcp.CreateSubgraphOutput{})
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "tail_curve", "scope": "tail"}, &polyformmcp.InstantiateSubgraphOutput{})
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "deg2rad", "scope": "tail"}, &polyformmcp.InstantiateSubgraphOutput{})

	var out polyformmcp.CheckOutlineOutput
	callTool(t, session, "check_outline", map[string]any{}, &out)
	for _, f := range out.Findings {
		assert.NotEqual(t, "extra-subgraph", f.Kind, "helpers are planned: %s", f.Message)
		assert.NotEqual(t, "missing-leaf", f.Kind, "the curve lives in the helper: %s", f.Message)
	}
}

func TestCheckOutlineBudgetsAMarchByWhatFeedsIt(t *testing.T) {
	session := testSession(t)
	var set polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": dogOutline}, &set)
	require.True(t, set.Accepted, issueMessages(set.Issues))

	// Each part exposes a field; the marches sit in the root, which is
	// not a part of the plan. That is the case the old per-scope budget
	// got wrong: with no part to key on it fell back to the whole plan's
	// thinnest, so the tail's 0.015 tip set the resolution for the torso.
	placed := map[string]string{}
	for _, part := range []string{"torso", "tail"} {
		callTool(t, session, "create_subgraph", map[string]any{
			"id": part, "name": part,
			"outputs": []map[string]any{{"name": "Field", "type": "github.com/EliCDavis/polyform/math/sample.Vec3ToFloat", "nodeId": "blob", "port": "Field"}},
			"nodes":   []map[string]any{{"alias": "blob", "type": sdfSphereType}},
		}, &polyformmcp.CreateSubgraphOutput{})

		var inst polyformmcp.InstantiateSubgraphOutput
		callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": part}, &inst)
		placed[part] = inst.NodeId
	}

	var made polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []map[string]any{
		{"alias": "torsoMarch", "type": marchNodeType, "inputs": map[string]any{
			"Field":      map[string]any{"nodeId": placed["torso"], "port": "Field"},
			"Resolution": map[string]any{"value": "120"},
		}},
		{"alias": "tailMarch", "type": marchNodeType, "inputs": map[string]any{
			"Field":      map[string]any{"nodeId": placed["tail"], "port": "Field"},
			"Resolution": map[string]any{"value": "120"},
		}},
	}}, &made)

	var out polyformmcp.CheckOutlineOutput
	callTool(t, session, "check_outline", map[string]any{}, &out)
	var coarse []string
	for _, f := range out.Findings {
		if f.Kind == "under-resolved" {
			coarse = append(coarse, f.Child)
		}
	}
	assert.Equal(t, []string{made.Nodes["tailMarch"]}, coarse,
		"only the march that actually reaches the 0.015 tail tip is under-resolved")
}

// A tooth inside a head: thinness is a property of one leaf, but
// resolution is a property of the whole march, and the cost of conflating
// them is cubic.
const moonOutline = `
version: 1
subject: moon
style: organic
extent: 2.5
parts:
  moon:
    children:
      body:  { count: 1, size: 0.9, thickness: 0.8, whole: true, primitive: sdf.Sphere }
      brow:  { count: 2, size: 0.3, thickness: 0.4, layout: mirror-x, primitive: sdf.RoundCube }
      tooth: { count: 6, size: 0.12, thickness: 0.06, layout: line, primitive: sdf.RoundCube }
`

func TestCheckOutlineFlagsAThinFeatureSettingTheResolutionForAWholeBody(t *testing.T) {
	session := testSession(t)
	var set polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": moonOutline}, &set)
	require.True(t, set.Accepted, issueMessages(set.Issues))

	callTool(t, session, "create_subgraph", map[string]any{
		"id": "moon", "name": "moon",
		"nodes": []map[string]any{
			{"alias": "blob", "type": sdfSphereType},
			{"alias": "march", "type": marchNodeType, "inputs": map[string]any{
				"Field":      map[string]any{"nodeId": "blob", "port": "Field"},
				"Resolution": map[string]any{"value": "90"},
				"Domain":     map[string]any{"value": `{"center":{"x":0,"y":0,"z":0},"extents":{"x":1.25,"y":1.25,"z":1.25}}`},
			}},
		},
	}, &polyformmcp.CreateSubgraphOutput{})

	var out polyformmcp.CheckOutlineOutput
	callTool(t, session, "check_outline", map[string]any{}, &out)

	var split []string
	for _, f := range out.Findings {
		if f.Kind == "thin-feature-sets-resolution" {
			split = append(split, f.Message)
		}
	}
	require.Len(t, split, 1, "%v", out.Findings)
	assert.Contains(t, split[0], "moon/tooth", "names the leaf forcing the resolution")
	assert.Contains(t, split[0], "separate: true", "names the way out")
}

func TestCheckOutlineLeavesAnEvenlyDetailedMarchAlone(t *testing.T) {
	session := testSession(t)
	var set polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": dogOutline}, &set)
	require.True(t, set.Accepted, issueMessages(set.Issues))

	callTool(t, session, "create_subgraph", map[string]any{
		"id": "torso", "name": "torso",
		"nodes": []map[string]any{
			{"alias": "blob", "type": sdfSphereType},
			{"alias": "march", "type": marchNodeType, "inputs": map[string]any{
				"Field":      map[string]any{"nodeId": "blob", "port": "Field"},
				"Resolution": map[string]any{"value": "40"},
				"Domain":     map[string]any{"value": `{"center":{"x":0,"y":0,"z":0},"extents":{"x":0.5,"y":0.5,"z":0.5}}`},
			}},
		},
	}, &polyformmcp.CreateSubgraphOutput{})

	var out polyformmcp.CheckOutlineOutput
	callTool(t, session, "check_outline", map[string]any{}, &out)
	for _, f := range out.Findings {
		assert.NotEqual(t, "thin-feature-sets-resolution", f.Kind,
			"the torso's chest and hips are the same order of thickness: %s", f.Message)
	}
}

func TestSetOutlineIncludeNodesIsThreeWay(t *testing.T) {
	session := testSession(t)

	t.Run("omitted returns them for a small plan", func(t *testing.T) {
		var out polyformmcp.SetOutlineOutput
		callTool(t, session, "set_outline", map[string]any{"yaml": dogOutline}, &out)
		require.True(t, out.Accepted, issueMessages(out.Issues))
		assert.NotEmpty(t, out.Nodes)
		assert.Zero(t, out.NodeCount)
	})

	t.Run("false suppresses them even for a small plan", func(t *testing.T) {
		var out polyformmcp.SetOutlineOutput
		callTool(t, session, "set_outline", map[string]any{"yaml": dogOutline, "includeNodes": false}, &out)
		require.True(t, out.Accepted, issueMessages(out.Issues))
		assert.Empty(t, out.Nodes, "includeNodes:false must be able to turn the list off")
		assert.Greater(t, out.NodeCount, 0, "and report the count instead")
	})

	t.Run("true returns them", func(t *testing.T) {
		var out polyformmcp.SetOutlineOutput
		callTool(t, session, "set_outline", map[string]any{"yaml": dogOutline, "includeNodes": true}, &out)
		require.True(t, out.Accepted, issueMessages(out.Issues))
		assert.NotEmpty(t, out.Nodes)
	})
}

const perSideOutline = `
version: 1
subject: jaw
style: organic
extent: 1.0
parts:
  jaw:
    children:
      gum:   { count: 1, size: 0.8, thickness: 0.5, whole: true, primitive: sdf.RoundCube }
      tooth: { count: %d, size: 0.1, thickness: 0.4, layout: %s, primitive: sdf.RoundCube }
`

func TestSetOutlineAllowsSeveralCopiesPerMirroredSide(t *testing.T) {
	tests := map[string]struct {
		count  int
		layout string
		warn   bool
	}{
		"one per side":            {2, "mirror-x", false},
		"three per side":          {6, "mirror-x", false},
		"five per side":           {10, "mirror-x", false},
		"odd count on one axis":   {3, "mirror-x", true},
		"two axes, one each":      {4, "mirror-xz", false},
		"two axes, three each":    {12, "mirror-xz", false},
		"two axes, not divisible": {6, "mirror-xz", true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			session := testSession(t)
			var out polyformmcp.SetOutlineOutput
			callTool(t, session, "set_outline", map[string]any{
				"yaml": fmt.Sprintf(perSideOutline, tc.count, tc.layout),
			}, &out)
			require.True(t, out.Accepted, issueMessages(out.Issues))

			var mirrorWarn string
			for _, i := range out.Issues {
				if strings.Contains(i.Message, "mirror") {
					mirrorWarn = i.Message
				}
			}
			if tc.warn {
				assert.NotEmpty(t, mirrorWarn, "count %d on %s should warn", tc.count, tc.layout)
			} else {
				assert.Empty(t, mirrorWarn, "count %d on %s is legitimate", tc.count, tc.layout)
			}
		})
	}
}
