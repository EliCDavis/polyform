package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/nodes"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"gopkg.in/yaml.v3"
)

// BoundaryInputSpec declares one input port on a subgraph's interface.
type BoundaryInputSpec struct {
	Name    string `json:"name" jsonschema:"the port name callers see on every instance, e.g. \"Radius\". Inside this call it also works as a nodeId in 'nodes' inputs: {\"nodeId\": \"Radius\", \"port\": \"Value\"}."`
	Type    string `json:"type" jsonschema:"the port's data type key, e.g. float64, github.com/EliCDavis/vector/vector3.Vector[float64], github.com/EliCDavis/polyform/modeling.Mesh"`
	Default string `json:"default,omitempty" jsonschema:"JSON literal the port reads on an instance that leaves it unwired, e.g. \"2.5\". Without one an unwired port reads the type's zero value"`
}

func setBoundaryDefault(child *graph.Graph, boundaryID, value string) error {
	return wireOneInput(child, boundaryID, subgraph.DefaultPortName, NodeInputElement{Value: value}, child.Node(boundaryID).Inputs(), nil)
}

func schemaVariableValue(sch schema.Graph, path string) (any, bool) {
	var value any
	found := false
	sch.Variables.Traverse(func(p string, v schema.Variable) bool {
		if p == path {
			value, found = v.Value, true
		}
		return !found
	})
	return value, found
}

func feederDefault(sch schema.Graph, scope, connection string) (string, bool) {
	nodeID, _, _ := strings.Cut(connection, ".")
	holder := sch.Nodes
	if scope != "" {
		holder = sch.SubGraphs[scope].Nodes
	}
	n, ok := holder[nodeID]
	if !ok {
		return "", false
	}
	var value any
	switch {
	case n.VariablePath != "":
		value, ok = schemaVariableValue(sch, n.VariablePath)
	case n.Parameter != nil:
		value = parameterValue(n.Parameter)
		ok = value != nil
	default:
		return "", false
	}
	if !ok {
		return "", false
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "", false
	}
	return string(data), true
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
			if port.Default == "" {
				continue
			}
			if e := setBoundaryDefault(child, id, port.Default); e != nil {
				_ = s.graph.DeleteSubGraph(in.Id)
				return fmt.Errorf("input %q default: %w", port.Name, e)
			}
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
			var nodeIDs map[string]string
			var errs []string
			dropped, e := s.graph.Batch(func() (e error) {
				nodeIDs, errs, e = s.createNodeBatch(child, in.Nodes, out.Inputs)
				return e
			})
			if e != nil {
				_ = s.graph.DeleteSubGraph(in.Id)
				return e
			}
			out.Nodes = nodeIDs
			out.Errors = append(out.Errors, errs...)
			aliases := maps.Clone(nodeIDs)
			maps.Copy(aliases, out.Inputs)
			out.Errors = append(out.Errors, batchErrorLines(dropped, aliases)...)
		}

		var outputErrors []string
		for _, port := range in.Outputs {
			if port.NodeId == "" {
				continue
			}
			e := func() error {
				if port.Port == "" {
					return fmt.Errorf("port is required when nodeId is set")
				}
				src := port.NodeId
				if id, ok := out.Nodes[src]; ok {
					src = id
				} else if id, ok := out.Inputs[src]; ok {
					src = id
				}
				upstream := lookupNode(child, src)
				outPort := resolveOutputPortName(upstream, port.Port)
				if upstream != nil {
					if _, ok := upstream.Outputs()[outPort]; !ok {
						return fmt.Errorf("node %q has no output port %q; it has %s", port.NodeId, port.Port, strings.Join(sortedOutputNames(upstream.Outputs()), ", "))
					}
				}
				if port.Type == "" {
					typeKey, e := outputPortTypeKey(upstream, outPort)
					if e != nil {
						return e
					}
					id, e := createBoundary(child, subgraph.OutputNodeTypeKey, port.Name, typeKey)
					if e != nil {
						return e
					}
					out.Outputs[port.Name] = id
				}
				return child.ConnectNodes(src, outPort, out.Outputs[port.Name], subgraph.ValuePortName)
			}()
			if e != nil {
				outputErrors = append(outputErrors, fmt.Sprintf("output %q: %v", port.Name, e))
			}
		}
		if len(outputErrors) > 0 {
			return fmt.Errorf("nothing was created, because the subgraph's outputs could not be wired:\n%s%s",
				strings.Join(outputErrors, "\n"), interiorErrors(out.Errors))
		}
		return nil
	})
	return nil, out, err
}

