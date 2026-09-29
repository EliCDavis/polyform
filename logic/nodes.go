package logic

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[SelectNode]](factory)
	refutil.RegisterType[nodes.Struct[SwitchNode]](factory)

	generator.RegisterTypes(factory)
}

type (
	SelectType nodes.DynamicType
	SwitchType nodes.DynamicType
)

type SelectNode struct {
	Condition nodes.Output[bool]            `description:"Which of the two inputs to pass through."`
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
	if nodes.TryGetOutputValue(out, s.Condition, false) {
		out.Forward(s.A)
		return
	}
	out.Forward(s.B)
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
	index := nodes.TryGetOutputValue(out, s.Index, 0)
	if index < 0 || index >= len(s.Cases) {
		out.Forward(s.Fallback)
		return
	}
	out.Forward(s.Cases[index])
}

