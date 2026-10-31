package graph

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/EliCDavis/polyform/formats/swagger"
	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/named"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/generator/variant"
	"github.com/EliCDavis/polyform/nodes"
)

// A graph's own enums are resolved here; every other type comes from the
// app's factory.
func (a *Instance) buildVariable(typeKey string) (variable.Variable, error) {
	if id := enum.RuntimeTypeID(typeKey); id != "" {
		def, ok := a.enums[id]
		if !ok {
			return nil, fmt.Errorf("this graph has no enum %q", id)
		}
		return variable.NewEnumVariable(def.table), nil
	}
	if a.variableFactory == nil {
		return nil, fmt.Errorf("no variable factory to build a %q variable with", typeKey)
	}
	return a.variableFactory(typeKey)
}

// NewVariableOfType returns the variable and the node type its reference
// nodes are created with.
func (a *Instance) NewVariableOfType(variablePath, typeKey string) (created variable.Variable, nodeType string, err error) {
	_, err = a.write(refuseConflicts, func() (err error) {
		created, err = a.buildVariable(typeKey)
		if err != nil {
			return err
		}
		nodeType, err = a.newVariable(variablePath, created)
		return err
	})
	return created, nodeType, err
}

// NewVariable returns the node type its reference nodes are created with.
func (a *Instance) NewVariable(variablePath string, variable variable.Variable) (nodeType string, err error) {
	_, err = a.write(refuseConflicts, func() (err error) {
		nodeType, err = a.newVariable(variablePath, variable)
		return err
	})
	return nodeType, err
}

func (a *Instance) newVariable(variablePath string, variable variable.Variable) (string, error) {
	if variable == nil {
		return "", fmt.Errorf("trying to add a nil variable %q to graph", variablePath)
	}
	if err := a.variables.Add(variablePath, variable); err != nil {
		return "", fmt.Errorf("failed to add variable to graph: %w", err)
	}

	a.typeFactory.RegisterBuilder(variablePath, func() any {
		return variable.NodeReference()
	})
	nodes.DiscoverNodePortTypes(variable.NodeReference())

	return variablePath, nil
}

// VariableDeletion is everything RemoveVariable took out along with the
// variable. Node entries are "nodeId" in the root graph or "subgraphId/nodeId".
type VariableDeletion struct {
	References         []string
	Unfed              []string
	Profiles           []string
	VariantSets        []string
	VariantSetsDeleted []string
}

func (a *Instance) DeleteVariable(variablePath string) error {
	_, err := a.RemoveVariable(variablePath)
	return err
}

// Also removes every reference node and the variable's key in profiles and variant sets.
func (a *Instance) RemoveVariable(variablePath string) (report VariableDeletion, err error) {
	_, err = a.write(dropConflicts, func() error {
		return a.removeVariable(variablePath, &report)
	})
	return report, err
}

func (a *Instance) removeVariable(variablePath string, report *VariableDeletion) error {
	deleted, err := a.variables.Variable(variablePath)
	if err != nil {
		return fmt.Errorf("trying to delete a variable at the path %q which doesn't contain a variable", variablePath)
	}
	if err := a.removeReferences(deleted, report); err != nil {
		return err
	}

	for _, name := range a.profiles.Names() {
		profile, _ := a.profiles.Get(name)
		if _, ok := profile[variablePath]; !ok {
			continue
		}
		delete(profile, variablePath)
		a.profiles.Set(name, profile)
		report.Profiles = append(report.Profiles, name)
	}

	for _, name := range a.variantSets.Names() {
		set, _ := a.variantSets.Get(name)
		stripped, changed := setWithoutVariable(set, variablePath)
		if !changed {
			continue
		}
		if len(stripped.Dimensions) == 0 {
			_ = a.variantSets.Delete(name)
			report.VariantSetsDeleted = append(report.VariantSetsDeleted, name)
			continue
		}
		a.variantSets.Set(name, stripped)
		report.VariantSets = append(report.VariantSets, name)
	}

	if err := a.variables.Remove(variablePath); err != nil {
		return err
	}
	a.typeFactory.Unregister(variablePath)
	return nil
}

func (a *Instance) removeReferences(deleted variable.Variable, report *VariableDeletion) error {
	for _, scope := range a.scopes() {
		for _, id := range slices.Sorted(maps.Keys(scope.graph.nodesByID)) {
			node := scope.graph.nodesByID[id]
			ref, ok := node.(variable.Reference)
			if !ok || ref.Reference() != deleted {
				continue
			}
			report.References = append(report.References, scope.prefix+id)
			report.Unfed = append(report.Unfed, scope.graph.readersOf(node, scope.prefix)...)
			if err := scope.graph.deleteNode(node); err != nil {
				return err
			}
		}
	}
	return nil
}

type scopedGraph struct {
	prefix string
	graph  *Graph
}

// Root first, then every definition by id.
func (a *Instance) scopes() []scopedGraph {
	scopes := []scopedGraph{{graph: a.Graph}}
	for _, id := range slices.Sorted(maps.Keys(a.subGraphs)) {
		scopes = append(scopes, scopedGraph{prefix: id + "/", graph: a.subGraphs[id].instance})
	}
	return scopes
}

