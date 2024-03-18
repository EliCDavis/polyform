package nodes

import (
	"fmt"

	"github.com/EliCDavis/polyform/refutil"
)

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

func (tn ValueNode[T]) Inputs() []Input {
	return []Input{}
}

func (tn ValueNode[T]) Outputs() []Output {
	return []Output{
		{
			Name: "Data",
			Type: refutil.GetTypeWithPackage(*new(T)),
		},
	}
}

func (in ValueNode[T]) Data() T {
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
