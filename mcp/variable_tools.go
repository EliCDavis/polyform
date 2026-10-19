package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/generator/variable/variabletypes"
	"github.com/EliCDavis/polyform/generator/variant"
	"github.com/EliCDavis/polyform/nodes"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewTypedVariable is the same table the editor uses, so anything a graph
// file from either app can hold, load_graph can reconstruct.
func NewTypedVariable(typeKey string) (variable.Variable, error) {
	return variabletypes.New(typeKey)
}

type CreateVariableInput struct {
	Path        string `json:"path" jsonschema:"unique variable path, e.g. 'Radius'; this is also the type key passed to create_node to place a reference to this variable in the graph"`
	Type        string `json:"type" jsonschema:"one of: float64, int, string, bool, vector2.Vector[float64], vector2.Vector[int], vector3.Vector[float64], vector3.Vector[int], vector4.Vector[float64], []float64, []int, []string, []vector2.Vector[float64], []vector2.Vector[int], []vector3.Vector[float64], []vector3.Vector[int], geometry.AABB, quaternion.Quaternion, trs.TRS, []trs.TRS, coloring.Color, []coloring.Color, coloring.Gradient[github.com/EliCDavis/polyform/drawing/coloring.Color], image.Image, file"`
	Description string `json:"description,omitempty" jsonschema:"shown to whoever edits this variable's value later; explain what it controls"`
	Value       string `json:"value,omitempty" jsonschema:"the initial value as JSON text inside a string, never a bare JSON number/object: \"5\", \"\\\"red\\\"\", \"{\\\"x\\\":1,\\\"y\\\":2,\\\"z\\\":3}\"; omit to use the type's zero value"`
}

type CreateVariableOutput struct {
	Path string `json:"path"`
}

// A JSON string wrapping a number or object is unwrapped and retried: the
// schema calls the value "JSON text in a string" and it gets read that way.
func applyJSONText(value string, apply func([]byte) (bool, error)) (bool, error) {
	data := []byte(value)
	if !json.Valid(data) {
		return false, fmt.Errorf("value is not valid JSON: %s", value)
	}
	ok, err := apply(data)
	if err == nil {
		return ok, nil
	}
	var text string
	if json.Unmarshal(data, &text) == nil && json.Valid([]byte(text)) {
		if ok, retry := apply([]byte(text)); retry == nil {
			return ok, nil
		}
	}
	return false, fmt.Errorf("%s: %w", value, err)
}

// createOneVariable constructs and registers a single variable on
// s.graph. It's the shared core behind both create_variable and
// create_variables, so a batch of N variables behaves identically to N
// individual create_variable calls (same validation, same error text).
// Callers must already hold s.mu (i.e. call this from within s.atomic).
func (s *Server) createOneVariable(in CreateVariableInput) error {
	v, e := NewTypedVariable(in.Type)
	if e != nil {
		return e
	}

	if in.Value != "" {
		if _, e := applyJSONText(in.Value, v.ApplyMessage); e != nil {
			return fmt.Errorf("value doesn't match type %q: %w", in.Type, e)
		}
	}

	if _, e := s.graph.NewVariable(in.Path, v); e != nil {
		return e
	}

	if in.Description != "" {
		if e := s.graph.SetVariableDescription(in.Path, in.Description); e != nil {
			return e
		}
	}
	return nil
}

func (s *Server) createVariable(ctx context.Context, req *mcpsdk.CallToolRequest, in CreateVariableInput) (*mcpsdk.CallToolResult, CreateVariableOutput, error) {
	var out CreateVariableOutput
	var err error
	s.atomic(&err, func() error {
		if e := s.createOneVariable(in); e != nil {
			return e
		}
		out.Path = in.Path
		return nil
	})
	return nil, out, err
}

type CreateVariablesInput struct {
	Variables []CreateVariableInput `json:"variables" jsonschema:"one entry per variable to create, same shape as create_variable's arguments (path, type, description, value)"`
}

type CreateVariablesOutput struct {
	Paths []string `json:"paths" jsonschema:"the created variable paths, in the same order as the input list"`
}

