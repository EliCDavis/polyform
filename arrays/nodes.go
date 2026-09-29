// Package arrays contains array utilities that work with any element type,
// each registered once rather than once per type.
package arrays

import (
	"fmt"
	"reflect"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/math/chance"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[IndexNode]](factory)
	refutil.RegisterType[nodes.Struct[FirstNode]](factory)
	refutil.RegisterType[nodes.Struct[LastNode]](factory)
	refutil.RegisterType[nodes.Struct[LengthNode]](factory)
	refutil.RegisterType[nodes.Struct[SliceNode]](factory)
	refutil.RegisterType[nodes.Struct[ReverseNode]](factory)
	refutil.RegisterType[nodes.Struct[AppendNode]](factory)
	refutil.RegisterType[nodes.Struct[ConcatNode]](factory)
	refutil.RegisterType[nodes.Struct[FromElementsNode]](factory)
	refutil.RegisterType[nodes.Struct[RepeatNode]](factory)
	refutil.RegisterType[nodes.Struct[ChunkNode]](factory)
	refutil.RegisterType[nodes.Struct[PartitionNode]](factory)
	refutil.RegisterType[nodes.Struct[IndexOfNode]](factory)
	refutil.RegisterType[nodes.Struct[ShuffleNode]](factory)
	refutil.RegisterType[nodes.Struct[SampleNode]](factory)
	refutil.RegisterType[nodes.Struct[PadNode]](factory)

	generator.RegisterTypes(factory)
}

type Element nodes.DynamicType

type IndexNode struct {
	Array nodes.DynamicPort[[]Element] `description:"The array to read from."`
	Index nodes.Output[int]            `description:"Which element to read, counting from 0. Negative counts back from the end."`
}

func (IndexNode) Description() string {
	return "Reads one element out of an array."
}

func (IndexNode) Keywords() []string {
	return []string{"at", "get"}
}

func (i IndexNode) Out(out *nodes.Dynamic[Element]) {
	values, ok := nodes.DynamicArrayValue(i.Array)
	if !ok {
		return
	}

	index := nodes.TryGetOutputValue(out, i.Index, 0)
	if index < 0 {
		index += values.Len()
	}
	if index < 0 || index >= values.Len() {
		out.CaptureError(fmt.Errorf("index %d is outside an array of %d element(s)", index, values.Len()))
		return
	}
	out.Set(values.At(index))
}

// ============================================================================

type FirstNode struct {
	Array nodes.DynamicPort[[]Element] `description:"The array to read from."`
}

func (FirstNode) Description() string {
	return "Reads the first element of an array."
}

func (FirstNode) Keywords() []string { return []string{"head", "front"} }

func (f FirstNode) Out(out *nodes.Dynamic[Element]) {
	values, ok := nodes.DynamicArrayValue(f.Array)
	if !ok || values.Len() == 0 {
		return
	}
	out.Set(values.At(0))
}

// ============================================================================

type LastNode struct {
	Array nodes.DynamicPort[[]Element] `description:"The array to read from."`
}

func (LastNode) Description() string {
	return "Reads the last element of an array."
}

func (LastNode) Keywords() []string { return []string{"tail", "end"} }

func (l LastNode) Out(out *nodes.Dynamic[Element]) {
	values, ok := nodes.DynamicArrayValue(l.Array)
	if !ok || values.Len() == 0 {
		return
	}
	out.Set(values.At(values.Len() - 1))
}

// ============================================================================

type LengthNode struct {
	Array nodes.DynamicPort[[]Element] `description:"The array to measure."`
}

func (LengthNode) Description() string {
	return "How many elements an array holds."
}

func (LengthNode) Keywords() []string { return []string{"count", "size"} }

func (l LengthNode) Out(out *nodes.StructOutput[int]) {
	values, ok := nodes.DynamicArrayValue(l.Array)
	if !ok {
		out.Set(0)
		return
	}
	out.Set(values.Len())
}

// ============================================================================

type SliceNode struct {
	Array nodes.DynamicPort[[]Element] `description:"The array to take a run out of."`
	Start nodes.Output[int]            `description:"First element to keep, counting from 0."`
	End   nodes.Output[int]            `description:"One past the last element to keep. Left at 0 it means the end of the array."`
}

func (SliceNode) Description() string {
	return "Takes a contiguous run out of an array, clamped to what is there."
}

func (SliceNode) Keywords() []string { return []string{"range", "trim"} }

func (s SliceNode) Out(out *nodes.Dynamic[[]Element]) {
	values, ok := nodes.DynamicArrayValue(s.Array)
	if !ok {
		return
	}

	start := nodes.TryGetOutputValue(out, s.Start, 0)
	end := nodes.TryGetOutputValue(out, s.End, 0)
	if s.End == nil || end == 0 {
		end = values.Len()
	}
	if start < 0 {
		start += values.Len()
	}
	if end < 0 {
		end += values.Len()
	}
	out.Set(values.Range(start, end))
}

// ============================================================================

type ReverseNode struct {
	Array nodes.DynamicPort[[]Element] `description:"The array to reverse."`
}

func (ReverseNode) Description() string {
	return "Reverses the order of an array."
}

