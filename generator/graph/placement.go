package graph

import (
	"fmt"
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

	mu     sync.Mutex
	copied *placementCopy

	// Kept across refreshes: an edge holds the port object it was made with.
	inputs  map[string]*subgraphInstanceInputPort
	outputs map[string]*subgraphInstanceOutputPort
}

type placementCopy struct {
	graph *Graph
	boundaryIndex
}

func NewRuntimeNode(owner *Instance, subGraphID string) *SubgraphInstanceNode {
	return &SubgraphInstanceNode{
		owner:      owner,
		subGraphID: subGraphID,
		inputs:     make(map[string]*subgraphInstanceInputPort),
		outputs:    make(map[string]*subgraphInstanceOutputPort),
	}
}

func (r *SubgraphInstanceNode) SubGraphID() string {
	return r.subGraphID
}

func (r *SubgraphInstanceNode) Name() string {
	if definition, ok := r.owner.subGraphs[r.subGraphID]; ok {
		return definition.name
	}
	return "SubGraph"
}

func (r *SubgraphInstanceNode) Description() string {
	if definition, ok := r.owner.subGraphs[r.subGraphID]; ok {
		return definition.description
	}
	return ""
}

func (r *SubgraphInstanceNode) Path() string {
	return "SubGraph"
}

func (r *SubgraphInstanceNode) definition() boundaryIndex {
	if definition, ok := r.owner.subGraphs[r.subGraphID]; ok {
		return definition.instance.boundaries()
	}
	return boundaryIndex{}
}

// LiveGraph is this placement's copy of the definition, built if it has to be.
func (r *SubgraphInstanceNode) LiveGraph() *Graph {
	if copied := r.ensureCopy(); copied != nil {
		return copied.graph
	}
	return nil
}

// BuiltGraph is this placement's copy if anything has needed one yet.
func (r *SubgraphInstanceNode) BuiltGraph() *Graph {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.copied == nil {
		return nil
	}
	return r.copied.graph
}

func (r *SubgraphInstanceNode) ensureCopy() *placementCopy {
	r.mu.Lock()
	copied := r.copied
	r.mu.Unlock()
	if copied != nil {
		return copied
	}

	_ = r.build()
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.copied
}

func (r *SubgraphInstanceNode) build() error {
	r.mu.Lock()
	externals := make(map[string]nodes.OutputPort, len(r.inputs))
	for name, input := range r.inputs {
		if input.external != nil {
			externals[name] = input.external
		}
	}
	r.mu.Unlock()

	graph, err := r.owner.cloneSubGraphDefinition(r.subGraphID, externals)
	if err != nil {
		return err
	}

	r.mu.Lock()
	r.copied = &placementCopy{graph: graph, boundaryIndex: graph.boundaries()}
	r.mu.Unlock()
	nodes.Touch()
	return nil
}

// invalidate drops the copy, which no longer matches the definition, and
// returns what was feeding any input the definition no longer has.
func (r *SubgraphInstanceNode) invalidate() (lost map[string]nodes.OutputPort) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.copied = nil
	nodes.Touch()
	return r.refreshInputsLocked()
}

