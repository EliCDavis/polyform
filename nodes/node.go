package nodes

type NodeState int

const (
	Stale NodeState = iota
	Processed
	Error
)

type Node interface {
	Versioned
	Stateful
	Dependent

	SetInput(input string, output Output)
	Outputs() []Output
	Inputs() []Input
}
