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

	normalized := point.Sub(min).
		DivByVector(max.Sub(min)).
		Clamp(0, 1).
		Scale(float64(maxVal))

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
	x = ExpandBits21(x)
	y = ExpandBits21(y)
	z = ExpandBits21(z)

	return x | (y << 1) | (z << 2)
}

// deinterleaveBits3D extracts the individual coordinates from a Morton code
func deinterleaveBits3D(morton uint64, resolution uint) (uint64, uint64, uint64) {
	x := CompactBits21(morton)
	y := CompactBits21(morton >> 1)
	z := CompactBits21(morton >> 2)

	// Mask to only keep the bits we care about
	mask := (uint64(1) << resolution) - 1
	return x & mask, y & mask, z & mask
}

// ExpandBits21 spreads the bits of a number so they can be interleaved
// Input:  00000000000000000000001011010010
// Output: 001000001000001000000001000000010
func ExpandBits21(v uint64) uint64 {
	v &= 0x1FFFFF // Ensure input is 21 bits max (0-20)
	v = (v | (v << 32)) & 0x1F00000000FFFF
	v = (v | (v << 16)) & 0x1F0000FF0000FF
	v = (v | (v << 8)) & 0x100F00F00F00F00F
	v = (v | (v << 4)) & 0x10C30C30C30C30C3
	v = (v | (v << 2)) & 0x1249249249249249
	return v
}

// CompactBits21 is the inverse of expandBits
func CompactBits21(v uint64) uint64 {
	v &= 0x1249249249249249
	v = (v | (v >> 2)) & 0x10C30C30C30C30C3
	v = (v | (v >> 4)) & 0x100F00F00F00F00F
	v = (v | (v >> 8)) & 0x1F0000FF0000FF
	v = (v | (v >> 16)) & 0x1F00000000FFFF
	v = (v | (v >> 32)) & 0x1FFFFF
	return v
}
