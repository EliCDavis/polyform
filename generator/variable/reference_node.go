package variable

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

const variableReferenceNodeOutputPortName = "Value"

type VariableReferenceNode[T any] struct {
	variable Variable
}

func (vrn *VariableReferenceNode[T]) Outputs() map[string]nodes.OutputPort {
	return map[string]nodes.OutputPort{
		variableReferenceNodeOutputPortName: &variableReferenceNodePort[T]{
			node: vrn,
		},
	}
}

func (vrn *VariableReferenceNode[T]) Reference() Variable {
	return vrn.variable
}

func (vrn *VariableReferenceNode[T]) Inputs() map[string]nodes.InputPort {
	return nil
}

func (tv *VariableReferenceNode[T]) Name() string {
	return tv.variable.Info().Name()
}

func (tv *VariableReferenceNode[T]) Description() string {
	return tv.variable.Info().Description()
}

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
	return vrn.node.variable.currentVersion()
}

func (vrn *variableReferenceNodePort[T]) Value() T {
	return vrn.node.variable.currentValue().(T)
}

func (so variableReferenceNodePort[T]) Type() string {
	return refutil.GetTypeWithPackage(new(T))
}
