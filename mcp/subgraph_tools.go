package mcp

import (
	"context"
	"fmt"
	"sort"

	"github.com/EliCDavis/polyform/generator/subgraph"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type CreateSubgraphInput struct {
	Id          string `json:"id" jsonschema:"unique id for the subgraph; used as its scope when adding nodes to it, and when instantiating it elsewhere"`
	Name        string `json:"name" jsonschema:"human-readable display name"`
	Description string `json:"description,omitempty"`
}

type CreateSubgraphOutput struct {
	Id string `json:"id"`
}

func (s *Server) createSubgraph(ctx context.Context, req *mcpsdk.CallToolRequest, in CreateSubgraphInput) (*mcpsdk.CallToolResult, CreateSubgraphOutput, error) {
	var out CreateSubgraphOutput
	var err error
	s.atomic(&err, func() error {
		if e := s.graph.CreateSubGraph(in.Id, in.Name, in.Description); e != nil {
			return e
		}
		out.Id = in.Id
		return nil
	})
	return nil, out, err
}

type SubgraphSummary struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type ListSubgraphsInput struct{}

type ListSubgraphsOutput struct {
	Subgraphs []SubgraphSummary `json:"subgraphs"`
}

func (s *Server) listSubgraphs(ctx context.Context, req *mcpsdk.CallToolRequest, in ListSubgraphsInput) (*mcpsdk.CallToolResult, ListSubgraphsOutput, error) {
	var out ListSubgraphsOutput
	var err error
	s.atomic(&err, func() error {
		sch := s.graph.Schema()
		for id, sg := range sch.SubGraphs {
			out.Subgraphs = append(out.Subgraphs, SubgraphSummary{
				Id:          id,
				Name:        sg.Name,
				Description: sg.Description,
			})
		}
		sort.Slice(out.Subgraphs, func(i, j int) bool { return out.Subgraphs[i].Id < out.Subgraphs[j].Id })
		return nil
	})
	return nil, out, err
}

type CreateBoundaryNodeInput struct {
	SubgraphId string `json:"subgraphId"`
	Kind       string `json:"kind" jsonschema:"either 'input' or 'output' — whether this port feeds a value into the subgraph or exposes one out of it"`
	PortType   string `json:"portType" jsonschema:"the boundary port's data type key, e.g. float64, github.com/EliCDavis/vector/vector3.Vector[float64]"`
	Name       string `json:"name" jsonschema:"the port name callers will see on nodes created by instantiate_subgraph"`
}

type CreateBoundaryNodeOutput struct {
	NodeId string `json:"nodeId"`
}

func (s *Server) createBoundaryNode(ctx context.Context, req *mcpsdk.CallToolRequest, in CreateBoundaryNodeInput) (*mcpsdk.CallToolResult, CreateBoundaryNodeOutput, error) {
	var out CreateBoundaryNodeOutput
	var err error
	s.atomic(&err, func() error {
		child, e := s.graph.SubGraphInstance(in.SubgraphId)
		if e != nil {
			return e
		}

		var typeKey string
		switch in.Kind {
		case "input":
			typeKey = subgraph.InputNodeTypeKey
		case "output":
			typeKey = subgraph.OutputNodeTypeKey
		default:
			return fmt.Errorf("kind must be \"input\" or \"output\", got %q", in.Kind)
		}

		_, id, e := child.CreateBoundaryNode(typeKey, in.PortType)
		if e != nil {
			return e
		}
		if e := child.SetBoundaryNodeInfo(id, in.Name); e != nil {
			return e
		}
		out.NodeId = id
		return nil
	})
	return nil, out, err
}

type InstantiateSubgraphInput struct {
	SubgraphId string                   `json:"subgraphId"`
	Scope      string                   `json:"scope,omitempty" jsonschema:"where to place the instance; omit for the root graph, or another subgraph id to nest it"`
	Inputs     map[string]NodeInputSpec `json:"inputs,omitempty" jsonschema:"optionally wire the instance's boundary inputs immediately — same shape as create_node's 'inputs': either {nodeId, port} to reference an existing node's output, or {value: <json text>} for a literal (only for types with a registered parameter node)"`
}

type InstantiateSubgraphOutput struct {
	NodeId string `json:"nodeId"`
}

func (s *Server) instantiateSubgraph(ctx context.Context, req *mcpsdk.CallToolRequest, in InstantiateSubgraphInput) (*mcpsdk.CallToolResult, InstantiateSubgraphOutput, error) {
	var out InstantiateSubgraphOutput
	var err error
	s.atomic(&err, func() error {
		inst, e := s.resolveScope(in.Scope)
		if e != nil {
			return e
		}
		node, id, e := inst.CreateNode(subgraph.RuntimeTypePath(in.SubgraphId))
		if e != nil {
			return e
		}
		out.NodeId = id
		return wireCreatedNodeInputs(inst, node, id, in.Inputs)
	})
	return nil, out, err
}

func (s *Server) registerSubgraphTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "create_subgraph",
		Description: "Create a new, initially empty subgraph definition — a reusable component (e.g. a 'wheel' or 'headlight') built once and instantiated many times via instantiate_subgraph.",
	}, s.createSubgraph)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "list_subgraphs",
		Description: "List all subgraph definitions that currently exist in the graph.",
	}, s.listSubgraphs)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "create_boundary_node",
		Description: "Add an input or output boundary port to a subgraph's public interface. Must be created inside the subgraph (pass its id as subgraphId) before wiring the subgraph's interior nodes to it.",
	}, s.createBoundaryNode)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "instantiate_subgraph",
		Description: "Place an instance of a previously created subgraph as a node in the root graph (or another subgraph). Call multiple times to place multiple copies, e.g. 4 wheel instances on a car. Optionally pass 'inputs' to wire its boundary inputs in the same call.",
	}, s.instantiateSubgraph)
}
