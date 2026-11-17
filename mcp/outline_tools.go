package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"gopkg.in/yaml.v3"
)

type SetOutlineInput struct {
	Yaml         string `json:"yaml" jsonschema:"the build plan, as YAML. Top level: version: 1, subject: <part name>, style: organic|mechanical, extent: <subject's longest dimension in world units>, accepted?: [{kind, part?, child?, reason}] (check_outline findings this build knowingly keeps; patch_outline adds them without a resend), parts: {<name>: {kind? (field, mesh, or helper - a helper is an equation/pivot/curve wrapper with no geometry, matched by name and exempt from size and children rules), material?, inputs?: {Port: type}, children: {<child>: {count, size, thickness?, layout?, primitive?, ref?, material?, whole?, paint?, separate?, relief?}}}}. material names a surface finish; parts and leaves inherit it from their parent, and one march exports as one mesh with one material, so a leaf or field part naming a different material than the march it is in is refused - parts with their own material are their own meshes (kind: mesh or separate: true), seated into the body, not blended into it. Every child needs count (how many) and size (fraction of the parent's extent). A child is either a part defined under parts (by its own name or ref) or a leaf with primitive: <package>.<Name> (a generic node by its bare name resolves to its float64 form - math.Multiply; an explicit parameter must be quoted in flow YAML: \"math.Add[int]\"); a leaf also needs thickness (smallest cross-section as a fraction of its own size). count > 1 needs a layout. whole: true on a leaf that really is a single primitive at its own size turns the >50% refusal into a warning; paint: true marks a color-only blob that needs no thickness and sets no resolution budget; separate: true marks a leaf marched on its own so it does not set the shared march budget; relief: true marks a deliberately thin surface detail (a ridge, a groove, a raised edge) that may blur at the march and is allowed to be under-resolved, so it does not set the budget either."`
	IncludeNodes *bool  `json:"includeNodes,omitempty" jsonschema:"whether to return every resolved placement in 'nodes'. true always does, false never does, and omitting it returns them only for a plan of 80 placements or fewer. The list is the largest thing this call returns, so false is worth passing when you only want the issues and the stats"`
}

const outlineNodesInlineLimit = 80

type SetOutlineOutput struct {
	Accepted  bool              `json:"accepted" jsonschema:"false when any error-level issue was found; the outline was not stored"`
	Issues    []OutlineIssue    `json:"issues,omitempty"`
	Stats     OutlineStats      `json:"stats"`
	Nodes     []OutlineNodeStat `json:"nodes,omitempty" jsonschema:"every resolved placement in the tree with its absolute size and depth, so the shape of the plan can be read at a glance; omitted past 80 placements unless includeNodes is set"`
	NodeCount int               `json:"nodeCount,omitempty" jsonschema:"how many placements the plan resolves to, when 'nodes' was left out for size"`
}

func (s *Server) setOutline(ctx context.Context, req *mcpsdk.CallToolRequest, in SetOutlineInput) (*mcpsdk.CallToolResult, SetOutlineOutput, error) {
	var out SetOutlineOutput
	var err error
	s.atomic(&err, func() error {
		o, e := parseOutline(in.Yaml)
		if e != nil {
			return e
		}
		resolver := newPrimitiveResolver(s.graph.BuildSchemaForAllNodeTypes())
		issues, nodes, stats := validateOutline(o, resolver)
		out.Issues = issues
		out.Stats = stats
		include := len(nodes) <= outlineNodesInlineLimit
		if in.IncludeNodes != nil {
			include = *in.IncludeNodes
		}
		if include {
			out.Nodes = nodes
		} else {
			out.NodeCount = len(nodes)
		}
		if hasOutlineErrors(issues) {
			return nil
		}
		s.graph.SetMetadata(outlineMetadataKey, map[string]any{
			"yaml": in.Yaml,
		})
		out.Accepted = true
		return nil
	})
	return nil, out, err
}

type GetOutlineInput struct{}

type GetOutlineOutput struct {
	Yaml string `json:"yaml,omitempty" jsonschema:"the stored outline, verbatim; empty when none has been set"`
}

func (s *Server) getOutline(ctx context.Context, req *mcpsdk.CallToolRequest, in GetOutlineInput) (*mcpsdk.CallToolResult, GetOutlineOutput, error) {
	var out GetOutlineOutput
	var err error
	s.atomic(&err, func() error {
		out.Yaml = storedOutlineYAML(s.graph)
		return nil
	})
	return nil, out, err
}

func storedOutlineYAML(inst *graph.Instance) string {
	stored, ok := inst.Metadata(outlineMetadataKey).(map[string]any)
	if !ok {
		return ""
	}
	text, _ := stored["yaml"].(string)
	return text
}

// OutlineFinding is one discrepancy between the plan and the graph.
type OutlineFinding struct {
	Kind    string `json:"kind" jsonschema:"missing-part | extra-subgraph | missing-input | count-mismatch | missing-leaf | inline-part | oversized-leaf | root-clutter | under-resolved | blend-too-wide | thin-feature-sets-resolution"`
	Part    string `json:"part,omitempty"`
	Child   string `json:"child,omitempty"`
	Message string `json:"message"`
}

type CheckOutlineInput struct{}

type CheckOutlineOutput struct {
	Findings []OutlineFinding `json:"findings"`
	Accepted []OutlineFinding `json:"accepted,omitempty" jsonschema:"findings the outline's accepted list covers, each with its reason; they are not open items"`
	Matched  int              `json:"matched" jsonschema:"parts in the outline that have a subgraph (the subject counts when the root graph is its assembly)"`
	Planned  int              `json:"planned" jsonschema:"parts in the outline"`
	Stats    GraphStats       `json:"stats"`
}

// splitAccepted moves the findings the plan has accepted out of the open
// list, keeping each one's reason on it so a report can quote it.
func splitAccepted(findings []OutlineFinding, accepted []OutlineAcceptance) (open, covered []OutlineFinding) {
	for _, f := range findings {
		var by *OutlineAcceptance
		for i := range accepted {
			if accepted[i].covers(f) {
				by = &accepted[i]
				break
			}
		}
		if by == nil {
			open = append(open, f)
			continue
		}
		f.Message = "accepted (" + by.Reason + "): " + f.Message
		covered = append(covered, f)
	}
	return open, covered
}

