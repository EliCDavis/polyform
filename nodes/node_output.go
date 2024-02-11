package nodes

type NodeOutput[T any] interface {
	ReferencesNode
	Data() T
}
