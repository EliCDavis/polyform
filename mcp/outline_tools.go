package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type SetOutlineInput struct {
	Yaml         string `json:"yaml" jsonschema:"the build plan, as YAML. Top level: version: 1, subject: <part name>, style: organic|mechanical, extent: <subject's longest dimension in world units>, parts: {<name>: {kind? (field, mesh, or helper - a helper is an equation/pivot/curve wrapper with no geometry, matched by name and exempt from size and children rules), inputs?: {Port: type}, children: {<child>: {count, size, thickness?, layout?, primitive?, ref?, whole?, paint?, separate?, relief?}}}}. Every child needs count (how many) and size (fraction of the parent's extent). A child is either a part defined under parts (by its own name or ref) or a leaf with primitive: <package>.<Name> (a generic node by its bare name resolves to its float64 form - math.Multiply; an explicit parameter must be quoted in flow YAML: \"math.Add[int]\"); a leaf also needs thickness (smallest cross-section as a fraction of its own size). count > 1 needs a layout. whole: true on a leaf that really is one primitive at its size (a cushion slab) turns the >50% refusal into a warning; paint: true marks a color-only blob that needs no thickness and sets no resolution budget; separate: true marks a leaf marched on its own (a tongue, a bead) so it does not set the shared march budget; relief: true marks a deliberately thin surface detail (a ridge, a panel line, a raised lip) that may blur at the march and is allowed to be under-resolved, so it does not set the budget either."`
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
	Matched  int              `json:"matched" jsonschema:"parts in the outline that have a subgraph"`
	Planned  int              `json:"planned" jsonschema:"parts in the outline"`
	Stats    GraphStats       `json:"stats"`
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
		out.Findings, out.Matched, out.Planned = compareOutline(s.graph, o, resolver)
		_, _, planStats := validateOutline(o, resolver)
		out.Findings = append(out.Findings, checkSofteningBudget(s.graph, planStats)...)
		out.Stats = computeGraphStats(s.graph)
		if out.Findings == nil {
			out.Findings = []OutlineFinding{}
		}
		return nil
	})
	return nil, out, err
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

func compareOutline(root *graph.Instance, o Outline, resolver primitiveResolver) ([]OutlineFinding, int, int) {
	var findings []OutlineFinding
	add := func(kind, part, child, format string, args ...any) {
		findings = append(findings, OutlineFinding{Kind: kind, Part: part, Child: child, Message: fmt.Sprintf(format, args...)})
	}

	sch := root.Schema()
	planned := map[string]string{}
	matched := 0
	for name := range o.Parts {
		id, ok := subgraphIDFor(sch, name)
		if !ok {
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
		if !plannedIDs[id] {
			add("extra-subgraph", "", "", "subgraph %q is not in the outline; add it to the plan (set_outline again) or say in the report why it exists", id)
		}
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
		child, e := root.SubGraphInstance(id)
		if e != nil {
			continue
		}
		childSchema := child.Schema()

		inputs := map[string]bool{}
		repeaters := 0
		instances := map[string]int{}
		leafTypes := map[string]int{}
		for _, n := range childSchema.Nodes {
			if n.SubGraphInputBoundary != nil {
				inputs[normalizePortKey(n.SubGraphInputBoundary.PortName)] = true
			}
			if n.SubGraphId != "" {
				instances[n.SubGraphId]++
				if !plannedIDs[n.SubGraphId] || helperIDs[n.SubGraphId] {
					// A helper instance (a tapered curve, an equation,
					// planned as kind: helper or not planned at all) has
					// no geometry of its own; its interior counts as
					// this part's own leaves.
					if helper, ok := sch.SubGraphs[n.SubGraphId]; ok {
						for _, hn := range helper.Nodes {
							leafTypes[hn.Type]++
						}
					}
				}
			}
			if isRepeaterType(n.Type) {
				repeaters++
			}
			leafTypes[n.Type]++
		}
		for portName := range part.Inputs {
			if !inputs[normalizePortKey(portName)] {
				add("missing-input", name, "", "outline declares input %q on %q but the subgraph has no such boundary input", portName, name)
			}
		}

		for childName, c := range part.Children {
			if c.Primitive != "" {
				key, err := resolver.resolve(c.Primitive)
				if err != nil {
					continue
				}
				have := leafTypes[key]
				switch {
				case have == 0:
					add("missing-leaf", name, childName, "%q should contain a %s for leaf %q; none found", name, c.Primitive, childName)
				case have < c.Count && repeaters == 0:
					add("count-mismatch", name, childName, "leaf %q plans %d x %s but %q holds %d and no mirror/repeat node", childName, c.Count, c.Primitive, name, have)
				}
				continue
			}
			target := c.Ref
			if target == "" {
				target = childName
			}
			targetID, ok := planned[target]
			if !ok {
				continue
			}
			have := instances[targetID]
			switch {
			case have == 0:
				add("inline-part", name, childName, "%q never instantiates %q; the outline says it holds %d - was it built inline instead? convert_to_subgraph the cluster, then instantiate_subgraph it here", name, target, c.Count)
			case have < c.Count && repeaters == 0:
				add("count-mismatch", name, childName, "%q plans %d x %q but holds %d instance(s) and no mirror/repeat node", name, c.Count, target, have)
			}
		}

		if structural := structuralNodeCount(childSchema); len(part.Children) > 0 && structural > 60 {
			add("oversized-leaf", name, "", "%q holds %d nodes that do work (of %d in total); a part that big is hiding parts that should be their own subgraphs", name, structural, len(childSchema.Nodes))
		}
	}

	clutter := 0
	for _, n := range sch.Nodes {
		if isAssemblyType(n) {
			continue
		}
		clutter++
	}
	if clutter > 0 {
		add("root-clutter", "", "", "%d node(s) in the root graph are neither assembly (Model/Manifest/Material/March), instances, literals nor variable references; geometry belongs inside a part subgraph", clutter)
	}

	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].Kind != findings[j].Kind {
			return findings[i].Kind < findings[j].Kind
		}
		return findings[i].Part+findings[i].Child < findings[j].Part+findings[j].Child
	})
	return findings, matched, len(o.Parts)
}