// needsCopy reports an input fed a different rank than the definition is
// wired for, which only a copy wired for it can type.
func (r *SubgraphInstanceNode) needsCopy() bool {
	definition := r.definition()

	r.mu.Lock()
	defer r.mu.Unlock()
	for name, input := range r.inputs {
		boundary, ok := definition.inputs[name]
		if ok && input.external != nil && input.Type() != boundaryRank(boundary) {
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

// feed has input read port. A source the inside cannot be wired around is
// refused, and the one before it put back.
func (r *SubgraphInstanceNode) feed(input *subgraphInstanceInputPort, port nodes.OutputPort) error {
	previous := input.Value()
	err := r.take(input, port)
	if err != nil && previous != port {
		if restoreErr := r.take(input, previous); restoreErr != nil {
			err = fmt.Errorf("%w (and the previous source could not be put back: %v)", err, restoreErr)
		}
	}
	return err
}

func (r *SubgraphInstanceNode) take(input *subgraphInstanceInputPort, port nodes.OutputPort) error {
	r.mu.Lock()
	input.external = port
	if r.copied != nil {
		// A copy stays only while it is still wired for what it is fed.
		boundary, ok := r.copied.inputs[input.portName]
		keep := false
		if ok {
			before := boundaryRank(boundary)
			boundary.SetExternalSource(port)
			keep = boundaryRank(boundary) == before
		}
		if !keep {
			r.copied = nil
		}
	}
	r.mu.Unlock()
	nodes.Touch()

	return r.buildIfNeeded()
}

func (r *SubgraphInstanceNode) Inputs() map[string]nodes.InputPort {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refreshInputsLocked()

	ports := make(map[string]nodes.InputPort, len(r.inputs))
	for name, port := range r.inputs {
		ports[name] = port
	}
	return ports
}

func (r *SubgraphInstanceNode) refreshInputsLocked() (lost map[string]nodes.OutputPort) {
	definition := r.definition()

	for name, boundary := range definition.inputs {
		port, ok := r.inputs[name]
		if !ok {
			port = &subgraphInstanceInputPort{subgraphNode: r, portName: name}
			r.inputs[name] = port
		}
		port.portType = boundary.BoundaryPortType()
	}

	for name, port := range r.inputs {
		if _, ok := definition.inputs[name]; ok {
			continue
		}
		if port.external != nil {
			if lost == nil {
				lost = make(map[string]nodes.OutputPort)
			}
			lost[name] = port.external
		}
		delete(r.inputs, name)
	}
	return lost
}

func (r *SubgraphInstanceNode) Outputs() map[string]nodes.OutputPort {
	definition := r.definition()

	r.mu.Lock()
	defer r.mu.Unlock()

	for name, boundary := range definition.outputs {
		port, ok := r.outputs[name]
		if !ok {
			port = &subgraphInstanceOutputPort{runtimeNode: r, portName: name}
			r.outputs[name] = port
		}
		port.portType = boundary.BoundaryPortType()

		// What consumers hold is rebuilt when an array starts or stops arriving.
		if effective := r.outputTypeLocked(name, definition); port.exposed == nil || port.exposedType != effective {
			port.exposed = port
			if builder, found := nodes.LookupPortTypeProxy(effective); found {
				port.exposed = builder.BuildProxyOutput(port)
			}
			port.exposedType = effective
		}
	}
	maps.DeleteFunc(r.outputs, func(name string, _ *subgraphInstanceOutputPort) bool {
		_, ok := definition.outputs[name]
		return !ok
	})

	ports := make(map[string]nodes.OutputPort, len(r.outputs))
	for name, port := range r.outputs {
		ports[name] = port.exposed
	}
	return ports
}

func (r *SubgraphInstanceNode) outputTypeLocked(name string, definition boundaryIndex) string {
	outputs := definition.outputs
	if r.copied != nil {
		outputs = r.copied.outputs
	}
	if output, ok := outputs[name]; ok {
		return output.EffectiveType()
	}
	return ""
}

// renameBoundaryPort keeps the port objects, which edges already hold.
func (r *SubgraphInstanceNode) renameBoundaryPort(oldName, newName string, kind BoundaryPortKind) {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch kind {
	case BoundaryPortKindInput:
		if port, ok := r.inputs[oldName]; ok {
			port.portName = newName
			delete(r.inputs, oldName)
			r.inputs[newName] = port
		}

	case BoundaryPortKindOutput:
		if port, ok := r.outputs[oldName]; ok {
			port.portName = newName
			delete(r.outputs, oldName)
			r.outputs[newName] = port
		}
	}
}

func boundaryRank(boundary subgraph.Boundary) string {
	if ranked, ok := boundary.(interface{ EffectiveType() string }); ok {
		return ranked.EffectiveType()
	}
	return boundary.BoundaryPortType()
}

// ============================================================================

type subgraphInstanceInputPort struct {
	subgraphNode *SubgraphInstanceNode
	portName     string
	portType     string
	external     nodes.OutputPort
}

func (p *subgraphInstanceInputPort) Node() nodes.Node {
	return p.subgraphNode
}

func (p *subgraphInstanceInputPort) Name() string {
	return p.portName
}

func (p *subgraphInstanceInputPort) Type() string {
	if typed, ok := p.external.(nodes.Typed); ok {
		if name := typed.Type(); name == subgraph.ArrayOf(p.portType) {
			return name
		}
	}
	return p.portType
}

func (p *subgraphInstanceInputPort) AcceptedTypes() []string {
	return subgraph.AcceptedBoundaryTypes(p.portType)
}

func (p *subgraphInstanceInputPort) Value() nodes.OutputPort {
	p.subgraphNode.mu.Lock()
	defer p.subgraphNode.mu.Unlock()
	return p.external
}

func (p *subgraphInstanceInputPort) Set(port nodes.OutputPort) error {
	return p.subgraphNode.feed(p, port)
}

func (p *subgraphInstanceInputPort) Clear() {
	if err := p.subgraphNode.feed(p, nil); err != nil {
		panic(err)
	}
}

// ============================================================================

// What a typed proxy handed to consumers reads through.
type subgraphInstanceOutputPort struct {
	runtimeNode *SubgraphInstanceNode
	portName    string
	portType    string

	exposed     nodes.OutputPort
	exposedType string
}

func (p *subgraphInstanceOutputPort) Node() nodes.Node {
	return p.runtimeNode
}

func (p *subgraphInstanceOutputPort) Name() string {
	return p.portName
}

func (p *subgraphInstanceOutputPort) Type() string {
	definition := p.runtimeNode.definition()

	p.runtimeNode.mu.Lock()
	defer p.runtimeNode.mu.Unlock()
	if effective := p.runtimeNode.outputTypeLocked(p.portName, definition); effective != "" {
		return effective
	}
	return p.portType
}

func (p *subgraphInstanceOutputPort) Version() int {
	if source := p.CurrentSource(); source != nil {
		return source.Version()
	}
	return 0
}

// CurrentSource is nil until something has evaluated through the placement.
func (p *subgraphInstanceOutputPort) CurrentSource() nodes.OutputPort {
	p.runtimeNode.mu.Lock()
	copied := p.runtimeNode.copied
	p.runtimeNode.mu.Unlock()
	return copied.source(p.portName)
}

func (p *subgraphInstanceOutputPort) SourceForValue() nodes.OutputPort {
	return p.runtimeNode.ensureCopy().source(p.portName)
}

func (c *placementCopy) source(output string) nodes.OutputPort {
	if c == nil {
		return nil
	}
	if node, ok := c.outputs[output]; ok {
		return node.ConnectedSource()
	}
	return nil
}
