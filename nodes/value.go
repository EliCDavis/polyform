package nodes

import "github.com/EliCDavis/polyform/refutil"

const valueOutputPortName = "Value"

// Implements Output[T any]
type valueOutputPort[T any] struct {
	Val *Value[T]
}

func (sno valueOutputPort[T]) Node() Node {
	return sno.Val
}

func (sno valueOutputPort[T]) Value() T {
	return sno.Val.value
}

func (sno valueOutputPort[T]) Name() string {
	return valueOutputPortName
}

func (sno valueOutputPort[T]) Version() int {
	return sno.Val.Version()
}

func (sno valueOutputPort[T]) Type() string {
	resolver := refutil.TypeResolution{
		IncludePackage:     true,
		IncludePointer:     true,
		StripSinglePointer: true,
	}
	return resolver.Resolve(new(T))
}

func (sno valueOutputPort[T]) BuildProxyOutput(source ProxySource) OutputPort {
	return NewProxyOutput[T](source)
}

func (sno valueOutputPort[T]) BuildDynamicOutput(source DynamicSource) OutputPort {
	return NewDynamicOutput[T](source)
}

func (sno valueOutputPort[T]) BuildDynamicArrayOutput(source DynamicSource) OutputPort {
	return NewDynamicOutput[[]T](source)
}

// ============================================================================

type Value[T any] struct {
	VersionData
	value T
}

func NewValue[T any](startingValue T) *Value[T] {
	return &Value[T]{
		value: startingValue,
	}
}

func (tn *Value[T]) Outputs() map[string]OutputPort {
	return map[string]OutputPort{
		valueOutputPortName: &valueOutputPort[T]{
			Val: tn,
		},
	}
}

func (tn *Value[T]) Inputs() map[string]InputPort {
	return nil
}

func (in *Value[T]) Set(value T) {
	in.value = value
	in.Increment()
}
