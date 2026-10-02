package graph

import (
	"fmt"
	"slices"
	"strings"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/persistence"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
)

func (a *Instance) loadSubGraphDefinition(subGraphID string, def persistence.SubGraph, decoder jbtf.Decoder) error {
	err := a.CreateSubGraph(subGraphID, def.Name, def.Description)
	if err != nil {
		return err
	}

	target, err := a.SubGraphInstance(subGraphID)
	if err != nil {
		return err
	}

	return populateInstanceFromSubGraphDef(target, def, decoder, nil)
}

func applyPersistedNodeData(nodeDefs map[string]persistence.Node, createdNodes map[string]nodes.Node, decoder jbtf.Decoder) error {
	for nodeID, instanceDetails := range nodeDefs {
		nodeI := createdNodes[nodeID]
		if p, ok := nodeI.(CustomGraphSerialization); ok {
			if err := p.FromJSON(decoder, instanceDetails.Data); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Instance) instantiateAppNode(nodeID string, instanceDetails persistence.Node) (nodes.Node, error) {
	if nodeID == "" {
		panic("attempting to create a node without an ID")
	}

	if instanceDetails.Variable != nil {
		variableInstance, err := a.variables.Variable(*instanceDetails.Variable)
		if err != nil {
			return nil, err
		}
		node := variableInstance.NodeReference()
		a.nodeIDs[node] = nodeID
		a.noteNodeID(nodeID)
		a.nodeTypeKeys[node] = instanceDetails.Type
		return node, nil
	}

	nodeType := instanceDetails.Type
	if subgraph.IsRuntimeNodeType(nodeType) && !a.typeFactory.KeyRegistered(nodeType) {
		subGraphID := subgraph.RuntimeTypeID(nodeType)
		if _, err := a.Root().RegisterSubGraphNodeType(subGraphID); err != nil {
			return nil, err
		}
	}

	newNode := a.typeFactory.New(nodeType)
	casted, ok := newNode.(nodes.Node)
	if !ok {
		panic(fmt.Errorf("graph definition contained type that instantiated a non node: %s", instanceDetails.Type))
	}
	a.nodeIDs[casted] = nodeID
	a.noteNodeID(nodeID)
	a.nodeTypeKeys[casted] = nodeType

	return casted, nil
}

func (a *Instance) bindDynamicTypes(nodeDefs map[string]persistence.Node, createdNodes map[string]nodes.Node) error {
	order := connectionOrder(nodeDefs)

	for range len(order) {
		bound := false

		for _, nodeID := range order {
			node := createdNodes[nodeID]
			inputs := node.Inputs()

			for _, sorted := range sortPortReferences(nodeDefs[nodeID].AssignedInput) {
				input, ok := inputs[strings.Split(sorted.name, ".")[0]]
				if !ok {
					continue
				}
				outNode, ok := createdNodes[sorted.port.NodeId]
				if !ok {
					continue
				}
				output, ok := outNode.Outputs()[sorted.port.PortName]
				if !ok {
					continue
				}

				release, err := bindDynamicPorts(output, input)
				if err != nil {
					return fmt.Errorf("node %s: %w", nodeID, err)
				}
				if release != nil {
					bound = true
				}
			}
		}

		if !bound {
			return nil
		}
	}
	return nil
}

// connectionOrder wires a node's own inputs before anything reads its
// outputs: a lifted output's type follows the rank of what is connected to
// it, and a consumer keeps the port object it was handed rather than asking
// again. A cycle is broken at whichever edge closes it.
func connectionOrder(nodeDefs map[string]persistence.Node) []string {
	ids := make([]string, 0, len(nodeDefs))
	for id := range nodeDefs {
		ids = append(ids, id)
	}
	slices.Sort(ids)

	const (
		unvisited = iota
		visiting
		done
	)

	state := make(map[string]int, len(ids))
	ordered := make([]string, 0, len(ids))

	var visit func(id string)
	visit = func(id string) {
		if state[id] != unvisited {
			return
		}
		state[id] = visiting
		for _, ref := range sortPortReferences(nodeDefs[id].AssignedInput) {
			if _, known := nodeDefs[ref.port.NodeId]; known {
				visit(ref.port.NodeId)
			}
		}
		state[id] = done
		ordered = append(ordered, id)
	}

	for _, id := range ids {
		visit(id)
	}
	return ordered
}

func (a *Instance) connectAppNodes(nodeDefs map[string]persistence.Node, createdNodes map[string]nodes.Node) error {
	if err := a.bindDynamicTypes(nodeDefs, createdNodes); err != nil {
		return err
	}

	for _, nodeID := range connectionOrder(nodeDefs) {
		instanceDetails := nodeDefs[nodeID]
		node := createdNodes[nodeID]
		inputs := node.Inputs()

		sortedInput := sortPortReferences(instanceDetails.AssignedInput)

		for _, sorted := range sortedInput {
			dirtyInputName := sorted.name
			dependency := sorted.port

			inputName := dirtyInputName
			components := strings.Split(inputName, ".")
			if len(components) > 1 {
				inputName = components[0]
			}

			input, ok := inputs[inputName]
			if !ok {
				panic(fmt.Errorf("Node %s has no input %s", nodeID, inputName))
			}

			outNode := createdNodes[dependency.NodeId]
			outNodeOutputs := outNode.Outputs()
			output, ok := outNodeOutputs[dependency.PortName]
			if !ok {
				panic(fmt.Errorf("Node %s has no output %s", dependency.NodeId, dependency.PortName))
			}

			if single, ok := input.(nodes.SingleValueInputPort); ok {
				if err := assignPort(nodeID, dirtyInputName, dependency, output, single.Set); err != nil {
					panic(err)
				}
			} else if array, ok := input.(nodes.ArrayValueInputPort); ok {
				if err := assignPort(nodeID, dirtyInputName, dependency, output, array.Add); err != nil {
					panic(err)
				}
			} else {
				panic(fmt.Errorf("not sure how to assign node %q's input %q", nodeID, inputName))
			}
		}
	}
	return nil
}

func assignPort(
	nodeID, inputName string,
	dependency schema.PortReference,
	output nodes.OutputPort,
	assign func(nodes.OutputPort) error,
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("connecting %s.%s to %s.%s: %v",
				dependency.NodeId, dependency.PortName, nodeID, inputName, r)
		}
	}()

	if err := assign(output); err != nil {
		return fmt.Errorf("connecting %s.%s to %s.%s: %w",
			dependency.NodeId, dependency.PortName, nodeID, inputName, err)
	}
	return nil
}
