package pattern_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/drawing/texturing/pattern"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type uvTexture = texturing.Texture[vector2.Float64]

func plainUV(w, h int) nodes.Output[uvTexture] {
	return nodes.ConstOutput[uvTexture]{Val: texturing.UVs(w, h)}
}

func number(v float64) nodes.Output[float64] {
	return nodes.ConstOutput[float64]{Val: v}
}

func texture[R, N any](t *testing.T, data N, port string) texturing.Texture[R] {
	t.Helper()
	out := nodes.GetNodeOutputPort[texturing.Texture[R]](&nodes.Struct[N]{Data: data}, port)
	value := out.Value()
	if observed, ok := out.(nodes.ObservableExecution); ok {
		require.Empty(t, observed.ExecutionReport().Errors)
	}
	return value
}

func TestCheckerAlternatesPerCell(t *testing.T) {
	tex := texture[float64](t, pattern.CheckerNode{UV: plainUV(4, 4), Columns: number(4), Rows: number(4)}, "Value")

	assert.Equal(t, 0., tex.Get(0, 0))
	assert.Equal(t, 1., tex.Get(1, 0))
	assert.Equal(t, 1., tex.Get(0, 1))
}

func TestCrossedStripesCountsTheStripesUnderEachPixel(t *testing.T) {
	tex := texture[float64](t, pattern.CrossedStripesNode{UV: plainUV(8, 8), Columns: number(4), Rows: number(4), Width: number(0.5)}, "Value")

	assert.Equal(t, 1., tex.Get(0, 0))
	assert.Equal(t, 0.5, tex.Get(1, 0))
	assert.Equal(t, 0., tex.Get(1, 1))
}

func TestCellGivesIndicesAndLocalUV(t *testing.T) {
	node := pattern.CellNode{UV: plainUV(4, 4), Columns: number(4), Rows: number(4)}

	assert.Equal(t, 1., texture[float64](t, node, "Column").Get(1, 2))
	assert.Equal(t, 2., texture[float64](t, node, "Row").Get(1, 2))
	assert.Equal(t, vector2.New(0.5, 0.5), texture[vector2.Float64](t, node, "Local UV").Get(1, 2))
}

func TestStripesTurnWithAngle(t *testing.T) {
	tex := texture[float64](t, pattern.StripesNode{UV: plainUV(1, 4), Count: number(2), Width: number(0.5), Angle: number(90)}, "Value")

	assert.Equal(t, 1., tex.Get(0, 0))
	assert.Equal(t, 0., tex.Get(0, 1))
}

func TestWeaveMasksTheStrandOnTop(t *testing.T) {
	node := pattern.WeaveNode{UV: plainUV(4, 4), Columns: number(4), Rows: number(4)}

	u := texture[float64](t, node, "Weft")
	v := texture[float64](t, node, "Warp")

	assert.Equal(t, 1., u.Get(0, 0))
	assert.Equal(t, 0., v.Get(0, 0))
	assert.Equal(t, 0., u.Get(1, 0))
	assert.Equal(t, 1., v.Get(1, 0))
}

func TestPatternsTakeTheSizeOfTheirUVs(t *testing.T) {
	tex := texture[float64](t, pattern.CheckerNode{UV: plainUV(3, 5)}, "Value")

	assert.Equal(t, 3, tex.Width())
	assert.Equal(t, 5, tex.Height())
}

func TestPatternWithoutUVsSaysSo(t *testing.T) {
	out := nodes.GetNodeOutputPort[texturing.Texture[float64]](&nodes.Struct[pattern.CheckerNode]{}, "Value")
	out.Value()

	assert.NotEmpty(t, out.(nodes.ObservableExecution).ExecutionReport().Errors)
}

func TestTransformUVScalesAroundTheCentre(t *testing.T) {
	tex := texture[vector2.Float64](t, pattern.TransformUVNode{
		In:    plainUV(2, 2),
		Scale: nodes.ConstOutput[vector2.Float64]{Val: vector2.New(2., 2.)},
	}, "UV")

	assert.Equal(t, vector2.New(0., 0.), tex.Get(0, 0))
	assert.Equal(t, vector2.New(1., 1.), tex.Get(1, 1))
}

func TestZigzagShiftsVAtEachPeak(t *testing.T) {
	uv := nodes.ConstOutput[uvTexture]{Val: texturing.FromArray([]vector2.Float64{vector2.New(0., 0.5), vector2.New(0.5, 0.5)}, 2, 1)}

	tex := texture[vector2.Float64](t, pattern.ZigzagUVNode{In: uv, Count: number(1), Amplitude: number(0.1)}, "UV")

	assert.Equal(t, vector2.New(0., 0.5), tex.Get(0, 0))
	assert.InDelta(t, 0.6, tex.Get(1, 0).Y(), 1e-12)
}

func TestDisplaceUVMovesEachPixelByItsAmount(t *testing.T) {
	amount := nodes.ConstOutput[texturing.Texture[float64]]{Val: texturing.FromArray([]float64{0, 1}, 2, 1)}

	tex := texture[vector2.Float64](t, pattern.DisplaceUVNode{
		In:     plainUV(2, 1),
		Amount: amount,
		Shift:  nodes.ConstOutput[vector2.Float64]{Val: vector2.New(0.5, 0.25)},
	}, "UV")

	assert.Equal(t, vector2.New(0.25, 0.5), tex.Get(0, 0))
	assert.Equal(t, vector2.New(1.25, 0.75), tex.Get(1, 0))
}

func TestDisplaceUVRefusesAMismatchedAmount(t *testing.T) {
	out := nodes.GetNodeOutputPort[uvTexture](&nodes.Struct[pattern.DisplaceUVNode]{Data: pattern.DisplaceUVNode{
		In:     plainUV(2, 2),
		Amount: nodes.ConstOutput[texturing.Texture[float64]]{Val: texturing.FromArray([]float64{0}, 1, 1)},
	}}, "UV")
	out.Value()

	assert.NotEmpty(t, out.(nodes.ObservableExecution).ExecutionReport().Errors)
}

func TestCellRowOffsetSlidesEveryOtherRow(t *testing.T) {
	node := pattern.CellNode{UV: plainUV(4, 2), Columns: number(2), Rows: number(2), RowOffset: number(0.5)}

	columns := texture[float64](t, node, "Column")
	local := texture[vector2.Float64](t, node, "Local UV")

	assert.Equal(t, 0., columns.Get(0, 0))
	assert.Equal(t, 0.25, local.Get(0, 0).X())
	assert.Equal(t, 0., columns.Get(0, 1))
	assert.Equal(t, 0.75, local.Get(0, 1).X())
	assert.Equal(t, 1., columns.Get(1, 1))
}

func TestDistanceMeasuresFromCenter(t *testing.T) {
	tex := texture[float64](t, pattern.DistanceNode{
		UV:     plainUV(2, 1),
		Center: nodes.ConstOutput[vector2.Float64]{Val: vector2.New(0.25, 0.5)},
	}, "Value")

	assert.Equal(t, 0., tex.Get(0, 0))
	assert.Equal(t, 0.5, tex.Get(1, 0))
}

