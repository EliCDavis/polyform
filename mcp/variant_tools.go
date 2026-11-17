package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/generator/variant"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// variantDimensionTypeKeys documents the set of variant.Dimension types
// UnmarshalDimension knows how to decode. Kept here purely for the tool
// description text; the actual accepted set is enforced by
// variant.UnmarshalDimension itself, so this can't drift into allowing a
// type that doesn't really exist.
const variantDimensionTypeKeys = "discrete, numericRange, intRange, vector2Range, vector3Range, vector2IntRange, vector3IntRange, rgbRange, hsvRange"

type VariantDimensionInput struct {
	Path string `json:"path" jsonschema:"variable path this dimension controls; must match an existing variable's path"`
	Type string `json:"type" jsonschema:"one of: discrete, numericRange, intRange, vector2Range, vector3Range, vector2IntRange, vector3IntRange, rgbRange, hsvRange"`
	Data string `json:"data" jsonschema:"literal JSON matching the chosen type's shape - discrete: {\"values\":[<json literal>, ...]} e.g. {\"values\":[1,2,3]} or {\"values\":[\"red\",\"blue\"]}; numericRange/intRange: {\"min\":N,\"max\":N,\"samples\":N}; vector2Range/vector2IntRange: {\"min\":{\"x\":N,\"y\":N},\"max\":{\"x\":N,\"y\":N},\"samples\":N}; vector3Range/vector3IntRange: same but with \"z\" too; rgbRange: {\"min\":\"#rrggbb\",\"max\":\"#rrggbb\",\"samples\":N}; hsvRange: {\"min\":{\"h\":N,\"s\":N,\"v\":N},\"max\":{\"h\":N,\"s\":N,\"v\":N},\"samples\":N} where h is degrees and s/v are 0-1. Samples below 1 behaves as 1 (just Min)."`
}

// buildDimension turns one tool-call dimension spec into a variant.Dimension
// by re-using variant.UnmarshalDimension against a synthesized envelope -
// the same {"type":...,"data":...} shape variant sets are persisted in, so
// what this tool accepts is exactly what save_graph/load_graph round-trip.
func buildDimension(in VariantDimensionInput) (variant.Dimension, error) {
	if in.Path == "" {
		return nil, fmt.Errorf("path is required")
	}
	if !json.Valid([]byte(in.Data)) {
		return nil, fmt.Errorf("data is not valid JSON: %s", in.Data)
	}

	envelope, err := json.Marshal(struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}{Type: in.Type, Data: json.RawMessage(in.Data)})
	if err != nil {
		return nil, err
	}

	return variant.UnmarshalDimension(in.Path, envelope)
}

const (
	variantModeCross      = "cross"
	variantModeOneAtATime = "oneAtATime"
)

type CreateVariantSetInput struct {
	Name       string                  `json:"name" jsonschema:"unique name for this variant set, e.g. 'ColorSweep'; replaces any existing variant set with the same name"`
	Dimensions []VariantDimensionInput `json:"dimensions" jsonschema:"one entry per variable path this set varies. Keep it to the 3-5 controls that genuinely vary independently."`
	AllowLarge bool                    `json:"allowLarge,omitempty" jsonschema:"required for a set of more than 500 combinations; without it the call refuses and reports the count"`
}

type CreateVariantSetOutput struct {
	Name              string `json:"name"`
	TotalCombinations int    `json:"totalCombinations" jsonschema:"how many variants a full cross sweep of this set would produce"`
	OneAtATime        int    `json:"oneAtATime" jsonschema:"how many variants a oneAtATime sweep of this set would produce"`
	Warning           string `json:"warning,omitempty"`
}

const (
	variantSetMaxCombinations = 500
	variantSweepWarnAbove     = 1000
)

