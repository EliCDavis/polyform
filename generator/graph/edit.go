package graph

import (
	"cmp"
	"maps"
	"slices"

	"github.com/EliCDavis/polyform/nodes"
)

type DroppedEdge struct {
	// The subgraph the edge was in; empty for the root graph.
	Scope string

	From     string
	FromPort string
	To       string
	// "Port", or "Port.N" for an element of an array input.
	ToPort string

	Reason string
}

type editState struct {
	depth int
	// Settled on close, with everything placing them.
	dirty map[*Graph]bool
	// Copies go stale; nothing to settle.
	changed map[*Graph]bool
	before  map[*Graph]map[nodes.Node]map[string][]source
}

// dirty has the edit settle this graph, and everything placing it, on close.
func (a *Graph) dirty() {
	edit := &a.project.edit
	if edit.dirty == nil {
		edit.dirty = map[*Graph]bool{}
		edit.before = map[*Graph]map[nodes.Node]map[string][]source{}
	}
	if !edit.dirty[a] {
		edit.dirty[a] = true
		edit.before[a] = cloneReads(a.reads)
	}
}

func (a *Graph) changed() {
	edit := &a.project.edit
	if edit.changed == nil {
		edit.changed = map[*Graph]bool{}
	}
	edit.changed[a] = true
}

func cloneReads(reads map[nodes.Node]map[string][]source) map[nodes.Node]map[string][]source {
	cloned := make(map[nodes.Node]map[string][]source, len(reads))
	for node, inputs := range reads {
		cloned[node] = maps.Clone(inputs)
	}
	return cloned
}

// Inside a Batch it only runs f; the batch closes the edit.
func (a *Instance) write(policy conflictPolicy, f func() error) ([]DroppedEdge, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.edit.depth++
	err := f()
	a.edit.depth--
	if err != nil || a.edit.depth > 0 {
		return nil, err
	}
	return a.close(policy)
}

// Batch settles once at the end; the edges that no longer fit are dropped and returned.
func (a *Instance) Batch(f func() error) ([]DroppedEdge, error) {
	a.lock.Lock()
	a.edit.depth++
	a.lock.Unlock()

	err := f()

	a.lock.Lock()
	defer a.lock.Unlock()
	a.edit.depth--
	if a.edit.depth > 0 {
		return nil, err
	}
	dropped, closeErr := a.close(dropConflicts)
	return dropped, cmp.Or(err, closeErr)
}

func (a *Instance) close(policy conflictPolicy) ([]DroppedEdge, error) {
	dirty, changed, before := a.edit.dirty, a.edit.changed, a.edit.before
	a.edit = editState{}
	a.modelVersion++

	stale := map[*Graph]bool{}
	maps.Copy(stale, dirty)
	maps.Copy(stale, changed)
	for _, graph := range a.affectedOrder(stale) {
		if definition, ok := a.subGraphs[graph.scope]; ok {
			definition.version++
		}
	}

	order := a.affectedOrder(dirty)
	var dropped []DroppedEdge
	for _, graph := range order {
		lost, err := graph.settle(policy)
		dropped = append(dropped, lost...)
		if err != nil {
			a.restore(dirty, before, order)
			return nil, err
		}
	}
	return dropped, nil
}

// Copies built while the refused edit was in place are stale again.
func (a *Instance) restore(dirty map[*Graph]bool, before map[*Graph]map[nodes.Node]map[string][]source, order []*Graph) {
	for graph := range dirty {
		graph.reads = before[graph]
	}
	for _, graph := range order {
		if definition, ok := a.subGraphs[graph.scope]; ok {
			definition.version++
		}
	}
	for _, graph := range order {
		graph.settle(refuseConflicts)
	}
}

// Innermost first: a placement's types are only right once what it places is settled.
func (a *Instance) affectedOrder(dirty map[*Graph]bool) []*Graph {
	if len(dirty) == 0 {
		return nil
	}
	graphs := map[string]*Graph{"": a.Graph}
	for id, definition := range a.subGraphs {
		graphs[id] = definition.instance
	}

	placed := func(graph *Graph) []string {
		var ids []string
		for node := range graph.nodeIDs {
			if placement, ok := node.(*SubgraphInstanceNode); ok {
				ids = append(ids, placement.subGraphID)
			}
		}
		return ids
	}

	order, _ := dependenciesFirst(graphs, placed)
	affected := map[string]bool{}
	var ordered []*Graph
	for _, scope := range order {
		graph := graphs[scope]
		if dirty[graph] || slices.ContainsFunc(placed(graph), func(id string) bool { return affected[id] }) {
			affected[scope] = true
			ordered = append(ordered, graph)
		}
	}
	return ordered
}