func (s *Server) checkOutline(ctx context.Context, req *mcpsdk.CallToolRequest, in CheckOutlineInput) (*mcpsdk.CallToolResult, CheckOutlineOutput, error) {
	var out CheckOutlineOutput
	var err error
	s.atomic(&err, func() error {
		text := storedOutlineYAML(s.graph)
		if text == "" {
			return fmt.Errorf("no outline has been set; call set_outline first")
		}
		o, e := parseOutline(text)
		if e != nil {
			return e
		}
		resolver := newPrimitiveResolver(s.graph.BuildSchemaForAllNodeTypes())
		findings, matched, planned := compareOutline(s.graph, o, resolver)
		_, _, planStats := validateOutline(o, resolver)
		findings = append(findings, checkSofteningBudget(s.graph, planStats, resolver.byKey)...)
		out.Findings, out.Accepted = splitAccepted(findings, o.Accepted)
		out.Matched, out.Planned = matched, planned
		out.Stats = computeGraphStats(s.graph)
		if out.Findings == nil {
			out.Findings = []OutlineFinding{}
		}
		return nil
	})
	return nil, out, err
}

type PatchOutlineInput struct {
	Yaml         string            `json:"yaml,omitempty" jsonschema:"a partial outline to merge into the stored one: parts listed here replace or add parts of the same name, accepted entries are appended, and extent/style/thresholds/subject are taken when given. Parts and accepted entries not mentioned are kept."`
	Remove       []string          `json:"remove,omitempty" jsonschema:"part names to drop from the plan (and from any accepted entries)"`
	Rename       map[string]string `json:"rename,omitempty" jsonschema:"old part name to new; children and refs that point at it, accepted entries and the subject follow"`
	IncludeNodes bool              `json:"includeNodes,omitempty" jsonschema:"return every resolved placement in 'nodes'. Off by default"`
}

// patchOutline edits the stored plan in place so one new helper or one
// accepted finding is a few lines, not a resend of the whole YAML.
func (s *Server) patchOutline(ctx context.Context, req *mcpsdk.CallToolRequest, in PatchOutlineInput) (*mcpsdk.CallToolResult, SetOutlineOutput, error) {
	var out SetOutlineOutput
	var err error
	s.atomic(&err, func() error {
		text := storedOutlineYAML(s.graph)
		if text == "" {
			return fmt.Errorf("no outline has been set; call set_outline first")
		}
		o, e := parseOutline(text)
		if e != nil {
			return e
		}
		if strings.TrimSpace(in.Yaml) != "" {
			var patch Outline
			if e := decodeOutline(in.Yaml, &patch); e != nil {
				return fmt.Errorf("patch is not valid outline YAML: %w", e)
			}
			if patch.Subject != "" {
				o.Subject = patch.Subject
			}
			if patch.Style != "" {
				o.Style = patch.Style
			}
			if patch.Extent > 0 {
				o.Extent = patch.Extent
			}
			if patch.Thresholds != (OutlineThresholds{}) {
				o.Thresholds = patch.Thresholds
			}
			if o.Parts == nil {
				o.Parts = map[string]OutlinePart{}
			}
			for name, part := range patch.Parts {
				o.Parts[name] = part
			}
			o.Accepted = append(o.Accepted, patch.Accepted...)
		}
		for _, name := range in.Remove {
			delete(o.Parts, name)
			o.Accepted = slices.DeleteFunc(o.Accepted, func(a OutlineAcceptance) bool { return a.Part == name })
		}
		for from, to := range in.Rename {
			if e := renameOutlinePart(&o, from, to); e != nil {
				return e
			}
		}
		merged, e := yaml.Marshal(o)
		if e != nil {
			return e
		}
		o, e = parseOutline(string(merged))
		if e != nil {
			return e
		}
		resolver := newPrimitiveResolver(s.graph.BuildSchemaForAllNodeTypes())
		issues, nodes, stats := validateOutline(o, resolver)
		out.Issues, out.Stats = issues, stats
		if in.IncludeNodes {
			out.Nodes = nodes
		} else {
			out.NodeCount = len(nodes)
		}
		if hasOutlineErrors(issues) {
			return nil
		}
		s.graph.SetMetadata(outlineMetadataKey, map[string]any{"yaml": string(merged)})
		out.Accepted = true
		return nil
	})
	return nil, out, err
}

func renameOutlinePart(o *Outline, from, to string) error {
	part, ok := o.Parts[from]
	if !ok {
		return fmt.Errorf("rename: no part %q in the outline", from)
	}
	if _, taken := o.Parts[to]; taken {
		return fmt.Errorf("rename: a part %q already exists", to)
	}
	delete(o.Parts, from)
	o.Parts[to] = part
	if o.Subject == from {
		o.Subject = to
	}
	for name, p := range o.Parts {
		for childName, c := range p.Children {
			switch {
			case c.Ref == from:
				c.Ref = to
				p.Children[childName] = c
			case c.Ref == "" && c.Primitive == "" && childName == from:
				delete(p.Children, childName)
				p.Children[to] = c
			}
		}
		o.Parts[name] = p
	}
	for i := range o.Accepted {
		if o.Accepted[i].Part == from {
			o.Accepted[i].Part = to
		}
	}
	return nil
}

// dropOutlinePart forgets a deleted subgraph's part so the plan does not
// report it missing. A part another part still places stays, since the
// plan is then out of date in a way only the agent can resolve.
func dropOutlinePart(inst *graph.Instance, id string) {
	text := storedOutlineYAML(inst)
	if text == "" {
		return
	}
	o, err := parseOutline(text)
	if err != nil {
		return
	}
	want := normalizePortKey(strings.ReplaceAll(id, "_", ""))
	name := ""
	for part := range o.Parts {
		if normalizePortKey(strings.ReplaceAll(part, "_", "")) == want {
			name = part
		}
	}
	if name == "" || name == o.Subject {
		return
	}
	for _, p := range o.Parts {
		for childName, c := range p.Children {
			if c.Ref == name || (c.Ref == "" && c.Primitive == "" && childName == name) {
				return
			}
		}
	}
	delete(o.Parts, name)
	o.Accepted = slices.DeleteFunc(o.Accepted, func(a OutlineAcceptance) bool { return a.Part == name })
	merged, err := yaml.Marshal(o)
	if err != nil {
		return
	}
	inst.SetMetadata(outlineMetadataKey, map[string]any{"yaml": string(merged)})
}

// subgraphIDFor finds the subgraph a part name refers to, tolerating
// case and underscore/space differences (convert_to_subgraph derives ids
// from display names, so "front leg" lands as Front_leg).
func subgraphIDFor(sch schema.Graph, part string) (string, bool) {
	if _, ok := sch.SubGraphs[part]; ok {
		return part, true
	}
	want := normalizePortKey(strings.ReplaceAll(part, "_", ""))
	for id := range sch.SubGraphs {
		if normalizePortKey(strings.ReplaceAll(id, "_", "")) == want {
			return id, true
		}
	}
	return "", false
}

