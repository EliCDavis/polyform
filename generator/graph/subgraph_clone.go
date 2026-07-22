package graph

import (
	"fmt"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/persistence"
	"github.com/EliCDavis/polyform/nodes"
)

// cloneSubGraphDefinition builds a private Instance that mirrors the editable
// sub-graph definition. Runtime nodes own these clones so each placement can
// evaluate independently while the definition remains the single edit target.
func (a *Instance) cloneSubGraphDefinition(subGraphID string) (*Instance, error) {
	root := a.Root()
	encoder := &jbtf.Encoder{}
	def, err := root.persistedSubGraphDefinition(subGraphID, encoder)
	if err != nil {
		return nil, err
	}

	payload, err := encoder.ToPgtf(persistence.App{
		SubGraphs: map[string]persistence.SubGraph{
			subGraphID: def,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("encode sub-graph %q for clone: %w", subGraphID, err)
	}

	decoder, err := jbtf.NewDecoder(payload)
	if err != nil {
		return nil, fmt.Errorf("decode sub-graph %q for clone: %w", subGraphID, err)
	}

	appSchema, err := jbtf.Unmarshal[persistence.App](payload)
	if err != nil {
		return nil, fmt.Errorf("unmarshal sub-graph %q for clone: %w", subGraphID, err)
	}

	clonedDef, ok := appSchema.SubGraphs[subGraphID]
	if !ok {
		return nil, fmt.Errorf("cloned payload missing sub-graph %q", subGraphID)
	}

	clone := newInstance(root)
	if err := populateInstanceFromSubGraphDef(clone, clonedDef, decoder); err != nil {
		return nil, err
	}
	return clone, nil
}

func (a *Instance) persistedSubGraphDefinition(id string, encoder *jbtf.Encoder) (persistence.SubGraph, error) {
	if err := a.assertRootGraph("encode subgraph definition"); err != nil {
		return persistence.SubGraph{}, err
	}
	a.initSubGraphs()

	runtime, exists := a.subGraphs[id]
	if !exists {
		return persistence.SubGraph{}, fmt.Errorf("sub-graph %q does not exist", id)
	}

	child := runtime.instance
	nodeInstances := make(map[string]persistence.Node)
	for node := range child.nodeIDs {
		nodeID := child.nodeIDs[node]
		encoded := child.buildNodeGraphInstanceSchema(node, encoder)
		// Prefer the factory registration key so clones rehydrate with the
		// same CreateNode type used when the definition was authored.
		if key, ok := child.nodeTypeKeys[node]; ok && key != "" {
			encoded.Type = key
		}
		nodeInstances[nodeID] = encoded
	}

	var noteMetadata map[string]any
	if notes := child.metadata.Get("notes"); notes != nil {
		if casted, ok := notes.(map[string]any); ok {
			noteMetadata = casted
		}
	}

	return persistence.SubGraph{
		Name:        runtime.name,
		Description: runtime.description,
		Nodes:       nodeInstances,
		Notes:       noteMetadata,
		Metadata:    child.metadata.Data(),
	}, nil
}

func populateInstanceFromSubGraphDef(target *Instance, def persistence.SubGraph, decoder jbtf.Decoder) error {
	if def.Notes != nil {
		target.metadata.Set("notes", def.Notes)
	}
	if def.Metadata != nil {
		target.metadata.OverwriteData(def.Metadata)
	}

	createdNodes := make(map[string]nodes.Node)
	for nodeID, instanceDetails := range def.Nodes {
		node, err := target.instantiateAppNode(nodeID, instanceDetails)
		if err != nil {
			return err
		}
		if node != nil {
			createdNodes[nodeID] = node
		}
	}

	if err := applyPersistedNodeData(def.Nodes, createdNodes, decoder); err != nil {
		return err
	}
	return target.connectAppNodes(def.Nodes, createdNodes)
}

// forEachSubGraphInstance visits every live placement of subGraphID across the
// root graph and all nested sub-graph definitions.
func forEachSubGraphInstance(graph *Instance, subGraphID string, fn func(*SubgraphInstanceNode)) {
	visit := func(inst *Instance) {
		if inst == nil {
			return
		}
		for node := range inst.nodeIDs {
			runtime, ok := node.(*SubgraphInstanceNode)
			if !ok || runtime.subGraphID != subGraphID {
				continue
			}
			fn(runtime)
		}
	}

	root := graph.Root()
	visit(root)
	root.initSubGraphs()
	for _, sg := range root.subGraphs {
		visit(sg.instance)
	}
}

func (a *Instance) rebuildSubGraphClones(subGraphID string) error {
	var firstErr error
	forEachSubGraphInstance(a, subGraphID, func(runtime *SubgraphInstanceNode) {
		if firstErr != nil {
			return
		}
		if err := runtime.rebuildClone(); err != nil {
			firstErr = fmt.Errorf("rebuild clone for sub-graph %q: %w", subGraphID, err)
		}
	})
	return firstErr
}

// notifyDefinitionMutation rebuilds every runtime clone of this definition so
// edits to the shared template are reflected in live placements.
func (a *Instance) notifyDefinitionMutation() error {
	if a.parent == nil {
		return nil
	}
	id := a.SubGraphScopeID()
	if id == "" {
		return nil
	}
	return a.parent.onSubGraphChildMutation(id)
}
