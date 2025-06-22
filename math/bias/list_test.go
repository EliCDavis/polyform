package bias_test

import (
	"math/rand/v2"
	"testing"

	"github.com/EliCDavis/polyform/math/bias"
	"github.com/stretchr/testify/assert"
)

func ptr(f float64) *float64 {
	return &f
}

func TestList(t *testing.T) {

	var tests = map[string]struct {
		config bias.ListConfig
		a      int
		b      int
		c      int
	}{
		"basic": {
			config: bias.ListConfig{Seed: rand.New(rand.NewPCG(0, 0))},
			a:      6998,
			b:      1961,
			c:      1041,
		},
		"basic-explicit temp": {
			config: bias.ListConfig{Seed: rand.New(rand.NewPCG(0, 0)), Temperature: ptr(1)},
			a:      6998,
			b:      1961,
			c:      1041,
		},
		"0 temp": {
			config: bias.ListConfig{Seed: rand.New(rand.NewPCG(0, 0)), Temperature: ptr(0)},
			a:      10000,
			b:      0,
			c:      0,
		},
		"5 temp": {
			config: bias.ListConfig{Seed: rand.New(rand.NewPCG(0, 0)), Temperature: ptr(5)},
			a:      4065,
			b:      3114,
			c:      2821,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			list := bias.NewList([]bias.ListItem[string]{
				{Item: "A", Weight: 0.7},
				{Item: "B", Weight: 0.2},
				{Item: "C", Weight: 0.1},
			}, test.config)

			results := map[string]int{
				"A": 0,
				"B": 0,
				"C": 0,
			}

			// ACT ====================================================================
			for range 10000 {
				results[list.Next()] += 1
			}

			// ASSERT =================================================================
			assert.Len(t, results, 3)
			assert.Equal(t, test.a, results["A"])
			assert.Equal(t, test.b, results["B"])
			assert.Equal(t, test.c, results["C"])
		})
	}

}

func TestNewListPanicsOnEmptyItems(t *testing.T) {
	assert.PanicsWithError(t, "No items provided", func() {
		bias.NewList[int](nil, bias.ListConfig{})
	})
}