func (ReverseNode) Keywords() []string { return []string{"flip", "backwards"} }

func (r ReverseNode) Out(out *nodes.Dynamic[[]Element]) {
	values, ok := nodes.DynamicArrayValue(r.Array)
	if !ok {
		return
	}

	reversed := values.Values()
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	out.Set(values.Build(reversed))
}

// ============================================================================

type AppendNode struct {
	Array    nodes.DynamicPort[[]Element] `description:"The array to add to."`
	Elements []nodes.DynamicPort[Element] `description:"The values to add to the end, in order."`
}

func (AppendNode) Description() string {
	return "Adds elements to the end of an array."
}

func (AppendNode) Keywords() []string { return []string{"add", "push"} }

func (a AppendNode) Out(out *nodes.Dynamic[[]Element]) {
	values, ok := nodes.DynamicArrayValue(a.Array)
	if !ok {
		return
	}

	appended := values.Values()
	for _, element := range a.Elements {
		value, ok := nodes.DynamicAnyValue(element)
		if !ok {
			continue
		}
		appended = append(appended, value)
	}
	out.Set(values.Build(appended))
}

// ============================================================================

type ConcatNode struct {
	Arrays []nodes.DynamicPort[[]Element] `description:"The arrays to join, in order."`
}

func (ConcatNode) Description() string {
	return "Joins several arrays into one, in order."
}

func (ConcatNode) Keywords() []string {
	return []string{"join", "merge"}
}

func (c ConcatNode) Out(out *nodes.Dynamic[[]Element]) {
	var joined []any
	var first nodes.DynamicArray
	found := false

	for _, port := range c.Arrays {
		values, ok := nodes.DynamicArrayValue(port)
		if !ok {
			continue
		}
		if !found {
			first, found = values, true
		}
		joined = append(joined, values.Values()...)
	}

	if !found {
		return
	}
	out.Set(first.Build(joined))
}

// ============================================================================

type FromElementsNode struct {
	Elements []nodes.DynamicPort[Element] `description:"The values to collect into an array, in order."`
}

func (FromElementsNode) Description() string {
	return "Collects individually wired values into one array, in order."
}

func (FromElementsNode) Keywords() []string {
	return []string{"collect", "gather"}
}

func (f FromElementsNode) Out(out *nodes.Dynamic[[]Element]) {
	if len(f.Elements) == 0 {
		return
	}

	values := make([]any, 0, len(f.Elements))
	for _, element := range f.Elements {
		value, ok := nodes.DynamicAnyValue(element)
		if !ok {
			continue
		}
		values = append(values, value)
	}

	built, ok := nodes.DynamicArrayOf(f.Elements[0], values)
	if !ok {
		return
	}
	out.Set(built)
}

// ============================================================================

type RepeatNode struct {
	Value nodes.DynamicPort[Element] `description:"The value to repeat."`
	Times nodes.Output[int]          `description:"How many copies the array should hold."`
}

func (RepeatNode) Description() string {
	return "Builds an array holding the same value several times."
}

func (RepeatNode) Keywords() []string { return []string{"duplicate", "fill"} }

func (r RepeatNode) Out(out *nodes.Dynamic[[]Element]) {
	value, ok := nodes.DynamicAnyValue(r.Value)
	if !ok {
		return
	}

	times := max(nodes.TryGetOutputValue(out, r.Times, 0), 0)
	values := make([]any, times)
	for i := range values {
		values[i] = value
	}

	built, ok := nodes.DynamicArrayOf(r.Value, values)
	if !ok {
		return
	}
	out.Set(built)
}

// ============================================================================

type ChunkNode struct {
	Array nodes.DynamicPort[[]Element] `description:"The array to split up."`
	Size  nodes.Output[int]            `description:"How many elements each chunk holds."`
}

func (ChunkNode) Description() string {
	return "Splits an array into runs of a fixed size, the last one shorter when the size does not divide evenly."
}

func (ChunkNode) Keywords() []string {
	return []string{"batch", "group"}
}

func (c ChunkNode) Out(out *nodes.Dynamic[[][]Element]) {
	values, ok := nodes.DynamicArrayValue(c.Array)
	if !ok {
		return
	}

	size := nodes.TryGetOutputValue(out, c.Size, 0)
	if size < 1 {
		out.CaptureError(fmt.Errorf("chunk size has to be at least 1, got %d", size))
		return
	}

	chunks := make([]any, 0, (values.Len()+size-1)/size)
	for i := 0; i < values.Len(); i += size {
		chunks = append(chunks, values.Range(i, i+size))
	}

	built, ok := nodes.DynamicArrayOf(c.Array, chunks)
	if !ok {
		return
	}
	out.Set(built)
}

// ============================================================================

type PartitionNode struct {
	Array nodes.DynamicPort[[]Element] `description:"The array to split in two."`
	At    nodes.Output[int]            `description:"How many elements go to Before. Negative counts back from the end."`
}

func (PartitionNode) Description() string {
	return "Splits an array in two at a position."
}

func (PartitionNode) Keywords() []string {
	return []string{"split", "halve"}
}

