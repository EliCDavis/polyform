package extrude_test

import (
	"testing"

	"github.com/EliCDavis/polyform/arrays"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/csg"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextNodeLabelsEachStringAndTheLabelsPlaceAndCombine(t *testing.T) {
	names := []string{"TX1", "RX0", "D2"}
	labels := &nodes.Struct[extrude.TextNode]{Data: extrude.TextNode{
		Text: nodes.ConstOutput[[]string]{Val: names},
	}}
	placed := &nodes.Struct[meshops.TransformNode]{Data: meshops.TransformNode{
		Mesh: nodes.GetNodeOutputPort[[]modeling.Mesh](labels, "Out"),
		TRS: nodes.ConstOutput[[]trs.TRS]{Val: []trs.TRS{
			trs.Position(vector3.New(0., 0., 0.)),
			trs.Position(vector3.New(10., 0., 0.)),
			trs.Position(vector3.New(20., 0., 0.)),
		}},
	}}
	spread := &nodes.Struct[arrays.SpreadNode]{}
	require.NoError(t, spread.Inputs()["Array"].(nodes.SingleValueInputPort).Set(placed.Outputs()["Out"]))

	combined := nodes.GetNodeOutputPort[modeling.Mesh](&nodes.Struct[meshops.CombineNode]{Data: meshops.CombineNode{
		Meshes: []nodes.Output[modeling.Mesh]{spread.Outputs()["Out"].(nodes.Output[modeling.Mesh])},
	}}, "Out").Value()

	triangles := 0
	for _, name := range names {
		mesh, err := extrude.Text{Text: name, Height: 1, CurveSegments: 6}.Mesh()
		require.NoError(t, err)
		triangles += mesh.PrimitiveCount()
	}
	assert.Equal(t, triangles, combined.PrimitiveCount())
	assert.Greater(t, combined.BoundingBox(modeling.PositionAttribute).Max().X(), 20.)
}

func TestTextCapitalsStandAtTheirHeightOnTheBaseline(t *testing.T) {
	mesh, err := extrude.Text{Text: "HE", Height: 2}.Mesh()
	require.NoError(t, err)

	box := mesh.BoundingBox(modeling.PositionAttribute)
	assert.InDelta(t, 0, box.Min().Y(), 1e-6)
	assert.InDelta(t, 2, box.Max().Y(), 1e-6)
	assert.InDelta(t, 0, box.Size().Z(), 1e-9, "no depth leaves a flat face")
}

func TestTextWithDepthIsAClosedSolid(t *testing.T) {
	mesh, err := extrude.Text{Text: "PCB 470µF", Height: 1, Depth: 0.1}.Mesh()
	require.NoError(t, err)

	require.NoError(t, csg.CheckClosed(mesh))
	assert.InDelta(t, 0.1, mesh.BoundingBox(modeling.PositionAttribute).Size().Z(), 1e-9)
}

func TestTextLeavesTheCounterOfAnOOpen(t *testing.T) {
	solid, err := extrude.Text{Text: "I", Height: 1, Depth: 1}.Mesh()
	require.NoError(t, err)
	ring, err := extrude.Text{Text: "O", Height: 1, Depth: 1}.Mesh()
	require.NoError(t, err)

	box := ring.BoundingBox(modeling.PositionAttribute).Size()
	assert.Less(t, signedVolume(ring), 0.5*box.X()*box.Y(), "the hole is not filled")
	assert.Positive(t, signedVolume(solid))
}

func TestTextAlignsEachLineAndStacksThemDownward(t *testing.T) {
	centered, err := extrude.Text{Text: "WIDE\nI", Height: 1, Align: "center"}.Mesh()
	require.NoError(t, err)
	box := centered.BoundingBox(modeling.PositionAttribute)
	assert.InDelta(t, 0, box.Center().X(), 0.1)
	assert.Less(t, box.Min().Y(), -1., "the second line sits below the first")

	right, err := extrude.Text{Text: "I", Height: 1, Align: "right"}.Mesh()
	require.NoError(t, err)
	assert.LessOrEqual(t, right.BoundingBox(modeling.PositionAttribute).Max().X(), 0.)
}

func TestTextNamesWhatItCannotSet(t *testing.T) {
	mesh, err := extrude.Text{Text: "A͸B", Height: 1}.Mesh()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "͸")
	assert.Positive(t, mesh.PrimitiveCount(), "the rest is still set")

	_, err = extrude.Text{Text: "A", Height: 1, Font: "comic"}.Mesh()
	assert.ErrorContains(t, err, `unknown font "comic"`)
}
