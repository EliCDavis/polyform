package graph

import "github.com/EliCDavis/polyform/nodes"

type namedOutputEntry[T any] struct {
	node     nodes.Node
	portName string
	port     nodes.Output[T]
}

type namedOutputManager[T any] struct {
	namedPorts map[string]namedOutputEntry[T]
}

func newNamedOutputManager[T any]() *namedOutputManager[T] {
	return &namedOutputManager[T]{namedPorts: make(map[string]namedOutputEntry[T])}
}

func (pd *namedOutputManager[T]) NamePort(name, portName string, node nodes.Node, port nodes.Output[T]) {
	// Make sure this node + portName isn't already named something else. If so,
	// delete it
	for name, entry := range pd.namedPorts {
		if entry.node == node && entry.portName == portName {
			delete(pd.namedPorts, name)
		}
	}

	pd.namedPorts[name] = namedOutputEntry[T]{
		node:     node,
		portName: portName,
		port:     port,
	}
}

func (pd *namedOutputManager[T]) IsPortNamed(node nodes.Node, portName string) (string, bool) {
	for name, entry := range pd.namedPorts {
		if entry.node == node && entry.portName == portName {
			return name, true
		}
	}
	return "", false
}

func (pd *namedOutputManager[T]) DeleteNode(node nodes.Node) {
	for name, producer := range pd.namedPorts {
		if producer.node == node {
			delete(pd.namedPorts, name)
		}
	}
}
