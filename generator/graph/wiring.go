package graph

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/EliCDavis/polyform/nodes"
)

// A source is the output an input reads. Graph.reads holds every edge as
// one; the port objects in the nodes' own fields are what settle makes of it.
type source struct {
	node nodes.Node
	port string
}

// An edge is one source of one input, with everything needed to act on it or
// describe it.
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
	return fmt.Sprintf("%s.%s -> %s.%s", e.producerID, e.output, e.consumerID, e.inputName())
}

// The "Port" or "Port.N" form ConnectNodes is called with.
func (e edge) inputName() string {
	if e.element == nextElement {
		return e.input
	}
	return e.input + "." + strconv.Itoa(e.element)
}

// splitElement reads "Port.N" as element N of Port. Any other name is a
// whole port name, dots and all.
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

func (a *Graph) edgeBetween(fromID, fromPort, toID, toInput string) (e edge, err error) {
	e = edge{producerID: fromID, output: fromPort, consumerID: toID}
	e.input, e.element = splitElement(toInput)

	var ok bool
	if e.producer, ok = a.nodesByID[fromID]; !ok {
		return e, fmt.Errorf("no node exists with id %q", fromID)
	}
	if e.consumer, ok = a.nodesByID[toID]; !ok {
		return e, fmt.Errorf("no node exists with id %q", toID)
	}
	if e.in, ok = e.consumer.Inputs()[e.input]; !ok {
		return e, fmt.Errorf("node %q contains no in-port %q", e.consumerID, e.input)
	}
	if e.out, ok = e.producer.Outputs()[e.output]; !ok {
		return e, fmt.Errorf("node %q contains no out-port %q", e.producerID, e.output)
	}
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

// setSources replaces what an input reads. The input's own field follows
// when it can; settle has the last word on whether it fits.
func (a *Graph) setSources(consumer nodes.Node, input string, sources []source) {
	if a.reads[consumer] == nil {
		a.reads[consumer] = make(map[string][]source)
	}
	a.reads[consumer][input] = sources
	if len(sources) == 0 {
		delete(a.reads[consumer], input)
	}

	if in, ok := consumer.Inputs()[input]; ok {
		want := make([]nodes.OutputPort, len(sources))
		for i, from := range sources {
			want[i] = from.node.Outputs()[from.port]
		}
		_, _ = hold(in, want)
	}
}

func (a *Graph) ConnectNodes(nodeOutId, outPortName, nodeInId, inPortName string) error {
	a.mu().Lock()
	defer a.mu().Unlock()
	return a.connectNodes(nodeOutId, outPortName, nodeInId, inPortName)
}

func (a *Graph) connectNodes(nodeOutId, outPortName, nodeInId, inPortName string) error {
	e, err := a.edgeBetween(nodeOutId, outPortName, nodeInId, inPortName)
	if err != nil {
		return err
	}
	if e.producer == e.consumer {
		return fmt.Errorf(
			"connecting node %q's %q output into its own %q input would make it depend on itself",
			e.producerID, e.output, e.input)
	}
	if a.dependsOn(e.producer, e.consumer) {
		return fmt.Errorf(
			"connecting node %q into node %q's %q input would create a cycle: %q already feeds %q, directly or through other nodes",
			e.producerID, e.consumerID, e.input, e.consumerID, e.producerID)
	}
	if err := checkPortTypes(e); err != nil {
		return err
	}

	previous := a.reads[e.consumer][e.input]
	next, err := withEdge(previous, e)
	if err != nil {
		return err
	}

	a.setSources(e.consumer, e.input, next)
	if _, err := a.commitEdit(refuseConflicts); err != nil {
		a.setSources(e.consumer, e.input, previous)
		if _, restoreErr := a.commitEdit(refuseConflicts); restoreErr != nil {
			err = fmt.Errorf("%w (and the graph could not be restored: %v)", err, restoreErr)
		}
		return fmt.Errorf("connecting %s: %w", e, err)
	}

	a.incModelVersion()
	return nil
}

func withEdge(sources []source, e edge) ([]source, error) {
	added := source{node: e.producer, port: e.output}
	_, isArray := e.in.(nodes.ArrayValueInputPort)

	switch {
	case !isArray && e.element != nextElement:
		return nil, fmt.Errorf("node %q's input %q takes a single value, so %q has no element to replace", e.consumerID, e.input, e.inputName())
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
	return nil, fmt.Errorf("node %q's input %q has %d element(s), so %q would leave a gap; use index %d to append", e.consumerID, e.input, len(sources), e.inputName(), len(sources))
}

// DeleteNodeInputConnection returns the other edges that no longer fit
// without this one.
func (a *Graph) DeleteNodeInputConnection(nodeId, portName string) ([]DroppedEdge, error) {
	a.mu().Lock()
	defer a.mu().Unlock()

	node, ok := a.nodesByID[nodeId]
	if !ok {
		return nil, fmt.Errorf("no node exists with id %q", nodeId)
	}
	port, element := splitElement(portName)
	input, ok := node.Inputs()[port]
	if !ok {
		return nil, fmt.Errorf("node %s contains no input port %s", nodeId, port)
	}

	var remaining []source
	if element != nextElement {
		if _, ok := input.(nodes.ArrayValueInputPort); !ok {
			return nil, fmt.Errorf("node %q port %q is not an array, so it has no element %d to remove", nodeId, port, element)
		}
		sources := a.reads[node][port]
		if element >= len(sources) {
			return nil, fmt.Errorf("node %q port %q has %d element(s), so there is no index %d to remove", nodeId, port, len(sources), element)
		}
		remaining = slices.Delete(slices.Clone(sources), element, element+1)
	}
	a.setSources(node, port, remaining)

	dropped, err := a.commitEdit(dropConflicts)
	if err != nil {
		return nil, err
	}
	a.incModelVersion()
	return dropped, nil
}

// dependsOn reports whether node reads target, directly or through any
// chain of nodes. An edge from node into target would then close a loop,
// which evaluates as infinite recursion rather than a graph error.
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

// checkPortTypes refuses a connection whose ends declare different types.
// Both ends have to say what they are for this to apply: an untyped port
// is left alone rather than guessed at.
func checkPortTypes(e edge) error {
	from := portTypeOf(e.out)
	if from == "" {
		return nil
	}

	// A port taking more than one type reports whichever it holds now, so
	// asking for "the" type would pin it there and refuse the other.
	accepted := []string{portTypeOf(e.in)}
	if options, ok := e.in.(nodes.TypeOptions); ok && len(options.AcceptedTypes()) > 0 {
		accepted = options.AcceptedTypes()
	}
	if accepted[0] == "" || slices.Contains(accepted, from) {
		return nil
	}

	hint := ""
	for _, want := range accepted {
		if hint = connectionHint(from, want, e.input); hint != "" {
			break
		}
	}
	return fmt.Errorf(
		"node %q's %q output is %s, but node %q's %q input takes %s%s",
		e.producerID, e.output, from, e.consumerID, e.input, strings.Join(accepted, " or "), hint)
}

func connectionHint(from, to, inPort string) string {
	switch {
	case to == "[]"+from:
		return fmt.Sprintf(" - %q takes the whole array as one value, so build the array first (an arrays.FromElements node) and connect its single output, rather than connecting one element at a time", inPort)
	case from == "int" && to == "float64":
		return " - put a math.IntToFloatNode between them, or feed a float64 source (a float64 parameter or boundary input) in the first place"
	case from == "float64" && to == "int":
		return " - put a math.RoundNode (its Int output) between them, or feed an int source in the first place"
	}
	return ""
}

func (a *Graph) renamePort(node nodes.Node, kind BoundaryPortKind, from, to string) {
	if kind == BoundaryPortKindInput {
		if sources, ok := a.reads[node][from]; ok {
			a.reads[node][to] = sources
			delete(a.reads[node], from)
		}
		return
	}
	for _, inputs := range a.reads {
		for _, sources := range inputs {
			for i, read := range sources {
				if read.node == node && read.port == from {
					sources[i].port = to
				}
			}
		}
	}
}