func (p PartitionNode) at(recorder nodes.ExecutionRecorder, length int) int {
	at := nodes.TryGetOutputValue(recorder, p.At, 0)
	if at < 0 {
		at += length
	}
	return min(max(at, 0), length)
}

func (p PartitionNode) Before(out *nodes.Dynamic[[]Element]) {
	values, ok := nodes.DynamicArrayValue(p.Array)
	if !ok {
		return
	}
	out.Set(values.Range(0, p.at(out, values.Len())))
}

func (PartitionNode) BeforeDescription() string {
	return "The elements up to the split."
}

func (p PartitionNode) After(out *nodes.Dynamic[[]Element]) {
	values, ok := nodes.DynamicArrayValue(p.Array)
	if !ok {
		return
	}
	out.Set(values.Range(p.at(out, values.Len()), values.Len()))
}

func (PartitionNode) AfterDescription() string {
	return "The elements from the split onwards."
}

// ============================================================================

type IndexOfNode struct {
	Array nodes.DynamicPort[[]Element] `description:"The array to search."`
	Value nodes.DynamicPort[Element]   `description:"The element to look for."`
}

func (IndexOfNode) Description() string {
	return "Finds where a value sits in an array."
}

func (IndexOfNode) Keywords() []string {
	return []string{"find", "search"}
}

func (i IndexOfNode) search() int {
	values, ok := nodes.DynamicArrayValue(i.Array)
	if !ok {
		return -1
	}

	target, ok := nodes.DynamicAnyValue(i.Value)
	if !ok {
		return -1
	}

	for index := range values.Len() {
		if reflect.DeepEqual(values.At(index), target) {
			return index
		}
	}
	return -1
}

func (i IndexOfNode) Index(out *nodes.StructOutput[int]) {
	out.Set(i.search())
}

func (IndexOfNode) IndexDescription() string {
	return "The position of the first match, counting from 0, or -1 when there is none."
}

func (i IndexOfNode) Found(out *nodes.StructOutput[bool]) {
	out.Set(i.search() >= 0)
}

func (IndexOfNode) FoundDescription() string {
	return "Whether the value is in the array at all."
}

// ============================================================================

type ShuffleNode struct {
	Array nodes.DynamicPort[[]Element] `description:"The array to reorder."`
	Seed  nodes.Output[int]            `description:"The same seed always produces the same order."`
}

func (ShuffleNode) Description() string {
	return "Reorders an array randomly."
}

func (ShuffleNode) Keywords() []string {
	return []string{"random", "scramble"}
}

func (s ShuffleNode) Out(out *nodes.Dynamic[[]Element]) {
	values, ok := nodes.DynamicArrayValue(s.Array)
	if !ok {
		return
	}

	shuffled := values.Values()
	random := chance.FromSeed(nodes.TryGetOutputValue(out, s.Seed, 0))
	random.Shuffle(len(shuffled), func(a, b int) {
		shuffled[a], shuffled[b] = shuffled[b], shuffled[a]
	})

	out.Set(values.Build(shuffled))
}

// ============================================================================

type SampleNode struct {
	Array nodes.DynamicPort[[]Element] `description:"The array to pick from."`
	Count nodes.Output[int]            `description:"How many elements to pick, capped at what the array holds."`
	Seed  nodes.Output[int]            `description:"The same seed always picks the same elements."`
}

func (SampleNode) Description() string {
	return "Picks elements at random from an array, never the same one twice."
}

func (SampleNode) Keywords() []string {
	return []string{"random", "pick"}
}

func (s SampleNode) Out(out *nodes.Dynamic[[]Element]) {
	values, ok := nodes.DynamicArrayValue(s.Array)
	if !ok {
		return
	}

	pool := values.Values()
	count := min(max(nodes.TryGetOutputValue(out, s.Count, 0), 0), len(pool))
	random := chance.FromSeed(nodes.TryGetOutputValue(out, s.Seed, 0))

	picked := make([]any, 0, count)
	for range count {
		index := random.IntN(len(pool))
		picked = append(picked, pool[index])
		pool = append(pool[:index], pool[index+1:]...)
	}

	out.Set(values.Build(picked))
}

// ============================================================================

type PadNode struct {
	Array  nodes.DynamicPort[[]Element] `description:"The array to extend."`
	Length nodes.Output[int]            `description:"The length to reach. An array already that long is left alone."`
	Value  nodes.DynamicPort[Element]   `description:"The value to repeat on the end. Left unwired, the type's zero is used."`
}

func (PadNode) Description() string {
	return "Extends an array to a length by repeating a value on the end."
}

func (PadNode) Keywords() []string {
	return []string{"extend", "grow"}
}

func (p PadNode) Out(out *nodes.Dynamic[[]Element]) {
	values, ok := nodes.DynamicArrayValue(p.Array)
	if !ok {
		return
	}

	padded := values.Values()
	length := nodes.TryGetOutputValue(out, p.Length, 0)
	if length > len(padded) {
		value, _ := nodes.DynamicAnyValue(p.Value)
		for len(padded) < length {
			padded = append(padded, value)
		}
	}

	out.Set(values.Build(padded))
}
