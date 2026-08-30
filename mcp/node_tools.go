package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/nodes"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// PortSummary describes a single input or output port on a node type.
type PortSummary struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	IsArray     bool   `json:"isArray,omitempty"`
	Description string `json:"description,omitempty"`
}

// NodeTypeSummary is a lightweight, search-result-sized view of a
// registered node type — enough to recognize whether it's the right one
// and to pass its Type on to get_node_types, but deliberately without full
// port lists (those can be large; fetching them for every match in a broad
// search is the main cost search_node_types used to have). Derived from
// graph.Instance.BuildSchemaForAllNodeTypes.
type NodeTypeSummary struct {
	Type        string `json:"type" jsonschema:"the exact key to pass to create_node, or to get_node_types for full port details"`
	DisplayName string `json:"displayName"`
	Path        string `json:"path"`
	Info        string `json:"info,omitempty"`
	IsParameter bool   `json:"isParameter,omitempty" jsonschema:"true if this is a literal-value parameter node; use set_parameter to assign its value"`
	InputCount  int    `json:"inputCount,omitempty" jsonschema:"number of input ports; call get_node_types to see their names/types"`
}

// NodeTypeDetail is the full view of a node type, including every port —
// what get_node_types returns for a specific type key once search has
// narrowed things down to one or two candidates.
type NodeTypeDetail struct {
	Type        string        `json:"type"`
	DisplayName string        `json:"displayName"`
	Path        string        `json:"path"`
	Info        string        `json:"info,omitempty"`
	IsParameter bool          `json:"isParameter,omitempty"`
	Inputs      []PortSummary `json:"inputs,omitempty"`
	Outputs     []PortSummary `json:"outputs,omitempty"`
}

func summarizeNodeType(nt schema.NodeType) NodeTypeSummary {
	return NodeTypeSummary{
		Type:        nt.Type,
		DisplayName: nt.DisplayName,
		Path:        nt.Path,
		Info:        nt.Info,
		IsParameter: nt.Parameter != nil,
		InputCount:  len(nt.Inputs),
	}
}

func detailNodeType(nt schema.NodeType) NodeTypeDetail {
	detail := NodeTypeDetail{
		Type:        nt.Type,
		DisplayName: nt.DisplayName,
		Path:        nt.Path,
		Info:        nt.Info,
		IsParameter: nt.Parameter != nil,
	}

	for name, in := range nt.Inputs {
		detail.Inputs = append(detail.Inputs, PortSummary{
			Name:        name,
			Type:        in.Type,
			IsArray:     in.IsArray,
			Description: in.Description,
		})
	}
	sort.Slice(detail.Inputs, func(i, j int) bool { return detail.Inputs[i].Name < detail.Inputs[j].Name })

	for name, out := range nt.Outputs {
		detail.Outputs = append(detail.Outputs, PortSummary{
			Name:        name,
			Type:        out.Type,
			Description: out.Description,
		})
	}
	sort.Slice(detail.Outputs, func(i, j int) bool { return detail.Outputs[i].Name < detail.Outputs[j].Name })

	return detail
}

// nodeTypeHaystack is what search_node_types matches a query against:
// the type key itself, display name, path, description, and every
// input/output port name — not just the description — so a query like
// "radius" finds CylinderNode even though "radius" never appears in its
// own Info text, only as a port name, and a regex query can target a
// specific generic instantiation via the type key (e.g. `\[float64\]$`).
func nodeTypeHaystack(nt schema.NodeType) string {
	var b strings.Builder
	b.WriteString(nt.Type)
	b.WriteByte(' ')
	b.WriteString(nt.DisplayName)
	b.WriteByte(' ')
	b.WriteString(nt.Path)
	b.WriteByte(' ')
	b.WriteString(nt.Info)
	for name := range nt.Inputs {
		b.WriteByte(' ')
		b.WriteString(name)
	}
	for name := range nt.Outputs {
		b.WriteByte(' ')
		b.WriteString(name)
	}
	return b.String()
}

// SearchNodeTypesInput filters the catalog of registered node types.
type SearchNodeTypesInput struct {
	Query      string `json:"query,omitempty" jsonschema:"by default, whitespace-separated substring terms (case-insensitive, all must match) against the type key, display name, path, description, and every input/output port name. If 'regex' is true, this is instead a single Go regular expression (RE2 syntax) matched against that same text, case-sensitive unless the pattern starts with (?i) — e.g. 'sphere|cylinder', or '\\[float64\\]$' to find a specific generic instantiation by its type key."`
	Regex      bool   `json:"regex,omitempty" jsonschema:"treat 'query' as a single regular expression instead of substring terms"`
	PathPrefix string `json:"pathPrefix,omitempty" jsonschema:"only return types whose path starts with this prefix, e.g. 'modeling/primitives'"`
	Limit      int    `json:"limit,omitempty" jsonschema:"maximum number of results to return; defaults to 50"`
}

