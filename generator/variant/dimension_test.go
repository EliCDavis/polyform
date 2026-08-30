package variant_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/generator/persistence"
	"github.com/EliCDavis/polyform/generator/variant"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func rawFloat(t *testing.T, v float64) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	return data
}

func TestDiscrete(t *testing.T) {
	d := variant.NewDiscrete("Fur Color", rawFloat(t, 1), rawFloat(t, 2), rawFloat(t, 3))

	assert.Equal(t, "Fur Color", d.Path())
	assert.Equal(t, 3, d.Count())

	v, err := d.Value(1)
	require.NoError(t, err)
	assert.JSONEq(t, `2`, string(v))

	_, err = d.Value(3)
	assert.Error(t, err, "index equal to Count() should be out of range")

	rng := rand.New(rand.NewSource(1))
	random, err := d.Random(rng)
	require.NoError(t, err)
	assert.Contains(t, []string{"1", "2", "3"}, string(random))
}

func TestNumericRangeSweepValues(t *testing.T) {
	r := variant.NewNumericRange("Scale", 0, 10, 5)
	assert.Equal(t, 5, r.Count())

	want := []float64{0, 2.5, 5, 7.5, 10}
	for i, w := range want {
		v, err := r.Value(i)
		require.NoError(t, err)
		var got float64
		require.NoError(t, json.Unmarshal(v, &got))
		assert.InDelta(t, w, got, 1e-9, "sample %d", i)
	}
}

func TestNumericRangeSingleSampleReturnsMin(t *testing.T) {
	r := variant.NewNumericRange("Scale", 5, 10, 1)
	assert.Equal(t, 1, r.Count())

	v, err := r.Value(0)
	require.NoError(t, err)
	var got float64
	require.NoError(t, json.Unmarshal(v, &got))
	assert.InDelta(t, 5, got, 1e-9, "a single sample should not divide by zero and should return Min")
}

func TestNumericRangeRandomStaysInBounds(t *testing.T) {
	r := variant.NewNumericRange("Scale", 2, 4, 10)
	rng := rand.New(rand.NewSource(42))

	for range 50 {
		v, err := r.Random(rng)
		require.NoError(t, err)
		var got float64
		require.NoError(t, json.Unmarshal(v, &got))
		assert.GreaterOrEqual(t, got, 2.0)
		assert.Less(t, got, 4.0)
	}
}

func TestVector2RangeLerpsMinAndMaxTogether(t *testing.T) {
	r := variant.NewVector2Range("Position", 0, 10, 0, 100, 3)
	assert.Equal(t, 3, r.Count(), "Samples describes the whole sweep, not a per-axis count")

	want := []vector2.Float64{
		vector2.New(0.0, 0.0),
		vector2.New(5.0, 50.0),
		vector2.New(10.0, 100.0),
	}
	for i, w := range want {
		v, err := r.Value(i)
		require.NoError(t, err)
		var got vector2.Float64
		require.NoError(t, json.Unmarshal(v, &got))
		assert.InDelta(t, w.X(), got.X(), 1e-9, "sample %d", i)
		assert.InDelta(t, w.Y(), got.Y(), 1e-9, "sample %d", i)
	}
}

func TestVector3RangeLerpsMinAndMaxTogether(t *testing.T) {
	r := variant.NewVector3Range("Position", 0, 10, 0, 100, 0, 1000, 3)
	assert.Equal(t, 3, r.Count(), "Samples describes the whole sweep, not a per-axis count")

	want := []vector3.Float64{
		vector3.New(0.0, 0.0, 0.0),
		vector3.New(5.0, 50.0, 500.0),
		vector3.New(10.0, 100.0, 1000.0),
	}
	for i, w := range want {
		v, err := r.Value(i)
		require.NoError(t, err)
		var got vector3.Float64
		require.NoError(t, json.Unmarshal(v, &got))
		assert.InDelta(t, w.X(), got.X(), 1e-9, "sample %d", i)
		assert.InDelta(t, w.Y(), got.Y(), 1e-9, "sample %d", i)
		assert.InDelta(t, w.Z(), got.Z(), 1e-9, "sample %d", i)
	}
}

func TestRGBRangeLerpsMinAndMaxTogether(t *testing.T) {
	r := variant.NewRGBRange("Fur Color", persistence.WebColor{}, persistence.WebColor{R: 255, G: 255, B: 255}, 3)
	assert.Equal(t, 3, r.Count(), "Samples describes the whole sweep, not a per-channel count")

	seen := map[string]bool{}
	for i := range r.Count() {
		v, err := r.Value(i)
		require.NoError(t, err)

		var c coloring.Color
		require.NoError(t, json.Unmarshal(v, &c), "must decode as a drawing/coloring.Color")
		assert.InDelta(t, c.R, c.G, 1.0/255, "channels move together")
		assert.InDelta(t, c.G, c.B, 1.0/255, "channels move together")

		seen[string(v)] = true
	}
	assert.Len(t, seen, 3, "all 3 samples should be distinct")
}

