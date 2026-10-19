package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
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
	for _, keyword := range nt.Keywords {
		b.WriteByte(' ')
		b.WriteString(keyword)
	}
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
	Elements []NodeInputElement `json:"elements,omitempty" jsonschema:"for an array input port (isArray:true from search_node_types) only: wire several elements at once, in order. If any element fails, none are wired. Use this instead of one create_node plus follow-up connect_nodes calls. Mutually exclusive with nodeId, value and variable."`

	NodeId   string `json:"nodeId,omitempty" jsonschema:"wire this input to an existing node's output port; mutually exclusive with 'value' and 'variable'"`
	Port     string `json:"port,omitempty" jsonschema:"the output port name on nodeId; required when nodeId is set"`
	Value    string `json:"value,omitempty" jsonschema:"literal JSON text for this input's value, e.g. \"2.5\", \"true\", \"{\\\"x\\\":1,\\\"y\\\":2,\\\"z\\\":3}\", \"\\\"#cc3333\\\"\"; a matching parameter node is created and wired in automatically. Only works for input types that have a registered literal parameter node: float64, int, bool, string, vector2/3/4 (and int variants), []float64/[]int/[]string, []vector2, []vector3, quaternion.Quaternion, trs.TRS, []trs.TRS, coloring.Color, []coloring.Color, coloring.Gradient, geometry.AABB — for anything else (a mesh, a material), create that node yourself and reference it via nodeId/port instead. Mutually exclusive with nodeId and variable."`
	Variable string `json:"variable,omitempty" jsonschema:"wire this input to a live reference to an existing variable, by its path (from create_variable/list_variables/create_variables); a reference node is created and wired in automatically, and updates whenever the variable does. Mutually exclusive with nodeId and value."`
}

type CreateNodeInput struct {
	Type   string                   `json:"type" jsonschema:"registered node type key, as returned by search_node_types"`
	Scope  string                   `json:"scope,omitempty" jsonschema:"subgraph id to create the node inside; omit for the root graph"`
	Inputs map[string]NodeInputSpec `json:"inputs,omitempty" jsonschema:"wire this node's input ports immediately. Key is the port name. For an array port, pass 'elements' with one entry each."`
}

type CreateNodeOutput struct {
	NodeId   string            `json:"nodeId"`
	Literals map[string]string `json:"literals,omitempty" jsonschema:"'Port' (or 'Port.N') -> the id of the literal parameter node a 'value' input created, for set_parameter/delete_node later"`
}

// wireCreatedNodeInputs applies a CreateNodeInput.Inputs spec to a
// freshly-created node: connecting an existing node's output, creating a
// live reference to an existing variable, or (for a literal value)
// creating and wiring a matching generator/parameter.Value[T] node — the
// same type-key construction (Value[<port's Go type>]) that only succeeds
// when that instantiation was actually registered (see
// generator/parameter/types.go), so unsupported port types fail with a
// clear error rather than silently doing nothing.
func wireCreatedNodeInputs(inst *graph.Graph, node nodes.Node, nodeID string, inputs map[string]NodeInputSpec) error {
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
	bindFromReferences(inst, node, ports, inputs)
	pool := literalPool{}
	// Sorted: the literal nodes this creates take ids in this order.
	for _, rawPortName := range slices.Sorted(maps.Keys(inputs)) {
		spec := inputs[rawPortName]
		// Tolerate a guessed port name that's right except for
		// whitespace/case (e.g. "ColorTexture" for the real port "Color
		// Texture") - see resolvePortName's own doc for why this comes up.
		portName := resolveInputPortName(node, rawPortName)

		if base, _, indexed := splitPortIndex(node, rawPortName); indexed {
			if _, isPort := ports[portName]; !isPort {
				failures = append(failures, fmt.Sprintf("input %q: a new node has no elements to index; pass %q with 'elements' listing every element in order", rawPortName, resolveInputPortName(node, base)))
				continue
			}
		}

		if len(spec.Elements) > 0 {
			if spec.NodeId != "" || spec.Value != "" || spec.Variable != "" {
				failures = append(failures, fmt.Sprintf("input %q: use either elements or one of nodeId/value/variable, not both", rawPortName))
				continue
			}
			if _, ok := ports[portName].(nodes.ArrayValueInputPort); !ok {
				failures = append(failures, fmt.Sprintf("input %q: elements only applies to a port that takes many connections (isArray:true, like Meshes[]); this port takes one connection%s, so pass the whole list as a single value or variable", rawPortName, describePortType(ports[portName])))
				continue
			}
			existing := map[string]bool{}
			for _, id := range inst.NodeIds() {
				existing[id] = true
			}
			var elementFailures []string
			for i, element := range spec.Elements {
				if err := wireOneInput(inst, nodeID, portName, element, ports, pool); err != nil {
					elementFailures = append(elementFailures, fmt.Sprintf("input %q element %d: %v", rawPortName, i, err))
				}
			}
			if len(elementFailures) > 0 {
				_, _ = inst.DeleteNodeInputConnection(nodeID, portName)
				for _, id := range inst.NodeIds() {
					if !existing[id] {
						_, _ = inst.DeleteNodeById(id)
					}
				}
				failures = append(failures, elementFailures...)
				failures = append(failures, noElementWired(rawPortName))
			}
			continue
		}

		if spec.Value != "" && spec.NodeId == "" && spec.Variable == "" && literalIsPortDefault(node, portName, spec.Value) {
			continue
		}

		element := NodeInputElement{
			NodeId:   spec.NodeId,
			Port:     spec.Port,
			Value:    spec.Value,
			Variable: spec.Variable,
		}
		if err := wireOneInput(inst, nodeID, portName, element, ports, pool); err != nil {
			failures = append(failures, fmt.Sprintf("input %q: %v", rawPortName, err))
		}
	}

	if len(failures) > 0 {
		sort.Strings(failures)
		return fmt.Errorf("%s (every other input on this node was wired)", strings.Join(failures, "; "))
	}
	return nil
}

func noElementWired(port string) string {
	return fmt.Sprintf("input %q: no element was wired, so none moved into another's place", port)
}

