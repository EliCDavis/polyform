package norm_test

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/norm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEveryUNormWidthClampsToBothEnds(t *testing.T) {
	widths := map[string]func(float64) norm.UNorm{
		"float": func(v float64) norm.UNorm { return norm.NewUFloat64(v) },
		"8":     func(v float64) norm.UNorm { return norm.NewU8(v) },
		"16":    func(v float64) norm.UNorm { return norm.NewU16(v) },
		"32":    func(v float64) norm.UNorm { return norm.NewU32(v) },
		"64":    func(v float64) norm.UNorm { return norm.NewU64(v) },
	}
	for name, build := range widths {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, 0., build(-3).Value())
			assert.Equal(t, 0., build(0).Value())
			assert.Equal(t, 1., build(1).Value())
			assert.Equal(t, 1., build(3).Value())
		})
	}
}

func TestEverySNormWidthClampsToBothEnds(t *testing.T) {
	widths := map[string]func(float64) norm.SNorm{
		"float": func(v float64) norm.SNorm { return norm.NewSFloat64(v) },
		"8":     func(v float64) norm.SNorm { return norm.NewS8(v) },
		"16":    func(v float64) norm.SNorm { return norm.NewS16(v) },
		"32":    func(v float64) norm.SNorm { return norm.NewS32(v) },
		"64":    func(v float64) norm.SNorm { return norm.NewS64(v) },
	}
	for name, build := range widths {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, -1., build(-3).Value())
			assert.Equal(t, -1., build(-1).Value())
			assert.Equal(t, 0., build(0).Value())
			assert.Equal(t, 1., build(1).Value())
			assert.Equal(t, 1., build(3).Value())
		})
	}
}

func TestValueClampsWhatAConversionLetsThrough(t *testing.T) {
	assert.Equal(t, 1., norm.UFloat64(5).Value())
	assert.Equal(t, 0., norm.UFloat64(-5).Value())
	assert.Equal(t, 1., norm.SFloat64(5).Value())
	assert.Equal(t, -1., norm.SFloat64(-5).Value())
}

func TestTheMostNegativeIntegersStayInRange(t *testing.T) {
	assert.Equal(t, -1., norm.S8(math.MinInt8).Value())
	assert.Equal(t, -1., norm.S16(math.MinInt16).Value())
	assert.Equal(t, -1., norm.S32(math.MinInt32).Value())
	assert.Equal(t, -1., norm.S64(math.MinInt64).Value())
}

func TestZeroValuesReadAsZero(t *testing.T) {
	var u norm.UFloat64
	var s norm.SFloat64
	assert.Equal(t, 0., u.Value())
	assert.Equal(t, 0., s.Value())
}

func TestNarrowWidthsQuantize(t *testing.T) {
	assert.Equal(t, 127./255., norm.NewU8(0.5).Value())
	assert.Equal(t, 63./127., norm.NewS8(0.5).Value())
}

func TestFloatBackedNormsKeepTheirValue(t *testing.T) {
	assert.Equal(t, 0.4, norm.NewUFloat64(0.4).Value())
	assert.Equal(t, -0.4, norm.NewSFloat64(-0.4).Value())
}

func TestEveryWidthMarshalsAsAPlainNumber(t *testing.T) {
	data, err := json.Marshal([]norm.UNorm{norm.NewUFloat64(0.25), norm.NewU8(1), norm.NewU16(1), norm.NewU32(1), norm.NewU64(0)})
	require.NoError(t, err)
	assert.JSONEq(t, `[0.25, 1, 1, 1, 0]`, string(data))

	data, err = json.Marshal([]norm.SNorm{norm.NewSFloat64(-0.75), norm.NewS8(-1), norm.NewS16(1), norm.NewS32(1), norm.NewS64(0)})
	require.NoError(t, err)
	assert.JSONEq(t, `[-0.75, -1, 1, 1, 0]`, string(data))
}

func TestQuantizedWidthsMarshalTheirNormalizedValue(t *testing.T) {
	for name, value := range map[string]interface{ Value() float64 }{
		"u8": norm.NewU8(0.5), "u16": norm.NewU16(0.5), "u32": norm.NewU32(0.5), "u64": norm.NewU64(0.5),
		"s8": norm.NewS8(-0.5), "s16": norm.NewS16(-0.5), "s32": norm.NewS32(-0.5), "s64": norm.NewS64(-0.5),
	} {
		t.Run(name, func(t *testing.T) {
			data, err := json.Marshal(value)
			require.NoError(t, err)

			var got float64
			require.NoError(t, json.Unmarshal(data, &got))
			assert.Equal(t, value.Value(), got)
			assert.InDelta(t, 0.5, math.Abs(got), 0.01)
		})
	}
}

func TestFloatNormsReadPlainNumbersAndClampOnValue(t *testing.T) {
	var values []norm.UFloat64
	require.NoError(t, json.Unmarshal([]byte(`[-2, 0.5, 2]`), &values))
	assert.Equal(t, 0., values[0].Value())
	assert.Equal(t, 0.5, values[1].Value())
	assert.Equal(t, 1., values[2].Value())
}
