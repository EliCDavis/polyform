package graph

import (
	"fmt"
	"strings"

	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
)

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
	return p.portType
}

func (p *subgraphInstanceInputPort) Clear() {
	p.external = nil
	_ = p.subgraphNode.syncInputToClone(p.portName, nil)
}

func (p *subgraphInstanceInputPort) Value() nodes.OutputPort {
	return p.external
}

func (p *subgraphInstanceInputPort) Set(port nodes.OutputPort) error {
	p.external = port
	return p.subgraphNode.syncInputToClone(p.portName, port)
}

type subgraphInstanceOutputPort struct {
	runtimeNode *SubgraphInstanceNode
	portName    string
	portType    string
}

func (p *subgraphInstanceOutputPort) Node() nodes.Node {
	return p.runtimeNode
}

func (p *subgraphInstanceOutputPort) Name() string {
	return p.portName
}

func (p *subgraphInstanceOutputPort) Type() string {
	return p.portType
}

func (p *subgraphInstanceOutputPort) Version() int {
	source := p.connectedSource()
	if source == nil {
		return 0
	}
	return source.Version()
}

func (p *subgraphInstanceOutputPort) CurrentSource() nodes.OutputPort {
	return p.connectedSource()
}

func (p *subgraphInstanceOutputPort) connectedSource() nodes.OutputPort {
	clone := p.runtimeNode.ensureClone()
	if clone == nil {
		return nil
	}
	for node := range clone.nodeIDs {
		outputNode, ok := node.(*subgraph.OutputNode)
		if !ok {
			continue
		}
		if outputNode.BoundaryPortName() == p.portName {
			return outputNode.ConnectedSource()
		}
	}
	return nil
}

// SubgraphInstanceNode is a placement of a sub-graph on a parent graph. Each
// placement owns a private clone of the definition so evaluation is isolated,
// while edits to the shared definition rebuild every clone.
type SubgraphInstanceNode struct {
	owner         *Instance
	subGraphID    string
	clone         *Instance
	inputs        map[string]nodes.InputPort
	outputs       map[string]nodes.OutputPort
	outputSources map[string]*subgraphInstanceOutputPort
}

func NewRuntimeNode(owner *Instance, subGraphID string) *SubgraphInstanceNode {
	n := &SubgraphInstanceNode{
		owner:      owner,
		subGraphID: subGraphID,
	}
	_ = n.rebuildClone()
	return n
}

func (r *SubgraphInstanceNode) SubGraphID() string {
	return r.subGraphID
}

func (r *SubgraphInstanceNode) ensureClone() *Instance {
	if r.clone == nil {
		_ = r.rebuildClone()
	}
	return r.clone
}

func (r *SubgraphInstanceNode) rebuildClone() error {
	externals := r.snapshotExternals()

	clone, err := r.owner.Root().cloneSubGraphDefinition(r.subGraphID)
	if err != nil {
		return err
	}
	r.clone = clone

	if err := r.applyExternals(externals); err != nil {
		return err
	}
	_ = r.Outputs()
	return nil
}

// renameBoundaryPort: boundary rename. Move map key first. Keep same port
// object so wires still hold it.
func (r *SubgraphInstanceNode) renameBoundaryPort(oldName, newName string, kind BoundaryPortKind) {
	if oldName == "" || oldName == newName {
		return
	}

	switch kind {
	case BoundaryPortKindInput:
		if r.inputs == nil {
			return
		}
		port, ok := r.inputs[oldName]
		if !ok {
			return
		}
		if sip, ok := port.(*subgraphInstanceInputPort); ok {
			sip.portName = newName
		}
		delete(r.inputs, oldName)
		r.inputs[newName] = port

	case BoundaryPortKindOutput:
		if r.outputSources != nil {
			if src, ok := r.outputSources[oldName]; ok {
				src.portName = newName
				delete(r.outputSources, oldName)
				r.outputSources[newName] = src
			}
		}
		if r.outputs != nil {
			if out, ok := r.outputs[oldName]; ok {
				delete(r.outputs, oldName)
				r.outputs[newName] = out
			}
		}
	}
}

func (r *SubgraphInstanceNode) snapshotExternals() map[string]nodes.OutputPort {
	if r.inputs == nil {
		return nil
	}
	out := make(map[string]nodes.OutputPort, len(r.inputs))
	for name, input := range r.inputs {
		sip, ok := input.(*subgraphInstanceInputPort)
		if !ok || sip.external == nil {
			continue
		}
		out[name] = sip.external
	}
	return out
}

func (r *SubgraphInstanceNode) applyExternals(externals map[string]nodes.OutputPort) error {
	r.Inputs() // refresh port map against current definition boundaries
	for name, port := range externals {
		if err := r.syncInputToClone(name, port); err != nil {
			// Port may have been removed from the definition; drop the wire.
			if sip, ok := r.inputs[name].(*subgraphInstanceInputPort); ok {
				sip.external = nil
			}
			continue
		}
		if sip, ok := r.inputs[name].(*subgraphInstanceInputPort); ok {
			sip.external = port
		}
	}
	return nil
}

func (r *SubgraphInstanceNode) syncInputToClone(portName string, port nodes.OutputPort) error {
	clone := r.ensureClone()
	if clone == nil {
		return fmt.Errorf("sub-graph %q has no clone", r.subGraphID)
	}

	for node := range clone.nodeIDs {
		inputBoundary, ok := subgraph.IsInputBoundary(node)
		if !ok || inputBoundary.BoundaryPortName() != portName {
			continue
		}
		inputBoundary.SetExternalSource(port)
		return nil
	}
	return fmt.Errorf("boundary input port %q not found in clone of %q", portName, r.subGraphID)
}

