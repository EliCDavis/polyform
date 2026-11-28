package norm

import (
	"encoding/json"
	"math"

	"github.com/EliCDavis/polyform/math/scalar"
)

type UNorm interface {
	Value() float64
}

type UFloat64 float64

func (u UFloat64) Value() float64    { return scalar.Clamp(float64(u), 0, 1) }
func NewUFloat64(v float64) UFloat64 { return UFloat64(scalar.Clamp(v, 0, 1)) }

// ============================================================================

type U8 uint8

func (u U8) Value() float64               { return float64(u) / math.MaxUint8 }
func (u U8) MarshalJSON() ([]byte, error) { return json.Marshal(u.Value()) }
func NewU8(v float64) U8                  { return U8(scalar.Clamp(v, 0, 1) * math.MaxUint8) }

// ============================================================================

type U16 uint16

func (u U16) Value() float64               { return float64(u) / math.MaxUint16 }
func (u U16) MarshalJSON() ([]byte, error) { return json.Marshal(u.Value()) }
func NewU16(v float64) U16                 { return U16(scalar.Clamp(v, 0, 1) * math.MaxUint16) }

// ============================================================================

type U32 uint32

func (u U32) Value() float64               { return float64(u) / math.MaxUint32 }
func (u U32) MarshalJSON() ([]byte, error) { return json.Marshal(u.Value()) }
func NewU32(v float64) U32                 { return U32(scalar.Clamp(v, 0, 1) * math.MaxUint32) }

// ============================================================================

type U64 uint64

func (u U64) Value() float64               { return float64(u) / math.MaxUint64 }
func (u U64) MarshalJSON() ([]byte, error) { return json.Marshal(u.Value()) }

func NewU64(v float64) U64 {
	if v >= 1 {
		return U64(math.MaxUint64)
	}
	return U64(scalar.Clamp(v, 0, 1) * math.MaxUint64)
}
