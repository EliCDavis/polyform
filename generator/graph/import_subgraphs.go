package graph

import (
	"cmp"
	"fmt"
	"maps"
	"slices"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/persistence"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/subgraph"
)

type ImportedSubGraph struct {
	ID string
	// Set when ID had to differ from the id in the payload.
	OriginalID string
	Name       string
	NodeType   schema.NodeType
	Inputs     []VariableInput
}

type ImportSubGraphsResult struct {
	Group    string
	Imported []ImportedSubGraph
}

// ImportSubGraphDefinitions adds every subgraph definition in a saved graph
// to this one, and nothing else from it. An id already taken gets a numeric
// suffix (Adder becomes Adder_2), and placements of it inside the other
// imported definitions follow. Variables they read become inputs, and they
// are all filed under group, or the saved graph's name when group is empty.
func (a *Instance) ImportSubGraphDefinitions(payload []byte, group string) (result ImportSubGraphsResult, err error) {
	app, decoder, err := parseSavedGraph(payload)
	if err != nil {
		return result, err
	}
	result.Group = cmp.Or(group, app.Name)
	if _, cyclic := dependenciesFirst(app.SubGraphs, placedSubGraphs); cyclic != "" {
		return result, fmt.Errorf("imported sub-graph %q contains itself through its placements", cyclic)
	}

	_, err = a.write(refuseConflicts, func() error {
		result, err = a.importSubGraphDefinitions(app, decoder, result)
		return err
	})
	return result, err
}

func (a *Instance) importSubGraphDefinitions(app persistence.App, decoder jbtf.Decoder, result ImportSubGraphsResult) (_ ImportSubGraphsResult, err error) {
	defs, inputs, err := a.liftVariables(app)
	if err != nil {
		return result, err
	}

	// Every definition exists before any is filled in, so one that places
	// another finds its type.
	newIDs := make(map[string]string, len(app.SubGraphs))
	defer func() {
		if err != nil {
			for _, id := range newIDs {
				_ = a.deleteSubGraph(id)
			}
		}
	}()
	for _, id := range slices.Sorted(maps.Keys(app.SubGraphs)) {
		newID := a.freeSubGraphID(cmp.Or(id, "Subgraph"))
		if err := a.createSubGraph(newID, app.SubGraphs[id].Name, app.SubGraphs[id].Description); err != nil {
			return result, err
		}
		a.subGraphs[newID].group = result.Group
		newIDs[id] = newID
	}

	for _, id := range subGraphLoadOrder(app.SubGraphs) {
		def, newID := defs[id], newIDs[id]

		renamed := make(map[string]persistence.Node, len(def.Nodes))
		for nodeID, node := range def.Nodes {
			if subgraph.IsRuntimeNodeType(node.Type) {
				placed, ok := newIDs[subgraph.RuntimeTypeID(node.Type)]
				if !ok {
					return result, fmt.Errorf("imported sub-graph %q references unknown sub-graph %q", def.Name, subgraph.RuntimeTypeID(node.Type))
				}
				node.Type = subgraph.RuntimeTypePath(placed)
			}
			renamed[nodeID] = node
		}
		def.Nodes = renamed

		if err := a.subGraphs[newID].instance.loadSubGraphContents(def, decoder); err != nil {
			return result, fmt.Errorf("populate imported sub-graph %q: %w", newID, err)
		}

		imported := ImportedSubGraph{
			ID:       newID,
			Name:     def.Name,
			NodeType: BuildNodeTypeSchema(subgraph.RuntimeTypePath(newID), NewRuntimeNode(a, newID)),
			Inputs:   inputs[id],
		}
		if newID != id {
			imported.OriginalID = id
		}
		result.Imported = append(result.Imported, imported)
	}

	return result, nil
}
