package graph

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/persistence"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/generator/variant"
	"github.com/EliCDavis/polyform/nodes"
)

func (a *Instance) ApplyAppSchema(jsonPayload []byte) error {
	a.mu().Lock()
	defer a.mu().Unlock()

	// Parsed before anything is cleared: loading is destructive from here on.
	app, err := jbtf.Unmarshal[persistence.App](jsonPayload)
	if err != nil {
		return fmt.Errorf("unable to parse graph as a jbtf: %w", err)
	}
	decoder, err := jbtf.NewDecoder(jsonPayload)
	if err != nil {
		return fmt.Errorf("unable to build a jbtf decoder: %w", err)
	}

	a.Reset()

	a.details = Details{Name: app.Name, Version: app.Version, Description: app.Description, Authors: app.Authors}
	a.metadata.OverwriteData(app.Metadata)

	app.Variables.Traverse(func(path string, saved persistence.Variable) bool {
		var loaded variable.Variable
		loaded, err = variable.DeserializePersistantVariableJSON(saved.Data, decoder, a.variableFactory)
		if err != nil {
			return false
		}
		if _, err = a.newVariable(path, loaded); err != nil {
			return false
		}
		loaded.Info().SetDescription(saved.Description)
		return true
	})
	if err != nil {
		return err
	}

	for name, profile := range app.Profiles {
		a.profiles.Set(name, profile.Data)
	}

	for name, saved := range app.Variants {
		dimensions := make([]variant.Dimension, 0, len(saved.Dimensions))
		for path, raw := range saved.Dimensions {
			dimension, err := variant.UnmarshalDimension(path, raw)
			if err != nil {
				return fmt.Errorf("decoding variant dimension %q in set %q: %w", path, name, err)
			}
			dimensions = append(dimensions, dimension)
		}
		a.variantSets.Set(name, variant.Set{Dimensions: dimensions})
	}

	for _, id := range subGraphLoadOrder(app.SubGraphs) {
		if err := a.loadSubGraphDefinition(id, app.SubGraphs[id], decoder); err != nil {
			return err
		}
	}

	if err := a.loadNodes(savedGraph{nodes: app.Nodes, decoder: decoder}, nil); err != nil {
		return err
	}

	for name, saved := range app.Producers {
		producer, ok := a.nodesByID[saved.NodeID]
		if !ok {
			return fmt.Errorf("producer %q: no node exists with id %q", name, saved.NodeID)
		}
		output, ok := producer.Outputs()[saved.Port].(nodes.Output[manifest.Manifest])
		if !ok {
			return fmt.Errorf("producer %q: node %q has no %q output producing a manifest", name, saved.NodeID, saved.Port)
		}
		a.namedManifests.NamePort(name, saved.Port, producer, output)
	}

	a.incModelVersion()
	return nil
}

func (a *Graph) loadNodes(saved savedGraph, externals map[string]nodes.OutputPort) error {
	a.clearCopySource()
	order, cyclic := dependenciesFirst(saved.nodes, func(node persistence.Node) []string {
		producers := make([]string, 0, len(node.AssignedInput))
		for _, from := range node.AssignedInput {
			producers = append(producers, from.NodeId)
		}
		return producers
	})
	if cyclic != "" {
		return fmt.Errorf("node %s reads its own output, directly or through other nodes", cyclic)
	}

	created := make(map[string]nodes.Node, len(saved.nodes))
	for _, id := range order {
		node, err := a.instantiateAppNode(id, saved.nodes[id])
		if err != nil {
			return err
		}
		created[id] = node

		if err := saved.restore(id, node); err != nil {
			return fmt.Errorf("node %s: %w", id, err)
		}
		a.clearBoundaryCache()
		if boundary, ok := subgraph.IsInputBoundary(node); ok {
			if external, fed := externals[boundary.BoundaryPortName()]; fed {
				boundary.SetExternalSource(external)
			}
		}
	}

	for _, id := range order {
		details := saved.nodes[id]
		reads := make(map[string][]source)
		// Sorted, so an array's elements are appended by index.
		for _, input := range slices.SortedFunc(maps.Keys(details.AssignedInput), compareInputNames) {
			from := details.AssignedInput[input]
			producer, ok := created[from.NodeId]
			if !ok {
				return fmt.Errorf("node %s's %q input reads node %s, which does not exist", id, input, from.NodeId)
			}
			port, _ := splitElement(input)
			reads[port] = append(reads[port], source{node: producer, port: from.PortName})
		}
		a.reads[created[id]] = reads
	}

	_, err := a.settle(refuseConflicts)
	return err
}

