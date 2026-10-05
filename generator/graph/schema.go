package graph

import (
	"fmt"

	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

// assignedInputs is every edge into node, keyed "Port" or "Port.N" for an
// element of an array input.
func (a *Graph) assignedInputs(node nodes.Node) map[string]schema.PortReference {
	assigned := make(map[string]schema.PortReference)
	if len(a.reads[node]) == 0 {
		return assigned
	}

	inputs := node.Inputs()
	for name, sources := range a.reads[node] {
		_, isArray := inputs[name].(nodes.ArrayValueInputPort)
		for i, from := range sources {
			key := name
			if isArray {
				key = fmt.Sprintf("%s.%d", name, i)
			}
			assigned[key] = schema.PortReference{NodeId: a.nodeIDs[from.node], PortName: from.port}
		}
	}
	return assigned
}

func nodeTypeName(node nodes.Node) string {
	if placement, ok := node.(*SubgraphInstanceNode); ok {
		return subgraph.RuntimeTypePath(placement.SubGraphID())
	}
	return refutil.TypeResolution{IncludePackage: true}.Resolve(node)
}

func (a *Graph) NodeInstanceSchema(node nodes.Node) schema.Node {
	var metadata map[string]any
	if data, ok := a.Metadata("nodes." + a.nodeIDs[node]).(map[string]any); ok {
		metadata = data
	}

	var dynamicTypes map[string]string
	if dynamic, ok := node.(nodes.DynamicallyTyped); ok {
		dynamicTypes = dynamic.DynamicTypes()
	}

	nodeInstance := schema.Node{
		Name:          "Unamed",
		Type:          nodeTypeName(node),
		DynamicTypes:  dynamicTypes,
		AssignedInput: a.assignedInputs(node),
		Output:        make(map[string]schema.NodeOutputPort),
		Metadata:      metadata,
	}

	if reference, ok := node.(variable.Reference); ok {
		referenced := reference.Reference()
		nodeInstance.Name = referenced.Info().Name()
		nodeInstance.Variable = referenced
		a.project.variables.Traverse(func(path string, _ variable.Info, v variable.Variable) {
			if v == referenced {
				nodeInstance.VariablePath = path
			}
		})
	}

	if param, ok := node.(Parameter); ok {
		nodeInstance.Name = param.DisplayName()
		nodeInstance.Parameter = param.Schema()
	} else if named, ok := node.(nodes.Named); ok {
		nodeInstance.Name = named.Name()
	}

	if boundary, ok := subgraph.IsBoundaryNode(node); ok {
		portBoundary := &schema.SubGraphPortBoundary{
			PortName: boundary.BoundaryPortName(),
			PortType: boundary.BoundaryPortType(),
		}
		if _, isInput := subgraph.IsInputBoundary(node); isInput {
			nodeInstance.SubGraphInputBoundary = portBoundary
		} else {
			nodeInstance.SubGraphOutputBoundary = portBoundary
		}
	}

	if placement, ok := node.(*SubgraphInstanceNode); ok {
		nodeInstance.SubGraphId = placement.SubGraphID()
	}

	for name, port := range node.Outputs() {
		output := schema.NodeOutputPort{Version: port.Version()}
		if instance, ok := port.(nodes.InstanceTyped); ok {
			output.Type = instance.InstanceType()
		}
		nodeInstance.Output[name] = output
	}

	return nodeInstance
}

func (a *Graph) ExecutionReport() schema.GraphExecutionReport {
	a.mu().RLock()
	defer a.mu().RUnlock()

	reports := make(map[string]schema.NodeExecutionReport, len(a.nodesByID))
	for id, node := range a.nodesByID {
		report := schema.NodeExecutionReport{Output: make(map[string]nodes.ExecutionReport)}
		for name, port := range node.Outputs() {
			if observable, ok := port.(nodes.ObservableExecution); ok {
				report.Output[name] = observable.ExecutionReport()
			}
		}
		reports[id] = report
	}
	return schema.GraphExecutionReport{Nodes: reports}
}

// Schema of one graph: its nodes and notes. The root graph's is the project's.
func (a *Graph) Schema() schema.Graph {
	if a.IsRoot() {
		return a.project.Schema()
	}
	a.mu().RLock()
	defer a.mu().RUnlock()
	return a.schema()
}

func (a *Graph) schema() schema.Graph {
	notes, _ := a.metadata.Get("notes").(map[string]any)
	graph := schema.Graph{
		Notes: notes,
		Nodes: make(map[string]schema.Node, len(a.nodesByID)),
	}
	for id, node := range a.nodesByID {
		graph.Nodes[id] = a.NodeInstanceSchema(node)
	}
	return graph
}

// Schema of the whole project.
func (a *Instance) Schema() schema.Graph {
	a.mu().RLock()
	defer a.mu().RUnlock()

	variableSchema, err := a.variables.RuntimeSchema()
	if err != nil {
		panic(err)
	}

	graph := a.Graph.schema()
	graph.Variables = variableSchema
	graph.Profiles = a.profiles.Names()
	graph.Variants = a.variantSets.Names()
	graph.Producers = a.producerSchema()
	for name, producer := range graph.Producers {
		node := graph.Nodes[producer.NodeID]
		node.Name = name
		graph.Nodes[producer.NodeID] = node
	}

	graph.SubGraphs = make(map[string]schema.SubGraph, len(a.subGraphs))
	for id, definition := range a.subGraphs {
		inside := definition.instance.schema()
		graph.SubGraphs[id] = schema.SubGraph{
			Name:        definition.name,
			Description: definition.description,
			Nodes:       inside.Nodes,
			Notes:       inside.Notes,
		}
	}

	return graph
}
