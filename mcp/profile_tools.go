package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/variable"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type SaveProfileInput struct {
	Name      string            `json:"name" jsonschema:"the profile's name, e.g. \"Default\" or \"Tall\"; saving to an existing name replaces it"`
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
		if e := restore(); e != nil {
			return e
		}
		out.Name = name
		out.Profiles = s.graph.Profiles().Names()
		return nil
	})
	return nil, out, err
}

func (s *Server) liveVariablePaths() map[string]bool {
	live := map[string]bool{}
	s.graph.Schema().Variables.Traverse(func(path string, _ schema.Variable) bool {
		live[path] = true
		return true
	})
	return live
}

// profileCoverage splits a profile against the live variables: missing are
// variables it does not pin, stale are keys for variables that no longer exist.
func profileCoverage(profile variable.Profile, live map[string]bool) (missing, stale []string) {
	for path := range live {
		if _, ok := profile[path]; !ok {
			missing = append(missing, path)
		}
	}
	for path := range profile {
		if !live[path] {
			stale = append(stale, path)
		}
	}
	sort.Strings(missing)
	sort.Strings(stale)
	return missing, stale
}

func profileValues(profile variable.Profile) map[string]string {
	values := make(map[string]string, len(profile))
	for path, value := range profile {
		values[path] = string(value)
	}
	return values
}

type ProfileSummary struct {
	Name    string   `json:"name"`
	Pinned  int      `json:"pinned" jsonschema:"how many live variables this profile sets"`
	Missing []string `json:"missing,omitempty" jsonschema:"live variables this profile does not set, usually ones added after it was saved; applying it leaves them at whatever the last look put there. update_profile to fold them in."`
	Stale   []string `json:"stale,omitempty" jsonschema:"keys for variables that no longer exist; ignored on apply"`
}

type ListProfilesInput struct{}

type ListProfilesOutput struct {
	Profiles []ProfileSummary `json:"profiles"`
	Note     string           `json:"note,omitempty"`
}

func (s *Server) listProfiles(ctx context.Context, req *mcpsdk.CallToolRequest, in ListProfilesInput) (*mcpsdk.CallToolResult, ListProfilesOutput, error) {
	var out ListProfilesOutput
	var err error
	s.atomic(&err, func() error {
		live := s.liveVariablePaths()
		incomplete := 0
		out.Profiles = []ProfileSummary{}
		for _, name := range s.graph.Profiles().Names() {
			profile, _ := s.graph.Profiles().Get(name)
			missing, stale := profileCoverage(profile, live)
			if len(missing) > 0 {
				incomplete++
			}
			out.Profiles = append(out.Profiles, ProfileSummary{
				Name:    name,
				Pinned:  len(profile) - len(stale),
				Missing: missing,
				Stale:   stale,
			})
		}
		if incomplete > 0 {
			out.Note = fmt.Sprintf("%d profile(s) do not set every live variable; update_profile each with the values it should pin before the final render, or those variables keep the previous look's values", incomplete)
		}
		return nil
	})
	return nil, out, err
}

type ProfileNameInput struct {
	Name string `json:"name"`
}

type GetProfileOutput struct {
	Name      string            `json:"name"`
	Variables map[string]string `json:"variables" jsonschema:"variable path -> saved JSON value text"`
	Missing   []string          `json:"missing,omitempty" jsonschema:"live variables this profile does not set"`
	Stale     []string          `json:"stale,omitempty" jsonschema:"keys for variables that no longer exist"`
}

func (s *Server) getProfile(ctx context.Context, req *mcpsdk.CallToolRequest, in ProfileNameInput) (*mcpsdk.CallToolResult, GetProfileOutput, error) {
	var out GetProfileOutput
	var err error
	s.atomic(&err, func() error {
		profile, e := s.graph.Profiles().Get(in.Name)
		if e != nil {
			return fmt.Errorf("%w; it has %s", e, strings.Join(s.graph.Profiles().Names(), ", "))
		}
		out.Name = in.Name
		out.Variables = profileValues(profile)
		out.Missing, out.Stale = profileCoverage(profile, s.liveVariablePaths())
		return nil
	})
	return nil, out, err
}

type ApplyProfileOutput struct {
	Name    string   `json:"name"`
	Unset   []string `json:"unset,omitempty" jsonschema:"variables this profile does not cover, so they keep whatever value they had. A profile only holds the variables that existed when it was saved, so these are usually variables added afterward - update_profile to fold them in."`
	Skipped []string `json:"skipped,omitempty" jsonschema:"keys for variables that no longer exist, left out of the apply"`
	Note    string   `json:"note,omitempty"`
}

func (s *Server) applyProfile(ctx context.Context, req *mcpsdk.CallToolRequest, in ProfileNameInput) (*mcpsdk.CallToolResult, ApplyProfileOutput, error) {
	var out ApplyProfileOutput
	var err error
	s.atomic(&err, func() error {
		profile, e := s.graph.Profiles().Get(in.Name)
		if e != nil {
			return fmt.Errorf("%w; it has %s", e, strings.Join(s.graph.Profiles().Names(), ", "))
		}
		out.Unset, out.Skipped = profileCoverage(profile, s.liveVariablePaths())
		if e := s.graph.LoadProfile(in.Name); e != nil {
			return e
		}
		out.Name = in.Name

		var notes []string
		if len(out.Unset) > 0 {
			notes = append(notes, fmt.Sprintf("%d variable(s) are not in this profile and kept their current values; update_profile to fold them in", len(out.Unset)))
		}
		if len(out.Skipped) > 0 {
			notes = append(notes, fmt.Sprintf("%d key(s) name variables that no longer exist and were skipped", len(out.Skipped)))
		}
		out.Note = strings.Join(notes, "; ")
		return nil
	})
	return nil, out, err
}

