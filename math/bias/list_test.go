package bias_test

import (
	"math"
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
		config  bias.ListConfig
		list    []bias.ListItem[string]
		results map[string]int
	}{
		"basic": {
			config:  bias.ListConfig{Seed: rand.New(rand.NewPCG(0, 0))},
			list:    []bias.ListItem[string]{{Item: "A", Weight: 0.7}, {Item: "B", Weight: 0.2}, {Item: "C", Weight: 0.1}},
			results: map[string]int{"A": 6998, "B": 1961, "C": 1041},
		},
		"basic-explicit temp": {
			config:  bias.ListConfig{Seed: rand.New(rand.NewPCG(0, 0)), Temperature: ptr(1)},
			list:    []bias.ListItem[string]{{Item: "A", Weight: 0.7}, {Item: "B", Weight: 0.2}, {Item: "C", Weight: 0.1}},
			results: map[string]int{"A": 6998, "B": 1961, "C": 1041},
		},
		"0 temp": {
			config:  bias.ListConfig{Seed: rand.New(rand.NewPCG(0, 0)), Temperature: ptr(0)},
			list:    []bias.ListItem[string]{{Item: "A", Weight: 0.7}, {Item: "B", Weight: 0.2}, {Item: "C", Weight: 0.1}},
			results: map[string]int{"A": 10000},
		},
		"5 temp": {
			config:  bias.ListConfig{Seed: rand.New(rand.NewPCG(0, 0)), Temperature: ptr(5)},
			list:    []bias.ListItem[string]{{Item: "A", Weight: 0.7}, {Item: "B", Weight: 0.2}, {Item: "C", Weight: 0.1}},
			results: map[string]int{"A": 4065, "B": 3114, "C": 2821},
		},
		"inf temp": {
			config:  bias.ListConfig{Seed: rand.New(rand.NewPCG(0, 0)), Temperature: ptr(math.Inf(1))},
			list:    []bias.ListItem[string]{{Item: "A", Weight: 0.7}, {Item: "B", Weight: 0.2}, {Item: "C", Weight: 0.1}},
			results: map[string]int{"A": 3332, "B": 3276, "C": 3392},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			list := bias.NewList(test.list, test.config)

			results := map[string]int{}

			// ACT ====================================================================
			for range 10000 {
				results[list.Next()] += 1
			}

			// ASSERT =================================================================
			assert.Len(t, results, len(test.results))
			for key, val := range test.results {
				assert.Equal(t, val, results[key], key)
			}
		})
	}
}

func TestNewListPanicsOnEmptyItems(t *testing.T) {
	assert.PanicsWithError(t, "No items provided", func() {
		bias.NewList[int](nil, bias.ListConfig{})
	})
}
