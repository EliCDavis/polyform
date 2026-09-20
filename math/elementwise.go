package math

import (
	"fmt"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
)

func elementwise[T any, G any](out *nodes.StructOutput[[]G], a, b []T, f func(a, b T) G) {
	n := len(a)
	if len(b) != n {
		out.CaptureError(fmt.Errorf("A has %d elements but B has %d; only the first %d were combined", len(a), len(b), min(len(a), len(b))))
		n = min(len(a), len(b))
	}
	result := make([]G, n)
	for i := range result {
		result[i] = f(a[i], b[i])
	}
	out.Set(result)
}

// ============================================================================

type AddArraysNode[T vector.Number] struct {
	A nodes.Output[[]T]
	B nodes.Output[[]T]
}

func (n AddArraysNode[T]) Description() string {
	return "Adds two arrays element by element: Sums[i] = A[i] + B[i]."
}

func (n AddArraysNode[T]) Sums(out *nodes.StructOutput[[]T]) {
	elementwise(out,
		nodes.TryGetOutputValue(out, n.A, nil),
		nodes.TryGetOutputValue(out, n.B, nil),
		func(a, b T) T { return a + b },
	)
}

// ============================================================================

type SubtractArraysNode[T vector.Number] struct {
	A nodes.Output[[]T] `description:"The values being subtracted from."`
	B nodes.Output[[]T] `description:"The values being subtracted."`
}

func (n SubtractArraysNode[T]) Description() string {
	return "Subtracts two arrays element by element: Differences[i] = A[i] - B[i]."
}

func (n SubtractArraysNode[T]) Differences(out *nodes.StructOutput[[]T]) {
	elementwise(out,
		nodes.TryGetOutputValue(out, n.A, nil),
		nodes.TryGetOutputValue(out, n.B, nil),
		func(a, b T) T { return a - b },
	)
}

// ============================================================================

type MultiplyArraysNode[T vector.Number] struct {
	A nodes.Output[[]T]
	B nodes.Output[[]T]
}

func (n MultiplyArraysNode[T]) Description() string {
	return "Multiplies two arrays element by element: Products[i] = A[i] * B[i]."
}

func (n MultiplyArraysNode[T]) Products(out *nodes.StructOutput[[]T]) {
	elementwise(out,
		nodes.TryGetOutputValue(out, n.A, nil),
		nodes.TryGetOutputValue(out, n.B, nil),
		func(a, b T) T { return a * b },
	)
}

// ============================================================================

type DivideArraysNode[T vector.Number] struct {
	A nodes.Output[[]T] `description:"The dividends."`
	B nodes.Output[[]T] `description:"The divisors. A zero here yields 0 at that index."`
}

func (n DivideArraysNode[T]) Description() string {
	return "Divides two arrays element by element: Quotients[i] = A[i] / B[i]."
}

func (n DivideArraysNode[T]) Quotients(out *nodes.StructOutput[[]T]) {
	dividedByZero := false
	elementwise(out,
		nodes.TryGetOutputValue(out, n.A, nil),
		nodes.TryGetOutputValue(out, n.B, nil),
		func(a, b T) T {
			if b == 0 {
				dividedByZero = true
				return 0
			}
			return a / b
		},
	)
	if dividedByZero {
		out.CaptureError(cantDivideByZeroErr)
	}
}

// ============================================================================

type MinArraysNode[T vector.Number] struct {
	A nodes.Output[[]T]
	B nodes.Output[[]T]
}

func (n MinArraysNode[T]) Description() string {
	return "The smaller of two arrays at each index: Minimums[i] = min(A[i], B[i])."
}

func (n MinArraysNode[T]) Minimums(out *nodes.StructOutput[[]T]) {
	elementwise(out,
		nodes.TryGetOutputValue(out, n.A, nil),
		nodes.TryGetOutputValue(out, n.B, nil),
		func(a, b T) T { return min(a, b) },
	)
}

// ============================================================================

type MaxArraysNode[T vector.Number] struct {
	A nodes.Output[[]T]
	B nodes.Output[[]T]
}

func (n MaxArraysNode[T]) Description() string {
	return "The larger of two arrays at each index: Maximums[i] = max(A[i], B[i])."
}

func (n MaxArraysNode[T]) Maximums(out *nodes.StructOutput[[]T]) {
	elementwise(out,
		nodes.TryGetOutputValue(out, n.A, nil),
		nodes.TryGetOutputValue(out, n.B, nil),
		func(a, b T) T { return max(a, b) },
	)
}
