package graph

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
)

// A placement of a subgraph. Ports are the definition's boundary nodes keyed
// by node id; the copy it evaluates through is derived, never maintained.
type SubgraphInstanceNode struct {
	owner      *Instance
	subGraphID string

	mu     sync.Mutex
	inputs map[string]nodes.OutputPort
	copied *copy

	each    bool
	runs    []*copy
	outputs map[string]*eachOutput
}

type copy struct {
	graph     *Graph
	version   uint64
	signature string
}

func NewRuntimeNode(owner *Instance, subGraphID string) *SubgraphInstanceNode {
	return &SubgraphInstanceNode{
		owner:      owner,
		subGraphID: subGraphID,
		inputs:     make(map[string]nodes.OutputPort),
	}
}

func (r *SubgraphInstanceNode) SubGraphID() string {
	return r.subGraphID
}

func (r *SubgraphInstanceNode) runtime() *subGraphRuntime {
	if definition, ok := r.owner.subGraphs[r.subGraphID]; ok {
		return definition
	}
	return &subGraphRuntime{name: "SubGraph", instance: newGraph(r.owner, "")}
}

func (r *SubgraphInstanceNode) Name() string {
	return r.runtime().name
}

func (r *SubgraphInstanceNode) Description() string {
	return r.runtime().description
}

func (r *SubgraphInstanceNode) Path() string {
	return "SubGraph"
}

func (r *SubgraphInstanceNode) definition() *Graph {
	return r.runtime().instance
}

// BuiltGraph is the copy if something has evaluated through the placement since the last edit.
func (r *SubgraphInstanceNode) BuiltGraph() *Graph {
	return r.current("")
}

func (r *SubgraphInstanceNode) LiveGraph() *Graph {
	live, _ := r.Live()
	return live
}

func (r *SubgraphInstanceNode) Live() (*Graph, error) {
	if r.ForEach() {
		return nil, fmt.Errorf("the placement runs once per item, so it has runs rather than one live graph")
	}
	signature, _ := r.signature()
	if current := r.current(signature); current != nil {
		return current, nil
	}
	return r.build(signature)
}

// An empty signature skips the fed-type check: settle leaves no copy fed something else.
func (r *SubgraphInstanceNode) current(signature string) *Graph {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.copied == nil || r.copied.version != r.runtime().version || (signature != "" && r.copied.signature != signature) {
		return nil
	}
	return r.copied.graph
}

func (r *SubgraphInstanceNode) build(signature string) (*Graph, error) {
	version := r.runtime().version
	graph, err := r.owner.copyDefinition(r, func(id, _ string) subgraph.External {
		return subgraphInstanceInputPort{node: r, id: id}
	})
	if err != nil {
		return nil, err
	}
	r.mu.Lock()
	r.copied = &copy{graph: graph, version: version, signature: signature}
	r.mu.Unlock()
	nodes.Touch()
	return graph, nil
}

// Called by settle once the inputs are held; a copy is built only when an
// array is fed where the definition has a single value.
func (r *SubgraphInstanceNode) refresh() error {
	if r.ForEach() {
		return nil
	}
	signature, lifted := r.signature()
	if r.current(signature) != nil {
		return nil
	}
	r.mu.Lock()
	r.copied = nil
	r.mu.Unlock()
	if !lifted {
		return nil
	}
	_, err := r.build(signature)
	return err
}

func (r *SubgraphInstanceNode) signature() (signature string, lifted bool) {
	definition := r.definition()
	r.mu.Lock()
	defer r.mu.Unlock()
	var parts []string
	for _, id := range slices.Sorted(maps.Keys(definition.boundaryNodes)) {
		boundary, ok := definition.inputPort(id)
		if !ok {
			continue
		}
		fed := fedType(boundary.BoundaryPortType(), r.inputs[id])
		lifted = lifted || fed != boundary.BoundaryPortType()
		parts = append(parts, id+":"+fed)
	}
	return strings.Join(parts, ";"), lifted
}

func (r *SubgraphInstanceNode) Inputs() map[string]nodes.InputPort {
	definition := r.definition()
	ports := make(map[string]nodes.InputPort, len(definition.boundaryNodes))
	for id := range definition.boundaryNodes {
		if boundary, ok := definition.inputPort(id); ok {
			ports[id] = subgraphInstanceInputPort{node: r, id: id, label: boundary.BoundaryPortName(), portType: boundary.BoundaryPortType()}
		}
	}
	return ports
}

