package utils_test

import (
	"testing"

	"github.com/EliCDavis/polyform/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSort(t *testing.T) {
	// ARRANGE ================================================================
	input := map[string]int{
		"a": 3,
		"b": 2,
		"c": 1,
	}

	// ACT ====================================================================
	sorted := utils.SortMapByKey(input)

	// ASSERT =================================================================
	require.Len(t, sorted, 3)

	assert.Equal(t, "a", sorted[0].Key)
	assert.Equal(t, "b", sorted[1].Key)
	assert.Equal(t, "c", sorted[2].Key)

	assert.Equal(t, 3, sorted[0].Val)
	assert.Equal(t, 2, sorted[1].Val)
	assert.Equal(t, 1, sorted[2].Val)
}
