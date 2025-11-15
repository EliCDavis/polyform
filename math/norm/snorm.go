package norm

import "math"

type SNorm interface {
	Value() float64
}

type s8 struct{ v int8 }

func (u s8) Value() float64 { return float64(u.v) / math.MaxInt8 }
func S8(v float64) SNorm    { return s8{v: int8(clampS(v) * math.MaxInt8)} }

// ============================================================================

type s16 struct{ v int16 }

func (u s16) Value() float64 { return float64(u.v) / math.MaxInt16 }
func S16(v float64) SNorm    { return s16{v: int16(clampS(v) * math.MaxInt16)} }

// ============================================================================

type s32 struct{ v int32 }

func (u s32) Value() float64 { return float64(u.v) / math.MaxInt32 }
func S32(v float64) SNorm    { return s32{v: int32(clampS(v) * math.MaxInt32)} }

// ============================================================================

type s64 struct{ v int64 }

func (u s64) Value() float64 { return float64(u.v) / math.MaxInt64 }
func S64(v float64) SNorm    { return s64{v: int64(clampS(v) * math.MaxInt64)} }

// ============================================================================

func clampS(val float64) float64 {
	return math.Max(math.Min(val, 1), -1)
}
