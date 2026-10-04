package graph

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/persistence"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

type ConvertSelectionResult struct {
	SubGraphID    string
	Name          string
	RuntimeNodeID string
	NodeType      schema.NodeType
}

// How far a new boundary node sits from the nodes it was cut from, and from
// the boundary before it.
var boundaryLayoutGap = vector2.New(220., 100.)

// ConvertSelectionToSubGraph moves the given nodes out of scope into a new
// subgraph definition and puts one placement of it where they were. Every
// edge that crossed the selection becomes a port on the new subgraph.
func (a *Instance) ConvertSelectionToSubGraph(scope Scope, nodeIDs []string, name, description string) (result ConvertSelectionResult, err error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return result, fmt.Errorf("name is required")
	}
	if len(nodeIDs) == 0 {
		return result, fmt.Errorf("at least one node id is required")
	}

	a.mu().Lock()
	defer a.mu().Unlock()

	parent, err := scope.ResolveInstance(a)
	if err != nil {
		return result, err
	}
	selection, err := parent.convertibleNodes(nodeIDs)
	if err != nil {
		return result, err
	}
	inbound, outbound, err := parent.edgesCrossing(selection)
	if err != nil {
		return result, err
	}
	if err := parent.refuseConvertCycles(selection, inbound); err != nil {
		return result, err
	}

	// The scope is a half wired graph until the last edit here, so nothing
	// placing it is brought in line until then.
	end := parent.beginCompoundEdit()
	defer func() {
		if endErr := end(); endErr != nil && err == nil {
			err = endErr
		}
	}()

	id := a.freeSubGraphID(name)
	if err := a.createSubGraph(id, name, strings.TrimSpace(description)); err != nil {
		return result, err
	}
	defer func() {
		if err != nil {
			_ = a.deleteSubGraph(id)
		}
	}()
	child := a.subGraphs[id].instance

	if err := copySelection(parent, child, selection); err != nil {
		return result, err
	}
	layout := parent.layoutOf(selection)

	// One input per outside value, however many selected nodes read it.
	inputs := map[string]string{}
	inputBoundaries := map[string]string{}
	for _, cut := range inbound {
		key := inboundSourceKey(cut)
		if _, made := inputs[key]; !made {
			port := fmt.Sprintf("Input %d", len(inputs)+1)
			at := vector2.New(layout.min.X()-boundaryLayoutGap.X(), layout.min.Y()+float64(len(inputs))*boundaryLayoutGap.Y())
			boundary, err := child.addBoundary(subgraph.InputNodeTypeKey, portTypeOf(cut.out), port, layout, at)
			if err != nil {
				return result, err
			}
			inputs[key], inputBoundaries[key] = port, boundary
		}
		// Elements inside the selection were appended by the copy, so this
		// one is appended behind them rather than put back at its index.
		if err := child.connectNodes(inputBoundaries[key], subgraph.ValuePortName, cut.consumerID, cut.input); err != nil {
			return result, err
		}
	}

	// One output per selected port read from outside.
	outputs := map[string]string{}
	for _, cut := range outbound {
		key := sourceKey(cut)
		if _, made := outputs[key]; made {
			continue
		}
		port := fmt.Sprintf("Output %d", len(outputs)+1)
		at := vector2.New(layout.max.X()+boundaryLayoutGap.X(), layout.min.Y()+float64(len(outputs))*boundaryLayoutGap.Y())
		boundary, err := child.addBoundary(subgraph.OutputNodeTypeKey, portTypeOf(cut.out), port, layout, at)
		if err != nil {
			return result, err
		}
		if err := child.connectNodes(cut.producerID, cut.output, boundary, subgraph.ValuePortName); err != nil {
			return result, err
		}
		outputs[key] = port
	}

	typePath := subgraph.RuntimeTypePath(id)
	_, placed, err := parent.createNode(typePath, "")
	if err != nil {
		return result, err
	}
	if layout.known {
		parent.setNodePosition(placed, layout.center)
	}

	fed := map[string]bool{}
	for _, cut := range inbound {
		key := inboundSourceKey(cut)
		if fed[key] {
			continue
		}
		fed[key] = true
		if err := parent.connectNodes(cut.producerID, cut.output, placed, inputs[key]); err != nil {
			return result, err
		}
	}
	for _, cut := range outbound {
		if err := parent.connectNodes(placed, outputs[sourceKey(cut)], cut.consumerID, cut.inputName()); err != nil {
			return result, err
		}
	}

	// Deleted only now: an array reader keeps its slot by having the new edge
	// replace the old one in place.
	for _, selected := range slices.Sorted(maps.Keys(selection)) {
		if _, err := parent.deleteNodeByID(selected); err != nil {
			return result, err
		}
	}

	return ConvertSelectionResult{
		SubGraphID:    id,
		Name:          name,
		RuntimeNodeID: placed,
		NodeType:      BuildNodeTypeSchema(typePath, NewRuntimeNode(a, id)),
	}, nil
}