func (r *SubgraphInstanceNode) Name() string {
	runtime, ok := r.owner.subGraphs[r.subGraphID]
	if !ok {
		return "SubGraph"
	}
	return runtime.name
}

func (r *SubgraphInstanceNode) Path() string {
	return "SubGraph"
}

func (r *SubgraphInstanceNode) Inputs() map[string]nodes.InputPort {
	if r.inputs == nil {
		r.inputs = make(map[string]nodes.InputPort)
	}

	boundaryPorts, err := r.owner.CollectBoundaryPorts(r.subGraphID)
	if err != nil {
		return r.inputs
	}

	active := make(map[string]struct{})
	for _, bp := range boundaryPorts {
		if bp.Kind != BoundaryPortKindInput {
			continue
		}
		name := bp.Name
		active[name] = struct{}{}

		port, ok := r.inputs[name].(*subgraphInstanceInputPort)
		if !ok {
			port = &subgraphInstanceInputPort{
				subgraphNode: r,
				portName:     name,
			}
			r.inputs[name] = port
		}
		port.portType = bp.Type
	}

	for name := range r.inputs {
		if _, ok := active[name]; !ok {
			delete(r.inputs, name)
		}
	}

	return r.inputs
}

func (r *SubgraphInstanceNode) Outputs() map[string]nodes.OutputPort {
	if r.outputs == nil {
		r.outputs = make(map[string]nodes.OutputPort)
	}
	if r.outputSources == nil {
		r.outputSources = make(map[string]*subgraphInstanceOutputPort)
	}

	boundaryPorts, err := r.owner.CollectBoundaryPorts(r.subGraphID)
	if err != nil {
		return r.outputs
	}

	active := make(map[string]struct{})
	for _, bp := range boundaryPorts {
		if bp.Kind != BoundaryPortKindOutput {
			continue
		}
		name := bp.Name
		active[name] = struct{}{}

		port, ok := r.outputSources[name]
		if !ok {
			port = &subgraphInstanceOutputPort{
				runtimeNode: r,
				portName:    name,
			}
			r.outputSources[name] = port
		}
		typeChanged := port.portType != bp.Type
		port.portType = bp.Type

		// Expose the boundary as a strongly typed output when the port type
		// is known, so downstream typed inputs can connect to it.
		if _, ok := r.outputs[name]; !ok || typeChanged {
			var exposed nodes.OutputPort = port
			if builder, found := subgraph.LookupPortTypeProxy(bp.Type); found {
				exposed = builder.BuildProxyOutput(port)
			}
			r.outputs[name] = exposed
		}
	}

	for name := range r.outputs {
		if _, ok := active[name]; !ok {
			delete(r.outputs, name)
			delete(r.outputSources, name)
		}
	}

	return r.outputs
}

func (r *SubgraphInstanceNode) Description() string {
	runtime, ok := r.owner.subGraphs[r.subGraphID]
	if !ok {
		return ""
	}
	return runtime.description
}

func validateBoundaryPortName(child *Instance, portName string, kind BoundaryPortKind, excludeNode nodes.Node) error {
	for node := range child.nodeIDs {
		if node == excludeNode {
			continue
		}
		boundary, ok := subgraph.IsBoundaryNode(node)
		if !ok {
			continue
		}
		if boundary.BoundaryPortName() != portName {
			continue
		}
		if GetBoundaryKind(boundary) == kind {
			return fmt.Errorf("boundary port name %q already used by another %s node", portName, kind)
		}
	}
	return nil
}

func validateBoundaryPortNameOnly(portName string) error {
	if strings.TrimSpace(portName) == "" {
		return fmt.Errorf("boundary port name is required")
	}
	return nil
}

func (a *Instance) SetBoundaryNodeInfo(nodeID, portName string) error {
	if err := validateBoundaryPortNameOnly(portName); err != nil {
		return err
	}

	node := a.Node(nodeID)

	if inputNode, ok := node.(*subgraph.InputNode); ok {
		oldName := inputNode.PortName
		if err := validateBoundaryPortName(a, portName, BoundaryPortKindInput, node); err != nil {
			return err
		}
		inputNode.PortName = portName
		a.migrateRuntimeBoundaryPortName(oldName, portName, BoundaryPortKindInput)
		return a.notifyDefinitionMutation()
	}

	if outputNode, ok := node.(*subgraph.OutputNode); ok {
		oldName := outputNode.PortName
		if err := validateBoundaryPortName(a, portName, BoundaryPortKindOutput, node); err != nil {
			return err
		}
		outputNode.PortName = portName
		a.migrateRuntimeBoundaryPortName(oldName, portName, BoundaryPortKindOutput)
		return a.notifyDefinitionMutation()
	}

	return fmt.Errorf("node %q is not a sub-graph boundary node", nodeID)
}

// migrateRuntimeBoundaryPortName rewrites port keys on every live placement of
// this definition before clones rebuild, so AssignedInput / externals keyed by
// the old name are not dropped as "removed ports".
func (a *Instance) migrateRuntimeBoundaryPortName(oldName, newName string, kind BoundaryPortKind) {
	if oldName == "" || oldName == newName {
		return
	}
	subGraphID := a.SubGraphScopeID()
	if subGraphID == "" {
		return
	}
	forEachSubGraphInstance(a.Root(), subGraphID, func(runtime *SubgraphInstanceNode) {
		runtime.renameBoundaryPort(oldName, newName, kind)
	})
}

func (a *Instance) SubGraphScopeID() string {
	if a.parent == nil {
		return ""
	}
	for id, runtime := range a.parent.subGraphs {
		if runtime.instance == a {
			return id
		}
	}
	return ""
}
