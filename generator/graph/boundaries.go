package graph

import (
	"fmt"
	"strings"

	"github.com/EliCDavis/polyform/generator/subgraph"
)

// The ports a graph offers wherever it is placed, by name.
type boundaryIndex struct {
	inputs  map[string]subgraph.InputBoundary
	outputs map[string]*subgraph.OutputNode
}

// A boundary is a port once it has both a type and a name. The index is
// kept until boundariesChanged, and must not be modified.
func (a *Graph) boundaries() boundaryIndex {
	if cached := a.boundaryCache.Load(); cached != nil {
		return *cached
	}

	index := boundaryIndex{
		inputs:  make(map[string]subgraph.InputBoundary),
		outputs: make(map[string]*subgraph.OutputNode),
	}
	defer a.boundaryCache.Store(&index)

	for node := range a.nodeIDs {
		boundary, ok := subgraph.IsBoundaryNode(node)
		if !ok || boundary.BoundaryPortType() == "" || !subgraph.BoundaryPortNameConfigured(boundary) {
			continue
		}
		switch boundary := boundary.(type) {
		case *subgraph.OutputNode:
			index.outputs[boundary.BoundaryPortName()] = boundary
		case subgraph.InputBoundary:
			index.inputs[boundary.BoundaryPortName()] = boundary
		}
	}
	return index
}

// boundariesChanged has to follow anything that adds, removes, names or
// types a boundary node.
func (a *Graph) boundariesChanged() {
	a.boundaryCache.Store(nil)
}

func (a *Graph) SetBoundaryNodeInfo(nodeID, portName string) error {
	a.mu().Lock()
	defer a.mu().Unlock()
	return a.setBoundaryNodeInfo(nodeID, portName)
}

func (a *Graph) setBoundaryNodeInfo(nodeID, portName string) error {
	if strings.TrimSpace(portName) == "" {
		return fmt.Errorf("boundary port name is required")
	}

	node := a.Node(nodeID)
	boundary, ok := subgraph.IsBoundaryNode(node)
	if !ok {
		return fmt.Errorf("node %q is not a sub-graph boundary node", nodeID)
	}
	kind := GetBoundaryKind(boundary)

	for other := range a.nodeIDs {
		taken, ok := subgraph.IsBoundaryNode(other)
		if ok && other != node && GetBoundaryKind(taken) == kind && taken.BoundaryPortName() == portName {
			return fmt.Errorf("boundary port name %q already used by another %s node", portName, kind)
		}
	}

	var oldName string
	switch boundary := boundary.(type) {
	case *subgraph.InputNode:
		oldName, boundary.PortName = boundary.PortName, portName
	case *subgraph.OutputNode:
		oldName, boundary.PortName = boundary.PortName, portName
	default:
		return fmt.Errorf("node %q is not a sub-graph boundary node", nodeID)
	}
	a.boundariesChanged()

	// Renamed on every placement first, so what feeds the old name is kept
	// rather than dropped as feeding a port that is gone.
	if scope := a.SubGraphScopeID(); scope != "" && oldName != "" && oldName != portName {
		forEachSubGraphInstance(a.Root(), scope, func(holder *Graph, placement *SubgraphInstanceNode) {
			placement.renameBoundaryPort(oldName, portName, kind)
			holder.renamePort(placement, kind, oldName, portName)
		})
	}
	_, err := a.commitEdit(refuseConflicts)
	return err
}
