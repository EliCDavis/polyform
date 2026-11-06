package logic

import (
	"reflect"

	"github.com/EliCDavis/polyform/nodes"
)

type EqualsNode struct {
	A nodes.DynamicPort[EqualsType] `description:"Compared with B."`
	B nodes.DynamicPort[EqualsType] `description:"Compared with A."`
}

func (EqualsNode) Description() string {
	return "True when A and B hold the same value."
}

func (EqualsNode) Keywords() []string {
	return []string{"compare", "same", "=="}
}

func (n EqualsNode) Out(out *nodes.StructOutput[bool]) {
	a, _ := nodes.DynamicAnyValue(out, n.A)
	b, _ := nodes.DynamicAnyValue(out, n.B)
	out.Set(reflect.DeepEqual(a, b))
}