func isRepeaterType(typeKey string) bool {
	return strings.Contains(typeKey, "sdf.MirrorNode") ||
		strings.Contains(typeKey, "meshops.MirrorNode") ||
		strings.Contains(typeKey, "sdf.RepeatNode") ||
		strings.Contains(typeKey, "/modeling/repeat.")
}

// repeats reports whether a node makes copies: a repeat or mirror node, a
// model drawing GPU instances, or a node lifted over an array so that one
// primitive comes out as many.
func repeats(n schema.Node) bool {
	if isRepeaterType(n.Type) {
		return true
	}
	if strings.Contains(n.Type, "gltf.ModelNode") {
		if _, ok := n.AssignedInput["Gpu Instances"]; ok {
			return true
		}
	}
	for _, out := range n.Output {
		if strings.HasPrefix(out.Type, "[]") {
			return true
		}
	}
	return false
}

// rootIsAssembly is a root graph that places parts: at least one subgraph
// instance or model lives there. An empty root, or one holding a stray
// primitive, has not built the subject at all.
func rootIsAssembly(graphNodes map[string]schema.Node) bool {
	for _, n := range graphNodes {
		if n.SubGraphId != "" || strings.Contains(n.Type, "gltf.ModelNode") || strings.Contains(n.Type, "gltf.ManifestNode") {
			return true
		}
	}
	return false
}

// isGeometryType is a field, mesh or coloured field: the types whose
// presence makes a node geometry rather than the math around it.
func isGeometryType(t string) bool {
	return strings.Contains(t, "modeling.Mesh") || strings.Contains(t, "sample.Vec3To") || strings.Contains(t, "ColoredField")
}

// outputTypes is what a node's ports carry: the type's declaration, with
// the settled type where lifting changed it.
func outputTypes(types map[string]schema.NodeType, n schema.Node) map[string]string {
	out := map[string]string{}
	if t, ok := types[n.Type]; ok {
		for port, o := range t.Outputs {
			out[port] = o.Type
		}
	}
	for port, o := range n.Output {
		if o.Type != "" {
			out[port] = o.Type
		}
	}
	return out
}

func producesGeometry(types map[string]schema.NodeType, n schema.Node) bool {
	for _, t := range outputTypes(types, n) {
		if isGeometryType(t) {
			return true
		}
	}
	return false
}

// isSourcePrimitive is a node that makes geometry out of numbers alone:
// a sphere, a line, a cube. Unions, transforms and marches take geometry in.
func isSourcePrimitive(types map[string]schema.NodeType, n schema.Node) bool {
	t, ok := types[n.Type]
	if !ok || n.SubGraphId != "" {
		return false
	}
	for _, in := range t.Inputs {
		if isGeometryType(in.Type) {
			return false
		}
	}
	return producesGeometry(types, n)
}

// subgraphProducesGeometry follows a subgraph's output boundaries to what
// feeds them. A subgraph with none - an equation, a layout helper - needs
// no place in the plan.
func subgraphProducesGeometry(sch schema.Graph, types map[string]schema.NodeType, id string, seen map[string]bool) bool {
	if seen[id] {
		return false
	}
	seen[id] = true
	sub, ok := sch.SubGraphs[id]
	if !ok {
		return false
	}
	for _, n := range sub.Nodes {
		if n.SubGraphOutputBoundary == nil {
			continue
		}
		ref, ok := n.AssignedInput["Value"]
		if !ok {
			continue
		}
		src, ok := sub.Nodes[ref.NodeId]
		if !ok {
			continue
		}
		if src.SubGraphId != "" {
			if subgraphProducesGeometry(sch, types, src.SubGraphId, seen) {
				return true
			}
			continue
		}
		if producesGeometry(types, src) {
			return true
		}
	}
	return false
}

// partScan is what a part's subtree holds, counted the way the plan counts:
// placements multiplied through nesting, and a primitive feeding several
// transforms counted once per transform. A repeat anywhere below the part
// explains a count the same way one in its own scope does.
type partScan struct {
	inputs    map[string]bool
	repeaters int
	instances map[string]int
	leaves    map[string]int
}

func scanPart(sch schema.Graph, scope map[string]schema.Node, scopeID string, multiplier int, depth int, visited map[string]bool, out *partScan) {
	if visited[scopeID] {
		return
	}
	visited[scopeID] = true
	defer delete(visited, scopeID)

	readers := map[string][]schema.Node{}
	for _, n := range scope {
		for _, ref := range n.AssignedInput {
			readers[ref.NodeId] = append(readers[ref.NodeId], n)
		}
	}
	for id, n := range scope {
		if depth == 0 && n.SubGraphInputBoundary != nil {
			out.inputs[normalizePortKey(n.SubGraphInputBoundary.PortName)] = true
		}
		if repeats(n) {
			out.repeaters++
		}
		if n.SubGraphId != "" {
			out.instances[n.SubGraphId] += multiplier
			if child, ok := sch.SubGraphs[n.SubGraphId]; ok {
				scanPart(sch, child.Nodes, n.SubGraphId, multiplier, depth+1, visited, out)
			}
			continue
		}
		placements := 0
		for _, reader := range readers[id] {
			if strings.Contains(reader.Type, "TransformNode") {
				placements++
			}
		}
		if placements < 1 {
			placements = 1
		}
		out.leaves[n.Type] += multiplier * placements
	}
}