// bindFromReferences gives each unbound dynamic input the type of a node
// wired into it, so a literal beside that reference has a type to parse as
// whatever order they are listed in. Settling the graph derives it again.
func bindFromReferences(inst *graph.Graph, node nodes.Node, ports map[string]nodes.InputPort, inputs map[string]NodeInputSpec) {
	for rawPortName, spec := range inputs {
		dynamic, ok := ports[resolveInputPortName(node, rawPortName)].(nodes.DynamicallyTypedPort)
		if !ok || portType(dynamic) != "" {
			continue
		}
		references := append([]NodeInputElement{{NodeId: spec.NodeId, Port: spec.Port}}, spec.Elements...)
		for _, reference := range references {
			upstream := lookupNode(inst, reference.NodeId)
			if reference.NodeId == "" || upstream == nil {
				continue
			}
			if concrete := portType(upstream.Outputs()[resolveOutputPortName(upstream, reference.Port)]); concrete != "" && dynamic.BindType(concrete) == nil {
				break
			}
		}
	}
}

func portType(port any) string {
	if typed, ok := port.(nodes.Typed); ok {
		return typed.Type()
	}
	return ""
}

func describePortType(port nodes.InputPort) string {
	typed, ok := port.(nodes.Typed)
	if !ok {
		return ""
	}
	return fmt.Sprintf(" of type %s", typed.Type())
}

// literalPool shares one parameter node between elements of the same array
// port that ask for the same literal of the same type. Keyed by port, type
// and the compacted JSON text; nil disables sharing.
//
// The port is part of the key because two ports holding equal values today
// are not the same number: pooling a NoiseNode's Width and Height at 512
// left one literal driving both, so changing the width later silently
// changed the height too.
type literalPool map[string]string

// literalIsPortDefault reports a literal the port already yields when
// left unconnected, so wiring it would only add a node: a zero component
// on a vector2/3/4 NewNode, a zero Position or identity Rotation on
// trs.NewNode.
func literalIsPortDefault(node nodes.Node, port, value string) bool {
	typeName := fmt.Sprintf("%T", node)
	switch {
	case strings.Contains(typeName, "vector2.NewNode") || strings.Contains(typeName, "vector3.NewNode") || strings.Contains(typeName, "vector4.NewNode"):
		switch port {
		case "X", "Y", "Z", "W":
			var n float64
			return json.Unmarshal([]byte(value), &n) == nil && n == 0
		}
	case strings.Contains(typeName, "trs.NewNode"):
		switch port {
		case "Position":
			var v vector3.Float64
			return json.Unmarshal([]byte(value), &v) == nil && v == vector3.Zero[float64]()
		case "Rotation":
			var q quaternion.Quaternion
			return json.Unmarshal([]byte(value), &q) == nil && q == quaternion.Identity()
		}
	}
	return false
}

// variableReferenceNode returns a node in inst that reads the variable at
// path, creating one only if none exists. A second reference to the same
// variable in the same scope is never wanted: it reads the same value and
// only adds a node.
func variableReferenceNode(inst *graph.Graph, path string) (string, error) {
	if target := lookupVariable(inst, path); target != nil {
		for _, id := range inst.NodeIds() {
			if ref, ok := inst.Node(id).(variable.Reference); ok && ref.Reference() == target {
				return id, nil
			}
		}
	}
	_, id, err := inst.CreateNode(path)
	if err != nil {
		return "", fmt.Errorf("no variable at path %q: %w", path, err)
	}
	return id, nil
}

func lookupVariable(inst *graph.Graph, path string) variable.Variable {
	v, err := inst.Root().GetVariable(path)
	if err != nil {
		return nil
	}
	return v
}

// canonicalJSON renders equal values identically (1e-4 and 0.0001, keys
// in any order) so they can share one literal.
func canonicalJSON(text string) string {
	var v any
	if err := json.Unmarshal([]byte(text), &v); err != nil {
		return text
	}
	out, err := json.Marshal(v)
	if err != nil {
		return text
	}
	return string(out)
}

// wireOneInput makes a single connection into portName. Called once per
// input, or once per element for an array port.
func wireOneInput(
	inst *graph.Graph,
	nodeID string,
	portName string,
	spec NodeInputElement,
	ports map[string]nodes.InputPort,
	pool literalPool,
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
		if _, ok := ports[portName]; !ok {
			return fmt.Errorf("no such input port; it has %s", strings.Join(sortedPortNames(ports), ", "))
		}
		upstream := lookupNode(inst, spec.NodeId)
		if upstream == nil {
			return fmt.Errorf("no node exists with id %q", spec.NodeId)
		}
		outPort := resolveOutputPortName(upstream, spec.Port)
		if _, ok := upstream.Outputs()[outPort]; !ok {
			return fmt.Errorf("node %q has no output port %q; it has %s", spec.NodeId, spec.Port, strings.Join(sortedOutputNames(upstream.Outputs()), ", "))
		}
		return inst.ConnectNodes(spec.NodeId, outPort, nodeID, portName)
	}

	if hasVariable {
		if _, ok := ports[portName]; !ok {
			return fmt.Errorf("no such input port; it has %s", strings.Join(sortedPortNames(ports), ", "))
		}
		varNodeID, err := variableReferenceNode(inst, spec.Variable)
		if err != nil {
			return err
		}
		return inst.ConnectNodes(varNodeID, "Value", nodeID, portName)
	}

	port, ok := ports[portName]
	if !ok {
		return fmt.Errorf("no such input port; it has %s", strings.Join(sortedPortNames(ports), ", "))
	}
	literalType := literalTypeFor(port, spec.Value)
	if literalType == "" {
		return fmt.Errorf("its type is decided by whatever gets connected to it, so there is nothing to make a literal of yet; create the value node yourself and connect it via nodeId/port")
	}
	if !json.Valid([]byte(spec.Value)) {
		return fmt.Errorf("value is not valid JSON: %s", spec.Value)
	}

	poolKey := portName + "\x00" + literalType + "\x00" + canonicalJSON(spec.Value)
	if shared, ok := pool[poolKey]; ok {
		return inst.ConnectNodes(shared, "Value", nodeID, portName)
	}

	paramType := fmt.Sprintf("github.com/EliCDavis/polyform/generator/parameter.Value[%s]", literalType)
	_, paramID, err := inst.CreateNode(paramType)
	if err != nil {
		return fmt.Errorf("type %s: no literal parameter node available for this type — create the value yourself and connect it via nodeId/port: %w", literalType, err)
	}
	if _, err := inst.UpdateParameter(paramID, []byte(spec.Value)); err != nil {
		return err
	}
	if pool != nil {
		pool[poolKey] = paramID
	}
	return inst.ConnectNodes(paramID, "Value", nodeID, portName)
}

