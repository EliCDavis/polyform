package mcp

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/nodes"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// BoundaryInputSpec declares one input port on a subgraph's interface.
type BoundaryInputSpec struct {
	Name string `json:"name" jsonschema:"the port name callers see on every instance, e.g. \"Radius\". Inside this call it also works as a nodeId in 'nodes' inputs: {\"nodeId\": \"Radius\", \"port\": \"Value\"}."`
	Type string `json:"type" jsonschema:"the port's data type key, e.g. float64, github.com/EliCDavis/vector/vector3.Vector[float64], github.com/EliCDavis/polyform/modeling.Mesh"`
}

// BoundaryOutputSpec declares one output port on a subgraph's interface
// and what feeds it.
type BoundaryOutputSpec struct {
	Name   string `json:"name" jsonschema:"the port name callers see on every instance, e.g. \"Mesh\" or \"Field\""`
	Type   string `json:"type,omitempty" jsonschema:"the port's data type key. Omit it when nodeId is set and the port takes the exact type of the port feeding it - the safe choice for pointer-valued types like *github.com/EliCDavis/polyform/formats/gltf.PolyformModel, whose spelling is easy to get wrong."`
	NodeId string `json:"nodeId,omitempty" jsonschema:"what feeds this output: an alias from 'nodes', an input port name, or an existing node id inside the subgraph. Omit to wire it later with connect_nodes."`
	Port   string `json:"port,omitempty" jsonschema:"output port on nodeId; required when nodeId is set"`
}

type CreateSubgraphInput struct {
	Id          string `json:"id" jsonschema:"unique id for the subgraph; used as its scope when adding nodes to it, and when instantiating it elsewhere"`
	Name        string `json:"name" jsonschema:"human-readable display name"`
	Description string `json:"description,omitempty"`

	Inputs  []BoundaryInputSpec  `json:"inputs,omitempty" jsonschema:"the subgraph's tunable inputs, created before 'nodes' so entries can wire to them by name"`
	Nodes   []CreateNodesEntry   `json:"nodes,omitempty" jsonschema:"the subgraph's interior, in create_nodes form. Aliases may reference each other and any input port by name."`
	Outputs []BoundaryOutputSpec `json:"outputs,omitempty" jsonschema:"what the subgraph exposes, each optionally wired from an alias in 'nodes'"`
}

type CreateSubgraphOutput struct {
	Id      string            `json:"id"`
	Inputs  map[string]string `json:"inputs,omitempty" jsonschema:"input port name -> its boundary node id inside the subgraph (its port is always 'Value')"`
	Outputs map[string]string `json:"outputs,omitempty" jsonschema:"output port name -> its boundary node id inside the subgraph (its port is always 'Value')"`
	Nodes   map[string]string `json:"nodes,omitempty" jsonschema:"alias (or index) -> created node id, as create_nodes returns"`
	Errors  []string          `json:"errors,omitempty" jsonschema:"per-entry failures. The subgraph and everything outside these still exist - fix them with follow-up calls scoped to the subgraph id."`
}

