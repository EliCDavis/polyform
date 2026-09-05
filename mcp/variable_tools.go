package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// variableTypeKeys documents (and NewTypedVariable enforces) the set of
// variable types this server knows how to construct. It mirrors the
// VariableFactory wired up in cmd/polyform-mcp/main.go, which is what lets
// load_graph reconstruct variables from an existing graph file, so the two
// must stay in sync.
const variableTypeKeys = "float64, string, int, bool, vector2.vector[float64], vector2.vector[int], vector3.vector[float64], vector3.vector[int], []vector3.vector[float64], geometry.aabb, coloring.color, image.image, file"

// NewTypedVariable constructs an empty variable.Variable for the given type
// key. It's exported so cmd/polyform-mcp/main.go can also use it as the
// graph.Config.VariableFactory, keeping variable creation (this file) and
// variable deserialization (on load_graph) backed by the same type table.
func NewTypedVariable(typeKey string) (variable.Variable, error) {
	switch strings.ToLower(typeKey) {
	case "float64":
		return &variable.TypeVariable[float64]{}, nil
	case "string":
		return &variable.TypeVariable[string]{}, nil
	case "int":
		return &variable.TypeVariable[int]{}, nil
	case "bool":
		return &variable.TypeVariable[bool]{}, nil
	case "vector2.vector[float64]":
		return &variable.TypeVariable[vector2.Float64]{}, nil
	case "vector2.vector[int]":
		return &variable.TypeVariable[vector2.Int]{}, nil
	case "vector3.vector[float64]":
		return &variable.TypeVariable[vector3.Float64]{}, nil
	case "vector3.vector[int]":
		return &variable.TypeVariable[vector3.Int]{}, nil
	case "[]vector3.vector[float64]":
		return &variable.TypeVariable[[]vector3.Float64]{}, nil
	case "geometry.aabb":
		return &variable.TypeVariable[geometry.AABB]{}, nil
	case "coloring.color":
		return &variable.TypeVariable[coloring.Color]{}, nil
	case "image.image":
		return &variable.ImageVariable{}, nil
	case "file":
		return &variable.FileVariable{}, nil
	default:
		return nil, fmt.Errorf("unsupported variable type %q, expected one of: %s", typeKey, variableTypeKeys)
	}
}

type CreateVariableInput struct {
	Path        string `json:"path" jsonschema:"unique variable path, e.g. 'Radius'; this is also the type key passed to create_node to place a reference to this variable in the graph"`
	Type        string `json:"type" jsonschema:"one of: float64, string, int, bool, vector2.vector[float64], vector2.vector[int], vector3.vector[float64], vector3.vector[int], []vector3.vector[float64], geometry.aabb, coloring.color, image.image, file"`
	Description string `json:"description,omitempty" jsonschema:"shown to whoever edits this variable's value later; explain what it controls"`
	Value       string `json:"value,omitempty" jsonschema:"literal JSON text for the initial value, e.g. 5, \"red\", {\"x\":1,\"y\":2,\"z\":3}; omit to use the type's zero value"`
}

type CreateVariableOutput struct {
	Path string `json:"path"`
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
		if !json.Valid([]byte(in.Value)) {
			return fmt.Errorf("value is not valid JSON: %s", in.Value)
		}
		if _, e := v.ApplyMessage([]byte(in.Value)); e != nil {
			return fmt.Errorf("value doesn't match type %q: %w", in.Type, e)
		}
	}

	s.graph.NewVariable(in.Path, v)

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
	Path  string `json:"path"`
	Value string `json:"value" jsonschema:"literal JSON text matching the variable's type, e.g. 5, \"red\", {\"x\":1,\"y\":2,\"z\":3}"`
}

type UpdateVariableOutput struct {
	Updated bool `json:"updated"`
}

func (s *Server) updateVariable(ctx context.Context, req *mcpsdk.CallToolRequest, in UpdateVariableInput) (*mcpsdk.CallToolResult, UpdateVariableOutput, error) {
	var out UpdateVariableOutput
	var err error
	s.atomic(&err, func() error {
		data := []byte(in.Value)
		if !json.Valid(data) {
			return fmt.Errorf("value is not valid JSON: %s", in.Value)
		}
		ok, e := s.graph.UpdateVariable(in.Path, data)
		if e != nil {
			return e
		}
		out.Updated = ok
		return nil
	})
	return nil, out, err
}

type DeleteVariableInput struct {
	Path string `json:"path"`
}

type DeleteVariableOutput struct {
	Deleted bool `json:"deleted"`
}

func (s *Server) deleteVariable(ctx context.Context, req *mcpsdk.CallToolRequest, in DeleteVariableInput) (*mcpsdk.CallToolResult, DeleteVariableOutput, error) {
	var out DeleteVariableOutput
	var err error
	s.atomic(&err, func() error {
		s.graph.DeleteVariable(in.Path)
		out.Deleted = true
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
	Path string `json:"path"`
}

func (s *Server) renameVariable(ctx context.Context, req *mcpsdk.CallToolRequest, in RenameVariableInput) (*mcpsdk.CallToolResult, RenameVariableOutput, error) {
	var out RenameVariableOutput
	var err error
	s.atomic(&err, func() error {
		v := s.graph.GetVariable(in.Path)
		description := v.Info().Description()
		if in.Description != "" {
			description = in.Description
		}
		if e := s.graph.SetVariableInfo(in.Path, in.NewPath, description); e != nil {
			return e
		}
		out.Path = in.NewPath
		return nil
	})
	return nil, out, err
}

type ListVariablesInput struct{}

type ListVariablesOutput struct {
	Variables []VariableSummary `json:"variables"`
}

func (s *Server) listVariables(ctx context.Context, req *mcpsdk.CallToolRequest, in ListVariablesInput) (*mcpsdk.CallToolResult, ListVariablesOutput, error) {
	var out ListVariablesOutput
	var err error
	s.atomic(&err, func() error {
		s.graph.Schema().Variables.Traverse(func(path string, v schema.Variable) bool {
			out.Variables = append(out.Variables, VariableSummary{
				Path:        path,
				Description: v.Description,
				Type:        v.Type,
				Value:       v.Value,
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
		Description: "Create a high-level, user-facing variable (e.g. \"Radius\", \"Wheel Count\") that lives outside the node graph proper. Once created, pass its path as create_node/instantiate_subgraph's 'inputs' {\"variable\": \"<path>\"} (or as create_node's 'type', for a standalone reference node) to reference it anywhere in the graph — the same variable can back many nodes, and editing it later updates every place it's referenced without touching graph structure.",
	}, s.createVariable)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "create_variables",
		Description: "Create several variables in one call — same effect as calling create_variable once per entry, just fewer round trips. Stops at the first invalid entry; variables before it are still created.",
	}, s.createVariables)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "update_variable",
		Description: "Set a variable's current value. Every node referencing this variable picks up the new value immediately.",
	}, s.updateVariable)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "delete_variable",
		Description: "Delete a variable and any nodes in the graph that reference it.",
	}, s.deleteVariable)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "rename_variable",
		Description: "Move a variable to a new path and/or update its description.",
	}, s.renameVariable)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "list_variables",
		Description: "List every variable currently defined on the graph, with its path, type, description, and current value.",
	}, s.listVariables)
}
