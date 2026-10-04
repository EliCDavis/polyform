package graph

import (
	"fmt"
	"maps"
	"slices"
)

// A DroppedEdge was removed because an edit left its producer carrying a
// type its input cannot take, or took the port it was wired to.
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

// commitEdit settles a after an edit to its nodes or edges, then brings
// every graph placing it back in line. It returns the edges that had to go.
func (a *Graph) commitEdit(policy conflictPolicy) ([]DroppedEdge, error) {
	a.clearCopySource()
	if a.deferredToCompoundEdit() {
		return nil, nil
	}

	dropped, err := a.settle(policy)
	if err != nil {
		return dropped, err
	}
	elsewhere, err := a.refreshPlacers(policy)
	return append(dropped, elsewhere...), err
}

// updateCopies repeats an edit in every copy of this definition. Only for an
// edit that cannot change a type: anything else has to go through commitEdit.
func (a *Graph) updateCopies(edit func(copied *Graph) error) error {
	a.clearCopySource()
	scope := a.SubGraphScopeID()
	if scope == "" || a.deferredToCompoundEdit() {
		return nil
	}

	root := a.Root()
	root.incModelVersion()

	var err error
	forEachSubGraphInstance(root, scope, func(_ *Graph, placement *SubgraphInstanceNode) {
		if copied := placement.BuiltGraph(); copied != nil && err == nil {
			err = edit(copied)
		}
	})
	return err
}

// refreshPlacers brings every graph that places this definition, directly
// or through another definition, back in line with it.
func (a *Graph) refreshPlacers(policy conflictPolicy) ([]DroppedEdge, error) {
	changed := a.SubGraphScopeID()
	if changed == "" {
		return nil, nil
	}

	root := a.Root()
	root.incModelVersion()

	var dropped []DroppedEdge
	placers, affected := root.placersOf(changed)
	for _, placer := range placers {
		var lost []DroppedEdge
		err := placer.refreshPlacements(affected)
		if err == nil {
			lost, err = placer.settle(policy)
		}
		if len(lost) > 0 {
			placer.clearCopySource()
		}
		dropped = append(dropped, lost...)
		if err != nil {
			return dropped, err
		}
	}
	return dropped, nil
}

// Innermost first: a graph's placements may only be rebuilt once every
// definition they place is final.
func (root *Instance) placersOf(changed string) (innermostFirst []*Graph, affected map[string]bool) {
	graphs := map[string]*Graph{"": root.Graph}
	for id, definition := range root.subGraphs {
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
	affected = map[string]bool{changed: true}
	for _, scope := range order {
		placesAffected := slices.ContainsFunc(placed(graphs[scope]), func(id string) bool { return affected[id] })
		if scope != changed && placesAffected {
			affected[scope] = true
			innermostFirst = append(innermostFirst, graphs[scope])
		}
	}
	return innermostFirst, affected
}

// refreshPlacements drops the copy held by each of a's placements of an
// affected definition.
func (a *Graph) refreshPlacements(affected map[string]bool) error {
	for _, id := range slices.Sorted(maps.Keys(a.nodesByID)) {
		placement, ok := a.nodesByID[id].(*SubgraphInstanceNode)
		if !ok || !affected[placement.subGraphID] {
			continue
		}
		placement.invalidate()
		if err := placement.buildIfNeeded(); err != nil {
			return fmt.Errorf("rebuilding placement %s of subgraph %q: %w", id, placement.subGraphID, err)
		}
	}
	return nil
}