func compareOutline(root *graph.Instance, o Outline, resolver primitiveResolver) ([]OutlineFinding, int, int) {
	var findings []OutlineFinding
	add := func(kind, part, child, format string, args ...any) {
		findings = append(findings, OutlineFinding{Kind: kind, Part: part, Child: child, Message: fmt.Sprintf(format, args...)})
	}

	sch := root.Schema()
	types := resolver.byKey
	planned := map[string]string{}
	matched := 0
	for name := range o.Parts {
		id, ok := subgraphIDFor(sch, name)
		if !ok {
			// The root graph is the assembly when the subject has no
			// subgraph of its own and the root is where the parts are
			// placed: it is checked as the subject's scope, and
			// root-clutter says whether anything else leaked into it.
			if name == o.Subject && rootIsAssembly(sch.Nodes) {
				planned[name] = ""
				matched++
				continue
			}
			add("missing-part", name, "", "outline part %q has no subgraph; build it with create_subgraph (or convert_to_subgraph the nodes that make it up)", name)
			continue
		}
		planned[name] = id
		matched++
	}
	plannedIDs := map[string]bool{}
	for _, id := range planned {
		plannedIDs[id] = true
	}
	for id := range sch.SubGraphs {
		if plannedIDs[id] || !subgraphProducesGeometry(sch, types, id, map[string]bool{}) {
			continue
		}
		add("extra-subgraph", "", "", "subgraph %q is not in the outline; add it to the plan (set_outline again) or say in the report why it exists", id)
	}

	helperIDs := map[string]bool{}
	for name, part := range o.Parts {
		if o.kindOf(part) == "helper" {
			if id, ok := planned[name]; ok {
				helperIDs[id] = true
			}
		}
	}

	for name, part := range o.Parts {
		id, ok := planned[name]
		if !ok || helperIDs[id] {
			continue
		}
		scope := sch
		if id != "" {
			child, e := root.SubGraphInstance(id)
			if e != nil {
				continue
			}
			scope = child.Schema()
		}

		scan := partScan{inputs: map[string]bool{}, instances: map[string]int{}, leaves: map[string]int{}}
		scanPart(sch, scope.Nodes, id, 1, 0, map[string]bool{}, &scan)

		if id != "" {
			for portName := range part.Inputs {
				if !scan.inputs[normalizePortKey(portName)] {
					add("missing-input", name, "", "outline declares input %q on %q but the subgraph has no such boundary input", portName, name)
				}
			}
		}

		for childName, c := range part.Children {
			if c.Primitive != "" {
				key, err := resolver.resolve(c.Primitive)
				if err != nil {
					continue
				}
				have := scan.leaves[key]
				switch {
				case have == 0:
					add("missing-leaf", name, childName, "%q should contain a %s for leaf %q; none found", name, c.Primitive, childName)
				case have < c.Count && scan.repeaters == 0:
					add("count-mismatch", name, childName, "leaf %q plans %d x %s but %q holds %d and no mirror/repeat node", childName, c.Count, c.Primitive, name, have)
				}
				continue
			}
			target := c.Ref
			if target == "" {
				target = childName
			}
			targetID, ok := planned[target]
			if !ok || targetID == "" {
				continue
			}
			have := scan.instances[targetID]
			switch {
			case have == 0:
				add("inline-part", name, childName, "%q never instantiates %q; the outline says it holds %d - was it built inline instead? convert_to_subgraph the cluster, then instantiate_subgraph it here", name, target, c.Count)
			case have < c.Count && scan.repeaters == 0:
				add("count-mismatch", name, childName, "%q plans %d x %q but holds %d instance(s) and no mirror/repeat node", name, c.Count, target, have)
			}
		}

		if structural := structuralNodeCount(scope); id != "" && len(part.Children) > 0 && structural > 60 {
			add("oversized-leaf", name, "", "%q holds %d nodes that do work (of %d in total); a part that big is hiding parts that should be their own subgraphs", name, structural, len(scope.Nodes))
		}
	}
	if clutter := rootGeometry(sch.Nodes, types); len(clutter) > 0 {
		add("root-clutter", "", "", "%d node(s) in the root graph are neither assembly (Model/Manifest/Material/March), material plumbing, instances, literals nor variable references; geometry belongs inside a part subgraph: %s", len(clutter), describeNodes(sch.Nodes, clutter))
	}

	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].Kind != findings[j].Kind {
			return findings[i].Kind < findings[j].Kind
		}
		return findings[i].Part+findings[i].Child < findings[j].Part+findings[j].Child
	})
	return findings, matched, len(o.Parts)
}

// blendTolerance is how far over the budget a radius has to be before it
// is reported. The budget is a bound on swallowing a feature, not a
// tolerance to the fourth decimal; 2% over it swallows nothing.
const blendTolerance = 1.05

// checkSofteningBudget compares every march, smooth union and normal weld
// in the graph against the thinnest feature that actually reaches it.
func checkSofteningBudget(root *graph.Instance, plan OutlineStats, types map[string]schema.NodeType) []OutlineFinding {
	if plan.ThinnestWorld <= 0 {
		return nil
	}
	var findings []OutlineFinding
	add := func(kind, scope, nodeID, format string, args ...any) {
		findings = append(findings, OutlineFinding{Kind: kind, Part: scope, Child: nodeID, Message: fmt.Sprintf(format, args...)})
	}

	scopes := map[string]*graph.Graph{"root": root.Graph}
	for id := range root.Schema().SubGraphs {
		if child, err := root.SubGraphInstance(id); err == nil {
			scopes[id] = child
		}
	}
	scopeNames := make([]string, 0, len(scopes))
	for name := range scopes {
		scopeNames = append(scopeNames, name)
	}
	sort.Strings(scopeNames)

	tracer := scaleTracer{root: root, scopes: scopes, types: types, memo: map[string]float64{}}

	partForScope := map[string]string{}
	for part := range plan.partThinnest {
		if id, ok := subgraphIDFor(root.Schema(), part); ok {
			partForScope[id] = part
		}
	}

	for _, scope := range scopeNames {
		inst := scopes[scope]
		for _, id := range inst.NodeIds() {
			node := inst.Node(id)
			typeKey := inst.Schema().Nodes[id].Type
			// Only the three kinds below need a budget, and working one
			// out walks the graph, so it stays behind the type check.
			budget := func() (thinnestLeaf, float64, float64, bool) {
				return tracer.budgetForNode(plan, partForScope, scope, id, "")
			}
			switch {
			case strings.Contains(typeKey, "marching.MarchNode"):
				if f := splitFinding(plan, partForScope[scope], scope, id, node); f != nil {
					findings = append(findings, *f)
				}
				thinnest, _, _, budgeted := budget()
				res, ok := inputFloat(node, "Resolution")
				if !budgeted || !ok || res <= 0 {
					continue
				}
				// The march samples in its own space; what it has to
				// resolve is the feature as it is there, before the
				// transforms and instances downstream scale it to world.
				scale := tracer.downstream(scope, id, nil)
				minResolution := math.Ceil(voxelsAcrossThinnest * scale / thinnest.world)
				if res < minResolution {
					add("under-resolved", scope, id, "MarchNode %s in %q has Resolution %v (voxels %.4f across); the thinnest feature it covers, %s, is %.4f in world units (x%.3g downstream scale = %.4f here), which needs Resolution >= %v to span %v voxels", id, scope, res, 1/res, thinnest.path, thinnest.world, scale, thinnest.world/scale, minResolution, voxelsAcrossThinnest)
				}
			case strings.Contains(typeKey, "sdf.SmoothUnionNode") || strings.Contains(typeKey, "sdf.SmoothUnionColoredNode"):
				r, ok := inputFloat(node, "Radius")
				if !ok {
					continue
				}
				// A colored union that is never marched sets a color fade,
				// not a shape, so the thinnest-feature limit does not apply.
				if strings.Contains(typeKey, "Colored") && !tracer.reachesMarch(scope, id, map[string]bool{}) {
					continue
				}
				thinnest, _, maxBlend, budgeted := budget()
				if !budgeted {
					continue
				}
				scale := tracer.downstream(scope, id, nil)
				if world := r * scale; world > maxBlend*blendTolerance {
					add("blend-too-wide", scope, id, "%s in %q blends with Radius %v (x%.3g downstream scale = %.4f in world units); the thinnest feature it blends, %s, is %.4f, so any world radius above %.4f bridges or swallows it", id, scope, r, scale, world, thinnest.path, thinnest.world, maxBlend)
				}
			case strings.Contains(typeKey, "SmoothNormalsImplicitWeldNode"):
				thinnest, _, maxBlend, budgeted := budget()
				if !budgeted {
					continue
				}
				if d, ok := inputFloat(node, "Distance"); ok && d > maxBlend*blendTolerance {
					add("blend-too-wide", scope, id, "%s in %q welds normals over Distance %v; the thinnest feature it covers, %s, is %.4f, so anything above %.4f averages its two faces into a tube", id, scope, d, thinnest.path, thinnest.world, maxBlend)
				}
			}
		}
	}
	return findings
}

