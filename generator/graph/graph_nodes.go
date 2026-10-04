package graph

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/EliCDavis/polyform/generator/persistence"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
)

func (a *Graph) register(node nodes.Node, id, typeKey string) {
	a.clearBoundaryCache()
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
	a.clearBoundaryCache()
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
	for name, input := range node.Inputs() {
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

func (a *Graph) CreateNode(nodeType string) (nodes.Node, string, error) {
	a.mu().Lock()
	defer a.mu().Unlock()
	return a.createNode(nodeType, "")
}

func (a *Graph) CreateBoundaryNode(nodeType, portType string) (nodes.Node, string, error) {
	if strings.TrimSpace(portType) == "" {
		return nil, "", fmt.Errorf("boundary port type is required")
	}
	a.mu().Lock()
	defer a.mu().Unlock()
	return a.createNode(nodeType, portType)
}

func (a *Graph) createNode(nodeType, portType string) (nodes.Node, string, error) {
	factory := a.project.typeFactory
	if !factory.KeyRegistered(nodeType) {
		return nil, "", fmt.Errorf("no factory registered with ID %s", nodeType)
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
	id := a.nodeIDs[node]

	// A boundary is not a port until it is named, so nothing is reshaped yet.
	err := a.updateCopies(func(copied *Graph) error {
		created, err := copied.instantiateAppNode(id, persistence.Node{Type: nodeType})
		if boundary, isBoundary := subgraph.IsBoundaryNode(created); isBoundary && err == nil {
			err = subgraph.ConfigureBoundaryPortType(boundary, portType)
		}
		return err
	})
	return node, id, err
}

// DeleteNodeById returns the edges elsewhere that no longer fit without it.
func (a *Graph) DeleteNodeById(nodeId string) ([]DroppedEdge, error) {
	a.mu().Lock()
	defer a.mu().Unlock()
	return a.deleteNodeByID(nodeId)
}

func (a *Graph) deleteNodeByID(nodeId string) ([]DroppedEdge, error) {
	node, ok := a.nodesByID[nodeId]
	if !ok {
		return nil, fmt.Errorf("can't delete, no node registered with ID %s", nodeId)
	}
	return a.deleteNode(node)
}

func (a *Graph) deleteNode(nodeToDelete nodes.Node) ([]DroppedEdge, error) {
	if _, ok := a.nodeIDs[nodeToDelete]; !ok {
		return nil, fmt.Errorf("can't delete a node that is not in the graph")
	}

	a.project.namedManifests.DeleteNode(nodeToDelete)
	a.forget(nodeToDelete)

	for consumer, inputs := range a.reads {
		for input, sources := range inputs {
			kept := slices.DeleteFunc(slices.Clone(sources), func(from source) bool { return from.node == nodeToDelete })
			if len(kept) != len(sources) {
				a.setSources(consumer, input, kept)
			}
		}
	}

	dropped, err := a.commitEdit(dropConflicts)
	if err != nil {
		return nil, err
	}
	a.incModelVersion()
	return dropped, nil
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

func (a *Graph) UpdateParameter(nodeId string, data []byte) (bool, error) {
	a.mu().Lock()
	defer a.mu().Unlock()

	param, err := a.Parameter(nodeId)
	if err != nil {
		return false, err
	}
	changed, err := param.ApplyMessage(data)
	if err != nil {
		return false, err
	}
	a.incModelVersion()
	return changed, a.updateCopies(func(copied *Graph) error {
		param, err := copied.Parameter(nodeId)
		if err != nil {
			return err
		}
		_, err = param.ApplyMessage(data)
		return err
	})
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