// literalTypeFor is the type to build a parameter node of. A lifted port
// reports nothing until something is wired in, but it always accepts its
// element type, and its array type when the literal is a JSON array.
func literalTypeFor(port nodes.InputPort, value string) string {
	element := ""
	if options, ok := port.(nodes.TypeOptions); ok {
		accepted := options.AcceptedTypes()
		for _, t := range accepted {
			if slices.Contains(accepted, "[]"+t) {
				element = t
				break
			}
		}
		if element == "" {
			for _, t := range accepted {
				if !strings.HasPrefix(t, "[]") {
					element = t
					break
				}
			}
		}
		if element != "" && slices.Contains(accepted, "[]"+element) && jsonArrayDepth(value) > strings.Count(element, "[]") {
			return "[]" + element
		}
	}

	// One element of an array port holds a single value. A lifted port
	// reports the array type as soon as an array is wired into any
	// element, so asking it directly would parse the next element's
	// literal against []T and refuse a plain number.
	if _, isArray := port.(nodes.ArrayValueInputPort); isArray && element != "" {
		return element
	}

	if typed, ok := port.(nodes.Typed); ok {
		if t := typed.Type(); t != "" {
			return t
		}
	}
	return element
}

func jsonArrayDepth(value string) int {
	depth := 0
	for _, r := range value {
		switch {
		case r == '[':
			depth++
		case r == ' ' || r == '\t' || r == '\n' || r == '\r':
		default:
			return depth
		}
	}
	return depth
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
		if e := wireCreatedNodeInputs(inst, node, id, in.Inputs); e != nil {
			// A single create is all-or-nothing: a half-wired node left
			// behind has to be found and deleted by hand.
			if _, deleteErr := inst.DeleteNodeById(id); deleteErr != nil {
				return errors.Join(e, deleteErr)
			}
			return fmt.Errorf("%w; the node was not created - fix the inputs and call again", e)
		}
		out.NodeId = id
		for key, literalID := range literalsFeeding(inst, map[string]string{"": id}) {
			if out.Literals == nil {
				out.Literals = map[string]string{}
			}
			out.Literals[strings.TrimPrefix(key, ".")] = literalID
		}
		return nil
	})
	return nil, out, err
}

type DeleteNodeEntry struct {
	NodeId string `json:"nodeId"`
	Scope  string `json:"scope,omitempty" jsonschema:"subgraph id this node lives in; omit to use the call's top-level scope"`
}

type DeleteNodeInput struct {
	NodeId string            `json:"nodeId,omitempty" jsonschema:"id of the one node to delete; ignored when 'nodes' is set"`
	Scope  string            `json:"scope,omitempty" jsonschema:"subgraph id the node lives in; omit for the root graph"`
	Nodes  []DeleteNodeEntry `json:"nodes,omitempty" jsonschema:"delete several nodes in one call, in order. Preferred over one call per node. When set, the single-node fields are ignored."`
}

type DeleteNodeOutput struct {
	Deleted         int                  `json:"deleted"`
	DeletedLiterals []string             `json:"deletedLiterals,omitempty" jsonschema:"literal parameter nodes that fed only a deleted node and were removed with it (a 'value' input's literal); nothing else referenced them"`
	Orphaned        []string             `json:"orphaned,omitempty" jsonschema:"single-value input ports ('nodeId.Port') a deleted node fed, now empty; each reads its zero value until rewired. Ports on nodes this same call deleted are not listed."`
	Shrunk          []ShrunkArray        `json:"shrunk,omitempty" jsonschema:"array input ports that lost elements a deleted node fed. The array closes up - later elements move down - so nothing is left reading empty."`
	Dropped         []DroppedEdgeSummary `json:"dropped,omitempty" jsonschema:"other edges removed because their producer now carries a type their input cannot take - a lifted node whose array input was deleted turns scalar, and an input that only takes an array loses that edge. Includes edges in graphs that place this subgraph. Edges into nodes this same call deleted are not listed. Rewire any that were meant to stay."`
	Errors          []string             `json:"errors,omitempty" jsonschema:"per-entry failures naming the index, when a 'nodes' batch was passed. Every other node in the batch was still deleted. A single-node call reports its failure as a tool error instead."`
}

type ShrunkArray struct {
	Port    string `json:"port" jsonschema:"'nodeId.Port'"`
	Removed []int  `json:"removed" jsonschema:"the indices that went, numbered as they were before this call"`
}

// cutReaders finds the inputs fed by any node in doomed: single-value
// ports left empty, and array ports with the elements that go. Ports
// belonging to a doomed node are left out: they go with it.
//
// One pass over the whole set, before anything is deleted, is what keeps
// the array indices meaningful - deleting an element shifts its siblings
// down, so scanning again between deletions reports the same index twice.
func cutReaders(inst *graph.Graph, doomed map[string]bool) (orphaned []string, shrunk []ShrunkArray) {
	for _, consumerID := range inst.NodeIds() {
		if doomed[consumerID] {
			continue
		}
		for _, port := range inst.Node(consumerID).Inputs() {
			portName := portLabel(port)
			switch p := port.(type) {
			case nodes.SingleValueInputPort:
				if up := p.Value(); up != nil && doomed[inst.NodeId(up.Node())] {
					orphaned = append(orphaned, consumerID+"."+portName)
				}
			case nodes.ArrayValueInputPort:
				cut := ShrunkArray{Port: consumerID + "." + portName}
				for i, up := range p.Value() {
					if up != nil && doomed[inst.NodeId(up.Node())] {
						cut.Removed = append(cut.Removed, i)
					}
				}
				if len(cut.Removed) > 0 {
					shrunk = append(shrunk, cut)
				}
			}
		}
	}
	sort.Strings(orphaned)
	sort.Slice(shrunk, func(i, j int) bool { return shrunk[i].Port < shrunk[j].Port })
	return orphaned, shrunk
}

// doomedScope is the nodes one delete_node call removes from a single
// scope, in the order they were asked for.
type doomedScope struct {
	inst *graph.Graph
	ids  []string
	set  map[string]bool
}