func interiorErrors(errs []string) string {
	if len(errs) == 0 {
		return ""
	}
	return "\nInterior entries that failed, which an output may have pointed at:\n" + strings.Join(errs, "\n")
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

func createBoundary(child *graph.Graph, kind, name, portType string) (string, error) {
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
	NodeIds             []string          `json:"nodeIds" jsonschema:"the nodes to move into the new subgraph, all in the same scope. Every outside node feeding the selection becomes one input port (one source wired to three selected nodes is one port, not three), and every selected output read from outside becomes one output port. Literal parameter nodes (the ones a 'value' input created for you) that feed only these nodes move in with them automatically, so they don't each become a port; a literal also used outside the selection stays outside and becomes a port. Variable reference nodes never move in, even when listed: each variable the selection reads becomes one input port named after it. Every port fed by a variable or a literal defaults to its current value, so an unwired instance still computes the same."`
	Scope               string            `json:"scope,omitempty" jsonschema:"where the nodes currently live; omit for the root graph"`
	Name                string            `json:"name" jsonschema:"display name for the new subgraph; its id is derived from this"`
	Description         string            `json:"description,omitempty"`
	KeepLiteralsOutside bool              `json:"keepLiteralsOutside,omitempty" jsonschema:"true leaves every literal parameter node outside, so each becomes an input port of the new subgraph. Only for when you really want every literal to be tunable per instance."`
	Ports               map[string]string `json:"ports,omitempty" jsonschema:"final names for the generated ports, applied in this same call: {generated port name or its connection: desired name}. A key is matched against the names the conversion would generate (see the result's inputs/outputs for the rules) or against a port's connection - an input's outside 'nodeId.port' feeder, which you already know, or an output's interior 'nodeId.port'. Prefer connections for an output you know by its producer, and names for a port you can predict. An unknown key fails the whole call and lists the generated ports, so the retry can key on them."`
}

// ConvertedPort describes one boundary port the conversion generated.
type ConvertedPort struct {
	BoundaryNodeId string `json:"boundaryNodeId" jsonschema:"the boundary node inside the new subgraph; pass it to rename_boundary_port to give the port a different name"`
	Connection     string `json:"connection" jsonschema:"for an input: the outside 'nodeId.port' feeding it. For an output: the interior 'nodeId.port' it exposes."`
}

type ConvertToSubgraphOutput struct {
	Id        string                   `json:"id" jsonschema:"the new subgraph's id - its scope for further edits, and what instantiate_subgraph takes to place more copies"`
	NodeId    string                   `json:"nodeId" jsonschema:"the instance that replaced the selection in the original scope, already wired to everything the selection was"`
	Inputs    map[string]ConvertedPort `json:"inputs,omitempty" jsonschema:"input port name -> its details, after any 'ports' renames. A generated name is, in order of preference: the variable it carries; the feeding node's own name (a subgraph instance's name, a parameter's label, the outer input port it carries, or a literal wired into the node's Name input, with the output appended when several of that node's outputs cross); the feeding output's name when it says more than Out/Value/Float; else the consuming input's name. A clash gets ' 2', ' 3'."`
	Outputs   map[string]ConvertedPort `json:"outputs,omitempty" jsonschema:"output port name -> its details, after any 'ports' renames. A generated name is the producing node's own name (same sources as for inputs, with the output appended when several of its outputs cross), else the output port's name; a clash gets ' 2', ' 3'."`
	Literals  []string                 `json:"literals,omitempty" jsonschema:"literal parameter nodes that moved into the subgraph alongside nodeIds because nothing outside the selection used them"`
	NoDefault []string                 `json:"noDefault,omitempty" jsonschema:"inputs fed by a variable or literal whose value could not be set as the port's default, with why"`
}

// literalFeeders finds parameter nodes outside the selection whose every
// consumer is inside it.
func literalFeeders(inst *graph.Graph, selected map[string]bool) []string {
	usedOutside := map[string]bool{}
	usedInside := map[string]bool{}
	for _, consumerID := range inst.NodeIds() {
		consumer := inst.Node(consumerID)
		mark := func(upstream nodes.OutputPort) {
			if upstream == nil {
				return
			}
			id := inst.NodeId(upstream.Node())
			if selected[id] {
				return
			}
			if _, isParam := upstream.Node().(graph.Parameter); !isParam {
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

	literals := make([]string, 0)
	for id := range usedInside {
		if !usedOutside[id] {
			literals = append(literals, id)
		}
	}
	sort.Strings(literals)
	return literals
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

		var nodeIDs []string
		selected := map[string]bool{}
		for _, id := range in.NodeIds {
			id = strings.TrimSpace(id)
			if node := lookupNode(parent, id); node != nil {
				if _, isVar := node.(variable.Reference); isVar {
					continue
				}
			}
			nodeIDs = append(nodeIDs, id)
			selected[id] = true
		}
		if !in.KeepLiteralsOutside {
			out.Literals = literalFeeders(parent, selected)
			nodeIDs = append(nodeIDs, out.Literals...)
		}

		result, e := s.graph.ConvertSelectionToSubGraph(scope, nodeIDs, in.Name, in.Description)
		if e != nil {
			return e
		}
		if len(result.Dropped) > 0 {
			unfit := make([]string, len(result.Dropped))
			for i, drop := range result.Dropped {
				unfit[i] = fmt.Sprintf("%s.%s -> %s.%s: %s", drop.From, drop.FromPort, drop.To, drop.ToPort, drop.Reason)
			}
			return fmt.Errorf("nothing was converted: these edges would not fit the new subgraph: %s", strings.Join(unfit, "; "))
		}
		out.Id = result.SubGraphID
		out.NodeId = result.RuntimeNodeID
		child, e := s.graph.SubGraphInstance(result.SubGraphID)
		if e != nil {
			return e
		}

		out.Inputs, out.Outputs = convertedPorts(parent, child, result.RuntimeNodeID)
		sch := s.graph.Schema()
		for _, name := range slices.Sorted(maps.Keys(out.Inputs)) {
			port := out.Inputs[name]
			value, ok := feederDefault(sch, in.Scope, port.Connection)
			if !ok {
				continue
			}
			if e := setBoundaryDefault(child, port.BoundaryNodeId, value); e != nil {
				out.NoDefault = append(out.NoDefault, fmt.Sprintf("%s: %v", name, e))
			}
		}
		if len(in.Ports) == 0 {
			return nil
		}
		renames := make(map[string]string, len(in.Ports))
		for key, name := range in.Ports {
			id, e := resolveConvertedPort(key, out.Inputs, out.Outputs)
			if e != nil {
				return fmt.Errorf("ports: %w", e)
			}
			renames[id] = name
		}
		if e := child.RenameBoundaryPorts(renames); e != nil {
			return fmt.Errorf("ports: %w", e)
		}
		out.Inputs, out.Outputs = convertedPorts(parent, child, result.RuntimeNodeID)
		return nil
	})
	return nil, out, err
}

func convertedPorts(parent, child *graph.Graph, instanceID string) (inputs, outputs map[string]ConvertedPort) {
	outside := map[string]string{}
	for name, port := range parent.Node(instanceID).Inputs() {
		if single, ok := port.(nodes.SingleValueInputPort); ok && single.Value() != nil {
			outside[name] = parent.NodeId(single.Value().Node()) + "." + portLabel(single.Value())
		}
	}

	inputs = map[string]ConvertedPort{}
	outputs = map[string]ConvertedPort{}
	for _, id := range child.NodeIds() {
		switch boundary := child.Node(id).(type) {
		case *subgraph.InputNode:
			inputs[boundary.PortName] = ConvertedPort{BoundaryNodeId: id, Connection: outside[id]}
		case *subgraph.OutputNode:
			port := ConvertedPort{BoundaryNodeId: id}
			if fed, ok := boundary.Inputs()[subgraph.ValuePortName].(nodes.SingleValueInputPort); ok && fed.Value() != nil {
				port.Connection = child.NodeId(fed.Value().Node()) + "." + portLabel(fed.Value())
			}
			outputs[boundary.PortName] = port
		}
	}
	return inputs, outputs
}

func resolveConvertedPort(key string, inputs, outputs map[string]ConvertedPort) (string, error) {
	var matched []string
	for _, ports := range []map[string]ConvertedPort{inputs, outputs} {
		for name, port := range ports {
			if name == key || port.Connection == key {
				matched = append(matched, port.BoundaryNodeId)
			}
		}
	}
	switch len(matched) {
	case 1:
		return matched[0], nil
	case 0:
		return "", fmt.Errorf("no generated port is named or connected as %q; inputs: %s; outputs: %s", key, describePorts(inputs), describePorts(outputs))
	default:
		return "", fmt.Errorf("%q names both an input and an output port; key it by connection instead: inputs: %s; outputs: %s", key, describePorts(inputs), describePorts(outputs))
	}
}

func describePorts(ports map[string]ConvertedPort) string {
	if len(ports) == 0 {
		return "none"
	}
	described := make([]string, 0, len(ports))
	for _, name := range slices.Sorted(maps.Keys(ports)) {
		described = append(described, fmt.Sprintf("%q (%s)", name, ports[name].Connection))
	}
	return strings.Join(described, ", ")
}

type BoundaryRename struct {
	NodeId string `json:"nodeId,omitempty" jsonschema:"the boundary node inside the subgraph (from create_subgraph's inputs/outputs map, convert_to_subgraph, or describe_graph scoped to the subgraph). Either this or port."`
	Port   string `json:"port,omitempty" jsonschema:"the port's current name, instead of nodeId. Refused when an input and an output share it."`
	Name   string `json:"name" jsonschema:"the new port name callers see on every instance"`
}

type RenameBoundaryPortInput struct {
	SubgraphId string           `json:"subgraphId"`
	NodeId     string           `json:"nodeId,omitempty" jsonschema:"the boundary node inside the subgraph (from create_subgraph's inputs/outputs map, convert_to_subgraph, or describe_graph scoped to the subgraph)"`
	Name       string           `json:"name,omitempty" jsonschema:"the new port name callers see on every instance"`
	Renames    []BoundaryRename `json:"renames,omitempty" jsonschema:"several renames in one call; nodeId and name are then ignored. Applied as one edit, so names may trade places, and nothing is renamed when any entry fails."`
}

type RenameBoundaryPortOutput struct {
	Renamed bool `json:"renamed"`
	Count   int  `json:"count" jsonschema:"how many ports were renamed"`
}

func (s *Server) renameBoundaryPort(ctx context.Context, req *mcpsdk.CallToolRequest, in RenameBoundaryPortInput) (*mcpsdk.CallToolResult, RenameBoundaryPortOutput, error) {
	var out RenameBoundaryPortOutput
	var err error
	s.atomic(&err, func() error {
		child, e := s.graph.SubGraphInstance(in.SubgraphId)
		if e != nil {
			return e
		}
		renames := in.Renames
		if len(renames) == 0 {
			renames = []BoundaryRename{{NodeId: in.NodeId, Name: in.Name}}
		}
		byID := make(map[string]string, len(renames))
		for i, rename := range renames {
			id, e := s.resolveBoundary(child, in.SubgraphId, rename)
			if e != nil {
				if len(in.Renames) > 0 {
					return fmt.Errorf("renames[%d]: %w", i, e)
				}
				return e
			}
			if _, twice := byID[id]; twice {
				return fmt.Errorf("renames[%d]: boundary node %q is renamed twice", i, id)
			}
			byID[id] = rename.Name
		}
		if e := child.RenameBoundaryPorts(byID); e != nil {
			return e
		}
		out.Renamed = true
		out.Count = len(byID)
		return nil
	})
	return nil, out, err
}

func (s *Server) resolveBoundary(child *graph.Graph, scope string, rename BoundaryRename) (string, error) {
	switch {
	case rename.NodeId != "" && rename.Port != "":
		return "", fmt.Errorf("give nodeId or port, not both")
	case rename.NodeId != "":
		if !child.HasNodeWithId(rename.NodeId) {
			return "", s.explainMissingNode(scope, rename.NodeId, fmt.Errorf("no node exists with id %q", rename.NodeId))
		}
		return rename.NodeId, nil
	case rename.Port == "":
		return "", fmt.Errorf("nodeId or port is required")
	}

	var matched []string
	for _, id := range child.NodeIds() {
		if boundary, ok := subgraph.IsBoundaryNode(child.Node(id)); ok && boundary.BoundaryPortName() == rename.Port {
			matched = append(matched, id)
		}
	}
	switch len(matched) {
	case 1:
		return matched[0], nil
	case 0:
		return "", fmt.Errorf("no boundary port is named %q", rename.Port)
	default:
		return "", fmt.Errorf("both an input and an output are named %q; use nodeId (%s)", rename.Port, strings.Join(matched, ", "))
	}
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
		dropOutlinePart(s.graph, in.Id)
		out.Deleted = true
		return nil
	})
	return nil, out, err
}