func (s *Server) createVariables(ctx context.Context, req *mcpsdk.CallToolRequest, in CreateVariablesInput) (*mcpsdk.CallToolResult, CreateVariablesOutput, error) {
	var out CreateVariablesOutput
	var err error
	s.atomic(&err, func() error {
		for i, v := range in.Variables {
			if e := s.createOneVariable(v); e != nil {
				return fmt.Errorf("variable %d (%q): %w — variables before this one were already created", i, v.Path, e)
			}
			out.Paths = append(out.Paths, v.Path)
		}
		return nil
	})
	return nil, out, err
}

type UpdateVariableInput struct {
	Path        string `json:"path"`
	Value       string `json:"value,omitempty" jsonschema:"the new value as JSON text inside a string, never a bare JSON number/object: \"5\", \"\\\"red\\\"\", \"{\\\"x\\\":1,\\\"y\\\":2,\\\"z\\\":3}\". Omit to change only the description."`
	Description string `json:"description,omitempty" jsonschema:"if set, replaces the variable's description; references and the value are untouched"`
	Reset       bool   `json:"reset,omitempty" jsonschema:"set the variable back to its creation default. The graph does not record that default, so this currently reports that instead of changing anything; pass value to set a default you know."`
}

type UpdateVariableOutput struct {
	Updated bool `json:"updated"`
}

func (s *Server) updateOneVariable(in UpdateVariableInput) (bool, error) {
	if in.Reset {
		return false, fmt.Errorf("variable %q has no recorded default: the graph stores only a variable's current value, so reset has nothing to restore - pass value instead", in.Path)
	}
	if in.Value == "" && in.Description == "" {
		return false, fmt.Errorf("pass a value, a description, or both")
	}
	updated := false
	if in.Description != "" {
		if e := s.graph.SetVariableDescription(in.Path, in.Description); e != nil {
			return false, e
		}
		updated = true
	}
	if in.Value == "" {
		return updated, nil
	}
	ok, e := applyJSONText(in.Value, func(data []byte) (bool, error) {
		return s.graph.UpdateVariable(in.Path, data)
	})
	if e != nil {
		return false, e
	}
	return updated || ok, nil
}

func (s *Server) updateVariable(ctx context.Context, req *mcpsdk.CallToolRequest, in UpdateVariableInput) (*mcpsdk.CallToolResult, UpdateVariableOutput, error) {
	var out UpdateVariableOutput
	var err error
	s.atomic(&err, func() error {
		updated, e := s.updateOneVariable(in)
		out.Updated = updated
		return e
	})
	return nil, out, err
}

type UpdateVariablesInput struct {
	Variables []UpdateVariableInput `json:"variables" jsonschema:"one entry per variable, same shape as update_variable's arguments (path, value, description, reset)"`
}

type UpdateVariableStatus struct {
	Path    string `json:"path"`
	Updated bool   `json:"updated"`
	Error   string `json:"error,omitempty" jsonschema:"why this entry was not applied; the other entries still were"`
}

type UpdateVariablesOutput struct {
	Results []UpdateVariableStatus `json:"results" jsonschema:"one per input entry, in order"`
	Failed  int                    `json:"failed,omitempty"`
}

func (s *Server) updateVariables(ctx context.Context, req *mcpsdk.CallToolRequest, in UpdateVariablesInput) (*mcpsdk.CallToolResult, UpdateVariablesOutput, error) {
	var out UpdateVariablesOutput
	var err error
	s.atomic(&err, func() error {
		if len(in.Variables) == 0 {
			return fmt.Errorf("variables is empty")
		}
		for _, entry := range in.Variables {
			status := UpdateVariableStatus{Path: entry.Path}
			updated, e := s.updateOneVariable(entry)
			status.Updated = updated
			if e != nil {
				status.Error = e.Error()
				out.Failed++
			}
			out.Results = append(out.Results, status)
		}
		if out.Failed == len(in.Variables) {
			return fmt.Errorf("no entry could be applied: %s", out.Results[0].Error)
		}
		return nil
	})
	return nil, out, err
}

