package logic

import (
	"fmt"

	"github.com/EliCDavis/polyform/generator/registry"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[SelectNode]](factory)
	refutil.RegisterType[nodes.Struct[SwitchNode]](factory)
	refutil.RegisterType[nodes.Struct[EqualsNode]](factory)
	refutil.RegisterType[nodes.Struct[AndNode]](factory)
	refutil.RegisterType[nodes.Struct[OrNode]](factory)

	registry.Register(factory)
}

type (
	SelectType nodes.DynamicType
	SwitchType nodes.DynamicType
	EqualsType nodes.DynamicType
)

type SelectNode struct {
	Condition nodes.LiftedPort[bool]        `description:"A when true, B when false."`
	A         nodes.DynamicPort[SelectType] `description:"Passed through when Condition is true."`
	B         nodes.DynamicPort[SelectType] `description:"Passed through when Condition is false."`
}

func (SelectNode) Description() string {
	return "Passes through one of two values."
}

func (SelectNode) Keywords() []string {
	return []string{"if", "ternary"}
}

func (s SelectNode) Out(out *nodes.Dynamic[SelectType]) {
	condition, _ := nodes.DynamicAnyValue(out, s.Condition)
	flags, perElement := condition.([]bool)
	if !perElement {
		if on, _ := condition.(bool); on {
			out.Forward(s.A)
			return
		}
		out.Forward(s.B)
		return
	}

	a := readSelectSide(out, s.A, len(flags))
	b := readSelectSide(out, s.B, len(flags))
	if a.length != len(flags) && a.length != 1 || b.length != len(flags) && b.length != 1 {
		out.CaptureError(fmt.Errorf("%d condition(s) for A of %d and B of %d element(s); a side of one element applies to every condition", len(flags), a.length, b.length))
		return
	}

	picked := make([]any, len(flags))
	for i, on := range flags {
		if on {
			picked[i] = a.at(i)
		} else {
			picked[i] = b.at(i)
		}
	}

	switch {
	case a.isArray:
		out.Set(a.array.Build(picked))
	case b.isArray:
		out.Set(b.array.Build(picked))
	case s.A != nil:
		built, _ := nodes.DynamicArrayOf(s.A, picked)
		out.Set(built)
	case s.B != nil:
		built, _ := nodes.DynamicArrayOf(s.B, picked)
		out.Set(built)
	}
}

type selectSide struct {
	array   nodes.DynamicArray
	isArray bool
	value   any
	length  int
}

func readSelectSide(out *nodes.Dynamic[SelectType], port nodes.OutputPort, unwiredLength int) selectSide {
	if port == nil {
		return selectSide{length: unwiredLength}
	}
	if array, ok := nodes.DynamicArrayValue(out, port); ok {
		return selectSide{array: array, isArray: true, length: array.Len()}
	}
	value, _ := nodes.DynamicAnyValue(out, port)
	return selectSide{value: value, length: 1}
}

func (s selectSide) at(i int) any {
	if !s.isArray {
		return s.value
	}
	if s.array.Len() == 1 {
		return s.array.At(0)
	}
	return s.array.At(i)
}

// ============================================================================

type SwitchNode struct {
	Index    nodes.Output[int]               `description:"Which case to pass through, counting from 0."`
	Cases    []nodes.DynamicPort[SwitchType] `description:"The values to choose between, in order."`
	Fallback nodes.DynamicPort[SwitchType]   `description:"Passed through when Index falls outside Cases."`
}

func (SwitchNode) Description() string {
	return "Passes through one of a list of cases."
}

func (SwitchNode) Keywords() []string {
	return []string{"case", "mux"}
}

func (s SwitchNode) Out(out *nodes.Dynamic[SwitchType]) {
	index := nodes.GetOutputValue(out, s.Index)
	if index < 0 {
		out.Forward(s.Fallback)
		return
	}
	for _, c := range s.Cases {
		if !nodes.IsSpread(c) {
			if index == 0 {
				out.Forward(c)
				return
			}
			index--
			continue
		}
		values := nodes.DynamicAnyValues(out, c)
		if index < len(values) {
			out.Set(values[index])
			return
		}
		index -= len(values)
	}
	out.Forward(s.Fallback)
}
