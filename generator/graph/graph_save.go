package graph

import (
	"encoding/json"
	"fmt"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/persistence"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/nodes"
)

func (a *Instance) EncodeToAppSchema() ([]byte, error) {
	a.mu().RLock()
	defer a.mu().RUnlock()

	encoder := &jbtf.Encoder{}

	app := persistence.App{
		Name:        a.details.Name,
		Version:     a.details.Version,
		Description: a.details.Description,
		Authors:     a.details.Authors,
		Nodes:       a.savedNodes(encoder),
		Producers:   a.producerSchema(),
		Profiles:    make(map[string]persistence.Profile),
		Variants:    make(map[string]persistence.VariantSet),
		Metadata:    a.metadata.Data(),
	}

	variablePaths := make(map[variable.Variable]string)
	a.variables.Traverse(func(path string, info variable.Info, v variable.Variable) {
		variablePaths[v] = path
	})
	for id, node := range a.nodesByID {
		if reference, ok := node.(variable.Reference); ok {
			saved := app.Nodes[id]
			path := variablePaths[reference.Reference()]
			saved.Variable = &path
			app.Nodes[id] = saved
		}
	}

	for name, data := range a.profiles.All() {
		app.Profiles[name] = persistence.Profile{Data: data}
	}

	for name, set := range a.variantSets.All() {
		encoded := make(map[string]json.RawMessage, len(set.Dimensions))
		for _, dim := range set.Dimensions {
			raw, err := json.Marshal(dim)
			if err != nil {
				return nil, fmt.Errorf("encoding variant dimension %q in set %q: %w", dim.Path(), name, err)
			}
			encoded[dim.Path()] = raw
		}
		app.Variants[name] = persistence.VariantSet{Dimensions: encoded}
	}

	var err error
	if app.Variables, err = a.variables.PersistedSchema(encoder); err != nil {
		return nil, err
	}
	if app.SubGraphs, err = a.encodeSubGraphDefinitions(encoder); err != nil {
		return nil, err
	}

	return encoder.ToPgtf(app)
}

func (a *Graph) savedNodes(encoder *jbtf.Encoder) map[string]persistence.Node {
	saved := make(map[string]persistence.Node, len(a.nodesByID))
	for id, node := range a.nodesByID {
		saved[id] = a.savedNode(node, encoder)
	}
	return saved
}

func (a *Graph) savedNode(node nodes.Node, encoder *jbtf.Encoder) persistence.Node {
	saved := persistence.Node{
		Type:          nodeTypeName(node),
		AssignedInput: a.assignedInputs(node),
	}

	// The key the node was created under is what load can create it from again.
	if key := a.nodeTypeKeys[node]; key != "" {
		saved.Type = key
	}

	if custom, ok := node.(CustomGraphSerialization); ok {
		data, err := custom.ToJSON(encoder)
		if err != nil {
			panic(err)
		}
		saved.Data = data
	}

	return saved
}
