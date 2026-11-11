package arrays

import (
	"fmt"

	"github.com/EliCDavis/polyform/nodes"
)

type GatherNode struct {
	Array   nodes.DynamicPort[[]Element] `description:"The array to read from."`
	Indices nodes.Output[[]int]          `description:"Which elements to read, in order, counting from 0. Negative counts back from the end; an index may repeat."`
}

func (GatherNode) Description() string {
	return "Reads several elements out of an array by index, in the order given."
}

func (GatherNode) Keywords() []string {
	return []string{"pick"}
}

func (g GatherNode) Out(out *nodes.Dynamic[[]Element]) {
	values, ok := nodes.DynamicArrayValue(out, g.Array)
	if !ok {
		return
	}

	indices := nodes.TryGetOutputValue(out, g.Indices, nil)
	picked := make([]any, 0, len(indices))
	for _, index := range indices {
		at := index
		if at < 0 {
			at += values.Len()
		}
		if at < 0 || at >= values.Len() {
			out.CaptureError(fmt.Errorf("index %d is outside an array of %d element(s)", index, values.Len()))
			return
		}
		picked = append(picked, values.At(at))
	}
	out.Set(values.Build(picked))
}

// ============================================================================

type FilterNode struct {
	Array nodes.DynamicPort[[]Element] `description:"The array to filter."`
	Keep  nodes.Output[[]bool]         `description:"One flag per element, true to keep it."`
}

func (FilterNode) Description() string {
	return "Splits an array by a list of flags, keeping order: Kept holds the flagged elements, Dropped the rest."
}

func (FilterNode) Keywords() []string {
	return []string{"mask", "where"}
}

func (f FilterNode) split(recorder nodes.ExecutionRecorder, keep bool) (nodes.DynamicArray, []any, bool) {
	values, ok := nodes.DynamicArrayValue(recorder, f.Array)
	if !ok {
		return values, nil, false
	}
	flags := nodes.TryGetOutputValue(recorder, f.Keep, nil)
	if len(flags) != values.Len() {
		recorder.CaptureError(fmt.Errorf("%d flag(s) for %d element(s)", len(flags), values.Len()))
		return values, nil, false
	}

	picked := make([]any, 0, values.Len())
	for i, flag := range flags {
		if flag == keep {
			picked = append(picked, values.At(i))
		}
	}
	return values, picked, true
}

func (f FilterNode) Kept(out *nodes.Dynamic[[]Element]) {
	if values, picked, ok := f.split(out, true); ok {
		out.Set(values.Build(picked))
	}
}

func (FilterNode) KeptDescription() string {
	return "The elements whose flag is true."
}

func (f FilterNode) Dropped(out *nodes.Dynamic[[]Element]) {
	if values, picked, ok := f.split(out, false); ok {
		out.Set(values.Build(picked))
	}
}

func (FilterNode) DroppedDescription() string {
	return "The elements whose flag is false."
}

// ============================================================================

type SplitNode struct {
	Array   nodes.DynamicPort[[]Element] `description:"The array to cut up."`
	Lengths nodes.Output[[]int]          `description:"How many elements each run takes, in order. They add up to the array's length."`
}

func (SplitNode) Description() string {
	return "Cuts an array into consecutive runs of the given lengths."
}

func (SplitNode) Keywords() []string {
	return []string{"unflatten"}
}

func (s SplitNode) Out(out *nodes.Dynamic[[][]Element]) {
	values, ok := nodes.DynamicArrayValue(out, s.Array)
	if !ok {
		return
	}

	lengths := nodes.TryGetOutputValue(out, s.Lengths, nil)
	total := 0
	for _, length := range lengths {
		if length < 0 {
			out.CaptureError(fmt.Errorf("run length %d is negative", length))
			return
		}
		total += length
	}
	if total != values.Len() {
		out.CaptureError(fmt.Errorf("lengths add up to %d, not %d", total, values.Len()))
		return
	}

	runs := make([]any, 0, len(lengths))
	start := 0
	for _, length := range lengths {
		runs = append(runs, values.Range(start, start+length))
		start += length
	}

	built, ok := nodes.DynamicArrayOf(s.Array, runs)
	if !ok {
		return
	}
	out.Set(built)
}

// ============================================================================

type GroupNode struct {
	Array  nodes.DynamicPort[[]Element] `description:"The array to sort into groups."`
	Groups nodes.Output[[]int]          `description:"Which group each element joins, counting from 0. -1 leaves it out."`
	Count  nodes.Output[int]            `description:"How many groups to make. Empty groups are kept, so group i is always output i."`
}

func (GroupNode) Description() string {
	return "Sorts an array's elements into numbered groups, keeping their order within each group. Elements in group -1 are dropped."
}

func (GroupNode) Keywords() []string {
	return []string{"bucket", "partition"}
}

func (g GroupNode) Out(out *nodes.Dynamic[[][]Element]) {
	values, ok := nodes.DynamicArrayValue(out, g.Array)
	if !ok {
		return
	}

	groups := nodes.TryGetOutputValue(out, g.Groups, nil)
	if len(groups) != values.Len() {
		out.CaptureError(fmt.Errorf("%d group(s) for %d element(s)", len(groups), values.Len()))
		return
	}

	count := nodes.GetOutputValue(out, g.Count)
	if count < 0 {
		out.CaptureError(fmt.Errorf("group count %d is negative", count))
		return
	}

	members := make([][]any, count)
	for i := range members {
		members[i] = make([]any, 0)
	}
	for i, group := range groups {
		if group == -1 {
			continue
		}
		if group < 0 || group >= count {
			out.CaptureError(fmt.Errorf("element %d is in group %d, outside 0 to %d", i, group, count-1))
			return
		}
		members[group] = append(members[group], values.At(i))
	}

	built := make([]any, count)
	for i, picked := range members {
		built[i] = values.Build(picked)
	}

	result, ok := nodes.DynamicArrayOf(g.Array, built)
	if !ok {
		return
	}
	out.Set(result)
}

// ============================================================================

type ZipNode struct {
	Arrays []nodes.DynamicPort[[]Element] `description:"Arrays of equal length, in order."`
}

func (ZipNode) Description() string {
	return "Turns parallel arrays into one array per position: element i of every input, in input order."
}

func (ZipNode) Keywords() []string {
	return []string{"transpose"}
}

func (z ZipNode) Out(out *nodes.Dynamic[[][]Element]) {
	var columns []nodes.DynamicArray
	var sample nodes.OutputPort
	for _, port := range z.Arrays {
		for _, column := range nodes.DynamicArrayValues(out, port) {
			if sample == nil {
				sample = port
			}
			columns = append(columns, column)
		}
	}
	if len(columns) == 0 {
		return
	}

	length := columns[0].Len()
	for _, column := range columns[1:] {
		if column.Len() != length {
			out.CaptureError(fmt.Errorf("arrays of %d and %d element(s) cannot be zipped", length, column.Len()))
			return
		}
	}

	rows := make([]any, length)
	for i := range rows {
		row := make([]any, len(columns))
		for j, column := range columns {
			row[j] = column.At(i)
		}
		rows[i] = columns[0].Build(row)
	}

	built, ok := nodes.DynamicArrayOf(sample, rows)
	if !ok {
		return
	}
	out.Set(built)
}
