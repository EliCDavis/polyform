package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
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

// searchAllTerms is search_node_types' default: keep only types whose
// searchable text contains every term, in catalog order.
func searchAllTerms(candidates []schema.NodeType, terms []string, limit int) (int, []NodeTypeSummary) {
	return searchFiltered(candidates, limit, func(haystack string) bool {
		lower := strings.ToLower(haystack)
		for _, term := range terms {
			if !strings.Contains(lower, term) {
				return false
			}
		}
		return true
	})
}

// searchRegex keeps types whose searchable text the pattern matches.
func searchRegex(candidates []schema.NodeType, re *regexp.Regexp, limit int) (int, []NodeTypeSummary) {
	return searchFiltered(candidates, limit, re.MatchString)
}

func searchFiltered(candidates []schema.NodeType, limit int, keep func(haystack string) bool) (int, []NodeTypeSummary) {
	total := 0
	results := make([]NodeTypeSummary, 0)
	for _, nt := range candidates {
		if !keep(nodeTypeHaystack(nt)) {
			continue
		}
		total++
		if len(results) < limit {
			results = append(results, summarizeNodeType(nt))
		}
	}
	return total, results
}

// searchAnyTerm is search_node_types' fallback ranking: keep every type
// matching at least one term, best-first. A term hitting the type key or
// display name counts double a hit buried in the description or a port
// name, so searching "wedge ramp prism" surfaces an actual wedge-shaped
// primitive ahead of something that merely mentions ramps in prose.
func searchAnyTerm(candidates []schema.NodeType, terms []string, limit int) (int, []NodeTypeSummary) {
	type scored struct {
		nt    schema.NodeType
		score int
	}

	matches := make([]scored, 0)
	for _, nt := range candidates {
		lower := strings.ToLower(nodeTypeHaystack(nt))
		name := strings.ToLower(nt.Type + " " + nt.DisplayName)

		score := 0
		for _, term := range terms {
			if !strings.Contains(lower, term) {
				continue
			}
			score++
			if strings.Contains(name, term) {
				score++
			}
		}
		if score > 0 {
			matches = append(matches, scored{nt: nt, score: score})
		}
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score > matches[j].score
		}
		return matches[i].nt.Type < matches[j].nt.Type
	})

	results := make([]NodeTypeSummary, 0, min(limit, len(matches)))
	for _, m := range matches {
		if len(results) >= limit {
			break
		}
		results = append(results, summarizeNodeType(m.nt))
	}
	return len(matches), results
}

// SearchNodeTypesInput filters the catalog of registered node types.
type SearchNodeTypesInput struct {
	Query      string `json:"query,omitempty" jsonschema:"by default, whitespace-separated substring terms (case-insensitive) against the type key, display name, path, description, and every input/output port name. Every term must match; if that yields nothing, the search automatically retries matching ANY term and ranks by how many matched, so a list of synonyms ('wedge ramp prism') finds the closest thing instead of dead-ending. If 'regex' is true, this is instead a single Go regular expression (RE2 syntax) matched against that same text. Case-insensitive (say (?-i) if you need otherwise), and whitespace is literal, so use alternation, not a space-separated list. — e.g. 'sphere|cylinder', or '\\[float64\\]$' to find a specific generic instantiation by its type key. A regex matching nothing falls back to the plain-term search."`
	Regex      bool   `json:"regex,omitempty" jsonschema:"treat 'query' as a single regular expression instead of substring terms"`
	PathPrefix string `json:"pathPrefix,omitempty" jsonschema:"only return types whose path starts with this prefix, e.g. 'modeling/primitives'"`
	Limit      int    `json:"limit,omitempty" jsonschema:"maximum number of results to return; defaults to 50"`
}

