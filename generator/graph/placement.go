package graph

import (
	"cmp"
	"maps"
	"sync"

	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
)

// SubgraphInstanceNode is a placement of a subgraph in another graph.
//
// Evaluating through it needs a private copy of the definition, which is
// built the first time a value is asked for. Until then the placement
// answers for its ports and their types from the definition itself, which
// holds as long as every input is fed the rank the definition was wired for.
// An array arriving where the definition has a single value changes how the
// inside is typed, so that builds the copy straight away.
type SubgraphInstanceNode struct {
	owner      *Instance
	subGraphID string

	mu        sync.Mutex
	copied    *Graph
	externals map[string]nodes.OutputPort
}

func NewRuntimeNode(owner *Instance, subGraphID string) *SubgraphInstanceNode {
	return &SubgraphInstanceNode{
		owner:      owner,
		subGraphID: subGraphID,
		externals:  make(map[string]nodes.OutputPort),
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

// BuiltGraph is this placement's copy if anything has needed one yet.
func (r *SubgraphInstanceNode) BuiltGraph() *Graph {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.copied
}

// LiveGraph is this placement's copy of the definition, built if it has to
// be. It is nil when the definition cannot be copied.
func (r *SubgraphInstanceNode) LiveGraph() *Graph {
	if r.BuiltGraph() == nil {
		_ = r.build()
	}
	return r.BuiltGraph()
}

func (r *SubgraphInstanceNode) build() error {
	r.mu.Lock()
	externals := maps.Clone(r.externals)
	r.mu.Unlock()

	graph, err := r.owner.cloneSubGraphDefinition(r.subGraphID, externals)
	if err != nil {
		return err
	}

	r.mu.Lock()
	r.copied = graph
	r.mu.Unlock()
	nodes.Touch()
	return nil
}

// invalidate drops the copy, which no longer matches the definition.
func (r *SubgraphInstanceNode) invalidate() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.copied = nil
	nodes.Touch()
}

// needsCopy reports an input fed a different rank than the definition is
// wired for, which only a copy wired for it can type.
func (r *SubgraphInstanceNode) needsCopy() bool {
	definition := r.definition()

	r.mu.Lock()
	defer r.mu.Unlock()
	for name, external := range r.externals {
		boundary, ok := definition.inputs[name]
		if ok && fedType(boundary.BoundaryPortType(), external) != boundaryRank(boundary) {
			return true
		}
	}
	return false
}

func (r *SubgraphInstanceNode) buildIfNeeded() error {
	if r.BuiltGraph() == nil && r.needsCopy() {
		return r.build()
	}
	return nil
}

// feed has an input read port, or nothing when port is nil. A copy stays
// only while it is still wired for what it is fed.
func (r *SubgraphInstanceNode) feed(input string, port nodes.OutputPort) error {
	r.mu.Lock()
	if port == nil {
		delete(r.externals, input)
	} else {
		r.externals[input] = port
	}
	if r.copied != nil {
		boundary, ok := r.copied.boundaries().inputs[input]
		if ok {
			before := boundaryRank(boundary)
			boundary.SetExternalSource(port)
			ok = boundaryRank(boundary) == before
		}
		if !ok {
			r.copied = nil
		}
	}
	r.mu.Unlock()
	nodes.Touch()

	return r.buildIfNeeded()
}

func (r *SubgraphInstanceNode) Inputs() map[string]nodes.InputPort {
	definition := r.definition()

	r.mu.Lock()
	maps.DeleteFunc(r.externals, func(name string, _ nodes.OutputPort) bool {
		_, ok := definition.inputs[name]
		return !ok
	})
	r.mu.Unlock()

	ports := make(map[string]nodes.InputPort, len(definition.inputs))
	for name, boundary := range definition.inputs {
		ports[name] = subgraphInstanceInputPort{node: r, name: name, portType: boundary.BoundaryPortType()}
	}
	return ports
}

func (r *SubgraphInstanceNode) Outputs() map[string]nodes.OutputPort {
	definition := r.definition()

	ports := make(map[string]nodes.OutputPort, len(definition.outputs))
	for name, boundary := range definition.outputs {
		port := subgraphInstanceOutputPort{node: r, name: name, portType: boundary.BoundaryPortType()}
		ports[name] = port
		if builder, found := nodes.LookupPortTypeProxy(port.Type()); found {
			ports[name] = builder.BuildProxyOutput(port)
		}
	}
	return ports
}

func boundaryRank(boundary subgraph.Boundary) string {
	if ranked, ok := boundary.(interface{ EffectiveType() string }); ok {
		return ranked.EffectiveType()
	}
	return boundary.BoundaryPortType()
}

// An input takes on the array of its type while an array feeds it.
func fedType(portType string, external nodes.OutputPort) string {
	if typed, ok := external.(nodes.Typed); ok && typed.Type() == subgraph.ArrayOf(portType) {
		return typed.Type()
	}
	return portType
}

// ============================================================================

type subgraphInstanceInputPort struct {
	node     *SubgraphInstanceNode
	name     string
	portType string
}

func (p subgraphInstanceInputPort) Node() nodes.Node {
	return p.node
}

func (p subgraphInstanceInputPort) Name() string {
	return p.name
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
	return p.node.externals[p.name]
}

func (p subgraphInstanceInputPort) Set(port nodes.OutputPort) error {
	return p.node.feed(p.name, port)
}

func (p subgraphInstanceInputPort) Clear() {
	if err := p.node.feed(p.name, nil); err != nil {
		panic(err)
	}
}

// ============================================================================

// What a typed proxy handed to consumers reads through.
type subgraphInstanceOutputPort struct {
	node     *SubgraphInstanceNode
	name     string
	portType string
}

func (p subgraphInstanceOutputPort) Node() nodes.Node {
	return p.node
}

func (p subgraphInstanceOutputPort) Name() string {
	return p.name
}

func (p subgraphInstanceOutputPort) Type() string {
	outputs := p.node.definition().outputs
	if copied := p.node.BuiltGraph(); copied != nil {
		outputs = copied.boundaries().outputs
	}
	if output, ok := outputs[p.name]; ok {
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
	return p.readFrom(p.node.BuiltGraph())
}

func (p subgraphInstanceOutputPort) SourceForValue() nodes.OutputPort {
	return p.readFrom(p.node.LiveGraph())
}

func (p subgraphInstanceOutputPort) readFrom(copied *Graph) nodes.OutputPort {
	if copied == nil {
		return nil
	}
	if output, ok := copied.boundaries().outputs[p.name]; ok {
		return output.ConnectedSource()
	}
	return nil
}
