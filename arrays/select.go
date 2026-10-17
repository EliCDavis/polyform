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
	return []string{"pick", "lookup", "take", "select", "index"}
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
	return []string{"mask", "where", "select", "remove", "partition"}
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
	return "Cuts an array into consecutive runs of the given lengths, like a flat list of points into one path per run."
}

func (SplitNode) Keywords() []string {
	return []string{"runs", "group", "unflatten", "paths", "chunk"}
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

type ZipNode struct {
	Arrays []nodes.DynamicPort[[]Element] `description:"Arrays of equal length, in order."`
}

func (ZipNode) Description() string {
	return "Turns parallel arrays into one array per position: element i of every input, in input order. All first points, all second points and so on become one path each."
}

func (ZipNode) Keywords() []string {
	return []string{"transpose", "interleave", "columns", "rows", "paths"}
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
