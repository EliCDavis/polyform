package mcp

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/EliCDavis/polyform/generator/schema"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type SaveProfileInput struct {
	Name      string            `json:"name" jsonschema:"the profile's name, e.g. \"Barn\" or \"French Doors\"; saving to an existing name replaces it"`
	Variables map[string]string `json:"variables,omitempty" jsonschema:"variable path -> JSON value text to apply on top of the current values for this snapshot only; the live values are restored afterward. One look-sheet entry becomes one profile without update_variable calls."`
}

type SaveProfileOutput struct {
	Name     string   `json:"name"`
	Profiles []string `json:"profiles" jsonschema:"every profile the graph now holds"`
}

func (s *Server) saveProfile(ctx context.Context, req *mcpsdk.CallToolRequest, in SaveProfileInput) (*mcpsdk.CallToolResult, SaveProfileOutput, error) {
	var out SaveProfileOutput
	var err error
	s.atomic(&err, func() error {
		name := strings.TrimSpace(in.Name)
		if name == "" {
			return fmt.Errorf("name is required")
		}
		restore, e := s.overrideVariables(in.Variables)
		if e != nil {
			return e
		}
		s.graph.SaveProfile(name)
		restore()
		out.Name = name
		out.Profiles = s.graph.Profiles()
		return nil
	})
	return nil, out, err
}

type ListProfilesInput struct{}

type ListProfilesOutput struct {
	Profiles []string `json:"profiles"`
}

func (s *Server) listProfiles(ctx context.Context, req *mcpsdk.CallToolRequest, in ListProfilesInput) (*mcpsdk.CallToolResult, ListProfilesOutput, error) {
	var out ListProfilesOutput
	var err error
	s.atomic(&err, func() error {
		out.Profiles = s.graph.Profiles()
		return nil
	})
	return nil, out, err
}

type ProfileNameInput struct {
	Name string `json:"name"`
}

type ApplyProfileOutput struct {
	Name  string   `json:"name"`
	Unset []string `json:"unset,omitempty" jsonschema:"variables this profile does not cover, so they keep whatever value they had. A profile only holds the variables that existed when it was saved, so these are usually variables added afterward - re-save the profile to fold them in."`
	Note  string   `json:"note,omitempty"`
}

func (s *Server) applyProfile(ctx context.Context, req *mcpsdk.CallToolRequest, in ProfileNameInput) (*mcpsdk.CallToolResult, ApplyProfileOutput, error) {
	var out ApplyProfileOutput
	var err error
	s.atomic(&err, func() error {
		before := s.graph.Schema().Variables
		if e := s.graph.LoadProfile(in.Name); e != nil {
			return fmt.Errorf("%w; it has %s", e, strings.Join(s.graph.Profiles(), ", "))
		}
		out.Name = in.Name

		// LoadProfile writes only the paths the profile holds, so anything
		// added since it was saved silently keeps the previous look's
		// value. Naming them beats a catalog that drifts.
		covered := map[string]bool{}
		for _, path := range s.graph.ProfileVariables(in.Name) {
			covered[path] = true
		}
		before.Traverse(func(path string, _ schema.Variable) bool {
			if !covered[path] {
				out.Unset = append(out.Unset, path)
			}
			return true
		})
		sort.Strings(out.Unset)
		if len(out.Unset) > 0 {
			out.Note = fmt.Sprintf("%d variable(s) are not in this profile and kept their current values; save_profile over it to fold them in", len(out.Unset))
		}
		return nil
	})
	return nil, out, err
}

type DeleteProfileOutput struct {
	Profiles []string `json:"profiles" jsonschema:"every profile the graph still holds"`
}

func (s *Server) deleteProfile(ctx context.Context, req *mcpsdk.CallToolRequest, in ProfileNameInput) (*mcpsdk.CallToolResult, DeleteProfileOutput, error) {
	var out DeleteProfileOutput
	var err error
	s.atomic(&err, func() error {
		if e := s.graph.DeleteProfile(in.Name); e != nil {
			return e
		}
		out.Profiles = s.graph.Profiles()
		return nil
	})
	return nil, out, err
}

func (s *Server) registerProfileTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "save_profile",
		Description: "Snapshot every variable's current value under a name. Profiles are saved with the graph (save_graph) and switchable in the polyform editor, so the named looks of a catalog - \"Barn\", \"French Doors\", \"Gothic\" - travel with the file. Pass variables to snapshot a look without changing the live values, exactly as render_preview's override does.",
	}, s.saveProfile)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "list_profiles",
		Description: "List the graph's saved variable profiles by name.",
	}, s.listProfiles)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "apply_profile",
		Description: "Set every variable the profile holds to its saved value. This changes the live graph; save_profile the current state first if you need to come back to it. A profile only covers the variables that existed when it was saved, so the result lists any it leaves alone - re-save old profiles after adding variables, or a catalog drifts as the graph grows.",
	}, s.applyProfile)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "delete_profile",
		Description: "Remove a saved profile. Variables keep their current values.",
	}, s.deleteProfile)
}