func (s *Server) deleteNode(ctx context.Context, req *mcpsdk.CallToolRequest, in DeleteNodeInput) (*mcpsdk.CallToolResult, DeleteNodeOutput, error) {
	// A single delete keeps the strict contract: a bad id is a tool
	// error. Only a batch downgrades failures to per-entry reports, the
	// same trade set_parameter makes.
	batched := len(in.Nodes) > 0

	entries := in.Nodes
	if !batched {
		if in.NodeId == "" {
			return nil, DeleteNodeOutput{}, fmt.Errorf("pass 'nodeId' or a 'nodes' list")
		}
		entries = []DeleteNodeEntry{{NodeId: in.NodeId, Scope: in.Scope}}
	}

	var out DeleteNodeOutput
	var err error
	s.atomic(&err, func() error {
		// Resolve every entry before removing anything, so the orphaned
		// ports are reported against the graph the caller was looking at.
		doomed := map[string]*doomedScope{}
		var scopeOrder []string

		for i, entry := range entries {
			scope := entry.Scope
			if scope == "" {
				scope = in.Scope
			}

			e := func() error {
				inst, e := s.resolveScope(scope)
				if e != nil {
					return fmt.Errorf("node %d (%s): %w", i, entry.NodeId, e)
				}
				if !inst.HasNodeWithId(entry.NodeId) {
					return s.explainMissingNode(scope, entry.NodeId,
						fmt.Errorf("node %d: can't delete, no node registered with ID %s", i, entry.NodeId))
				}
				target, seen := doomed[scope]
				if !seen {
					target = &doomedScope{inst: inst, set: map[string]bool{}}
					doomed[scope] = target
					scopeOrder = append(scopeOrder, scope)
				}
				if !target.set[entry.NodeId] {
					target.set[entry.NodeId] = true
					target.ids = append(target.ids, entry.NodeId)
				}
				return nil
			}()
			if e != nil {
				if !batched {
					return e
				}
				out.Errors = append(out.Errors, e.Error())
			}
		}

		for _, scope := range scopeOrder {
			target := doomed[scope]
			orphaned, shrunk := cutReaders(target.inst, target.set)
			out.Orphaned = append(out.Orphaned, orphaned...)
			out.Shrunk = append(out.Shrunk, shrunk...)

			// Literals feeding only doomed nodes are unreachable garbage
			// now. Asking once for the whole set also catches one shared
			// by two of them, which node-at-a-time would leave behind.
			literals := literalFeeders(target.inst, target.set)

			for _, id := range target.ids {
				dropped, e := target.inst.DeleteNodeById(id)
				if e != nil {
					return e
				}
				for _, drop := range dropped {
					if drop.Scope == scope && target.set[drop.To] {
						continue
					}
					out.Dropped = append(out.Dropped, summarizeDrops([]graph.DroppedEdge{drop})...)
				}
				out.Deleted++
			}
			for _, id := range literals {
				if target.inst.HasNodeWithId(id) {
					if _, e := target.inst.DeleteNodeById(id); e != nil {
						return e
					}
					out.DeletedLiterals = append(out.DeletedLiterals, id)
				}
			}
		}
		return nil
	})
	return nil, out, err
}

// NodeConnection is one edge in a connect_nodes batch.
type NodeConnection struct {
	OutNodeId string `json:"outNodeId,omitempty" jsonschema:"id of the node providing the value; mutually exclusive with 'variable'"`
	OutPort   string `json:"outPort,omitempty" jsonschema:"name of the output port on outNodeId; required when outNodeId is set"`
	Variable  string `json:"variable,omitempty" jsonschema:"path of an existing variable to feed the input instead of a node; a reference node is created and wired in, and follows the variable's value from then on. Mutually exclusive with outNodeId/outPort."`
	InNodeId  string `json:"inNodeId" jsonschema:"id of the node receiving the value"`
	InPort    string `json:"inPort" jsonschema:"input port on inNodeId. A single-value port is replaced; a bare array port name appends; 'Port.N' replaces element N in place; 'Port.@nodeId' replaces whichever element that node currently feeds."`
	Scope     string `json:"scope,omitempty" jsonschema:"subgraph id both nodes live in; omit to use the call's top-level scope"`
}

type ConnectNodesInput struct {
	Connections []NodeConnection `json:"connections,omitempty" jsonschema:"wire several edges in one call, in order. Preferred over one call per edge. When set, the single-edge fields are ignored."`

	OutNodeId string `json:"outNodeId,omitempty" jsonschema:"id of the node providing the value; mutually exclusive with 'variable'"`
	OutPort   string `json:"outPort,omitempty" jsonschema:"name of the output port on outNodeId; required when outNodeId is set"`
	Variable  string `json:"variable,omitempty" jsonschema:"path of an existing variable to feed the input instead of a node; a reference node is created and wired in, and follows the variable's value from then on. Mutually exclusive with outNodeId/outPort."`
	InNodeId  string `json:"inNodeId,omitempty" jsonschema:"id of the node receiving the value"`
	InPort    string `json:"inPort,omitempty" jsonschema:"input port on inNodeId. A single-value port is replaced; a bare array port name appends; 'Port.N' replaces element N in place; 'Port.@nodeId' replaces whichever element that node currently feeds (no need to look up the index). Use disconnect with 'Port.N' to remove one element."`
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
			if in.OutNodeId == "" && in.Variable == "" && in.InNodeId == "" {
				return fmt.Errorf("pass either 'connections' or a single outNodeId/outPort (or variable) plus inNodeId/inPort")
			}
			connections = []NodeConnection{{
				OutNodeId: in.OutNodeId,
				OutPort:   in.OutPort,
				Variable:  in.Variable,
				InNodeId:  in.InNodeId,
				InPort:    in.InPort,
				Scope:     in.Scope,
			}}
		}

		connectAll := func() error {
			for i, c := range connections {
				if e := s.connectOne(i, c, in.Scope, batched, &out); e != nil {
					return e
				}
			}
			return nil
		}
		if batched {
			dropped, e := s.graph.Batch(connectAll)
			if e != nil {
				return e
			}
			out.Errors = append(out.Errors, batchErrorLines(dropped, nil)...)
		} else if e := connectAll(); e != nil {
			return e
		}
		out.Connected = len(out.Errors) == 0
		return nil
	})
	return nil, out, err
}

