package opearations

import (
	"slices"

	"github.com/EliCDavis/polyform/nodes"
)

type Element nodes.DynamicType

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

type SeperateNode struct {
	Array     nodes.DynamicPort[[]Element]
	Selection nodes.Output[[]bool] `description:"One flag per entry. Entries past its end count as unselected."`
}

func (node SeperateNode) Description() string {
	return "Splits an array into selected and unselected entries."
}

func (node SeperateNode) split(recorder nodes.ExecutionRecorder) (nodes.DynamicArray, []any, []any, bool) {
	array, ok := nodes.DynamicArrayValue(node.Array)
	if !ok {
		return array, nil, nil, false
	}

	selection := nodes.TryGetOutputValue(recorder, node.Selection, nil)

	kept := make([]any, 0, array.Len())
	removed := make([]any, 0, array.Len())
	for i := 0; i < array.Len(); i++ {
		if i < len(selection) && selection[i] {
			kept = append(kept, array.At(i))
			continue
		}
		removed = append(removed, array.At(i))
	}

	// Seperate packs the removed entries from the end of one buffer, so they
	// come back reversed. Kept here so both paths agree.
	slices.Reverse(removed)

	return array, kept, removed, true
}

func (node SeperateNode) Selected(out *nodes.Dynamic[[]Element]) {
	array, kept, _, ok := node.split(out)
	if !ok {
		return
	}
	out.Set(array.Build(kept))
}

func (node SeperateNode) Removed(out *nodes.Dynamic[[]Element]) {
	array, _, removed, ok := node.split(out)
	if !ok {
		return
	}
	out.Set(array.Build(removed))
}