type SearchNodeTypesOutput struct {
	TotalMatches int               `json:"totalMatches" jsonschema:"total number of matches, which may exceed len(results) if limit truncated the list"`
	Results      []NodeTypeSummary `json:"results" jsonschema:"lightweight summaries only — no port lists. Call get_node_types with a candidate's 'type' to see its full inputs/outputs before create_node-ing it."`
}

func (s *Server) searchNodeTypes(ctx context.Context, req *mcpsdk.CallToolRequest, in SearchNodeTypesInput) (*mcpsdk.CallToolResult, SearchNodeTypesOutput, error) {
	var out SearchNodeTypesOutput
	var err error
	s.atomic(&err, func() error {
		limit := in.Limit
		if limit <= 0 {
			limit = 50
		}

		var re *regexp.Regexp
		var terms []string
		if in.Regex {
			var e error
			re, e = regexp.Compile(in.Query)
			if e != nil {
				return fmt.Errorf("invalid regex %q: %w", in.Query, e)
			}
		} else {
			terms = strings.Fields(strings.ToLower(in.Query))
		}

		for _, nt := range s.graph.BuildSchemaForAllNodeTypes() {
			if in.PathPrefix != "" && !strings.HasPrefix(nt.Path, in.PathPrefix) {
				continue
			}

			haystack := nodeTypeHaystack(nt)
			matched := true
			if re != nil {
				matched = re.MatchString(haystack)
			} else {
				lower := strings.ToLower(haystack)
				for _, term := range terms {
					if !strings.Contains(lower, term) {
						matched = false
						break
					}
				}
			}
			if !matched {
				continue
			}

			out.TotalMatches++
			if len(out.Results) < limit {
				out.Results = append(out.Results, summarizeNodeType(nt))
			}
		}
		return nil
	})
	return nil, out, err
}

type GetNodeTypesInput struct {
	Types []string `json:"types" jsonschema:"exact type keys (the 'type' field from search_node_types results) to fetch full port details for — usually 1-3 candidates you've narrowed a search down to"`
}

type GetNodeTypesOutput struct {
	Results  []NodeTypeDetail `json:"results"`
	NotFound []string         `json:"notFound,omitempty" jsonschema:"any requested type keys that aren't registered — typo, or search_node_types first to get the exact key"`
}

func (s *Server) getNodeTypes(ctx context.Context, req *mcpsdk.CallToolRequest, in GetNodeTypesInput) (*mcpsdk.CallToolResult, GetNodeTypesOutput, error) {
	var out GetNodeTypesOutput

	byType := make(map[string]schema.NodeType)
	for _, nt := range s.graph.BuildSchemaForAllNodeTypes() {
		byType[nt.Type] = nt
	}

	for _, t := range in.Types {
		nt, ok := byType[t]
		if !ok {
			out.NotFound = append(out.NotFound, t)
			continue
		}
		out.Results = append(out.Results, detailNodeType(nt))
	}

	return nil, out, nil
}

type NodeInputSpec struct {
	NodeId   string `json:"nodeId,omitempty" jsonschema:"wire this input to an existing node's output port; mutually exclusive with 'value' and 'variable'"`
	Port     string `json:"port,omitempty" jsonschema:"the output port name on nodeId; required when nodeId is set"`
	Value    string `json:"value,omitempty" jsonschema:"literal JSON text for this input's value, e.g. \"2.5\", \"true\", \"{\\\"x\\\":1,\\\"y\\\":2,\\\"z\\\":3}\", \"\\\"#cc3333\\\"\"; a matching parameter node is created and wired in automatically. Only works for input types that have a registered literal parameter node (float64, int, bool, string, vector2/vector3 (and int variants), []vector3, geometry.AABB, coloring.Color) — for anything else (e.g. a quaternion rotation), create that node yourself and reference it via nodeId/port instead. Mutually exclusive with nodeId and variable."`
	Variable string `json:"variable,omitempty" jsonschema:"wire this input to a live reference to an existing variable, by its path (from create_variable/list_variables/create_variables); a reference node is created and wired in automatically, and updates whenever the variable does. Mutually exclusive with nodeId and value."`
}

type CreateNodeInput struct {
	Type   string                   `json:"type" jsonschema:"registered node type key, as returned by search_node_types"`
	Scope  string                   `json:"scope,omitempty" jsonschema:"subgraph id to create the node inside; omit for the root graph"`
	Inputs map[string]NodeInputSpec `json:"inputs,omitempty" jsonschema:"optionally wire this node's input ports immediately, instead of following up with separate connect_nodes/set_parameter calls. Key is the input port name. Covers single-value ports; for an array port (isArray:true from search_node_types) this wires only its first element — use connect_nodes for additional elements."`
}

