package variable

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

const variableReferenceNodeOutputPortName = "Value"

type VariableReferenceNode[T any] struct {
	variable TypeVariable[T]
}

func (vrn *VariableReferenceNode[T]) Outputs() map[string]nodes.OutputPort {
	return map[string]nodes.OutputPort{
		variableReferenceNodeOutputPortName: &variableReferenceNodePort[T]{
			node: vrn,
		},
	}
}

func (vrn *VariableReferenceNode[T]) Inputs() map[string]nodes.InputPort {
	return nil
}

func (vrn *VariableReferenceNode[T]) Name() string {
	return vrn.variable.name
}

func (vrn *VariableReferenceNode[T]) Description() string {
	return vrn.variable.description
}

// ============================================================================
// Output Port Interface Implementation =======================================
// ============================================================================
//
// type Output interface {
// 	  Node() Node
// 	  Name() string
// 	  Version() int
// 	  Value() T
// }
//
// ============================================================================

type variableReferenceNodePort[T any] struct {
	node *VariableReferenceNode[T]
}

func (vrn *variableReferenceNodePort[T]) Node() nodes.Node {
	return vrn.node
}

func (vrn *variableReferenceNodePort[T]) Name() string {
	return variableReferenceNodeOutputPortName
}

func (vrn *variableReferenceNodePort[T]) Version() int {
	return vrn.node.variable.Version()
}

func (vrn *variableReferenceNodePort[T]) Value() T {
	return vrn.node.variable.value
}

func (so variableReferenceNodePort[T]) Type() string {
	return refutil.GetTypeWithPackage(new(T))
}
