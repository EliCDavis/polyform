package graph

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode"

	"encoding/json"
	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/persistence"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/nodes"
)

type graphEnum struct {
	name        string
	description string
	table       enum.Table
}

type EnumUpdate struct {
	Name        string
	Description string
	Members     []enum.Member
	// Old member name to new, for members whose name changed rather than
	// being removed and added.
	Renamed map[string]string
}

func (a *Instance) CreateEnum(name, description string, members []enum.Member) (id string, err error) {
	_, err = a.write(refuseConflicts, func() error {
		id = a.freeEnumID(enumIDBase(name))
		return a.createEnum(id, name, description, members)
	})
	return id, err
}

func (a *Instance) createEnum(id, name, description string, members []enum.Member) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("an enum needs a name")
	}
	if _, exists := a.enums[id]; exists {
		return fmt.Errorf("enum %q already exists", id)
	}
	table, err := enum.NewTable(enum.RuntimeTypePath(id), members)
	if err != nil {
		return fmt.Errorf("enum %q: %w", name, err)
	}
	a.enums[id] = &graphEnum{name: name, description: description, table: table}
	a.typeFactory.RegisterBuilder(EnumParameterTypePath(id), func() any {
		return NewEnumParameter(a.enums[id].table, a.enums[id].name)
	})
	a.typeFactory.RegisterBuilder(EnumMembersTypePath(id), func() any {
		return NewEnumMembers(a.enums[id].table, a.enums[id].name)
	})
	nodes.DiscoverNodePortTypes(NewEnumParameter(table, name))
	nodes.DiscoverNodePortTypes(NewEnumMembers(table, name))
	return nil
}

func (a *Instance) UpdateEnum(id string, update EnumUpdate) error {
	_, err := a.write(refuseConflicts, func() error {
		return a.updateEnum(id, update)
	})
	return err
}