type CreateNodeOutput struct {
	NodeId string `json:"nodeId"`
}

// wireCreatedNodeInputs applies a CreateNodeInput.Inputs spec to a
// freshly-created node: connecting an existing node's output, creating a
// live reference to an existing variable, or (for a literal value)
// creating and wiring a matching generator/parameter.Value[T] node — the
// same type-key construction (Value[<port's Go type>]) that only succeeds
// when that instantiation was actually registered (see
// generator/parameter/types.go), so unsupported port types fail with a
// clear error rather than silently doing nothing.
func wireCreatedNodeInputs(inst *graph.Instance, node nodes.Node, nodeID string, inputs map[string]NodeInputSpec) error {
	if len(inputs) == 0 {
		return nil
	}

	ports := node.Inputs()
	for rawPortName, spec := range inputs {
		hasRef := spec.NodeId != ""
		hasValue := spec.Value != ""
		hasVariable := spec.Variable != ""
		set := 0
		for _, b := range []bool{hasRef, hasValue, hasVariable} {
			if b {
				set++
			}
		}
		if set != 1 {
			return fmt.Errorf("input %q: specify exactly one of nodeId, value, or variable", rawPortName)
		}

		// Tolerate a guessed port name that's right except for
		// whitespace/case (e.g. "ColorTexture" for the real port "Color
		// Texture") - see resolvePortName's own doc for why this comes up.
		portName := resolveInputPortName(node, rawPortName)

		if hasRef {
			if spec.Port == "" {
				return fmt.Errorf("input %q: \"port\" is required when \"nodeId\" is set", rawPortName)
			}
			outPort := resolveOutputPortName(inst.Node(spec.NodeId), spec.Port)
			inst.ConnectNodes(spec.NodeId, outPort, nodeID, portName)
			continue
		}

		if hasVariable {
			_, varNodeID, err := inst.CreateNode(spec.Variable)
			if err != nil {
				return fmt.Errorf("input %q: no variable at path %q: %w", rawPortName, spec.Variable, err)
			}
			inst.ConnectNodes(varNodeID, "Value", nodeID, portName)
			continue
		}

		port, ok := ports[portName]
		if !ok {
			return fmt.Errorf("input %q: no such input port", rawPortName)
		}
		typed, ok := port.(nodes.Typed)
		if !ok {
			return fmt.Errorf("input %q: can't determine its type to create a literal value for it; connect an existing node instead", portName)
		}
		if !json.Valid([]byte(spec.Value)) {
			return fmt.Errorf("input %q: value is not valid JSON: %s", portName, spec.Value)
		}

		paramType := fmt.Sprintf("github.com/EliCDavis/polyform/generator/parameter.Value[%s]", typed.Type())
		_, paramID, err := inst.CreateNode(paramType)
		if err != nil {
			return fmt.Errorf("input %q (type %s): no literal parameter node available for this type — create the value yourself and connect it via nodeId/port: %w", portName, typed.Type(), err)
		}
		if _, err := inst.UpdateParameter(paramID, []byte(spec.Value)); err != nil {
			return fmt.Errorf("input %q: %w", portName, err)
		}
		inst.ConnectNodes(paramID, "Value", nodeID, portName)
	}
	return nil
}

func (s *Server) createNode(ctx context.Context, req *mcpsdk.CallToolRequest, in CreateNodeInput) (*mcpsdk.CallToolResult, CreateNodeOutput, error) {
	var out CreateNodeOutput
	var err error
	s.atomic(&err, func() error {
		inst, e := s.resolveScope(in.Scope)
		if e != nil {
			return e
		}
		node, id, e := inst.CreateNode(in.Type)
		if e != nil {
			return e
		}
		out.NodeId = id
		return wireCreatedNodeInputs(inst, node, id, in.Inputs)
	})
	return nil, out, err
}

type DeleteNodeInput struct {
	NodeId string `json:"nodeId"`
	Scope  string `json:"scope,omitempty"`
}

type DeleteNodeOutput struct {
	Deleted bool `json:"deleted"`
}

func (s *Server) deleteNode(ctx context.Context, req *mcpsdk.CallToolRequest, in DeleteNodeInput) (*mcpsdk.CallToolResult, DeleteNodeOutput, error) {
	var out DeleteNodeOutput
	var err error
	s.atomic(&err, func() error {
		inst, e := s.resolveScope(in.Scope)
		if e != nil {
			return e
		}
		inst.DeleteNodeById(in.NodeId)
		out.Deleted = true
		return nil
	})
	return nil, out, err
}

