package graph

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/EliCDavis/polyform/nodes"
)

// Graph.reads holds every edge as one; node fields are what settle makes of it.
type source struct {
	node nodes.Node
	port string
}

type edge struct {
	producerID string
	producer   nodes.Node
	output     string

	consumerID string
	consumer   nodes.Node
	input      string

	// Which element of an array input. nextElement appends; a single-value
	// input takes nextElement too.
	element int

	in  nodes.InputPort
	out nodes.OutputPort
}

const nextElement = -1

func (e edge) String() string {
	return fmt.Sprintf("%s.%s -> %s.%s", e.producerID, e.outputLabel(), e.consumerID, e.inputLabel())
}

// The "Port" or "Port.N" form of the input's key.
func (e edge) inputName() string {
	return elementName(e.input, e.element)
}

func elementName(port string, element int) string {
	if element == nextElement {
		return port
	}
	return port + "." + strconv.Itoa(element)
}

// What to call the ports in a message: a placement's are labeled by name.
func (e edge) inputLabel() string {
	if e.in == nil {
		return e.inputName()
	}
	return elementName(label(e.in), e.element)
}

func (e edge) outputLabel() string {
	if e.out == nil {
		return e.output
	}
	return label(e.out)
}

func (e edge) dropped(scope, reason string) DroppedEdge {
	return DroppedEdge{Scope: scope, From: e.producerID, FromPort: e.outputLabel(), To: e.consumerID, ToPort: e.inputLabel(), Reason: reason}
}

// Any name not ending in ".N" is a whole port name, dots and all.
func splitElement(name string) (port string, element int) {
	dot := strings.LastIndex(name, ".")
	if dot == -1 {
		return name, nextElement
	}
	element, err := strconv.Atoi(name[dot+1:])
	if err != nil || element < 0 {
		return name, nextElement
	}
	return name[:dot], element
}

func inputKey(inputs map[string]nodes.InputPort, name string) (key string, element int, ok bool) {
	if _, whole := inputs[name]; whole {
		return name, nextElement, true
	}
	port, element := splitElement(name)
	if _, ok := inputs[port]; ok {
		return port, element, true
	}
	for key, in := range inputs {
		if label(in) == name {
			return key, nextElement, true
		}
		if label(in) == port {
			return key, element, true
		}
	}
	return name, nextElement, false
}

func outputKey(outputs map[string]nodes.OutputPort, name string) (string, bool) {
	if _, ok := outputs[name]; ok {
		return name, true
	}
	for key, out := range outputs {
		if label(out) == name {
			return key, true
		}
	}
	return name, false
}

func (a *Graph) edgeBetween(fromID, fromPort, toID, toInput string) (e edge, err error) {
	e = edge{producerID: fromID, output: fromPort, consumerID: toID, input: toInput, element: nextElement}

	var ok bool
	if e.producer, ok = a.nodesByID[fromID]; !ok {
		return e, fmt.Errorf("no node exists with id %q", fromID)
	}
	if e.consumer, ok = a.nodesByID[toID]; !ok {
		return e, fmt.Errorf("no node exists with id %q", toID)
	}
	inputs := e.consumer.Inputs()
	if e.input, e.element, ok = inputKey(inputs, toInput); !ok {
		return e, fmt.Errorf("node %q contains no in-port %q", e.consumerID, e.input)
	}
	e.in = inputs[e.input]
	outputs := e.producer.Outputs()
	if e.output, ok = outputKey(outputs, fromPort); !ok {
		return e, fmt.Errorf("node %q contains no out-port %q", e.producerID, e.output)
	}
	e.out = outputs[e.output]
	return e, nil
}

// edges lists every edge, sorted by consumer, input and element.
func (a *Graph) edges() []edge {
	var edges []edge
	for _, id := range slices.Sorted(maps.Keys(a.nodesByID)) {
		consumer := a.nodesByID[id]
		inputs := consumer.Inputs()
		for _, input := range slices.Sorted(maps.Keys(a.reads[consumer])) {
			for i := range a.reads[consumer][input] {
				edges = append(edges, a.edgeAt(consumer, input, inputs[input], i))
			}
		}
	}
	return edges
}

func (a *Graph) edgeAt(consumer nodes.Node, input string, in nodes.InputPort, i int) edge {
	from := a.reads[consumer][input][i]
	e := edge{
		producerID: a.nodeIDs[from.node], producer: from.node, output: from.port, out: from.node.Outputs()[from.port],
		consumerID: a.nodeIDs[consumer], consumer: consumer, input: input, in: in, element: nextElement,
	}
	if _, isArray := in.(nodes.ArrayValueInputPort); isArray {
		e.element = i
	}
	return e
}

// The input's field follows when the edit closes.
func (a *Graph) setSources(consumer nodes.Node, input string, sources []source) {
	a.dirty()
	if a.reads[consumer] == nil {
		a.reads[consumer] = make(map[string][]source)
	}
	a.reads[consumer][input] = sources
	if len(sources) == 0 {
		delete(a.reads[consumer], input)
	}
}

func (a *Graph) ConnectNodes(nodeOutId, outPortName, nodeInId, inPortName string) error {
	_, err := a.project.write(refuseConflicts, func() error {
		return a.connectNodes(nodeOutId, outPortName, nodeInId, inPortName)
	})
	return err
}