func (s *Server) createSubgraph(ctx context.Context, req *mcpsdk.CallToolRequest, in CreateSubgraphInput) (*mcpsdk.CallToolResult, CreateSubgraphOutput, error) {
	var out CreateSubgraphOutput
	var err error
	s.atomic(&err, func() error {
		if e := s.graph.CreateSubGraph(in.Id, in.Name, in.Description); e != nil {
			return e
		}
		out.Id = in.Id
		if len(in.Inputs) == 0 && len(in.Nodes) == 0 && len(in.Outputs) == 0 {
			return nil
		}

		child, e := s.graph.SubGraphInstance(in.Id)
		if e != nil {
			return e
		}

		// Boundaries first: a malformed port declaration is a hard error
		// (the interface is the whole point), and nothing has been built
		// on it yet, so the subgraph can be discarded cleanly.
		out.Inputs = make(map[string]string, len(in.Inputs))
		for _, port := range in.Inputs {
			id, e := createBoundary(child, subgraph.InputNodeTypeKey, port.Name, port.Type)
			if e != nil {
				_ = s.graph.DeleteSubGraph(in.Id)
				return fmt.Errorf("input %q: %w", port.Name, e)
			}
			out.Inputs[port.Name] = id
		}
		out.Outputs = make(map[string]string, len(in.Outputs))
		for _, port := range in.Outputs {
			if port.Type == "" {
				if port.NodeId == "" {
					_ = s.graph.DeleteSubGraph(in.Id)
					return fmt.Errorf("output %q: type is required unless nodeId names what feeds it", port.Name)
				}
				continue
			}
			id, e := createBoundary(child, subgraph.OutputNodeTypeKey, port.Name, port.Type)
			if e != nil {
				_ = s.graph.DeleteSubGraph(in.Id)
				return fmt.Errorf("output %q: %w", port.Name, e)
			}
			out.Outputs[port.Name] = id
		}

		if len(in.Nodes) > 0 {
			nodeIDs, errs, e := s.createNodeBatch(child, in.Nodes, out.Inputs)
			if e != nil {
				_ = s.graph.DeleteSubGraph(in.Id)
				return e
			}
			out.Nodes = nodeIDs
			out.Errors = append(out.Errors, errs...)
		}

		for _, port := range in.Outputs {
			if port.NodeId == "" {
				continue
			}
			e := func() (e error) {
				defer func() {
					if r := recover(); r != nil {
						e = fmt.Errorf("%v", r)
					}
				}()
				if port.Port == "" {
					return fmt.Errorf("port is required when nodeId is set")
				}
				src := port.NodeId
				if id, ok := out.Nodes[src]; ok {
					src = id
				} else if id, ok := out.Inputs[src]; ok {
					src = id
				}
				outPort := resolveOutputPortName(child.Node(src), port.Port)
				if upstream := child.Node(src); upstream != nil {
					if _, ok := upstream.Outputs()[outPort]; !ok {
						return fmt.Errorf("node %q has no output port %q; it has %s", port.NodeId, port.Port, strings.Join(sortedOutputNames(upstream.Outputs()), ", "))
					}
				}
				if port.Type == "" {
					typeKey, e := outputPortTypeKey(child.Node(src), outPort)
					if e != nil {
						return e
					}
					id, e := createBoundary(child, subgraph.OutputNodeTypeKey, port.Name, typeKey)
					if e != nil {
						return e
					}
					out.Outputs[port.Name] = id
				}
				child.ConnectNodes(src, outPort, out.Outputs[port.Name], subgraph.ValuePortName)
				return nil
			}()
			if e != nil {
				out.Errors = append(out.Errors, fmt.Sprintf("output %q: %v", port.Name, e))
			}
		}
		return nil
	})
	return nil, out, err
}

func outputPortTypeKey(node nodes.Node, portName string) (string, error) {
	if node == nil {
		return "", fmt.Errorf("no node to take the output type from")
	}
	port, ok := node.Outputs()[portName]
	if !ok {
		return "", fmt.Errorf("no output port %q to take the type from", portName)
	}
	typed, ok := port.(nodes.Typed)
	if !ok || typed.Type() == "" {
		return "", fmt.Errorf("output port %q has no type to take; pass type explicitly", portName)
	}
	return typed.Type(), nil
}

func createBoundary(child *graph.Instance, kind, name, portType string) (string, error) {
	_, id, err := child.CreateBoundaryNode(kind, portType)
	if err != nil {
		return "", err
	}
	if err := child.SetBoundaryNodeInfo(id, name); err != nil {
		return "", err
	}
	return id, nil
}

type ConvertToSubgraphInput struct {
	NodeIds             []string `json:"nodeIds" jsonschema:"the nodes to move into the new subgraph, all in the same scope. Every outside source feeding the selection becomes one input port (a variable wired to three selected nodes is one port, not three), and every selected output read from outside becomes one output port. Literal parameter nodes (the ones a 'value' input created for you) that feed only these nodes move in with them automatically, so they don't each become an 'Input N' port; a literal also used outside the selection stays outside and becomes a port."`
	Scope               string   `json:"scope,omitempty" jsonschema:"where the nodes currently live; omit for the root graph"`
	Name                string   `json:"name" jsonschema:"display name for the new subgraph; its id is derived from this"`
	Description         string   `json:"description,omitempty"`
	KeepLiteralsOutside bool     `json:"keepLiteralsOutside,omitempty" jsonschema:"true leaves every literal parameter node outside, so each becomes an input port of the new subgraph. Only for when you really want every literal to be tunable per instance."`
}

// ConvertedPort describes one boundary port the conversion generated.
type ConvertedPort struct {
	BoundaryNodeId string `json:"boundaryNodeId" jsonschema:"the boundary node inside the new subgraph; pass it to rename_boundary_port to give the port a real name"`
	Connection     string `json:"connection" jsonschema:"for an input: the outside 'nodeId.port' feeding it. For an output: the interior 'nodeId.port' it exposes."`
}