type SearchNodeTypesOutput struct {
	TotalMatches int               `json:"totalMatches" jsonschema:"total number of matches, which may exceed len(results) if limit truncated the list"`
	MatchMode    string            `json:"matchMode,omitempty" jsonschema:"set when the search had to loosen what you asked for, so check the results before use. 'substring' means the regex matched nothing and the query was retried as plain terms. 'any-term' means no node matched every term, so these match at least one, best-first. Absent for a normal all-terms or regex match."`
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
		terms := strings.Fields(strings.ToLower(in.Query))
		if in.Regex {
			var e error
			// Everything a regex matches against here is a Go identifier,
			// so it's CamelCase: a lowercase pattern like "torus" hits
			// none of TorusNode, Torus, or the type key. Case-sensitive by
			// default made every casually typed regex return zero. Wrap in
			// (?i); a caller who wants case sensitivity can say (?-i).
			re, e = regexp.Compile("(?i)(?:" + in.Query + ")")
			if e != nil {
				return fmt.Errorf("invalid regex %q: %w", in.Query, e)
			}
		}

		candidates := make([]schema.NodeType, 0)
		for _, nt := range s.graph.BuildSchemaForAllNodeTypes() {
			if in.PathPrefix != "" && !strings.HasPrefix(nt.Path, in.PathPrefix) {
				continue
			}
			candidates = append(candidates, nt)
		}

		if re != nil {
			out.TotalMatches, out.Results = searchRegex(candidates, re, limit)
		} else {
			out.TotalMatches, out.Results = searchAllTerms(candidates, terms, limit)
		}

		// A regex that matches nothing is usually a caller reaching for
		// regex mode without needing it: "torus disc" as a pattern is a
		// literal with a space in it, which occurs nowhere. Retry it as
		// ordinary terms rather than dead-ending.
		if out.TotalMatches == 0 && re != nil && len(terms) > 0 {
			out.MatchMode = "substring"
			out.TotalMatches, out.Results = searchAllTerms(candidates, terms, limit)
		}

		// Requiring every term is the right default when the caller knows
		// what they're narrowing toward ("gradient color"), but it's a
		// guaranteed dead end for the other common shape: a list of
		// synonyms hoping one lands ("wedge prism ramp pyramid"), where no
		// single node ever contains all of them. Rather than return
		// nothing, retry matching any term and rank by how many hit.
		if out.TotalMatches == 0 && len(terms) > 1 {
			out.MatchMode = "any-term"
			out.TotalMatches, out.Results = searchAnyTerm(candidates, terms, limit)
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

// NodeInputElement is one connection into an input port. Deliberately not
// the same type as NodeInputSpec: elements can't themselves hold elements,
// and the MCP SDK's schema generator rejects self-referential types.
type NodeInputElement struct {
	NodeId   string `json:"nodeId,omitempty" jsonschema:"wire this element to an existing node's output port; mutually exclusive with 'value' and 'variable'"`
	Port     string `json:"port,omitempty" jsonschema:"the output port name on nodeId; required when nodeId is set"`
	Value    string `json:"value,omitempty" jsonschema:"literal JSON text for this element's value; a matching parameter node is created and wired in automatically. Mutually exclusive with nodeId and variable."`
	Variable string `json:"variable,omitempty" jsonschema:"wire this element to a live reference to an existing variable, by its path. Mutually exclusive with nodeId and value."`
}

type NodeInputSpec struct {
	Elements []NodeInputElement `json:"elements,omitempty" jsonschema:"for an array input port (isArray:true from search_node_types) only: wire several elements at once, in order. Use this instead of one create_node plus follow-up connect_nodes calls. Mutually exclusive with nodeId, value and variable."`

	NodeId   string `json:"nodeId,omitempty" jsonschema:"wire this input to an existing node's output port; mutually exclusive with 'value' and 'variable'"`
	Port     string `json:"port,omitempty" jsonschema:"the output port name on nodeId; required when nodeId is set"`
	Value    string `json:"value,omitempty" jsonschema:"literal JSON text for this input's value, e.g. \"2.5\", \"true\", \"{\\\"x\\\":1,\\\"y\\\":2,\\\"z\\\":3}\", \"\\\"#cc3333\\\"\"; a matching parameter node is created and wired in automatically. Only works for input types that have a registered literal parameter node (float64, int, bool, string, vector2/vector3 (and int variants), []vector3, geometry.AABB, coloring.Color) — for anything else (e.g. a quaternion rotation), create that node yourself and reference it via nodeId/port instead. Mutually exclusive with nodeId and variable."`
	Variable string `json:"variable,omitempty" jsonschema:"wire this input to a live reference to an existing variable, by its path (from create_variable/list_variables/create_variables); a reference node is created and wired in automatically, and updates whenever the variable does. Mutually exclusive with nodeId and value."`
}

type CreateNodeInput struct {
	Type   string                   `json:"type" jsonschema:"registered node type key, as returned by search_node_types"`
	Scope  string                   `json:"scope,omitempty" jsonschema:"subgraph id to create the node inside; omit for the root graph"`
	Inputs map[string]NodeInputSpec `json:"inputs,omitempty" jsonschema:"wire this node's input ports immediately. Key is the port name. For an array port, pass 'elements' with one entry each."`
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

	// Every input is attempted, and every failure reported. Returning at
	// the first bad one left the rest of the entry silently unwired -
	// and since Go map iteration is random, which ports survived changed
	// between runs. One stale alias reference took a ModelNode's Mesh and
	// Rotation down with it, and the part simply didn't appear, with
	// nothing in the error naming the ports that were skipped.
	var failures []string

	ports := node.Inputs()
	for rawPortName, spec := range inputs {
		// Tolerate a guessed port name that's right except for
		// whitespace/case (e.g. "ColorTexture" for the real port "Color
		// Texture") - see resolvePortName's own doc for why this comes up.
		portName := resolveInputPortName(node, rawPortName)

		if len(spec.Elements) > 0 {
			if spec.NodeId != "" || spec.Value != "" || spec.Variable != "" {
				failures = append(failures, fmt.Sprintf("input %q: use either elements or one of nodeId/value/variable, not both", rawPortName))
				continue
			}
			if _, ok := ports[portName].(nodes.ArrayValueInputPort); !ok {
				failures = append(failures, fmt.Sprintf("input %q: elements only applies to an array input port", rawPortName))
				continue
			}
			for i, element := range spec.Elements {
				if err := wireOneInput(inst, nodeID, portName, element, ports); err != nil {
					failures = append(failures, fmt.Sprintf("input %q element %d: %v", rawPortName, i, err))
				}
			}
			continue
		}

		element := NodeInputElement{
			NodeId:   spec.NodeId,
			Port:     spec.Port,
			Value:    spec.Value,
			Variable: spec.Variable,
		}
		if err := wireOneInput(inst, nodeID, portName, element, ports); err != nil {
			failures = append(failures, fmt.Sprintf("input %q: %v", rawPortName, err))
		}
	}

	if len(failures) > 0 {
		sort.Strings(failures)
		return fmt.Errorf("%s (every other input on this node was wired)", strings.Join(failures, "; "))
	}
	return nil
}

// wireOneInput makes a single connection into portName. Called once per
// input, or once per element for an array port.
func wireOneInput(
	inst *graph.Instance,
	nodeID string,
	portName string,
	spec NodeInputElement,
	ports map[string]nodes.InputPort,
) error {
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
		return fmt.Errorf("specify exactly one of nodeId, value, or variable")
	}

	if hasRef {
		if spec.Port == "" {
			return fmt.Errorf("\"port\" is required when \"nodeId\" is set")
		}
		outPort := resolveOutputPortName(inst.Node(spec.NodeId), spec.Port)
		inst.ConnectNodes(spec.NodeId, outPort, nodeID, portName)
		return nil
	}

	if hasVariable {
		_, varNodeID, err := inst.CreateNode(spec.Variable)
		if err != nil {
			return fmt.Errorf("no variable at path %q: %w", spec.Variable, err)
		}
		inst.ConnectNodes(varNodeID, "Value", nodeID, portName)
		return nil
	}

	port, ok := ports[portName]
	if !ok {
		return fmt.Errorf("no such input port")
	}
	typed, ok := port.(nodes.Typed)
	if !ok {
		return fmt.Errorf("can't determine its type to create a literal value for it; connect an existing node instead")
	}
	if !json.Valid([]byte(spec.Value)) {
		return fmt.Errorf("value is not valid JSON: %s", spec.Value)
	}

	paramType := fmt.Sprintf("github.com/EliCDavis/polyform/generator/parameter.Value[%s]", typed.Type())
	_, paramID, err := inst.CreateNode(paramType)
	if err != nil {
		return fmt.Errorf("type %s: no literal parameter node available for this type — create the value yourself and connect it via nodeId/port: %w", typed.Type(), err)
	}
	if _, err := inst.UpdateParameter(paramID, []byte(spec.Value)); err != nil {
		return err
	}
	inst.ConnectNodes(paramID, "Value", nodeID, portName)
	return nil
}

// explainCreateFailure turns a bare factory miss into the answer the
// caller needs. A subgraph is not a registered node type - it has to go
// through instantiate_subgraph - but reaching for create_nodes with a
// subgraph id is a natural mistake once every other node in the batch is
// being created that way, and "no factory registered with ID x" gives no
// hint that the id was right and only the tool was wrong.
func (s *Server) explainCreateFailure(nodeType string, err error) error {
	if err == nil {
		return nil
	}

	subgraphs := s.graph.Schema().SubGraphs
	if _, isSubgraph := subgraphs[nodeType]; isSubgraph {
		return fmt.Errorf("%q is a subgraph, not a node type: use instantiate_subgraph to place it, then wire the node id it returns (aliases in this batch can't name it)", nodeType)
	}

	// A near miss on a subgraph id is worth naming too, since the ids are
	// the caller's own and easy to mistype.
	for id := range subgraphs {
		if strings.EqualFold(id, nodeType) {
			return fmt.Errorf("%q is a subgraph (spelled %q), not a node type: use instantiate_subgraph to place it, then wire the node id it returns", nodeType, id)
		}
	}

	return err
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
			return s.explainCreateFailure(in.Type, e)
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

// NodeConnection is one edge in a connect_nodes batch.
type NodeConnection struct {
	OutNodeId string `json:"outNodeId" jsonschema:"id of the node providing the value"`
	OutPort   string `json:"outPort" jsonschema:"name of the output port on outNodeId"`
	InNodeId  string `json:"inNodeId" jsonschema:"id of the node receiving the value"`
	InPort    string `json:"inPort" jsonschema:"name of the input port on inNodeId"`
	Scope     string `json:"scope,omitempty" jsonschema:"subgraph id both nodes live in; omit to use the call's top-level scope"`
}

type ConnectNodesInput struct {
	Connections []NodeConnection `json:"connections,omitempty" jsonschema:"wire several edges in one call, in order. Preferred over one call per edge. When set, the single-edge fields are ignored."`

	OutNodeId string `json:"outNodeId,omitempty" jsonschema:"id of the node providing the value"`
	OutPort   string `json:"outPort,omitempty" jsonschema:"name of the output port on outNodeId"`
	InNodeId  string `json:"inNodeId,omitempty" jsonschema:"id of the node receiving the value"`
	InPort    string `json:"inPort,omitempty" jsonschema:"input port on inNodeId. A single-value port is replaced; an array port appends. Use disconnect with 'Port.N' to remove one element."`
	Scope     string `json:"scope,omitempty" jsonschema:"subgraph id both nodes live in; omit for the root graph"`
}

type ConnectNodesOutput struct {
	Connected bool     `json:"connected" jsonschema:"true when every requested connection was made"`
	Made      int      `json:"made,omitempty" jsonschema:"how many connections succeeded, when a batch was passed"`
	Errors    []string `json:"errors,omitempty" jsonschema:"per-connection failures naming the index. Connections outside these were still made."`
}

func (s *Server) connectNodes(ctx context.Context, req *mcpsdk.CallToolRequest, in ConnectNodesInput) (*mcpsdk.CallToolResult, ConnectNodesOutput, error) {
	var out ConnectNodesOutput
	var err error
	s.atomic(&err, func() error {
		// A single-edge call keeps the strict contract: a bad port is a
		// tool error. Only a batch downgrades failures to per-entry
		// reports, since there the rest of the batch is worth keeping.
		batched := len(in.Connections) > 0

		connections := in.Connections
		if len(connections) == 0 {
			if in.OutNodeId == "" && in.InNodeId == "" {
				return fmt.Errorf("pass either 'connections' or a single outNodeId/outPort/inNodeId/inPort")
			}
			connections = []NodeConnection{{
				OutNodeId: in.OutNodeId,
				OutPort:   in.OutPort,
				InNodeId:  in.InNodeId,
				InPort:    in.InPort,
				Scope:     in.Scope,
			}}
		}

		for i, c := range connections {
			scope := c.Scope
			if scope == "" {
				scope = in.Scope
			}

			// One bad edge shouldn't discard the rest of the batch: the
			// caller would have to re-derive which of them landed.
			e := func() (e error) {
				defer func() {
					if r := recover(); r != nil {
						e = fmt.Errorf("%v", r)
					}
				}()
				inst, e := s.resolveScope(scope)
				if e != nil {
					return e
				}
				outPort := resolveOutputPortName(inst.Node(c.OutNodeId), c.OutPort)
				inPort := resolveInputPortName(inst.Node(c.InNodeId), c.InPort)
				inst.ConnectNodes(c.OutNodeId, outPort, c.InNodeId, inPort)
				return nil
			}()
			if e != nil {
				if !batched {
					return e
				}
				out.Errors = append(out.Errors, fmt.Sprintf("connection %d (%s.%s -> %s.%s): %v",
					i, c.OutNodeId, c.OutPort, c.InNodeId, c.InPort, e))
				continue
			}
			out.Made++
		}
		out.Connected = len(out.Errors) == 0
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

// ParameterAssignment is one value to set in a set_parameter batch.
type ParameterAssignment struct {
	NodeId string `json:"nodeId" jsonschema:"the parameter node's id, or with 'port', the node whose port is being set"`
	Port   string `json:"port,omitempty" jsonschema:"input port on nodeId; sets whatever feeds it, creating a literal if empty. Append '.N' for one element of an array port, e.g. 'Values.1'. Omit to address a parameter node by its own id."`
	Value  string `json:"value" jsonschema:"Literal JSON text matching the parameter node's type, e.g. 2, \"red\", true, {\"x\":1,\"y\":2,\"z\":3}"`
	Scope  string `json:"scope,omitempty" jsonschema:"subgraph id this node lives in; omit to use the call's top-level scope"`
}

type SetParameterInput struct {
	Parameters []ParameterAssignment `json:"parameters,omitempty" jsonschema:"set several values in one call, in order. Preferred over one call per value. When set, the single-value fields are ignored."`

	NodeId string `json:"nodeId,omitempty" jsonschema:"id of the parameter node to set, or - with 'port' also given - the node whose port is being set"`
	Port   string `json:"port,omitempty" jsonschema:"input port name on nodeId. Sets whatever parameter node feeds that port, so you do not have to know its id; if nothing feeds it yet, a literal is created and wired in. Omit to address a parameter node directly by its own id."`
	Value  string `json:"value,omitempty" jsonschema:"Literal JSON text matching the parameter node's type, e.g. 2, \"red\", true, {\"x\":1,\"y\":2,\"z\":3}"`
	Scope  string `json:"scope,omitempty" jsonschema:"subgraph id the nodes live in; omit for the root graph"`
}

type SetParameterOutput struct {
	Updated bool     `json:"updated" jsonschema:"true when every requested value was set"`
	Set     int      `json:"set,omitempty" jsonschema:"how many values were set, when a batch was passed"`
	Errors  []string `json:"errors,omitempty" jsonschema:"per-assignment failures naming the index. Assignments outside these still applied."`
}

// parameterFeedingPort returns the id of the parameter node wired into
// nodeID's portName. When the port is unconnected it creates a literal
// holding value, wires it in, and returns "" to report that the assignment
// is already done.
//
// This exists because create_node and create_nodes quietly create a
// literal parameter node for any input given as a value, and never report
// that node's id. Without addressing by port, changing such a value costs
// a describe_graph first just to go find the id.
func parameterFeedingPort(inst *graph.Instance, nodeID, portName, value string) (string, error) {
	node := inst.Node(nodeID)
	if node == nil {
		return "", fmt.Errorf("no node exists with id %q", nodeID)
	}

	inputs := node.Inputs()

	// "Values.1" addresses one element of an array port, the same spelling
	// disconnect takes. Without it, changing a single element's literal
	// meant disconnect-by-index, create a fresh literal, reconnect - three
	// calls for what reads like one.
	base, index, hasIndex := splitPortIndex(portName)
	resolved := resolveInputPortName(node, base)
	port, ok := inputs[resolved]
	if !ok {
		return "", fmt.Errorf("no input port %q; it has %s", base, strings.Join(sortedPortNames(inputs), ", "))
	}

	if array, isArray := port.(nodes.ArrayValueInputPort); isArray {
		if !hasIndex {
			return "", fmt.Errorf("port %q holds an array of connections; address one element as %q, or set the element's own parameter node by its id", resolved, resolved+".0")
		}
		elements := array.Value()
		if index >= len(elements) {
			return "", fmt.Errorf("port %q has %d element(s), so there is no index %d", resolved, len(elements), index)
		}
		upstream := elements[index]
		if upstream == nil {
			return "", fmt.Errorf("nothing is wired into %s.%d", resolved, index)
		}
		return parameterNodeID(inst, upstream, fmt.Sprintf("%s.%d", resolved, index))
	}

	if hasIndex {
		return "", fmt.Errorf("port %q is not an array, so %q has no element to address", resolved, portName)
	}

	single, ok := port.(nodes.SingleValueInputPort)
	if !ok {
		return "", fmt.Errorf("port %q is neither a single value nor an array", resolved)
	}

	upstream := single.Value()
	if upstream == nil {
		if e := wireOneInput(inst, nodeID, resolved, NodeInputElement{Value: value}, inputs); e != nil {
			return "", e
		}
		return "", nil
	}

	return parameterNodeID(inst, upstream, resolved)
}

// parameterNodeID names the literal behind a port, or explains why there
// isn't one.
func parameterNodeID(inst *graph.Instance, upstream nodes.OutputPort, portLabel string) (string, error) {
	upstreamID := inst.NodeId(upstream.Node())
	if _, isParam := upstream.Node().(graph.Parameter); !isParam {
		return "", fmt.Errorf("port %q is fed by node %q, which computes its value rather than holding a literal; set the parameter feeding that node instead, or disconnect the port first", portLabel, upstreamID)
	}
	return upstreamID, nil
}

// splitPortIndex separates a trailing ".N" array index from a port name.
func splitPortIndex(portName string) (base string, index int, ok bool) {
	name, suffix, found := strings.Cut(portName, ".")
	if !found {
		return portName, 0, false
	}
	i, err := strconv.Atoi(suffix)
	if err != nil || i < 0 {
		// Not an index - the port name itself contains a dot.
		return portName, 0, false
	}
	return name, i, true
}

func sortedPortNames(ports map[string]nodes.InputPort) []string {
	names := make([]string, 0, len(ports))
	for name := range ports {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func setOneParameter(inst *graph.Instance, nodeID, portName, value string) (bool, error) {
	data := []byte(value)
	if !json.Valid(data) {
		return false, fmt.Errorf("value is not valid JSON: %s", value)
	}

	targetID := nodeID
	if portName != "" {
		id, e := parameterFeedingPort(inst, nodeID, portName, value)
		if e != nil {
			return false, e
		}
		if id == "" {
			// Port was unconnected; a literal holding value was wired in.
			return true, nil
		}
		targetID = id
	}

	if inst.Node(targetID) == nil {
		return false, fmt.Errorf("no node exists with id %q", targetID)
	}
	return inst.UpdateParameter(targetID, data)
}

func (s *Server) setParameter(ctx context.Context, req *mcpsdk.CallToolRequest, in SetParameterInput) (*mcpsdk.CallToolResult, SetParameterOutput, error) {
	var out SetParameterOutput
	var err error
	s.atomic(&err, func() error {
		// A single assignment keeps the strict contract: a bad node or
		// port is a tool error. Only a batch downgrades failures to
		// per-entry reports, since there the rest is worth keeping.
		batched := len(in.Parameters) > 0

		assignments := in.Parameters
		if !batched {
			if in.NodeId == "" {
				return fmt.Errorf("pass either 'parameters' or a single nodeId/value")
			}
			assignments = []ParameterAssignment{{
				NodeId: in.NodeId,
				Port:   in.Port,
				Value:  in.Value,
				Scope:  in.Scope,
			}}
		}

		for i, a := range assignments {
			scope := a.Scope
			if scope == "" {
				scope = in.Scope
			}

			changed, e := func() (changed bool, e error) {
				defer func() {
					if r := recover(); r != nil {
						e = fmt.Errorf("%v", r)
					}
				}()
				inst, e := s.resolveScope(scope)
				if e != nil {
					return false, e
				}
				return setOneParameter(inst, a.NodeId, a.Port, a.Value)
			}()
			if e != nil {
				if !batched {
					return e
				}
				target := a.NodeId
				if a.Port != "" {
					target += "." + a.Port
				}
				out.Errors = append(out.Errors, fmt.Sprintf("parameter %d (%s): %v", i, target, e))
				continue
			}
			out.Set++
			if !batched {
				out.Updated = changed
			}
		}
		if batched {
			out.Updated = len(out.Errors) == 0
		}
		return nil
	})
	return nil, out, err
}

func (s *Server) registerNodeTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "search_node_types",
		Description: "Search available node types by name, path, description, or port name. Returns lightweight summaries without port lists; follow up with get_node_types on the 1-3 best candidates to see actual ports. Common type keys are already listed in your reference docs - check those before searching.",
	}, s.searchNodeTypes)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "get_node_types",
		Description: "Get full input/output port details (names, types, array-ness, descriptions) for one or more specific node types by their exact type key, as returned by search_node_types. Use this after search to inspect candidates, not as a way to browse — for browsing, search_node_types' lightweight results are cheaper.",
	}, s.getNodeTypes)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "create_node",
		Description: "Create one node of a registered type in the graph (or a named subgraph); returns its id. To create several, use create_nodes instead. Optionally pass 'inputs' to wire its ports in the same call: an existing node's output, a variable reference, or a literal value (a parameter node is made for you).",
	}, s.createNode)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "create_nodes",
		Description: "Create many nodes at once and wire them together in one call. Use this, not repeated create_node, whenever placing more than one node. Give each entry an 'alias'; any entry's 'inputs' may reference another entry's alias in place of a node id, in either direction. Returns the alias-to-id map. A failing entry does not stop the batch - check 'errors'.",
	}, s.createNodes)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "delete_node",
		Description: "Delete a node, clearing any connections other nodes had to it.",
	}, s.deleteNode)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "connect_nodes",
		Description: "Connect one node's output port to another's input port. Connecting into an array port appends rather than replaces, so repeated calls build it up in order. A connection that would create a cycle is refused. Pass 'connections' to make many in one call; the single-edge fields are then ignored.",
	}, s.connectNodes)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "disconnect",
		Description: "Remove an existing connection from a node's input port.",
	}, s.disconnect)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "set_parameter",
		Description: "Set the literal value held by a parameter node, addressed by its own node id, or by 'nodeId' plus 'port' to set whatever feeds that port (wiring a literal in if the port is empty). Pass 'parameters' to set many values in one call; the single-value fields are then ignored.",
	}, s.setParameter)
}

