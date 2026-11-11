package utils_test

import (
	"sync"
	"testing"

	"github.com/EliCDavis/polyform/utils"
	"github.com/stretchr/testify/assert"
)

func TestParallelRangeVisitsEveryIndexOnce(t *testing.T) {
	for _, n := range []int{0, 1, 7, 1000, 100003} {
		visits := make([]int, n)
		var mu sync.Mutex
		calls := 0
		utils.ParallelRange(n, func(start, end int) {
			mu.Lock()
			calls++
			mu.Unlock()
			for i := start; i < end; i++ {
				visits[i]++
			}
		})
		for i, v := range visits {
			assert.Equal(t, 1, v, "n=%d index %d", n, i)
		}
		assert.GreaterOrEqual(t, calls, 1, "n=%d", n)
	}
}
