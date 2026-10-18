package graph

import (
	"fmt"
	"maps"
	"slices"

	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
)

// copyDefinition builds the graph a placement evaluates through: the
// definition's nodes under the same ids, wired the same way, with each input
// boundary reading what the placement is fed.
//
// A node with no inputs and no dynamic types is the definition's own node,
// not a copy: its outputs depend on nothing wired, so one instance serves
// every placement. That is every node with state of its own.
func (root *Instance) copyDefinition(placement *SubgraphInstanceNode) (*Graph, error) {
	definition, ok := root.subGraphs[placement.subGraphID]
	if !ok {
		return nil, fmt.Errorf("sub-graph %q does not exist", placement.subGraphID)
	}
	from := definition.instance

	clone := newGraph(root, "")
	copied := make(map[nodes.Node]nodes.Node, len(from.nodesByID))
	for _, id := range slices.Sorted(maps.Keys(from.nodesByID)) {
		original := from.nodesByID[id]
		node, err := root.copyNode(original, from.nodeTypeKeys[original], placement, id)
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

func (root *Instance) copyNode(original nodes.Node, typeKey string, placement *SubgraphInstanceNode, id string) (nodes.Node, error) {
	switch boundary := original.(type) {
	case *subgraph.InputNode:
		in := subgraph.NewInputNode(boundary.PortName, boundary.PortType)
		in.ReadFrom(subgraphInstanceInputPort{node: placement, id: id})
		return in, nil
	case *subgraph.OutputNode:
		return subgraph.NewOutputNode(boundary.PortName, boundary.PortType), nil
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