// "Port.N" pushes element N and the ones after it back rather than replacing it.
func (a *Graph) InsertNodeConnection(nodeOutId, outPortName, nodeInId, inPortName string) error {
	_, err := a.project.write(refuseConflicts, func() error {
		return a.wire(nodeOutId, outPortName, nodeInId, inPortName, insertedEdge)
	})
	return err
}

func (a *Graph) connectNodes(nodeOutId, outPortName, nodeInId, inPortName string) error {
	return a.wire(nodeOutId, outPortName, nodeInId, inPortName, withEdge)
}

func (a *Graph) wire(nodeOutId, outPortName, nodeInId, inPortName string, place func([]source, edge) ([]source, error)) error {
	e, err := a.edgeBetween(nodeOutId, outPortName, nodeInId, inPortName)
	if err != nil {
		return err
	}
	if e.producer == e.consumer {
		return fmt.Errorf(
			"connecting node %q's %q output into its own %q input would make it depend on itself",
			e.producerID, e.outputLabel(), e.inputLabel())
	}
	if a.dependsOn(e.producer, e.consumer) {
		return fmt.Errorf(
			"connecting node %q into node %q's %q input would create a cycle: %q already feeds %q, directly or through other nodes",
			e.producerID, e.consumerID, e.inputLabel(), e.consumerID, e.producerID)
	}

	next, err := place(a.reads[e.consumer][e.input], e)
	if err != nil {
		return err
	}
	a.setSources(e.consumer, e.input, next)
	return nil
}

func insertedEdge(sources []source, e edge) ([]source, error) {
	added := source{node: e.producer, port: e.output}
	if _, isArray := e.in.(nodes.ArrayValueInputPort); !isArray {
		return nil, fmt.Errorf("node %q's input %q takes a single value, so there is nothing to insert %q between", e.consumerID, label(e.in), e.inputLabel())
	}
	if e.element == nextElement {
		return append(slices.Clone(sources), added), nil
	}
	if e.element > len(sources) {
		return nil, fmt.Errorf("node %q's input %q has %d element(s), so inserting at %q would leave a gap", e.consumerID, label(e.in), len(sources), e.inputLabel())
	}
	return slices.Insert(slices.Clone(sources), e.element, added), nil
}

func withEdge(sources []source, e edge) ([]source, error) {
	added := source{node: e.producer, port: e.output}
	_, isArray := e.in.(nodes.ArrayValueInputPort)

	switch {
	case !isArray && e.element != nextElement:
		return nil, fmt.Errorf("node %q's input %q takes a single value, so %q has no element to replace", e.consumerID, label(e.in), e.inputLabel())
	case !isArray:
		return []source{added}, nil

	// The editor appends by connecting "Port.N" with N == len.
	case e.element == nextElement || e.element == len(sources):
		return append(slices.Clone(sources), added), nil
	case e.element < len(sources):
		replaced := slices.Clone(sources)
		replaced[e.element] = added
		return replaced, nil
	}
	return nil, fmt.Errorf("node %q's input %q has %d element(s), so %q would leave a gap; use index %d to append", e.consumerID, label(e.in), len(sources), e.inputLabel(), len(sources))
}

// Returns the other edges that no longer fit without this one.
func (a *Graph) DeleteNodeInputConnection(nodeId, portName string) ([]DroppedEdge, error) {
	return a.project.write(dropConflicts, func() error {
		node, ok := a.nodesByID[nodeId]
		if !ok {
			return fmt.Errorf("no node exists with id %q", nodeId)
		}
		inputs := node.Inputs()
		port, element, ok := inputKey(inputs, portName)
		if !ok {
			return fmt.Errorf("node %s contains no input port %s", nodeId, port)
		}

		var remaining []source
		if element != nextElement {
			if _, ok := inputs[port].(nodes.ArrayValueInputPort); !ok {
				return fmt.Errorf("node %q port %q is not an array, so it has no element %d to remove", nodeId, portName, element)
			}
			sources := a.reads[node][port]
			if element >= len(sources) {
				return fmt.Errorf("node %q port %q has %d element(s), so there is no index %d to remove", nodeId, portName, len(sources), element)
			}
			remaining = slices.Delete(slices.Clone(sources), element, element+1)
		}
		a.setSources(node, port, remaining)
		return nil
	})
}

// A loop evaluates as infinite recursion, not a graph error, so it is refused here.
func (a *Graph) dependsOn(node, target nodes.Node) bool {
	visited := make(map[nodes.Node]bool)

	var walk func(nodes.Node) bool
	walk = func(n nodes.Node) bool {
		if n == target {
			return true
		}
		if visited[n] {
			return false
		}
		visited[n] = true
		return slices.ContainsFunc(a.readBy(n), walk)
	}

	return walk(node)
}

func (a *Graph) readBy(node nodes.Node) []nodes.Node {
	var read []nodes.Node
	for _, sources := range a.reads[node] {
		for _, from := range sources {
			read = append(read, from.node)
		}
	}
	return read
}

func portTypeOf(port any) string {
	typed, ok := port.(nodes.Typed)
	if !ok {
		return ""
	}
	return typed.Type()
}
