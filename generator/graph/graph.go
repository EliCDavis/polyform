package graph

import (
	gsync "sync"

	"github.com/EliCDavis/polyform/generator/sync"
	"github.com/EliCDavis/polyform/nodes"
)

// A Graph is nodes and the edges between them: the root graph, a subgraph
// definition, or the copy of a definition one of its placements holds.
type Graph struct {
	project *Instance

	// The definition this is; empty for the root and for copies.
	scope string

	// Written only through register and forget, which keep them in step.
	nodeIDs         map[nodes.Node]string
	nodesByID       map[string]nodes.Node
	nodeTypeKeys    map[nodes.Node]string
	nodeIDHighWater int // so a deleted id is never handed out again

	// Every edge: what each input of each node reads, an array input's
	// sources in element order. A placement's inputs are keyed by boundary
	// node id. Node fields are a projection of this, written only by settle.
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
	a.reads = make(map[nodes.Node]map[string][]source)
	a.nodeIDHighWater = 0
	a.metadata = sync.NewNestedSyncMap()
}

// Root is the project this graph belongs to.
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

// SubGraphScopeID is the id of the definition this graph is, or empty for
// the root graph and for a placement's copy.
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
