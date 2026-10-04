package graph

import (
	"fmt"
	"maps"
	"reflect"
	"slices"
	"sort"

	"github.com/EliCDavis/polyform/nodes"
)

// settle derives every type the edges imply, from nothing, and has each
// input hold the port its producer offers for that type. The result depends
// only on which edges exist, never on the order they were made in.
//
//  1. forget: release every dynamic type.
//  2. infer:  bind each dynamic type an edge pins down.
//  3. reseat: swap each held port for the one its producer offers now.
//
// Reseating changes what inputs report, which can pin down more types, so 2
// and 3 repeat until a reseat changes nothing.
func (a *Graph) settle() error {
	edges := a.heldEdges()
	a.forgetDynamicTypes()

	rounds := len(edges) + 2
	for range rounds {
		if err := inferTypes(edges); err != nil {
			return err
		}
		reseated, err := reseat(edges)
		if err != nil {
			return err
		}
		if !reseated {
			return nil
		}
	}
	return fmt.Errorf("the graph's types did not settle after %d rounds", rounds)
}

func (a *Graph) settleDroppingConflicts() ([]DroppedEdge, error) {
	var dropped []DroppedEdge
	for {
		err := a.settle()
		if err == nil {
			return dropped, nil
		}
		conflict, ok := err.(*edgeConflict)
		if !ok {
			return dropped, err
		}
		if err := conflict.edge.unhold(); err != nil {
			return dropped, fmt.Errorf("removing %s: %w", conflict.edge, err)
		}

		drop := conflict.edge.dropped(a.SubGraphScopeID(), conflict.Error())
		a.recordDrop(drop)
		dropped = append(dropped, drop)
	}
}

type edgeConflict struct {
	edge *heldEdge
	err  error
}

func (c *edgeConflict) Error() string { return c.err.Error() }

// Its ports are built once and may go stale; their types and bindings are
// read live, so that is safe.
type heldEdge struct {
	consumerID string
	input      string
	element    int
	in         nodes.InputPort

	producerID string
	producer   nodes.Node
	output     string
	out        nodes.OutputPort

	// Whether either end is a dynamic port, and so has a type to bind.
	bindable bool
}

func (e *heldEdge) String() string {
	return edge{
		from:    portEnd{id: e.producerID, port: e.output},
		to:      portEnd{id: e.consumerID, port: e.input},
		element: e.element,
	}.String()
}

func (e *heldEdge) held() nodes.OutputPort {
	switch slot := e.in.(type) {
	case nodes.SingleValueInputPort:
		return slot.Value()
	case nodes.ArrayValueInputPort:
		return slot.Value()[e.element]
	}
	return nil
}

func (e *heldEdge) hold(port nodes.OutputPort) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	switch slot := e.in.(type) {
	case nodes.SingleValueInputPort:
		err = slot.Set(port)
	case nodes.ArrayValueInputPort:
		err = slot.Replace(e.element, port)
	}
	if err != nil {
		return err
	}
	if e.held() != port {
		return fmt.Errorf("the input dropped it")
	}
	e.out = port
	return nil
}

func (e *heldEdge) dropped(scope, reason string) DroppedEdge {
	return DroppedEdge{
		Scope: scope,
		From:  e.producerID, FromPort: e.output,
		To: e.consumerID, ToPort: edge{to: portEnd{port: e.input}, element: e.element}.inputName(),
		Reason: reason,
	}
}

func (e *heldEdge) unhold() error {
	switch slot := e.in.(type) {
	case nodes.SingleValueInputPort:
		slot.Clear()
		if slot.Value() != nil {
			return fmt.Errorf("the input kept it")
		}
	case nodes.ArrayValueInputPort:
		count := len(slot.Value())
		if err := slot.Remove(slot.Value()[e.element]); err != nil {
			return err
		}
		if len(slot.Value()) != count-1 {
			return fmt.Errorf("the input kept it")
		}
	}
	return nil
}