func (s *Server) createVariantSet(ctx context.Context, req *mcpsdk.CallToolRequest, in CreateVariantSetInput) (*mcpsdk.CallToolResult, CreateVariantSetOutput, error) {
	var out CreateVariantSetOutput
	var err error
	s.atomic(&err, func() error {
		if len(in.Dimensions) == 0 {
			return fmt.Errorf("variant set %q has no dimensions", in.Name)
		}

		set := variant.Set{}
		for i, d := range in.Dimensions {
			dim, e := buildDimension(d)
			if e != nil {
				return fmt.Errorf("dimension %d (%q): %w", i, d.Path, e)
			}
			set.Dimensions = append(set.Dimensions, dim)
		}

		total := set.TotalCombinations()
		if total > variantSetMaxCombinations && !in.AllowLarge {
			return fmt.Errorf("%q would be %d combinations, over the %d limit. Dimensions multiply: keep a set to the 3-5 controls that genuinely vary independently, sweep it with mode oneAtATime for a per-control robustness check, or pass allowLarge:true if every combination is really wanted", in.Name, total, variantSetMaxCombinations)
		}

		s.graph.VariantSets().Set(in.Name, set)

		out.Name = in.Name
		out.TotalCombinations = total
		out.OneAtATime = oneAtATimeCount(set)
		if total > variantSweepWarnAbove {
			out.Warning = fmt.Sprintf("a full sweep is %d outputs; run_variant_sweep with limit and sample to check a handful before committing to all of them", total)
		}
		return nil
	})
	return nil, out, err
}

type VariantDimensionSummary struct {
	Path  string `json:"path"`
	Count int    `json:"count" jsonschema:"how many distinct values this dimension contributes"`
}

type VariantSetSummary struct {
	Name              string                    `json:"name"`
	Dimensions        []VariantDimensionSummary `json:"dimensions,omitempty"`
	TotalCombinations int                       `json:"totalCombinations" jsonschema:"variants in a cross sweep"`
	OneAtATime        int                       `json:"oneAtATime" jsonschema:"variants in a oneAtATime sweep"`
}

func oneAtATimeCount(set variant.Set) int {
	total := 0
	for _, d := range set.Dimensions {
		total += d.Count()
	}
	return total
}

// Dimensions[0] is the least significant digit, matching variant.Set.Sweep's order.
func crossCombination(set variant.Set, index int) (variable.Profile, error) {
	profile := make(variable.Profile, len(set.Dimensions))
	for _, d := range set.Dimensions {
		count := d.Count()
		value, err := d.Value(index % count)
		if err != nil {
			return nil, fmt.Errorf("dimension %q: %w", d.Path(), err)
		}
		profile[d.Path()] = value
		index /= count
	}
	return profile, nil
}

func oneAtATimeCombination(set variant.Set, index int) (variable.Profile, error) {
	for _, d := range set.Dimensions {
		if index >= d.Count() {
			index -= d.Count()
			continue
		}
		value, err := d.Value(index)
		if err != nil {
			return nil, fmt.Errorf("dimension %q: %w", d.Path(), err)
		}
		return variable.Profile{d.Path(): value}, nil
	}
	return nil, fmt.Errorf("variant %d is past the end of the set", index)
}

type ListVariantSetsInput struct{}

type ListVariantSetsOutput struct {
	VariantSets []VariantSetSummary `json:"variantSets"`
}

func (s *Server) listVariantSets(ctx context.Context, req *mcpsdk.CallToolRequest, in ListVariantSetsInput) (*mcpsdk.CallToolResult, ListVariantSetsOutput, error) {
	var out ListVariantSetsOutput
	var err error
	s.atomic(&err, func() error {
		for _, name := range s.graph.VariantSets().Names() {
			set, e := s.graph.VariantSets().Get(name)
			if e != nil {
				return e
			}

			summary := VariantSetSummary{
				Name:              name,
				TotalCombinations: set.TotalCombinations(),
				OneAtATime:        oneAtATimeCount(set),
			}
			for _, d := range set.Dimensions {
				summary.Dimensions = append(summary.Dimensions, VariantDimensionSummary{
					Path:  d.Path(),
					Count: d.Count(),
				})
			}
			out.VariantSets = append(out.VariantSets, summary)
		}
		return nil
	})
	return nil, out, err
}

type RenameVariantSetInput struct {
	Name    string `json:"name"`
	NewName string `json:"newName"`
}

type RenameVariantSetOutput struct {
	Name string `json:"name"`
}

