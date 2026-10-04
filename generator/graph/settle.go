package graph

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"

	"github.com/EliCDavis/polyform/nodes"
)

type conflictPolicy int

const (
	// An edit that adds refuses instead of costing an existing edge.
	refuseConflicts conflictPolicy = iota
	// An edit that removes drops the edges that depended on what it took,
	// and reports them.
	dropConflicts
)

// settle makes every node's fields agree with the edges: each dynamic type
// is derived from nothing, and each input holds the port its sources offer
// for those types. The result depends only on which edges exist. An edge
// that cannot hold is an error, or under dropConflicts is removed and
// returned.
func (a *Graph) settle(policy conflictPolicy) ([]DroppedEdge, error) {
	var dropped []DroppedEdge
	for {
		err := a.settleOnce()

		var conflict *edgeConflict
		if policy == refuseConflicts || !errors.As(err, &conflict) {
			return dropped, err
		}

		e := conflict.edge
		at := max(e.element, 0)
		a.setSources(e.consumer, e.input, slices.Delete(slices.Clone(a.reads[e.consumer][e.input]), at, at+1))
		dropped = append(dropped, DroppedEdge{
			Scope: a.SubGraphScopeID(),
			From:  e.producerID, FromPort: e.output,
			To: e.consumerID, ToPort: e.inputName(),
			Reason: conflict.Error(),
		})
	}
}

type edgeConflict struct {
	edge edge
	err  error
}

func (c *edgeConflict) Error() string { return c.err.Error() }

// A node takes its type from what feeds it, so nodes are seated producers
// first. Only a node nothing typed feeds takes its type from what it feeds,
// which can then seat what was waiting on it, hence the rounds.
func (a *Graph) settleOnce() error {
	ids, _ := dependenciesFirst(a.nodesByID, func(node nodes.Node) []string {
		var producers []string
		for _, read := range a.readBy(node) {
			producers = append(producers, a.nodeIDs[read])
		}
		return producers
	})
	seats := make([]seat, len(ids))
	for i, id := range ids {
		seats[i] = seat{graph: a, node: a.nodesByID[id], inputs: a.nodesByID[id].Inputs()}
	}

	a.forgetDynamicTypes()

	for range len(seats) + 2 {
		offered := offers{}
		var waiting *edgeConflict
		for _, seat := range seats {
			if err := seat.fill(offered, &waiting); err != nil {
				return err
			}
		}

		bound := 0
		for _, seat := range slices.Backward(seats) {
			n, err := seat.typeProducers(offered)
			if err != nil {
				return err
			}
			bound += n
		}
		if bound == 0 {
			if waiting != nil {
				return waiting
			}
			return nil
		}
	}
	return fmt.Errorf("the graph's types did not settle after %d rounds", len(seats)+2)
}

// What each producer's outputs are right now, asked for once per round.
type offers map[nodes.Node]map[string]nodes.OutputPort

func (o offers) of(from source) nodes.OutputPort {
	if _, built := o[from.node]; !built {
		o[from.node] = from.node.Outputs()
	}
	return o[from.node][from.port]
}

type seat struct {
	graph  *Graph
	node   nodes.Node
	inputs map[string]nodes.InputPort
}

func (s seat) conflict(input string, i int, format string, args ...any) *edgeConflict {
	e := s.graph.edgeAt(s.node, input, s.inputs[input], i)
	return &edgeConflict{edge: e, err: fmt.Errorf(format, args...)}
}

