package variable

import (
	"encoding/json"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

type TypeVariable[T any] struct {
	value       T
	name        string
	description string

	version int
}

func (tv TypeVariable[T]) Name() string {
	return tv.name
}

func (tv *TypeVariable[T]) SetName(name string) {
	tv.name = name
}

func (tv *TypeVariable[T]) SetValue(v T) {
	tv.value = v
	tv.version++
}

func (tv *TypeVariable[T]) GetValue() T {
	return tv.value
}

func (tv *TypeVariable[T]) Version() int {
	return tv.version
}

func (tv *TypeVariable[T]) SetDescription(description string) {
	tv.description = description
}

func (tv TypeVariable[T]) ApplyMessage(msg []byte) (bool, error) {
	return false, nil
}

func (tv TypeVariable[T]) ToMessage() []byte {
	return nil
}

func (tv TypeVariable[T]) NodeReference() nodes.Node {
	return &VariableReferenceNode[T]{
		variable: tv,
	}
}

func (tv TypeVariable[T]) MarshalJSON() ([]byte, error) {
	var t T
	return json.Marshal(typedVariableSchema[T]{
		variableSchemaBase: variableSchemaBase{
			Name:        tv.name,
			Type:        refutil.GetTypeName(t),
			Description: tv.description,
		},
		Value: tv.value,
	})
}
