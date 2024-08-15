package nodes

type NodeOutput[T any] interface {
	NodeOutputReference
	Value() T
}

type Processor[T any] interface {
	Process() (T, error)
}

func TryGetOutputValue[T any](output NodeOutput[T], fallback T) T {
	if output == nil {
		return fallback
	}
	return output.Value()
}