type SubgraphSummary struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Placements  []string `json:"placements,omitempty" jsonschema:"'scope/nodeId' of every instance; scope is the subgraph definition holding it, empty for the root graph ('/Node-12')"`
	Instances   []string `json:"instances,omitempty" jsonschema:"live paths from the root graph to every instance, 'Node-81/Node-158' for one nested inside another part's instance: what evaluate_node, sample_field, raycast_field and describe_mesh take as instanceId"`
}

type ListSubgraphsInput struct{}

type ListSubgraphsOutput struct {
	Subgraphs []SubgraphSummary `json:"subgraphs"`
}

// instancePaths walks from the root graph down to every live instance of id.
func instancePaths(sch schema.Graph, id string, visiting map[string]bool) []string {
	if visiting[id] {
		return nil
	}
	visiting[id] = true
	defer delete(visiting, id)

	var paths []string
	for nodeID, n := range sch.Nodes {
		if n.SubGraphId == id {
			paths = append(paths, nodeID)
		}
	}
	for scope, sg := range sch.SubGraphs {
		for nodeID, n := range sg.Nodes {
			if n.SubGraphId != id {
				continue
			}
			for _, outer := range instancePaths(sch, scope, visiting) {
				paths = append(paths, outer+"/"+nodeID)
			}
		}
	}
	sort.Strings(paths)
	return paths
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
				Placements:  subgraphInstances(sch, id),
				Instances:   instancePaths(sch, id, map[string]bool{}),
			})
		}
		sort.Slice(out.Subgraphs, func(i, j int) bool { return out.Subgraphs[i].Id < out.Subgraphs[j].Id })
		return nil
	})
	return nil, out, err
}

