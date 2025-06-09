package variable

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

const variableReferenceNodeOutputPortName = "Value"

type VariableReferenceNode[T any] struct {
	variable *TypeVariable[T]
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
	return tv.variable.info.Name()
}

func (tv *VariableReferenceNode[T]) Description() string {
	return tv.variable.Info().Description()
}

// CUSTOM JTF Serialization ===================================================

// func (pn *VariableReferenceNode[T]) ToJSON(encoder *jbtf.Encoder) ([]byte, error) {
// 	return json.Marshal(parameterNodeGraphSchema[T]{
// 		Name:         pn.Name,
// 		Description:  pn.Description,
// 		CurrentValue: pn.Value(),
// 		DefaultValue: pn.DefaultValue,
// 		CLI:          pn.CLI,
// 	})
// }

// func (pn *VariableReferenceNode[T]) FromJSON(decoder jbtf.Decoder, body []byte) (err error) {
// 	gn := parameterNodeGraphSchema[T]{}
// 	err = json.Unmarshal(body, &gn)
// 	if err != nil {
// 		return
// 	}

// 	pn.Name = gn.Name
// 	pn.Description = gn.Description
// 	pn.DefaultValue = gn.DefaultValue
// 	pn.CLI = gn.CLI
// 	pn.appliedProfile = &gn.CurrentValue
// 	return
// }

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
