package mcp

import (
	"cmp"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/EliCDavis/polyform/generator/schema"
	"gopkg.in/yaml.v3"
)

const outlineMetadataKey = "outline"

// Outline is the build plan an agent commits to before touching the
// graph. Every named part becomes a subgraph; every child records how
// many of it there are and how big it is relative to its parent, so
// depth and repetition are declared decisions a tool can check rather
// than judgment calls that quietly default to "one blob".
type Outline struct {
	Version    int                    `yaml:"version" json:"version"`
	Subject    string                 `yaml:"subject" json:"subject"`
	Style      string                 `yaml:"style" json:"style"`
	Extent     float64                `yaml:"extent" json:"extent"`
	Thresholds OutlineThresholds      `yaml:"thresholds,omitempty" json:"thresholds"`
	Parts      map[string]OutlinePart `yaml:"parts" json:"parts"`
	Accepted   []OutlineAcceptance    `yaml:"accepted,omitempty" json:"accepted,omitempty"`
}

// OutlineAcceptance is a check_outline finding the plan knowingly keeps:
// a part that is big because it is data-driven, a blend that is wide on
// purpose. It silences that finding until the plan changes, instead of
// having it re-explained every round.
type OutlineAcceptance struct {
	Kind   string `yaml:"kind" json:"kind"`
	Part   string `yaml:"part,omitempty" json:"part,omitempty"`
	Child  string `yaml:"child,omitempty" json:"child,omitempty"`
	Reason string `yaml:"reason" json:"reason"`
}

var outlineFindingKinds = map[string]bool{
	"missing-part": true, "extra-subgraph": true, "missing-input": true, "count-mismatch": true,
	"missing-leaf": true, "inline-part": true, "oversized-leaf": true, "root-clutter": true,
	"under-resolved": true, "blend-too-wide": true, "thin-feature-sets-resolution": true,
}

// covers reports whether a finding is the one this acceptance names. An
// empty part or child matches any; a part is matched the way parts are
// matched to subgraphs, since budget findings name the scope.
func (a OutlineAcceptance) covers(f OutlineFinding) bool {
	if a.Kind != f.Kind {
		return false
	}
	same := func(want, got string) bool {
		return want == "" || normalizePortKey(strings.ReplaceAll(want, "_", "")) == normalizePortKey(strings.ReplaceAll(got, "_", ""))
	}
	return same(a.Part, f.Part) && same(a.Child, f.Child)
}

// voxelsAcrossThinnest is how many marching voxels the thinnest planned
// feature must span for it to survive the march with its shape.
const voxelsAcrossThinnest = 4.0

const (
	// How much thicker the rest of a part has to be before its thinnest
	// leaf counts as the one feature dragging the whole march up.
	outlierThicknessRatio = 2.0

	// Below this a march is cheap enough that splitting it is not worth
	// the extra node, whatever the ratio.
	voxelsWorthSplitting = 2_000_000.
)

type OutlineThresholds struct {
	// A part whose absolute size is above this must have children.
	Decompose float64 `yaml:"decompose" json:"decompose"`
	// A part whose absolute size is below this should be a leaf.
	Leaf float64 `yaml:"leaf" json:"leaf"`
	// Roughly how many triangles one march may produce before the plan is
	// refused. A march resolves the whole of its part at the resolution
	// its thinnest feature demands, so a small detail on a large mass
	// multiplies out across the entire surface.
	Triangles float64 `yaml:"triangles" json:"triangles"`
}

// defaultTriangleBudget is what a part may cost before the plan is refused.
// A real-time prop runs 5k-50k triangles; this sits well above that so a
// dense hero asset still passes, and only catches the case where one thin
// feature has multiplied a whole body into something unusable.
const defaultTriangleBudget = 50_000.

// estimateTriangles is roughly what marching a blob of this world size at
// this resolution produces: about two triangles per surface cell, over the
// area of a sphere that size. Only the order of magnitude matters - the
// question is whether a plan is 20k or 400k, not 391k against 406k.
func estimateTriangles(worldSize, resolution float64) float64 {
	area := math.Pi * worldSize * worldSize
	return 2 * area * resolution * resolution
}

