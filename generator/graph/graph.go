package graph

import (
	gsync "sync"

	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/generator/sync"
	"github.com/EliCDavis/polyform/nodes"
)

type Graph struct {
	project *Instance

	// The definition this is; empty for the root and for copies.
	scope string

	// Written only through register and forget, which keep them in step.
	nodeIDs         map[nodes.Node]string
	nodesByID       map[string]nodes.Node
	nodeTypeKeys    map[nodes.Node]string
	boundaryNodes   map[string]subgraph.Boundary
	nodeIDHighWater int // so a deleted id is never handed out again

	// Every edge, an array input's sources in element order; a placement's
	// inputs keyed by boundary node id. Node fields follow this, via settle only.
	reads map[nodes.Node]map[string][]source

	metadata *sync.NestedSyncMap
}

func newGraph(project *Instance, scope string) *Graph {
	graph := &Graph{project: project, scope: scope}
	graph.clear()
	return graph
}

func (a *Graph) clear() {
	a.nodeIDs = make(map[nodes.Node]string)
	a.nodesByID = make(map[string]nodes.Node)
	a.nodeTypeKeys = make(map[nodes.Node]string)
	a.boundaryNodes = make(map[string]subgraph.Boundary)
	a.reads = make(map[nodes.Node]map[string][]source)
	a.nodeIDHighWater = 0
	a.metadata = sync.NewNestedSyncMap()
}

func (a *Graph) Root() *Instance {
	return a.project
}

func (a *Graph) IsRoot() bool {
	return a == a.project.Graph
}

// One lock for the whole project: an edit inside a subgraph reaches every graph
// that places it. Exported methods take it; unexported ones expect it held.
func (a *Graph) mu() *gsync.RWMutex {
	return &a.project.lock
}

func (a *Graph) SubGraphScopeID() string {
	return a.scope
}

func (a *Graph) SetMetadata(key string, value any) {
	a.metadata.Set(key, value)
}

func (a *Graph) Metadata(key string) any {
	if !a.metadata.PathExists(key) {
		return nil
	}
	return a.metadata.Get(key)
}

func (a *Graph) DeleteMetadata(key string) {
	a.metadata.Delete(key)
}
