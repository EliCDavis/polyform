package nodes

type NodeState int

const (
	Stale NodeState = iota
	Processed
	Error
)

type Port interface {

	// Node the port belongs to
	Node() Node

	// Name of the port
	Name() string
}

type OutputPort interface {
	Port
	Version() int
}

// A port whose Name is a key rather than what to show.
type Labeled interface {
	Label() string
}

type Output[T any] interface {
	OutputPort
	Value() T
}

type InputPort interface {
	Port

	// Remove any connections to this port
	Clear()
}

type SingleValueInputPort interface {
	InputPort
	Value() OutputPort
	Set(port OutputPort) error
}

type ArrayValueInputPort interface {
	InputPort
	Value() []OutputPort
	Add(port OutputPort) error
	Remove(port OutputPort) error
	Replace(index int, port OutputPort) error
}

type Node interface {
	Outputs() map[string]OutputPort
	Inputs() map[string]InputPort
}
