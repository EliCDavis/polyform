package graph

import (
	"fmt"
	"sort"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/persistence"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/subgraph"
)

// ImportedSubGraph describes one definition merged into the live graph.
type ImportedSubGraph struct {
	ID         string
	OriginalID string
	Name       string
	NodeType   schema.NodeType
}

// ImportSubGraphsResult is returned by ImportSubGraphDefinitions.
type ImportSubGraphsResult struct {
	Imported []ImportedSubGraph
}

// ImportSubGraphDefinitions merges every sub-graph definition from a persisted
// app payload into this (root) instance. Root nodes, variables, producers, and
// profiles in the payload are ignored. Conflicting IDs are renamed with a
// numeric suffix (Adder → Adder_2); nested subgraph/<id> type references are
// rewritten to match.
func (a *Instance) ImportSubGraphDefinitions(payload []byte) (ImportSubGraphsResult, error) {
	if err := a.assertRootGraph("import subgraph definitions"); err != nil {
		return ImportSubGraphsResult{}, err
	}
	a.initSubGraphs()

	appSchema, err := jbtf.Unmarshal[persistence.App](payload)
	if err != nil {
		return ImportSubGraphsResult{}, fmt.Errorf("unable to parse graph as a jbtf: %w", err)
	}

	decoder, err := jbtf.NewDecoder(payload)
	if err != nil {
		return ImportSubGraphsResult{}, fmt.Errorf("unable to build a jbtf decoder: %w", err)
	}

	if len(appSchema.SubGraphs) == 0 {
		return ImportSubGraphsResult{}, nil
	}

	oldIDs := make([]string, 0, len(appSchema.SubGraphs))
	for id := range appSchema.SubGraphs {
		oldIDs = append(oldIDs, id)
	}
	sort.Strings(oldIDs)

	remap := make(map[string]string, len(oldIDs))
	reserved := make(map[string]struct{}, len(oldIDs))
	for _, oldID := range oldIDs {
		newID := allocateImportSubGraphID(a, oldID, reserved)
		remap[oldID] = newID
		reserved[newID] = struct{}{}
	}

	created := make([]string, 0, len(oldIDs))
	rollback := func() {
		for i := len(created) - 1; i >= 0; i-- {
			_ = a.DeleteSubGraph(created[i])
		}
	}

	// Register all definitions first so nested runtime types resolve while
	// populating dependents.
	for _, oldID := range oldIDs {
		newID := remap[oldID]
		def := appSchema.SubGraphs[oldID]
		if err := a.CreateSubGraph(newID, def.Name, def.Description); err != nil {
			rollback()
			return ImportSubGraphsResult{}, err
		}
		created = append(created, newID)
	}

	result := ImportSubGraphsResult{
		Imported: make([]ImportedSubGraph, 0, len(oldIDs)),
	}

	for _, oldID := range subGraphLoadOrder(appSchema.SubGraphs) {
		newID := remap[oldID]
		remapped, err := remapSubGraphDefTypes(appSchema.SubGraphs[oldID], remap)
		if err != nil {
			rollback()
			return ImportSubGraphsResult{}, err
		}

		target, err := a.SubGraphInstance(newID)
		if err != nil {
			rollback()
			return ImportSubGraphsResult{}, err
		}
		if err := populateInstanceFromSubGraphDef(target, remapped, decoder); err != nil {
			rollback()
			return ImportSubGraphsResult{}, fmt.Errorf("populate imported sub-graph %q: %w", newID, err)
		}

		a.refreshSubGraphNodeType(newID)
		typePath := subgraph.RuntimeTypePath(newID)
		nodeType := BuildNodeTypeSchema(typePath, NewRuntimeNode(a, newID))

		entry := ImportedSubGraph{
			ID:       newID,
			Name:     remapped.Name,
			NodeType: nodeType,
		}
		if newID != oldID {
			entry.OriginalID = oldID
		}
		result.Imported = append(result.Imported, entry)
	}

	return result, nil
}

func allocateImportSubGraphID(root *Instance, base string, reserved map[string]struct{}) string {
	root.initSubGraphs()
	if base == "" {
		base = "Subgraph"
	}
	if importIDAvailable(root, base, reserved) {
		return base
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s_%d", base, i)
		if importIDAvailable(root, candidate, reserved) {
			return candidate
		}
	}
}

func importIDAvailable(root *Instance, id string, reserved map[string]struct{}) bool {
	if _, exists := root.subGraphs[id]; exists {
		return false
	}
	if _, taken := reserved[id]; taken {
		return false
	}
	return true
}

func remapSubGraphDefTypes(def persistence.SubGraph, idRemap map[string]string) (persistence.SubGraph, error) {
	out := persistence.SubGraph{
		Name:        def.Name,
		Description: def.Description,
		Notes:       def.Notes,
		Metadata:    def.Metadata,
		Nodes:       make(map[string]persistence.Node, len(def.Nodes)),
	}
	for nodeID, node := range def.Nodes {
		n := node
		if subgraph.IsRuntimeNodeType(n.Type) {
			oldRef := subgraph.RuntimeTypeID(n.Type)
			newRef, ok := idRemap[oldRef]
			if !ok {
				return persistence.SubGraph{}, fmt.Errorf(
					"imported sub-graph %q references unknown sub-graph %q",
					def.Name, oldRef,
				)
			}
			n.Type = subgraph.RuntimeTypePath(newRef)
		}
		out.Nodes[nodeID] = n
	}
	return out, nil
}