// connectOne makes connection i. In a batch a failure is recorded on out
// and the rest go on; alone it is the tool's error.
func (s *Server) connectOne(i int, c NodeConnection, defaultScope string, batched bool, out *ConnectNodesOutput) error {
	scope := c.Scope
	if scope == "" {
		scope = defaultScope
	}

	e := func() error {
		inst, e := s.resolveScope(scope)
		if e != nil {
			return e
		}
		outNodeID, outPort := c.OutNodeId, c.OutPort
		switch {
		case c.Variable != "" && c.OutNodeId != "":
			return fmt.Errorf("pass either outNodeId/outPort or variable, not both")
		case c.Variable != "":
			varNodeID, e := variableReferenceNode(inst, c.Variable)
			if e != nil {
				return e
			}
			outNodeID, outPort = varNodeID, "Value"
		case c.OutPort == "":
			return fmt.Errorf("outPort is required when outNodeId is set")
		default:
			upstream := lookupNode(inst, c.OutNodeId)
			outPort = resolveOutputPortName(upstream, c.OutPort)
			if upstream != nil {
				if _, ok := upstream.Outputs()[outPort]; !ok {
					return fmt.Errorf("node %q has no output port %q; it has %s", c.OutNodeId, c.OutPort, strings.Join(sortedOutputNames(upstream.Outputs()), ", "))
				}
			}
		}
		inPort, e := resolveInputPortByFeeder(inst, lookupNode(inst, c.InNodeId), c.InPort)
		if e != nil {
			return e
		}
		return inst.ConnectNodes(outNodeID, outPort, c.InNodeId, inPort)
	}()
	if e == nil {
		out.Made++
		return nil
	}

	for _, id := range []string{c.OutNodeId, c.InNodeId} {
		e = s.explainMissingNode(scope, id, e)
	}
	if !batched {
		return e
	}
	source := c.OutNodeId + "." + c.OutPort
	if c.Variable != "" {
		source = "variable " + c.Variable
	}
	out.Errors = append(out.Errors, fmt.Sprintf("connection %d (%s -> %s.%s): %v",
		i, source, c.InNodeId, c.InPort, e))
	return nil
}

type DisconnectEntry struct {
	NodeId string `json:"nodeId"`
	Port   string `json:"port" jsonschema:"input port name, 'Port.N' for one element of an array input, or 'Port.@nodeId' for whichever element that node feeds"`
	Scope  string `json:"scope,omitempty" jsonschema:"subgraph id this node lives in; omit to use the call's top-level scope"`
}

type DisconnectInput struct {
	NodeId      string            `json:"nodeId,omitempty"`
	Port        string            `json:"port,omitempty" jsonschema:"input port name; append '.N' to remove a single element of an array input, e.g. 'Items.0', or '.@nodeId' to remove whichever element that node feeds"`
	Scope       string            `json:"scope,omitempty"`
	Connections []DisconnectEntry `json:"connections,omitempty" jsonschema:"remove several in one call; the single fields are then ignored. Every entry is resolved against the graph as it is now, before anything is removed, so 'Items.0', 'Items.2' and 'Items.@Node-7' all mean what they mean today - no index shifts between entries."`
}

type RemainingElements struct {
	Port      string                 `json:"port" jsonschema:"'nodeId.Port' (prefixed 'scope/' inside a subgraph)"`
	Remaining []PortReferenceSummary `json:"remaining" jsonschema:"what is still wired in, in order. Index i here is now 'Port.i'."`
}

type DisconnectOutput struct {
	Disconnected bool                   `json:"disconnected" jsonschema:"true when every requested connection was removed"`
	Removed      int                    `json:"removed,omitempty" jsonschema:"how many were removed, when 'connections' was passed"`
	Remaining    []PortReferenceSummary `json:"remaining,omitempty" jsonschema:"for an array port: what is still wired in, in order, after the removal. Index i here is now 'Port.i'."`
	Arrays       []RemainingElements    `json:"arrays,omitempty" jsonschema:"for a 'connections' batch: every array port touched, with what remains"`
	Dropped      []DroppedEdgeSummary   `json:"dropped,omitempty" jsonschema:"other edges this removed because their producer now carries a type their input cannot take - taking the array off a lifted node's input makes its output a scalar, and an input that only takes an array loses that edge. Includes edges in graphs that place this subgraph. Rewire any that were meant to stay."`
	Errors       []string               `json:"errors,omitempty" jsonschema:"per-entry failures naming the index, for a 'connections' batch. Nothing is removed when any entry fails to resolve."`
}

type resolvedDisconnect struct {
	scope  string
	inst   *graph.Graph
	nodeID string
	base   string
	index  int
	single bool
}

func (s *Server) resolveDisconnect(entry DisconnectEntry, defaultScope string) (resolvedDisconnect, error) {
	scope := entry.Scope
	if scope == "" {
		scope = defaultScope
	}
	inst, err := s.resolveScope(scope)
	if err != nil {
		return resolvedDisconnect{}, err
	}
	if !inst.HasNodeWithId(entry.NodeId) {
		return resolvedDisconnect{}, s.explainMissingNode(scope, entry.NodeId, fmt.Errorf("no node exists with id %q", entry.NodeId))
	}
	node := inst.Node(entry.NodeId)
	port, err := resolveInputPortByFeeder(inst, node, entry.Port)
	if err != nil {
		return resolvedDisconnect{}, err
	}
	resolved := resolvedDisconnect{scope: scope, inst: inst, nodeID: entry.NodeId, base: port, single: true}
	if base, index, indexed := splitPortIndex(node, port); indexed {
		resolved.base, resolved.index, resolved.single = base, index, false
	}
	input, isPort := node.Inputs()[resolved.base]
	if !isPort {
		return resolvedDisconnect{}, fmt.Errorf("node %q has no input port %q; it has %s", entry.NodeId, entry.Port, strings.Join(sortedPortNames(node.Inputs()), ", "))
	}
	if !resolved.single {
		array, isArray := input.(nodes.ArrayValueInputPort)
		if !isArray {
			return resolvedDisconnect{}, fmt.Errorf("node %q port %q is not an array, so it has no element %d to remove", entry.NodeId, resolved.base, resolved.index)
		}
		if resolved.index >= len(array.Value()) {
			return resolvedDisconnect{}, fmt.Errorf("node %q port %q has %d element(s), so there is no index %d to remove", entry.NodeId, resolved.base, len(array.Value()), resolved.index)
		}
	}
	return resolved, nil
}

