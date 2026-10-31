package enum

import (
	"fmt"
	"slices"

	"github.com/EliCDavis/polyform/nodes"
)

type valueType nodes.DynamicType

type MemberNameNode struct {
	Value nodes.DynamicPort[valueType] `description:"An enum value, from a Go enum or one of this graph's enums."`
}

func (MemberNameNode) Description() string {
	return "The name of an enum value, as the editor shows it."
}

func (n MemberNameNode) Out(out *nodes.StructOutput[string]) {
	name, _, err := memberOf(out, n.Value)
	out.CaptureError(err)
	out.Set(name)
}

type MemberIndexNode struct {
	Value nodes.DynamicPort[valueType] `description:"An enum value, from a Go enum or one of this graph's enums."`
}

func (MemberIndexNode) Description() string {
	return "The position of an enum value among its enum's members, counting from 0."
}

func (MemberIndexNode) Keywords() []string {
	return []string{"ordinal", "switch"}
}

func (n MemberIndexNode) Out(out *nodes.StructOutput[int]) {
	_, index, err := memberOf(out, n.Value)
	out.CaptureError(err)
	out.Set(index)
}

func memberOf(out nodes.ExecutionRecorder, port nodes.OutputPort) (string, int, error) {
	if port == nil {
		return "", 0, nil
	}
	value, _ := nodes.DynamicAnyValue(out, port)
	if choice, ok := value.(Choice); ok {
		return choice.Name, choice.Index, nil
	}
	if typed, ok := port.(nodes.Typed); ok {
		if def, found := Lookup(typed.Type()); found {
			name, _ := def.NameOf(value)
			return name, slices.Index(def.Names(), name), nil
		}
	}
	return "", 0, fmt.Errorf("%T is not an enum", value)
}
