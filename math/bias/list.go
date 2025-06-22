package bias

import (
	"errors"
	"math"
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

type ListConfig struct {
	Seed        *rand.Rand
	Temperature *float64
}

func NewList[T any](dirtyItems []ListItem[T], config ListConfig) *List[T] {

	temp := 1.
	if config.Temperature != nil {
		temp = *config.Temperature
	}

	items := rewight(dirtyItems, temp)

	n := len(items)

	if n == 0 {
		panic(errors.New("No items provided"))
	}

	var seed *rand.Rand = config.Seed
	if seed == nil {
		seed = rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano())))
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

func rewight[T any](items []ListItem[T], temperature float64) []ListItem[T] {
	if len(items) == 0 {
		return nil
	}

	if temperature == 1 {
		return items
	}

	// 	/*
	// 		https://www.boristhebrave.com/2025/05/07/fiddling-weights-with-temperature/
	// 		def reweight(weights: list[float], temperature: float) -> list[float]:
	// 			if temperature == 0:
	// 				# At temperature 0, only the maximum weight is ever selected
	// 				max_weight = max(weights)
	// 				return [1.0 if w == max_weight else 0.0 for w in weights]

	// 			# Rescale weights (for numerical stability)
	// 			max_weight = max(weights)
	// 			weights = [w / max_weight for w in weights]

	// 			# Convert to logits and apply temperature
	// 			logits = [math.log(w) for w in weights]
	// 			scaled_logits = [l / temperature for l in logits]

	// 			# Handle overflow
	// 			if any(math.isinf(sl) for sl in scaled_logits):
	// 				return [1.0 if math.isinf(sl) else 0.0 for sl in scaled_logits]

	// 			# Convert back to weights
	// 			max_logit = max(scaled_logits)
	// 			exp_logits = [math.exp(l - max_logit) for l in scaled_logits]

	// 			# We don't need to divide by the sum here as that will be done in randomChoice,
	// 			# but I leave it in for clarity.
	// 			sum_exp = sum(exp_logits)
	// 			return [exp / sum_exp for exp in exp_logits]
	// 	*/

	weights := make([]float64, len(items))
	for i, item := range items {
		weights[i] = item.Weight
	}

	maxWeight := math.Inf(-1)
	for _, w := range weights {
		maxWeight = max(maxWeight, w)
	}

	// Handle temperature == 0: only max weight survives
	if temperature == 0 {
		out := make([]ListItem[T], 0)
		for _, item := range items {
			if item.Weight == maxWeight {
				out = append(out, ListItem[T]{Item: item.Item, Weight: 1.0})
			}
		}
		return out
	}

	// Normalize weights
	//     max_weight = max(weights)
	//     weights = [w / max_weight for w in weights]
	for i := range weights {
		weights[i] /= maxWeight
	}

	// Convert to logits and apply temperature
	//     logits = [math.log(w) for w in weights]
	//     scaled_logits = [l / temperature for l in logits]
	scaled_logits := make([]float64, len(weights))
	hasInf := false
	for i, l := range weights {
		scaled_logits[i] = math.Log(l) / temperature
		if math.IsInf(scaled_logits[i], 0) {
			hasInf = true
		}
	}

	// Handle overflow
	//     if any(math.isinf(sl) for sl in scaled_logits):
	//         return [1.0 if math.isinf(sl) else 0.0 for sl in scaled_logits]
	if hasInf {
		out := make([]ListItem[T], 0)
		for i := range items {
			if math.IsInf(scaled_logits[i], 0) {
				out = append(out, ListItem[T]{Item: items[i].Item, Weight: 1.0})
			}
		}
		return out
	}

	// Convert back to weights
	//     max_logit = max(scaled_logits)
	//     exp_logits = [math.exp(l - max_logit) for l in scaled_logits]
	maxLogit := scaled_logits[0]
	for _, l := range scaled_logits[1:] {
		maxLogit = max(maxLogit, l)
	}

	exp := make([]float64, len(scaled_logits))
	sum := 0.0
	for i, l := range scaled_logits {
		exp[i] = math.Exp(l - maxLogit)
		sum += exp[i]
	}

	// return [exp / sum_exp for exp in exp_logits]
	out := make([]ListItem[T], len(items))
	for i := range items {
		out[i] = ListItem[T]{Item: items[i].Item, Weight: exp[i] / sum}
	}
	return out
}