// Sorted: which edge binds or drops first must not depend on map order.
func (a *Graph) heldEdges() []heldEdge {
	ids := slices.Sorted(maps.Keys(a.nodesByID))

	var edges []heldEdge
	for _, id := range ids {
		inputs := a.nodesByID[id].Inputs()

		names := make([]string, 0, len(inputs))
		for name := range inputs {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			in := inputs[name]
			add := func(element int, out nodes.OutputPort) {
				if out == nil {
					return
				}
				edges = append(edges, heldEdge{
					consumerID: id, input: name, element: element, in: in,
					producerID: a.nodeIDs[out.Node()], producer: out.Node(), output: out.Name(), out: out,
					bindable: isDynamic(in) || isDynamic(out),
				})
			}

			switch slot := in.(type) {
			case nodes.SingleValueInputPort:
				add(nextElement, slot.Value())
			case nodes.ArrayValueInputPort:
				for element, out := range slot.Value() {
					add(element, out)
				}
			}
		}
	}
	return edges
}

func isDynamic(port any) bool {
	dynamic, ok := port.(nodes.DynamicallyTypedPort)
	return ok && dynamic.DynamicPattern() != ""
}

func (a *Graph) forgetDynamicTypes() {
	release := func(port any) {
		if dynamic, ok := port.(nodes.DynamicallyTypedPort); ok {
			dynamic.ReleaseType()
		}
	}
	for node := range a.nodeIDs {
		typed, ok := node.(nodes.DynamicallyTyped)
		if !ok || len(typed.DynamicTypes()) == 0 {
			continue
		}
		for _, port := range node.Inputs() {
			release(port)
		}
		for _, port := range node.Outputs() {
			release(port)
		}
	}
}

type bindDirection int

const (
	// A node takes its type from what feeds it...
	fromProducer bindDirection = iota
	// ...and only a node nothing typed feeds takes it from what it feeds.
	fromConsumer
)

// Downstream reaches a fixed point before anything binds upstream, so a
// producer's type wins wherever both directions could type a node.
func inferTypes(edges []heldEdge) error {
	for {
		bound, err := bindAcross(edges, fromProducer)
		if err != nil {
			return err
		}
		if bound == 0 {
			bound, err = bindAcross(edges, fromConsumer)
			if err != nil {
				return err
			}
		}
		if bound == 0 {
			return nil
		}
	}
}

func bindAcross(edges []heldEdge, direction bindDirection) (int, error) {
	bound := 0
	for i := range edges {
		e := &edges[i]
		if !e.bindable {
			continue
		}
		outType, inType := portTypeOf(e.out), portTypeOf(e.in)

		var untyped any
		var to string
		switch {
		case direction == fromProducer && inType == "" && outType != "":
			untyped, to = e.in, outType
		case direction == fromConsumer && outType == "" && inType != "":
			untyped, to = e.out, inType
		default:
			continue
		}

		if !isDynamic(untyped) {
			continue
		}
		if err := untyped.(nodes.DynamicallyTypedPort).BindType(to); err != nil {
			return bound, &edgeConflict{edge: e, err: fmt.Errorf("%s: %w", e, err)}
		}
		bound++
	}
	return bound, nil
}

func reseat(edges []heldEdge) (bool, error) {
	offers := map[nodes.Node]map[string]nodes.OutputPort{}
	changed := false
	for i := range edges {
		e := &edges[i]
		if _, built := offers[e.producer]; !built {
			offers[e.producer] = e.producer.Outputs()
		}
		offered := offers[e.producer][e.output]
		if offered == nil {
			return changed, &edgeConflict{edge: e, err: fmt.Errorf(
				"node %q no longer has an output %q for node %q's %q input to read",
				e.producerID, e.output, e.consumerID, e.input)}
		}
		if reflect.TypeOf(offered) == reflect.TypeOf(e.out) {
			continue
		}
		if err := e.hold(offered); err != nil {
			return changed, &edgeConflict{edge: e, err: fmt.Errorf(
				"node %q's %q output now carries %s, which node %q's %q input cannot take (%v)",
				e.producerID, e.output, portTypeOf(offered), e.consumerID, e.input, err)}
		}
		changed = true
	}
	return changed, nil
}
