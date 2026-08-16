package graph

import (
	"sort"

	"github.com/EliCDavis/polyform/generator/persistence"
	"github.com/EliCDavis/polyform/generator/subgraph"
)

// subGraphLoadOrder returns subgraph IDs ordered so that any subgraph
// referenced as a node type (subgraph/<id>) within another subgraph's nodes
// is ordered before that referencing subgraph.
//
// This matters because instantiating a subgraph-type node eagerly clones the
// referenced subgraph's current definition (SubgraphInstanceNode.rebuildClone,
// via cloneSubGraphDefinition). If the referenced subgraph hasn't had its own
// nodes populated yet, the clone comes back with no boundary ports, and wiring
// any input into it panics. Go's map iteration order is randomized, so loading
// subgraph definitions in map order intermittently panics on any graph with a
// subgraph nested inside another subgraph.
//
// Falls back to alphabetical order among ties for determinism, and tolerates
// cycles (each subgraph is only ever visited once) rather than erroring,
// since a cycle would be a different, pre-existing problem this function
// isn't responsible for catching.
func subGraphLoadOrder(subGraphs map[string]persistence.SubGraph) []string {
	ids := make([]string, 0, len(subGraphs))
	for id := range subGraphs {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	ordered := make([]string, 0, len(ids))
	visited := make(map[string]bool, len(ids))
	visiting := make(map[string]bool, len(ids))

	var visit func(id string)
	visit = func(id string) {
		if visited[id] || visiting[id] {
			return
		}
		def, ok := subGraphs[id]
		if !ok {
			return
		}
		visiting[id] = true

		deps := make(map[string]struct{})
		for _, node := range def.Nodes {
			if !subgraph.IsRuntimeNodeType(node.Type) {
				continue
			}
			dep := subgraph.RuntimeTypeID(node.Type)
			if dep == "" || dep == id {
				continue
			}
			if _, exists := subGraphs[dep]; exists {
				deps[dep] = struct{}{}
			}
		}

		depIDs := make([]string, 0, len(deps))
		for dep := range deps {
			depIDs = append(depIDs, dep)
		}
		sort.Strings(depIDs)
		for _, dep := range depIDs {
			visit(dep)
		}

		visiting[id] = false
		visited[id] = true
		ordered = append(ordered, id)
	}

	for _, id := range ids {
		visit(id)
	}

	return ordered
}
