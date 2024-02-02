package nodes

import "fmt"

type InputNode[T any] struct {
	VersionData
	subs  []Alertable
	value T
}

func Input[T any](startingValue T) *InputNode[T] {
	return &InputNode[T]{
		value: startingValue,
	}
}

func InputFromFunc[T any](f func() T) *InputNode[T] {
	return &InputNode[T]{
		value: f(),
	}
}

func (in InputNode[T]) Name() string {
	switch any(in.value).(type) {
	case int, string, float32, float64:
		return fmt.Sprintf("%v", in.value)
	default:
		return "Input"
	}
}

func (in InputNode[T]) Data() T {
	return in.value
}

func (in *InputNode[T]) Set(value T) {
	in.value = value
	in.version++
	in.alertSubscribers()
}

func (v *InputNode[T]) AddSubscription(a Alertable) {
	v.subs = append(v.subs, a)
}

func (v *InputNode[T]) alertSubscribers() {
	for _, a := range v.subs {
		if a != nil {
			a.Alert(v.version, Processed)
		}
	}
}

func (v *InputNode[T]) State() NodeState {
	return Processed
}

func (v *InputNode[T]) Dependencies() []NodeDependency {
	return nil
}