type ConvertToSubgraphOutput struct {
	Id       string                   `json:"id" jsonschema:"the new subgraph's id - its scope for further edits, and what instantiate_subgraph takes to place more copies"`
	NodeId   string                   `json:"nodeId" jsonschema:"the instance that replaced the selection in the original scope, already wired to everything the selection was"`
	Inputs   map[string]ConvertedPort `json:"inputs,omitempty" jsonschema:"generated input port name -> its details. Names are positional ('Input 1', ...); rename them with rename_boundary_port so instances read as an interface."`
	Outputs  map[string]ConvertedPort `json:"outputs,omitempty" jsonschema:"generated output port name -> its details"`
	Literals []string                 `json:"literals,omitempty" jsonschema:"literal parameter nodes that moved into the subgraph alongside nodeIds because nothing outside the selection used them"`
}

// literalFeeders finds parameter nodes outside the selection whose every
// consumer is inside it.
func literalFeeders(inst *graph.Instance, selected map[string]bool) []string {
	usedOutside := map[string]bool{}
	usedInside := map[string]bool{}
	for _, consumerID := range inst.NodeIds() {
		consumer := inst.Node(consumerID)
		mark := func(upstream nodes.OutputPort) {
			if upstream == nil {
				return
			}
			if _, isParam := upstream.Node().(graph.Parameter); !isParam {
				return
			}
			if _, isVar := upstream.Node().(variable.Reference); isVar {
				return
			}
			id := inst.NodeId(upstream.Node())
			if selected[id] {
				return
			}
			if selected[consumerID] {
				usedInside[id] = true
			} else {
				usedOutside[id] = true
			}
		}
		for _, port := range consumer.Inputs() {
			switch p := port.(type) {
			case nodes.SingleValueInputPort:
				mark(p.Value())
			case nodes.ArrayValueInputPort:
				for _, upstream := range p.Value() {
					mark(upstream)
				}
			}
		}
	}

	feeders := make([]string, 0)
	for id := range usedInside {
		if !usedOutside[id] {
			feeders = append(feeders, id)
		}
	}
	sort.Strings(feeders)
	return feeders
}

func (s *Server) convertToSubgraph(ctx context.Context, req *mcpsdk.CallToolRequest, in ConvertToSubgraphInput) (*mcpsdk.CallToolResult, ConvertToSubgraphOutput, error) {
	var out ConvertToSubgraphOutput
	var err error
	s.atomic(&err, func() error {
		scope := graph.RootScope
		if in.Scope != "" {
			scope = graph.SubGraphScope(in.Scope)
		}
		parent, e := s.resolveScope(in.Scope)
		if e != nil {
			return e
		}

		nodeIDs := in.NodeIds
		if !in.KeepLiteralsOutside {
			selected := map[string]bool{}
			for _, id := range nodeIDs {
				selected[strings.TrimSpace(id)] = true
			}
			out.Literals = literalFeeders(parent, selected)
			nodeIDs = append(append([]string{}, nodeIDs...), out.Literals...)
		}

		result, e := s.graph.ConvertSelectionToSubGraph(scope, nodeIDs, in.Name, in.Description)
		if e != nil {
			return e
		}
		out.Id = result.SubGraphID
		out.NodeId = result.RuntimeNodeID
		child, e := s.graph.SubGraphInstance(result.SubGraphID)
		if e != nil {
			return e
		}

		outside := map[string]string{}
		for name, port := range parent.Node(result.RuntimeNodeID).Inputs() {
			if single, ok := port.(nodes.SingleValueInputPort); ok && single.Value() != nil {
				outside[name] = parent.NodeId(single.Value().Node()) + "." + single.Value().Name()
			}
		}

		out.Inputs = map[string]ConvertedPort{}
		out.Outputs = map[string]ConvertedPort{}
		for _, id := range child.NodeIds() {
			switch boundary := child.Node(id).(type) {
			case *subgraph.InputNode:
				out.Inputs[boundary.PortName] = ConvertedPort{BoundaryNodeId: id, Connection: outside[boundary.PortName]}
			case *subgraph.OutputNode:
				port := ConvertedPort{BoundaryNodeId: id}
				if fed, ok := boundary.Inputs()[subgraph.ValuePortName].(nodes.SingleValueInputPort); ok && fed.Value() != nil {
					port.Connection = child.NodeId(fed.Value().Node()) + "." + fed.Value().Name()
				}
				out.Outputs[boundary.PortName] = port
			}
		}
		return nil
	})
	return nil, out, err
}

