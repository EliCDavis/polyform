package variant_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/EliCDavis/polyform/generator/variant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetTotalCombinationsMultipliesAcrossDimensions(t *testing.T) {
	set := variant.Set{
		Dimensions: []variant.Dimension{
			variant.NewNumericRange("A", 0, 1, 2),
			variant.NewNumericRange("B", 0, 1, 3),
			variant.NewDiscrete("C", rawFloat(t, 1), rawFloat(t, 2), rawFloat(t, 3), rawFloat(t, 4)),
		},
	}
	assert.Equal(t, 2*3*4, set.TotalCombinations())
}

func TestSetSweepProducesEveryDistinctCombination(t *testing.T) {
	set := variant.Set{
		Dimensions: []variant.Dimension{
			variant.NewNumericRange("A", 0, 1, 2),
			variant.NewDiscrete("B", rawFloat(t, 10), rawFloat(t, 20), rawFloat(t, 30)),
		},
	}

	profiles, err := set.Sweep()
	require.NoError(t, err)
	assert.Len(t, profiles, 6)

	seen := map[string]bool{}
	for _, p := range profiles {
		require.Contains(t, p, "A")
		require.Contains(t, p, "B")
		key := string(p["A"]) + "|" + string(p["B"])
		seen[key] = true
	}
	assert.Len(t, seen, 6, "all 6 combinations should be distinct")

	// every A paired with every B at least once
	wantA := map[string]bool{"0": false, "1": false}
	wantB := map[string]bool{"10": false, "20": false, "30": false}
	for _, p := range profiles {
		wantA[string(p["A"])] = true
		wantB[string(p["B"])] = true
	}
	for v, ok := range wantA {
		assert.True(t, ok, "A=%s should appear in the sweep", v)
	}
	for v, ok := range wantB {
		assert.True(t, ok, "B=%s should appear in the sweep", v)
	}
}

func TestSetSweepMixedRadixOrderAcrossThreeDimensions(t *testing.T) {
	rawStr := func(v string) json.RawMessage {
		data, err := json.Marshal(v)
		require.NoError(t, err)
		return data
	}

	set := variant.Set{
		Dimensions: []variant.Dimension{
			variant.NewDiscrete("A", rawStr("a0"), rawStr("a1")),
			variant.NewDiscrete("B", rawStr("b0"), rawStr("b1"), rawStr("b2")),
			variant.NewDiscrete("C", rawStr("c0"), rawStr("c1")),
		},
	}

	profiles, err := set.Sweep()
	require.NoError(t, err)

	// A (count 2) is least significant, then B (count 3), then C (count 2)
	// most significant - this is the exact digit order Sweep must produce.
	want := [][3]string{
		{"a0", "b0", "c0"}, {"a1", "b0", "c0"},
		{"a0", "b1", "c0"}, {"a1", "b1", "c0"},
		{"a0", "b2", "c0"}, {"a1", "b2", "c0"},
		{"a0", "b0", "c1"}, {"a1", "b0", "c1"},
		{"a0", "b1", "c1"}, {"a1", "b1", "c1"},
		{"a0", "b2", "c1"}, {"a1", "b2", "c1"},
	}
	require.Len(t, profiles, len(want))

	for i, w := range want {
		assert.JSONEq(t, `"`+w[0]+`"`, string(profiles[i]["A"]), "profile %d, dimension A", i)
		assert.JSONEq(t, `"`+w[1]+`"`, string(profiles[i]["B"]), "profile %d, dimension B", i)
		assert.JSONEq(t, `"`+w[2]+`"`, string(profiles[i]["C"]), "profile %d, dimension C", i)
	}
}

func TestSetSweepSingleDimension(t *testing.T) {
	set := variant.Set{
		Dimensions: []variant.Dimension{
			variant.NewNumericRange("A", 0, 10, 5),
		},
	}

	profiles, err := set.Sweep()
	require.NoError(t, err)
	require.Len(t, profiles, 5)

	var got float64
	require.NoError(t, json.Unmarshal(profiles[2]["A"], &got))
	assert.InDelta(t, 5, got, 1e-9)
}

func TestSetSampleProducesExactlyNProfilesWithinRange(t *testing.T) {
	set := variant.Set{
		Dimensions: []variant.Dimension{
			variant.NewNumericRange("A", 0, 1, 100),
			variant.NewDiscrete("B", rawFloat(t, 1), rawFloat(t, 2)),
		},
	}

	rng := rand.New(rand.NewSource(7))
	profiles, err := set.Sample(25, rng)
	require.NoError(t, err)
	assert.Len(t, profiles, 25)

	for _, p := range profiles {
		var a float64
		require.NoError(t, json.Unmarshal(p["A"], &a))
		assert.GreaterOrEqual(t, a, 0.0)
		assert.Less(t, a, 1.0)
		assert.Contains(t, []string{"1", "2"}, string(p["B"]))
	}
}

func TestSetSampleIsDeterministicGivenTheSameSeed(t *testing.T) {
	set := variant.Set{
		Dimensions: []variant.Dimension{
			variant.NewNumericRange("A", 0, 100, 1000),
		},
	}

	a, err := set.Sample(5, rand.New(rand.NewSource(99)))
	require.NoError(t, err)
	b, err := set.Sample(5, rand.New(rand.NewSource(99)))
	require.NoError(t, err)

	for i := range a {
		assert.JSONEq(t, string(a[i]["A"]), string(b[i]["A"]))
	}
}

func TestSetSampleRejectsNilRand(t *testing.T) {
	set := variant.Set{Dimensions: []variant.Dimension{variant.NewNumericRange("A", 0, 1, 2)}}
	_, err := set.Sample(5, nil)
	assert.Error(t, err)
}

func TestSetSampleRejectsNegativeCount(t *testing.T) {
	set := variant.Set{Dimensions: []variant.Dimension{variant.NewNumericRange("A", 0, 1, 2)}}
	_, err := set.Sample(-1, rand.New(rand.NewSource(1)))
	assert.Error(t, err)
}