// CreateNodesEntry is one node in a create_nodes batch.
type CreateNodesEntry struct {
	Alias  string                   `json:"alias,omitempty" jsonschema:"a short name you pick so other entries can wire to it before it has a real id. Unique within the batch. Optional; it also labels this entry in the returned map and in errors."`
	Type   string                   `json:"type" jsonschema:"registered node type key, as returned by search_node_types"`
	Inputs map[string]NodeInputSpec `json:"inputs,omitempty" jsonschema:"same shape as create_node's 'inputs'. A 'nodeId' may be a real node id or any entry's alias, including an entry listed later."`
}

type CreateNodesInput struct {
	Scope string             `json:"scope,omitempty" jsonschema:"subgraph id to create every node inside; omit for the root graph"`
	Nodes []CreateNodesEntry `json:"nodes" jsonschema:"the nodes to create, in any order"`
}

type CreateNodesOutput struct {
	Nodes  map[string]string `json:"nodes" jsonschema:"alias (or the entry's index as a string, when it had no alias) mapped to the created node id"`
	Errors []string          `json:"errors,omitempty" jsonschema:"per-entry failures, naming the entry. ALWAYS check: entries outside these still applied."`
}

// createNodes is create_node's batch form, and exists purely to cut turns.
// A node id is only known after the call that made it, so wiring a graph
// one create_node at a time is forced to serialize: measured over real
// builds, ~93 create_node calls landed across ~55 separate turns, each one
// re-sending the whole conversation. Aliases break that dependency - every
// node in the batch is created first, so entries can reference each other
// in any order, and a whole part collapses into one call.
//
// Per-entry failures are reported in the output rather than returned as a
// tool error, because an error result discards structured content: a batch
// that failed on entry 30 would otherwise take the ids of the 29 that
// worked down with it, and the caller would have no way to find them
// again short of describe_graph.
func (s *Server) createNodes(ctx context.Context, req *mcpsdk.CallToolRequest, in CreateNodesInput) (*mcpsdk.CallToolResult, CreateNodesOutput, error) {
	var out CreateNodesOutput
	var err error
	s.atomic(&err, func() error {
		if len(in.Nodes) == 0 {
			return fmt.Errorf("\"nodes\" is empty; pass at least one node to create")
		}

		inst, e := s.resolveScope(in.Scope)
		if e != nil {
			return e
		}

		label := func(i int) string {
			if in.Nodes[i].Alias != "" {
				return in.Nodes[i].Alias
			}
			return strconv.Itoa(i)
		}

		seen := make(map[string]int, len(in.Nodes))
		for i, entry := range in.Nodes {
			if entry.Alias == "" {
				continue
			}
			if prev, dup := seen[entry.Alias]; dup {
				return fmt.Errorf("alias %q used by both entry %d and entry %d; aliases must be unique within a batch", entry.Alias, prev, i)
			}
			seen[entry.Alias] = i
		}

		out.Nodes = make(map[string]string, len(in.Nodes))
		created := make([]nodes.Node, len(in.Nodes))
		ids := make([]string, len(in.Nodes))

		for i, entry := range in.Nodes {
			node, id, e := inst.CreateNode(entry.Type)
			if e != nil {
				out.Errors = append(out.Errors, fmt.Sprintf("entry %s: %v", label(i), s.explainCreateFailure(entry.Type, e)))
				continue
			}
			created[i], ids[i] = node, id
			out.Nodes[label(i)] = id
		}

		// Second pass: every id exists now, so an entry can reference an
		// alias regardless of where it sits in the list.
		for i, entry := range in.Nodes {
			if created[i] == nil {
				continue
			}

			// Wiring reaches inst.ConnectNodes, which panics rather than
			// returning an error for things like an unknown node id. Left
			// unrecovered that panic escapes to s.atomic and fails the
			// whole call, which discards the structured content - taking
			// down every id the batch just created, the exact outcome the
			// per-entry errors field exists to avoid.
			e := func() (e error) {
				defer func() {
					if r := recover(); r != nil {
						e = fmt.Errorf("%v", r)
					}
				}()
				// Wire whatever resolved even when some references
				// didn't, so one stale alias can't silently strip a node
				// of every other input it was given.
				resolved, resolveErr := resolveAliasedInputs(entry.Inputs, out.Nodes, seen, inst)
				wireErr := wireCreatedNodeInputs(inst, created[i], ids[i], resolved)

				problems := []string{}
				if resolveErr != nil {
					problems = append(problems, resolveErr.Error())
				}
				if wireErr != nil {
					problems = append(problems, wireErr.Error())
				}
				if len(problems) == 0 {
					return nil
				}
				return fmt.Errorf("%s (every other input on this node was wired)",
					strings.Join(problems, "; "))
			}()
			if e != nil {
				out.Errors = append(out.Errors, fmt.Sprintf("entry %s: %v", label(i), e))
			}
		}
		return nil
	})
	return nil, out, err
}