type RenameBoundaryPortInput struct {
	SubgraphId string `json:"subgraphId"`
	NodeId     string `json:"nodeId" jsonschema:"the boundary node inside the subgraph (from create_subgraph's inputs/outputs map, convert_to_subgraph, or describe_graph scoped to the subgraph)"`
	Name       string `json:"name" jsonschema:"the new port name callers see on every instance"`
}

type RenameBoundaryPortOutput struct {
	Renamed bool `json:"renamed"`
}

func (s *Server) renameBoundaryPort(ctx context.Context, req *mcpsdk.CallToolRequest, in RenameBoundaryPortInput) (*mcpsdk.CallToolResult, RenameBoundaryPortOutput, error) {
	var out RenameBoundaryPortOutput
	var err error
	s.atomic(&err, func() error {
		child, e := s.graph.SubGraphInstance(in.SubgraphId)
		if e != nil {
			return e
		}
		if !child.HasNodeWithId(in.NodeId) {
			return s.explainMissingNode(in.SubgraphId, in.NodeId, fmt.Errorf("no node exists with id %q", in.NodeId))
		}
		if e := child.SetBoundaryNodeInfo(in.NodeId, in.Name); e != nil {
			return e
		}
		out.Renamed = true
		return nil
	})
	return nil, out, err
}

type DeleteSubgraphInput struct {
	Id string `json:"id" jsonschema:"id of the subgraph definition to delete; refused while any instance of it still exists"`
}

type DeleteSubgraphOutput struct {
	Deleted bool `json:"deleted"`
}

func (s *Server) deleteSubgraph(ctx context.Context, req *mcpsdk.CallToolRequest, in DeleteSubgraphInput) (*mcpsdk.CallToolResult, DeleteSubgraphOutput, error) {
	var out DeleteSubgraphOutput
	var err error
	s.atomic(&err, func() error {
		if e := s.graph.DeleteSubGraph(in.Id); e != nil {
			return e
		}
		out.Deleted = true
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
	Name       string `json:"name" jsonschema:"the port name callers will see on nodes created by instantiate_subgraph. This is the OUTSIDE name only; the boundary node's own ports are always called 'Value' regardless of what you name it here"`
}

type CreateBoundaryNodeOutput struct {
	NodeId    string   `json:"nodeId"`
	Port      string   `json:"port" jsonschema:"the port name to pass to connect_nodes for this boundary node inside the subgraph - always 'Value', never the name you gave it. An input boundary's 'Value' is an output port (wire it into the subgraph's interior); an output boundary's 'Value' is an input port (wire the interior into it)."`
	Instances []string `json:"instances,omitempty" jsonschema:"for a new input: 'scope/nodeId' of every existing instance of this subgraph, each of which now has this port unconnected and reads the type's zero value until wired. Empty when there are no instances yet."`
}

// subgraphInstances lists every instance of id, in the root graph and
// inside other subgraphs, as "scope/nodeId" ("" scope for root).
func subgraphInstances(sch schema.Graph, id string) []string {
	var found []string
	for nodeID, n := range sch.Nodes {
		if n.SubGraphId == id {
			found = append(found, "/"+nodeID)
		}
	}
	for scope, sg := range sch.SubGraphs {
		for nodeID, n := range sg.Nodes {
			if n.SubGraphId == id {
				found = append(found, scope+"/"+nodeID)
			}
		}
	}
	sort.Strings(found)
	return found
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
		out.Port = subgraph.ValuePortName
		if in.Kind == "input" {
			out.Instances = subgraphInstances(s.graph.Schema(), in.SubgraphId)
		}
		return nil
	})
	return nil, out, err
}

