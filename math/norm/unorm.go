package norm

import "math"

type UNorm interface {
	Value() float64
}

type u8 struct{ v uint8 }

func (u u8) Value() float64 { return float64(u.v) / math.MaxUint8 }
func U8(v float64) UNorm    { return u8{v: uint8(clampU(v) * math.MaxUint8)} }

// ============================================================================

type u16 struct{ v uint16 }

func (u u16) Value() float64 { return float64(u.v) / math.MaxUint16 }
func U16(v float64) UNorm    { return u16{v: uint16(clampU(v) * math.MaxUint16)} }

// ============================================================================

type u32 struct{ v uint32 }

func (u u32) Value() float64 { return float64(u.v) / math.MaxUint32 }
func U32(v float64) UNorm    { return u32{v: uint32(clampU(v) * math.MaxUint32)} }

// ============================================================================

type u64 struct{ v uint64 }

func (u u64) Value() float64 { return float64(u.v) / math.MaxUint64 }
func U64(v float64) UNorm    { return u64{v: uint64(clampU(v) * math.MaxUint64)} }

// ============================================================================

func clampU(val float64) float64 {
	return math.Max(math.Min(val, 1), 0)
}