func (a *Instance) updateEnum(id string, update EnumUpdate) error {
	def, ok := a.enums[id]
	if !ok {
		return fmt.Errorf("enum %q does not exist", id)
	}
	if strings.TrimSpace(update.Name) == "" {
		return fmt.Errorf("an enum needs a name")
	}
	table, err := enum.NewTable(enum.RuntimeTypePath(id), update.Members)
	if err != nil {
		return fmt.Errorf("enum %q: %w", update.Name, err)
	}
	for old, renamed := range update.Renamed {
		if def.table.Index(old) < 0 {
			return fmt.Errorf("enum %q has no member %q to rename", def.name, old)
		}
		if table.Index(renamed) < 0 {
			return fmt.Errorf("member %q is renamed to %q, which is not in the new members", old, renamed)
		}
	}

	var removed []string
	for _, old := range def.table.Names() {
		if _, renamed := update.Renamed[old]; renamed || table.Index(old) >= 0 {
			continue
		}
		removed = append(removed, old)
	}
	if len(removed) > 0 {
		held := a.enumMemberHolders(id, removed)
		if len(held) > 0 {
			return fmt.Errorf("%s still used by %s", quoteJoin(removed), strings.Join(held, ", "))
		}
	}

	def.name = update.Name
	def.description = update.Description
	def.table = table
	err = a.forEachEnumHolder(id, func(set func(table enum.Table), value string, rename func(string) error) error {
		set(table)
		if renamed, ok := update.Renamed[value]; ok {
			return rename(renamed)
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, members := range a.enumMembersNodes(id) {
		members.SetTable(table)
	}
	a.renameEnumMembersInProfiles(id, update.Renamed)
	return nil
}

func (a *Instance) enumMembersNodes(id string) []*EnumMembersNode {
	typeName := enum.RuntimeTypePath(id)
	var found []*EnumMembersNode
	for _, scope := range a.scopes() {
		for node := range scope.graph.nodeIDs {
			if members, ok := node.(*EnumMembersNode); ok && members.Table().TypeName() == typeName {
				found = append(found, members)
			}
		}
	}
	return found
}

func (a *Instance) DeleteEnum(id string) error {
	_, err := a.write(refuseConflicts, func() error {
		return a.deleteEnum(id)
	})
	return err
}

func (a *Instance) deleteEnum(id string) error {
	def, ok := a.enums[id]
	if !ok {
		return fmt.Errorf("enum %q does not exist", id)
	}
	if holders := a.enumTypeHolders(id); len(holders) > 0 {
		return fmt.Errorf("enum %q is still used by %s", def.name, strings.Join(holders, ", "))
	}
	a.typeFactory.Unregister(EnumParameterTypePath(id))
	a.typeFactory.Unregister(EnumMembersTypePath(id))
	delete(a.enums, id)
	return nil
}

func (a *Instance) Enums() map[string]schema.Enum {
	a.mu().RLock()
	defer a.mu().RUnlock()
	return a.enumSchemas()
}

func (a *Instance) enumSchemas() map[string]schema.Enum {
	out := make(map[string]schema.Enum, len(a.enums))
	for id, def := range a.enums {
		out[id] = schema.Enum{Name: def.name, Description: def.description, Members: def.table.Members()}
	}
	return out
}

func (a *Instance) persistedEnums() map[string]persistence.Enum {
	out := make(map[string]persistence.Enum, len(a.enums))
	for id, def := range a.enums {
		out[id] = persistence.Enum{Name: def.name, Description: def.description, Members: def.table.Members()}
	}
	return out
}

func (a *Instance) loadEnums(saved map[string]persistence.Enum) error {
	for _, id := range slices.Sorted(maps.Keys(saved)) {
		def := saved[id]
		if err := a.createEnum(id, def.Name, def.Description, def.Members); err != nil {
			return err
		}
	}
	return nil
}

// EnumTable resolves a run-time enum port or variable type to its table.
func (a *Instance) EnumTable(typeName string) (enum.Table, bool) {
	a.mu().RLock()
	defer a.mu().RUnlock()
	return a.enumTable(typeName)
}

func (a *Instance) enumTable(typeName string) (enum.Table, bool) {
	def, ok := a.enums[enum.RuntimeTypeID(strings.TrimPrefix(typeName, "[]"))]
	if !ok {
		return enum.Table{}, false
	}
	return def.table, true
}

func (a *Instance) EnumParameterNodeType(id string) (schema.NodeType, error) {
	a.mu().RLock()
	defer a.mu().RUnlock()
	def, ok := a.enums[id]
	if !ok {
		return schema.NodeType{}, fmt.Errorf("enum %q does not exist", id)
	}
	typePath := EnumParameterTypePath(id)
	return BuildNodeTypeSchema(typePath, NewEnumParameter(def.table, def.name)), nil
}

func (a *Instance) EnumMembersNodeType(id string) (schema.NodeType, error) {
	a.mu().RLock()
	defer a.mu().RUnlock()
	def, ok := a.enums[id]
	if !ok {
		return schema.NodeType{}, fmt.Errorf("enum %q does not exist", id)
	}
	typePath := EnumMembersTypePath(id)
	return BuildNodeTypeSchema(typePath, NewEnumMembers(def.table, def.name)), nil
}

func (a *Instance) freeEnumID(base string) string {
	if _, taken := a.enums[base]; !taken {
		return base
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s_%d", base, i)
		if _, taken := a.enums[candidate]; !taken {
			return candidate
		}
	}
}

func enumIDBase(name string) string {
	var b strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "Enum"
	}
	return b.String()
}

// Everything holding a value of the enum: variables by path, parameter nodes
// as "scope/nodeId" (root nodes have no scope), and the setter for each.
func (a *Instance) forEachEnumHolder(id string, visit func(set func(enum.Table), value string, rename func(string) error) error) error {
	typeName := enum.RuntimeTypePath(id)
	var err error
	a.variables.Traverse(func(path string, info variable.Info, v variable.Variable) {
		ev, ok := v.(*variable.EnumVariable)
		if !ok || ev.Table().TypeName() != typeName || err != nil {
			return
		}
		err = visit(ev.SetTable, ev.GetValue(), ev.SetValue)
	})
	if err != nil {
		return err
	}
	for _, scope := range a.scopes() {
		for node := range scope.graph.nodeIDs {
			param, ok := node.(*EnumParameterNode)
			if !ok || param.Table().TypeName() != typeName {
				continue
			}
			if err := visit(param.SetTable, param.Value(), param.SetValue); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Instance) enumMemberHolders(id string, members []string) []string {
	var held []string
	typeName := enum.RuntimeTypePath(id)
	a.variables.Traverse(func(path string, info variable.Info, v variable.Variable) {
		if ev, ok := v.(*variable.EnumVariable); ok && ev.Table().TypeName() == typeName && slices.Contains(members, ev.GetValue()) {
			held = append(held, "variable "+path)
		}
	})
	for _, scope := range a.scopes() {
		for node, nodeID := range scope.graph.nodeIDs {
			if param, ok := node.(*EnumParameterNode); ok && param.Table().TypeName() == typeName && slices.Contains(members, param.Value()) {
				held = append(held, "parameter "+scope.prefix+nodeID)
			}
		}
	}
	paths := a.enumVariablePaths(id)
	for name, profile := range a.profiles.All() {
		for _, path := range paths {
			raw, ok := profile[path]
			if !ok {
				continue
			}
			var value string
			if json.Unmarshal(raw, &value) == nil && slices.Contains(members, value) {
				held = append(held, "profile "+name)
				break
			}
		}
	}
	slices.Sort(held)
	return held
}

func (a *Instance) enumTypeHolders(id string) []string {
	var held []string
	typeName := enum.RuntimeTypePath(id)
	for _, path := range a.enumVariablePaths(id) {
		held = append(held, "variable "+path)
	}
	for _, scope := range a.scopes() {
		for node, nodeID := range scope.graph.nodeIDs {
			if param, ok := node.(*EnumParameterNode); ok && param.Table().TypeName() == typeName {
				held = append(held, "parameter "+scope.prefix+nodeID)
			}
			if members, ok := node.(*EnumMembersNode); ok && members.Table().TypeName() == typeName {
				held = append(held, "members "+scope.prefix+nodeID)
			}
			if boundary, ok := subgraph.IsBoundaryNode(node); ok && boundary.BoundaryPortType() == typeName {
				held = append(held, "port "+scope.prefix+nodeID)
			}
		}
	}
	slices.Sort(held)
	return held
}

func (a *Instance) enumVariablePaths(id string) []string {
	var paths []string
	typeName := enum.RuntimeTypePath(id)
	a.variables.Traverse(func(path string, info variable.Info, v variable.Variable) {
		if ev, ok := v.(*variable.EnumVariable); ok && ev.Table().TypeName() == typeName {
			paths = append(paths, path)
		}
	})
	slices.Sort(paths)
	return paths
}

func (a *Instance) renameEnumMembersInProfiles(id string, renamed map[string]string) {
	if len(renamed) == 0 {
		return
	}
	paths := a.enumVariablePaths(id)
	for name, profile := range a.profiles.All() {
		changed := false
		for _, path := range paths {
			raw, ok := profile[path]
			if !ok {
				continue
			}
			var value string
			if json.Unmarshal(raw, &value) != nil {
				continue
			}
			if to, ok := renamed[value]; ok {
				profile[path], _ = json.Marshal(to)
				changed = true
			}
		}
		if changed {
			a.profiles.Set(name, profile)
		}
	}
}

func quoteJoin(names []string) string {
	quoted := make([]string, len(names))
	for i, n := range names {
		quoted[i] = fmt.Sprintf("%q", n)
	}
	return strings.Join(quoted, ", ")
}
