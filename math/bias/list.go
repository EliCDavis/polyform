package bias

import (
	"errors"
	"math/rand/v2"
	"time"
)

type ListItem[T any] struct {
	Item   T
	Weight float64
}

type List[T any] struct {
	items []T
	prob  []float64
	alias []int
	rand  *rand.Rand
}

// Next returns a biased random element
func (l *List[T]) Next() T {
	col := l.rand.IntN(len(l.prob))

	if l.rand.Float64() < l.prob[col] {
		return l.items[col]
	}

	return l.items[l.alias[col]]
}

func unzipAndAverage[T any](items []ListItem[T]) ([]float64, []T, float64) {
	weight := make([]float64, len(items))
	vals := make([]T, len(items))
	sum := float64(0)
	for i, item := range items {
		sum += item.Weight
		weight[i] = item.Weight
		vals[i] = item.Item
	}
	return weight, vals, sum / float64(len(items))
}

func NewList[T any](items []ListItem[T]) *List[T] {
	// TODO: Find a better seed
	return NewSeededList(items, rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano()))))
}

func NewSeededList[T any](items []ListItem[T], seed *rand.Rand) *List[T] {

	n := len(items)

	if n == 0 {
		panic(errors.New("No items provided"))
	}

	weights, vals, average := unzipAndAverage(items)
	wr := &List[T]{
		prob:  make([]float64, n),
		alias: make([]int, n),
		rand:  seed,
		items: vals,
	}

	// Fan out weights indexes to small or large
	small := make([]int, 0, n)
	large := make([]int, 0, n)
	for i, weight := range weights {
		if weight >= average {
			large = append(large, i)
		} else {
			small = append(small, i)
		}
	}

	// Fan out small and large into prob and alias
	for len(small) > 0 && len(large) > 0 {
		smallIdx := small[0]
		small = small[1:]
		largeIdx := large[0]
		large = large[1:]

		wr.prob[smallIdx] = weights[smallIdx] / average
		wr.alias[smallIdx] = largeIdx
		weights[largeIdx] -= average - weights[smallIdx]

		if weights[largeIdx] < average {
			small = append(small, largeIdx)
		} else {
			large = append(large, largeIdx)
		}
	}

	// Any indexes remaining in small or large assume normalized average
	for _, smallIdx := range small {
		wr.prob[smallIdx] = 1.0
	}
	for _, largeIdx := range large {
		wr.prob[largeIdx] = 1.0
	}

	return wr
}
