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
}

type BoundaryPortKind string

const (
	BoundaryPortKindInput  BoundaryPortKind = "input"
	BoundaryPortKindOutput BoundaryPortKind = "output"
)

type SubgraphBoundaryPort struct {
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
	a.mu().Lock()
	defer a.mu().Unlock()
	return a.createSubGraph(id, name, description)
}

func (a *Instance) createSubGraph(id, name, description string) error {
	if _, exists := a.subGraphs[id]; exists {
		return fmt.Errorf("sub-graph %q already exists", id)
	}

	a.subGraphs[id] = &subGraphRuntime{
		name:        name,
		description: description,
		instance:    newGraph(a),
	}
	a.RegisterSubGraphNodeType(id)
	a.incModelVersion()
	return nil
}

func (a *Instance) DeleteSubGraph(id string) error {
	a.mu().Lock()
	defer a.mu().Unlock()
	return a.deleteSubGraph(id)
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
	delete(a.copySources, id)
	a.incModelVersion()
	return nil
}

func (a *Instance) RenameSubGraph(oldID, newID string) error {
	a.mu().Lock()
	defer a.mu().Unlock()

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
	a.typeFactory.Unregister(subgraph.RuntimeTypePath(oldID))
	a.RegisterSubGraphNodeType(newID)

	typePath := subgraph.RuntimeTypePath(newID)
	forEachSubGraphInstance(a, oldID, func(holder *Graph, placement *SubgraphInstanceNode) {
		placement.subGraphID = newID
		holder.nodeTypeKeys[placement] = typePath
	})
	// Copy sources name placements by the old type path.
	a.copySources = make(map[string]savedGraph)
	a.incModelVersion()
	return nil
}

func (a *Instance) SetSubGraphInfo(id, name, description string) error {
	a.mu().Lock()
	defer a.mu().Unlock()

	runtime, exists := a.subGraphs[id]
	if !exists {
		return fmt.Errorf("sub-graph %q does not exist", id)
	}
	runtime.name = name
	runtime.description = description
	a.incModelVersion()
	return nil
}

// SubGraphInstance is the definition every placement of the subgraph copies.
func (a *Instance) SubGraphInstance(id string) (*Graph, error) {
	runtime, exists := a.subGraphs[id]
	if !exists {
		return nil, fmt.Errorf("sub-graph %q does not exist", id)
	}
	return runtime.instance, nil
}

// wouldContainItself reports whether placing placed inside the definition
// scope lets scope reach itself through placements.
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

	index := child.boundaries()
	ports := make([]SubgraphBoundaryPort, 0, len(index.inputs)+len(index.outputs))
	for name, boundary := range index.inputs {
		ports = append(ports, SubgraphBoundaryPort{Name: name, Type: boundary.BoundaryPortType(), Kind: BoundaryPortKindInput})
	}
	for name, boundary := range index.outputs {
		ports = append(ports, SubgraphBoundaryPort{Name: name, Type: boundary.BoundaryPortType(), Kind: BoundaryPortKindOutput})
	}
	return ports, nil
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
