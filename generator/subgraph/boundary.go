package subgraph

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/nodes"
)

const (
	ValuePortName   = "Value"
	DefaultPortName = "default"
)

type Boundary interface {
	nodes.Node
	PortName() string
	PortType() string
	PortDescription() string
	SetPortName(name string)
	SetPortDescription(description string)
}

type InputBoundary interface {
	Boundary
	ReadFrom(external External)
}

// What a placement feeds the boundary, read when asked; nil Value means unfed.
type External interface {
	Value() nodes.OutputPort
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
	PortName        string `json:"portName"`
	PortType        string `json:"portType"`
	PortDescription string `json:"portDescription,omitempty"`
}

type InputNode struct {
	portName        string
	portType        string
	portDescription string

	external    External
	defaultPort *inputNodeDefaultPort
	version     int
}

func NewInputNode(portName, portType string) *InputNode {
	n := &InputNode{
		portName: portName,
		portType: portType,
	}
	n.defaultPort = &inputNodeDefaultPort{node: n}
	return n
}

func NewOutputNode(portName, portType string) *OutputNode {
	n := &OutputNode{
		portName: portName,
		portType: portType,
	}
	n.inputPort = &outputNodeInputPort{node: n}
	return n
}

func (n *InputNode) Copy() *InputNode {
	copied := NewInputNode(n.portName, n.portType)
	copied.portDescription = n.portDescription
	return copied
}

func (n *OutputNode) Copy() *OutputNode {
	copied := NewOutputNode(n.portName, n.portType)
	copied.portDescription = n.portDescription
	return copied
}

func ConfigureBoundaryPortType(boundary Boundary, portType string) error {
	portType = strings.TrimSpace(portType)
	if portType == "" {
		return fmt.Errorf("boundary port type is required")
	}
	switch n := boundary.(type) {
	case *InputNode:
		if n.portType != "" && n.portType != portType {
			return fmt.Errorf("boundary port type cannot be changed")
		}
		n.portType = portType
	case *OutputNode:
		if n.portType != "" && n.portType != portType {
			return fmt.Errorf("boundary port type cannot be changed")
		}
		n.portType = portType
	default:
		return fmt.Errorf("node is not a sub-graph boundary node")
	}
	return nil
}

func BoundaryPortNameConfigured(boundary Boundary) bool {
	switch n := boundary.(type) {
	case *InputNode:
		return strings.TrimSpace(n.portName) != ""
	case *OutputNode:
		return strings.TrimSpace(n.portName) != ""
	default:
		return false
	}
}

func (n *InputNode) PortName() string {
	if n.portName == "" {
		return "Input"
	}
	return n.portName
}

func (n *InputNode) PortType() string {
	return n.portType
}

func (n *InputNode) PortDescription() string {
	return n.portDescription
}

func (n *InputNode) SetPortName(name string) {
	n.portName = name
}

func (n *InputNode) SetPortDescription(description string) {
	n.portDescription = description
}

func ArrayOf(portType string) string {
	return "[]" + portType
}

// Rank only goes one level: an array boundary does not take arrays of arrays.
func AcceptedBoundaryTypes(portType string) []string {
	if strings.HasPrefix(portType, "[]") {
		return []string{portType}
	}
	return []string{portType, ArrayOf(portType)}
}

func effectivePortType(declared string, source nodes.OutputPort) string {
	if source == nil {
		return declared
	}
	typed, ok := source.(nodes.Typed)
	if !ok {
		return declared
	}
	if strings.TrimSpace(typed.Type()) == ArrayOf(declared) {
		return ArrayOf(declared)
	}
	return declared
}

func (n *InputNode) EffectiveType() string {
	return effectivePortType(n.portType, n.source())
}

func (n *InputNode) source() nodes.OutputPort {
	if n.external != nil {
		if fed := n.external.Value(); fed != nil {
			return fed
		}
	}
	return n.defaultPort.Value()
}

func (n *InputNode) Name() string {
	return n.PortName()
}

func (n *InputNode) ReadFrom(external External) {
	n.external = external
	n.version++
	nodes.Touch()
}

