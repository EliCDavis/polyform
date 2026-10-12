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
//
// A selected variable reference is copied in rather than cut: reading a
// variable is free anywhere, so it stays behind as well while anything
// outside still reads it, and it never becomes a port.
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
	ports, err := planPorts(selection, inbound, outbound, parent.layoutOf(selection))
	if err != nil {
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

	id := a.freeSubGraphID(strings.Join(strings.Fields(name), "_"))
	if err := a.createSubGraph(id, name, strings.TrimSpace(description)); err != nil {
		return result, err
	}
	defer func() {
		if err != nil {
			_ = a.deleteSubGraph(id)
		}
	}()
	child := a.subGraphs[id].instance

	// Loaded as one graph, boundaries included, so a node typed by what
	// feeds it from outside settles with that in place.
	if err := copySelection(parent, child, selection, ports); err != nil {
		return result, err
	}

	typePath := subgraph.RuntimeTypePath(id)
	_, placed, err := parent.createNode(typePath, "")
	if err != nil {
		return result, err
	}
	if ports.layout.known {
		parent.setNodePosition(placed, ports.layout.center)
	}

	fed := map[string]bool{}
	for _, cut := range inbound {
		key := inboundSourceKey(cut)
		if fed[key] {
			continue
		}
		fed[key] = true
		if err := parent.connectNodes(cut.producerID, cut.output, placed, ports.inputs[key].name); err != nil {
			return result, err
		}
	}
	for _, cut := range outbound {
		if err := parent.connectNodes(placed, ports.outputs[sourceKey(cut)].name, cut.consumerID, cut.inputName()); err != nil {
			return result, err
		}
	}

	// Deleted only now: an array reader keeps its slot by having the new edge
	// replace the old one in place. References go last, once only outside
	// readers can be left holding them.
	ordered := slices.Sorted(maps.Keys(selection))
	slices.SortStableFunc(ordered, func(x, y string) int {
		return boolToInt(isVariableReference(selection[x])) - boolToInt(isVariableReference(selection[y]))
	})
	for _, selected := range ordered {
		if isVariableReference(selection[selected]) && parent.isRead(selection[selected]) {
			continue
		}
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
		selection[id] = node
	}
	return selection, nil
}

