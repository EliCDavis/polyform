package morton

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
)

// Encoder3D provides 3D Morton encoding/decoding functionality for float64 coordinates
type Encoder3D struct {
	// Bounds define the space that will be mapped to Morton codes
	Bounds geometry.AABB

	// Resolution determines the precision of the encoding (number of bits per dimension)
	Resolution uint
}

// Encode converts a 3D float64 point to a Morton code
func (m *Encoder3D) Encode(point vector3.Float64) uint64 {
	min := m.Bounds.Min()
	max := m.Bounds.Max()
	maxVal := (1 << m.Resolution) - 1

	// Normalize coordinates to [0, 1] range
	normalized := point.Sub(min).
		DivByVector(max.Sub(min)).
		Clamp(0, 1).
		Scale(float64(maxVal))

	// Interleave the bits
	return interleaveBits3D(
		uint64(normalized.X()),
		uint64(normalized.Y()),
		uint64(normalized.Z()),
	)
}

func (m *Encoder3D) EncodeArray(points []vector3.Float64) []uint64 {
	min := m.Bounds.Min()
	max := m.Bounds.Max()

	extents := max.Sub(min)
	maxVal := (1 << m.Resolution) - 1

	results := make([]uint64, len(points))

	for i, p := range points {
		normalized := p.Sub(min).DivByVector(extents).Clamp(0, 1).Scale(float64(maxVal))

		// Interleave the bits
		results[i] = interleaveBits3D(
			uint64(normalized.X()),
			uint64(normalized.Y()),
			uint64(normalized.Z()),
		)
	}

	return results
}

// Decode converts a Morton code back to a 3D float64 point
func (m *Encoder3D) Decode(morton uint64) vector3.Float64 {
	// Deinterleave the bits
	x, y, z := deinterleaveBits3D(morton, m.Resolution)

	maxVal := float64((uint64(1) << m.Resolution) - 1)

	min := m.Bounds.Min()
	max := m.Bounds.Max()

	return vector3.New(float64(x), float64(y), float64(z)).
		DivByConstant(maxVal).
		MultByVector(max.Sub(min)).
		Add(min)
}

// Decode converts a Morton code back to a 3D float64 point
func (m *Encoder3D) DecodeArray(mortons []uint64) []vector3.Float64 {
	maxVal := float64((uint64(1) << m.Resolution) - 1)
	min := m.Bounds.Min()
	max := m.Bounds.Max()
	extents := max.Sub(min)

	results := make([]vector3.Float64, len(mortons))
	for i, morton := range mortons {
		// Deinterleave the bits
		x, y, z := deinterleaveBits3D(morton, m.Resolution)

		// Convert back to normalized coordinates
		results[i] = vector3.New(float64(x), float64(y), float64(z)).
			DivByConstant(maxVal).
			MultByVector(extents).
			Add(min)
	}

	return results
}

// interleaveBits3D interleaves the bits of three integers for 3D Morton encoding
func interleaveBits3D(x, y, z uint64) uint64 {
	x = expandBits(x)
	y = expandBits(y)
	z = expandBits(z)

	return x | (y << 1) | (z << 2)
}

// deinterleaveBits3D extracts the individual coordinates from a Morton code
func deinterleaveBits3D(morton uint64, resolution uint) (uint64, uint64, uint64) {
	x := compactBits(morton)
	y := compactBits(morton >> 1)
	z := compactBits(morton >> 2)

	// Mask to only keep the bits we care about
	mask := (uint64(1) << resolution) - 1
	return x & mask, y & mask, z & mask
}

// expandBits spreads the bits of a number so they can be interleaved
// Input:  00000000000000000000001011010010
// Output: 001000001000001000000001000000010
func expandBits(v uint64) uint64 {
	v = (v * 0x00010001) & 0xFF0000FF
	v = (v * 0x00000101) & 0x0F00F00F
	v = (v * 0x00000011) & 0xC30C30C3
	v = (v * 0x00000005) & 0x49249249
	return v
}

// compactBits is the inverse of expandBits
func compactBits(v uint64) uint64 {
	v &= 0x49249249
	v = (v | (v >> 2)) & 0xC30C30C3
	v = (v | (v >> 4)) & 0x0F00F00F
	v = (v | (v >> 8)) & 0xFF0000FF
	v = (v | (v >> 16)) & 0x0000FFFF
	return v
}