// resolveAliasedInputs rewrites any nodeId naming a batch alias into the
// real id that alias was given. A nodeId matching no alias is passed
// through untouched, so referencing nodes that already existed before the
// batch keeps working.
func resolveAliasedInputs(
	inputs map[string]NodeInputSpec,
	aliases map[string]string,
	declared map[string]int,
	inst *graph.Instance,
) (map[string]NodeInputSpec, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	// A bad reference drops only its own input. Rejecting the whole map
	// left every other port on the node unwired - including ones with
	// nothing wrong with them - and the part then silently failed to
	// appear rather than erroring.
	var failures []string

	resolved := make(map[string]NodeInputSpec, len(inputs))
	for portName, spec := range inputs {
		id, e := resolveOneReference(spec.NodeId, aliases, declared, inst)
		if e != nil {
			failures = append(failures, fmt.Sprintf("input %q: %v", portName, e))
			continue
		}
		spec.NodeId = id

		if len(spec.Elements) > 0 {
			elements := make([]NodeInputElement, 0, len(spec.Elements))
			for i, element := range spec.Elements {
				id, e := resolveOneReference(element.NodeId, aliases, declared, inst)
				if e != nil {
					failures = append(failures, fmt.Sprintf("input %q element %d: %v", portName, i, e))
					continue
				}
				element.NodeId = id
				elements = append(elements, element)
			}
			spec.Elements = elements
		}
		resolved[portName] = spec
	}

	if len(failures) > 0 {
		sort.Strings(failures)
		return resolved, fmt.Errorf("%s", strings.Join(failures, "; "))
	}
	return resolved, nil
}

// resolveOneReference turns one nodeId reference into a real node id,
// distinguishing the three ways it can go wrong. Left to the graph layer
// all three collapse into the same "no node exists with id X" panic, which
// names the alias the caller wrote and so reads as though aliases don't
// work at all.
func resolveOneReference(ref string, aliases map[string]string, declared map[string]int, inst *graph.Instance) (string, error) {
	if ref == "" {
		return ref, nil
	}
	if id, ok := aliases[ref]; ok {
		return id, nil
	}
	if i, ok := declared[ref]; ok {
		return "", fmt.Errorf("references alias %q, but entry %d failed to be created - see the other errors", ref, i)
	}
	if inst.HasNodeWithId(ref) {
		return ref, nil
	}
	return "", fmt.Errorf("%q is neither a node in this graph nor an alias in this batch. Aliases only last for the one call that declares them; to reference a node from an earlier call, use the real id that call returned", ref)
}