func (s *Server) renameVariantSet(ctx context.Context, req *mcpsdk.CallToolRequest, in RenameVariantSetInput) (*mcpsdk.CallToolResult, RenameVariantSetOutput, error) {
	var out RenameVariantSetOutput
	var err error
	s.atomic(&err, func() error {
		if e := s.graph.VariantSets().Rename(in.Name, in.NewName); e != nil {
			return e
		}
		out.Name = in.NewName
		return nil
	})
	return nil, out, err
}

type DeleteVariantSetInput struct {
	Name string `json:"name"`
}

type DeleteVariantSetOutput struct {
	Deleted bool `json:"deleted"`
}

func (s *Server) deleteVariantSet(ctx context.Context, req *mcpsdk.CallToolRequest, in DeleteVariantSetInput) (*mcpsdk.CallToolResult, DeleteVariantSetOutput, error) {
	var out DeleteVariantSetOutput
	var err error
	s.atomic(&err, func() error {
		if e := s.graph.VariantSets().Delete(in.Name); e != nil {
			return e
		}
		out.Deleted = true
		return nil
	})
	return nil, out, err
}

type RunVariantSweepInput struct {
	Name          string `json:"name" jsonschema:"name of an existing variant set to sweep"`
	Mode          string `json:"mode,omitempty" jsonschema:"cross (default): every combination of the dimensions, their counts multiplied. oneAtATime: one variant per dimension value with every other variable at its live value, their counts added - the robustness check for each control's extremes."`
	OutputDir     string `json:"outputDir" jsonschema:"folder to write one numbered subfolder per variant into, e.g. 'polyform-output/sweep', plus variants.json"`
	Limit         int    `json:"limit,omitempty" jsonschema:"run at most this many variants; omitted or 0 runs the whole set"`
	Sample        string `json:"sample,omitempty" jsonschema:"which variants limit keeps: first (default), spaced (evenly through the set, both ends included) or random (seeded)"`
	Seed          int64  `json:"seed,omitempty" jsonschema:"seed for sample random; omitted picks one and reports it"`
	KeepLast      bool   `json:"keepLast,omitempty" jsonschema:"leave the live variables at the last variant's values instead of putting back the values they had before the sweep"`
	Confirm       bool   `json:"confirm,omitempty" jsonschema:"must be true to proceed when more variants would run than warnThreshold (default 1000); omitted/false is refused with the actual count so the size can be reviewed before writing that many folders"`
	WarnThreshold int    `json:"warnThreshold,omitempty" jsonschema:"variant count above which the sweep refuses to run without confirm:true; defaults to 1000 when omitted or zero"`
}

type VariantResult struct {
	Index       int               `json:"index" jsonschema:"position in this run; the folder is variant-<index>"`
	Combination int               `json:"combination" jsonschema:"index into the full enumeration of the set in this sweep's mode"`
	Folder      string            `json:"folder"`
	Variables   map[string]string `json:"variables" jsonschema:"variable path -> the JSON value this variant ran with"`
	Triangles   int               `json:"triangles" jsonschema:"triangles across every manifest exported, instances included"`
	EvalMs      int64             `json:"evalMs"`
	NodeErrors  []NodeError       `json:"nodeErrors,omitempty"`
}

type RunVariantSweepOutput struct {
	Mode              string          `json:"mode"`
	TotalCombinations int             `json:"totalCombinations" jsonschema:"size of the whole set in this mode, whether or not limit kept all of it"`
	Ran               int             `json:"ran"`
	Seed              int64           `json:"seed,omitempty" jsonschema:"the seed sample random used; pass it back to rerun the same pick"`
	Folders           []string        `json:"folders" jsonschema:"one subfolder path per variant written, in run order"`
	Variants          []VariantResult `json:"variants" jsonschema:"one row per variant run: values, triangles, eval time and node errors - the review table, also written to variants.json"`
	Manifest          string          `json:"manifest" jsonschema:"path of variants.json"`
	Restored          bool            `json:"restored" jsonschema:"the variables the sweep changed are back at their pre-sweep values"`
}

