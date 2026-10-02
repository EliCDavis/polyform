package mcp

import (
	"fmt"
	"math"
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
	Thresholds OutlineThresholds      `yaml:"thresholds" json:"thresholds"`
	Parts      map[string]OutlinePart `yaml:"parts" json:"parts"`
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
}

type OutlinePart struct {
	Kind     string                  `yaml:"kind" json:"kind,omitempty"`
	Inputs   map[string]string       `yaml:"inputs" json:"inputs,omitempty"`
	Children map[string]OutlineChild `yaml:"children" json:"children,omitempty"`
}

type OutlineChild struct {
	Count     int     `yaml:"count" json:"count"`
	Size      float64 `yaml:"size" json:"size"`
	Thickness float64 `yaml:"thickness" json:"thickness,omitempty"`
	Layout    string  `yaml:"layout" json:"layout,omitempty"`
	Primitive string  `yaml:"primitive" json:"primitive,omitempty"`
	Ref       string  `yaml:"ref" json:"ref,omitempty"`
	Whole     bool    `yaml:"whole" json:"whole,omitempty"`
	Paint     bool    `yaml:"paint" json:"paint,omitempty"`
	Separate  bool    `yaml:"separate" json:"separate,omitempty"`
	Relief    bool    `yaml:"relief" json:"relief,omitempty"`
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
	path  string
	world float64
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

func parseOutline(text string) (Outline, error) {
	var o Outline
	dec := yaml.NewDecoder(strings.NewReader(text))
	dec.KnownFields(true)
	if err := dec.Decode(&o); err != nil {
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
		errorf("", "", "extent is required: the subject's longest dimension in world units (a house cat is about 0.5 long, a car about 4.5); every size below is a fraction of it")
	}
	if len(o.Parts) == 0 {
		errorf("", "", "parts is empty")
	}
	if _, ok := o.Parts[o.Subject]; o.Subject != "" && !ok {
		errorf("", "", "subject %q is not defined under parts", o.Subject)
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
					errorf(name, childName, "thickness is required on a leaf: its smallest cross-section as a fraction of its own size (a leg about 0.25, a torso about 0.6, a whisker 0.02); it sets the march resolution and every blend radius")
				}
				if o.Style == "organic" && o.kindOf(part) == "field" && isBoxPrimitive(resolvedPrimitives[child.Primitive]) {
					warnf(name, childName, "%s is a box-family primitive in an organic part; unless this piece really has hard corners, match on topology (sdf.Sphere / sdf.Capsule / sdf.RoundedCone plus a TransformNode scale) or the part reads as a brick", child.Primitive)
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
			if child.Count == 1 && strings.HasSuffix(childName, "s") && !strings.HasSuffix(childName, "ss") {
				warnf(name, childName, "plural name with count 1 - if there are several, say count: N with a layout; if one, name it in the singular")
			}
		}
	}

	for name := range o.Parts {
		if !referenced[name] {
			warnf(name, "", "defined but never used as a child of anything")
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
	stats := OutlineStats{Parts: len(o.Parts), partThinnest: map[string]thinnestLeaf{}, partLeaves: map[string][]thinnestLeaf{}}
	leafSet := map[string]bool{}
	var depthWeighted float64
	var leafInstances int

	var walk func(path, partName string, absSize float64, depth, instances int, stack []string)
	walk = func(path, partName string, absSize float64, depth, instances int, stack []string) {
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
		if absSize > o.Thresholds.Decompose && len(part.Children) < 2 {
			errorf(partName, "", "%s is %.0f%% of the subject with only %d child; a part that large needs real structure (2+ children)", path, absSize*100, len(part.Children))
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
				switch {
				case childAbs > 2*o.Thresholds.Decompose && !child.Whole && !child.Paint:
					errorf(partName, childName, "%s is %.0f%% of the subject but is a single %s; anything above %.0f%% must be a part with children, unless it really is one primitive (a cushion slab, a plate) - then say whole: true", childPath, childAbs*100, child.Primitive, 2*o.Thresholds.Decompose*100)
				case childAbs > o.Thresholds.Decompose && !child.Paint:
					warnf(partName, childName, "%s is %.0f%% of the subject as a single %s; fine for a core mass (a skull, a chest) surrounded by siblings, wrong if it is the whole thing", childPath, childAbs*100, child.Primitive)
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
					for _, owner := range append(stack, partName) {
						if cur, ok := stats.partThinnest[owner]; !ok || thickness < cur.world {
							stats.partThinnest[owner] = thinnestLeaf{path: childPath, world: thickness}
						}
						stats.partLeaves[owner] = append(stats.partLeaves[owner], thinnestLeaf{path: childPath, world: thickness})
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
			walk(childPath, target, childAbs, depth+1, childInstances, append(stack, partName))
		}
	}
	nodes = append(nodes, OutlineNodeStat{Path: o.Subject, Part: o.Subject, Count: 1, AbsoluteSize: 1, WorldSize: o.Extent, Depth: 0})
	walk(o.Subject, o.Subject, 1, 0, 1, nil)

	stats.Leaves = len(leafSet)
	if leafInstances > 0 {
		stats.MeanDepth = depthWeighted / float64(leafInstances)
	}
	if stats.ThinnestWorld > 0 {
		stats.MinResolution = math.Ceil(voxelsAcrossThinnest / stats.ThinnestWorld)
		stats.MaxBlend = stats.ThinnestWorld / 2
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