type DeleteVariableInput struct {
	Path string `json:"path"`
}

type DeleteVariableOutput struct {
	Deleted            bool     `json:"deleted"`
	References         []string `json:"references,omitempty" jsonschema:"reference nodes removed, as 'nodeId' (root) or 'subgraphId/nodeId'"`
	Unfed              []string `json:"unfed,omitempty" jsonschema:"inputs those references fed and that now read nothing, as 'nodeId.Port' or 'subgraphId/nodeId.Port'; wire or set_parameter each one, or the node computes from a zero value"`
	Profiles           []string `json:"profiles,omitempty" jsonschema:"profiles the variable's key was removed from"`
	VariantSets        []string `json:"variantSets,omitempty" jsonschema:"variant sets the variable's dimension or row values were removed from"`
	VariantSetsDeleted []string `json:"variantSetsDeleted,omitempty" jsonschema:"variant sets removed because this variable was all they varied"`
}

func (s *Server) deleteVariable(ctx context.Context, req *mcpsdk.CallToolRequest, in DeleteVariableInput) (*mcpsdk.CallToolResult, DeleteVariableOutput, error) {
	var out DeleteVariableOutput
	var err error
	s.atomic(&err, func() error {
		report, e := s.graph.RemoveVariable(in.Path)
		if e != nil {
			return e
		}
		out = DeleteVariableOutput{
			Deleted:            true,
			References:         report.References,
			Unfed:              report.Unfed,
			Profiles:           report.Profiles,
			VariantSets:        report.VariantSets,
			VariantSetsDeleted: report.VariantSetsDeleted,
		}
		return nil
	})
	return nil, out, err
}

type RenameVariableInput struct {
	Path        string `json:"path"`
	NewPath     string `json:"newPath"`
	Description string `json:"description,omitempty" jsonschema:"if set, replaces the variable's description; if omitted the existing description is kept"`
}

type RenameVariableOutput struct {
	Path        string   `json:"path"`
	Profiles    []string `json:"profiles,omitempty" jsonschema:"profiles whose key for this variable was rewritten to the new path"`
	VariantSets []string `json:"variantSets,omitempty" jsonschema:"variant sets whose dimension for this variable now uses the new path"`
}

func (s *Server) renameVariable(ctx context.Context, req *mcpsdk.CallToolRequest, in RenameVariableInput) (*mcpsdk.CallToolResult, RenameVariableOutput, error) {
	var out RenameVariableOutput
	var err error
	s.atomic(&err, func() error {
		v, e := s.graph.GetVariable(in.Path)
		if e != nil {
			return e
		}
		description := v.Info().Description()
		if in.Description != "" {
			description = in.Description
		}
		for _, name := range s.graph.Profiles().Names() {
			if profile, _ := s.graph.Profiles().Get(name); profile[in.Path] != nil {
				out.Profiles = append(out.Profiles, name)
			}
		}
		for _, name := range s.graph.VariantSets().Names() {
			if set, _ := s.graph.VariantSets().Get(name); setVariesVariable(set, in.Path) {
				out.VariantSets = append(out.VariantSets, name)
			}
		}
		if e := s.graph.SetVariableInfo(in.Path, in.NewPath, description); e != nil {
			return e
		}
		out.Path = in.NewPath
		if in.NewPath == in.Path {
			out.Profiles, out.VariantSets = nil, nil
		}
		return nil
	})
	return nil, out, err
}

func setVariesVariable(set variant.Set, path string) bool {
	for _, d := range set.Dimensions {
		if d.Path() == path {
			return true
		}
	}
	return false
}