type NodeListing struct {
	Id         string `json:"id"`
	Type       string `json:"type" jsonschema:"the type key with the common github.com/EliCDavis/polyform/ prefix and nodes.Struct wrapper removed, e.g. math.AddNode[float64]"`
	Name       string `json:"name,omitempty"`
	SubgraphId string `json:"subgraphId,omitempty" jsonschema:"set when the node is a subgraph instance"`
}

type ListNodesInput struct {
	Scope string `json:"scope,omitempty" jsonschema:"subgraph id whose interior to list; omit for the root graph"`
}

type ListNodesOutput struct {
	Scope string        `json:"scope,omitempty"`
	Nodes []NodeListing `json:"nodes"`
}

func (s *Server) listNodes(ctx context.Context, req *mcpsdk.CallToolRequest, in ListNodesInput) (*mcpsdk.CallToolResult, ListNodesOutput, error) {
	var out ListNodesOutput
	var err error
	s.atomic(&err, func() error {
		inst, e := s.resolveScope(in.Scope)
		if e != nil {
			return e
		}
		out.Scope = in.Scope
		out.Nodes = []NodeListing{}
		for id, n := range inst.Schema().Nodes {
			listing := NodeListing{Id: id, Type: shortTypeKey(n.Type), Name: n.Name, SubgraphId: n.SubGraphId}
			if listing.Name == "Unamed" {
				listing.Name = ""
			}
			out.Nodes = append(out.Nodes, listing)
		}
		sort.Slice(out.Nodes, func(i, j int) bool { return out.Nodes[i].Id < out.Nodes[j].Id })
		return nil
	})
	return nil, out, err
}

type BoundaryPortSpec struct {
	Kind          string            `json:"kind" jsonschema:"'input' or 'output'"`
	PortType      string            `json:"portType" jsonschema:"the port's data type key"`
	Name          string            `json:"name" jsonschema:"the port name seen on instances"`
	WireInstances *NodeInputElement `json:"wireInstances,omitempty" jsonschema:"for an input: what to feed every existing instance in this same call - {variable: path}, {value: json literal}, or {nodeId, port} (looked up in each instance's own scope)"`
	Default       string            `json:"default,omitempty" jsonschema:"for an input: JSON literal an unwired instance reads. Omitted, it is the value or the variable's current value that wireInstances feeds"`
}