// scaleTracer follows a field node's output toward the march, through
// TransformNodes in its own scope and out through subgraph instances,
// and returns the largest scale factor the field is drawn at.
type scaleTracer struct {
	root   *graph.Instance
	scopes map[string]*graph.Graph
	types  map[string]schema.NodeType
	memo   map[string]float64
}

// consumerScale is the scale a node applies to the geometry flowing
// through it on its way to the world: a transform's TRS, a model's own
// Scale and the largest of its GPU instances.
func consumerScale(n schema.Node, node nodes.Node) float64 {
	switch {
	case strings.Contains(n.Type, "sdf.TransformNode"):
		return transformScale(node, "Transform")
	case strings.Contains(n.Type, "meshops.TransformNode"):
		return transformScale(node, "TRS")
	case strings.Contains(n.Type, "gltf.ModelNode"):
		factor := 1.0
		if in, ok := node.Inputs()["Scale"].(nodes.SingleValueInputPort); ok && in.Value() != nil {
			if out, ok := in.Value().(nodes.Output[vector3.Float64]); ok {
				factor = maxAbsComponent(out.Value())
			}
		}
		if in, ok := node.Inputs()["Gpu Instances"].(nodes.SingleValueInputPort); ok && in.Value() != nil {
			if out, ok := in.Value().(nodes.Output[[]trs.TRS]); ok {
				largest := 0.0
				for _, t := range out.Value() {
					largest = math.Max(largest, maxAbsComponent(t.Scale()))
				}
				if largest > 0 {
					factor *= largest
				}
			}
		}
		return factor
	}
	return 1
}

func maxAbsComponent(v vector3.Float64) float64 {
	return math.Max(math.Abs(v.X()), math.Max(math.Abs(v.Y()), math.Abs(v.Z())))
}

func (t scaleTracer) downstream(scope, nodeID string, stack map[string]bool) float64 {
	key := scope + "\x00" + nodeID
	if v, ok := t.memo[key]; ok {
		return v
	}
	if stack == nil {
		stack = map[string]bool{}
	}
	if stack[key] {
		return 1
	}
	stack[key] = true
	defer delete(stack, key)

	inst := t.scopes[scope]
	node := inst.Node(nodeID)
	sch := inst.Schema()

	best := 0.0
	seen := false
	consider := func(factor float64) {
		if !seen || factor > best {
			best, seen = factor, true
		}
	}

	if _, isBoundary := node.(*subgraph.OutputNode); isBoundary {
		for parentScope, parent := range t.scopes {
			for id, n := range parent.Schema().Nodes {
				if n.SubGraphId == scope {
					consider(t.downstream(parentScope, id, stack))
				}
			}
		}
	} else {
		for consumerID, n := range sch.Nodes {
			for _, ref := range n.AssignedInput {
				if ref.NodeId != nodeID {
					continue
				}
				consider(consumerScale(n, inst.Node(consumerID)) * t.downstream(scope, consumerID, stack))
				break
			}
		}
	}

	if !seen {
		best = 1
	}
	t.memo[key] = best
	return best
}

// feedingScopes is every scope whose geometry actually reaches nodeID's
// port, following inputs up and stepping into a subgraph instance through
// the one output boundary that was read.
func (t scaleTracer) feedingScopes(scope, nodeID, port string) map[string]bool {
	hit := map[string]bool{}
	t.walkUp(scope, nodeID, port, map[string]bool{}, hit)
	return hit
}

func (t scaleTracer) walkUp(scope, nodeID, port string, visited, hit map[string]bool) {
	key := scope + "\x00" + nodeID + "\x00" + port
	if visited[key] {
		return
	}
	visited[key] = true

	inst := t.scopes[scope]
	if inst == nil {
		return
	}
	n, ok := inst.Schema().Nodes[nodeID]
	if !ok {
		return
	}
	hit[scope] = true

	// A subgraph instance computes its value inside its own scope, so the
	// walk continues from the boundary that produced the port just read.
	if n.SubGraphId != "" {
		if child := t.scopes[n.SubGraphId]; child != nil {
			for id, cn := range child.Schema().Nodes {
				if b := cn.SubGraphOutputBoundary; b != nil && (port == "" || b.PortName == port) {
					t.walkUp(n.SubGraphId, id, "", visited, hit)
				}
			}
		}
	}

	for _, ref := range n.AssignedInput {
		t.walkUp(scope, ref.NodeId, ref.PortName, visited, hit)
	}
}

// budgetForNode is the tightest budget among the outline parts whose
// geometry reaches nodeID, rather than whichever part the node happens to
// sit in. A separately marched leg no longer sets the resolution for the
// upholstery it never touches.
//
// Within those parts, the budget is the thinnest of the leaves the node's
// own operands implement, not the thinnest leaf the part has anywhere: a
// union of two head spheres is not judged by the ear it never touches.
//
// With no planned field leaf anywhere upstream there is nothing the node
// could swallow, and no budget: ok is false and the node is left alone.
func (t scaleTracer) budgetForNode(plan OutlineStats, partForScope map[string]string, scope, nodeID, port string) (leaf thinnestLeaf, minResolution, maxBlend float64, ok bool) {
	var candidates []thinnestLeaf
	seen := map[string]bool{}
	for reached := range t.feedingScopes(scope, nodeID, port) {
		for _, leaf := range plan.partLeaves[partForScope[reached]] {
			if !seen[leaf.path] {
				seen[leaf.path] = true
				candidates = append(candidates, leaf)
			}
		}
	}
	if len(candidates) == 0 {
		return thinnestLeaf{}, 0, 0, false
	}

	best := thinnestLeaf{}
	for _, src := range t.sourcesFeeding(scope, nodeID, port) {
		inst := t.scopes[src.scope]
		if inst == nil {
			continue
		}
		world := primitiveThickness(inst.Node(src.id)) * t.downstream(src.scope, src.id, nil)
		leaf, ok := matchLeaf(candidates, src.node.Type, world)
		if ok && (best.world == 0 || leaf.world < best.world) {
			best = leaf
		}
	}
	if best.world == 0 {
		for _, leaf := range candidates {
			if best.world == 0 || leaf.world < best.world {
				best = leaf
			}
		}
	}
	return best, math.Ceil(voxelsAcrossThinnest / best.world), best.world / 2, true
}