func (a *Graph) convertibleNodes(nodeIDs []string) (map[string]nodes.Node, error) {
	selection := make(map[string]nodes.Node, len(nodeIDs))
	for _, id := range nodeIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			return nil, fmt.Errorf("node id cannot be empty")
		}
		if _, dup := selection[id]; dup {
			return nil, fmt.Errorf("duplicate node id %q", id)
		}
		node, ok := a.nodesByID[id]
		if !ok {
			return nil, fmt.Errorf("no node exists with id %q", id)
		}
		if _, isBoundary := subgraph.IsBoundaryNode(node); isBoundary {
			return nil, fmt.Errorf("cannot convert boundary node %q into a sub-graph", id)
		}
		if _, isVar := node.(variable.Reference); isVar {
			return nil, fmt.Errorf("cannot convert variable reference node %q into a sub-graph", id)
		}
		selection[id] = node
	}
	return selection, nil
}

// edgesCrossing splits the edges with one end in selection into those read
// by it and those read from it. Both come back in a fixed order, which is
// the order their ports are numbered in.
func (a *Graph) edgesCrossing(selection map[string]nodes.Node) (inbound, outbound []heldEdge, err error) {
	for _, held := range a.heldEdges() {
		_, producerSelected := selection[held.producerID]
		_, consumerSelected := selection[held.consumerID]
		if producerSelected == consumerSelected {
			continue
		}
		if portTypeOf(held.out) == "" {
			return nil, nil, fmt.Errorf("unable to resolve port type for output %q of node %s", held.output, held.producerID)
		}
		if consumerSelected {
			inbound = append(inbound, held)
		} else {
			outbound = append(outbound, held)
		}
	}

	slices.SortStableFunc(outbound, func(x, y heldEdge) int {
		return strings.Compare(sourceKey(x), sourceKey(y))
	})
	return inbound, outbound, nil
}

func (e heldEdge) inputName() string {
	return edge{to: portEnd{port: e.input}, element: e.element}.inputName()
}

func sourceKey(e heldEdge) string {
	return e.producerID + "\x00" + e.output
}

// Separate reference nodes to one variable are one value, so they share a port.
func inboundSourceKey(e heldEdge) string {
	if ref, ok := e.producer.(variable.Reference); ok {
		return fmt.Sprintf("var:%p", ref.Reference())
	}
	return sourceKey(e)
}

// refuseConvertCycles refuses a selection that an outside node both reads
// and feeds: as one subgraph it would have to read its own output.
func (a *Graph) refuseConvertCycles(selection map[string]nodes.Node, inbound []heldEdge) error {
	selected := make(map[nodes.Node]bool, len(selection))
	for _, node := range selection {
		selected[node] = true
	}

	checked := map[string]bool{}
	for _, cut := range inbound {
		if checked[cut.producerID] {
			continue
		}
		checked[cut.producerID] = true

		if reached, through, ok := a.pathBackTo(selected, cut.producer); ok {
			return fmt.Errorf("converting would create a cycle: %s feeds %s.%s, but %s itself depends on selected node %s (through %s); include %s in the selection or leave %s out",
				cut.producerID, cut.consumerID, cut.inputName(), cut.producerID, reached, strings.Join(through, " -> "), cut.producerID, cut.consumerID)
		}
	}
	return nil
}

