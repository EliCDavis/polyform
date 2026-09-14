package sfc_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/sfc"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHilbertIndexWalksTheOrder2Curve(t *testing.T) {
	want := [][2]uint32{
		{0, 0}, {1, 0}, {1, 1}, {0, 1},
		{0, 2}, {0, 3}, {1, 3}, {1, 2},
		{2, 2}, {2, 3}, {3, 3}, {3, 2},
		{3, 1}, {2, 1}, {2, 0}, {3, 0},
	}
	for d, cell := range want {
		assert.Equal(t, uint64(d), sfc.HilbertIndex(2, cell[0], cell[1]), "cell %v", cell)
		x, y := sfc.HilbertCell(2, uint64(d))
		assert.Equal(t, cell, [2]uint32{x, y}, "distance %d", d)
	}
}

func TestHilbertConsecutiveIndicesAreAdjacentCells(t *testing.T) {
	const order = 6
	px, py := sfc.HilbertCell(order, 0)
	for d := uint64(1); d < 1<<(2*order); d++ {
		x, y := sfc.HilbertCell(order, d)
		dx, dy := int(x)-int(px), int(y)-int(py)
		require.Equal(t, 1, dx*dx+dy*dy, "distance %d jumps from (%d,%d) to (%d,%d)", d, px, py, x, y)
		px, py = x, y
	}
}

func TestHilbertCellIsTheInverseOfHilbertIndex(t *testing.T) {
	const order = 32
	for _, cell := range [][2]uint32{{0, 0}, {1, 0}, {0xFFFFFFFF, 0xFFFFFFFF}, {123456789, 987654321}} {
		x, y := sfc.HilbertCell(order, sfc.HilbertIndex(order, cell[0], cell[1]))
		assert.Equal(t, cell, [2]uint32{x, y})
	}
}

func TestHilbert2DRoundTrips(t *testing.T) {
	encoder := sfc.Hilbert2D{
		Min:        vector2.New(-1., -1.),
		Max:        vector2.New(1., 1.),
		Resolution: 16,
	}

	for _, p := range []vector2.Float64{
		vector2.New(-1., -1.),
		vector2.New(1., 1.),
		vector2.New(0.25, -0.75),
	} {
		back := encoder.Decode(encoder.Encode(p))
		assert.InDelta(t, p.X(), back.X(), 1e-4)
		assert.InDelta(t, p.Y(), back.Y(), 1e-4)
	}

	assert.Equal(t, uint64(0), encoder.Encode(vector2.New(-5., -5.)), "clamped to the corner")
	assert.Equal(t, encoder.EncodeArray([]vector2.Float64{vector2.New(0.25, -0.75)})[0], encoder.Encode(vector2.New(0.25, -0.75)))
}