func transformScale(node nodes.Node, port string) float64 {
	in, ok := node.Inputs()[port].(nodes.SingleValueInputPort)
	if !ok || in.Value() == nil {
		return 1
	}
	out, ok := in.Value().(nodes.Output[trs.TRS])
	if !ok {
		return 1
	}
	return maxAbsComponent(out.Value().Scale())
}

// primitiveThickness is a source primitive's smallest cross-section in its
// own space, read off the ports that size it: a radius doubled, the
// smallest extent of a box. Zero when the node has none of those ports.
func primitiveThickness(node nodes.Node) float64 {
	best := 0.0
	consider := func(v float64) {
		if v > 0 && (best == 0 || v < best) {
			best = v
		}
	}
	for name, port := range node.Inputs() {
		in, ok := port.(nodes.SingleValueInputPort)
		if !ok || in.Value() == nil {
			continue
		}
		key := normalizePortKey(name)
		switch out := in.Value().(type) {
		case nodes.Output[float64]:
			switch {
			case strings.HasPrefix(key, "radius"):
				consider(2 * out.Value())
			case key == "thickness", key == "height", key == "width", key == "depth", key == "size":
				consider(out.Value())
			}
		case nodes.Output[vector3.Float64]:
			if key == "size" || key == "dimensions" || key == "extents" {
				v := out.Value()
				consider(math.Min(math.Abs(v.X()), math.Min(math.Abs(v.Y()), math.Abs(v.Z()))))
			}
		}
	}
	return best
}

// sourcesFeeding is every source primitive whose geometry reaches nodeID's
// port, with the scope it lives in.
func (t scaleTracer) sourcesFeeding(scope, nodeID, port string) []primitiveRef {
	var sources []primitiveRef
	t.walkUpSources(scope, nodeID, port, map[string]bool{}, &sources)
	return sources
}

type primitiveRef struct {
	scope, id string
	node      schema.Node
}

func (t scaleTracer) walkUpSources(scope, nodeID, port string, visited map[string]bool, out *[]primitiveRef) {
	key := scope + "\x00" + nodeID + "\x00" + port
	if visited[key] {
		return
	}
	visited[key] = true

	inst := t.scopes[scope]
	if inst == nil {
		return
	}
	n, ok := inst.Schema().Nodes[nodeID]
	if !ok {
		return
	}
	if isSourcePrimitive(t.types, n) {
		*out = append(*out, primitiveRef{scope: scope, id: nodeID, node: n})
	}
	if n.SubGraphId != "" {
		if child := t.scopes[n.SubGraphId]; child != nil {
			for id, cn := range child.Schema().Nodes {
				if b := cn.SubGraphOutputBoundary; b != nil && (port == "" || b.PortName == port) {
					t.walkUpSources(n.SubGraphId, id, "", visited, out)
				}
			}
		}
	}
	for _, ref := range n.AssignedInput {
		t.walkUpSources(scope, ref.NodeId, ref.PortName, visited, out)
	}
}

// matchLeaf pairs a source primitive with the planned leaf it most likely
// implements: same primitive type, then the closest thickness once the
// node's own size is scaled to world. With no size to go on it takes the
// thinnest candidate, which is the old answer.
func matchLeaf(candidates []thinnestLeaf, typeKey string, worldThickness float64) (thinnestLeaf, bool) {
	var pool []thinnestLeaf
	for _, leaf := range candidates {
		if leaf.primitive == typeKey {
			pool = append(pool, leaf)
		}
	}
	if len(pool) == 0 {
		pool = candidates
	}
	if len(pool) == 0 {
		return thinnestLeaf{}, false
	}
	best := pool[0]
	if worldThickness <= 0 {
		for _, leaf := range pool[1:] {
			if leaf.world < best.world {
				best = leaf
			}
		}
		return best, true
	}
	bestDist := math.Inf(1)
	for _, leaf := range pool {
		if leaf.world <= 0 {
			continue
		}
		if d := math.Abs(math.Log(leaf.world / worldThickness)); d < bestDist {
			best, bestDist = leaf, d
		}
	}
	return best, true
}

func inputFloat(node nodes.Node, port string) (float64, bool) {
	in, ok := node.Inputs()[port].(nodes.SingleValueInputPort)
	if !ok || in.Value() == nil {
		return 0, false
	}
	out, ok := in.Value().(nodes.Output[float64])
	if !ok {
		return 0, false
	}
	return out.Value(), true
}

// rootGeometry lists the nodes that build geometry in the root graph, plus
// the math that exists only to feed them. Math whose readers are all
// assembly, materials or instances - a placement vector into an instance's
// Transform, a color conversion into a material - is wiring, not geometry.
func rootGeometry(graphNodes map[string]schema.Node, types map[string]schema.NodeType) []string {
	readers := map[string][]string{}
	for id, n := range graphNodes {
		for _, ref := range n.AssignedInput {
			readers[ref.NodeId] = append(readers[ref.NodeId], id)
		}
	}

	clutter := map[string]bool{}
	for id, n := range graphNodes {
		if !isAssemblyType(n) && !isMaterialType(n.Type) && producesGeometry(types, n) {
			clutter[id] = true
		}
	}
	// Math is clutter when nothing but clutter (or nothing at all) reads
	// it. Iterate to a fixed point so a chain feeding a root sphere counts.
	for changed := true; changed; {
		changed = false
		for id, n := range graphNodes {
			if clutter[id] || isAssemblyType(n) || isMaterialType(n.Type) {
				continue
			}
			only := true
			for _, reader := range readers[id] {
				if !clutter[reader] {
					only = false
					break
				}
			}
			if only {
				clutter[id] = true
				changed = true
			}
		}
	}

	geometry := make([]string, 0, len(clutter))
	for id := range clutter {
		geometry = append(geometry, id)
	}
	sort.Strings(geometry)
	return geometry
}

// describeNodes names the first few of a list so a finding points at
// something: "Node-12 (sdf.Sphere), Node-14 (math.Multiply)".
func describeNodes(graphNodes map[string]schema.Node, ids []string) string {
	const show = 8
	parts := make([]string, 0, show+1)
	for i, id := range ids {
		if i == show {
			parts = append(parts, fmt.Sprintf("and %d more", len(ids)-show))
			break
		}
		t := graphNodes[id].Type
		if i := strings.LastIndex(t, "/"); i >= 0 {
			t = t[i+1:]
		}
		t = strings.TrimSuffix(strings.TrimSuffix(t, "]"), "Node")
		parts = append(parts, fmt.Sprintf("%s (%s)", id, t))
	}
	return strings.Join(parts, ", ")
}

