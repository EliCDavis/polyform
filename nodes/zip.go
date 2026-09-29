package nodes

import "fmt"

// []T is checked for first so that a T which is itself a slice still
// resolves: a port satisfies exactly one of the two interfaces.
func liftedValues[T any](port OutputPort) ([]T, int, bool) {
	if port == nil {
		return nil, rankScalar, false
	}
	if typed, ok := port.(Output[[]T]); ok {
		return typed.Value(), rankArray, true
	}
	if typed, ok := port.(Output[T]); ok {
		return []T{typed.Value()}, rankScalar, true
	}
	return nil, rankScalar, false
}

// liftedOperand is one input's values, already widened to a slice so the
// kernel loop does not care what rank it came in at.
type liftedOperand[T any] struct {
	values []T
	rank   int
}

func operandOf[T any](port OutputPort) liftedOperand[T] {
	values, rank, ok := liftedValues[T](port)
	if !ok {
		// Nothing connected reads as a single zero, which keeps an
		// unwired node scalar rather than erroring.
		var zero T
		return liftedOperand[T]{values: []T{zero}, rank: rankScalar}
	}
	return liftedOperand[T]{values: values, rank: rank}
}

// A length of one broadcasts across the whole result.
func (o liftedOperand[T]) at(i int) T {
	if len(o.values) == 1 {
		return o.values[0]
	}
	if i >= len(o.values) {
		var zero T
		return zero
	}
	return o.values[i]
}

// A length of one stretches to meet anything; past that the lengths have to
// agree, because a mismatch is a wiring mistake rather than a guess.
func broadcastLength(lengths ...int) (int, error) {
	n := 1
	for _, length := range lengths {
		if length == 1 {
			continue
		}
		if n == 1 {
			n = length
			continue
		}
		if length != n {
			return 0, fmt.Errorf("arrays of %s cannot be combined; they have to be the same length, or of length 1 to apply to every element", describeLengths(lengths))
		}
	}
	return n, nil
}

func describeLengths(lengths []int) string {
	out := ""
	for i, length := range lengths {
		if i > 0 {
			out += " and "
		}
		out += fmt.Sprintf("%d", length)
	}
	return out
}

func finish[R any](out *Lifted[R], rank int, results []R) {
	if rank == rankScalar {
		out.rank = rankScalar
		if len(results) > 0 {
			out.scalar = results[0]
		}
		return
	}
	out.rank = rankArray
	out.array = results
}

func Zip1[A, R any](out *Lifted[R], a LiftedPort[A], kernel func(A) R) {
	operand := operandOf[A](a)

	results := make([]R, len(operand.values))
	for i := range results {
		results[i] = kernel(operand.at(i))
	}
	finish(out, operand.rank, results)
}

func Zip2[A, B, R any](out *Lifted[R], a LiftedPort[A], b LiftedPort[B], kernel func(A, B) R) {
	left, right := operandOf[A](a), operandOf[B](b)

	rank := max(left.rank, right.rank)
	count, err := broadcastLength(len(left.values), len(right.values))
	if err != nil {
		out.CaptureError(err)
		finish(out, rank, nil)
		return
	}

	results := make([]R, count)
	for i := range results {
		results[i] = kernel(left.at(i), right.at(i))
	}
	finish(out, rank, results)
}

func Zip3[A, B, C, R any](out *Lifted[R], a LiftedPort[A], b LiftedPort[B], c LiftedPort[C], kernel func(A, B, C) R) {
	left, middle, right := operandOf[A](a), operandOf[B](b), operandOf[C](c)

	rank := max(left.rank, middle.rank, right.rank)
	count, err := broadcastLength(len(left.values), len(middle.values), len(right.values))
	if err != nil {
		out.CaptureError(err)
		finish(out, rank, nil)
		return
	}

	results := make([]R, count)
	for i := range results {
		results[i] = kernel(left.at(i), middle.at(i), right.at(i))
	}
	finish(out, rank, results)
}

func Zip4[A, B, C, D, R any](
	out *Lifted[R],
	a LiftedPort[A],
	b LiftedPort[B],
	c LiftedPort[C],
	d LiftedPort[D],
	kernel func(A, B, C, D) R,
) {
	first, second, third, fourth := operandOf[A](a), operandOf[B](b), operandOf[C](c), operandOf[D](d)

	rank := max(first.rank, second.rank, third.rank, fourth.rank)
	count, err := broadcastLength(
		len(first.values), len(second.values), len(third.values), len(fourth.values))
	if err != nil {
		out.CaptureError(err)
		finish(out, rank, nil)
		return
	}

	results := make([]R, count)
	for i := range results {
		results[i] = kernel(first.at(i), second.at(i), third.at(i), fourth.at(i))
	}
	finish(out, rank, results)
}

// An absent port is indistinguishable from one carrying the zero value, so a
// default has to be substituted here rather than inside the kernel.
func LiftedOr[T any](port LiftedPort[T], value T) LiftedPort[T] {
	if port == nil {
		return ConstOutput[T]{Val: value}
	}
	return port
}

// Hands the kernel one element from each connection, unwired holes left out.
// The slice is reused between elements, so a kernel keeping it has to copy.
func ZipAll[A, R any](out *Lifted[R], ports []LiftedPort[A], kernel func([]A) R) {
	operands := make([]liftedOperand[A], 0, len(ports))
	lengths := make([]int, 0, len(ports))
	rank := rankScalar

	for _, port := range ports {
		if port == nil {
			continue
		}
		operand := operandOf[A](port)
		operands = append(operands, operand)
		lengths = append(lengths, len(operand.values))
		rank = max(rank, operand.rank)
	}

	count, err := broadcastLength(lengths...)
	if err != nil {
		out.CaptureError(err)
		finish(out, rank, nil)
		return
	}

	results := make([]R, count)
	row := make([]A, len(operands))
	for i := range results {
		for j := range operands {
			row[j] = operands[j].at(i)
		}
		results[i] = kernel(row)
	}
	finish(out, rank, results)
}