// pathBackTo walks upstream from start and reports the first selected node
// it reaches, with the ids on the way there.
func (a *Graph) pathBackTo(selected map[nodes.Node]bool, start nodes.Node) (reached string, through []string, found bool) {
	visited := map[nodes.Node]bool{}

	var walk func(node nodes.Node, trail []string) bool
	walk = func(node nodes.Node, trail []string) bool {
		if visited[node] {
			return false
		}
		visited[node] = true

		for _, read := range flattenNodeInputReferences(node) {
			if selected[read] {
				reached, through = a.nodeIDs[read], trail
				return true
			}
			if walk(read, append(trail, a.nodeIDs[read])) {
				return true
			}
		}
		return false
	}

	found = walk(start, []string{a.nodeIDs[start]})
	return reached, through, found
}

func (root *Instance) freeSubGraphID(name string) string {
	base := strings.Join(strings.Fields(name), "_")
	if _, taken := root.subGraphs[base]; !taken {
		return base
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s_%d", base, i)
		if _, taken := root.subGraphs[candidate]; !taken {
			return candidate
		}
	}
}

// copySelection copies the selected nodes, their layout, and the edges
// between them. Edges to anything outside are left for boundaries.
func copySelection(parent, child *Graph, selection map[string]nodes.Node) error {
	encoder := &jbtf.Encoder{}
	saved := make(map[string]persistence.Node, len(selection))
	for id, node := range selection {
		copied := parent.savedNode(node, encoder)
		maps.DeleteFunc(copied.AssignedInput, func(_ string, from schema.PortReference) bool {
			_, inside := selection[from.NodeId]
			return !inside
		})
		saved[id] = copied

		if layout, ok := parent.Metadata("nodes." + id).(map[string]any); ok {
			child.metadata.Set("nodes."+id, cloneMetadataMap(layout))
		}
	}

	decoder, err := decoderOf(encoder)
	if err != nil {
		return fmt.Errorf("copying the selection: %w", err)
	}
	return child.loadNodes(savedGraph{nodes: saved, decoder: decoder, originals: selection}, nil)
}

func cloneMetadataMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		if nested, ok := v.(map[string]any); ok {
			v = cloneMetadataMap(nested)
		}
		out[k] = v
	}
	return out
}

func (a *Graph) addBoundary(typeKey, portType, name string, layout selectionLayout, at vector2.Float64) (string, error) {
	_, id, err := a.createNode(typeKey, portType)
	if err != nil {
		return "", fmt.Errorf("create boundary %q: %w", name, err)
	}
	if err := a.setBoundaryNodeInfo(id, name); err != nil {
		return "", err
	}
	if layout.known {
		a.setNodePosition(id, at)
	}
	return id, nil
}

// Where the selection sat in the editor. Not known when none of it had been
// laid out, and then nothing new is given a position either.
type selectionLayout struct {
	min, max, center vector2.Float64
	known            bool
}

func (a *Graph) layoutOf(selection map[string]nodes.Node) selectionLayout {
	var layout selectionLayout
	var sum vector2.Float64
	count := 0.
	for id := range selection {
		at, ok := a.nodePosition(id)
		if !ok {
			continue
		}
		if !layout.known {
			layout.min, layout.max, layout.known = at, at, true
		}
		layout.min, layout.max = vector2.Min(layout.min, at), vector2.Max(layout.max, at)
		sum = sum.Add(at)
		count++
	}
	if layout.known {
		layout.center = sum.DivByConstant(count)
	}
	return layout
}

func (a *Graph) nodePosition(id string) (vector2.Float64, bool) {
	meta, _ := a.Metadata("nodes." + id).(map[string]any)
	position, _ := meta["position"].(map[string]any)
	x, xOk := asFloat64(position["x"])
	y, yOk := asFloat64(position["y"])
	return vector2.New(x, y), xOk && yOk
}

func (a *Graph) setNodePosition(id string, at vector2.Float64) {
	a.metadata.Set("nodes."+id+".position", map[string]any{"x": at.X(), "y": at.Y()})
}

func asFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}
