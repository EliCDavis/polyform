package nodes

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

func FuncValue[T any](f func() T) *Value[T] {
	return &Value[T]{
		value: f(),
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
