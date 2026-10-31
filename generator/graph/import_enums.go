package graph

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/persistence"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/subgraph"
)

type ImportedEnum struct {
	ID string
	// Set when ID had to differ from the id in the payload.
	OriginalID string
	Name       string
	// The target already had this enum, by name and members, and the import
	// points at it instead of adding a copy.
	Reused bool
}

func referencedEnumIDs(app persistence.App) []string {
	referenced := map[string]bool{}
	note := func(typeName string) {
		if id := enum.RuntimeTypeID(strings.TrimPrefix(typeName, "[]")); id != "" {
			referenced[id] = true
		}
	}
	for _, def := range app.SubGraphs {
		for _, node := range def.Nodes {
			if id := EnumParameterTypeID(node.Type); id != "" {
				referenced[id] = true
			}
			if id := EnumMembersTypeID(node.Type); id != "" {
				referenced[id] = true
			}
			if node.Type == subgraph.InputNodeTypeKey || node.Type == subgraph.OutputNodeTypeKey {
				note(boundaryPortType(node.Data))
			}
			if node.Variable != nil {
				if saved, ok := savedVariable(app.Variables, *node.Variable); ok {
					note(savedVariableType(saved))
				}
			}
		}
	}
	return slices.Sorted(maps.Keys(referenced))
}

func boundaryPortType(data json.RawMessage) string {
	var boundary struct {
		PortType string `json:"portType"`
	}
	_ = json.Unmarshal(data, &boundary)
	return boundary.PortType
}

func savedVariableType(saved persistence.Variable) string {
	var header struct {
		Type string `json:"type"`
	}
	_ = json.Unmarshal(saved.Data, &header)
	return header.Type
}

// Same name and same members reuses the target's enum; the same name with
// other members becomes a separate enum with a suffixed id and name.
func (a *Instance) mergeEnums(app persistence.App, referenced []string) (idMap map[string]string, imported []ImportedEnum, created []string, err error) {
	idMap = make(map[string]string, len(referenced))
	for _, oldID := range referenced {
		def, ok := app.Enums[oldID]
		if !ok {
			return nil, nil, created, fmt.Errorf("imported sub-graphs reference enum %q, which the imported graph does not define", oldID)
		}
		if existingID, ok := a.matchingEnum(def); ok {
			idMap[oldID] = existingID
			imported = append(imported, ImportedEnum{ID: existingID, OriginalID: differing(oldID, existingID), Name: a.enums[existingID].name, Reused: true})
			continue
		}
		newID := a.freeEnumID(oldID)
		name := a.freeEnumName(def.Name)
		if err := a.createEnum(newID, name, def.Description, def.Members); err != nil {
			return nil, nil, created, err
		}
		created = append(created, newID)
		idMap[oldID] = newID
		imported = append(imported, ImportedEnum{ID: newID, OriginalID: differing(oldID, newID), Name: name})
	}
	return idMap, imported, created, nil
}

func differing(original, actual string) string {
	if original == actual {
		return ""
	}
	return original
}

func (a *Instance) matchingEnum(def persistence.Enum) (string, bool) {
	wanted := foldedNames(def.Members)
	for _, id := range slices.Sorted(maps.Keys(a.enums)) {
		existing := a.enums[id]
		if !strings.EqualFold(existing.name, def.Name) {
			continue
		}
		if slices.Equal(foldedNames(existing.table.Members()), wanted) {
			return id, true
		}
	}
	return "", false
}

func foldedNames(members []enum.Member) []string {
	names := make([]string, len(members))
	for i, m := range members {
		names[i] = strings.ToLower(strings.TrimSpace(m.Name))
	}
	slices.Sort(names)
	return names
}

func (a *Instance) freeEnumName(base string) string {
	taken := func(name string) bool {
		for _, def := range a.enums {
			if strings.EqualFold(def.name, name) {
				return true
			}
		}
		return false
	}
	if !taken(base) {
		return base
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s_%d", base, i)
		if !taken(candidate) {
			return candidate
		}
	}
}

func rewriteEnumIDs(app persistence.App, idMap map[string]string) persistence.App {
	rewriteType := func(typeName string) string {
		prefix, elem := "", typeName
		for strings.HasPrefix(elem, "[]") {
			prefix += "[]"
			elem = elem[2:]
		}
		if newID, ok := idMap[enum.RuntimeTypeID(elem)]; ok {
			return prefix + enum.RuntimeTypePath(newID)
		}
		return typeName
	}

	subGraphs := make(map[string]persistence.SubGraph, len(app.SubGraphs))
	for id, def := range app.SubGraphs {
		def.Nodes = maps.Clone(def.Nodes)
		for nodeID, node := range def.Nodes {
			if oldID := EnumParameterTypeID(node.Type); oldID != "" {
				if newID, ok := idMap[oldID]; ok {
					node.Type = EnumParameterTypePath(newID)
				}
			}
			if oldID := EnumMembersTypeID(node.Type); oldID != "" {
				if newID, ok := idMap[oldID]; ok {
					node.Type = EnumMembersTypePath(newID)
				}
			}
			if node.Type == subgraph.InputNodeTypeKey || node.Type == subgraph.OutputNodeTypeKey {
				node.Data = rewriteJSONField(node.Data, "portType", rewriteType)
			}
			def.Nodes[nodeID] = node
		}
		subGraphs[id] = def
	}
	app.SubGraphs = subGraphs
	app.Variables = rewriteVariableTypes(app.Variables, rewriteType)
	return app
}

func rewriteVariableTypes(group schema.NestedGroup[persistence.Variable], rewrite func(string) string) schema.NestedGroup[persistence.Variable] {
	out := schema.NestedGroup[persistence.Variable]{
		Variables: make(map[string]persistence.Variable, len(group.Variables)),
		SubGroups: make(map[string]schema.NestedGroup[persistence.Variable], len(group.SubGroups)),
	}
	for name, v := range group.Variables {
		v.Data = rewriteJSONField(v.Data, "type", rewrite)
		out.Variables[name] = v
	}
	for name, sub := range group.SubGroups {
		out.SubGroups[name] = rewriteVariableTypes(sub, rewrite)
	}
	return out
}

func rewriteJSONField(data json.RawMessage, field string, rewrite func(string) string) json.RawMessage {
	var fields map[string]json.RawMessage
	if json.Unmarshal(data, &fields) != nil {
		return data
	}
	var value string
	if json.Unmarshal(fields[field], &value) != nil {
		return data
	}
	rewritten := rewrite(value)
	if rewritten == value {
		return data
	}
	fields[field], _ = json.Marshal(rewritten)
	out, err := json.Marshal(fields)
	if err != nil {
		return data
	}
	return out
}
