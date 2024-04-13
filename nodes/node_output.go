package nodes

type NodeOutput[T any] interface {
	ReferencesNode
	Value() T
}

type Processor[T any] interface {
	Process() (T, error)
}