func (n *InputNode) Inputs() map[string]nodes.InputPort {
	return map[string]nodes.InputPort{
		DefaultPortName: n.defaultPort,
	}
}

func (n *InputNode) Outputs() map[string]nodes.OutputPort {
	return map[string]nodes.OutputPort{
		ValuePortName: buildInputOutputPort(n),
	}
}

func (n *InputNode) ToJSON(encoder *jbtf.Encoder) ([]byte, error) {
	return json.Marshal(boundaryData{
		PortName:        n.portName,
		PortType:        n.portType,
		PortDescription: n.portDescription,
	})
}

func (n *InputNode) FromJSON(decoder jbtf.Decoder, body []byte) error {
	var data boundaryData
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}
	n.portName = data.PortName
	n.portType = data.PortType
	n.portDescription = data.PortDescription
	return nil
}

func buildInputOutputPort(n *InputNode) nodes.OutputPort {
	source := &inputNodeOutputPort{node: n}
	if builder, ok := nodes.LookupPortTypeProxy(n.EffectiveType()); ok {
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
	return p.node.EffectiveType()
}

func (p *inputNodeOutputPort) Version() int {
	if source := p.node.source(); source != nil {
		return source.Version()
	}
	return p.node.version
}

func (p *inputNodeOutputPort) CurrentSource() nodes.OutputPort {
	return p.node.source()
}

// Read when the placement leaves the input unfed.
type inputNodeDefaultPort struct {
	node      *InputNode
	connected nodes.OutputPort
}

func (p *inputNodeDefaultPort) Node() nodes.Node {
	return p.node
}

func (p *inputNodeDefaultPort) Name() string {
	return DefaultPortName
}

func (p *inputNodeDefaultPort) Type() string {
	return p.node.EffectiveType()
}

func (p *inputNodeDefaultPort) AcceptedTypes() []string {
	return AcceptedBoundaryTypes(p.node.portType)
}

func (p *inputNodeDefaultPort) Clear() {
	p.connected = nil
	p.node.version++
	nodes.Touch()
}

func (p *inputNodeDefaultPort) Value() nodes.OutputPort {
	return p.connected
}

func (p *inputNodeDefaultPort) Set(port nodes.OutputPort) error {
	p.connected = port
	p.node.version++
	nodes.Touch()
	return nil
}

type OutputNode struct {
	portName        string
	portType        string
	portDescription string

	inputPort *outputNodeInputPort
}

func (n *OutputNode) PortName() string {
	if n.portName == "" {
		return "Output"
	}
	return n.portName
}

func (n *OutputNode) PortType() string {
	return n.portType
}

func (n *OutputNode) PortDescription() string {
	return n.portDescription
}

func (n *OutputNode) SetPortName(name string) {
	n.portName = name
}

func (n *OutputNode) SetPortDescription(description string) {
	n.portDescription = description
}

func (n *OutputNode) EffectiveType() string {
	return effectivePortType(n.portType, n.inputPort.Value())
}

func (n *OutputNode) Name() string {
	return n.PortName()
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
		PortName:        n.portName,
		PortType:        n.portType,
		PortDescription: n.portDescription,
	})
}

func (n *OutputNode) FromJSON(decoder jbtf.Decoder, body []byte) error {
	var data boundaryData
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}
	n.portName = data.PortName
	n.portType = data.PortType
	n.portDescription = data.PortDescription
	return nil
}

type outputNodeInputPort struct {
	node      *OutputNode
	connected nodes.OutputPort
}

func (p *outputNodeInputPort) Node() nodes.Node {
	return p.node
}

func (p *outputNodeInputPort) Name() string {
	return ValuePortName
}

func (p *outputNodeInputPort) Type() string {
	return p.node.EffectiveType()
}

func (p *outputNodeInputPort) AcceptedTypes() []string {
	return AcceptedBoundaryTypes(p.node.portType)
}

func (p *outputNodeInputPort) Clear() {
	p.connected = nil
	nodes.Touch()
}

func (p *outputNodeInputPort) Value() nodes.OutputPort {
	return p.connected
}

func (p *outputNodeInputPort) Set(port nodes.OutputPort) error {
	p.connected = port
	nodes.Touch()
	return nil
}
