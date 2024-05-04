package nodes

type NodeOutput[T any] interface {
	NodeOutputReference
	Value() T
}

type Processor[T any] interface {
	Process() (T, error)
}