func (r resolvedDisconnect) portLabel() string {
	label := r.nodeID + "." + r.base
	if r.scope != "" {
		label = r.scope + "/" + label
	}
	return label
}

func remainingElements(inst *graph.Graph, node nodes.Node, base string) []PortReferenceSummary {
	array, ok := node.Inputs()[base].(nodes.ArrayValueInputPort)
	if !ok {
		return nil
	}
	remaining := []PortReferenceSummary{}
	for _, element := range array.Value() {
		remaining = append(remaining, PortReferenceSummary{NodeId: inst.NodeId(element.Node()), Port: portLabel(element)})
	}
	return remaining
}

type DroppedEdgeSummary struct {
	Scope  string `json:"scope,omitempty" jsonschema:"subgraph the edge was in; empty for the root graph"`
	From   string `json:"from" jsonschema:"'nodeId.Port' that fed it"`
	To     string `json:"to" jsonschema:"'nodeId.Port' or 'nodeId.Port.N' it fed"`
	Reason string `json:"reason"`
}

func summarizeDrops(drops []graph.DroppedEdge) []DroppedEdgeSummary {
	var out []DroppedEdgeSummary
	for _, drop := range drops {
		out = append(out, DroppedEdgeSummary{
			Scope:  drop.Scope,
			From:   drop.From + "." + drop.FromPort,
			To:     drop.To + "." + drop.ToPort,
			Reason: drop.Reason,
		})
	}
	return out
}

// batchErrorLines reports what a kept batch lost, naming nodes by the
// aliases (alias to id) the batch gave them.
func batchErrorLines(dropped []graph.DroppedEdge, aliases map[string]string) []string {
	names := make(map[string]string, len(aliases))
	for alias, id := range aliases {
		names[id] = fmt.Sprintf("%s (%s)", alias, id)
	}
	name := func(id string) string {
		if named, ok := names[id]; ok {
			return named
		}
		return id
	}
	lines := make([]string, len(dropped))
	for i, drop := range dropped {
		edge := fmt.Sprintf("%s.%s -> %s.%s", name(drop.From), drop.FromPort, name(drop.To), drop.ToPort)
		if drop.Scope != "" {
			edge = fmt.Sprintf("%s (in %q)", edge, drop.Scope)
		}
		lines[i] = fmt.Sprintf("removed %s: %s", edge, drop.Reason)
	}
	return lines
}

