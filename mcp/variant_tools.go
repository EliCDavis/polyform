package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/variant"
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

type CreateVariantSetInput struct {
	Name       string                  `json:"name" jsonschema:"unique name for this variant set, e.g. 'ColorSweep'"`
	Dimensions []VariantDimensionInput `json:"dimensions" jsonschema:"one entry per variable path this variant set controls; replaces any existing variant set with the same name"`
}

type CreateVariantSetOutput struct {
	Name              string `json:"name"`
	TotalCombinations int    `json:"totalCombinations" jsonschema:"how many variants a full sweep of this set would produce (product of every dimension's sample/value count)"`
}

func (s *Server) createVariantSet(ctx context.Context, req *mcpsdk.CallToolRequest, in CreateVariantSetInput) (*mcpsdk.CallToolResult, CreateVariantSetOutput, error) {
	var out CreateVariantSetOutput
	var err error
	s.atomic(&err, func() error {
		dimensions := make([]variant.Dimension, 0, len(in.Dimensions))
		for i, d := range in.Dimensions {
			dim, e := buildDimension(d)
			if e != nil {
				return fmt.Errorf("dimension %d (%q): %w", i, d.Path, e)
			}
			dimensions = append(dimensions, dim)
		}

		set := variant.Set{Dimensions: dimensions}
		if e := s.graph.SetVariantSet(in.Name, set); e != nil {
			return e
		}

		out.Name = in.Name
		out.TotalCombinations = set.TotalCombinations()
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
	Dimensions        []VariantDimensionSummary `json:"dimensions"`
	TotalCombinations int                       `json:"totalCombinations"`
}

type ListVariantSetsInput struct{}

type ListVariantSetsOutput struct {
	VariantSets []VariantSetSummary `json:"variantSets"`
}

func (s *Server) listVariantSets(ctx context.Context, req *mcpsdk.CallToolRequest, in ListVariantSetsInput) (*mcpsdk.CallToolResult, ListVariantSetsOutput, error) {
	var out ListVariantSetsOutput
	var err error
	s.atomic(&err, func() error {
		for _, name := range s.graph.VariantSets() {
			set, e := s.graph.VariantSet(name)
			if e != nil {
				return e
			}

			summary := VariantSetSummary{
				Name:              name,
				TotalCombinations: set.TotalCombinations(),
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
		if e := s.graph.RenameVariantSet(in.Name, in.NewName); e != nil {
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
		if e := s.graph.DeleteVariantSet(in.Name); e != nil {
			return e
		}
		out.Deleted = true
		return nil
	})
	return nil, out, err
}

type RunVariantSweepInput struct {
	Name          string `json:"name" jsonschema:"name of an existing variant set to sweep"`
	OutputDir     string `json:"outputDir" jsonschema:"folder to write one numbered subfolder per variant into, e.g. 'polyform-output/sweep'"`
	Confirm       bool   `json:"confirm,omitempty" jsonschema:"must be true to proceed when the sweep would produce more variants than warnThreshold (default 1000); omitted/false is refused with the actual combination count so the size can be reviewed before writing that many folders"`
	WarnThreshold int    `json:"warnThreshold,omitempty" jsonschema:"combination count above which the sweep refuses to run without confirm:true; defaults to 1000 when omitted or zero"`
}

type RunVariantSweepOutput struct {
	TotalCombinations int      `json:"totalCombinations"`
	Folders           []string `json:"folders" jsonschema:"one subfolder path per variant written, in sweep order"`
}

// runVariantSweep applies each of the variant set's combinations to the live
// graph in turn and writes it to its own subfolder, via the same
// generator.RunVariants the CLI's Sweep command uses. This leaves the
// graph's variables set to the LAST combination swept, not restored to
// whatever they were before the call - callers relying on the graph's
// values afterward (e.g. a subsequent render_preview) should account for
// that, same as the CLI.
func (s *Server) runVariantSweep(ctx context.Context, req *mcpsdk.CallToolRequest, in RunVariantSweepInput) (*mcpsdk.CallToolResult, RunVariantSweepOutput, error) {
	var out RunVariantSweepOutput
	var err error
	s.atomic(&err, func() error {
		set, e := s.graph.VariantSet(in.Name)
		if e != nil {
			return e
		}

		warnThreshold := in.WarnThreshold
		if warnThreshold <= 0 {
			warnThreshold = generator.DefaultSweepWarnThreshold
		}

		total := set.TotalCombinations()
		if total > warnThreshold && !in.Confirm {
			return fmt.Errorf("sweep %q would produce %d variants, over the %d warning threshold - pass confirm:true to proceed anyway, or narrow the variant set's ranges", in.Name, total, warnThreshold)
		}

		profiles, e := set.Sweep()
		if e != nil {
			return e
		}

		if e := generator.RunVariants(s.graph, profiles, in.OutputDir); e != nil {
			return e
		}

		out.TotalCombinations = total
		out.Folders = make([]string, len(profiles))
		for i := range profiles {
			out.Folders[i] = filepath.Join(in.OutputDir, fmt.Sprintf("variant-%04d", i))
		}
		return nil
	})
	return nil, out, err
}

func (s *Server) registerVariantSetTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "create_variant_set",
		Description: "Create (or replace) a named variant set: a collection of per-variable value ranges used to sweep every combination or sample random combinations of variables, producing many outputs from one graph. Each dimension targets one existing variable's path. Types: " + variantDimensionTypeKeys + " - see each dimension's 'data' field description for its exact JSON shape. This tool only defines the set; run_variant_sweep executes it (a full sweep is a deliberate, potentially large action - don't call run_variant_sweep unless the user actually asks to run one), and the 'polyform sample' CLI command covers random sampling.",
	}, s.createVariantSet)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "list_variant_sets",
		Description: "List every variant set currently defined on the graph, with each dimension's variable path, value count, and the set's total combination count if fully swept.",
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
		Description: "Run every combination of a saved variant set against the live graph, writing each variant to its own numbered subfolder under outputDir (same behavior as the 'polyform sweep' CLI command). Refuses to run past warnThreshold combinations (default 1000) unless confirm is true, since a sweep writes one full set of output files per combination. Leaves the graph's variables set to the last combination swept, not restored afterward - re-apply a profile or update_variable calls if you need the graph back to a specific state after sweeping.",
	}, s.runVariantSweep)
}