type OutlinePart struct {
	Kind     string                  `yaml:"kind,omitempty" json:"kind,omitempty"`
	Material string                  `yaml:"material,omitempty" json:"material,omitempty"`
	Inputs   map[string]string       `yaml:"inputs,omitempty" json:"inputs,omitempty"`
	Children map[string]OutlineChild `yaml:"children,omitempty" json:"children,omitempty"`
}

type OutlineChild struct {
	Count     int     `yaml:"count" json:"count"`
	Size      float64 `yaml:"size" json:"size"`
	Thickness float64 `yaml:"thickness,omitempty" json:"thickness,omitempty"`
	Layout    string  `yaml:"layout,omitempty" json:"layout,omitempty"`
	Primitive string  `yaml:"primitive,omitempty" json:"primitive,omitempty"`
	Ref       string  `yaml:"ref,omitempty" json:"ref,omitempty"`
	Material  string  `yaml:"material,omitempty" json:"material,omitempty"`
	Whole     bool    `yaml:"whole,omitempty" json:"whole,omitempty"`
	Paint     bool    `yaml:"paint,omitempty" json:"paint,omitempty"`
	Separate  bool    `yaml:"separate,omitempty" json:"separate,omitempty"`
	Relief    bool    `yaml:"relief,omitempty" json:"relief,omitempty"`
}

type outlineMesh struct {
	root     string
	material string
}

func describeMaterial(material string) string {
	if material == "" {
		return "the unnamed default material"
	}
	return fmt.Sprintf("material %q", material)
}

var outlineLayouts = map[string]bool{
	"mirror-x": true, "mirror-y": true, "mirror-z": true,
	"mirror-xz": true, "mirror-xy": true, "mirror-yz": true,
	"line": true, "arc": true, "ring": true, "grid": true, "free": true,
}

// OutlineIssue is one validation finding. Errors block set_outline;
// warnings are returned alongside a successful set.
type OutlineIssue struct {
	Level   string `json:"level" jsonschema:"error or warning"`
	Part    string `json:"part,omitempty" jsonschema:"the part definition the issue is in"`
	Child   string `json:"child,omitempty" jsonschema:"the child entry within that part, when the issue is about one"`
	Message string `json:"message"`
}

// OutlineNodeStat is one resolved node in the outline tree, with its
// absolute size (product of every size on the path from the subject).
type OutlineNodeStat struct {
	Path         string  `json:"path" jsonschema:"subject/child/grandchild"`
	Part         string  `json:"part,omitempty" jsonschema:"the part definition this resolves to; empty for a primitive leaf"`
	Primitive    string  `json:"primitive,omitempty" jsonschema:"resolved node type key for a leaf"`
	Count          int     `json:"count"`
	AbsoluteSize   float64 `json:"absoluteSize"`
	WorldSize      float64 `json:"worldSize" jsonschema:"absoluteSize times the subject's extent, in world units"`
	WorldThickness float64 `json:"worldThickness,omitempty" jsonschema:"the smallest cross-section in world units, for a leaf"`
	Depth          int     `json:"depth"`
	Leaf           bool    `json:"leaf"`
}

