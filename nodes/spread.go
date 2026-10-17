package nodes

import (
	"fmt"
	"time"
)

// A node whose dynamic outputs each carry a whole array while typed as one
// element, so a list input reads every element as if it had its own wire.
type SpreadingNode interface {
	SpreadsOutputs()
}

type spreadPort interface {
	Spreads() bool
}

type spreadValuesPort[T any] interface {
	SpreadValues() ([]T, bool)
}

func IsSpread(port OutputPort) bool {
	spread, ok := port.(spreadPort)
	return ok && spread.Spreads()
}

// RefuseSpread is for any input that reads one value, where a spread would
// silently collapse to a single element.
func RefuseSpread(port OutputPort, input string) error {
	if IsSpread(port) {
		return fmt.Errorf("input %q takes one value, not a spread", input)
	}
	return nil
}

func SpreadValues[T any](port OutputPort) ([]T, bool) {
	if !IsSpread(port) {
		return nil, false
	}
	spread, ok := port.(spreadValuesPort[T])
	if !ok {
		return nil, false
	}
	return spread.SpreadValues()
}

type spreadAnyPort interface {
	SpreadAny() any
}

// DynamicAnyValues is every value one wire carries: each element of a
// spread, or its one value.
func DynamicAnyValues(recorder ExecutionRecorder, port OutputPort) []any {
	defer timeRead(recorder, port, time.Now())
	return dynamicAnyValues(port)
}

func dynamicAnyValues(port OutputPort) []any {
	if !IsSpread(port) {
		if value, ok := dynamicAnyValue(port); ok {
			return []any{value}
		}
		return nil
	}
	spread, ok := port.(spreadAnyPort)
	if !ok {
		return nil
	}
	array, ok := dynamicArrayFrom(spread.SpreadAny())
	if !ok {
		return nil
	}
	return array.Values()
}

// DynamicArrayValues is every array one wire carries into an array input.
func DynamicArrayValues(recorder ExecutionRecorder, port OutputPort) []DynamicArray {
	defer timeRead(recorder, port, time.Now())
	if !IsSpread(port) {
		if array, ok := dynamicArrayValue(port); ok {
			return []DynamicArray{array}
		}
		return nil
	}
	var arrays []DynamicArray
	for _, value := range dynamicAnyValues(port) {
		if array, ok := dynamicArrayFrom(value); ok {
			arrays = append(arrays, array)
		}
	}
	return arrays
}

func (d *dynamicOutput) SpreadAny() any   { return d.AnyValue() }
func (d dynamicTyped[T]) SpreadAny() any { return d.source.AnyValue() }

func (d *dynamicOutput) Spreads() bool {
	_, ok := d.data.(SpreadingNode)
	return ok
}

func (d dynamicTyped[T]) Spreads() bool {
	return IsSpread(d.source)
}

func (d dynamicTyped[T]) SpreadValues() ([]T, bool) {
	if !d.Spreads() {
		return nil, false
	}
	values, ok := d.source.AnyValue().([]T)
	return values, ok
}