func (s *Server) disconnect(ctx context.Context, req *mcpsdk.CallToolRequest, in DisconnectInput) (*mcpsdk.CallToolResult, DisconnectOutput, error) {
	var out DisconnectOutput
	var err error
	s.atomic(&err, func() error {
		if len(in.Connections) == 0 {
			resolved, e := s.resolveDisconnect(DisconnectEntry{NodeId: in.NodeId, Port: in.Port}, in.Scope)
			if e != nil {
				return e
			}
			port := resolved.base
			if !resolved.single {
				port = fmt.Sprintf("%s.%d", resolved.base, resolved.index)
			}
			dropped, e := resolved.inst.DeleteNodeInputConnection(resolved.nodeID, port)
			if e != nil {
				return e
			}
			out.Dropped = summarizeDrops(dropped)
			out.Disconnected = true
			out.Remaining = remainingElements(resolved.inst, resolved.inst.Node(resolved.nodeID), resolved.base)
			return nil
		}

		resolved := make([]resolvedDisconnect, 0, len(in.Connections))
		for i, entry := range in.Connections {
			r, e := s.resolveDisconnect(entry, in.Scope)
			if e != nil {
				out.Errors = append(out.Errors, fmt.Sprintf("connection %d (%s.%s): %v", i, entry.NodeId, entry.Port, e))
				continue
			}
			resolved = append(resolved, r)
		}
		if len(out.Errors) > 0 {
			return fmt.Errorf("nothing was removed:\n%s", strings.Join(out.Errors, "\n"))
		}

		// Highest element first within a port, so each removal leaves the
		// lower indices where the caller saw them.
		sort.SliceStable(resolved, func(i, j int) bool {
			if resolved[i].portLabel() != resolved[j].portLabel() {
				return resolved[i].portLabel() < resolved[j].portLabel()
			}
			return resolved[i].index > resolved[j].index
		})
		touched := map[string]resolvedDisconnect{}
		var order []string
		for i, r := range resolved {
			if i > 0 && resolved[i-1].portLabel() == r.portLabel() && resolved[i-1].index == r.index && resolved[i-1].single == r.single {
				continue
			}
			port := r.base
			if !r.single {
				port = fmt.Sprintf("%s.%d", r.base, r.index)
			}
			dropped, e := r.inst.DeleteNodeInputConnection(r.nodeID, port)
			if e != nil {
				return fmt.Errorf("%s: %w", r.portLabel(), e)
			}
			out.Dropped = append(out.Dropped, summarizeDrops(dropped)...)
			out.Removed++
			if _, seen := touched[r.portLabel()]; !seen {
				order = append(order, r.portLabel())
			}
			touched[r.portLabel()] = r
		}
		for _, label := range order {
			r := touched[label]
			if remaining := remainingElements(r.inst, r.inst.Node(r.nodeID), r.base); remaining != nil {
				out.Arrays = append(out.Arrays, RemainingElements{Port: label, Remaining: remaining})
			}
		}
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
func parameterFeedingPort(inst *graph.Graph, nodeID, portName, value string) (string, error) {
	node := lookupNode(inst, nodeID)
	if node == nil {
		return "", fmt.Errorf("no node exists with id %q", nodeID)
	}

	inputs := node.Inputs()

	// "Values.1" addresses one element of an array port, the same spelling
	// disconnect takes. Without it, changing a single element's literal
	// meant disconnect-by-index, create a fresh literal, reconnect - three
	// calls for what reads like one.
	base, index, hasIndex := splitPortIndex(node, portName)
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
		if index == len(elements) {
			// The next free slot: append a fresh literal holding value.
			if e := wireOneInput(inst, nodeID, resolved, NodeInputElement{Value: value}, inputs, nil); e != nil {
				return "", e
			}
			return "", nil
		}
		if index > len(elements) {
			return "", fmt.Errorf("port %q has %d element(s), so there is no index %d; %d would append", resolved, len(elements), index, len(elements))
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
		if e := wireOneInput(inst, nodeID, resolved, NodeInputElement{Value: value}, inputs, nil); e != nil {
			return "", e
		}
		return "", nil
	}

	return parameterNodeID(inst, upstream, resolved)
}

// parameterNodeID names the literal behind a port, or explains why there
// isn't one.
func parameterNodeID(inst *graph.Graph, upstream nodes.OutputPort, portLabel string) (string, error) {
	upstreamID := inst.NodeId(upstream.Node())
	if _, isParam := upstream.Node().(graph.Parameter); !isParam {
		return "", fmt.Errorf("port %q is fed by node %q, which computes its value rather than holding a literal; set the parameter feeding that node instead, or disconnect the port first", portLabel, upstreamID)
	}
	return upstreamID, nil
}

// splitPortIndex separates a trailing ".N" array index from a port name. A
// port whose own name ends in ".N", like one named after "Switch 6x3.5", is
// that port.
func splitPortIndex(node nodes.Node, portName string) (base string, index int, ok bool) {
	if node != nil {
		if _, whole := node.Inputs()[resolveInputPortName(node, portName)]; whole {
			return portName, 0, false
		}
	}
	dot := strings.LastIndex(portName, ".")
	if dot == -1 {
		return portName, 0, false
	}
	i, err := strconv.Atoi(portName[dot+1:])
	if err != nil || i < 0 {
		return portName, 0, false
	}
	return portName[:dot], i, true
}

func setOneParameter(inst *graph.Graph, nodeID, portName, value string) (bool, error) {
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

	if !inst.HasNodeWithId(targetID) {
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

			changed, e := func() (bool, error) {
				inst, e := s.resolveScope(scope)
				if e != nil {
					return false, e
				}
				return setOneParameter(inst, a.NodeId, a.Port, a.Value)
			}()
			if e != nil {
				e = s.explainMissingNode(scope, a.NodeId, e)
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
		Description: "Create many nodes at once and wire them together in one call. Use this, not repeated create_node, whenever placing more than one node. Give each entry an 'alias'; any entry's 'inputs' may reference another entry's alias in place of a node id, in either direction. An entry with 'subgraphId' instead of 'type' places an instance of that subgraph (an equation helper, a reusable part) in the same batch. Returns the alias-to-id map. A failing entry does not stop the batch - check 'errors'.",
	}, s.createNodes)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "delete_node",
		Description: "Delete a node, clearing any connections other nodes had to it. Pass 'nodes' to delete many in one call (each with its own scope); the single-node fields are then ignored. A bad id in a 'nodes' batch does not stop the rest - it comes back in 'errors' and every other node is still deleted. Deleting a node does not delete what fed it - list the whole dead chain.",
	}, s.deleteNode)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "connect_nodes",
		Description: "Connect one node's output port to another's input port. Connecting into a bare array port name appends; 'Port.N' sets element N in place (or appends when N is the next free slot; anything beyond is an error), so an existing element can be swapped without renumbering the rest. A connection that would create a cycle is refused. The source may be a variable (by path) instead of a node. Pass 'connections' to make many in one call; the single-edge fields are then ignored. Swapping two inputs is one batch connecting each one's current source into the other.",
	}, s.connectNodes)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "disconnect",
		Description: "Remove an existing connection from a node's input port. Removing 'Port.N' from an array port shifts every later element down by one; the result lists what remains, in their new positions. Pass 'connections' to remove several in one call: every entry ('Port.N' or 'Port.@nodeId') is resolved before anything is removed, so indices mean what they mean now and several elements of one array go in a single call.",
	}, s.disconnect)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "set_parameter",
		Description: "Set the literal value held by a parameter node, addressed by its own node id, or by 'nodeId' plus 'port' to set whatever feeds that port (wiring a literal in if the port is empty; on an array port, 'Port.N' with N equal to the current length appends a new literal element). A port fed by a computing node (not a literal) is refused, never silently replaced: disconnect it first, or set the literal upstream of it. Pass 'parameters' to set many values in one call; the single-value fields are then ignored.",
	}, s.setParameter)
}

// CreateNodesEntry is one node in a create_nodes batch.
type CreateNodesEntry struct {
	Alias      string                   `json:"alias,omitempty" jsonschema:"a short name you pick so other entries can wire to it before it has a real id. Unique within the batch. Optional; it also labels this entry in the returned map and in errors."`
	Type       string                   `json:"type,omitempty" jsonschema:"registered node type key, as returned by search_node_types. Exactly one of type or subgraphId."`
	SubgraphId string                   `json:"subgraphId,omitempty" jsonschema:"instead of type: place an instance of this existing subgraph (an equation helper, a reusable part) as this entry, exactly as instantiate_subgraph would, so its boundary ports can be wired in the same batch."`
	Inputs     map[string]NodeInputSpec `json:"inputs,omitempty" jsonschema:"same shape as create_node's 'inputs'. A 'nodeId' may be a real node id or any entry's alias, including an entry listed later."`
	Value      string                   `json:"value,omitempty" jsonschema:"for a parameter.Value[T] entry only: its initial value as JSON text (\"0.1\", \"\\\"red\\\"\"), so a named literal several entries share needs no follow-up set_parameter"`
}

func (e CreateNodesEntry) typeKey() (string, error) {
	switch {
	case e.Type != "" && e.SubgraphId != "":
		return "", fmt.Errorf("give either type or subgraphId, not both")
	case e.SubgraphId != "":
		return subgraph.RuntimeTypePath(e.SubgraphId), nil
	case e.Type != "":
		return e.Type, nil
	default:
		return "", fmt.Errorf("type is required (or subgraphId, to place an instance)")
	}
}

type CreateNodesInput struct {
	Scope string             `json:"scope,omitempty" jsonschema:"subgraph id to create every node inside; omit for the root graph"`
	Nodes []CreateNodesEntry `json:"nodes" jsonschema:"the nodes to create, in any order"`
}