type OutlineStats struct {
	Parts           int     `json:"parts" jsonschema:"distinct part definitions (each becomes a subgraph)"`
	Leaves          int     `json:"leaves" jsonschema:"distinct primitive leaf entries"`
	MaxDepth        int     `json:"maxDepth"`
	TotalInstances  int     `json:"totalInstances" jsonschema:"every placement counted, with counts multiplied down the tree"`
	Repeated        int     `json:"repeated" jsonschema:"children with count > 1"`
	MeanDepth       float64 `json:"meanDepth" jsonschema:"mean depth of the leaves, weighted by instance count"`
	ThinnestFeature string  `json:"thinnestFeature,omitempty" jsonschema:"path of the leaf with the smallest world thickness"`
	ThinnestWorld   float64 `json:"thinnestWorld,omitempty" jsonschema:"that thickness in world units; every softening length in the build is bounded by it"`
	MinResolution   float64 `json:"minResolution,omitempty" jsonschema:"the smallest MarchNode Resolution (voxels per unit) that keeps the thinnest feature; check_outline flags a march below it"`
	MaxBlend        float64 `json:"maxBlend,omitempty" jsonschema:"upper bound for SmoothUnionNode Radius, SmoothNormalsImplicitWeldNode Distance and RoundCube Roundness: half the thinnest feature"`

	partThinnest map[string]thinnestLeaf
	partLeaves   map[string][]thinnestLeaf
	partSize     map[string]float64
}

// outlierLeaf is the part's thinnest field leaf, and the next one up, when
// the thinnest is far enough below the rest to be setting the resolution
// for geometry that does not need it.
func (s OutlineStats) outlierLeaf(part string) (thin, next thinnestLeaf, ok bool) {
	leaves := s.partLeaves[part]
	if len(leaves) < 2 {
		return thin, next, false
	}

	sorted := append([]thinnestLeaf(nil), leaves...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].world < sorted[j].world })

	thin = sorted[0]
	for _, leaf := range sorted[1:] {
		if leaf.world > thin.world {
			next = leaf
			break
		}
	}
	if next.world == 0 || thin.world <= 0 {
		return thin, next, false
	}
	return thin, next, next.world/thin.world >= outlierThicknessRatio
}

type thinnestLeaf struct {
	path      string
	world     float64
	primitive string
}

// budgetFor is the thinnest field leaf inside part's own subtree, falling
// back to the whole plan's for scopes the outline doesn't name.
func (s OutlineStats) budgetFor(part string) (thinnestLeaf, float64, float64) {
	leaf, ok := s.partThinnest[part]
	if !ok {
		return thinnestLeaf{path: s.ThinnestFeature, world: s.ThinnestWorld}, s.MinResolution, s.MaxBlend
	}
	return leaf, math.Ceil(voxelsAcrossThinnest / leaf.world), leaf.world / 2
}

func isBoxPrimitive(typeKey string) bool {
	return strings.Contains(typeKey, "Cube")
}

var unknownOutlineField = regexp.MustCompile(`^(line \d+): field (.+) not found in type mcp\.(\w+)$`)

var outlineTypeNames = map[string]string{
	"Outline":           "outline",
	"OutlineThresholds": "thresholds",
	"OutlineAcceptance": "accepted entry",
	"OutlinePart":       "part",
	"OutlineChild":      "child",
}

func decodeOutline(text string, o *Outline) error {
	dec := yaml.NewDecoder(strings.NewReader(text))
	dec.KnownFields(true)
	err := dec.Decode(o)
	var typeErr *yaml.TypeError
	if !errors.As(err, &typeErr) {
		return err
	}
	msgs := make([]string, len(typeErr.Errors))
	for i, msg := range typeErr.Errors {
		msgs[i] = msg
		match := unknownOutlineField.FindStringSubmatch(msg)
		switch {
		case match == nil:
		case strings.Contains(match[2], " "):
			msgs[i] = fmt.Sprintf("%s: unquoted comma split %q", match[1], match[2])
		default:
			msgs[i] = fmt.Sprintf("%s: unknown key %q in %s", match[1], match[2], cmp.Or(outlineTypeNames[match[3]], match[3]))
		}
	}
	return errors.New(strings.Join(msgs, "; "))
}

func parseOutline(text string) (Outline, error) {
	var o Outline
	if err := decodeOutline(text, &o); err != nil {
		return o, fmt.Errorf("outline is not valid YAML for this schema: %w", err)
	}
	if o.Thresholds.Decompose == 0 {
		o.Thresholds.Decompose = 0.25
	}
	if o.Thresholds.Leaf == 0 {
		o.Thresholds.Leaf = 0.02
	}
	return o, nil
}