func isMaterialType(typeKey string) bool {
	for _, marker := range []string{"gltf.Material", "gltf.Texture", "gltf.NormalTexture", "gltf.Sampler"} {
		if strings.Contains(typeKey, marker) {
			return true
		}
	}
	return false
}

func isAssemblyType(n schema.Node) bool {
	if n.SubGraphId != "" || n.Parameter != nil || n.Variable != nil {
		return true
	}
	for _, marker := range []string{
		"gltf.ModelNode", "gltf.ManifestNode", "gltf.MaterialNode", "gltf.TextureNode",
		"marching.MarchNode", "SmoothNormals", "FlatNormals", "meshops.CombineNode",
		"meshops.TransformNode", "meshops.FillVertexColorNode", "modeling.SetAttribute",
		"modeling.FillAttribute", "basics.",
	} {
		if strings.Contains(n.Type, marker) {
			return true
		}
	}
	return false
}

// GraphStats is the shape of a graph in numbers, for comparing builds
// against each other regardless of who made them.
type GraphStats struct {
	TotalNodes      int            `json:"totalNodes"`
	RootNodes       int            `json:"rootNodes"`
	RootGeometry    int            `json:"rootGeometry" jsonschema:"root nodes that are not assembly, instances, literals or variable references"`
	Subgraphs       int            `json:"subgraphs"`
	Instances       int            `json:"instances" jsonschema:"subgraph instance nodes across every scope"`
	MaxNesting      int            `json:"maxNesting" jsonschema:"deepest chain of subgraph-inside-subgraph"`
	BoundaryInputs  int            `json:"boundaryInputs"`
	Variables       int            `json:"variables"`
	NodesByScope    map[string]int `json:"nodesByScope"`
	InstancesOf     map[string]int `json:"instancesOf" jsonschema:"subgraph id -> how many times it is placed, across every scope"`
	InSubgraphShare float64        `json:"inSubgraphShare" jsonschema:"fraction of all nodes that live inside a subgraph definition"`
}

func computeGraphStats(root *graph.Instance) GraphStats {
	sch := root.Schema()
	stats := GraphStats{
		NodesByScope: map[string]int{"root": len(sch.Nodes)},
		InstancesOf:  map[string]int{},
		Subgraphs:    len(sch.SubGraphs),
		RootNodes:    len(sch.Nodes),
		TotalNodes:   len(sch.Nodes),
	}
	sch.Variables.Traverse(func(string, schema.Variable) bool { stats.Variables++; return true })

	count := func(nodes map[string]schema.Node) {
		for _, n := range nodes {
			if n.SubGraphId != "" {
				stats.Instances++
				stats.InstancesOf[n.SubGraphId]++
			}
			if n.SubGraphInputBoundary != nil {
				stats.BoundaryInputs++
			}
		}
	}
	count(sch.Nodes)
	stats.RootGeometry = len(rootGeometry(sch.Nodes, newPrimitiveResolver(root.BuildSchemaForAllNodeTypes()).byKey))

	children := map[string]map[string]bool{}
	for id := range sch.SubGraphs {
		child, err := root.SubGraphInstance(id)
		if err != nil {
			continue
		}
		cs := child.Schema()
		stats.NodesByScope[id] = len(cs.Nodes)
		stats.TotalNodes += len(cs.Nodes)
		count(cs.Nodes)
		children[id] = map[string]bool{}
		for _, n := range cs.Nodes {
			if n.SubGraphId != "" {
				children[id][n.SubGraphId] = true
			}
		}
	}

	var depth func(id string, seen map[string]bool) int
	depth = func(id string, seen map[string]bool) int {
		if seen[id] {
			return 0
		}
		seen[id] = true
		best := 0
		for c := range children[id] {
			if d := depth(c, seen); d > best {
				best = d
			}
		}
		delete(seen, id)
		return best + 1
	}
	for _, n := range sch.Nodes {
		if n.SubGraphId != "" {
			if d := depth(n.SubGraphId, map[string]bool{}); d > stats.MaxNesting {
				stats.MaxNesting = d
			}
		}
	}
	if stats.TotalNodes > 0 {
		stats.InSubgraphShare = float64(stats.TotalNodes-stats.RootNodes) / float64(stats.TotalNodes)
	}
	return stats
}

type GraphStatsInput struct{}

func (s *Server) graphStats(ctx context.Context, req *mcpsdk.CallToolRequest, in GraphStatsInput) (*mcpsdk.CallToolResult, GraphStats, error) {
	var out GraphStats
	var err error
	s.atomic(&err, func() error {
		out = computeGraphStats(s.graph)
		return nil
	})
	return nil, out, err
}

// GraphStatsJSON is computeGraphStats for an offline graph file, so runs
// by different agents can be scored the same way after the fact.
func GraphStatsJSON(inst *graph.Instance) ([]byte, error) {
	type report struct {
		Stats    GraphStats       `json:"stats"`
		Outline  string           `json:"outline,omitempty"`
		Findings []OutlineFinding `json:"findings,omitempty"`
		Matched  int              `json:"matched,omitempty"`
		Planned  int              `json:"planned,omitempty"`
	}
	r := report{Stats: computeGraphStats(inst)}
	if text := storedOutlineYAML(inst); text != "" {
		r.Outline = text
		if o, err := parseOutline(text); err == nil {
			resolver := newPrimitiveResolver(inst.BuildSchemaForAllNodeTypes())
			r.Findings, r.Matched, r.Planned = compareOutline(inst, o, resolver)
			_, _, planStats := validateOutline(o, resolver)
			r.Findings = append(r.Findings, checkSofteningBudget(inst, planStats, resolver.byKey)...)
			r.Findings, _ = splitAccepted(r.Findings, o.Accepted)
		}
	}
	return json.MarshalIndent(r, "", "  ")
}

