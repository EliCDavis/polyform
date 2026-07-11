package subgraph

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/nodes"
)

const (
	ValuePortName = "Value"
)

type Boundary interface {
	nodes.Node
	BoundaryPortName() string
	BoundaryPortType() string
}

type InputBoundary interface {
	Boundary
	SetExternalSource(port nodes.OutputPort)
	ExternalSource() nodes.OutputPort
}

func IsBoundaryNode(node nodes.Node) (Boundary, bool) {
	b, ok := node.(Boundary)
	return b, ok
}

func IsInputBoundary(node nodes.Node) (InputBoundary, bool) {
	b, ok := node.(InputBoundary)
	return b, ok
}

type boundaryData struct {
	PortName string `json:"portName"`
	PortType string `json:"portType"`
}

type InputNode struct {
	PortName string
	PortType string

	externalSource nodes.OutputPort
	version        int
}

func NewInputNode(portName, portType string) *InputNode {
	return &InputNode{
		PortName: portName,
		PortType: portType,
	}
}

func NewOutputNode(portName, portType string) *OutputNode {
	n := &OutputNode{
		PortName: portName,
		PortType: portType,
	}
	n.inputPort = &outputNodeInputPort{node: n}
	return n
}

func ConfigureBoundaryPortType(node nodes.Node, portType string) error {
	portType = strings.TrimSpace(portType)
	if portType == "" {
		return fmt.Errorf("boundary port type is required")
	}
	switch n := node.(type) {
	case *InputNode:
		if n.PortType != "" && n.PortType != portType {
			return fmt.Errorf("boundary port type cannot be changed")
		}
		n.PortType = portType
	case *OutputNode:
		if n.PortType != "" && n.PortType != portType {
			return fmt.Errorf("boundary port type cannot be changed")
		}
		n.PortType = portType
	default:
		return fmt.Errorf("node is not a sub-graph boundary node")
	}
	return nil
}

func BoundaryPortNameConfigured(node nodes.Node) bool {
	switch n := node.(type) {
	case *InputNode:
		return strings.TrimSpace(n.PortName) != ""
	case *OutputNode:
		return strings.TrimSpace(n.PortName) != ""
	default:
		return false
	}
}

func (n *InputNode) BoundaryPortName() string {
	if n.PortName == "" {
		return "Input"
	}
	return n.PortName
}

func (n *InputNode) BoundaryPortType() string {
	return n.PortType
}

func (n *InputNode) Name() string {
	return n.BoundaryPortName()
}

func (n *InputNode) SetExternalSource(port nodes.OutputPort) {
	n.externalSource = port
	n.version++
}

func (n *InputNode) ExternalSource() nodes.OutputPort {
	return n.externalSource
}

func (n *InputNode) Inputs() map[string]nodes.InputPort {
	return nil
}

func (n *InputNode) Outputs() map[string]nodes.OutputPort {
	return map[string]nodes.OutputPort{
		ValuePortName: buildInputOutputPort(n),
	}
}

func (n *InputNode) ToJSON(encoder *jbtf.Encoder) ([]byte, error) {
	return json.Marshal(boundaryData{
		PortName: n.PortName,
		PortType: n.PortType,
	})
}

func (n *InputNode) FromJSON(decoder jbtf.Decoder, body []byte) error {
	var data boundaryData
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}
	n.PortName = data.PortName
	n.PortType = data.PortType
	return nil
}

// buildInputOutputPort exposes the boundary's value as a strongly typed
// output when its port type has been discovered, falling back to an untyped
// port otherwise.
func buildInputOutputPort(n *InputNode) nodes.OutputPort {
	source := &inputNodeOutputPort{node: n}
	if builder, ok := LookupPortTypeProxy(n.PortType); ok {
		return builder.BuildProxyOutput(source)
	}
	return source
}

// inputNodeOutputPort is the untyped fallback port for an input boundary. It
// also acts as the nodes.ProxySource that typed proxy ports forward to.
type inputNodeOutputPort struct {
	node *InputNode
}

func (p *inputNodeOutputPort) Node() nodes.Node {
	return p.node
}

func (p *inputNodeOutputPort) Name() string {
	return ValuePortName
}

func (p *inputNodeOutputPort) Type() string {
	return p.node.PortType
}

func (p *inputNodeOutputPort) Version() int {
	if p.node.externalSource != nil {
		return p.node.externalSource.Version()
	}
	return p.node.version
}

func (p *inputNodeOutputPort) CurrentSource() nodes.OutputPort {
	return p.node.externalSource
}

type OutputNode struct {
	PortName string
	PortType string

	inputPort *outputNodeInputPort
}

func (n *OutputNode) BoundaryPortName() string {
	if n.PortName == "" {
		return "Output"
	}
	return n.PortName
}

func (n *OutputNode) BoundaryPortType() string {
	return n.PortType
}

func (n *OutputNode) Name() string {
	return n.BoundaryPortName()
}

func (n *OutputNode) Inputs() map[string]nodes.InputPort {
	return map[string]nodes.InputPort{
		ValuePortName: n.inputPort,
	}
}

func (n *OutputNode) Outputs() map[string]nodes.OutputPort {
	return nil
}

func (n *OutputNode) ConnectedSource() nodes.OutputPort {
	return n.inputPort.Value()
}

func (n *OutputNode) ToJSON(encoder *jbtf.Encoder) ([]byte, error) {
	return json.Marshal(boundaryData{
		PortName: n.PortName,
		PortType: n.PortType,
	})
}

func (n *OutputNode) FromJSON(decoder jbtf.Decoder, body []byte) error {
	var data boundaryData
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}
	n.PortName = data.PortName
	n.PortType = data.PortType
	return nil
}

type outputNodeInputPort struct {
	node     *OutputNode
	connected nodes.OutputPort
}

func (p *outputNodeInputPort) Node() nodes.Node {
	return p.node
}

func (p *outputNodeInputPort) Name() string {
	return ValuePortName
}

func (p *outputNodeInputPort) Type() string {
	return p.node.PortType
}

func (p *outputNodeInputPort) Clear() {
	p.connected = nil
}

func (p *outputNodeInputPort) Value() nodes.OutputPort {
	return p.connected
}

func (p *outputNodeInputPort) Set(port nodes.OutputPort) error {
	p.connected = port
	return nil
}