func isVariableReference(node nodes.Node) bool {
	_, ok := node.(variable.Reference)
	return ok
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (a *Graph) isRead(node nodes.Node) bool {
	for _, inputs := range a.reads {
		for _, sources := range inputs {
			if slices.ContainsFunc(sources, func(from source) bool { return from.node == node }) {
				return true
			}
		}
	}
	return false
}

// edgesCrossing splits the edges with one end in selection into those read
// by it and those read from it. Both come back in a fixed order, which is
// the order their ports are numbered in.
func (a *Graph) edgesCrossing(selection map[string]nodes.Node) (inbound, outbound []edge, err error) {
	for _, held := range a.edges() {
		_, producerSelected := selection[held.producerID]
		_, consumerSelected := selection[held.consumerID]
		if producerSelected == consumerSelected {
			continue
		}
		if producerSelected && isVariableReference(held.producer) {
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

	slices.SortStableFunc(outbound, func(x, y edge) int {
		return strings.Compare(sourceKey(x), sourceKey(y))
	})
	return inbound, outbound, nil
}

func sourceKey(e edge) string {
	return e.producerID + "\x00" + e.output
}

var genericOutputNames = map[string]bool{
	"Out": true, "Value": true, "Output": true, "Result": true,
	"Float": true, "Int": true, "Bool": true, "String": true,
}

func givenNodeName(node nodes.Node) string {
	switch n := node.(type) {
	case variable.Reference:
		return n.Reference().Info().Name()
	case *SubgraphInstanceNode:
		return n.Name()
	case Parameter:
		return strings.TrimSpace(n.DisplayName())
	}
	if boundary, ok := subgraph.IsInputBoundary(node); ok {
		if subgraph.BoundaryPortNameConfigured(boundary) {
			return boundary.BoundaryPortName()
		}
		return ""
	}
	port, ok := node.Inputs()["Name"].(nodes.SingleValueInputPort)
	if !ok || port.Value() == nil {
		return ""
	}
	if _, isLiteral := port.Value().Node().(Parameter); !isLiteral {
		return ""
	}
	if named, ok := port.Value().(nodes.Output[string]); ok {
		return strings.TrimSpace(named.Value())
	}
	return ""
}

// A named producer crossing through several outputs is told apart by output.
type portNamer struct {
	outputsCrossing map[string]int
}

func newPortNamer(cuts []edge) portNamer {
	seen := map[string]bool{}
	namer := portNamer{outputsCrossing: map[string]int{}}
	for _, cut := range cuts {
		if key := sourceKey(cut); !seen[key] {
			seen[key] = true
			namer.outputsCrossing[cut.producerID]++
		}
	}
	return namer
}

func (namer portNamer) producerName(e edge) string {
	name := givenNodeName(e.producer)
	if name != "" && namer.outputsCrossing[e.producerID] > 1 {
		return name + " " + e.output
	}
	return name
}

func (namer portNamer) inbound(e edge) string {
	if name := namer.producerName(e); name != "" {
		return name
	}
	if !genericOutputNames[e.output] {
		return e.output
	}
	if _, isArray := e.in.(nodes.ArrayValueInputPort); isArray {
		return e.output
	}
	return e.input
}

func (namer portNamer) outbound(e edge) string {
	if name := namer.producerName(e); name != "" {
		return name
	}
	return e.output
}

type portNames map[string]bool

func (taken portNames) claim(name string) string {
	unique := name
	for i := 2; taken[unique]; i++ {
		unique = fmt.Sprintf("%s %d", name, i)
	}
	taken[unique] = true
	return unique
}

// Separate reference nodes to one variable are one value, so they share a port.
func inboundSourceKey(e edge) string {
	if ref, ok := e.producer.(variable.Reference); ok {
		return fmt.Sprintf("var:%p", ref.Reference())
	}
	return sourceKey(e)
}

// refuseConvertCycles refuses a selection that an outside node both reads
// and feeds: as one subgraph it would have to read its own output.
func (a *Graph) refuseConvertCycles(selection map[string]nodes.Node, inbound []edge) error {
	selected := make(map[nodes.Node]bool, len(selection))
	for _, node := range selection {
		if !isVariableReference(node) {
			selected[node] = true
		}
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

		for _, read := range a.readBy(node) {
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

func (root *Instance) freeSubGraphID(base string) string {
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

// A boundary node the conversion will create, with everything decided
// before anything is touched.
type plannedPort struct {
	id, kind, portType, name string
	at                       vector2.Float64
}

type portPlan struct {
	layout selectionLayout
	// Input ports by inboundSourceKey, output ports by sourceKey.
	inputs, outputs map[string]plannedPort
	// The boundary each cut input reads, by consumer id and "Port" or "Port.N".
	rewired map[string]string
}

func planPorts(selection map[string]nodes.Node, inbound, outbound []edge, layout selectionLayout) (portPlan, error) {
	plan := portPlan{
		layout:  layout,
		inputs:  map[string]plannedPort{},
		outputs: map[string]plannedPort{},
		rewired: map[string]string{},
	}
	ids := boundaryIDs{taken: map[string]bool{}, next: len(selection)}
	for id := range selection {
		ids.taken[id] = true
	}

	namer := newPortNamer(slices.Concat(inbound, outbound))

	// One input per outside value, however many selected nodes read it.
	inputNames := portNames{}
	for _, cut := range inbound {
		key := inboundSourceKey(cut)
		if _, made := plan.inputs[key]; !made {
			port := plannedPort{
				id:       ids.claim(),
				kind:     subgraph.InputNodeTypeKey,
				portType: portTypeOf(cut.out),
				name:     inputNames.claim(namer.inbound(cut)),
				at:       vector2.New(layout.min.X()-boundaryLayoutGap.X(), layout.min.Y()+float64(len(plan.inputs))*boundaryLayoutGap.Y()),
			}
			if err := port.check(); err != nil {
				return plan, err
			}
			plan.inputs[key] = port
		}
		plan.rewired[cut.consumerID+"\x00"+cut.inputName()] = plan.inputs[key].id
	}

	// One output per selected port read from outside.
	outputNames := portNames{}
	for _, cut := range outbound {
		key := sourceKey(cut)
		if _, made := plan.outputs[key]; made {
			continue
		}
		port := plannedPort{
			id:       ids.claim(),
			kind:     subgraph.OutputNodeTypeKey,
			portType: portTypeOf(cut.out),
			name:     outputNames.claim(namer.outbound(cut)),
			at:       vector2.New(layout.max.X()+boundaryLayoutGap.X(), layout.min.Y()+float64(len(plan.outputs))*boundaryLayoutGap.Y()),
		}
		if err := port.check(); err != nil {
			return plan, err
		}
		plan.outputs[key] = port
	}
	return plan, nil
}

func (port plannedPort) check() error {
	if !nodes.IsPortTypeKnown(port.portType) {
		return fmt.Errorf("create boundary %q: unknown boundary port type %q", port.name, port.portType)
	}
	return nil
}

func (port plannedPort) saved(encoder *jbtf.Encoder) (persistence.Node, error) {
	var boundary CustomGraphSerialization
	switch port.kind {
	case subgraph.InputNodeTypeKey:
		boundary = subgraph.NewInputNode(port.name, port.portType)
	default:
		boundary = subgraph.NewOutputNode(port.name, port.portType)
	}
	data, err := boundary.ToJSON(encoder)
	return persistence.Node{Type: port.kind, Data: data}, err
}

// Ids for the boundaries, clear of the ids the selection keeps.
type boundaryIDs struct {
	taken map[string]bool
	next  int
}

func (ids *boundaryIDs) claim() string {
	for ; ; ids.next++ {
		candidate := fmt.Sprintf("Node-%d", ids.next)
		if !ids.taken[candidate] {
			ids.taken[candidate] = true
			return candidate
		}
	}
}

// copySelection loads the selected nodes, their layout, the edges between
// them and the boundaries at every cut edge into child as one graph.
func copySelection(parent, child *Graph, selection map[string]nodes.Node, ports portPlan) error {
	encoder := &jbtf.Encoder{}
	saved := make(map[string]persistence.Node, len(selection)+len(ports.inputs)+len(ports.outputs))
	for id, node := range selection {
		copied := parent.savedNode(node, encoder)
		for input, from := range copied.AssignedInput {
			if _, inside := selection[from.NodeId]; inside {
				continue
			}
			copied.AssignedInput[input] = schema.PortReference{NodeId: ports.rewired[id+"\x00"+input], PortName: subgraph.ValuePortName}
		}
		saved[id] = copied

		if layout, ok := parent.Metadata("nodes." + id).(map[string]any); ok {
			child.metadata.Set("nodes."+id, cloneMetadataMap(layout))
		}
	}
	for _, port := range ports.inputs {
		node, err := port.saved(encoder)
		if err != nil {
			return err
		}
		saved[port.id] = node
	}
	for key, port := range ports.outputs {
		node, err := port.saved(encoder)
		if err != nil {
			return err
		}
		producerID, output, _ := strings.Cut(key, "\x00")
		node.AssignedInput = map[string]schema.PortReference{subgraph.ValuePortName: {NodeId: producerID, PortName: output}}
		saved[port.id] = node
	}
	if ports.layout.known {
		for _, port := range ports.inputs {
			child.setNodePosition(port.id, port.at)
		}
		for _, port := range ports.outputs {
			child.setNodePosition(port.id, port.at)
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
