package graph

import (
	"fmt"
	"maps"
	"slices"

	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
)

// A node with no inputs and no dynamic types is shared, not copied: its
// outputs depend on nothing wired, and that is every node with state.
// feed is what each input boundary reads, by boundary id and declared type.
func (root *Instance) copyDefinition(placement *SubgraphInstanceNode, feed func(id, portType string) subgraph.External) (*Graph, error) {
	definition, ok := root.subGraphs[placement.subGraphID]
	if !ok {
		return nil, fmt.Errorf("sub-graph %q does not exist", placement.subGraphID)
	}
	from := definition.instance

	clone := newGraph(root, "")
	copied := make(map[nodes.Node]nodes.Node, len(from.nodesByID))
	for _, id := range slices.Sorted(maps.Keys(from.nodesByID)) {
		original := from.nodesByID[id]
		node, err := root.copyNode(original, from.nodeTypeKeys[original], feed, id)
		if err != nil {
			return nil, fmt.Errorf("copying sub-graph %q: node %s: %w", placement.subGraphID, id, err)
		}
		clone.register(node, id, from.nodeTypeKeys[original])
		copied[original] = node
	}

	for consumer, inputs := range from.reads {
		reads := make(map[string][]source, len(inputs))
		for input, sources := range inputs {
			mapped := make([]source, len(sources))
			for i, s := range sources {
				mapped[i] = source{node: copied[s.node], port: s.port}
			}
			reads[input] = mapped
		}
		clone.reads[copied[consumer]] = reads
	}

	if _, err := clone.settle(refuseConflicts); err != nil {
		return nil, fmt.Errorf("copying sub-graph %q: %w", placement.subGraphID, err)
	}
	return clone, nil
}

func (root *Instance) copyNode(original nodes.Node, typeKey string, feed func(id, portType string) subgraph.External, id string) (nodes.Node, error) {
	switch n := original.(type) {
	case *subgraph.InputNode:
		in := subgraph.NewInputNode(n.PortName, n.PortType)
		in.ReadFrom(feed(id, n.PortType))
		return in, nil
	case *subgraph.OutputNode:
		return subgraph.NewOutputNode(n.PortName, n.PortType), nil
	case *SubgraphInstanceNode:
		placed := NewRuntimeNode(root, n.subGraphID)
		placed.each = n.ForEach()
		return placed, nil
	}

	if len(original.Inputs()) == 0 {
		if dynamic, ok := original.(nodes.DynamicallyTyped); !ok || len(dynamic.DynamicTypes()) == 0 {
			return original, nil
		}
	}

	if typeKey == "" {
		typeKey = nodeTypeName(original)
	}
	if !root.typeFactory.KeyRegistered(typeKey) {
		return nil, fmt.Errorf("type %q is not registered", typeKey)
	}
	node, ok := root.typeFactory.New(typeKey).(nodes.Node)
	if !ok {
		return nil, fmt.Errorf("type %q did not create a node", typeKey)
	}
	return node, nil
}
