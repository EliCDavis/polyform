package graph

import (
	"fmt"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/persistence"
	"github.com/EliCDavis/polyform/nodes"
)

// externals are in place before the clone is wired, so an array at an input lifts what reads it.
func (root *Instance) cloneSubGraphDefinition(subGraphID string, externals map[string]nodes.OutputPort) (*Graph, error) {
	source, err := root.copySource(subGraphID)
	if err != nil {
		return nil, err
	}

	clone := newGraph(root)
	if err := clone.loadNodes(source, externals); err != nil {
		return nil, err
	}
	return clone, nil
}

// A definition is encoded once and every copy is built from that, until an
// edit to it calls clearCopySource.
func (root *Instance) copySource(subGraphID string) (savedGraph, error) {
	if source, ok := root.copySources[subGraphID]; ok {
		return source, nil
	}

	encoder := &jbtf.Encoder{}
	def, err := root.persistedSubGraphDefinition(subGraphID, encoder)
	if err != nil {
		return savedGraph{}, err
	}

	decoder, err := decoderOf(encoder)
	if err != nil {
		return savedGraph{}, fmt.Errorf("copying sub-graph %q: %w", subGraphID, err)
	}

	source := savedGraph{
		nodes:     def.Nodes,
		decoder:   decoder,
		originals: root.subGraphs[subGraphID].instance.nodesByID,
	}
	root.copySources[subGraphID] = source
	return source, nil
}

// Through bytes: what nodes encoded can only be read back out of a decoder.
func decoderOf(encoder *jbtf.Encoder) (jbtf.Decoder, error) {
	payload, err := encoder.ToPgtf(persistence.App{})
	if err != nil {
		return jbtf.Decoder{}, err
	}
	return jbtf.NewDecoder(payload)
}

// clearCopySource has to be called by every edit to a definition's nodes
// or edges: copies are built from what copySource remembers of it.
func (a *Graph) clearCopySource() {
	if scope := a.SubGraphScopeID(); scope != "" {
		delete(a.project.copySources, scope)
	}
}

func (a *Instance) persistedSubGraphDefinition(id string, encoder *jbtf.Encoder) (persistence.SubGraph, error) {
	runtime, exists := a.subGraphs[id]
	if !exists {
		return persistence.SubGraph{}, fmt.Errorf("sub-graph %q does not exist", id)
	}

	child := runtime.instance
	notes, _ := child.metadata.Get("notes").(map[string]any)

	return persistence.SubGraph{
		Name:        runtime.name,
		Description: runtime.description,
		Nodes:       child.savedNodes(encoder),
		Notes:       notes,
		Metadata:    child.metadata.Data(),
	}, nil
}

// forEachSubGraphInstance visits every live placement of subGraphID across the
// root graph and all recursively nested sub-graph definitions.
func forEachSubGraphInstance(root *Instance, subGraphID string, fn func(holder *Graph, placement *SubgraphInstanceNode)) {
	visited := map[*Graph]bool{}

	var visit func(inst *Graph)
	visit = func(inst *Graph) {
		if inst == nil || visited[inst] {
			return
		}
		visited[inst] = true

		for node := range inst.nodeIDs {
			runtime, ok := node.(*SubgraphInstanceNode)
			if !ok {
				continue
			}
			if runtime.subGraphID == subGraphID {
				fn(inst, runtime)
			}

			visit(runtime.BuiltGraph())
		}
	}

	visit(root.Graph)
	for _, sg := range root.subGraphs {
		visit(sg.instance)
	}
}
