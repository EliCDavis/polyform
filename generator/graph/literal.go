package graph

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/generator/variable/variabletypes"
	"github.com/EliCDavis/polyform/nodes"
)

var ErrNoLiteral = errors.New("takes no literal")

// node is kept, never rebuilt: settle tells ports apart by their node.
type literal struct {
	value variable.Variable
	node  nodes.Node
}

func (l *literal) port() nodes.OutputPort {
	for _, port := range l.node.Outputs() {
		return port
	}
	return nil
}

type nodeLiterals map[string]*literal

func (l nodeLiterals) prune(inputs map[string]nodes.InputPort) (changed bool) {
	for name := range l {
		if _, ok := inputs[name]; !ok {
			delete(l, name)
			changed = true
		}
	}
	return changed
}

func literalVariableType(portType string) (string, bool) {
	variableType := variabletypes.KeyOfPortType(portType)
	if variableType == "" || strings.HasPrefix(variableType, "[]") || strings.EqualFold(variableType, "image.Image") || strings.EqualFold(variableType, "file") {
		return "", false
	}
	return variableType, true
}

func (a *Graph) buildLiteral(node nodes.Node, name string, input nodes.InputPort) (*literal, error) {
	if _, array := input.(nodes.ArrayValueInputPort); array || isDynamic(input) {
		return nil, nil
	}
	// A lifted port names its element first; its own type follows the wire.
	variableType, ok := literalVariableType(acceptedTypes(input)[0])
	if !ok {
		return nil, nil
	}
	built, err := a.project.buildVariable(variableType)
	if err != nil {
		return nil, nil
	}
	value, err := variable.ClearInfo(built)
	if err != nil {
		return nil, err
	}
	if initial, ok, err := a.literalDefault(node, name, input, variableType); err != nil {
		return nil, err
	} else if ok {
		if _, err := value.ApplyMessage(initial); err != nil {
			return nil, fmt.Errorf("default %s: %w", initial, err)
		}
	}
	return &literal{value: value, node: value.NodeReference()}, nil
}

func (a *Graph) literalDefault(node nodes.Node, name string, input nodes.InputPort, variableType string) (json.RawMessage, bool, error) {
	if defaulted, ok := input.(nodes.Defaulted); ok && defaulted.Default() != "" {
		parsed, err := variabletypes.ParseDefault(variableType, defaulted.Default())
		if err != nil {
			return nil, false, err
		}
		return parsed, true, nil
	}
	if placement, ok := node.(*SubgraphInstanceNode); ok {
		definition := placement.definition()
		if boundary, ok := definition.nodesByID[name]; ok {
			return definition.boundaryDefault(boundary)
		}
	}
	return nil, false, nil
}

func (a *Graph) boundaryDefault(boundary nodes.Node) (json.RawMessage, bool, error) {
	in, ok := boundary.Inputs()[subgraph.DefaultPortName].(nodes.SingleValueInputPort)
	if !ok {
		return nil, false, nil
	}
	port := in.Value()
	if port == nil {
		return nil, false, nil
	}
	if lit := a.literals[boundary][subgraph.DefaultPortName]; lit != nil && port.Node() == lit.node {
		return lit.value.ToMessage(), true, nil
	}
	held, ok := nodes.DynamicAnyValue(nil, port)
	if !ok {
		return nil, false, nil
	}
	// An enum variable is set by member name, not by the value's own JSON.
	if choice, ok := held.(enum.Choice); ok {
		held = choice.Name
	} else if def, ok := enum.Lookup(portTypeOf(port)); ok {
		if name, ok := def.NameOf(held); ok {
			held = name
		}
	}
	data, err := json.Marshal(held)
	if err != nil {
		return nil, false, nil
	}
	return data, true, nil
}

func (a *Graph) reconcileLiterals() (changed bool, err error) {
	for _, id := range slices.Sorted(maps.Keys(a.nodesByID)) {
		node := a.nodesByID[id]
		inputs := node.Inputs()
		have := a.literals[node]
		if have.prune(inputs) {
			changed = true
		}
		for _, name := range slices.Sorted(maps.Keys(inputs)) {
			if _, ok := have[name]; ok {
				continue
			}
			lit, err := a.buildLiteral(node, name, inputs[name])
			if err != nil {
				return changed, fmt.Errorf("node %s input %q: %w", id, label(inputs[name]), err)
			}
			if lit == nil {
				continue
			}
			if have == nil {
				have = nodeLiterals{}
				a.literals[node] = have
			}
			have[name] = lit
			changed = true
		}
	}
	return changed, nil
}

func (a *Graph) isLiteralNode(node nodes.Node) bool {
	reference, ok := node.(variable.Reference)
	return ok && a.project.variablePath(reference.Reference()) == ""
}

func (a *Graph) literalAt(nodeID, port string) (*literal, error) {
	node, ok := a.nodesByID[nodeID]
	if !ok {
		return nil, fmt.Errorf("no node exists with id %q", nodeID)
	}
	lit, err := a.literalOf(node, port)
	if err != nil {
		return nil, fmt.Errorf("node %q: %w", nodeID, err)
	}
	return lit, nil
}

func (a *Graph) literalOf(node nodes.Node, port string) (*literal, error) {
	inputs := node.Inputs()
	name, element, ok := inputKey(inputs, port)
	if !ok {
		return nil, fmt.Errorf("has no input %q", port)
	}
	if element != nextElement {
		return nil, fmt.Errorf("input %q is an array, which holds no literal", port)
	}
	lit := a.literals[node][name]
	if lit == nil {
		return nil, fmt.Errorf("input %q %w", label(inputs[name]), ErrNoLiteral)
	}
	return lit, nil
}

// A node made earlier in the same batch has no literals until the batch closes.
func (a *Graph) seatLiteral(nodeID, port string) error {
	node, ok := a.nodesByID[nodeID]
	if !ok {
		return nil
	}
	inputs := node.Inputs()
	name, element, ok := inputKey(inputs, port)
	if !ok || element != nextElement || a.literals[node][name] != nil {
		return nil
	}
	lit, err := a.buildLiteral(node, name, inputs[name])
	if err != nil || lit == nil {
		return err
	}
	if a.literals[node] == nil {
		a.literals[node] = nodeLiterals{}
	}
	a.literals[node][name] = lit
	return nil
}

// Neither dirties nor marks changed: the literal's own version carries the edit.
func (a *Graph) SetLiteral(nodeID, port string, value []byte) error {
	_, err := a.project.write(refuseConflicts, func() error {
		if err := a.seatLiteral(nodeID, port); err != nil {
			return err
		}
		lit, err := a.literalAt(nodeID, port)
		if err != nil {
			return err
		}
		if _, err := lit.value.ApplyMessage(value); err != nil {
			return fmt.Errorf("node %q's %q literal: %w", nodeID, port, err)
		}
		return nil
	})
	return err
}

func (a *Graph) LiteralValue(nodeID, port string) ([]byte, error) {
	a.mu().RLock()
	defer a.mu().RUnlock()
	lit, err := a.literalAt(nodeID, port)
	if err != nil {
		return nil, err
	}
	return lit.value.ToMessage(), nil
}

func (a *Graph) literalValues(node nodes.Node, name func(nodes.Port) string) map[string]json.RawMessage {
	literals := a.literals[node]
	if len(literals) == 0 {
		return nil
	}
	inputs := node.Inputs()
	values := make(map[string]json.RawMessage, len(literals))
	for input, lit := range literals {
		if port, ok := inputs[input]; ok {
			values[name(port)] = lit.value.ToMessage()
		}
	}
	return values
}
