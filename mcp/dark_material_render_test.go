package mcp_test

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tonalRange reports the darkest and brightest luminance among the model's
// pixels, ignoring the light background the buffer is cleared to.
func tonalRange(t *testing.T, path string) (min, max int, pixels int) {
	t.Helper()

	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	img, err := png.Decode(f)
	require.NoError(t, err)

	min, max = 255, 0
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			cr, cg, cb, _ := img.At(x, y).RGBA()
			r, g, bl := int(cr>>8), int(cg>>8), int(cb>>8)
			if r > 200 && g > 200 && bl > 220 {
				continue // cleared background
			}
			pixels++
			lum := (r*299 + g*587 + bl*114) / 1000
			if lum < min {
				min = lum
			}
			if lum > max {
				max = lum
			}
		}
	}
	return min, max, pixels
}

// A near-black material is what a deep-sea creature, a black car or a tyre
// actually is, and the preview exists so a model can judge whether the
// geometry is right. Shaded straight, an 0.09 albedo spans roughly 0.04 to
// 0.11 and comes back a featureless black silhouette - which sent real
// builds into long blind debug loops, re-rendering something they could
// not read while the same model looked fine in a glTF viewer.
func TestDarkMaterialRendersWithReadableForm(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	var sphere polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   sphereNodeType,
		"inputs": map[string]any{"Radius": map[string]any{"value": "1"}},
	}, &sphere)

	var material polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.MaterialNode]",
		"inputs": map[string]any{"Color": map[string]any{"value": `"#17161c"`}},
	}, &material)

	var model polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.ModelNode]",
		"inputs": map[string]any{
			"Mesh":     map[string]any{"nodeId": sphere.NodeId, "port": "Out"},
			"Material": map[string]any{"nodeId": material.NodeId, "port": "Out"},
		},
	}, &model)

	var manifest polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.ManifestNode]",
		"inputs": map[string]any{
			"Models": map[string]any{
				"elements": []any{map[string]any{"nodeId": model.NodeId, "port": "Out"}},
			},
		},
	}, &manifest)

	out := filepath.Join(t.TempDir(), "dark.png")
	var render polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId": manifest.NodeId, "outputPath": out,
	}, &render)

	min, max, pixels := tonalRange(t, out)
	require.Greater(t, pixels, 500, "expected a rendered sphere")
	t.Logf("dark sphere luminance: min=%d max=%d spread=%d", min, max, max-min)

	// What makes a dark surface readable is that the whole object sits in
	// a visible band rather than crushed against black: shaded straight,
	// this sphere measured min 13 / max 30, where neighbouring levels are
	// indistinguishable on screen. Both ends are asserted because either
	// alone can be satisfied by a broken render - a black silhouette has a
	// fine ratio between its two black values.
	assert.Greater(t, min, 25, "even the shadow side must sit above the black floor")
	assert.Greater(t, max, 50, "the lit side of a dark sphere must be clearly visible")
	assert.Greater(t, max-min, 20, "a dark sphere must still show a shading gradient")
}

// The same treatment must not wash the color out of a saturated surface.
func TestBrightMaterialKeepsItsColor(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	var sphere polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   sphereNodeType,
		"inputs": map[string]any{"Radius": map[string]any{"value": "1"}},
	}, &sphere)

	var material polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.MaterialNode]",
		"inputs": map[string]any{"Color": map[string]any{"value": `"#00d0a0"`}},
	}, &material)

	var model polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.ModelNode]",
		"inputs": map[string]any{
			"Mesh":     map[string]any{"nodeId": sphere.NodeId, "port": "Out"},
			"Material": map[string]any{"nodeId": material.NodeId, "port": "Out"},
		},
	}, &model)

	var manifest polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.ManifestNode]",
		"inputs": map[string]any{
			"Models": map[string]any{
				"elements": []any{map[string]any{"nodeId": model.NodeId, "port": "Out"}},
			},
		},
	}, &manifest)

	out := filepath.Join(t.TempDir(), "bright.png")
	var render polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId": manifest.NodeId, "outputPath": out,
	}, &render)

	f, err := os.Open(out)
	require.NoError(t, err)
	defer f.Close()
	img, _, err := image.Decode(f)
	require.NoError(t, err)

	// Find the most saturated model pixel and check the hue survived.
	bestGap, bestG, bestR := 0, 0, 0
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			cr, cg, cb, _ := img.At(x, y).RGBA()
			r, g, bl := int(cr>>8), int(cg>>8), int(cb>>8)
			if r > 200 && g > 200 && bl > 220 {
				continue
			}
			if g-r > bestGap {
				bestGap, bestG, bestR = g-r, g, r
			}
		}
	}
	assert.Greater(t, bestGap, 60,
		"a saturated green-cyan surface must not be tone mapped toward gray (green=%d red=%d)", bestG, bestR)
}