func (o Outline) kindOf(part OutlinePart) string {
	if part.Kind != "" {
		return part.Kind
	}
	if o.Style == "mechanical" {
		return "mesh"
	}
	return "field"
}

// primitiveResolver turns a leaf's primitive spelling into a registered
// node type key: the full key, or "<path tail>.<DisplayName>" such as
// sdf.Sphere or primitives.UvSphere (case-insensitive, "Node" optional).
type primitiveResolver struct {
	byKey   map[string]schema.NodeType
	byShort map[string][]string
}

func newPrimitiveResolver(types []schema.NodeType) primitiveResolver {
	r := primitiveResolver{byKey: map[string]schema.NodeType{}, byShort: map[string][]string{}}
	for _, t := range types {
		r.byKey[t.Type] = t
		name := normalizePortKey(strings.TrimSuffix(strings.ReplaceAll(t.DisplayName, " ", ""), "Node"))
		names := []string{name}
		// A generic node ("Multiply[float64]") also answers to its bare
		// name; resolve() breaks the tie in favour of float64.
		if bare, _, generic := strings.Cut(name, "["); generic {
			names = append(names, bare)
		}
		segments := strings.Split(t.Path, "/")
		for i := range segments {
			for _, n := range names {
				short := strings.ToLower(strings.Join(segments[i:], "/")) + "." + n
				r.byShort[short] = append(r.byShort[short], t.Type)
			}
		}
	}
	return r
}

func (r primitiveResolver) resolve(spelling string) (string, error) {
	if _, ok := r.byKey[spelling]; ok {
		return spelling, nil
	}
	pkg, name, found := strings.Cut(spelling, ".")
	if !found {
		return "", fmt.Errorf("primitive %q must be spelled <package>.<Name>, e.g. sdf.Sphere or primitives.Cube", spelling)
	}
	name = strings.TrimSuffix(name, "Node")
	short := strings.ToLower(pkg) + "." + normalizePortKey(name)
	candidates := r.byShort[short]
	switch len(candidates) {
	case 0:
		return "", fmt.Errorf("primitive %q matches no registered node type; check search_node_types", spelling)
	case 1:
		return candidates[0], nil
	default:
		var floats []string
		for _, c := range candidates {
			if strings.Contains(c, "[float64]") {
				floats = append(floats, c)
			}
		}
		if len(floats) == 1 {
			return floats[0], nil
		}
		sort.Strings(candidates)
		return "", fmt.Errorf("primitive %q is ambiguous between %s; use the full type key", spelling, strings.Join(candidates, ", "))
	}
}

