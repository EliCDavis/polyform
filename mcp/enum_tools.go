package mcp

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/graph"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func decodeJSON(data json.RawMessage) any {
	var value any
	if len(data) == 0 || json.Unmarshal(data, &value) != nil {
		return nil
	}
	return value
}

func (s *Server) enumMembers(typeName string) []string {
	if table, ok := s.graph.EnumTable(typeName); ok {
		return table.Names()
	}
	if def, ok := enum.Lookup(typeName); ok {
		return def.Names()
	}
	return nil
}

type EnumMemberSpec struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

func enumMembersOf(specs []EnumMemberSpec) []enum.Member {
	members := make([]enum.Member, len(specs))
	for i, spec := range specs {
		members[i] = enum.Member{Name: spec.Name, Description: spec.Description}
	}
	return members
}

type EnumSummary struct {
	Id                string   `json:"id,omitempty" jsonschema:"set for an enum this graph declares; a Go enum has none and cannot be edited"`
	Type              string   `json:"type" jsonschema:"the type key: create_variable's 'type', a subgraph boundary's type"`
	Name              string   `json:"name,omitempty"`
	Description       string   `json:"description,omitempty"`
	Members           []string `json:"members"`
	ParameterNodeType string   `json:"parameterNodeType,omitempty" jsonschema:"create_node type: a node holding one member, for when one choice feeds several inputs"`
	MembersNodeType   string   `json:"membersNodeType,omitempty" jsonschema:"create_node type: a node listing every member, to run a For Each over them or index into them"`
}

func graphEnumSummary(id, name, description string, members []string) EnumSummary {
	return EnumSummary{
		Id:                id,
		Type:              enum.RuntimeTypePath(id),
		Name:              name,
		Description:       description,
		Members:           members,
		ParameterNodeType: graph.EnumParameterTypePath(id),
		MembersNodeType:   graph.EnumMembersTypePath(id),
	}
}

type DefineEnumInput struct {
	Name        string           `json:"name" jsonschema:"display name; the id is derived from it"`
	Description string           `json:"description,omitempty"`
	Members     []EnumMemberSpec `json:"members" jsonschema:"at least one; names must differ"`
}

func (s *Server) defineEnum(ctx context.Context, req *mcpsdk.CallToolRequest, in DefineEnumInput) (*mcpsdk.CallToolResult, EnumSummary, error) {
	var out EnumSummary
	var err error
	s.atomic(&err, func() error {
		members := enumMembersOf(in.Members)
		id, e := s.graph.CreateEnum(in.Name, in.Description, members)
		if e != nil {
			return e
		}
		names := make([]string, len(members))
		for i, member := range members {
			names[i] = member.Name
		}
		out = graphEnumSummary(id, in.Name, in.Description, names)
		return nil
	})
	return nil, out, err
}

type EditEnumInput struct {
	Id          string            `json:"id"`
	Name        string            `json:"name,omitempty" jsonschema:"if set, replaces the display name"`
	Description string            `json:"description,omitempty" jsonschema:"if set, replaces the description"`
	Members     []EnumMemberSpec  `json:"members,omitempty" jsonschema:"the full new member list, in order; omit to keep the members. A member dropped while something still holds it is refused, naming the holders."`
	Renamed     map[string]string `json:"renamed,omitempty" jsonschema:"old member name -> new, for members that changed name rather than went away: everything holding one follows, saved profiles included"`
}

func (s *Server) editEnum(ctx context.Context, req *mcpsdk.CallToolRequest, in EditEnumInput) (*mcpsdk.CallToolResult, EnumSummary, error) {
	var out EnumSummary
	var err error
	s.atomic(&err, func() error {
		current, ok := s.graph.Enums()[in.Id]
		if !ok {
			return fmt.Errorf("no enum %q", in.Id)
		}
		update := graph.EnumUpdate{
			Name:        cmp.Or(in.Name, current.Name),
			Description: cmp.Or(in.Description, current.Description),
			Members:     current.Members,
			Renamed:     in.Renamed,
		}
		if len(in.Members) > 0 {
			update.Members = enumMembersOf(in.Members)
		}
		if e := s.graph.UpdateEnum(in.Id, update); e != nil {
			return e
		}
		out = graphEnumSummary(in.Id, update.Name, update.Description, s.enumMembers(enum.RuntimeTypePath(in.Id)))
		return nil
	})
	return nil, out, err
}

type DeleteEnumInput struct {
	Id string `json:"id"`
}

type DeleteEnumOutput struct {
	Deleted bool `json:"deleted"`
}

func (s *Server) deleteEnum(ctx context.Context, req *mcpsdk.CallToolRequest, in DeleteEnumInput) (*mcpsdk.CallToolResult, DeleteEnumOutput, error) {
	var out DeleteEnumOutput
	var err error
	s.atomic(&err, func() error {
		if e := s.graph.DeleteEnum(in.Id); e != nil {
			return e
		}
		out.Deleted = true
		return nil
	})
	return nil, out, err
}

type ListEnumsInput struct{}

type ListEnumsOutput struct {
	Enums []EnumSummary `json:"enums"`
}

func (s *Server) listEnums(ctx context.Context, req *mcpsdk.CallToolRequest, in ListEnumsInput) (*mcpsdk.CallToolResult, ListEnumsOutput, error) {
	var out ListEnumsOutput
	var err error
	s.atomic(&err, func() error {
		out.Enums = []EnumSummary{}
		declared := s.graph.Enums()
		for _, id := range slices.Sorted(maps.Keys(declared)) {
			def := declared[id]
			out.Enums = append(out.Enums, graphEnumSummary(id, def.Name, def.Description, s.enumMembers(enum.RuntimeTypePath(id))))
		}
		for _, def := range enum.Definitions() {
			out.Enums = append(out.Enums, EnumSummary{Type: def.TypeName(), Members: def.Names()})
		}
		return nil
	})
	return nil, out, err
}

func (s *Server) registerEnumTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "define_enum",
		Description: "Declare a named set of choices on this graph. Its type key works as a variable type (create_variable) and a subgraph boundary type, and an input of that type takes a member name as its value. Use one instead of a string compared by hand or an int that means a choice.",
	}, s.defineEnum)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "edit_enum",
		Description: "Change an enum this graph declared: its name, description or members. 'renamed' carries holders across a member's new name.",
	}, s.editEnum)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "delete_enum",
		Description: "Delete an enum this graph declared. Refused while anything still uses its type.",
	}, s.deleteEnum)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "list_enums",
		Description: "List every enum: the ones this graph declares, then the ones built into node types, each with its members.",
	}, s.listEnums)
}