type ConnectNodesInput struct {
	OutNodeId string `json:"outNodeId" jsonschema:"id of the node providing the value"`
	OutPort   string `json:"outPort" jsonschema:"name of the output port on outNodeId"`
	InNodeId  string `json:"inNodeId" jsonschema:"id of the node receiving the value"`
	InPort    string `json:"inPort" jsonschema:"name of the input port on inNodeId"`
	Scope     string `json:"scope,omitempty" jsonschema:"subgraph id both nodes live in; omit for the root graph"`
}

type ConnectNodesOutput struct {
	Connected bool `json:"connected"`
}

func (s *Server) connectNodes(ctx context.Context, req *mcpsdk.CallToolRequest, in ConnectNodesInput) (*mcpsdk.CallToolResult, ConnectNodesOutput, error) {
	var out ConnectNodesOutput
	var err error
	s.atomic(&err, func() error {
		inst, e := s.resolveScope(in.Scope)
		if e != nil {
			return e
		}
		outPort := resolveOutputPortName(inst.Node(in.OutNodeId), in.OutPort)
		inPort := resolveInputPortName(inst.Node(in.InNodeId), in.InPort)
		inst.ConnectNodes(in.OutNodeId, outPort, in.InNodeId, inPort)
		out.Connected = true
		return nil
	})
	return nil, out, err
}

type DisconnectInput struct {
	NodeId string `json:"nodeId"`
	Port   string `json:"port" jsonschema:"input port name; append '.N' to remove a single element of an array input, e.g. 'Items.0'"`
	Scope  string `json:"scope,omitempty"`
}

type DisconnectOutput struct {
	Disconnected bool `json:"disconnected"`
}

func (s *Server) disconnect(ctx context.Context, req *mcpsdk.CallToolRequest, in DisconnectInput) (*mcpsdk.CallToolResult, DisconnectOutput, error) {
	var out DisconnectOutput
	var err error
	s.atomic(&err, func() error {
		inst, e := s.resolveScope(in.Scope)
		if e != nil {
			return e
		}
		port := resolveInputPortNameWithIndex(inst.Node(in.NodeId), in.Port)
		inst.DeleteNodeInputConnection(in.NodeId, port)
		out.Disconnected = true
		return nil
	})
	return nil, out, err
}

type SetParameterInput struct {
	NodeId string `json:"nodeId"`
	Value  string `json:"value" jsonschema:"Literal JSON text matching the parameter node's type, e.g. 2, \"red\", true, {\"x\":1,\"y\":2,\"z\":3}"`
	Scope  string `json:"scope,omitempty"`
}

type SetParameterOutput struct {
	Updated bool `json:"updated"`
}

func (s *Server) setParameter(ctx context.Context, req *mcpsdk.CallToolRequest, in SetParameterInput) (*mcpsdk.CallToolResult, SetParameterOutput, error) {
	var out SetParameterOutput
	var err error
	s.atomic(&err, func() error {
		inst, e := s.resolveScope(in.Scope)
		if e != nil {
			return e
		}
		data := []byte(in.Value)
		if !json.Valid(data) {
			return fmt.Errorf("value is not valid JSON: %s", in.Value)
		}
		ok, e := inst.UpdateParameter(in.NodeId, data)
		if e != nil {
			return e
		}
		out.Updated = ok
		return nil
	})
	return nil, out, err
}

func (s *Server) registerNodeTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "search_node_types",
		Description: "Search the catalog of available polyform node types (primitives, booleans, transforms, materials, exporters, etc.) by name, path, description, or port name. Returns lightweight summaries (no port lists) so a broad query stays cheap — follow up with get_node_types on the 1-3 candidates that look right to see their actual inputs/outputs before create_node-ing one.",
	}, s.searchNodeTypes)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "get_node_types",
		Description: "Get full input/output port details (names, types, array-ness, descriptions) for one or more specific node types by their exact type key, as returned by search_node_types. Use this after search to inspect candidates, not as a way to browse — for browsing, search_node_types' lightweight results are cheaper.",
	}, s.getNodeTypes)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "create_node",
		Description: "Instantiate a node of the given registered type in the graph (or a named subgraph). Returns the new node's id. Optionally pass 'inputs' to wire its input ports in the same call — referencing an existing node's output, a live reference to an existing variable, or a literal value (a matching parameter node is created and wired in automatically) — instead of following up with separate connect_nodes/set_parameter calls per port.",
	}, s.createNode)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "delete_node",
		Description: "Delete a node, clearing any connections other nodes had to it.",
	}, s.deleteNode)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "connect_nodes",
		Description: "Connect an output port of one node to an input port of another, wiring data flow between them.",
	}, s.connectNodes)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "disconnect",
		Description: "Remove an existing connection from a node's input port.",
	}, s.disconnect)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "set_parameter",
		Description: "Set the literal value held by a parameter node (e.g. a generator/parameter.Value[T] node created via create_node).",
	}, s.setParameter)
}
