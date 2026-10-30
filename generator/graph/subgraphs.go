package graph

import (
	"fmt"
	"strings"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/persistence"
	"github.com/EliCDavis/polyform/generator/subgraph"
)

type subGraphRuntime struct {
	name        string
	description string
	group       string
	instance    *Graph

	// Bumped by every edit to the definition and to anything it places;
	// a copy built from an older version is stale.
	version uint64
}

type BoundaryPortKind string

const (
	BoundaryPortKindInput  BoundaryPortKind = "input"
	BoundaryPortKindOutput BoundaryPortKind = "output"
)

type SubgraphBoundaryPort struct {
	// The boundary node's id, which is the port's key on a placement.
	ID   string
	Name string
	Type string
	Kind BoundaryPortKind
}

func GetBoundaryKind(boundary subgraph.Boundary) BoundaryPortKind {
	if _, isInput := subgraph.IsInputBoundary(boundary); isInput {
		return BoundaryPortKindInput
	}
	return BoundaryPortKindOutput
}

func (a *Instance) CreateSubGraph(id, name, description string) error {
	_, err := a.write(refuseConflicts, func() error {
		return a.createSubGraph(id, name, description)
	})
	return err
}

func (a *Instance) createSubGraph(id, name, description string) error {
	if _, exists := a.subGraphs[id]; exists {
		return fmt.Errorf("sub-graph %q already exists", id)
	}

	a.subGraphs[id] = &subGraphRuntime{
		name:        name,
		description: description,
		instance:    newGraph(a, id),
	}
	a.RegisterSubGraphNodeType(id)
	return nil
}

func (a *Instance) DeleteSubGraph(id string) error {
	_, err := a.write(refuseConflicts, func() error {
		return a.deleteSubGraph(id)
	})
	return err
}

func (a *Instance) deleteSubGraph(id string) error {
	if _, exists := a.subGraphs[id]; !exists {
		return fmt.Errorf("sub-graph %q does not exist", id)
	}

	placed := 0
	for node := range a.nodeIDs {
		if placement, ok := node.(*SubgraphInstanceNode); ok && placement.subGraphID == id {
			placed++
		}
	}
	if placed > 0 {
		return fmt.Errorf("sub-graph %q is still referenced by %d node instance(s)", id, placed)
	}

	a.typeFactory.Unregister(subgraph.RuntimeTypePath(id))
	delete(a.subGraphs, id)
	return nil
}

func (a *Instance) RenameSubGraph(oldID, newID string) error {
	_, err := a.write(refuseConflicts, func() error {
		newID = strings.TrimSpace(newID)
		runtime, exists := a.subGraphs[oldID]
		if !exists {
			return fmt.Errorf("sub-graph %q does not exist", oldID)
		}
		if newID == "" {
			return fmt.Errorf("new sub-graph id is required")
		}
		if newID == oldID {
			return nil
		}
		if _, taken := a.subGraphs[newID]; taken {
			return fmt.Errorf("sub-graph %q already exists", newID)
		}

		delete(a.subGraphs, oldID)
		a.subGraphs[newID] = runtime
		runtime.instance.scope = newID
		a.typeFactory.Unregister(subgraph.RuntimeTypePath(oldID))
		a.RegisterSubGraphNodeType(newID)

		typePath := subgraph.RuntimeTypePath(newID)
		for _, scope := range a.scopes() {
			for node := range scope.graph.nodeIDs {
				if placement, ok := node.(*SubgraphInstanceNode); ok && placement.subGraphID == oldID {
					placement.subGraphID = newID
					scope.graph.nodeTypeKeys[placement] = typePath
				}
			}
		}
		// Copies hold placements by the old id.
		for _, definition := range a.subGraphs {
			definition.version++
		}
		return nil
	})
	return err
}

func (a *Instance) SetSubGraphInfo(id, name, description, group string) error {
	_, err := a.write(refuseConflicts, func() error {
		runtime, exists := a.subGraphs[id]
		if !exists {
			return fmt.Errorf("sub-graph %q does not exist", id)
		}
		runtime.name = name
		runtime.description = description
		runtime.group = group
		return nil
	})
	return err
}

func (a *Instance) SubGraphInstance(id string) (*Graph, error) {
	runtime, exists := a.subGraphs[id]
	if !exists {
		return nil, fmt.Errorf("sub-graph %q does not exist", id)
	}
	return runtime.instance, nil
}

func (a *Instance) wouldContainItself(scope, placed string) bool {
	seen := map[string]bool{}
	var reaches func(id string) bool
	reaches = func(id string) bool {
		if id == scope {
			return true
		}
		runtime, exists := a.subGraphs[id]
		if seen[id] || !exists {
			return false
		}
		seen[id] = true
		for node := range runtime.instance.nodeIDs {
			if placement, ok := node.(*SubgraphInstanceNode); ok && reaches(placement.subGraphID) {
				return true
			}
		}
		return false
	}
	return reaches(placed)
}

func refuseSelfContainment(scope, placed string) error {
	return fmt.Errorf("sub-graph %q cannot be placed inside %q: it would contain itself", placed, scope)
}

func (a *Instance) RegisterSubGraphNodeType(subGraphID string) (string, error) {
	typePath := subgraph.RuntimeTypePath(subGraphID)
	a.typeFactory.RegisterBuilder(typePath, func() any {
		return NewRuntimeNode(a, subGraphID)
	})
	return typePath, nil
}

func (a *Instance) CollectBoundaryPorts(subGraphID string) ([]SubgraphBoundaryPort, error) {
	child, err := a.SubGraphInstance(subGraphID)
	if err != nil {
		return nil, err
	}

	ports := make([]SubgraphBoundaryPort, 0, len(child.boundaryNodes))
	for id, boundary := range child.boundaryNodes {
		if isPort(boundary) {
			ports = append(ports, SubgraphBoundaryPort{ID: id, Name: boundary.BoundaryPortName(), Type: boundary.BoundaryPortType(), Kind: GetBoundaryKind(boundary)})
		}
	}
	return ports, nil
}

func (a *Instance) persistedSubGraphDefinition(id string, encoder *jbtf.Encoder) (persistence.SubGraph, error) {
	runtime, exists := a.subGraphs[id]
	if !exists {
		return persistence.SubGraph{}, fmt.Errorf("sub-graph %q does not exist", id)
	}

	child := runtime.instance
	notes, _ := child.metadata.Get("notes").(map[string]any)

	return persistence.SubGraph{
		Name:        runtime.name,
		Description: runtime.description,
		Group:       runtime.group,
		Nodes:       child.savedNodes(encoder),
		Notes:       notes,
		Metadata:    child.metadata.Data(),
	}, nil
}

func (a *Instance) encodeSubGraphDefinitions(encoder *jbtf.Encoder) (map[string]persistence.SubGraph, error) {
	result := make(map[string]persistence.SubGraph, len(a.subGraphs))
	for id := range a.subGraphs {
		def, err := a.persistedSubGraphDefinition(id, encoder)
		if err != nil {
			return nil, err
		}
		result[id] = def
	}
	return result, nil
}
