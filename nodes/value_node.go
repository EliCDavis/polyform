package nodes

import (
	"fmt"

	"github.com/EliCDavis/polyform/refutil"
)

type ValueNodeOutput[T any] struct {
	Val *ValueNode[T]
}

func (sno ValueNodeOutput[T]) Value() T {
	return sno.Val.Value()
}

func (sno ValueNodeOutput[T]) Node() Node {
	return sno.Val
}

func (sno ValueNodeOutput[T]) Port() string {
	return "Out"
}

type ValueNode[T any] struct {
	VersionData
	subs  []Alertable
	value T
}

func Value[T any](startingValue T) *ValueNode[T] {
	return &ValueNode[T]{
		value: startingValue,
	}
}

func FuncValue[T any](f func() T) *ValueNode[T] {
	return &ValueNode[T]{
		value: f(),
	}
}

func (in ValueNode[T]) Name() string {
	switch any(in.value).(type) {
	case int, string, float32, float64:
		return fmt.Sprintf("%v", in.value)
	default:
		return "Value"
	}
}

func (in *ValueNode[T]) Node() Node {
	return in
}

func (in *ValueNode[T]) Port() string {
	return "Out"
}

func (in *ValueNode[T]) Out() ValueNodeOutput[T] {
	return ValueNodeOutput[T]{Val: in}
}

func (vn ValueNode[T]) SetInput(input string, output Output) {
	panic("input can not be set")
}

func (tn ValueNode[T]) Inputs() []Input {
	return []Input{}
}

func (tn *ValueNode[T]) Outputs() []Output {
	return []Output{
		{
			Type:       refutil.GetTypeWithPackage(new(T)),
			NodeOutput: ValueNodeOutput[T]{Val: tn},
		},
	}
}

func (in ValueNode[T]) Value() T {
	return in.value
}

func (in *ValueNode[T]) Set(value T) {
	in.value = value
	in.Increment()
	in.alertSubscribers()
}

func (v *ValueNode[T]) AddSubscription(a Alertable) {
	v.subs = append(v.subs, a)
}

func (v *ValueNode[T]) alertSubscribers() {
	for _, a := range v.subs {
		if a != nil {
			a.Alert(v.version, Processed)
		}
	}
}

func (v *ValueNode[T]) State() NodeState {
	return Processed
}

func (v *ValueNode[T]) Dependencies() []NodeDependency {
	return nil
}