func (s *Server) registerOutlineTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "set_outline",
		Description: "Commit the build plan before building. Validates the outline (extent in world units; every child needs count and size; a leaf names its primitive and its thickness; anything above a quarter of the subject must be decomposed; count > 1 needs a layout) and stores it with the graph. Returns the resolved tree with world sizes, the thinnest planned feature, and from it the minimum MarchNode Resolution and the maximum blend/weld radius the build may use. Refused with the issue list when there are errors - fix the plan and call again. Call again later if the plan changes.",
	}, s.setOutline)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "get_outline",
		Description: "Return the stored outline YAML, if any. Use it when resuming a project to see what was planned.",
	}, s.getOutline)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "patch_outline",
		Description: "Edit the stored outline without resending it: merge a partial YAML (parts to add or replace, accepted findings, a new extent), remove parts, rename parts. Validates and stores the result exactly as set_outline does; get_outline returns the merged YAML. Use it when a helper appears mid-build, a part is split, or a check_outline finding is right for this build and should stop being reported: accepted: [{kind: oversized-leaf, part: strips, reason: one strip per path is the design}].",
	}, s.patchOutline)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "check_outline",
		Description: "Diff the graph against the stored outline: parts with no subgraph (the root graph may be the subject's own assembly), subgraphs that build geometry but are not in the plan, declared inputs with no boundary port, children the parent never instantiates (built inline?), counts that don't match with nothing to explain them (a mirror or repeat node, GPU instances, an array-lifted primitive, or one primitive placed by several transforms, anywhere in the part's subtree, all count), oversized parts, a march whose resolution is set by one thin feature it pays for over a whole body, geometry sitting in the root graph (named), a MarchNode whose Resolution is too coarse in its own space for the thinnest feature it covers, and a SmoothUnion Radius or normal-weld Distance wider than the thinnest feature its own operands implement. Findings the outline's accepted list covers come back under accepted with their reasons, not as open items. Run it before save_graph and put its findings in the report.",
	}, s.checkOutline)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "graph_stats",
		Description: "The graph's shape in numbers: nodes per scope, subgraph and instance counts, nesting depth, how much of the graph lives inside subgraphs. Cheap; use it to see whether structure is going where the plan says.",
	}, s.graphStats)
}

// inputAABB reads a march's Domain so a finding can talk in voxels rather
// than in resolution alone: the same resolution is cheap over a tooth and
// ruinous over a whole body.
func inputAABB(node nodes.Node, port string) (geometry.AABB, bool) {
	in, ok := node.Inputs()[port].(nodes.SingleValueInputPort)
	if !ok || in.Value() == nil {
		return geometry.AABB{}, false
	}
	out, ok := in.Value().(nodes.Output[geometry.AABB])
	if !ok {
		return geometry.AABB{}, false
	}
	return out.Value(), true
}

func voxelCount(domain geometry.AABB, resolution float64) float64 {
	size := domain.Size()
	return math.Ceil(size.X()*resolution) * math.Ceil(size.Y()*resolution) * math.Ceil(size.Z()*resolution)
}

func formatVoxels(n float64) string {
	switch {
	case n >= 1e6:
		return fmt.Sprintf("%.1fM", n/1e6)
	case n >= 1e3:
		return fmt.Sprintf("%.0fk", n/1e3)
	default:
		return fmt.Sprintf("%.0f", n)
	}
}

// splitFinding reports a march whose resolution is set by one thin feature
// while it has to pay that resolution over a whole body.
//
// Resolution is a property of the march, but thinness is a property of one
// leaf, and the cost of the mismatch is cubic.
func splitFinding(plan OutlineStats, part, scope, nodeID string, node nodes.Node) *OutlineFinding {
	thin, next, ok := plan.outlierLeaf(part)
	if !ok {
		return nil
	}

	domain, ok := inputAABB(node, "Domain")
	if !ok {
		return nil
	}
	res, ok := inputFloat(node, "Resolution")
	if !ok || res <= 0 {
		return nil
	}

	have := voxelCount(domain, res)
	if have < voxelsWorthSplitting {
		return nil
	}

	// What the rest of the part would need if the thin leaf were not in it.
	without := math.Ceil(voxelsAcrossThinnest / next.world)
	if without >= res {
		return nil
	}
	would := voxelCount(domain, without)

	size := domain.Size()
	return &OutlineFinding{
		Kind: "thin-feature-sets-resolution", Part: scope, Child: nodeID,
		Message: fmt.Sprintf(
			"MarchNode %s spans %.2fx%.2fx%.2f at Resolution %v, which is %s voxels every time it runs. %s (%.4f thick) is what forces that resolution; everything else in %q is at least %.4f and would only need %v, or %s voxels - %.0fx cheaper. If it is a distinct piece that should not blend, mark it separate: true and march it on its own over a domain that just covers it. If it has to blend into the surface (a small bump or groove), it stays: a separate mesh shows a seam where it meets the body.",
			nodeID, size.X(), size.Y(), size.Z(), res, formatVoxels(have),
			thin.path, thin.world, part, next.world, without, formatVoxels(would), have/would),
	}
}

// structuralNodeCount is how many nodes a subgraph is actually built from.
// The literals and variable references created to feed ports are not
// structure, and neither is a node nothing reads and nothing feeds -
// prune_orphans exists to sweep those. Counting them made a modest part
// look oversized.
func structuralNodeCount(sch schema.Graph) int {
	consumed := make(map[string]bool, len(sch.Nodes))
	for _, n := range sch.Nodes {
		for _, ref := range n.AssignedInput {
			consumed[ref.NodeId] = true
		}
	}

	count := 0
	for id, n := range sch.Nodes {
		if n.Parameter != nil || n.Variable != nil {
			continue
		}
		if len(n.AssignedInput) == 0 && !consumed[id] {
			continue
		}
		count++
	}
	return count
}

// reachesMarch reports whether this node's value ends up being marched
// into geometry, rather than only coloring vertices that already exist.
//
// A SmoothUnionColoredNode that feeds nothing but ApplyColorFieldNode has
// a Radius that sets how softly one color fades into another. Holding it
// to the march's thinnest feature is a false alarm - nothing about it can
// swallow a thin shape, because it never becomes shape.
func (t scaleTracer) reachesMarch(scope, nodeID string, seen map[string]bool) bool {
	key := scope + "\x00" + nodeID
	if seen[key] {
		return false
	}
	seen[key] = true

	inst := t.scopes[scope]
	if inst == nil {
		return false
	}
	sch := inst.Schema()

	if node := inst.Node(nodeID); node != nil {
		if _, isBoundary := node.(*subgraph.OutputNode); isBoundary {
			for parentScope, parent := range t.scopes {
				for id, n := range parent.Schema().Nodes {
					if n.SubGraphId == scope && t.reachesMarch(parentScope, id, seen) {
						return true
					}
				}
			}
			return false
		}
	}

	for consumerID, n := range sch.Nodes {
		feeds := false
		for _, ref := range n.AssignedInput {
			if ref.NodeId == nodeID {
				feeds = true
				break
			}
		}
		if !feeds {
			continue
		}
		if strings.Contains(n.Type, "marching.MarchNode") {
			return true
		}
		if t.reachesMarch(scope, consumerID, seen) {
			return true
		}
	}
	return false
}