func (r *SubgraphInstanceNode) Outputs() map[string]nodes.OutputPort {
	definition := r.definition()
	each := r.ForEach()
	ports := make(map[string]nodes.OutputPort, len(definition.boundaryNodes))
	for id := range definition.boundaryNodes {
		boundary, ok := definition.outputPort(id)
		if !ok {
			continue
		}
		if each {
			port := eachOutputPort{node: r, id: id, label: boundary.BoundaryPortName(), portType: boundary.BoundaryPortType()}
			ports[id] = port
			if builder, found := nodes.LookupPortTypeDynamicArray(r.itemType(id, port.portType)); found {
				ports[id] = builder.BuildDynamicArrayOutput(port)
			}
			continue
		}
		port := subgraphInstanceOutputPort{node: r, id: id, label: boundary.BoundaryPortName(), portType: boundary.BoundaryPortType()}
		ports[id] = port
		if builder, found := nodes.LookupPortTypeProxy(port.Type()); found {
			ports[id] = builder.BuildProxyOutput(port)
		}
	}
	return ports
}

// An input takes on the array of its type while an array feeds it.
func fedType(portType string, external nodes.OutputPort) string {
	if typed, ok := external.(nodes.Typed); ok && typed.Type() == subgraph.ArrayOf(portType) {
		return typed.Type()
	}
	return portType
}

func label(port nodes.Port) string {
	if l, ok := port.(nodes.Labeled); ok {
		return l.Label()
	}
	return port.Name()
}

// ============================================================================

type subgraphInstanceInputPort struct {
	node     *SubgraphInstanceNode
	id       string
	label    string
	portType string
}

func (p subgraphInstanceInputPort) Node() nodes.Node {
	return p.node
}

func (p subgraphInstanceInputPort) Name() string {
	return p.id
}

func (p subgraphInstanceInputPort) Label() string {
	return p.label
}

func (p subgraphInstanceInputPort) Type() string {
	return fedType(p.portType, p.Value())
}

// Run once per item, a boundary declared as a list also takes a list of lists.
func (p subgraphInstanceInputPort) AcceptedTypes() []string {
	if p.node.ForEach() {
		return []string{p.portType, subgraph.ArrayOf(p.portType)}
	}
	return subgraph.AcceptedBoundaryTypes(p.portType)
}

func (p subgraphInstanceInputPort) Value() nodes.OutputPort {
	p.node.mu.Lock()
	defer p.node.mu.Unlock()
	return p.node.inputs[p.id]
}

func (p subgraphInstanceInputPort) Set(port nodes.OutputPort) error {
	p.node.mu.Lock()
	p.node.inputs[p.id] = port
	p.node.mu.Unlock()
	nodes.Touch()
	return nil
}

func (p subgraphInstanceInputPort) Clear() {
	p.node.mu.Lock()
	delete(p.node.inputs, p.id)
	p.node.mu.Unlock()
	nodes.Touch()
}

// ============================================================================

// What a typed proxy handed to consumers reads through.
type subgraphInstanceOutputPort struct {
	node     *SubgraphInstanceNode
	id       string
	label    string
	portType string
}

func (p subgraphInstanceOutputPort) Node() nodes.Node {
	return p.node
}

func (p subgraphInstanceOutputPort) Name() string {
	return p.id
}

func (p subgraphInstanceOutputPort) Label() string {
	return p.label
}

func (p subgraphInstanceOutputPort) Type() string {
	graph := p.node.definition()
	if copied := p.node.current(""); copied != nil {
		graph = copied
	}
	if output, ok := graph.outputPort(p.id); ok {
		return cmp.Or(output.EffectiveType(), p.portType)
	}
	return p.portType
}

func (p subgraphInstanceOutputPort) InstanceType() string {
	return p.Type()
}

func (p subgraphInstanceOutputPort) Version() int {
	if source := p.CurrentSource(); source != nil {
		return source.Version()
	}
	return 0
}

// CurrentSource is nil until something has evaluated through the placement.
func (p subgraphInstanceOutputPort) CurrentSource() nodes.OutputPort {
	return p.readFrom(p.node.current(""))
}

func (p subgraphInstanceOutputPort) SourceForValue() nodes.OutputPort {
	return p.readFrom(p.node.LiveGraph())
}

func (p subgraphInstanceOutputPort) readFrom(copied *Graph) nodes.OutputPort {
	if copied == nil {
		return nil
	}
	if output, ok := copied.outputPort(p.id); ok {
		return output.ConnectedSource()
	}
	return nil
}