// checkSofteningBudget compares every march, smooth union and normal weld
// in the graph against the thinnest feature that actually reaches it.
func checkSofteningBudget(root *graph.Instance, plan OutlineStats) []OutlineFinding {
	if plan.ThinnestWorld <= 0 {
		return nil
	}
	var findings []OutlineFinding
	add := func(kind, scope, nodeID, format string, args ...any) {
		findings = append(findings, OutlineFinding{Kind: kind, Part: scope, Child: nodeID, Message: fmt.Sprintf(format, args...)})
	}

	scopes := map[string]*graph.Instance{"root": root}
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

	tracer := scaleTracer{root: root, scopes: scopes, memo: map[string]float64{}}

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
			budget := func() (thinnestLeaf, float64, float64) {
				return tracer.budgetForNode(plan, partForScope, scope, id, "")
			}
			switch {
			case strings.Contains(typeKey, "marching.MarchNode"):
				thinnest, minResolution, _ := budget()
				if res, ok := inputFloat(node, "Resolution"); ok && res < minResolution {
					add("under-resolved", scope, id, "MarchNode %s in %q has Resolution %v (voxels %.4f across); the thinnest feature it covers, %s, is %.4f, which needs Resolution >= %v to span %v voxels", id, scope, res, 1/res, thinnest.path, thinnest.world, minResolution, voxelsAcrossThinnest)
				}
				if f := splitFinding(plan, partForScope[scope], scope, id, node); f != nil {
					findings = append(findings, *f)
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
				thinnest, _, maxBlend := budget()
				scale := tracer.downstream(scope, id, nil)
				if world := r * scale; world > maxBlend {
					add("blend-too-wide", scope, id, "%s in %q blends with Radius %v (x%.3g downstream scale = %.4f in world units); the thinnest feature it covers, %s, is %.4f, so any world radius above %.4f bridges or swallows it", id, scope, r, scale, world, thinnest.path, thinnest.world, maxBlend)
				}
			case strings.Contains(typeKey, "SmoothNormalsImplicitWeldNode"):
				thinnest, _, maxBlend := budget()
				if d, ok := inputFloat(node, "Distance"); ok && d > maxBlend {
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
	scopes map[string]*graph.Instance
	memo   map[string]float64
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
				factor := 1.0
				if strings.Contains(n.Type, "sdf.TransformNode") {
					factor = transformScale(inst.Node(consumerID))
				}
				consider(factor * t.downstream(scope, consumerID, stack))
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
func (t scaleTracer) budgetForNode(plan OutlineStats, partForScope map[string]string, scope, nodeID, port string) (thinnestLeaf, float64, float64) {
	best := thinnestLeaf{}
	for reached := range t.feedingScopes(scope, nodeID, port) {
		leaf, ok := plan.partThinnest[partForScope[reached]]
		if !ok {
			continue
		}
		if best.world == 0 || leaf.world < best.world {
			best = leaf
		}
	}
	if best.world == 0 {
		return plan.budgetFor("")
	}
	return best, math.Ceil(voxelsAcrossThinnest / best.world), best.world / 2
}

func transformScale(node nodes.Node) float64 {
	in, ok := node.Inputs()["Transform"].(nodes.SingleValueInputPort)
	if !ok || in.Value() == nil {
		return 1
	}
	out, ok := in.Value().(nodes.Output[trs.TRS])
	if !ok {
		return 1
	}
	s := out.Value().Scale()
	return math.Max(math.Abs(s.X()), math.Max(math.Abs(s.Y()), math.Abs(s.Z())))
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
	for _, n := range sch.Nodes {
		if !isAssemblyType(n) {
			stats.RootGeometry++
		}
	}

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
		Name:        "check_outline",
		Description: "Diff the graph against the stored outline: parts with no subgraph, subgraphs not in the plan, declared inputs with no boundary port, children the parent never instantiates (built inline?), counts that don't match with no mirror/repeat node to explain it, oversized parts, a march whose resolution is set by one thin feature it pays for over a whole body, geometry sitting in the root graph, a MarchNode whose Resolution is too coarse for the thinnest feature whose geometry actually reaches it, and a SmoothUnion Radius or normal-weld Distance wider than that feature allows. Budgets follow the graph, not the part tree, so a separately marched thin part does not constrain a march it never feeds. Run it before save_graph and put its findings in the report.",
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
// leaf, and the cost of the mismatch is cubic: a tooth that needs twice the
// resolution makes every voxel of the head eight times as many. The two
// belong in separate marches - a small dense domain for the detail, a big
// coarse one for the mass - which is what separate: true is for.
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
			"MarchNode %s spans %.2fx%.2fx%.2f at Resolution %v, which is %s voxels every time it runs. %s (%.4f thick) is what forces that resolution; everything else in %q is at least %.4f and would only need %v, or %s voxels - %.0fx cheaper. Mark %s separate: true and march it on its own over a domain that just covers it, so the detail is dense and the mass is not.",
			nodeID, size.X(), size.Y(), size.Z(), res, formatVoxels(have),
			thin.path, thin.world, part, next.world, without, formatVoxels(would), have/would, thin.path),
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
