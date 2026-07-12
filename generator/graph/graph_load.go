package graph

import (
	"fmt"
	"strings"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/persistence"
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

	return populateInstanceFromSubGraphDef(target, def, decoder)
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
	a.nodeTypeKeys[casted] = nodeType
	return casted, nil
}

func (a *Instance) connectAppNodes(nodeDefs map[string]persistence.Node, createdNodes map[string]nodes.Node) error {
	for nodeID, instanceDetails := range nodeDefs {
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
				if err := single.Set(output); err != nil {
					panic(err)
				}
			} else if array, ok := input.(nodes.ArrayValueInputPort); ok {
				if err := array.Add(output); err != nil {
					panic(err)
				}
			} else {
				panic(fmt.Errorf("not sure how to assign node %q's input %q", nodeID, inputName))
			}
		}
	}
	return nil
}