type CreateBoundaryNodeInput struct {
	SubgraphId    string             `json:"subgraphId"`
	Kind          string             `json:"kind,omitempty" jsonschema:"either 'input' or 'output' — whether this port feeds a value into the subgraph or exposes one out of it"`
	PortType      string             `json:"portType,omitempty" jsonschema:"the boundary port's data type key, e.g. float64, github.com/EliCDavis/vector/vector3.Vector[float64]"`
	Name          string             `json:"name,omitempty" jsonschema:"the port name callers will see on nodes created by instantiate_subgraph. This is the OUTSIDE name only; the boundary node's own ports are always called 'Value' regardless of what you name it here"`
	WireInstances *NodeInputElement  `json:"wireInstances,omitempty" jsonschema:"for an input: feed every existing instance of the subgraph in this same call - {variable: path}, {value: json literal} (one literal per instance), or {nodeId, port} (looked up in each instance's own scope). Without it the new port is unconnected on every instance and reads the type's zero value."`
	Default       string             `json:"default,omitempty" jsonschema:"for an input: JSON literal an unwired instance reads. Omitted, it is the value or the variable's current value that wireInstances feeds"`
	Through       []string           `json:"through,omitempty" jsonschema:"for an input: the subgraph ids that nest this one, outermost first - ['outer','middle'] when outer places middle and middle places subgraphId. Each gets a same-named input wired to its inner instances, so the value reaches this port from the outermost's instances in one call; wireInstances then feeds the outermost's instances."`
	Ports         []BoundaryPortSpec `json:"ports,omitempty" jsonschema:"several ports in one call, each with its own kind, name, portType, default and wireInstances; the single-port fields must then be left out. 'through' applies to every input in the list."`
}

type CreatedBoundaryPort struct {
	Name      string            `json:"name"`
	NodeId    string            `json:"nodeId"`
	Port      string            `json:"port" jsonschema:"the port name to pass to connect_nodes for this boundary node inside the subgraph - always 'Value', never the name you gave it. An input boundary's 'Value' is an output port (wire it into the subgraph's interior); an output boundary's 'Value' is an input port (wire the interior into it)."`
	Chain     map[string]string `json:"chain,omitempty" jsonschema:"for 'through': outer subgraph id -> the boundary node created there, already wired to its inner instances"`
	Instances []string          `json:"instances,omitempty" jsonschema:"for a new input: 'scope/nodeId' of every existing instance that got this port (of the outermost subgraph when 'through' is set). Each reads the type's zero value until wired, unless wireInstances fed it."`
	Wired     []string          `json:"wired,omitempty" jsonschema:"instances wireInstances fed"`
	Unwired   []string          `json:"unwired,omitempty" jsonschema:"instances left unfed, with why: a 'through' chain only reaches instances placed along it"`
}

