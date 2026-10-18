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

// A load that fails, even after it has started replacing the open graph,
// leaves the open graph as it was.
func (a *Instance) ApplyAppSchema(jsonPayload []byte) (err error) {
	a.mu().Lock()
	defer a.mu().Unlock()

	before, err := a.encodeToAppSchema()
	if err != nil {
		return fmt.Errorf("saving the open graph before loading over it: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("loading graph: %v", r)
		}
		if err == nil {
			return
		}
		if restoreErr := a.applyAppSchema(before); restoreErr != nil {
			err = fmt.Errorf("%w (and restoring the open graph failed: %v)", err, restoreErr)
		}
	}()
	return a.applyAppSchema(jsonPayload)
}

func parseSavedGraph(payload []byte) (app persistence.App, decoder jbtf.Decoder, err error) {
	if app, err = jbtf.Unmarshal[persistence.App](payload); err != nil {
		return app, decoder, fmt.Errorf("unable to parse graph as a jbtf: %w", err)
	}
	if decoder, err = jbtf.NewDecoder(payload); err != nil {
		return app, decoder, fmt.Errorf("unable to build a jbtf decoder: %w", err)
	}
	return app, decoder, nil
}

func (a *Instance) applyAppSchema(jsonPayload []byte) error {
	// Parsed before anything is cleared: loading is destructive from here on.
	app, decoder, err := parseSavedGraph(jsonPayload)
	if err != nil {
		return err
	}

	a.Reset()
	a.modelVersion++

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

	if err := a.loadNodes(app.Nodes, decoder); err != nil {
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
	return nil
}

// loadNodes fills an empty graph from its saved nodes and settles it.
func (a *Graph) loadNodes(saved map[string]persistence.Node, decoder jbtf.Decoder) error {
	order, cyclic := dependenciesFirst(saved, func(node persistence.Node) []string {
		producers := make([]string, 0, len(node.AssignedInput))
		for _, from := range node.AssignedInput {
			producers = append(producers, from.NodeId)
		}
		return producers
	})
	if cyclic != "" {
		return fmt.Errorf("node %s reads its own output, directly or through other nodes", cyclic)
	}

	created := make(map[string]nodes.Node, len(saved))
	for _, id := range order {
		node, err := a.instantiateAppNode(id, saved[id])
		if err != nil {
			return err
		}
		created[id] = node
		if custom, ok := node.(CustomGraphSerialization); ok {
			if err := custom.FromJSON(decoder, saved[id].Data); err != nil {
				return fmt.Errorf("node %s: %w", id, err)
			}
		}
	}

	for _, id := range order {
		details := saved[id]
		reads := make(map[string][]source)
		inputs := created[id].Inputs()
		// Sorted, so an array's elements are appended by index.
		for _, input := range slices.SortedFunc(maps.Keys(details.AssignedInput), compareInputNames) {
			from := details.AssignedInput[input]
			producer, ok := created[from.NodeId]
			if !ok {
				return fmt.Errorf("node %s's %q input reads node %s, which does not exist", id, input, from.NodeId)
			}
			port, _, _ := inputKey(inputs, input)
			reads[port] = append(reads[port], source{node: producer, port: from.PortName})
		}
		a.reads[created[id]] = reads
	}

	_, err := a.settle(refuseConflicts)
	return err
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
	ids := slices.Sorted(maps.Keys(items))

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
	ordered, _ := dependenciesFirst(subGraphs, placedSubGraphs)
	return ordered
}

func placedSubGraphs(def persistence.SubGraph) []string {
	var placed []string
	for _, node := range def.Nodes {
		if subgraph.IsRuntimeNodeType(node.Type) {
			placed = append(placed, subgraph.RuntimeTypeID(node.Type))
		}
	}
	return placed
}

func (a *Instance) loadSubGraphDefinition(id string, def persistence.SubGraph, decoder jbtf.Decoder) error {
	if err := a.createSubGraph(id, def.Name, def.Description); err != nil {
		return err
	}
	a.subGraphs[id].group = def.Group
	return a.subGraphs[id].instance.loadSubGraphContents(def, decoder)
}

func (a *Graph) loadSubGraphContents(def persistence.SubGraph, decoder jbtf.Decoder) error {
	if def.Notes != nil {
		a.metadata.Set("notes", def.Notes)
	}
	if def.Metadata != nil {
		a.metadata.OverwriteData(def.Metadata)
	}
	return a.loadNodes(def.Nodes, decoder)
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
		if !project.typeFactory.KeyRegistered(saved.Type) {
			return nil, fmt.Errorf("node %q has unknown type %q", nodeID, saved.Type)
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