func (s *Server) runVariantSweep(ctx context.Context, req *mcpsdk.CallToolRequest, in RunVariantSweepInput) (*mcpsdk.CallToolResult, RunVariantSweepOutput, error) {
	var out RunVariantSweepOutput
	var err error
	s.atomic(&err, func() error {
		set, e := s.graph.VariantSets().Get(in.Name)
		if e != nil {
			return e
		}
		if in.OutputDir == "" {
			return fmt.Errorf("outputDir is required")
		}

		mode := in.Mode
		var total int
		var combination func(variant.Set, int) (variable.Profile, error)
		switch mode {
		case "", variantModeCross:
			mode, total, combination = variantModeCross, set.TotalCombinations(), crossCombination
		case variantModeOneAtATime:
			total, combination = oneAtATimeCount(set), oneAtATimeCombination
		default:
			return fmt.Errorf("mode %q is not %s or %s", mode, variantModeCross, variantModeOneAtATime)
		}

		indices, seed, e := sweepIndices(total, in.Limit, in.Sample, in.Seed)
		if e != nil {
			return e
		}

		warnThreshold := in.WarnThreshold
		if warnThreshold <= 0 {
			warnThreshold = generator.DefaultSweepWarnThreshold
		}
		if len(indices) > warnThreshold && !in.Confirm {
			return fmt.Errorf("sweep %q would run %d variants, over the %d warning threshold - pass limit (with sample spaced or random) to check a few, or confirm:true to run them all", in.Name, len(indices), warnThreshold)
		}

		combinations := make([]variable.Profile, len(indices))
		touched := map[string]bool{}
		for i, index := range indices {
			if combinations[i], e = combination(set, index); e != nil {
				return e
			}
			for path := range combinations[i] {
				touched[path] = true
			}
		}
		before := make(map[string][]byte, len(touched))
		for path := range touched {
			if before[path], e = s.graph.VariableData(path); e != nil {
				return fmt.Errorf("variant set %q varies %q, which is not a variable: %w", in.Name, path, e)
			}
		}

		if e := os.MkdirAll(in.OutputDir, os.ModePerm); e != nil {
			return e
		}
		out.Mode = mode
		out.TotalCombinations = total
		out.Seed = seed
		out.Folders = []string{}
		out.Variants = []VariantResult{}
		for i, profile := range combinations {
			result, e := s.runVariant(in.OutputDir, i, indices[i], profile)
			if e != nil {
				return e
			}
			out.Variants = append(out.Variants, result)
			out.Folders = append(out.Folders, result.Folder)
		}
		out.Ran = len(out.Variants)

		out.Manifest = filepath.Join(in.OutputDir, "variants.json")
		table, e := json.MarshalIndent(struct {
			Set               string          `json:"set"`
			Mode              string          `json:"mode"`
			TotalCombinations int             `json:"totalCombinations"`
			Variants          []VariantResult `json:"variants"`
		}{Set: in.Name, Mode: mode, TotalCombinations: total, Variants: out.Variants}, "", "\t")
		if e != nil {
			return e
		}
		if e := os.WriteFile(out.Manifest, table, 0o644); e != nil {
			return e
		}

		if in.KeepLast {
			return nil
		}
		for _, path := range slices.Sorted(maps.Keys(before)) {
			if _, e := s.graph.UpdateVariable(path, before[path]); e != nil {
				return fmt.Errorf("restoring %q after the sweep: %w", path, e)
			}
		}
		out.Restored = true
		return nil
	})
	return nil, out, err
}

func (s *Server) runVariant(outputDir string, index, combination int, profile variable.Profile) (VariantResult, error) {
	start := time.Now()
	if e := s.graph.ApplyProfile(profile); e != nil {
		return VariantResult{}, fmt.Errorf("applying variant %d: %w", index, e)
	}

	folder := filepath.Join(outputDir, fmt.Sprintf("variant-%04d", index))
	if e := os.MkdirAll(folder, os.ModePerm); e != nil {
		return VariantResult{}, e
	}
	e := graph.WriteToFolder(s.graph, folder)
	nodeErrors := collectNodeErrors(s.graph.Graph)
	if e != nil {
		return VariantResult{}, withNodeErrors(fmt.Errorf("writing variant %d: %w", index, e), nodeErrors)
	}

	saved, e := json.MarshalIndent(profile, "", "\t")
	if e != nil {
		return VariantResult{}, e
	}
	if e := os.WriteFile(filepath.Join(folder, "profile.json"), saved, 0o644); e != nil {
		return VariantResult{}, e
	}

	return VariantResult{
		Index:       index,
		Combination: combination,
		Folder:      folder,
		Variables:   profileValues(profile),
		Triangles:   s.exportedTriangles(),
		EvalMs:      time.Since(start).Milliseconds(),
		NodeErrors:  nodeErrors,
	}, nil
}

