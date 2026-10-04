package graph

import (
	gsync "sync"
	"sync/atomic"

	"github.com/EliCDavis/polyform/generator/sync"
	"github.com/EliCDavis/polyform/nodes"
)

// A Graph is nodes and the edges between them: the root graph, a subgraph
// definition, or the copy of a definition one of its placements holds.
type Graph struct {
	project *Instance

	// Written only through register and forget, which keep them in step.
	nodeIDs         map[nodes.Node]string
	nodesByID       map[string]nodes.Node
	nodeTypeKeys    map[nodes.Node]string
	nodeIDHighWater int // so a deleted id is never handed out again

	// Every edge: what each input of each node reads, an array input's
	// sources in element order.
	reads map[nodes.Node]map[string][]source

	metadata *sync.NestedSyncMap

	boundaryCache atomic.Pointer[boundaryIndex]
}

func newGraph(project *Instance) *Graph {
	graph := &Graph{project: project}
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
	a.clearBoundaryCache()
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

func (a *Graph) incModelVersion() {
	a.project.modelVersion++
}

// SubGraphScopeID is the id of the definition this graph is, or empty for
// the root graph and for a placement's copy.
func (a *Graph) SubGraphScopeID() string {
	for id, definition := range a.project.subGraphs {
		if definition.instance == a {
			return id
		}
	}
	return ""
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
