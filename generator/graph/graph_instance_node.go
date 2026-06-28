package graph

import (
	"fmt"
	"strings"

	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
)

type graphInstanceInputPort struct {
	runtimeNode *GraphInstanceNode
	portName    string
	portType    string
	external    nodes.OutputPort
}

func (p *graphInstanceInputPort) Node() nodes.Node {
	return p.runtimeNode
}

func (p *graphInstanceInputPort) Name() string {
	return p.portName
}

func (p *graphInstanceInputPort) Type() string {
	return p.portType
}

func (p *graphInstanceInputPort) Clear() {
	p.external = nil
	p.syncToBoundaryInput(nil)
}

func (p *graphInstanceInputPort) Value() nodes.OutputPort {
	return p.external
}

func (p *graphInstanceInputPort) Set(port nodes.OutputPort) error {
	p.external = port
	p.syncToBoundaryInput(port)
	return nil
}

func (p *graphInstanceInputPort) syncToBoundaryInput(port nodes.OutputPort) {
	child, err := p.runtimeNode.owner.SubGraphInstance(p.runtimeNode.subGraphID)
	if err != nil {
		return
	}
	for node := range child.nodeIDs {
		inputBoundary, ok := subgraph.IsInputBoundary(node)
		if !ok {
			continue
		}
		if inputBoundary.BoundaryPortName() == p.portName {
			inputBoundary.SetExternalSource(port)
			return
		}
	}
}

type graphInstanceOutputPort struct {
	runtimeNode *GraphInstanceNode
	portName    string
	portType    string
}

func (p *graphInstanceOutputPort) Node() nodes.Node {
	return p.runtimeNode
}

func (p *graphInstanceOutputPort) Name() string {
	return p.portName
}

func (p *graphInstanceOutputPort) Type() string {
	return p.portType
}

func (p *graphInstanceOutputPort) Version() int {
	source := p.connectedSource()
	if source == nil {
		return 0
	}
	return source.Version()
}

func (p *graphInstanceOutputPort) connectedSource() nodes.OutputPort {
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

type GraphInstanceNode struct {
	owner      *Instance
	subGraphID string
}

func NewRuntimeNode(owner *Instance, subGraphID string) *GraphInstanceNode {
	return &GraphInstanceNode{
		owner:      owner,
		subGraphID: subGraphID,
	}
}

func (r *GraphInstanceNode) SubGraphID() string {
	return r.subGraphID
}

func (r *GraphInstanceNode) Name() string {
	runtime, ok := r.owner.subGraphs[r.subGraphID]
	if !ok {
		return "SubGraph"
	}
	return runtime.name
}

func (r *GraphInstanceNode) Path() string {
	return "SubGraph"
}

func (r *GraphInstanceNode) Inputs() map[string]nodes.InputPort {
	ports := make(map[string]nodes.InputPort)
	boundaryPorts, err := r.owner.CollectBoundaryPorts(r.subGraphID)
	if err != nil {
		return ports
	}
	for _, bp := range boundaryPorts {
		if bp.Kind != "input" {
			continue
		}
		name := bp.Name
		ports[name] = &graphInstanceInputPort{
			runtimeNode: r,
			portName:    name,
			portType:    bp.Type,
		}
	}
	return ports
}

func (r *GraphInstanceNode) Outputs() map[string]nodes.OutputPort {
	ports := make(map[string]nodes.OutputPort)
	boundaryPorts, err := r.owner.CollectBoundaryPorts(r.subGraphID)
	if err != nil {
		return ports
	}
	for _, bp := range boundaryPorts {
		if bp.Kind != "output" {
			continue
		}
		name := bp.Name
		ports[name] = &graphInstanceOutputPort{
			runtimeNode: r,
			portName:    name,
			portType:    bp.Type,
		}
	}
	return ports
}

func (r *GraphInstanceNode) Description() string {
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
		a.Root().onSubGraphChildMutation(a.subGraphScopeID())
		return nil
	}

	if outputNode, ok := node.(*subgraph.OutputNode); ok {
		if err := validateBoundaryPortName(a, portName, "output", node); err != nil {
			return err
		}
		outputNode.PortName = portName
		outputNode.PortType = portType
		a.Root().onSubGraphChildMutation(a.subGraphScopeID())
		return nil
	}

	return fmt.Errorf("node %q is not a sub-graph boundary node", nodeID)
}

func (a *Instance) SubGraphScopeID() string {
	return a.subGraphScopeID()
}

func (a *Instance) subGraphScopeID() string {
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