func (s *Server) exportedTriangles() int {
	total := 0
	_ = graph.ForeachManifestNodeOutput(s.graph, func(_ string, node nodes.Node, _ nodes.Output[manifest.Manifest]) error {
		models, ok := node.Inputs()["Models"].(nodes.ArrayValueInputPort)
		if !ok {
			return nil
		}
		for _, port := range models.Value() {
			if model, ok := port.(nodes.Output[*gltf.PolyformModel]); ok {
				total += modelTriangles(model.Value())
			}
		}
		return nil
	})
	return total
}

func modelTriangles(model *gltf.PolyformModel) int {
	if model == nil {
		return 0
	}
	total := 0
	if model.Mesh != nil && model.Mesh.HasFloat3Attribute(modeling.PositionAttribute) {
		total = model.Mesh.PrimitiveCount()
		if instances := len(model.GpuInstances); instances > 0 {
			total *= instances
		}
	}
	for _, child := range model.Children {
		total += modelTriangles(child)
	}
	return total
}

// sweepIndices picks which combinations of a set of size total a run covers.
func sweepIndices(total, limit int, sample string, seed int64) ([]int, int64, error) {
	if limit <= 0 || limit > total {
		limit = total
	}
	indices := make([]int, limit)
	switch sample {
	case "", "first":
		for i := range indices {
			indices[i] = i
		}
		return indices, 0, nil

	case "spaced":
		for i := range indices {
			if limit > 1 {
				indices[i] = i * (total - 1) / (limit - 1)
			}
		}
		return indices, 0, nil

	case "random":
		if seed == 0 {
			seed = time.Now().UnixNano()
		}
		rng := rand.New(rand.NewSource(seed))
		if total <= 1<<20 {
			copy(indices, rng.Perm(total))
		} else {
			chosen := make(map[int]bool, limit)
			for len(chosen) < limit {
				chosen[rng.Intn(total)] = true
			}
			indices = slices.Sorted(maps.Keys(chosen))
		}
		slices.Sort(indices)
		return indices, seed, nil
	}
	return nil, 0, fmt.Errorf("sample must be first, spaced or random, not %q", sample)
}

func (s *Server) registerVariantSetTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "create_variant_set",
		Description: "Create (or replace) a named variant set, saved with the graph: one dimension per variable it varies (types: " + variantDimensionTypeKeys + " - see 'data' for each shape). Refuses more than 500 combinations without allowLarge, because dimensions multiply - a set is the 3-5 controls that genuinely vary independently, not every variable. This tool only defines the set; run_variant_sweep executes it, as every combination or one control at a time. A few named looks are profiles, rendered as one render_preview with a view per profile.",
	}, s.createVariantSet)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "list_variant_sets",
		Description: "List every variant set currently defined on the graph, with each dimension's variable path and value count, and how many variants a cross and a oneAtATime sweep of it would run.",
	}, s.listVariantSets)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "rename_variant_set",
		Description: "Rename a variant set.",
	}, s.renameVariantSet)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "delete_variant_set",
		Description: "Delete a variant set.",
	}, s.deleteVariantSet)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "run_variant_sweep",
		Description: "Run a saved variant set against the live graph, writing each variant's full export to its own numbered subfolder under outputDir, and return one row per variant - its values, triangle count, eval time and node errors - also written as variants.json next to the folders. mode cross (default) runs every combination; mode oneAtATime varies one control at a time from the live values. Pass limit with sample spaced or random to check a few combinations of a big set without defining a second one. The variables the sweep changed are put back afterward unless keepLast is true. Refuses to run past warnThreshold variants (default 1000) without confirm, since each one is a full export.",
	}, s.runVariantSweep)
}
