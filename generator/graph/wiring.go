package graph

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/EliCDavis/polyform/nodes"
)

type portEnd struct {
	id   string
	node nodes.Node
	port string
}

// An input holds its edge as the port object of the output it reads. Only
// the node and port that object names are the edge; settle derives the rest.
type edge struct {
	from portEnd
	to   portEnd

	// Which element of an array input. nextElement appends; a single-value
	// input takes nextElement too.
	element int
}

const nextElement = -1

func (e edge) String() string {
	return fmt.Sprintf("%s.%s -> %s.%s", e.from.id, e.from.port, e.to.id, e.inputName())
}

// The "Port" or "Port.N" form ConnectNodes is called with.
func (e edge) inputName() string {
	if e.element == nextElement {
		return e.to.port
	}
	return e.to.port + "." + strconv.Itoa(e.element)
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

func (a *Graph) edgeBetween(fromID, fromPort, toID, toInput string) (edge, error) {
	toPort, element := splitElement(toInput)

	from, err := a.portEnd(fromID, fromPort)
	if err != nil {
		return edge{}, err
	}
	to, err := a.portEnd(toID, toPort)
	if err != nil {
		return edge{}, err
	}
	if _, ok := to.node.Inputs()[to.port]; !ok {
		return edge{}, fmt.Errorf("node %q contains no in-port %q", to.id, to.port)
	}
	if _, ok := from.node.Outputs()[from.port]; !ok {
		return edge{}, fmt.Errorf("node %q contains no out-port %q", from.id, from.port)
	}
	return edge{from: from, to: to, element: element}, nil
}

func (a *Graph) portEnd(id, port string) (portEnd, error) {
	node, ok := a.nodesByID[id]
	if !ok {
		return portEnd{}, fmt.Errorf("no node exists with id %q", id)
	}
	return portEnd{id: id, node: node, port: port}, nil
}

func refuseCycle(e edge) error {
	if e.from.node == e.to.node {
		return fmt.Errorf(
			"connecting node %q's %q output into its own %q input would make it depend on itself",
			e.from.id, e.from.port, e.to.port)
	}
	if dependsOn(e.from.node, e.to.node) {
		return fmt.Errorf(
			"connecting node %q into node %q's %q input would create a cycle: %q already feeds %q, directly or through other nodes",
			e.from.id, e.to.id, e.to.port, e.to.id, e.from.id)
	}
	return nil
}

// wire does not check for cycles. untyped reports an output with no type
// of its own yet, which only settle can give it.
func (a *Graph) wire(e edge) (unwire func(), untyped bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("connecting %s: %v", e, r)
		}
	}()

	input, ok := e.to.node.Inputs()[e.to.port]
	if !ok {
		return nil, false, fmt.Errorf("node %q contains no in-port %q", e.to.id, e.to.port)
	}
	output, ok := e.from.node.Outputs()[e.from.port]
	if !ok {
		return nil, false, fmt.Errorf("node %q contains no out-port %q", e.from.id, e.from.port)
	}
	untyped = portTypeOf(output) == ""
	release, err := bindDynamicPorts(output, input)
	if err != nil {
		return nil, untyped, fmt.Errorf("connecting node %q's %q output into node %q's %q input: %w",
			e.from.id, e.from.port, e.to.id, e.to.port, err)
	}
	stored := false
	defer func() {
		if !stored && release != nil {
			release()
		}
	}()
	if release != nil {
		// A producer that just took a type hands out a port typed for it.
		output = e.from.node.Outputs()[e.from.port]
	}

	if err := checkPortTypes(e.from.id, e.from.port, output, e.to.id, e.to.port, input); err != nil {
		return nil, untyped, err
	}

	switch slot := input.(type) {
	case nodes.SingleValueInputPort:
		unwire, err = storeSingle(e, slot, output, input)
	case nodes.ArrayValueInputPort:
		unwire, err = storeElement(e, slot, output, input)
	default:
		err = fmt.Errorf("can not determine type of node %q's input %q", e.to.id, e.to.port)
	}
	stored = err == nil
	return unwire, untyped, err
}