type CreateBoundaryNodeOutput struct {
	NodeId    string                `json:"nodeId,omitempty"`
	Port      string                `json:"port,omitempty" jsonschema:"always 'Value'; see ports[].port"`
	Chain     map[string]string     `json:"chain,omitempty"`
	Instances []string              `json:"instances,omitempty" jsonschema:"for a new input: 'scope/nodeId' of every existing instance of this subgraph, each of which now has this port unconnected and reads the type's zero value until wired. Empty when there are no instances yet."`
	Wired     []string              `json:"wired,omitempty"`
	Unwired   []string              `json:"unwired,omitempty"`
	Ports     []CreatedBoundaryPort `json:"ports,omitempty" jsonschema:"one entry per port when 'ports' was passed"`
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

// placementsIn lists the instances of id placed directly inside scope ("" for root).
func placementsIn(sch schema.Graph, scope, id string) []string {
	holder := sch.Nodes
	if scope != "" {
		holder = sch.SubGraphs[scope].Nodes
	}
	var found []string
	for nodeID, n := range holder {
		if n.SubGraphId == id {
			found = append(found, nodeID)
		}
	}
	sort.Strings(found)
	return found
}

func (s *Server) addBoundaryPort(subgraphID string, spec BoundaryPortSpec, through []string) (CreatedBoundaryPort, error) {
	var port CreatedBoundaryPort
	child, err := s.graph.SubGraphInstance(subgraphID)
	if err != nil {
		return port, err
	}

	var typeKey string
	switch spec.Kind {
	case "input":
		typeKey = subgraph.InputNodeTypeKey
	case "output":
		typeKey = subgraph.OutputNodeTypeKey
		if spec.WireInstances != nil || len(through) > 0 || spec.Default != "" {
			return port, fmt.Errorf("output %q: wireInstances, default and through only apply to an input", spec.Name)
		}
	default:
		return port, fmt.Errorf("kind must be \"input\" or \"output\", got %q", spec.Kind)
	}

	id, err := createBoundary(child, typeKey, spec.Name, spec.PortType)
	if err != nil {
		return port, err
	}
	port.Name, port.NodeId, port.Port = spec.Name, id, subgraph.ValuePortName
	if spec.Kind == "output" {
		return port, nil
	}

	sch := s.graph.Schema()
	defaultValue := spec.Default
	if defaultValue == "" && spec.WireInstances != nil {
		switch {
		case spec.WireInstances.Value != "":
			defaultValue = spec.WireInstances.Value
		case spec.WireInstances.Variable != "":
			if value, ok := schemaVariableValue(sch, spec.WireInstances.Variable); ok {
				if data, e := json.Marshal(value); e == nil {
					defaultValue = string(data)
				}
			}
		}
	}
	withDefault := func(g *graph.Graph, boundaryID string) error {
		if defaultValue == "" {
			return nil
		}
		if e := setBoundaryDefault(g, boundaryID, defaultValue); e != nil {
			return fmt.Errorf("default: %w", e)
		}
		return nil
	}
	if err := withDefault(child, id); err != nil {
		return port, err
	}
	target := subgraphID
	for i := len(through) - 1; i >= 0; i-- {
		outer := through[i]
		if outer == target {
			return port, fmt.Errorf("through: %q cannot nest itself", outer)
		}
		outerGraph, err := s.graph.SubGraphInstance(outer)
		if err != nil {
			return port, fmt.Errorf("through: %w", err)
		}
		inner := placementsIn(sch, outer, target)
		if len(inner) == 0 {
			return port, fmt.Errorf("through: subgraph %q places no instance of %q", outer, target)
		}
		for _, placement := range subgraphInstances(sch, target) {
			if !strings.HasPrefix(placement, outer+"/") {
				port.Unwired = append(port.Unwired, fmt.Sprintf("%s: an instance of %q outside the through chain", placement, target))
			}
		}
		outerID, err := createBoundary(outerGraph, subgraph.InputNodeTypeKey, spec.Name, spec.PortType)
		if err != nil {
			return port, fmt.Errorf("through: subgraph %q: %w", outer, err)
		}
		if err := withDefault(outerGraph, outerID); err != nil {
			return port, fmt.Errorf("through: subgraph %q: %w", outer, err)
		}
		for _, placement := range inner {
			if err := outerGraph.ConnectNodes(outerID, subgraph.ValuePortName, placement, spec.Name); err != nil {
				return port, fmt.Errorf("through: subgraph %q: %w", outer, err)
			}
		}
		if port.Chain == nil {
			port.Chain = map[string]string{}
		}
		port.Chain[outer] = outerID
		target = outer
	}

	port.Instances = subgraphInstances(sch, target)
	if spec.WireInstances == nil {
		return port, nil
	}
	for _, placement := range port.Instances {
		scope, nodeID, _ := strings.Cut(placement, "/")
		inst, err := s.resolveScope(scope)
		if err != nil {
			return port, err
		}
		node := inst.Node(nodeID)
		if err := wireOneInput(inst, nodeID, resolveInputPortName(node, spec.Name), *spec.WireInstances, node.Inputs(), nil); err != nil {
			port.Unwired = append(port.Unwired, fmt.Sprintf("%s: %v", placement, err))
			continue
		}
		port.Wired = append(port.Wired, placement)
	}
	return port, nil
}

func (s *Server) createBoundaryNode(ctx context.Context, req *mcpsdk.CallToolRequest, in CreateBoundaryNodeInput) (*mcpsdk.CallToolResult, CreateBoundaryNodeOutput, error) {
	var out CreateBoundaryNodeOutput
	var err error
	s.atomic(&err, func() error {
		if len(in.Ports) == 0 {
			port, e := s.addBoundaryPort(in.SubgraphId, BoundaryPortSpec{Kind: in.Kind, PortType: in.PortType, Name: in.Name, WireInstances: in.WireInstances, Default: in.Default}, in.Through)
			if e != nil {
				return e
			}
			out.NodeId, out.Port, out.Chain = port.NodeId, port.Port, port.Chain
			out.Instances, out.Wired, out.Unwired = port.Instances, port.Wired, port.Unwired
			return nil
		}
		for field, set := range map[string]bool{
			"kind": in.Kind != "", "name": in.Name != "", "portType": in.PortType != "",
			"default": in.Default != "", "wireInstances": in.WireInstances != nil,
		} {
			if set {
				return fmt.Errorf("%q belongs inside each \"ports\" entry", field)
			}
		}
		for i, spec := range in.Ports {
			through := in.Through
			if spec.Kind == "output" {
				through = nil
			}
			port, e := s.addBoundaryPort(in.SubgraphId, spec, through)
			if e != nil {
				return fmt.Errorf("ports[%d] (%s): %w; nothing was created", i, spec.Name, e)
			}
			out.Ports = append(out.Ports, port)
		}
		return nil
	})
	return nil, out, err
}

type InstantiateSubgraphInput struct {
	SubgraphId     string                   `json:"subgraphId"`
	Scope          string                   `json:"scope,omitempty" jsonschema:"where to place the instance; omit for the root graph, or another subgraph id to nest it"`
	Inputs         map[string]NodeInputSpec `json:"inputs,omitempty" jsonschema:"optionally wire the instance's boundary inputs immediately — same shape as create_node's 'inputs': either {nodeId, port} to reference an existing node's output, or {value: <json text>} for a literal (only for types with a registered parameter node)"`
	CopyInputsFrom string                   `json:"copyInputsFrom,omitempty" jsonschema:"id of an existing subgraph instance in the same scope: every input wired on it whose port name the new subgraph also has is wired identically on the new one (same upstream nodes, shared - not copied). The instance may be of another subgraph; ports it has that the new one lacks come back in 'skipped'. 'inputs' then overrides individual ports. The way to make a second pipeline that differs in one or two inputs (a low-poly march, a mirrored copy) without re-wiring the other twenty."`
}

type InstantiateSubgraphOutput struct {
	NodeId  string   `json:"nodeId"`
	Copied  []string `json:"copied,omitempty" jsonschema:"input ports wired from copyInputsFrom"`
	Skipped []string `json:"skipped,omitempty" jsonschema:"input ports wired on copyInputsFrom that could not be copied, with why: the new subgraph has no such input, or its type differs"`
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
			source := lookupNode(inst, in.CopyInputsFrom)
			if source == nil {
				return s.explainMissingNode(in.Scope, in.CopyInputsFrom, fmt.Errorf("copyInputsFrom: no node exists with id %q", in.CopyInputsFrom))
			}
			sub, ok := source.(*graph.SubgraphInstanceNode)
			if !ok {
				return fmt.Errorf("copyInputsFrom: node %q is not an instance of any subgraph", in.CopyInputsFrom)
			}
			template = sub
		}

		node, id, e := inst.CreateNode(subgraph.RuntimeTypePath(in.SubgraphId))
		if e != nil {
			return e
		}
		out.NodeId = id

		if template != nil {
			ownPorts := node.Inputs()
			templatePorts := template.Inputs()
			byLabel := map[string]string{}
			for key, port := range templatePorts {
				byLabel[portLabel(port)] = key
			}
			for _, portName := range slices.Sorted(maps.Keys(byLabel)) {
				if _, overridden := in.Inputs[portName]; overridden {
					continue
				}
				var upstreams []nodes.OutputPort
				switch p := templatePorts[byLabel[portName]].(type) {
				case nodes.SingleValueInputPort:
					if p.Value() != nil {
						upstreams = []nodes.OutputPort{p.Value()}
					}
				case nodes.ArrayValueInputPort:
					upstreams = p.Value()
				}
				if len(upstreams) == 0 {
					continue
				}
				ownKey := resolvePortName(portName, ownPorts)
				if _, has := ownPorts[ownKey]; !has {
					out.Skipped = append(out.Skipped, fmt.Sprintf("%s: %q has no such input", portName, in.SubgraphId))
					continue
				}
				var failed error
				for _, upstream := range upstreams {
					if upstream == nil {
						continue
					}
					if failed = inst.ConnectNodes(inst.NodeId(upstream.Node()), upstream.Name(), id, ownKey); failed != nil {
						break
					}
				}
				if failed != nil {
					out.Skipped = append(out.Skipped, fmt.Sprintf("%s: %v", portName, failed))
					continue
				}
				out.Copied = append(out.Copied, portName)
			}
		}
		return wireCreatedNodeInputs(inst, node, id, in.Inputs)
	})
	return nil, out, err
}