func (a *Graph) readersOf(producer nodes.Node, prefix string) []string {
	var readers []string
	for consumer, inputs := range a.reads {
		for input, sources := range inputs {
			if slices.ContainsFunc(sources, func(from source) bool { return from.node == producer }) {
				readers = append(readers, prefix+a.nodeIDs[consumer]+"."+input)
			}
		}
	}
	slices.Sort(readers)
	return readers
}

func setWithoutVariable(set variant.Set, variablePath string) (variant.Set, bool) {
	dimensions := slices.DeleteFunc(slices.Clone(set.Dimensions), func(d variant.Dimension) bool {
		return d.Path() == variablePath
	})
	return variant.Set{Dimensions: dimensions}, len(dimensions) != len(set.Dimensions)
}

func (a *Instance) GetVariable(variablePath string) (variable.Variable, error) {
	variable, err := a.variables.Variable(variablePath)
	if err != nil {
		return nil, fmt.Errorf("trying to get a variable at the path %q which doesn't exist", variablePath)
	}
	return variable, nil
}

func (a *Instance) SetVariableInfo(variablePath, newPath, description string) error {
	if err := a.SetVariableDescription(variablePath, description); err != nil {
		return err
	}
	_, err := a.write(refuseConflicts, func() error {
		clean := strings.TrimSpace(newPath)
		if err := a.variables.Move(variablePath, clean); err != nil {
			return err
		}
		if clean == variablePath {
			return nil
		}
		moved, err := a.variables.Variable(clean)
		if err != nil {
			return err
		}
		a.typeFactory.Unregister(variablePath)
		a.typeFactory.RegisterBuilder(clean, func() any {
			return moved.NodeReference()
		})
		a.renameInProfiles(variablePath, clean)
		return a.renameInVariantSets(variablePath, clean)
	})
	return err
}

func (a *Instance) renameInProfiles(oldPath, newPath string) {
	for _, name := range a.profiles.Names() {
		profile, _ := a.profiles.Get(name)
		value, ok := profile[oldPath]
		if !ok {
			continue
		}
		delete(profile, oldPath)
		profile[newPath] = value
		a.profiles.Set(name, profile)
	}
}

func (a *Instance) renameInVariantSets(oldPath, newPath string) error {
	for _, name := range a.variantSets.Names() {
		set, _ := a.variantSets.Get(name)
		changed := false
		dimensions := slices.Clone(set.Dimensions)
		for i, d := range dimensions {
			if d.Path() != oldPath {
				continue
			}
			raw, err := json.Marshal(d)
			if err != nil {
				return fmt.Errorf("variant set %q: %w", name, err)
			}
			if dimensions[i], err = variant.UnmarshalDimension(newPath, raw); err != nil {
				return fmt.Errorf("variant set %q: %w", name, err)
			}
			changed = true
		}
		if changed {
			a.variantSets.Set(name, variant.Set{Dimensions: dimensions})
		}
	}
	return nil
}

func (a *Instance) variablePath(v variable.Variable) string {
	path := ""
	a.variables.Traverse(func(candidatePath string, _ variable.Info, candidate variable.Variable) {
		if candidate == v {
			path = candidatePath
		}
	})
	return path
}

func (a *Instance) SetVariableDescription(variablePath, description string) error {
	a.mu().Lock()
	defer a.mu().Unlock()
	variable, err := a.variables.Variable(variablePath)
	if err != nil {
		return err
	}
	variable.Info().SetDescription(description)
	return nil
}

func (a *Instance) UpdateVariable(variablePath string, data []byte) (changed bool, err error) {
	_, err = a.write(refuseConflicts, func() error {
		variable, err := a.variables.Variable(variablePath)
		if err != nil {
			return err
		}
		changed, err = variable.ApplyMessage(data)
		if err != nil {
			return fmt.Errorf("%s: %w", variablePath, err)
		}
		return nil
	})
	return changed, err
}

func (a *Instance) VariableData(variablePath string) ([]byte, error) {
	a.mu().Lock()
	defer a.mu().Unlock()

	variable, err := a.variables.Variable(variablePath)
	if err != nil {
		return nil, err
	}
	return variable.ToMessage(), nil
}

func (a *Instance) SwaggerDefinition() swagger.Definition {
	a.mu().RLock()
	defer a.mu().RUnlock()
	return a.variables.SwaggerDefinition()
}

func (a *Instance) Profiles() *named.Collection[variable.Profile] {
	return a.profiles
}

func (a *Instance) SaveProfile(profileName string) {
	a.mu().Lock()
	defer a.mu().Unlock()
	a.Profiles().Set(profileName, a.variables.GetProfile())
}

// LoadProfile skips a saved key whose variable has since been deleted;
// ApplyProfile, which takes a profile from a caller, does not.
func (a *Instance) LoadProfile(profileName string) error {
	profile, err := a.Profiles().Get(profileName)
	if err != nil {
		return err
	}
	known := make(variable.Profile, len(profile))
	for path, value := range profile {
		if a.variables.Exists(path) {
			known[path] = value
		}
	}
	return a.ApplyProfile(known)
}

func (a *Instance) ApplyProfile(profile variable.Profile) error {
	_, err := a.write(refuseConflicts, func() error {
		return a.variables.ApplyProfile(profile)
	})
	return err
}

func (a *Instance) VariantSets() *named.Collection[variant.Set] {
	return a.variantSets
}