// fill has each of the node's inputs hold what its sources offer, first
// taking any dynamic type they pin down. An input reading an output that
// has no type yet is left for waiting.
func (s seat) fill(offered offers, waiting **edgeConflict) error {
	reads := s.graph.reads[s.node]
	for _, input := range slices.Sorted(maps.Keys(reads)) {
		if _, ok := s.inputs[input]; !ok {
			return s.conflict(input, 0, "node %q no longer has an input %q", s.graph.nodeIDs[s.node], input)
		}
	}

	for _, input := range slices.Sorted(maps.Keys(s.inputs)) {
		in := s.inputs[input]
		want := make([]nodes.OutputPort, len(reads[input]))
		untyped := -1

		for i, from := range reads[input] {
			out := offered.of(from)
			if out == nil {
				return s.conflict(input, i, "node %q no longer has an output %q for node %q's %q input to read",
					s.graph.nodeIDs[from.node], from.port, s.graph.nodeIDs[s.node], input)
			}
			want[i] = out

			outType := portTypeOf(out)
			switch {
			case outType == "" && isDynamic(out) && portTypeOf(in) != "":
				untyped = i
			case outType != "" && isDynamic(in) && portTypeOf(in) == "":
				if err := in.(nodes.DynamicallyTypedPort).BindType(outType); err != nil {
					c := s.conflict(input, i, "")
					c.err = fmt.Errorf("%s: %w", c.edge, err)
					return c
				}
			}
		}

		if untyped != -1 {
			if *waiting == nil {
				*waiting = s.conflict(input, untyped, "node %q's %q output has no type for node %q's %q input to take",
					s.graph.nodeIDs[reads[input][untyped].node], reads[input][untyped].port, s.graph.nodeIDs[s.node], input)
			}
			continue
		}
		if i, err := hold(in, want); err != nil {
			from := reads[input][i]
			return s.conflict(input, i, "node %q's %q output now carries %s, which node %q's %q input cannot take (%v)",
				s.graph.nodeIDs[from.node], from.port, portTypeOf(want[i]), s.graph.nodeIDs[s.node], input, err)
		}
	}
	return nil
}

// typeProducers gives each untyped output the node reads the type of the
// input reading it.
func (s seat) typeProducers(offered offers) (bound int, err error) {
	for _, input := range slices.Sorted(maps.Keys(s.graph.reads[s.node])) {
		for i, from := range s.graph.reads[s.node][input] {
			out := offered.of(from)
			if !isDynamic(out) || portTypeOf(out) != "" {
				continue
			}
			inType := portTypeOf(s.inputs[input])
			if inType == "" {
				continue
			}
			if err := out.(nodes.DynamicallyTypedPort).BindType(inType); err != nil {
				c := s.conflict(input, i, "")
				c.err = fmt.Errorf("%s: %w", c.edge, err)
				return bound, c
			}
			bound++
		}
	}
	return bound, nil
}

// hold has an input hold exactly the given ports, touching nothing when it
// already does. It reports which of them the input would not take.
func hold(in nodes.InputPort, want []nodes.OutputPort) (refused int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	switch slot := in.(type) {
	case nodes.SingleValueInputPort:
		held := slot.Value()
		switch {
		case len(want) == 0:
			if held != nil {
				slot.Clear()
			}
		case !samePort(held, want[0]):
			if err := slot.Set(want[0]); err != nil {
				return 0, err
			}
			// Reflection drops a port of the wrong type rather than complaining.
			if slot.Value() != want[0] {
				return 0, fmt.Errorf("the input dropped it")
			}
		}

	case nodes.ArrayValueInputPort:
		if slices.EqualFunc(slot.Value(), want, samePort) {
			return 0, nil
		}
		slot.Clear()
		for i, port := range want {
			if err := slot.Add(port); err != nil {
				return i, err
			}
			if len(slot.Value()) != i+1 {
				return i, fmt.Errorf("the input dropped it")
			}
		}
	}
	return 0, nil
}

// A port is rebuilt on every Outputs call, so the one held is never the one
// offered; what matters is that it reads the same output as the same type.
func samePort(held, offered nodes.OutputPort) bool {
	return held != nil &&
		reflect.TypeOf(held) == reflect.TypeOf(offered) &&
		held.Node() == offered.Node() &&
		held.Name() == offered.Name()
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
