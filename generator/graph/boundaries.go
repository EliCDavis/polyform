package graph

import (
	"fmt"
	"strings"

	"github.com/EliCDavis/polyform/generator/subgraph"
)

// A boundary is a port once it has both a type and a name.
func isPort(boundary subgraph.Boundary) bool {
	return boundary.PortType() != "" && subgraph.BoundaryPortNameConfigured(boundary)
}

// inputPort is the input boundary with this id, if it is a port.
func (a *Graph) inputPort(id string) (subgraph.InputBoundary, bool) {
	boundary, ok := a.boundaryNodes[id].(subgraph.InputBoundary)
	return boundary, ok && isPort(boundary)
}

func (a *Graph) outputPort(id string) (*subgraph.OutputNode, bool) {
	boundary, ok := a.boundaryNodes[id].(*subgraph.OutputNode)
	return boundary, ok && isPort(boundary)
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

func (a *Graph) SetBoundaryPortDescriptions(descriptions map[string]string) error {
	_, err := a.project.write(refuseConflicts, func() error {
		for id := range descriptions {
			if _, ok := subgraph.IsBoundaryNode(a.boundaryNodes[id]); !ok {
				return fmt.Errorf("node %q is not a sub-graph boundary node", id)
			}
		}
		for id, description := range descriptions {
			a.boundaryNodes[id].SetPortDescription(strings.TrimSpace(description))
		}
		a.dirty()
		return nil
	})
	return err
}

func (a *Graph) renameBoundaryPorts(names map[string]string) error {
	for id := range names {
		if _, exists := a.nodesByID[id]; !exists {
			return fmt.Errorf("no node exists with id %q", id)
		}
		if _, ok := a.boundaryNodes[id]; !ok {
			return fmt.Errorf("node %q is not a sub-graph boundary node", id)
		}
	}

	final := map[BoundaryPortKind]map[string]string{}
	for id, boundary := range a.boundaryNodes {
		name, renamed := names[id]
		name = strings.TrimSpace(name)
		switch {
		case renamed && name == "":
			return fmt.Errorf("boundary port name is required for %q", id)
		case !renamed && !subgraph.BoundaryPortNameConfigured(boundary):
			continue
		case !renamed:
			name = boundary.PortName()
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
		a.boundaryNodes[id].SetPortName(strings.TrimSpace(name))
	}
	a.dirty()
	return nil
}
