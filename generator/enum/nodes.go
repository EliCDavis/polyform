package enum

import (
	"fmt"
	"slices"
	"strings"

	"github.com/EliCDavis/polyform/nodes"
)

type valueType nodes.DynamicType

type MemberNameNode struct {
	Value nodes.DynamicPort[valueType] `description:"An enum value, or a list of them, from a Go enum or one of this graph's enums."`
}

func (MemberNameNode) Description() string {
	return "The name of an enum value, as the editor shows it."
}

func (n MemberNameNode) Out(out *nodes.Lifted[string]) {
	eachMember(out, n.Value, func(name string, _ int) string { return name })
}

type MemberIndexNode struct {
	Value nodes.DynamicPort[valueType] `description:"An enum value, or a list of them, from a Go enum or one of this graph's enums."`
}

func (MemberIndexNode) Description() string {
	return "The position of an enum value among its enum's members, counting from 0."
}

func (MemberIndexNode) Keywords() []string {
	return []string{"ordinal", "switch"}
}

func (n MemberIndexNode) Out(out *nodes.Lifted[int]) {
	eachMember(out, n.Value, func(_ string, index int) int { return index })
}

func eachMember[R any](out *nodes.Lifted[R], port nodes.OutputPort, read func(name string, index int) R) {
	element := ""
	if typed, ok := port.(nodes.Typed); ok {
		element = strings.TrimPrefix(typed.Type(), "[]")
	}
	nodes.ZipDynamic(out, port, func(value any) R {
		name, index, err := memberOf(value, element)
		out.CaptureError(err)
		return read(name, index)
	})
}

func memberOf(value any, element string) (string, int, error) {
	if value == nil {
		return "", 0, nil
	}
	if choice, ok := value.(Choice); ok {
		return choice.Name, choice.Index, nil
	}
	if def, found := Lookup(element); found {
		name, _ := def.NameOf(value)
		return name, slices.Index(def.Names(), name), nil
	}
	return "", 0, fmt.Errorf("%T is not an enum", value)
}
