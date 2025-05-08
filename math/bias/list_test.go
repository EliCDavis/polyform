package bias_test

import (
	"math/rand/v2"
	"testing"

	"github.com/EliCDavis/polyform/math/bias"
	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	list := bias.NewSeededList([]bias.ListItem[string]{
		{Item: "A", Weight: 0.7},
		{Item: "B", Weight: 0.2},
		{Item: "C", Weight: 0.1},
	}, rand.New(rand.NewPCG(0, 0)))

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
	assert.Equal(t, 6998, results["A"])
	assert.Equal(t, 1961, results["B"])
	assert.Equal(t, 1041, results["C"])
}

func TestNewListPanicsOnEmptyItems(t *testing.T) {
	assert.PanicsWithError(t, "No items provided", func() {
		bias.NewList[int](nil)
	})
}
