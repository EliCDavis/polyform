package norm

import (
	"encoding/json"
	"math"
)

type UNorm interface {
	Value() float64
}

type UFloat64 float64

func (u UFloat64) Value() float64    { return clampU(float64(u)) }
func NewUFloat64(v float64) UFloat64 { return UFloat64(clampU(v)) }

// ============================================================================

type U8 uint8

func (u U8) Value() float64               { return float64(u) / math.MaxUint8 }
func (u U8) MarshalJSON() ([]byte, error) { return json.Marshal(u.Value()) }
func NewU8(v float64) U8                  { return U8(clampU(v) * math.MaxUint8) }

// ============================================================================

type U16 uint16

func (u U16) Value() float64               { return float64(u) / math.MaxUint16 }
func (u U16) MarshalJSON() ([]byte, error) { return json.Marshal(u.Value()) }
func NewU16(v float64) U16                 { return U16(clampU(v) * math.MaxUint16) }

// ============================================================================

type U32 uint32

func (u U32) Value() float64               { return float64(u) / math.MaxUint32 }
func (u U32) MarshalJSON() ([]byte, error) { return json.Marshal(u.Value()) }
func NewU32(v float64) U32                 { return U32(clampU(v) * math.MaxUint32) }

// ============================================================================

type U64 uint64

func (u U64) Value() float64               { return float64(u) / math.MaxUint64 }
func (u U64) MarshalJSON() ([]byte, error) { return json.Marshal(u.Value()) }

func NewU64(v float64) U64 {
	if v >= 1 {
		return U64(math.MaxUint64)
	}
	return U64(clampU(v) * math.MaxUint64)
}

// ============================================================================

func clampU(val float64) float64 {
	return math.Max(math.Min(val, 1), 0)
}
