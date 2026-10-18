package graph

import (
	"fmt"
	"strings"

	"github.com/EliCDavis/polyform/generator/subgraph"
)

// The ports a graph offers wherever it is placed, by boundary node id.
type boundaryIndex struct {
	inputs  map[string]subgraph.InputBoundary
	outputs map[string]*subgraph.OutputNode
}

// A boundary is a port once it has both a type and a name.
func (a *Graph) boundaries() boundaryIndex {
	index := boundaryIndex{
		inputs:  make(map[string]subgraph.InputBoundary),
		outputs: make(map[string]*subgraph.OutputNode),
	}
	for node, id := range a.nodeIDs {
		boundary, ok := subgraph.IsBoundaryNode(node)
		if !ok || boundary.BoundaryPortType() == "" || !subgraph.BoundaryPortNameConfigured(boundary) {
			continue
		}
		switch boundary := boundary.(type) {
		case *subgraph.OutputNode:
			index.outputs[id] = boundary
		case subgraph.InputBoundary:
			index.inputs[id] = boundary
		}
	}
	return index
}

func (a *Graph) SetBoundaryNodeInfo(nodeID, portName string) error {
	return a.RenameBoundaryPorts(map[string]string{nodeID: portName})
}

// Names may trade places: each is checked against the state after all of them.
func (a *Graph) RenameBoundaryPorts(names map[string]string) error {
	_, err := a.project.write(refuseConflicts, func() error {
		return a.renameBoundaryPorts(names)
	})
	return err
}

func (a *Graph) renameBoundaryPorts(names map[string]string) error {
	final := map[BoundaryPortKind]map[string]string{}
	for node, id := range a.nodeIDs {
		boundary, ok := subgraph.IsBoundaryNode(node)
		if !ok {
			continue
		}
		name, renamed := names[id]
		name = strings.TrimSpace(name)
		switch {
		case renamed && name == "":
			return fmt.Errorf("boundary port name is required for %q", id)
		case !renamed && !subgraph.BoundaryPortNameConfigured(boundary):
			continue
		case !renamed:
			name = boundary.BoundaryPortName()
		}
		kind := GetBoundaryKind(boundary)
		if final[kind] == nil {
			final[kind] = map[string]string{}
		}
		if other, taken := final[kind][name]; taken {
			return fmt.Errorf("boundary port name %q would be used by both %s nodes %q and %q", name, kind, other, id)
		}
		final[kind][name] = id
	}

	for id, name := range names {
		node, exists := a.nodesByID[id]
		if !exists {
			return fmt.Errorf("no node exists with id %q", id)
		}
		switch boundary := node.(type) {
		case *subgraph.InputNode:
			boundary.PortName = strings.TrimSpace(name)
		case *subgraph.OutputNode:
			boundary.PortName = strings.TrimSpace(name)
		default:
			return fmt.Errorf("node %q is not a sub-graph boundary node", id)
		}
	}
	a.dirty()
	return nil
}