// validateOutline checks structure and the depth rules, and walks the
// tree to produce per-node stats. Structural errors stop the walk.
func validateOutline(o Outline, resolver primitiveResolver) ([]OutlineIssue, []OutlineNodeStat, OutlineStats) {
	var issues []OutlineIssue
	errorf := func(part, child, format string, args ...any) {
		issues = append(issues, OutlineIssue{Level: "error", Part: part, Child: child, Message: fmt.Sprintf(format, args...)})
	}
	warnf := func(part, child, format string, args ...any) {
		issues = append(issues, OutlineIssue{Level: "warning", Part: part, Child: child, Message: fmt.Sprintf(format, args...)})
	}

	if o.Version != 1 {
		errorf("", "", "version must be 1")
	}
	if o.Subject == "" {
		errorf("", "", "subject is required: the part name the whole build resolves to")
	}
	if o.Style != "organic" && o.Style != "mechanical" {
		errorf("", "", "style must be organic or mechanical, got %q", o.Style)
	}
	if o.Extent <= 0 {
		errorf("", "", "extent is required: the subject's longest dimension in world units; every size below is a fraction of it")
	}
	if len(o.Parts) == 0 {
		errorf("", "", "parts is empty")
	}
	if subject, ok := o.Parts[o.Subject]; o.Subject != "" && !ok {
		errorf("", "", "subject %q is not defined under parts", o.Subject)
	} else if ok && subject.Material == "" {
		warnf(o.Subject, "", "name the subject's material (material: <name>). Parts and leaves inherit it; anything that shines differently names its own, and cannot share a march with a different material")
	}

	resolvedPrimitives := map[string]string{}
	referenced := map[string]bool{o.Subject: true}

	for name, part := range o.Parts {
		if strings.TrimSpace(name) == "" {
			errorf(name, "", "part names cannot be blank")
		}
		if strings.Contains(name, " ") {
			errorf(name, "", "part names become subgraph ids; use camelCase, no spaces")
		}
		k := o.kindOf(part)
		if k != "field" && k != "mesh" && k != "helper" {
			errorf(name, "", "kind must be field, mesh or helper, got %q", k)
		}
		if k == "helper" {
			// A helper (an equation, a pivot, a tapered-curve wrapper)
			// has no geometry of its own: it is matched by name and
			// exempt from the size and children rules.
			if len(part.Children) > 0 {
				warnf(name, "", "a helper has no children; they are ignored")
			}
			referenced[name] = true
			continue
		}
		if len(part.Children) == 0 {
			errorf(name, "", "a part must have children; a piece that is one primitive belongs as a leaf child (primitive: ...) of its parent, not as a part")
		}
		for childName, child := range part.Children {
			target := child.Ref
			if target == "" {
				target = childName
			}
			_, isPart := o.Parts[target]
			isLeaf := child.Primitive != ""

			switch {
			case isPart && isLeaf:
				errorf(name, childName, "is both a reference to part %q and a primitive leaf; pick one", target)
			case !isPart && !isLeaf:
				errorf(name, childName, "is neither a part defined under parts nor a leaf with a primitive. Name the primitive it is made of, or define %q as a part with its own children", target)
			case isPart:
				referenced[target] = true
			case isLeaf:
				if _, done := resolvedPrimitives[child.Primitive]; !done {
					key, err := resolver.resolve(child.Primitive)
					if err != nil {
						errorf(name, childName, "%v", err)
					}
					resolvedPrimitives[child.Primitive] = key
				}
				if child.Thickness <= 0 && !child.Paint {
					errorf(name, childName, "thickness is required on a leaf: its smallest cross-section as a fraction of its own size (1 for a ball, about 0.25 for a rod four times longer than wide, 0.02 for a fine filament); it sets the march resolution and every blend radius")
				}
				if o.Style == "organic" && o.kindOf(part) == "field" && isBoxPrimitive(resolvedPrimitives[child.Primitive]) {
					warnf(name, childName, "%s is a box-family primitive in an organic part; unless this piece really has hard corners, match on topology (sdf.Sphere / sdf.Line / sdf.RoundedCone plus a TransformNode scale) or the part reads as a box", child.Primitive)
				}
			}
			if child.Thickness > 1 {
				warnf(name, childName, "thickness %.2f is above 1; it is the smallest cross-section as a fraction of this child's own size, so it cannot exceed 1", child.Thickness)
			}

			if child.Count < 1 {
				errorf(name, childName, "count is required and must be at least 1 (how many of this child the part places)")
			}
			if child.Size <= 0 && o.kindOf(o.Parts[target]) != "helper" {
				errorf(name, childName, "size is required: this child's extent as a fraction of %q's extent, e.g. 0.3", name)
			} else if child.Size > 1.5 {
				warnf(name, childName, "size %.2f means this child is much larger than its parent; check the fraction", child.Size)
			}
			if child.Count > 1 {
				if child.Layout == "" {
					errorf(name, childName, "count %d needs a layout saying how the copies are placed: mirror-x, mirror-xz, line, arc, ring, grid, or free", child.Count)
				} else if !outlineLayouts[child.Layout] {
					errorf(name, childName, "layout %q is not one of mirror-x/y/z/xz/xy/yz, line, arc, ring, grid, free", child.Layout)
				}
				// A mirror doubles per axis, but what it mirrors can be
				// several copies already: three teeth per side is count 6
				// on mirror-x. Divisibility is the rule, not equality.
				if per := mirrorCopies(child.Layout); per > 0 && child.Count%per != 0 {
					warnf(name, childName, "mirror-%s makes %d copies of whatever it mirrors, so count has to be a multiple of %d (%d per side); %d is not",
						strings.TrimPrefix(child.Layout, "mirror-"), per, per, child.Count/per+1, child.Count)
				}
			}
		}
	}

	for name := range o.Parts {
		if !referenced[name] {
			warnf(name, "", "defined but never used as a child of anything")
		}
	}

	for i, a := range o.Accepted {
		switch {
		case !outlineFindingKinds[a.Kind]:
			errorf(a.Part, a.Child, "accepted[%d]: kind %q is not a check_outline finding kind", i, a.Kind)
		case strings.TrimSpace(a.Reason) == "":
			errorf(a.Part, a.Child, "accepted[%d]: a reason is required - say why the %s finding is right for this build", i, a.Kind)
		}
	}

	hasErrors := false
	for _, issue := range issues {
		if issue.Level == "error" {
			hasErrors = true
			break
		}
	}
	if hasErrors {
		sortIssues(issues)
		return issues, nil, OutlineStats{}
	}

	// Walk the tree for sizes and depth. Cycles are an error found here
	// rather than above, since they need the same traversal.
	var nodes []OutlineNodeStat
	stats := OutlineStats{
		Parts:        len(o.Parts),
		partThinnest: map[string]thinnestLeaf{},
		partLeaves:   map[string][]thinnestLeaf{},
		partSize:     map[string]float64{},
	}
	leafSet := map[string]bool{}
	var depthWeighted float64
	var leafInstances int

	sharesMarch := func(part OutlinePart, child OutlineChild) bool {
		return o.kindOf(part) == "field" && !child.Separate
	}
	refuseMaterial := func(partName, childName, childPath, material string, mesh outlineMesh) {
		errorf(partName, childName, "%s is %s but is marched with %s, which is %s. One march is one mesh with one material: paint changes its color, never how it shines. Make %s its own mesh - a kind: mesh part, or separate: true - with its own MaterialNode, seated into %s rather than blended into it",
			childPath, describeMaterial(material), mesh.root, describeMaterial(mesh.material), childName, mesh.root)
	}

	var walk func(path, partName string, absSize float64, depth, instances int, stack []string, mesh outlineMesh)
	walk = func(path, partName string, absSize float64, depth, instances int, stack []string, mesh outlineMesh) {
		for _, s := range stack {
			if s == partName {
				errorf(partName, "", "part cycle: %s -> %s", strings.Join(stack, " -> "), partName)
				return
			}
		}
		part := o.Parts[partName]
		if depth > stats.MaxDepth {
			stats.MaxDepth = depth
		}
		if absSize > stats.partSize[partName] {
			stats.partSize[partName] = absSize
		}
		pieces := 0
		for _, child := range part.Children {
			pieces += child.Count
		}
		if absSize > o.Thresholds.Decompose && pieces < 2 {
			errorf(partName, "", "%s is %.0f%% of the subject but holds %d piece; a part that large needs real structure (2+ children, or one child repeated)", path, absSize*100, pieces)
		}
		childNames := make([]string, 0, len(part.Children))
		for n := range part.Children {
			childNames = append(childNames, n)
		}
		sort.Strings(childNames)
		for _, childName := range childNames {
			child := part.Children[childName]
			childAbs := absSize * child.Size
			childPath := path + "/" + childName
			childInstances := instances * child.Count
			if child.Count > 1 {
				stats.Repeated++
			}
			if child.Primitive != "" {
				// whole: true is the plan saying this leaf really is one
				// primitive at that size, so it is not asked again below
				// the refusal line; above it the refusal becomes a warning.
				switch {
				case childAbs > 2*o.Thresholds.Decompose && !child.Whole && !child.Paint:
					errorf(partName, childName, "%s is %.0f%% of the subject but is a single %s; anything above %.0f%% must be a part with children, unless it really is a single primitive at its own size - then say whole: true", childPath, childAbs*100, child.Primitive, 2*o.Thresholds.Decompose*100)
				case childAbs > 2*o.Thresholds.Decompose && child.Whole:
					warnf(partName, childName, "%s is %.0f%% of the subject as a single %s; fine for a core mass surrounded by siblings, wrong if it is the whole thing", childPath, childAbs*100, child.Primitive)
				case childAbs > o.Thresholds.Decompose && !child.Paint && !child.Whole:
					warnf(partName, childName, "%s is %.0f%% of the subject as a single %s; fine for a core mass surrounded by siblings, wrong if it is the whole thing", childPath, childAbs*100, child.Primitive)
				}
				if child.Material != "" && child.Material != mesh.material && sharesMarch(part, child) {
					refuseMaterial(partName, childName, childPath, child.Material, mesh)
				}
				thickness := childAbs * child.Thickness * o.Extent
				// Only field leaves go through the shared march; a mesh
				// part's whisker sets no voxel budget, a paint blob only
				// colors vertices that are already there, and a leaf
				// marched on its own (a tongue) sizes its own march.
				if o.kindOf(part) == "field" && !child.Paint && !child.Separate && !child.Relief {
					if stats.ThinnestFeature == "" || thickness < stats.ThinnestWorld {
						stats.ThinnestFeature = childPath
						stats.ThinnestWorld = thickness
					}
					leaf := thinnestLeaf{path: childPath, world: thickness, primitive: resolvedPrimitives[child.Primitive]}
					for _, owner := range append(stack, partName) {
						if cur, ok := stats.partThinnest[owner]; !ok || thickness < cur.world {
							stats.partThinnest[owner] = leaf
						}
						stats.partLeaves[owner] = append(stats.partLeaves[owner], leaf)
					}
				}
				nodes = append(nodes, OutlineNodeStat{
					Path: childPath, Primitive: resolvedPrimitives[child.Primitive],
					Count: child.Count, AbsoluteSize: childAbs, WorldSize: childAbs * o.Extent,
					WorldThickness: thickness, Depth: depth + 1, Leaf: true,
				})
				leafSet[childPath] = true
				if depth+1 > stats.MaxDepth {
					stats.MaxDepth = depth + 1
				}
				stats.TotalInstances += childInstances
				depthWeighted += float64(depth+1) * float64(childInstances)
				leafInstances += childInstances
				continue
			}
			target := child.Ref
			if target == "" {
				target = childName
			}
			if o.kindOf(o.Parts[target]) == "helper" {
				nodes = append(nodes, OutlineNodeStat{Path: childPath, Part: target, Count: child.Count, Depth: depth + 1})
				continue
			}
			if childAbs < o.Thresholds.Leaf {
				warnf(partName, childName, "%s is %.1f%% of the subject and still decomposed into part %q; below %.0f%% this is sub-pixel detail", childPath, childAbs*100, target, o.Thresholds.Leaf*100)
			}
			nodes = append(nodes, OutlineNodeStat{
				Path: childPath, Part: target,
				Count: child.Count, AbsoluteSize: childAbs, WorldSize: childAbs * o.Extent, Depth: depth + 1,
			})
			stats.TotalInstances += childInstances

			// separate: true on a part is a march boundary. What is inside
			// it sizes its own march, so its thin features stop setting the
			// resolution for everything above - which is the only way a
			// detailed face on a large body is affordable.
			owners := append(stack, partName)
			if child.Separate {
				owners = nil
			}
			childMesh := mesh
			material := cmp.Or(child.Material, o.Parts[target].Material)
			switch {
			case !sharesMarch(part, child) || o.kindOf(o.Parts[target]) != "field":
				childMesh = outlineMesh{root: childPath, material: cmp.Or(material, mesh.material)}
			case material != "" && material != mesh.material:
				refuseMaterial(partName, childName, childPath, material, mesh)
			}
			walk(childPath, target, childAbs, depth+1, childInstances, owners, childMesh)
		}
	}
	nodes = append(nodes, OutlineNodeStat{Path: o.Subject, Part: o.Subject, Count: 1, AbsoluteSize: 1, WorldSize: o.Extent, Depth: 0})
	walk(o.Subject, o.Subject, 1, 0, 1, nil, outlineMesh{root: o.Subject, material: o.Parts[o.Subject].Material})

	stats.Leaves = len(leafSet)
	if leafInstances > 0 {
		stats.MeanDepth = depthWeighted / float64(leafInstances)
	}
	if stats.ThinnestWorld > 0 {
		stats.MinResolution = math.Ceil(voxelsAcrossThinnest / stats.ThinnestWorld)
		stats.MaxBlend = stats.ThinnestWorld / 2
	}
	for _, issue := range affordabilityIssues(o, stats) {
		issues = append(issues, issue)
	}
	sortIssues(issues)
	return issues, nodes, stats
}

