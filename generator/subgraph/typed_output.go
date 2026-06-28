package subgraph

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

type typedInputOutput[T any] struct {
	node *InputNode
}

func (p typedInputOutput[T]) Node() nodes.Node {
	return p.node
}

func (p typedInputOutput[T]) Name() string {
	return ValuePortName
}

func (p typedInputOutput[T]) Type() string {
	return p.node.PortType
}

func (p typedInputOutput[T]) Version() int {
	if p.node.externalSource != nil {
		return p.node.externalSource.Version()
	}
	return p.node.version
}

func (p typedInputOutput[T]) Value() T {
	if p.node.externalSource != nil {
		if typed, ok := p.node.externalSource.(nodes.Output[T]); ok {
			return typed.Value()
		}
	}
	var zero T
	return zero
}

type inputOutputBuilder func(*InputNode) nodes.OutputPort

var inputOutputBuilders = map[string]inputOutputBuilder{}

func RegisterInputOutputType[T any](typeKeys ...string) {
	resolver := refutil.TypeResolution{
		IncludePackage: true,
		IncludePointer: false,
	}
	defaultKey := resolver.Resolve(new(T))
	keys := append([]string{defaultKey}, typeKeys...)
	builder := func(n *InputNode) nodes.OutputPort {
		return typedInputOutput[T]{node: n}
	}
	for _, key := range keys {
		if key == "" {
			continue
		}
		inputOutputBuilders[key] = builder
	}
}

func buildInputOutputPort(n *InputNode) nodes.OutputPort {
	if n.PortType == "" {
		return &inputNodeOutputPort{node: n}
	}
	if builder, ok := inputOutputBuilders[n.PortType]; ok {
		return builder(n)
	}
	return &inputNodeOutputPort{node: n}
}

func init() {
	RegisterInputOutputType[float64]()
	RegisterInputOutputType[int]()
	RegisterInputOutputType[bool]()
	RegisterInputOutputType[string]()
	RegisterInputOutputType[float32]()
}
