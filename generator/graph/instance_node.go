package graph

import (
	"fmt"
	"strings"

	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
)

type instanceInputPort struct {
	subgraphNode *SubgraphInstanceNode
	portName     string
	portType     string
	external     nodes.OutputPort
}

func (p *instanceInputPort) Node() nodes.Node {
	return p.subgraphNode
}

func (p *instanceInputPort) Name() string {
	return p.portName
}

func (p *instanceInputPort) Type() string {
	return p.portType
}

func (p *instanceInputPort) Clear() {
	p.external = nil
	err := p.syncToBoundaryInput(nil)
	if err != nil {
		panic(err)
	}
}

func (p *instanceInputPort) Value() nodes.OutputPort {
	return p.external
}

func (p *instanceInputPort) Set(port nodes.OutputPort) error {
	p.external = port
	return p.syncToBoundaryInput(port)
}

func (p *instanceInputPort) syncToBoundaryInput(port nodes.OutputPort) error {
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

type instanceOutputPort struct {
	runtimeNode *SubgraphInstanceNode
	portName    string
	portType    string
}

func (p *instanceOutputPort) Node() nodes.Node {
	return p.runtimeNode
}

func (p *instanceOutputPort) Name() string {
	return p.portName
}

func (p *instanceOutputPort) Type() string {
	return p.portType
}

func (p *instanceOutputPort) Version() int {
	source := p.connectedSource()
	if source == nil {
		return 0
	}
	return source.Version()
}

func (p *instanceOutputPort) connectedSource() nodes.OutputPort {
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
	return r.syncInputPorts()
}

func (r *SubgraphInstanceNode) Outputs() map[string]nodes.OutputPort {
	return r.syncOutputPorts()
}

func (r *SubgraphInstanceNode) syncInputPorts() map[string]nodes.InputPort {
	if r.inputs == nil {
		r.inputs = make(map[string]nodes.InputPort)
	}

	boundaryPorts, err := r.owner.CollectBoundaryPorts(r.subGraphID)
	if err != nil {
		return r.inputs
	}

	active := make(map[string]struct{})
	for _, bp := range boundaryPorts {
		if bp.Kind != "input" {
			continue
		}
		name := bp.Name
		active[name] = struct{}{}

		port, ok := r.inputs[name].(*instanceInputPort)
		if !ok {
			port = &instanceInputPort{
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

func (r *SubgraphInstanceNode) syncOutputPorts() map[string]nodes.OutputPort {
	if r.outputs == nil {
		r.outputs = make(map[string]nodes.OutputPort)
	}

	boundaryPorts, err := r.owner.CollectBoundaryPorts(r.subGraphID)
	if err != nil {
		return r.outputs
	}

	active := make(map[string]struct{})
	for _, bp := range boundaryPorts {
		if bp.Kind != "output" {
			continue
		}
		name := bp.Name
		active[name] = struct{}{}

		port, ok := r.outputs[name].(*instanceOutputPort)
		if !ok {
			port = &instanceOutputPort{
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

func validateBoundaryPortName(child *Instance, portName string, kind string, excludeNode nodes.Node) error {
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
		existingKind := "output"
		if _, isInput := subgraph.IsInputBoundary(node); isInput {
			existingKind = "input"
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
		if err := validateBoundaryPortName(a, portName, "input", node); err != nil {
			return err
		}
		inputNode.PortName = portName
		inputNode.PortType = portType
		a.Root().onSubGraphChildMutation(a.SubGraphScopeID())
		return nil
	}

	if outputNode, ok := node.(*subgraph.OutputNode); ok {
		if err := validateBoundaryPortName(a, portName, "output", node); err != nil {
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
