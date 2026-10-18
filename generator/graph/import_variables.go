package graph

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/persistence"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
)

// VariableInput is an input boundary that replaced a variable an imported
// sub-graph read from the graph it was exported from.
type VariableInput struct {
	Variable string
	Port     string
	// False when the variable's value could not be kept as the input's default.
	HasDefault bool
}

type constantNode interface {
	nodes.Node
	ApplyMessage(msg []byte) (bool, error)
	SetName(name string)
	SetDescription(description string)
	ToJSON(encoder *jbtf.Encoder) ([]byte, error)
}

type liftedVariable struct {
	boundaryID string
	input      VariableInput
}

type variableLifter struct {
	instance  *Instance
	variables schema.NestedGroup[persistence.Variable]
	constants map[string]string
}

// liftVariables rewrites every imported definition so that each variable it
// reads becomes an input boundary defaulting to the variable's value, and
// every placement of a lifted definition feeds those inputs from the
// placing definition's own, so one input still drives all of them.
func (a *Instance) liftVariables(app persistence.App) (map[string]persistence.SubGraph, map[string][]VariableInput, error) {
	lifter := variableLifter{instance: a, variables: app.Variables}
	defs := make(map[string]persistence.SubGraph, len(app.SubGraphs))
	lifted := make(map[string]map[string]liftedVariable, len(app.SubGraphs))
	inputs := make(map[string][]VariableInput, len(app.SubGraphs))

	for _, id := range subGraphLoadOrder(app.SubGraphs) {
		def := app.SubGraphs[id]
		def.Nodes = maps.Clone(def.Nodes)
		scope := liftScope{lifter: &lifter, def: &def, lifted: map[string]liftedVariable{}}

		if err := scope.liftReferences(); err != nil {
			return nil, nil, fmt.Errorf("sub-graph %q: %w", id, err)
		}
		if err := scope.feedPlacements(lifted); err != nil {
			return nil, nil, fmt.Errorf("sub-graph %q: %w", id, err)
		}

		defs[id] = def
		lifted[id] = scope.lifted
		for _, path := range slices.Sorted(maps.Keys(scope.lifted)) {
			inputs[id] = append(inputs[id], scope.lifted[path].input)
		}
	}
	return defs, inputs, nil
}

type liftScope struct {
	lifter *variableLifter
	def    *persistence.SubGraph
	lifted map[string]liftedVariable
}

func (s *liftScope) liftReferences() error {
	replaced := map[string]string{}
	for _, nodeID := range slices.Sorted(maps.Keys(s.def.Nodes)) {
		node := s.def.Nodes[nodeID]
		if node.Variable == nil {
			continue
		}
		if existing, ok := s.lifted[*node.Variable]; ok {
			replaced[nodeID] = existing.boundaryID
			delete(s.def.Nodes, nodeID)
			continue
		}
		if _, err := s.lift(*node.Variable, nodeID); err != nil {
			return err
		}
	}

	for nodeID, node := range s.def.Nodes {
		rewired := false
		for port, from := range node.AssignedInput {
			if to, ok := replaced[from.NodeId]; ok {
				if !rewired {
					node.AssignedInput = maps.Clone(node.AssignedInput)
					rewired = true
				}
				node.AssignedInput[port] = schema.PortReference{NodeId: to, PortName: from.PortName}
			}
		}
		s.def.Nodes[nodeID] = node
	}
	return nil
}

func (s *liftScope) feedPlacements(lifted map[string]map[string]liftedVariable) error {
	for _, nodeID := range slices.Sorted(maps.Keys(s.def.Nodes)) {
		node := s.def.Nodes[nodeID]
		if !subgraph.IsRuntimeNodeType(node.Type) {
			continue
		}
		needed := lifted[subgraph.RuntimeTypeID(node.Type)]
		if len(needed) == 0 {
			continue
		}

		node.AssignedInput = maps.Clone(node.AssignedInput)
		if node.AssignedInput == nil {
			node.AssignedInput = map[string]schema.PortReference{}
		}
		for _, path := range slices.Sorted(maps.Keys(needed)) {
			own, err := s.lift(path, "")
			if err != nil {
				return err
			}
			node.AssignedInput[needed[path].boundaryID] = schema.PortReference{
				NodeId:   own.boundaryID,
				PortName: subgraph.ValuePortName,
			}
		}
		s.def.Nodes[nodeID] = node
	}
	return nil
}

