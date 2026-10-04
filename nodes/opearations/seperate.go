package opearations

import (
	"slices"

	"github.com/EliCDavis/polyform/nodes"
)

type Element nodes.DynamicType

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

	// Last first: the order this node has always returned them in.
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
