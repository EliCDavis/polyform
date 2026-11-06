package norm

import (
	"encoding/json"
	"math"
)

type SNorm interface {
	Value() float64
}

type SFloat64 float64

func (u SFloat64) Value() float64    { return clampS(float64(u)) }
func NewSFloat64(v float64) SFloat64 { return SFloat64(clampS(v)) }

// ============================================================================

type S8 int8

func (u S8) Value() float64               { return clampS(float64(u) / math.MaxInt8) }
func (u S8) MarshalJSON() ([]byte, error) { return json.Marshal(u.Value()) }
func NewS8(v float64) S8                  { return S8(clampS(v) * math.MaxInt8) }

// ============================================================================

type S16 int16

func (u S16) Value() float64               { return clampS(float64(u) / math.MaxInt16) }
func (u S16) MarshalJSON() ([]byte, error) { return json.Marshal(u.Value()) }
func NewS16(v float64) S16                 { return S16(clampS(v) * math.MaxInt16) }

// ============================================================================

type S32 int32

func (u S32) Value() float64               { return clampS(float64(u) / math.MaxInt32) }
func (u S32) MarshalJSON() ([]byte, error) { return json.Marshal(u.Value()) }
func NewS32(v float64) S32                 { return S32(clampS(v) * math.MaxInt32) }

// ============================================================================

type S64 int64

func (u S64) Value() float64               { return clampS(float64(u) / math.MaxInt64) }
func (u S64) MarshalJSON() ([]byte, error) { return json.Marshal(u.Value()) }

func NewS64(v float64) S64 {
	// MaxInt64 rounds up as a float, so scaling 1 overflows the conversion.
	if v >= 1 {
		return S64(math.MaxInt64)
	}
	return S64(clampS(v) * math.MaxInt64)
}

// ============================================================================

func clampS(val float64) float64 {
	return math.Max(math.Min(val, 1), -1)
}