// lift returns the input boundary standing in for path, creating it under
// boundaryID (a fresh id when empty) the first time path is seen.
func (s *liftScope) lift(path, boundaryID string) (liftedVariable, error) {
	if existing, ok := s.lifted[path]; ok {
		return existing, nil
	}

	saved, ok := savedVariable(s.lifter.variables, path)
	if !ok {
		return liftedVariable{}, fmt.Errorf("reads variable %q, which the imported graph does not define", path)
	}
	portType, err := s.lifter.portType(saved)
	if err != nil {
		return liftedVariable{}, fmt.Errorf("variable %q: %w", path, err)
	}

	if boundaryID == "" {
		boundaryID = freeNodeID(s.def.Nodes)
	}
	name := s.freeInputName(path[strings.LastIndex(path, "/")+1:])
	boundaryData, err := json.Marshal(map[string]string{"portName": name, "portType": portType})
	if err != nil {
		return liftedVariable{}, err
	}
	boundary := persistence.Node{Type: subgraph.InputNodeTypeKey, Data: boundaryData}

	input := VariableInput{Variable: path, Port: name}
	if constant, ok := s.lifter.constant(saved, portType, name); ok {
		constantID := freeNodeID(s.def.Nodes, boundaryID)
		s.def.Nodes[constantID] = constant
		boundary.AssignedInput = map[string]schema.PortReference{
			subgraph.DefaultPortName: {NodeId: constantID, PortName: subgraph.ValuePortName},
		}
		input.HasDefault = true
	}
	s.def.Nodes[boundaryID] = boundary

	l := liftedVariable{boundaryID: boundaryID, input: input}
	s.lifted[path] = l
	return l, nil
}

func (s *liftScope) freeInputName(base string) string {
	taken := map[string]bool{}
	for _, node := range s.def.Nodes {
		if node.Type != subgraph.InputNodeTypeKey {
			continue
		}
		var data struct {
			PortName string `json:"portName"`
		}
		if json.Unmarshal(node.Data, &data) == nil {
			taken[data.PortName] = true
		}
	}
	name := base
	for i := 2; taken[name]; i++ {
		name = fmt.Sprintf("%s_%d", base, i)
	}
	return name
}

func (l *variableLifter) portType(saved persistence.Variable) (string, error) {
	var header struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(saved.Data, &header); err != nil {
		return "", err
	}
	v, err := l.instance.variableFactory(header.Type)
	if err != nil {
		return "", err
	}
	typed, ok := v.NodeReference().Outputs()[subgraph.ValuePortName].(nodes.Typed)
	if !ok {
		return "", fmt.Errorf("variable type %q has no typed value", header.Type)
	}
	return typed.Type(), nil
}

// constant builds a node holding the variable's value, if its value is plain
// JSON and some registered value node produces portType.
func (l *variableLifter) constant(saved persistence.Variable, portType, name string) (persistence.Node, bool) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(saved.Data, &fields) != nil {
		return persistence.Node{}, false
	}
	value, ok := fields["value"]
	if !ok {
		return persistence.Node{}, false
	}

	typeKey, ok := l.constantType(portType)
	if !ok {
		return persistence.Node{}, false
	}
	node := l.instance.typeFactory.New(typeKey).(constantNode)
	if _, err := node.ApplyMessage(value); err != nil {
		return persistence.Node{}, false
	}
	node.SetName(name)
	node.SetDescription(saved.Description)
	data, err := node.ToJSON(&jbtf.Encoder{})
	if err != nil {
		return persistence.Node{}, false
	}
	return persistence.Node{Type: typeKey, Data: data}, true
}

func (l *variableLifter) constantType(portType string) (string, bool) {
	if l.constants == nil {
		l.constants = map[string]string{}
		for _, key := range slices.Sorted(slices.Values(l.instance.typeFactory.Types())) {
			node, ok := l.instance.typeFactory.New(key).(constantNode)
			if !ok || len(node.Inputs()) != 0 {
				continue
			}
			outputs := node.Outputs()
			typed, ok := outputs[subgraph.ValuePortName].(nodes.Typed)
			if !ok || len(outputs) != 1 {
				continue
			}
			if _, taken := l.constants[typed.Type()]; !taken {
				l.constants[typed.Type()] = key
			}
		}
	}
	key, ok := l.constants[portType]
	return key, ok
}

func savedVariable(group schema.NestedGroup[persistence.Variable], path string) (persistence.Variable, bool) {
	if v, ok := group.Variables[path]; ok {
		return v, true
	}
	segments := strings.Split(path, "/")
	for _, segment := range segments[:len(segments)-1] {
		next, ok := group.SubGroups[segment]
		if !ok {
			return persistence.Variable{}, false
		}
		group = next
	}
	v, ok := group.Variables[segments[len(segments)-1]]
	return v, ok
}

func freeNodeID(taken map[string]persistence.Node, reserved ...string) string {
	for i := len(taken); ; i++ {
		id := fmt.Sprintf("Node-%d", i)
		if _, used := taken[id]; !used && !slices.Contains(reserved, id) {
			return id
		}
	}
}
