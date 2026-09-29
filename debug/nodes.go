// Package debug contains nodes for looking at what a graph is carrying.
package debug

import (
	"fmt"
	"reflect"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[ProbeNode]](factory)

	generator.RegisterTypes(factory)
}

type Probed nodes.DynamicType

type ProbeNode struct {
	In nodes.DynamicPort[Probed] `description:"The value to look at."`
}

func (ProbeNode) Description() string {
	return "Passes a value through untouched while reporting what it is as text."
}

func (ProbeNode) Keywords() []string {
	return []string{"inspect", "watch"}
}

func (p ProbeNode) Out(out *nodes.Dynamic[Probed]) {
	out.Forward(p.In)
}

func (ProbeNode) OutDescription() string {
	return "The value being looked at, passed through unchanged."
}

func (p ProbeNode) Type(out *nodes.StructOutput[string]) {
	if p.In == nil {
		return
	}

	if typed, ok := p.In.(nodes.Typed); ok {
		if name := typed.Type(); name != "" {
			out.Set(name)
			return
		}
	}

	value, ok := nodes.DynamicAnyValue(p.In)
	if !ok || value == nil {
		return
	}
	out.Set(reflect.TypeOf(value).String())
}

func (ProbeNode) TypeDescription() string {
	return "The name of the type flowing through, empty while nothing is connected."
}

func (p ProbeNode) Value(out *nodes.StructOutput[string]) {
	value, ok := nodes.DynamicAnyValue(p.In)
	if !ok {
		return
	}
	out.Set(fmt.Sprintf("%v", value))
}

func (ProbeNode) ValueDescription() string {
	return "The value rendered as text."
}

func (p ProbeNode) Connected(out *nodes.StructOutput[bool]) {
	out.Set(p.In != nil)
}

func (ProbeNode) ConnectedDescription() string {
	return "Whether anything is wired into the probe at all."
}
