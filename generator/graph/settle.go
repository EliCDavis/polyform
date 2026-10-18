package graph

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/EliCDavis/polyform/nodes"
)

type conflictPolicy int

const (
	// An edit that adds refuses instead of costing an existing edge.
	refuseConflicts conflictPolicy = iota
	// An edit that removes drops what no longer fits, and reports it.
	dropConflicts
)

// Derives every type and held port from the edges alone, so the result
// depends only on which edges exist.
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
		a.reads[e.consumer][e.input] = slices.Delete(slices.Clone(a.reads[e.consumer][e.input]), at, at+1)
		if len(a.reads[e.consumer][e.input]) == 0 {
			delete(a.reads[e.consumer], e.input)
		}
		dropped = append(dropped, e.dropped(a.scope, conflict.Error()))
	}
}

type edgeConflict struct {
	edge edge
	err  error
}

func (c *edgeConflict) Error() string { return c.err.Error() }

// Producers first; rounds because a node nothing typed feeds takes its type
// from what it feeds, which can then seat what was waiting on it.
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
			untyped, err := seat.fill(offered)
			if err != nil {
				return err
			}
			if waiting == nil {
				waiting = untyped
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

func (s seat) edge(input string, i int) edge {
	return s.graph.edgeAt(s.node, input, s.inputs[input], i)
}

// An input reading an output that has no type yet is left alone and returned as untyped.
func (s seat) fill(offered offers) (untyped *edgeConflict, err error) {
	reads := s.graph.reads[s.node]
	for _, input := range slices.Sorted(maps.Keys(reads)) {
		if _, ok := s.inputs[input]; !ok {
			e := s.edge(input, 0)
			return nil, &edgeConflict{e, fmt.Errorf("node %q no longer has an input %q", e.consumerID, input)}
		}
	}

	for _, input := range slices.Sorted(maps.Keys(s.inputs)) {
		in := s.inputs[input]
		want := make([]nodes.OutputPort, len(reads[input]))
		typed := true

		for i, from := range reads[input] {
			want[i] = offered.of(from)
			if want[i] == nil {
				e := s.edge(input, i)
				return nil, &edgeConflict{e, fmt.Errorf("node %q no longer has an output %q for node %q's %q input to read",
					e.producerID, e.output, e.consumerID, label(in))}
			}

			outType := portTypeOf(want[i])
			switch {
			case outType == "" && isDynamic(want[i]) && portTypeOf(in) != "":
				typed = false
				if e := s.edge(input, i); untyped == nil {
					untyped = &edgeConflict{e, fmt.Errorf("node %q's %q output has no type for node %q's %q input to take",
						e.producerID, e.outputLabel(), e.consumerID, label(in))}
				}
			case outType != "" && isDynamic(in) && portTypeOf(in) == "":
				if err := in.(nodes.DynamicallyTypedPort).BindType(outType); err != nil {
					e := s.edge(input, i)
					return nil, &edgeConflict{e, fmt.Errorf("%s: %w", e, err)}
				}
			}
			if outType != "" && !accepts(in, outType) {
				e := s.edge(input, i)
				return nil, &edgeConflict{e, typeMismatch(e)}
			}
		}

		if !typed {
			continue
		}
		if i, err := hold(in, want); err != nil {
			e := s.edge(input, i)
			return nil, &edgeConflict{e, fmt.Errorf("node %q's %q output now carries %s, which node %q's %q input cannot take (%v)",
				e.producerID, e.outputLabel(), portTypeOf(want[i]), e.consumerID, label(in), err)}
		}
	}

	// Fed, so its inside can be brought in line before anything reads its outputs.
	if placement, ok := s.node.(*SubgraphInstanceNode); ok {
		if err := placement.refresh(); err != nil {
			return nil, err
		}
	}
	return untyped, nil
}

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
				e := s.edge(input, i)
				return bound, &edgeConflict{e, fmt.Errorf("%s: %w", e, err)}
			}
			bound++
		}
	}
	return bound, nil
}

// Touches nothing the input already holds; reports which port it would not take.
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
			if err := nodes.RefuseSpread(want[0], label(in)); err != nil {
				return 0, err
			}
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
			// Set before Add, so a panic the deferred recover catches names this element.
			refused = i
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

// Ports are rebuilt on every Outputs call, so identity is by what they read, not by pointer.
func samePort(held, offered nodes.OutputPort) bool {
	return held != nil &&
		reflect.TypeOf(held) == reflect.TypeOf(offered) &&
		held.Node() == offered.Node() &&
		held.Name() == offered.Name() &&
		label(held) == label(offered)
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

func accepts(in nodes.InputPort, outType string) bool {
	accepted := acceptedTypes(in)
	return accepted[0] == "" || slices.Contains(accepted, outType)
}

func typeMismatch(e edge) error {
	return fmt.Errorf(
		"node %q's %q output is %s, but node %q's %q input takes %s%s",
		e.producerID, e.outputLabel(), portTypeOf(e.out), e.consumerID, label(e.in), strings.Join(acceptedTypes(e.in), " or "), edgeHint(e))
}

// A port taking several types reports only the one it holds now.
func acceptedTypes(in nodes.InputPort) []string {
	if options, ok := in.(nodes.TypeOptions); ok && len(options.AcceptedTypes()) > 0 {
		return options.AcceptedTypes()
	}
	return []string{portTypeOf(in)}
}

func edgeHint(e edge) string {
	from := portTypeOf(e.out)
	for _, want := range acceptedTypes(e.in) {
		if hint := connectionHint(from, want, label(e.in)); hint != "" {
			return hint
		}
	}
	return ""
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
