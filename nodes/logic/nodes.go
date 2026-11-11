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
	return "Passes through A when Condition is true, otherwise B."
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

	a, aOk := nodes.DynamicArrayValue(out, s.A)
	b, bOk := nodes.DynamicArrayValue(out, s.B)
	if !aOk && !bOk {
		return
	}
	builder := a
	if !aOk {
		builder = b
	}
	length := func(ok bool, arr nodes.DynamicArray) int {
		if !ok {
			return len(flags)
		}
		return arr.Len()
	}
	if (length(aOk, a) != len(flags) && length(aOk, a) != 1) || (length(bOk, b) != len(flags) && length(bOk, b) != 1) {
		out.CaptureError(fmt.Errorf("%d condition(s) for A of %d and B of %d element(s); a side of one element applies to every condition", len(flags), length(aOk, a), length(bOk, b)))
		return
	}

	// A one-element side stands in for a single value: wire a list of one
	// to use the same value wherever that side is chosen.
	at := func(ok bool, arr nodes.DynamicArray, i int) any {
		if !ok {
			return nil
		}
		if arr.Len() == 1 {
			return arr.At(0)
		}
		return arr.At(i)
	}

	picked := make([]any, len(flags))
	for i, on := range flags {
		if on {
			picked[i] = at(aOk, a, i)
		} else {
			picked[i] = at(bOk, b, i)
		}
	}
	out.Set(builder.Build(picked))
}

// ============================================================================

type SwitchNode struct {
	Index    nodes.Output[int]               `description:"Which case to pass through, counting from 0."`
	Cases    []nodes.DynamicPort[SwitchType] `description:"The values to choose between, in order."`
	Fallback nodes.DynamicPort[SwitchType]   `description:"Passed through when Index falls outside Cases."`
}

func (SwitchNode) Description() string {
	return "Passes through the case at Index, or Fallback when Index is outside the list."
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
