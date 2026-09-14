package sfc

import (
	"github.com/EliCDavis/vector/vector2"
)

// Hilbert2D maps float64 points to their distance along a 2D Hilbert curve
type Hilbert2D struct {
	// Bounds define the space that will be mapped to the curve
	Min, Max vector2.Float64

	// Resolution determines the precision of the encoding (number of bits per dimension, max 32)
	Resolution uint
}

func (e Hilbert2D) Encode(point vector2.Float64) uint64 {
	maxVal := float64((uint64(1) << e.Resolution) - 1)
	normalized := point.Sub(e.Min).
		DivByVector(e.Max.Sub(e.Min)).
		Clamp(0, 1).
		Scale(maxVal)
	return HilbertIndex(e.Resolution, uint32(normalized.X()), uint32(normalized.Y()))
}

func (e Hilbert2D) EncodeArray(points []vector2.Float64) []uint64 {
	extents := e.Max.Sub(e.Min)
	maxVal := float64((uint64(1) << e.Resolution) - 1)

	results := make([]uint64, len(points))
	for i, p := range points {
		normalized := p.Sub(e.Min).DivByVector(extents).Clamp(0, 1).Scale(maxVal)
		results[i] = HilbertIndex(e.Resolution, uint32(normalized.X()), uint32(normalized.Y()))
	}
	return results
}

func (e Hilbert2D) Decode(index uint64) vector2.Float64 {
	x, y := HilbertCell(e.Resolution, index)
	maxVal := float64((uint64(1) << e.Resolution) - 1)
	return vector2.New(float64(x), float64(y)).
		DivByConstant(maxVal).
		MultByVector(e.Max.Sub(e.Min)).
		Add(e.Min)
}

// HilbertIndex is the distance along the curve of cell (x, y) on a grid
// 2^order cells wide
func HilbertIndex(order uint, x, y uint32) uint64 {
	var d uint64
	for s := uint32(uint64(1) << order >> 1); s > 0; s >>= 1 {
		var rx, ry uint32
		if x&s > 0 {
			rx = 1
		}
		if y&s > 0 {
			ry = 1
		}
		d += uint64(s) * uint64(s) * uint64((3*rx)^ry)
		x, y = hilbertRotate(s, x, y, rx, ry)
	}
	return d
}

// HilbertCell is the inverse of HilbertIndex
func HilbertCell(order uint, d uint64) (x, y uint32) {
	n := uint64(1) << order
	for s := uint64(1); s < n; s <<= 1 {
		rx := uint32(1 & (d / 2))
		ry := uint32(1 & (d ^ uint64(rx)))
		x, y = hilbertRotate(uint32(s), x, y, rx, ry)
		x += uint32(s) * rx
		y += uint32(s) * ry
		d /= 4
	}
	return x, y
}

func hilbertRotate(s, x, y, rx, ry uint32) (uint32, uint32) {
	if ry == 0 {
		if rx == 1 {
			x = s - 1 - x
			y = s - 1 - y
		}
		return y, x
	}
	return x, y
}