func storeSingle(e edge, slot nodes.SingleValueInputPort, output nodes.OutputPort, input nodes.InputPort) (func(), error) {
	if e.element != nextElement {
		return nil, fmt.Errorf("node %q's input %q takes a single value, so %q has no element to replace", e.to.id, e.to.port, e.inputName())
	}

	previous := slot.Value()
	if err := slot.Set(output); err != nil {
		return nil, err
	}
	// Reflection drops a port of the wrong type rather than complaining.
	if slot.Value() != output {
		return nil, mismatchError(e.from.id, e.from.port, output, e.to.id, e.to.port, input)
	}

	return func() {
		if previous == nil {
			slot.Clear()
			return
		}
		_ = slot.Set(previous)
	}, nil
}

func storeElement(e edge, slot nodes.ArrayValueInputPort, output nodes.OutputPort, input nodes.InputPort) (func(), error) {
	count := len(slot.Value())

	switch {
	// The editor appends by connecting "Port.N" with N == len.
	case e.element == nextElement || e.element == count:
		if err := slot.Add(output); err != nil {
			return nil, err
		}
		if len(slot.Value()) != count+1 {
			return nil, mismatchError(e.from.id, e.from.port, output, e.to.id, e.to.port, input)
		}
		// By position: settle may have swapped the object since.
		return func() { _ = slot.Remove(slot.Value()[count]) }, nil

	case e.element < count:
		previous := slot.Value()[e.element]
		if err := slot.Replace(e.element, output); err != nil {
			return nil, err
		}
		if slot.Value()[e.element] != output {
			return nil, mismatchError(e.from.id, e.from.port, output, e.to.id, e.to.port, input)
		}
		return func() { _ = slot.Replace(e.element, previous) }, nil

	default:
		return nil, fmt.Errorf("node %q's input %q has %d element(s), so %q would leave a gap; use index %d to append", e.to.id, e.to.port, count, e.inputName(), count)
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
	if err := refuseCycle(e); err != nil {
		return err
	}
	unwire, _, err := a.wire(e)
	if err != nil {
		return err
	}
	if err := a.commitEdit(refuseConflicts); err != nil {
		unwire()
		if restoreErr := a.commitEdit(refuseConflicts); restoreErr != nil {
			err = fmt.Errorf("%w (and the graph could not be restored: %v)", err, restoreErr)
		}
		return fmt.Errorf("connecting %s: %w", e, err)
	}

	a.incModelVersion()
	return nil
}

// DeleteNodeInputConnection returns the other edges that no longer fit
// without this one.
func (a *Graph) DeleteNodeInputConnection(nodeId, portName string) (dropped []DroppedEdge, err error) {
	a.mu().Lock()
	defer a.mu().Unlock()

	stop := a.watchDrops()
	defer func() { dropped = stop() }()

	node, ok := a.nodesByID[nodeId]
	if !ok {
		return nil, fmt.Errorf("no node exists with id %q", nodeId)
	}
	port, element := splitElement(portName)
	input, ok := node.Inputs()[port]
	if !ok {
		return nil, fmt.Errorf("node %s contains no input port %s", nodeId, port)
	}

	if element == nextElement {
		input.Clear()
	} else {
		array, ok := input.(nodes.ArrayValueInputPort)
		if !ok {
			return nil, fmt.Errorf("node %q port %q is not an array, so it has no element %d to remove", nodeId, port, element)
		}

		elements := array.Value()
		if element >= len(elements) {
			return nil, fmt.Errorf("node %q port %q has %d element(s), so there is no index %d to remove", nodeId, port, len(elements), element)
		}
		if err := array.Remove(elements[element]); err != nil {
			return nil, err
		}
	}

	if err := a.commitEdit(dropConflicts); err != nil {
		return nil, err
	}
	a.incModelVersion()
	return nil, nil
}

// dependsOn reports whether node's inputs lead back to target, directly or
// through any chain of upstream nodes. Connecting target's output into
// node when this is true closes a loop, which evaluates as infinite
// recursion rather than a graph error.
func dependsOn(node, target nodes.Node) bool {
	visited := make(map[nodes.Node]bool)

	var walk func(nodes.Node) bool
	walk = func(n nodes.Node) bool {
		if n == target {
			return true
		}
		if n == nil || visited[n] {
			return false
		}
		visited[n] = true

		for _, input := range n.Inputs() {
			switch port := input.(type) {
			case nodes.SingleValueInputPort:
				if v := port.Value(); v != nil && walk(v.Node()) {
					return true
				}
			case nodes.ArrayValueInputPort:
				for _, v := range port.Value() {
					if v != nil && walk(v.Node()) {
						return true
					}
				}
			}
		}
		return false
	}

	return walk(node)
}

func portTypeOf(port any) string {
	typed, ok := port.(nodes.Typed)
	if !ok {
		return ""
	}
	return typed.Type()
}

// An untyped end takes the other end's type. Only an unbound variable is
// ever bound here, so the returned undo cannot disturb another connection.
func bindDynamicPorts(output nodes.OutputPort, input nodes.InputPort) (func(), error) {
	outType, inType := portTypeOf(output), portTypeOf(input)

	switch {
	case outType == "" && inType != "":
		return bindPortType(output, inType)
	case inType == "" && outType != "":
		return bindPortType(input, outType)
	}
	return nil, nil
}

func bindPortType(port any, to string) (func(), error) {
	dynamic, ok := port.(nodes.DynamicallyTypedPort)
	if !ok {
		return nil, nil
	}
	if err := dynamic.BindType(to); err != nil {
		return nil, err
	}

	// Ports are rebuilt on every Inputs/Outputs call, but the variable lives
	// on the node, so this stale port still releases the right binding.
	return dynamic.ReleaseType, nil
}

// checkPortTypes refuses a connection whose ends declare different types.
// Both ends have to say what they are for this to apply: an untyped port
// is left alone rather than guessed at.
func checkPortTypes(outID, outPort string, output nodes.OutputPort, inID, inPort string, input nodes.InputPort) error {
	outTyped, ok := output.(nodes.Typed)
	if !ok {
		return nil
	}

	// A port taking more than one type reports whichever it holds now, so
	// asking for "the" type would pin it there and refuse the other.
	if options, ok := input.(nodes.TypeOptions); ok {
		if accepted := options.AcceptedTypes(); len(accepted) > 0 {
			from := outTyped.Type()
			if from == "" || slices.Contains(accepted, from) {
				return nil
			}
			hint := ""
			for _, want := range accepted {
				if h := connectionHint(from, want, inPort); h != "" {
					hint = h
					break
				}
			}
			return fmt.Errorf(
				"node %q's %q output is %s, but node %q's %q input takes %s%s",
				outID, outPort, from, inID, inPort, strings.Join(accepted, " or "), hint)
		}
	}

	inTyped, ok := input.(nodes.Typed)
	if !ok {
		return nil
	}

	from, to := outTyped.Type(), inTyped.Type()
	if from == "" || to == "" || from == to {
		return nil
	}

	return fmt.Errorf(
		"node %q's %q output is %s, but node %q's %q input takes %s%s",
		outID, outPort, from, inID, inPort, to, connectionHint(from, to, inPort))
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

// mismatchError describes a connection the input port refused to hold,
// naming both types so the caller can see which end to change.
func mismatchError(outID, outPort string, output nodes.OutputPort, inID, inPort string, input nodes.InputPort) error {
	describe := func(p any) string {
		if typed, ok := p.(nodes.Typed); ok {
			return typed.Type()
		}
		return "unknown type"
	}
	return fmt.Errorf(
		"node %q's %q output (%s) doesn't fit node %q's %q input (%s), so the connection was refused; wire a node that produces the input's type instead",
		outID, outPort, describe(output), inID, inPort, describe(input))
}