type savedGraph struct {
	nodes   map[string]persistence.Node
	decoder jbtf.Decoder

	// The live nodes this was encoded from, when it is being copied.
	originals map[string]nodes.Node
}

func (saved savedGraph) restore(id string, node nodes.Node) error {
	if copier, ok := node.(stateCopier); ok {
		if original, ok := saved.originals[id]; ok && copier.CopyStateFrom(original) {
			return nil
		}
	}
	if custom, ok := node.(CustomGraphSerialization); ok {
		return custom.FromJSON(saved.decoder, saved.nodes[id].Data)
	}
	return nil
}

func compareInputNames(x, y string) int {
	xPort, xElement := splitElement(x)
	yPort, yElement := splitElement(y)
	if xPort != yPort {
		return strings.Compare(xPort, yPort)
	}
	return xElement - yElement
}

// The order is complete even when cyclic names a key that depends on itself.
func dependenciesFirst[V any](items map[string]V, dependencies func(V) []string) (ordered []string, cyclic string) {
	ids := make([]string, 0, len(items))
	for id := range items {
		ids = append(ids, id)
	}
	slices.Sort(ids)

	const (
		visiting = iota + 1
		visited
	)
	state := make(map[string]int, len(ids))
	ordered = make([]string, 0, len(ids))

	var visit func(id string)
	visit = func(id string) {
		item, known := items[id]
		if !known || state[id] == visited {
			return
		}
		if state[id] == visiting {
			if cyclic == "" {
				cyclic = id
			}
			return
		}
		state[id] = visiting

		needs := dependencies(item)
		slices.Sort(needs)
		for _, need := range needs {
			visit(need)
		}

		state[id] = visited
		ordered = append(ordered, id)
	}

	for _, id := range ids {
		visit(id)
	}
	return ordered, cyclic
}

// Placing a definition copies it as it stands, so it has to be loaded first.
func subGraphLoadOrder(subGraphs map[string]persistence.SubGraph) []string {
	ordered, _ := dependenciesFirst(subGraphs, func(def persistence.SubGraph) []string {
		var placed []string
		for _, node := range def.Nodes {
			if subgraph.IsRuntimeNodeType(node.Type) {
				placed = append(placed, subgraph.RuntimeTypeID(node.Type))
			}
		}
		return placed
	})
	return ordered
}

func (a *Instance) loadSubGraphDefinition(id string, def persistence.SubGraph, decoder jbtf.Decoder) error {
	if err := a.createSubGraph(id, def.Name, def.Description); err != nil {
		return err
	}
	definition := a.subGraphs[id].instance
	if err := definition.loadSubGraphContents(def, decoder); err != nil {
		return err
	}
	a.copySources[id] = savedGraph{nodes: def.Nodes, decoder: decoder, originals: definition.nodesByID}
	return nil
}

func (a *Graph) loadSubGraphContents(def persistence.SubGraph, decoder jbtf.Decoder) error {
	if def.Notes != nil {
		a.metadata.Set("notes", def.Notes)
	}
	if def.Metadata != nil {
		a.metadata.OverwriteData(def.Metadata)
	}
	return a.loadNodes(savedGraph{nodes: def.Nodes, decoder: decoder}, nil)
}

func (a *Graph) instantiateAppNode(nodeID string, saved persistence.Node) (nodes.Node, error) {
	if nodeID == "" {
		return nil, fmt.Errorf("attempting to create a node without an ID")
	}

	project := a.project
	var node nodes.Node
	if saved.Variable != nil {
		variable, err := project.variables.Variable(*saved.Variable)
		if err != nil {
			return nil, err
		}
		node = variable.NodeReference()
	} else {
		if subgraph.IsRuntimeNodeType(saved.Type) && !project.typeFactory.KeyRegistered(saved.Type) {
			project.RegisterSubGraphNodeType(subgraph.RuntimeTypeID(saved.Type))
		}
		built, ok := project.typeFactory.New(saved.Type).(nodes.Node)
		if !ok {
			return nil, fmt.Errorf("graph definition contained type that instantiated a non node: %s", saved.Type)
		}
		node = built
	}

	a.register(node, nodeID, saved.Type)
	return node, nil
}
