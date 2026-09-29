package trig_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/trig"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

func floats(vals ...float64) nodes.Output[[]float64] {
	return nodes.ConstOutput[[]float64]{Val: vals}
}

func scalar(val float64) nodes.Output[float64] {
	return nodes.ConstOutput[float64]{Val: val}
}

func TestSinScalar(t *testing.T) {
	node := &nodes.Struct[trig.SinNode]{Data: trig.SinNode{
		Input: scalar(math.Pi / 2),
	}}
	got := nodes.GetNodeOutputPort[float64](node, "Out").Value()
	assert.InDelta(t, 1, got, 1e-12)
}

func TestSinArray(t *testing.T) {
	node := &nodes.Struct[trig.SinNode]{Data: trig.SinNode{
		Input: floats(0, math.Pi/2, math.Pi),
	}}
	got := nodes.GetNodeOutputPort[[]float64](node, "Out").Value()
	assert.InDeltaSlice(t, []float64{0, 1, 0}, got, 1e-12)
}

func TestAmplitudeBroadcastsAcrossArray(t *testing.T) {
	node := &nodes.Struct[trig.SinNode]{Data: trig.SinNode{
		Input:     floats(math.Pi/2, math.Pi/2),
		Amplitude: scalar(3),
	}}
	got := nodes.GetNodeOutputPort[[]float64](node, "Out").Value()
	assert.InDeltaSlice(t, []float64{3, 3}, got, 1e-12)
}

func TestArrayAmplitudeLiftsScalarInput(t *testing.T) {
	node := &nodes.Struct[trig.SinNode]{Data: trig.SinNode{
		Input:     scalar(math.Pi / 2),
		Amplitude: floats(1, 2, 3),
	}}
	got := nodes.GetNodeOutputPort[[]float64](node, "Out").Value()
	assert.InDeltaSlice(t, []float64{1, 2, 3}, got, 1e-12)
}

func TestShiftIsAPhaseShift(t *testing.T) {
	node := &nodes.Struct[trig.CosNode]{Data: trig.CosNode{
		Input: floats(0),
		Shift: scalar(math.Pi),
	}}
	got := nodes.GetNodeOutputPort[[]float64](node, "Out").Value()
	assert.InDelta(t, -1, got[0], 1e-12)
}

func TestArcTan2Array(t *testing.T) {
	node := &nodes.Struct[trig.ArcTan2Node]{Data: trig.ArcTan2Node{
		Y: floats(0, 1, 0, -1),
		X: floats(1, 0, -1, 0),
	}}
	got := nodes.GetNodeOutputPort[[]float64](node, "Out").Value()
	assert.InDeltaSlice(t, []float64{0, math.Pi / 2, math.Pi, -math.Pi / 2}, got, 1e-12)
}

func TestArcTan2MismatchedLengths(t *testing.T) {
	node := &nodes.Struct[trig.ArcTan2Node]{Data: trig.ArcTan2Node{
		Y: floats(0, 1, 0, -1),
		X: floats(1, 0, -1),
	}}
	port := nodes.GetNodeOutputPort[[]float64](node, "Out")
	assert.Empty(t, port.Value())
	assert.Equal(
		t,
		[]string{"arrays of 4 and 3 cannot be combined; they have to be the same length, or of length 1 to apply to every element"},
		port.(nodes.ObservableExecution).ExecutionReport().Errors,
	)
}

func TestDegreesToRadians(t *testing.T) {
	node := &nodes.Struct[trig.DegreesToRadiansNode]{Data: trig.DegreesToRadiansNode{
		Input: floats(0, 90, 180),
	}}
	got := nodes.GetNodeOutputPort[[]float64](node, "Out").Value()
	assert.InDeltaSlice(t, []float64{0, math.Pi / 2, math.Pi}, got, 1e-12)
}
