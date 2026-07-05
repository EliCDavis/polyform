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
	err := p.syncToBoundaryInput(nil)
	if err != nil {
		panic(err)
	}
}

func (p *subgraphInstanceInputPort) Value() nodes.OutputPort {
	return p.external
}

func (p *subgraphInstanceInputPort) Set(port nodes.OutputPort) error {
	p.external = port
	return p.syncToBoundaryInput(port)
}

func (p *subgraphInstanceInputPort) syncToBoundaryInput(port nodes.OutputPort) error {
	child, err := p.subgraphNode.owner.SubGraphInstance(p.subgraphNode.subGraphID)
	if err != nil {
		return fmt.Errorf("failed to get subgraph instance: %w", err)
	}
	for node := range child.nodeIDs {
		inputBoundary, ok := subgraph.IsInputBoundary(node)
		if !ok {
			continue
		}
		if inputBoundary.BoundaryPortName() == p.portName {
			inputBoundary.SetExternalSource(port)
			return nil
		}
	}

	return fmt.Errorf("boundary input port %q not found", p.portName)
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

func (p *subgraphInstanceOutputPort) connectedSource() nodes.OutputPort {
	child, err := p.runtimeNode.owner.SubGraphInstance(p.runtimeNode.subGraphID)
	if err != nil {
		return nil
	}
	for node := range child.nodeIDs {
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

type SubgraphInstanceNode struct {
	owner      *Instance
	subGraphID string
	inputs     map[string]nodes.InputPort
	outputs    map[string]nodes.OutputPort
}

func NewRuntimeNode(owner *Instance, subGraphID string) *SubgraphInstanceNode {
	return &SubgraphInstanceNode{
		owner:      owner,
		subGraphID: subGraphID,
	}
}

func (r *SubgraphInstanceNode) SubGraphID() string {
	return r.subGraphID
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

		port, ok := r.outputs[name].(*subgraphInstanceOutputPort)
		if !ok {
			port = &subgraphInstanceOutputPort{
				runtimeNode: r,
				portName:    name,
			}
			r.outputs[name] = port
		}
		port.portType = bp.Type
	}

	for name := range r.outputs {
		if _, ok := active[name]; !ok {
			delete(r.outputs, name)
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
		existingKind := BoundaryPortKindOutput
		if _, isInput := subgraph.IsInputBoundary(node); isInput {
			existingKind = BoundaryPortKindInput
		}
		if existingKind == kind {
			return fmt.Errorf("boundary port name %q already used by another %s node", portName, kind)
		}
	}
	return nil
}

func validateBoundaryPortInfo(portName, portType string) error {
	if strings.TrimSpace(portType) == "" {
		return fmt.Errorf("boundary port type is required")
	}
	if strings.TrimSpace(portName) == "" {
		return fmt.Errorf("boundary port name is required")
	}
	return nil
}

func (a *Instance) SetBoundaryNodeInfo(nodeID, portName, portType string) error {
	if err := validateBoundaryPortInfo(portName, portType); err != nil {
		return err
	}

	node := a.Node(nodeID)

	if inputNode, ok := node.(*subgraph.InputNode); ok {
		if err := validateBoundaryPortName(a, portName, BoundaryPortKindInput, node); err != nil {
			return err
		}
		inputNode.PortName = portName
		inputNode.PortType = portType
		a.Root().onSubGraphChildMutation(a.SubGraphScopeID())
		return nil
	}

	if outputNode, ok := node.(*subgraph.OutputNode); ok {
		if err := validateBoundaryPortName(a, portName, BoundaryPortKindOutput, node); err != nil {
			return err
		}
		outputNode.PortName = portName
		outputNode.PortType = portType
		a.Root().onSubGraphChildMutation(a.SubGraphScopeID())
		return nil
	}

	return fmt.Errorf("node %q is not a sub-graph boundary node", nodeID)
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