type ReplaceInstanceInput struct {
	NodeId     string `json:"nodeId" jsonschema:"the subgraph instance to swap out; it keeps this id"`
	SubgraphId string `json:"subgraphId" jsonschema:"the subgraph to place in its stead"`
	Scope      string `json:"scope,omitempty" jsonschema:"subgraph id the instance lives in; omit for the root graph"`
}

type ReplaceInstanceOutput struct {
	NodeId      string   `json:"nodeId"`
	SubgraphId  string   `json:"subgraphId"`
	KeptInputs  []string `json:"keptInputs,omitempty" jsonschema:"input ports whose outside wiring carried over"`
	KeptOutputs []string `json:"keptOutputs,omitempty" jsonschema:"'Port -> consumer.Port' edges that carried over; the producer naming it, if any, followed"`
	LostInputs  []string `json:"lostInputs,omitempty" jsonschema:"input wiring that could not carry over, with why (no such port on the new subgraph, or a type that does not fit)"`
	LostOutputs []string `json:"lostOutputs,omitempty" jsonschema:"consumer edges that could not carry over, with why; those consumer ports now read their zero value"`
	OpenInputs  []string `json:"openInputs,omitempty" jsonschema:"the new instance's inputs with nothing wired in"`
}

func (s *Server) replaceInstance(ctx context.Context, req *mcpsdk.CallToolRequest, in ReplaceInstanceInput) (*mcpsdk.CallToolResult, ReplaceInstanceOutput, error) {
	var out ReplaceInstanceOutput
	var err error
	s.atomic(&err, func() error {
		inst, e := s.resolveScope(in.Scope)
		if e != nil {
			return e
		}
		if !inst.HasNodeWithId(in.NodeId) {
			return s.explainMissingNode(in.Scope, in.NodeId, fmt.Errorf("no node exists with id %q", in.NodeId))
		}
		result, e := inst.ReplaceSubgraphInstance(in.NodeId, in.SubgraphId)
		if e != nil {
			return e
		}
		out.NodeId, out.SubgraphId = in.NodeId, in.SubgraphId
		out.KeptInputs, out.KeptOutputs = result.KeptInputs, result.KeptOutputs
		out.LostInputs, out.LostOutputs = result.LostInputs, result.LostOutputs
		for _, port := range inst.Node(in.NodeId).Inputs() {
			if single, ok := port.(nodes.SingleValueInputPort); ok && single.Value() == nil {
				out.OpenInputs = append(out.OpenInputs, portLabel(port))
			}
		}
		sort.Strings(out.OpenInputs)
		return nil
	})
	return nil, out, err
}

type RenameSubgraphInput struct {
	Id    string `json:"id"`
	NewId string `json:"newId" jsonschema:"the new id; every instance follows, and so does the outline part of the same name"`
}

type RenameSubgraphOutput struct {
	Id             string   `json:"id"`
	Instances      []string `json:"instances,omitempty" jsonschema:"'scope/nodeId' of every instance, now of the renamed subgraph"`
	OutlineRenamed bool     `json:"outlineRenamed,omitempty" jsonschema:"true when the outline had a part by the old name and it was renamed too"`
	OutlineWarning string   `json:"outlineWarning,omitempty"`
}

