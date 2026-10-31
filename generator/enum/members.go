package enum

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

const membersOutputPortName = "Members"

type Members[E comparable] struct {
	set Set[E]
}

func NewMembers[E comparable](set Set[E]) Members[E] {
	return Members[E]{set: set}
}

func (m Members[E]) Name() string { return shortName[E]() + " Members" }

func (m Members[E]) Description() string {
	return "Every member of " + shortName[E]() + ", in declared order."
}

func (m *Members[E]) Outputs() map[string]nodes.OutputPort {
	return map[string]nodes.OutputPort{
		membersOutputPortName: membersOutput[E]{node: m},
	}
}

func (m Members[E]) Inputs() map[string]nodes.InputPort { return nil }

type membersOutput[E comparable] struct {
	node *Members[E]
}

func (o membersOutput[E]) Value() []E { return o.node.set.Values() }

func (o membersOutput[E]) Node() nodes.Node { return o.node }

func (o membersOutput[E]) Name() string { return membersOutputPortName }

func (o membersOutput[E]) Version() int { return 0 }

func (o membersOutput[E]) Type() string {
	return refutil.TypeResolution{IncludePackage: true}.Resolve(*new([]E))
}

func (o membersOutput[E]) BuildProxyOutput(source nodes.ProxySource) nodes.OutputPort {
	return nodes.NewProxyOutput[[]E](source)
}