type InstantiateSubgraphInput struct {
	SubgraphId     string                   `json:"subgraphId"`
	Scope          string                   `json:"scope,omitempty" jsonschema:"where to place the instance; omit for the root graph, or another subgraph id to nest it"`
	Inputs         map[string]NodeInputSpec `json:"inputs,omitempty" jsonschema:"optionally wire the instance's boundary inputs immediately — same shape as create_node's 'inputs': either {nodeId, port} to reference an existing node's output, or {value: <json text>} for a literal (only for types with a registered parameter node)"`
	CopyInputsFrom string                   `json:"copyInputsFrom,omitempty" jsonschema:"id of an existing instance of the same subgraph in the same scope: every input wired on it is wired identically on the new one (same upstream nodes, shared - not copied). 'inputs' then overrides individual ports. The way to make a second pipeline that differs in one or two inputs (a low-poly march, a mirrored limb) without re-wiring the other twenty."`
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
		var template *graph.SubgraphInstanceNode
		if in.CopyInputsFrom != "" {
			source := inst.Node(in.CopyInputsFrom)
			if source == nil {
				return s.explainMissingNode(in.Scope, in.CopyInputsFrom, fmt.Errorf("copyInputsFrom: no node exists with id %q", in.CopyInputsFrom))
			}
			sub, ok := source.(*graph.SubgraphInstanceNode)
			if !ok || sub.SubGraphID() != in.SubgraphId {
				return fmt.Errorf("copyInputsFrom: node %q is not an instance of subgraph %q", in.CopyInputsFrom, in.SubgraphId)
			}
			template = sub
		}

		node, id, e := inst.CreateNode(subgraph.RuntimeTypePath(in.SubgraphId))
		if e != nil {
			return e
		}
		out.NodeId = id

		if template != nil {
			for portName, port := range template.Inputs() {
				if _, overridden := in.Inputs[portName]; overridden {
					continue
				}
				switch p := port.(type) {
				case nodes.SingleValueInputPort:
					if upstream := p.Value(); upstream != nil {
						inst.ConnectNodes(inst.NodeId(upstream.Node()), upstream.Name(), id, portName)
					}
				case nodes.ArrayValueInputPort:
					for _, upstream := range p.Value() {
						if upstream != nil {
							inst.ConnectNodes(inst.NodeId(upstream.Node()), upstream.Name(), id, portName)
						}
					}
				}
			}
		}
		return wireCreatedNodeInputs(inst, node, id, in.Inputs)
	})
	return nil, out, err
}

func (s *Server) registerSubgraphTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "create_subgraph",
		Description: "Create a subgraph definition - a named part with its own interface (e.g. a 'paw' with a Size input and a Field output) - in one call: declare 'inputs', build its interior with 'nodes' (create_nodes form, where an input port's name works as a nodeId), and wire 'outputs' from an alias. Then instantiate_subgraph places it, once or many times. All three lists are optional; with none it creates an empty definition to fill via scoped create_nodes and create_boundary_node.",
	}, s.createSubgraph)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "convert_to_subgraph",
		Description: "Move existing nodes into a new subgraph after the fact. Every connection crossing the selection becomes a boundary port, and one instance replaces the selection in place, wired exactly as before - so the graph computes the same thing but the part now has an interface. Literal parameter nodes used only by the selection move in with it (they are not ports); only literals shared with nodes outside, and real upstream nodes, become inputs. Use it when a cluster of nodes you built inline turns out to be a part in its own right (a second copy is needed, or it should be tunable). Port names come out positional; rename them with rename_boundary_port.",
	}, s.convertToSubgraph)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "rename_boundary_port",
		Description: "Rename a subgraph's input or output port as seen on its instances. Existing wiring on every instance follows the rename.",
	}, s.renameBoundaryPort)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "list_subgraphs",
		Description: "List all subgraph definitions that currently exist in the graph.",
	}, s.listSubgraphs)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "delete_subgraph",
		Description: "Delete a subgraph definition that nothing instantiates any more (an abandoned approach, a superseded helper). Refused with the instance count while any instance of it still exists - delete those nodes first.",
	}, s.deleteSubgraph)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "create_boundary_node",
		Description: "Add an input or output boundary port to a subgraph's public interface. Must be created inside the subgraph (pass its id as subgraphId) before wiring the subgraph's interior nodes to it. The port you connect to on the boundary node itself is always called 'Value' - the 'name' argument only sets the port name seen from outside, on instances made by instantiate_subgraph.",
	}, s.createBoundaryNode)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "instantiate_subgraph",
		Description: "Place an instance of a previously created subgraph as a node in the root graph (or another subgraph). Call multiple times to place multiple copies, e.g. 4 wheel instances on a car. Optionally pass 'inputs' to wire its boundary inputs in the same call, or 'copyInputsFrom' (an existing instance) to inherit all of its wiring and override only what differs.",
	}, s.instantiateSubgraph)
}
