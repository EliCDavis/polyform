package graph

import (
	"cmp"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
)

// SubgraphInstanceNode is a placement of a subgraph in another graph. Its
// ports are the definition's boundary nodes, keyed by their node id; the
// boundary's name is the port's label.
//
// Evaluating through it needs a private copy of the definition, derived
// from the definition's version and from what the placement is fed. The
// definition itself answers for the ports' types as long as every input is
// fed the rank it was wired for; an array arriving where the definition has
// a single value changes how the inside is typed, so that needs the copy.
type SubgraphInstanceNode struct {
	owner      *Instance
	subGraphID string

	mu     sync.Mutex
	inputs map[string]nodes.OutputPort
	copied *copy
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

func (r *SubgraphInstanceNode) runtime() subGraphRuntime {
	if definition, ok := r.owner.subGraphs[r.subGraphID]; ok {
		return *definition
	}
	return subGraphRuntime{name: "SubGraph"}
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

func (r *SubgraphInstanceNode) definition() boundaryIndex {
	if graph := r.runtime().instance; graph != nil {
		return graph.boundaries()
	}
	return boundaryIndex{}
}

// LiveGraph is this placement's copy of the definition, built if it has to
// be. It is nil when the definition cannot be copied.
func (r *SubgraphInstanceNode) LiveGraph() *Graph {
	live, _ := r.Live()
	return live
}

// Live is LiveGraph with the reason a copy could not be built.
func (r *SubgraphInstanceNode) Live() (*Graph, error) {
	if current := r.current(); current != nil {
		return current, nil
	}
	return r.build()
}

// current is the copy if it was built from the definition as it is now and
// what the placement is fed now.
func (r *SubgraphInstanceNode) current() *Graph {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.copied == nil || r.copied.version != r.runtime().version || r.copied.signature != r.signature() {
		return nil
	}
	return r.copied.graph
}

func (r *SubgraphInstanceNode) build() (*Graph, error) {
	version, signature := r.runtime().version, r.signature()
	graph, err := r.owner.copyDefinition(r)
	if err != nil {
		return nil, err
	}
	r.mu.Lock()
	r.copied = &copy{graph: graph, version: version, signature: signature}
	r.mu.Unlock()
	nodes.Touch()
	return graph, nil
}

// refresh builds the copy when the placement is fed something the
// definition is not wired for, so its output types can be read. settle
// calls it before asking for the outputs.
func (r *SubgraphInstanceNode) refresh() error {
	if !r.lifts() || r.current() != nil {
		return nil
	}
	_, err := r.build()
	return err
}

// The effective type of each input, which is what the inside's types
// depend on. Callers hold r.mu or are settle-time reads under the lock.
func (r *SubgraphInstanceNode) signature() string {
	definition := r.definition()
	var parts []string
	for _, id := range slices.Sorted(maps.Keys(definition.inputs)) {
		parts = append(parts, id+":"+fedType(definition.inputs[id].BoundaryPortType(), r.inputs[id]))
	}
	return strings.Join(parts, ";")
}

// lifts reports an input fed an array where the definition has a single value.
func (r *SubgraphInstanceNode) lifts() bool {
	definition := r.definition()
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, boundary := range definition.inputs {
		if fedType(boundary.BoundaryPortType(), r.inputs[id]) != boundary.BoundaryPortType() {
			return true
		}
	}
	return false
}

func (r *SubgraphInstanceNode) Inputs() map[string]nodes.InputPort {
	definition := r.definition()
	ports := make(map[string]nodes.InputPort, len(definition.inputs))
	for id, boundary := range definition.inputs {
		ports[id] = subgraphInstanceInputPort{node: r, id: id, label: boundary.BoundaryPortName(), portType: boundary.BoundaryPortType()}
	}
	return ports
}

func (r *SubgraphInstanceNode) Outputs() map[string]nodes.OutputPort {
	definition := r.definition()
	ports := make(map[string]nodes.OutputPort, len(definition.outputs))
	for id, boundary := range definition.outputs {
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

func (p subgraphInstanceInputPort) AcceptedTypes() []string {
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
	outputs := p.node.definition().outputs
	if copied := p.node.current(); copied != nil {
		outputs = copied.boundaries().outputs
	}
	if output, ok := outputs[p.id]; ok {
		return cmp.Or(output.EffectiveType(), p.portType)
	}
	return p.portType
}

func (p subgraphInstanceOutputPort) Version() int {
	if source := p.CurrentSource(); source != nil {
		return source.Version()
	}
	return 0
}

// CurrentSource is nil until something has evaluated through the placement.
func (p subgraphInstanceOutputPort) CurrentSource() nodes.OutputPort {
	return p.readFrom(p.node.current())
}

func (p subgraphInstanceOutputPort) SourceForValue() nodes.OutputPort {
	return p.readFrom(p.node.LiveGraph())
}

func (p subgraphInstanceOutputPort) readFrom(copied *Graph) nodes.OutputPort {
	if copied == nil {
		return nil
	}
	if output, ok := copied.boundaries().outputs[p.id]; ok {
		return output.ConnectedSource()
	}
	return nil
}