type UpdateProfileInput struct {
	Name      string            `json:"name" jsonschema:"an existing profile"`
	Variables map[string]string `json:"variables" jsonschema:"variable path -> JSON value text to set in the profile; keys it already holds are replaced, the rest are kept. The live values are not changed."`
}

type UpdateProfileOutput struct {
	Name      string            `json:"name"`
	Variables map[string]string `json:"variables" jsonschema:"the profile's full contents after the merge"`
	Missing   []string          `json:"missing,omitempty" jsonschema:"live variables the profile still does not set"`
}

func (s *Server) updateProfile(ctx context.Context, req *mcpsdk.CallToolRequest, in UpdateProfileInput) (*mcpsdk.CallToolResult, UpdateProfileOutput, error) {
	var out UpdateProfileOutput
	var err error
	s.atomic(&err, func() error {
		profile, e := s.graph.Profiles().Get(in.Name)
		if e != nil {
			return fmt.Errorf("%w; it has %s", e, strings.Join(s.graph.Profiles().Names(), ", "))
		}
		if len(in.Variables) == 0 {
			return fmt.Errorf("variables is empty")
		}
		canonical, e := s.canonicalValues(in.Variables)
		if e != nil {
			return e
		}
		merged := make(variable.Profile, len(profile)+len(canonical))
		for path, value := range profile {
			merged[path] = value
		}
		for path, value := range canonical {
			merged[path] = value
		}
		s.graph.Profiles().Set(in.Name, merged)
		out.Name = in.Name
		out.Variables = profileValues(merged)
		out.Missing, _ = profileCoverage(merged, s.liveVariablePaths())
		return nil
	})
	return nil, out, err
}

// canonicalValues checks each value against its variable's type by applying
// it, reads back the variable's own encoding, and puts the live value back.
func (s *Server) canonicalValues(values map[string]string) (variable.Profile, error) {
	canonical := make(variable.Profile, len(values))
	paths := make([]string, 0, len(values))
	for path := range values {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		before, e := s.graph.VariableData(path)
		if e != nil {
			return nil, fmt.Errorf("variables[%q]: %w", path, e)
		}
		if _, e := applyJSONText(values[path], func(data []byte) (bool, error) {
			return s.graph.UpdateVariable(path, data)
		}); e != nil {
			return nil, fmt.Errorf("variables[%q]: %w", path, e)
		}
		after, e := s.graph.VariableData(path)
		if e != nil {
			return nil, e
		}
		canonical[path] = json.RawMessage(after)
		if _, e := s.graph.UpdateVariable(path, before); e != nil {
			return nil, fmt.Errorf("restoring variables[%q]: %w", path, e)
		}
	}
	return canonical, nil
}

type DeleteProfileOutput struct {
	Profiles []string `json:"profiles" jsonschema:"every profile the graph still holds"`
}

func (s *Server) deleteProfile(ctx context.Context, req *mcpsdk.CallToolRequest, in ProfileNameInput) (*mcpsdk.CallToolResult, DeleteProfileOutput, error) {
	var out DeleteProfileOutput
	var err error
	s.atomic(&err, func() error {
		if e := s.graph.Profiles().Delete(in.Name); e != nil {
			return e
		}
		out.Profiles = s.graph.Profiles().Names()
		return nil
	})
	return nil, out, err
}

func (s *Server) registerProfileTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "save_profile",
		Description: "Snapshot every variable's current value under a name. Profiles are saved with the graph (save_graph) and switchable in the polyform editor, so the named looks of a catalog - \"Default\", \"Tall\", \"Compact\" - travel with the file. Pass variables to snapshot a look without changing the live values: the same map you give one render_preview view, so a look sheet of N views becomes N save_profile calls with the same N maps. A profile is a snapshot - a variable added or re-defaulted later is not in it; list_profiles shows which, update_profile patches one key without re-saving.",
	}, s.saveProfile)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "list_profiles",
		Description: "List the graph's saved variable profiles with their coverage: the live variables each one does not set (added since it was saved, so applying it leaves them at the previous look's values) and any keys for variables that no longer exist. Check this before the final render; a profile missing a variable is a look that will drift.",
	}, s.listProfiles)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "get_profile",
		Description: "Read a profile's saved values without applying them, plus the live variables it does not set. The way to see what a look holds, or to copy one profile's values into another.",
	}, s.getProfile)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "apply_profile",
		Description: "Set every variable the profile holds to its saved value. This changes the live graph; save_profile the current state first if you need to come back to it. Variables the profile does not cover keep their values and are listed as unset; keys for variables that no longer exist are skipped and listed.",
	}, s.applyProfile)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "update_profile",
		Description: "Merge values into an existing profile without touching the live variables: keys given are set, the rest of the profile is kept. Use it to fold a newly added variable into every saved look, or to change one value of one look, instead of apply_profile + update_variable + save_profile.",
	}, s.updateProfile)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "delete_profile",
		Description: "Remove a saved profile. Variables keep their current values.",
	}, s.deleteProfile)
}
