package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	withColorType         = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sdf.WithColorNode]"
	smoothUnionColoredTyp = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sdf.SmoothUnionColoredNode]"
	applyColorFieldType   = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/marching.ApplyColorFieldNode]"
	coloredDistanceType   = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sdf.ColoredFieldDistanceNode]"
)

// A SmoothUnionColoredNode that only feeds ApplyColorFieldNode sets how
// softly one color fades into another. It never becomes geometry, so the
// march's thinnest-feature limit does not apply to it - holding it to that
// limit fired on every pass of the raccoon build from pass 5 on.
func TestCheckOutlineSparesAColorOnlyUnionTheBlendLimit(t *testing.T) {
	session := testSession(t)
	var set polyformmcp.SetOutlineOutput
	callTool(t, session, "set_outline", map[string]any{"yaml": dogOutline}, &set)
	require.True(t, set.Accepted, issueMessages(set.Issues))
	// The dog's thinnest feature is a 0.015 tail tip, so MaxBlend is tiny.
	require.InDelta(t, 0.0075, set.Stats.MaxBlend, 1e-9)

	callTool(t, session, "create_subgraph", map[string]any{
		"id": "tail", "name": "tail",
		"nodes": []map[string]any{
			{"alias": "blob", "type": sdfSphereType},
			{"alias": "paintA", "type": withColorType, "inputs": map[string]any{
				"Field": map[string]any{"nodeId": "blob", "port": "Field"},
				"Color": map[string]any{"value": `"#ff0000"`},
			}},
			{"alias": "paintB", "type": withColorType, "inputs": map[string]any{
				"Field": map[string]any{"nodeId": "blob", "port": "Field"},
				"Color": map[string]any{"value": `"#0000ff"`},
			}},
			// Radius 0.05 is far over the 0.0075 shape budget, but this
			// chain is only ever sampled for color.
			{"alias": "paint", "type": smoothUnionColoredTyp, "inputs": map[string]any{
				"Fields": map[string]any{"elements": []map[string]any{
					{"nodeId": "paintA", "port": "Out"},
					{"nodeId": "paintB", "port": "Out"},
				}},
				"Radius": map[string]any{"value": "0.05"},
			}},
			{"alias": "distance", "type": coloredDistanceType, "inputs": map[string]any{
				"Field": map[string]any{"nodeId": "paintA", "port": "Out"},
			}},
			{"alias": "march", "type": marchNodeType, "inputs": map[string]any{
				"Field":      map[string]any{"nodeId": "distance", "port": "Out"},
				"Resolution": map[string]any{"value": "300"},
			}},
			{"alias": "apply", "type": applyColorFieldType, "inputs": map[string]any{
				"Mesh":  map[string]any{"nodeId": "march", "port": "Mesh"},
				"Field": map[string]any{"nodeId": "paint", "port": "Union"},
			}},
		},
	}, &polyformmcp.CreateSubgraphOutput{})

	var out polyformmcp.CheckOutlineOutput
	callTool(t, session, "check_outline", map[string]any{}, &out)

	for _, f := range out.Findings {
		assert.NotEqual(t, "blend-too-wide", f.Kind,
			"a color-only union is not held to the shape budget: %s", f.Message)
	}
}
