package graph

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
)

func (a *Graph) register(node nodes.Node, id, typeKey string) {
	a.nodeIDs[node] = id
	a.nodesByID[id] = node
	if typeKey != "" {
		a.nodeTypeKeys[node] = typeKey
	}

	number, isNumbered := strings.CutPrefix(id, "Node-")
	if n, err := strconv.Atoi(number); isNumbered && err == nil && n+1 > a.nodeIDHighWater {
		a.nodeIDHighWater = n + 1
	}
}

func (a *Graph) forget(node nodes.Node) {
	delete(a.nodesByID, a.nodeIDs[node])
	delete(a.nodeIDs, node)
	delete(a.nodeTypeKeys, node)
	delete(a.reads, node)
}

func (a *Graph) nextNodeID() string {
	number := max(len(a.nodeIDs), a.nodeIDHighWater)
	for a.nodesByID[fmt.Sprintf("Node-%d", number)] != nil {
		number++
	}
	return fmt.Sprintf("Node-%d", number)
}

// A node built in code can arrive already holding the outputs of others.
// Those get ids first, and what it holds becomes its edges.
func (a *Graph) registerWithInputs(node nodes.Node, typeKey string) {
	if _, ok := a.nodeIDs[node]; ok {
		return
	}

	reads := make(map[string][]source)
	inputs := node.Inputs()
	// Sorted: the nodes this registers take ids in this order.
	for _, name := range slices.Sorted(maps.Keys(inputs)) {
		input := inputs[name]
		var held []nodes.OutputPort
		switch slot := input.(type) {
		case nodes.SingleValueInputPort:
			held = []nodes.OutputPort{slot.Value()}
		case nodes.ArrayValueInputPort:
			held = slot.Value()
		}
		for _, port := range held {
			if port != nil {
				a.registerWithInputs(port.Node(), "")
				reads[name] = append(reads[name], source{node: port.Node(), port: port.Name()})
			}
		}
	}

	a.register(node, a.nextNodeID(), typeKey)
	a.reads[node] = reads
	a.dirty()
}

func (a *Graph) NodeId(node nodes.Node) string {
	return a.nodeIDs[node]
}

func (a *Graph) NodeIds() []string {
	return slices.Collect(maps.Keys(a.nodesByID))
}

func (a *Graph) HasNodeWithId(nodeId string) bool {
	_, ok := a.nodesByID[nodeId]
	return ok
}

func (a *Graph) Node(nodeId string) nodes.Node {
	node, ok := a.nodesByID[nodeId]
	if !ok {
		panic(fmt.Errorf("no node exists with id %q", nodeId))
	}
	return node
}

func (a *Graph) CreateNode(nodeType string) (node nodes.Node, id string, err error) {
	_, err = a.project.write(refuseConflicts, func() (err error) {
		node, id, err = a.createNode(nodeType, "")
		return err
	})
	return node, id, err
}

func (a *Graph) CreateBoundaryNode(nodeType, portType string) (node nodes.Node, id string, err error) {
	if strings.TrimSpace(portType) == "" {
		return nil, "", fmt.Errorf("boundary port type is required")
	}
	_, err = a.project.write(refuseConflicts, func() (err error) {
		node, id, err = a.createNode(nodeType, portType)
		return err
	})
	return node, id, err
}

func (a *Graph) createNode(nodeType, portType string) (nodes.Node, string, error) {
	factory := a.project.typeFactory
	if !factory.KeyRegistered(nodeType) {
		return nil, "", fmt.Errorf("no factory registered with ID %s", nodeType)
	}
	if placed := subgraph.RuntimeTypeID(nodeType); placed != "" {
		if a.scope != "" && a.project.wouldContainItself(a.scope, placed) {
			return nil, "", refuseSelfContainment(a.scope, placed)
		}
	}

	node, ok := factory.New(nodeType).(nodes.Node)
	if !ok {
		panic(fmt.Errorf("registered type %s did not create a node", nodeType))
	}

	if boundary, isBoundary := subgraph.IsBoundaryNode(node); isBoundary {
		if portType == "" {
			return nil, "", fmt.Errorf("boundary port type is required")
		}
		if !nodes.IsPortTypeKnown(portType) {
			return nil, "", fmt.Errorf("unknown boundary port type %q", portType)
		}
		if err := subgraph.ConfigureBoundaryPortType(boundary, portType); err != nil {
			return nil, "", err
		}
	} else if portType != "" {
		return nil, "", fmt.Errorf("port type cannot be set on non-boundary node type %q", nodeType)
	}

	a.registerWithInputs(node, nodeType)
	return node, a.nodeIDs[node], nil
}

// DeleteNodeById returns the edges elsewhere that no longer fit without it.
func (a *Graph) DeleteNodeById(nodeId string) ([]DroppedEdge, error) {
	return a.project.write(dropConflicts, func() error {
		return a.deleteNodeByID(nodeId)
	})
}

func (a *Graph) deleteNodeByID(nodeId string) error {
	node, ok := a.nodesByID[nodeId]
	if !ok {
		return fmt.Errorf("can't delete, no node registered with ID %s", nodeId)
	}
	return a.deleteNode(node)
}

func (a *Graph) deleteNode(nodeToDelete nodes.Node) error {
	if _, ok := a.nodeIDs[nodeToDelete]; !ok {
		return fmt.Errorf("can't delete a node that is not in the graph")
	}

	a.project.namedManifests.DeleteNode(nodeToDelete)
	a.dirty()
	a.forget(nodeToDelete)

	for consumer, inputs := range a.reads {
		for input, sources := range inputs {
			kept := slices.DeleteFunc(slices.Clone(sources), func(from source) bool { return from.node == nodeToDelete })
			if len(kept) != len(sources) {
				a.setSources(consumer, input, kept)
			}
		}
	}
	return nil
}

func (a *Graph) Parameter(nodeId string) (Parameter, error) {
	node, ok := a.nodesByID[nodeId]
	if !ok {
		return nil, fmt.Errorf("no node exists with id %q", nodeId)
	}
	param, ok := node.(Parameter)
	if !ok {
		return nil, fmt.Errorf("node %q is not a parameter", nodeId)
	}
	return param, nil
}

// A parameter inside a definition is the same node in every copy, so the
// change reaches every placement without an edit to the graph.
func (a *Graph) UpdateParameter(nodeId string, data []byte) (changed bool, err error) {
	_, err = a.project.write(refuseConflicts, func() (err error) {
		param, err := a.Parameter(nodeId)
		if err != nil {
			return err
		}
		changed, err = param.ApplyMessage(data)
		return err
	})
	return changed, err
}

func (a *Graph) ParameterData(nodeId string) ([]byte, error) {
	a.mu().Lock()
	defer a.mu().Unlock()

	param, err := a.Parameter(nodeId)
	if err != nil {
		return nil, err
	}
	return param.ToMessage(), nil
}