// variableReferences maps each variable path to the input ports its
// reference nodes feed, across the root graph and every subgraph
// definition.
func (s *Server) variableReferences() map[string][]string {
	byVariable := map[variable.Variable]string{}
	s.graph.Schema().Variables.Traverse(func(path string, _ schema.Variable) bool {
		if v, err := s.graph.GetVariable(path); err == nil {
			byVariable[v] = path
		}
		return true
	})

	refs := map[string][]string{}
	collect := func(inst *graph.Graph, prefix string) {
		for _, consumerID := range inst.NodeIds() {
			for _, port := range inst.Node(consumerID).Inputs() {
				portName := portLabel(port)
				var upstreams []nodes.OutputPort
				switch p := port.(type) {
				case nodes.SingleValueInputPort:
					upstreams = append(upstreams, p.Value())
				case nodes.ArrayValueInputPort:
					upstreams = append(upstreams, p.Value()...)
				}
				for _, up := range upstreams {
					if up == nil {
						continue
					}
					ref, ok := up.Node().(variable.Reference)
					if !ok {
						continue
					}
					if path, known := byVariable[ref.Reference()]; known {
						refs[path] = append(refs[path], prefix+consumerID+"."+portName)
					}
				}
			}
		}
	}
	collect(s.graph.Graph, "")
	for id := range s.graph.Schema().SubGraphs {
		if child, e := s.graph.SubGraphInstance(id); e == nil {
			collect(child, id+"/")
		}
	}
	for path := range refs {
		sort.Strings(refs[path])
	}
	return refs
}

type ListVariablesInput struct{}

type ListVariablesOutput struct {
	Variables []VariableSummary `json:"variables"`
}

func (s *Server) listVariables(ctx context.Context, req *mcpsdk.CallToolRequest, in ListVariablesInput) (*mcpsdk.CallToolResult, ListVariablesOutput, error) {
	var out ListVariablesOutput
	var err error
	s.atomic(&err, func() error {
		refs := s.variableReferences()
		s.graph.Schema().Variables.Traverse(func(path string, v schema.Variable) bool {
			out.Variables = append(out.Variables, VariableSummary{
				Path:        path,
				Description: v.Description,
				Type:        v.Type,
				Value:       v.Value,
				References:  refs[path],
			})
			return true
		})
		sort.Slice(out.Variables, func(i, j int) bool { return out.Variables[i].Path < out.Variables[j].Path })
		return nil
	})
	return nil, out, err
}

func (s *Server) registerVariableTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "create_variable",
		Description: "Create a high-level, user-facing variable (e.g. \"Radius\", \"Segment Count\") that lives outside the node graph proper. Once created, pass its path as create_node/instantiate_subgraph's 'inputs' {\"variable\": \"<path>\"} (or as create_node's 'type', for a standalone reference node) to reference it anywhere in the graph — the same variable can back many nodes, and editing it later updates every place it's referenced without touching graph structure.",
	}, s.createVariable)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "create_variables",
		Description: "Create several variables in one call — same effect as calling create_variable once per entry, just fewer round trips. An invalid entry fails the whole call and none of them are created, so a reported error never leaves half a batch behind.",
	}, s.createVariables)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "update_variable",
		Description: "Set a variable's current value and/or description. Every node referencing this variable picks up the new value immediately. The value is JSON text; a number or object wrapped in a JSON string is unwrapped. Saved profiles are snapshots and do not follow this change - update_profile the ones that should.",
	}, s.updateVariable)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "update_variables",
		Description: "Set several variables in one call, as one undoable edit. Each entry is applied on its own and reports its own status, so one bad value does not block the rest; the call errors only when nothing could be applied.",
	}, s.updateVariables)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "delete_variable",
		Description: "Delete a variable, every node referencing it in the root graph and inside every subgraph definition, and its key in every saved profile and variant set. The result names what went: the reference nodes, the inputs they fed that now read nothing (wire or set_parameter each, or that node computes from a zero value), and the profiles and variant sets touched.",
	}, s.deleteVariable)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "rename_variable",
		Description: "Move a variable to a new path and/or update its description. Reference nodes anywhere in the graph, including nested subgraphs, keep reading it; saved profiles and variant sets are rewritten to the new path and the result names them.",
	}, s.renameVariable)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "list_variables",
		Description: "List every variable currently defined on the graph, with its path, type, description, and current value.",
	}, s.listVariables)
}