func TestRGBRangeRandomStaysInBounds(t *testing.T) {
	r := variant.NewRGBRange("Fur Color", persistence.WebColor{}, persistence.WebColor{R: 0x7f, G: 0x7f, B: 0x7f}, 10)
	rng := rand.New(rand.NewSource(3))

	for range 50 {
		v, err := r.Random(rng)
		require.NoError(t, err)

		var c coloring.Color
		require.NoError(t, json.Unmarshal(v, &c))
		assert.GreaterOrEqual(t, c.R, 0.0)
		assert.LessOrEqual(t, c.R, 0.51, "byte-quantization can push slightly above the float bound")
		assert.GreaterOrEqual(t, c.G, 0.0)
		assert.LessOrEqual(t, c.G, 0.51)
		assert.GreaterOrEqual(t, c.B, 0.0)
		assert.LessOrEqual(t, c.B, 0.51)
	}
}

func TestHSVRangeKnownColors(t *testing.T) {
	tests := []struct {
		name    string
		h, s, v float64
		r, g, b float64
	}{
		{"red", 0, 1, 1, 1, 0, 0},
		{"green", 120, 1, 1, 0, 1, 0},
		{"blue", 240, 1, 1, 0, 0, 1},
		{"white", 0, 0, 1, 1, 1, 1},
		{"black", 0, 0, 0, 0, 0, 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := variant.NewHSVRange("Fur Color", test.h, test.h, test.s, test.s, test.v, test.v, 1)
			value, err := r.Value(0)
			require.NoError(t, err)

			var c coloring.Color
			require.NoError(t, json.Unmarshal(value, &c))
			assert.InDelta(t, test.r, c.R, 1.0/255)
			assert.InDelta(t, test.g, c.G, 1.0/255)
			assert.InDelta(t, test.b, c.B, 1.0/255)
		})
	}
}

func TestCombinationSumsCounts(t *testing.T) {
	c := variant.NewCombination("Scale",
		variant.NewDiscrete("Scale", rawFloat(t, -1), rawFloat(t, -2)),
		variant.NewNumericRange("Scale", 0, 10, 5),
	)

	assert.Equal(t, "Scale", c.Path())
	assert.Equal(t, 7, c.Count(), "2 discrete values + 5 numeric samples")
}

func TestCombinationValueDelegatesToCoveringDimension(t *testing.T) {
	c := variant.NewCombination("Scale",
		variant.NewDiscrete("Scale", rawFloat(t, -1), rawFloat(t, -2)),
		variant.NewNumericRange("Scale", 0, 10, 5),
	)

	v, err := c.Value(0)
	require.NoError(t, err)
	assert.JSONEq(t, `-1`, string(v))

	v, err = c.Value(1)
	require.NoError(t, err)
	assert.JSONEq(t, `-2`, string(v))

	v, err = c.Value(2)
	require.NoError(t, err)
	var got float64
	require.NoError(t, json.Unmarshal(v, &got))
	assert.InDelta(t, 0, got, 1e-9, "index 2 is the first numeric range sample")

	_, err = c.Value(7)
	assert.Error(t, err, "index equal to Count() should be out of range")
}

func TestCombinationRandomStaysWithinAllDimensions(t *testing.T) {
	c := variant.NewCombination("Scale",
		variant.NewDiscrete("Scale", rawFloat(t, -1), rawFloat(t, -2)),
		variant.NewNumericRange("Scale", 0, 10, 5),
	)

	rng := rand.New(rand.NewSource(7))
	for range 50 {
		v, err := c.Random(rng)
		require.NoError(t, err)
		var got float64
		require.NoError(t, json.Unmarshal(v, &got))
		assert.GreaterOrEqual(t, got, -2.0)
		assert.LessOrEqual(t, got, 10.0)
	}
}

func TestCombinationRandomRejectsEmptyDimensions(t *testing.T) {
	c := variant.NewCombination("Scale")
	_, err := c.Random(rand.New(rand.NewSource(1)))
	assert.Error(t, err)
}

func TestDimensionRoundTripsThroughJSON(t *testing.T) {
	tests := map[string]variant.Dimension{
		"discrete":      variant.NewDiscrete("Color", rawFloat(t, 1), rawFloat(t, 2)),
		"numeric range": variant.NewNumericRange("Scale", 0, 10, 5),
		"vector2 range": variant.NewVector2Range("Position2D", 0, 1, 0, 1, 2),
		"vector3 range": variant.NewVector3Range("Position3D", 0, 1, 0, 1, 0, 1, 2),
		"rgb range":     variant.NewRGBRange("Fur Color", persistence.WebColor{}, persistence.WebColor{R: 255, G: 255, B: 255}, 2),
		"hsv range":     variant.NewHSVRange("Fur Color", 0, 360, 0, 1, 0, 1, 2),
		"combination": variant.NewCombination("Scale",
			variant.NewDiscrete("Scale", rawFloat(t, -1), rawFloat(t, -2)),
			variant.NewNumericRange("Scale", 0, 10, 5),
		),
	}

	for name, dim := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := json.Marshal(dim)
			require.NoError(t, err)

			decoded, err := variant.UnmarshalDimension(dim.Path(), data)
			require.NoError(t, err)

			assert.Equal(t, dim.Path(), decoded.Path())
			assert.Equal(t, dim.Count(), decoded.Count())

			for i := range dim.Count() {
				want, err := dim.Value(i)
				require.NoError(t, err)
				got, err := decoded.Value(i)
				require.NoError(t, err)
				assert.JSONEq(t, string(want), string(got))
			}
		})
	}
}