func sortIssues(issues []OutlineIssue) {
	sort.SliceStable(issues, func(i, j int) bool {
		if issues[i].Level != issues[j].Level {
			return issues[i].Level == "error"
		}
		if issues[i].Part != issues[j].Part {
			return issues[i].Part < issues[j].Part
		}
		return issues[i].Child < issues[j].Child
	})
}

func hasOutlineErrors(issues []OutlineIssue) bool {
	for _, issue := range issues {
		if issue.Level == "error" {
			return true
		}
	}
	return false
}

// mirrorCopies is how many copies a mirror layout produces: two for one
// axis, four for two. Zero when the layout is not a mirror.
func mirrorCopies(layout string) int {
	axes, ok := strings.CutPrefix(layout, "mirror-")
	if !ok {
		return 0
	}
	switch len(axes) {
	case 1:
		return 2
	case 2:
		return 4
	default:
		return 0
	}
}

func affordabilityIssues(o Outline, stats OutlineStats) []OutlineIssue {
	budget := o.Thresholds.Triangles
	if budget <= 0 {
		budget = defaultTriangleBudget
	}

	var issues []OutlineIssue
	for part, thin := range stats.partThinnest {
		size := stats.partSize[part] * o.Extent
		if size <= 0 || thin.world <= 0 {
			continue
		}

		resolution := math.Ceil(voxelsAcrossThinnest / thin.world)
		triangles := estimateTriangles(size, resolution)
		if triangles <= budget {
			continue
		}

		relief := ""
		if _, next, ok := stats.outlierLeaf(part); ok && next.world > 0 {
			cheaper := estimateTriangles(size, math.Ceil(voxelsAcrossThinnest/next.world))
			relief = fmt.Sprintf(" Without it the rest of %q needs only Resolution %v, about %s triangles - %.0fx less.",
				part, math.Ceil(voxelsAcrossThinnest/next.world), formatCount(cheaper), triangles/cheaper)
		}

		issues = append(issues, OutlineIssue{
			Level: "warning", Part: part, Child: thin.path,
			Message: fmt.Sprintf(
				"%q is %.2f units across and its thinnest feature %s is %.4f, which forces Resolution %v over the whole of it - roughly %s triangles, past the %s budget.%s If %s is thinner than it needs to be, raise its thickness. If it is a distinct piece that should not blend into the surface, it belongs in its own mesh with separate: true. If it has to blend (a small bump, groove or ridge), it stays in this march: keep the cost, or set thresholds.triangles to say this is meant to be a dense asset.",
				part, size, thin.path, thin.world, resolution, formatCount(triangles), formatCount(budget), relief, thin.path),
		})
	}
	return issues
}

func formatCount(n float64) string {
	switch {
	case n >= 1e6:
		return fmt.Sprintf("%.1fM", n/1e6)
	case n >= 1e3:
		return fmt.Sprintf("%.0fk", n/1e3)
	default:
		return fmt.Sprintf("%.0f", n)
	}
}