func renameOutlinePartFor(inst *graph.Instance, from, to string) (renamed bool, err error) {
	text := storedOutlineYAML(inst)
	if text == "" {
		return false, nil
	}
	o, err := parseOutline(text)
	if err != nil {
		return false, err
	}
	want := normalizePortKey(strings.ReplaceAll(from, "_", ""))
	name := ""
	for part := range o.Parts {
		if normalizePortKey(strings.ReplaceAll(part, "_", "")) == want {
			name = part
		}
	}
	if name == "" {
		return false, nil
	}
	if err := renameOutlinePart(&o, name, to); err != nil {
		return false, err
	}
	merged, err := yaml.Marshal(o)
	if err != nil {
		return false, err
	}
	inst.SetMetadata(outlineMetadataKey, map[string]any{"yaml": string(merged)})
	return true, nil
}

func (s *Server) renameSubgraph(ctx context.Context, req *mcpsdk.CallToolRequest, in RenameSubgraphInput) (*mcpsdk.CallToolResult, RenameSubgraphOutput, error) {
	var out RenameSubgraphOutput
	var err error
	s.atomic(&err, func() error {
		if e := s.graph.RenameSubGraph(in.Id, in.NewId); e != nil {
			return e
		}
		out.Id = in.NewId
		out.Instances = subgraphInstances(s.graph.Schema(), in.NewId)
		renamed, e := renameOutlinePartFor(s.graph, in.Id, in.NewId)
		if e != nil {
			out.OutlineWarning = fmt.Sprintf("the outline was not updated: %v", e)
			return nil
		}
		out.OutlineRenamed = renamed
		return nil
	})
	return nil, out, err
}

func (s *Server) registerSubgraphTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "create_subgraph",
		Description: "Create a subgraph definition - a named part with its own interface (e.g. a part with a Size input and a Field output) - in one call: declare 'inputs', build its interior with 'nodes' (create_nodes form, where an input port's name works as a nodeId), and wire 'outputs' from an alias. Then instantiate_subgraph places it, once or many times. All three lists are optional; with none it creates an empty definition to fill via scoped create_nodes and create_boundary_node.",
	}, s.createSubgraph)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "convert_to_subgraph",
		Description: "Move existing nodes into a new subgraph after the fact. Every connection crossing the selection becomes a boundary port, and one instance replaces the selection in place, wired exactly as before - so the graph computes the same thing but the part now has an interface. Literal parameter nodes used only by the selection move in with it (they are not ports); each variable it reads becomes an input port defaulting to the variable's value, and so do literals shared with nodes outside and real upstream nodes. Use it when a cluster of nodes you built inline turns out to be a part in its own right (a second copy is needed, or it should be tunable). Ports are named after what they carry: the variable, the feeding or producing node's own name (a subgraph instance's name, a parameter's label, a literal in its Name input), else a port name that says something; clashes get ' 2', ' 3'. Pass 'ports' to set the final names in the same call, keyed by the generated name or by the port's connection; rename_boundary_port with 'renames' fixes any left in one more call.",
	}, s.convertToSubgraph)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "rename_boundary_port",
		Description: "Rename a subgraph's input or output ports as seen on its instances. Existing wiring on every instance follows the rename. Pass 'renames' to rename several in one call (each by nodeId or current port name); they are applied as one edit, so two ports may trade names.",
	}, s.renameBoundaryPort)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "list_subgraphs",
		Description: "List every subgraph definition with where it is placed: 'placements' as scope/nodeId, and 'instances' as the live root paths (outer/inner) that evaluate_node, sample_field and describe_mesh take as instanceId.",
	}, s.listSubgraphs)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "list_nodes",
		Description: "List a scope's nodes as id, short type and name only - the cheap way to find an id or see what a subgraph holds. describe_graph has the wiring.",
	}, s.listNodes)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "rename_subgraph",
		Description: "Change a subgraph's id. Every instance follows, and an outline part of the old name is renamed with it.",
	}, s.renameSubgraph)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "replace_instance",
		Description: "Swap a subgraph instance for an instance of another subgraph, in place: the node keeps its id, and every outside connection whose port name and type match on both sides is kept - inputs fed into it and consumers reading its outputs, including a producer naming it. Connections that could not be kept come back in lostInputs/lostOutputs. Use it to move a part onto a rebuilt or corrected definition instead of delete_node + instantiate_subgraph + rewiring every consumer.",
	}, s.replaceInstance)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "delete_subgraph",
		Description: "Delete a subgraph definition that nothing instantiates any more (an abandoned approach, a superseded helper). Refused with the instance count while any instance of it still exists - delete those nodes first.",
	}, s.deleteSubgraph)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "create_boundary_node",
		Description: "Add input or output boundary ports to an existing subgraph's interface; pass 'ports' for several in one call. The port you connect to on the boundary node itself is always called 'Value' - the 'name' argument only sets the port name seen from outside, on instances. For an input, 'wireInstances' feeds every existing instance in the same call ({variable}, {value} or {nodeId, port}), and 'through' carries a port down through nested subgraphs (outer, middle, this) so a value reaches a grandchild in one call.",
	}, s.createBoundaryNode)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "instantiate_subgraph",
		Description: "Place an instance of a previously created subgraph as a node in the root graph (or another subgraph). Call multiple times to place multiple copies, e.g. 4 instances of one part. Optionally pass 'inputs' to wire its boundary inputs in the same call, or 'copyInputsFrom' (an existing instance) to inherit all of its wiring and override only what differs.",
	}, s.instantiateSubgraph)
}
