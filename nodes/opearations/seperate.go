package opearations

import "github.com/EliCDavis/polyform/nodes"

func Seperate[T any](in []T, keep []bool) (kept, removed []T) {
	keptLen := 0
	removedLen := 0
	seperated := make([]T, len(in))

	for i, v := range in {
		if i < len(keep) && keep[i] {
			seperated[keptLen] = v
			keptLen++
		} else {
			seperated[len(in)-removedLen-1] = v
			removedLen++
		}
	}

	kept = seperated[:keptLen]
	removed = seperated[keptLen:]
	return
}

type SeperateNode[T any] struct {
	Array     nodes.Output[[]T]
	Selection nodes.Output[[]bool]
}

func (node SeperateNode[T]) Selected(out *nodes.StructOutput[[]T]) {
	kept, _ := Seperate(
		nodes.TryGetOutputValue(out, node.Array, nil),
		nodes.TryGetOutputValue(out, node.Selection, nil),
	)
	out.Set(kept)
}

func (node SeperateNode[T]) Removed(out *nodes.StructOutput[[]T]) {
	_, removed := Seperate(
		nodes.TryGetOutputValue(out, node.Array, nil),
		nodes.TryGetOutputValue(out, node.Selection, nil),
	)
	out.Set(removed)
}