type CreateNodesOutput struct {
	Nodes    map[string]string `json:"nodes" jsonschema:"alias (or the entry's index as a string, when it had no alias) mapped to the created node id"`
	Literals map[string]string `json:"literals,omitempty" jsonschema:"'alias.Port' (or 'alias.Port.N' for an array element) -> the id of the literal parameter node feeding it, for every port a 'value' created; pass that id to set_parameter or delete_node later without a describe_graph"`
	Errors   []string          `json:"errors,omitempty" jsonschema:"per-entry failures, naming the entry. ALWAYS check: entries outside these still applied."`
}

// literalsFeeding maps each parameter-fed input of the named nodes to
// the literal's id, keyed by the alias the caller used.
func literalsFeeding(inst *graph.Graph, created map[string]string) map[string]string {
	found := map[string]string{}
	for alias, id := range created {
		if !inst.HasNodeWithId(id) {
			continue
		}
		for _, port := range inst.Node(id).Inputs() {
			portName := portLabel(port)
			record := func(upstream nodes.OutputPort, key string) {
				if upstream == nil {
					return
				}
				if _, isParam := upstream.Node().(graph.Parameter); isParam {
					found[key] = inst.NodeId(upstream.Node())
				}
			}
			switch p := port.(type) {
			case nodes.SingleValueInputPort:
				record(p.Value(), alias+"."+portName)
			case nodes.ArrayValueInputPort:
				for i, up := range p.Value() {
					record(up, fmt.Sprintf("%s.%s.%d", alias, portName, i))
				}
			}
		}
	}
	if len(found) == 0 {
		return nil
	}
	return found
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

		dropped, e := s.graph.Batch(func() (hard error) {
			out.Nodes, out.Errors, hard = s.createNodeBatch(inst, in.Nodes, nil)
			return hard
		})
		if e != nil {
			return e
		}
		out.Errors = append(out.Errors, batchErrorLines(dropped, out.Nodes)...)
		out.Literals = literalsFeeding(inst, out.Nodes)
		return nil
	})
	return nil, out, err
}

// createNodeBatch creates every entry, then wires them, so entries can
// reference each other's aliases in any order. aliases seeds the alias
// map with names already bound to node ids (create_subgraph passes its
// boundary ports this way); those names are reserved, and an entry that
// reuses one is a hard error since the reference would be ambiguous.
func (s *Server) createNodeBatch(inst *graph.Graph, entries []CreateNodesEntry, aliases map[string]string) (map[string]string, []string, error) {
	var errs []string

	label := func(i int) string {
		if entries[i].Alias != "" {
			return entries[i].Alias
		}
		return strconv.Itoa(i)
	}

	// A duplicate alias is ambiguous for anything that references it, so
	// every entry carrying one is refused; the rest of the batch goes on.
	seen := make(map[string]int, len(entries))
	skip := make(map[int]bool)
	for i, entry := range entries {
		if entry.Alias == "" {
			continue
		}
		if _, reserved := aliases[entry.Alias]; reserved {
			return nil, nil, fmt.Errorf("alias %q on entry %d is already the name of a boundary port; pick another", entry.Alias, i)
		}
		if prev, dup := seen[entry.Alias]; dup {
			if !skip[prev] {
				errs = append(errs, fmt.Sprintf("entry %d: alias %q is also used by entry %d; aliases must be unique within a batch, so neither was created", prev, entry.Alias, i))
				skip[prev] = true
			}
			skip[i] = true
			continue
		}
		seen[entry.Alias] = i
	}

	named := make(map[string]string, len(entries)+len(aliases))
	for k, v := range aliases {
		named[k] = v
	}
	created := make([]nodes.Node, len(entries))
	ids := make([]string, len(entries))

	for i, entry := range entries {
		if skip[i] {
			continue
		}
		typeKey, e := entry.typeKey()
		if e != nil {
			errs = append(errs, fmt.Sprintf("entry %s: %v", label(i), e))
			continue
		}
		node, id, e := inst.CreateNode(typeKey)
		if e != nil {
			if entry.SubgraphId != "" {
				errs = append(errs, fmt.Sprintf("entry %s: no subgraph %q to instantiate: %v", label(i), entry.SubgraphId, e))
				continue
			}
			errs = append(errs, fmt.Sprintf("entry %s: %v", label(i), s.explainCreateFailure(entry.Type, e)))
			continue
		}
		created[i], ids[i] = node, id
		named[label(i)] = id
		if entry.Value != "" {
			if _, e := setOneParameter(inst, id, "", entry.Value); e != nil {
				errs = append(errs, fmt.Sprintf("entry %s: value: %v (the node was still created)", label(i), e))
			}
		}
	}

	// Second pass: every id exists now, so an entry can reference an
	// alias regardless of where it sits in the list.
	for i, entry := range entries {
		if created[i] == nil {
			continue
		}

		// Wire whatever resolved even when some references didn't, so one
		// stale alias can't silently strip a node of every other input it
		// was given.
		resolved, resolveErr := resolveAliasedInputs(entry.Inputs, named, seen, inst)
		wireErr := wireCreatedNodeInputs(inst, created[i], ids[i], resolved)

		// wireErr already carries the "every other input was wired"
		// suffix; only the alias failures need it added.
		var e error
		switch {
		case resolveErr != nil && wireErr != nil:
			e = fmt.Errorf("%s; %w", resolveErr, wireErr)
		case resolveErr != nil:
			e = fmt.Errorf("%s (every other input on this node was wired)", resolveErr)
		default:
			e = wireErr
		}
		if e != nil {
			errs = append(errs, fmt.Sprintf("entry %s: %v", label(i), e))
		}
	}

	result := make(map[string]string, len(entries))
	for i := range entries {
		if created[i] != nil {
			result[label(i)] = ids[i]
		}
	}
	return result, errs, nil
}

// resolveAliasedInputs rewrites any nodeId naming a batch alias into the
// real id that alias was given. A nodeId matching no alias is passed
// through untouched, so referencing nodes that already existed before the
// batch keeps working.
func resolveAliasedInputs(
	inputs map[string]NodeInputSpec,
	aliases map[string]string,
	declared map[string]int,
	inst *graph.Graph,
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
			if len(elements) < len(spec.Elements) {
				failures = append(failures, noElementWired(portName))
				continue
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
// all three collapse into the same "no node exists with id X" error, which
// names the alias the caller wrote and so reads as though aliases don't
// work at all.
func resolveOneReference(ref string, aliases map[string]string, declared map[string]int, inst *graph.Graph) (string, error) {
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
