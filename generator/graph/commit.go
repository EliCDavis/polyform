package graph

import (
	"fmt"
	"slices"
	"sort"
)

type conflictPolicy int

const (
	// An edit that adds refuses instead of costing an existing edge.
	refuseConflicts conflictPolicy = iota
	// An edit that removes drops the edges that depended on what it took,
	// and reports them.
	dropConflicts
)

func (a *Graph) settleUnder(policy conflictPolicy) ([]DroppedEdge, error) {
	if policy == refuseConflicts {
		return nil, a.settle()
	}
	return a.settleDroppingConflicts()
}

func (a *Graph) commitEdit(policy conflictPolicy) error {
	a.definitionChanged()
	if a.deferredToCompoundEdit() {
		return nil
	}
	if _, err := a.settleUnder(policy); err != nil {
		return err
	}
	return a.refreshPlacers(policy)
}

// updateCopies repeats an edit in every copy of this definition. Only for an
// edit that cannot change a type: anything else has to go through refreshPlacers.
func (a *Graph) updateCopies(edit func(copied *Graph) error) error {
	a.definitionChanged()
	scope := a.SubGraphScopeID()
	if scope == "" || a.deferredToCompoundEdit() {
		return nil
	}

	root := a.Root()
	root.incModelVersion()

	var err error
	forEachSubGraphInstance(root, scope, func(placement *SubgraphInstanceNode) {
		if copied := placement.BuiltGraph(); copied != nil && err == nil {
			err = edit(copied)
		}
	})
	return err
}

// refreshPlacers brings every graph that places this definition, directly
// or through another definition, back in line with it.
func (a *Graph) refreshPlacers(policy conflictPolicy) error {
	a.definitionChanged()
	if a.deferredToCompoundEdit() {
		return nil
	}
	changed := a.SubGraphScopeID()
	if changed == "" {
		return nil
	}

	root := a.Root()
	root.refreshSubGraphNodeType(changed)

	placers, affected := root.placersOf(changed)
	for _, placer := range placers {
		lost, err := placer.refreshPlacements(affected, policy)
		if err != nil {
			return err
		}
		dropped, err := placer.settleUnder(policy)
		if err != nil {
			return err
		}
		if len(lost)+len(dropped) > 0 {
			placer.definitionChanged()
		}
	}
	return nil
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
// affected definition, and unwires what fed any input that is gone.
func (a *Graph) refreshPlacements(affected map[string]bool, policy conflictPolicy) ([]DroppedEdge, error) {
	ids := make([]string, 0)
	for node, id := range a.nodeIDs {
		if placement, ok := node.(*SubgraphInstanceNode); ok && affected[placement.subGraphID] {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)

	var lost []DroppedEdge
	for _, id := range ids {
		placement := a.Node(id).(*SubgraphInstanceNode)

		unwired := placement.invalidate()
		if err := placement.buildIfNeeded(); err != nil {
			return lost, fmt.Errorf("rebuilding placement %s of subgraph %q: %w", id, placement.subGraphID, err)
		}
		inputs := make([]string, 0, len(unwired))
		for input := range unwired {
			inputs = append(inputs, input)
		}
		sort.Strings(inputs)
		for _, input := range inputs {
			source := unwired[input]
			drop := DroppedEdge{
				Scope: a.SubGraphScopeID(),
				From:  a.nodeIDs[source.Node()], FromPort: source.Name(),
				To: id, ToPort: input,
				Reason: fmt.Sprintf("subgraph %q no longer has an input %q", placement.subGraphID, input),
			}
			if policy == refuseConflicts {
				return lost, fmt.Errorf("%s", drop.Reason)
			}
			a.recordDrop(drop)
			lost = append(lost, drop)
		}
	}
	return lost, nil
}
