package enum

import (
	"fmt"
	"strings"
)

const runtimeTypePrefix = "enum/"

func RuntimeTypeID(typeName string) string {
	if len(typeName) <= len(runtimeTypePrefix) || typeName[:len(runtimeTypePrefix)] != runtimeTypePrefix {
		return ""
	}
	return typeName[len(runtimeTypePrefix):]
}

func RuntimeTypePath(id string) string {
	return runtimeTypePrefix + id
}

type Member struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Table is a Definition declared at run time rather than by a Go type; a
// graph's own enums are tables.
type Table struct {
	typeName string
	members  []Member
}

func NewTable(typeName string, members []Member) (Table, error) {
	if len(members) == 0 {
		return Table{}, fmt.Errorf("an enum needs at least one member")
	}
	seen := make(map[string]bool, len(members))
	for i, m := range members {
		if strings.TrimSpace(m.Name) == "" {
			return Table{}, fmt.Errorf("member %d has no name", i)
		}
		if seen[fold(m.Name)] {
			return Table{}, fmt.Errorf("member %q listed twice", m.Name)
		}
		seen[fold(m.Name)] = true
	}
	return Table{typeName: typeName, members: append([]Member(nil), members...)}, nil
}

func (t Table) TypeName() string { return t.typeName }

func (t Table) Members() []Member {
	return append([]Member(nil), t.members...)
}

func (t Table) Names() []string {
	out := make([]string, len(t.members))
	for i, m := range t.members {
		out[i] = m.Name
	}
	return out
}

func (t Table) Describe(name string) string {
	if i := t.Index(name); i >= 0 {
		return t.members[i].Description
	}
	return ""
}

func (t Table) Index(name string) int {
	folded := fold(name)
	for i, m := range t.members {
		if fold(m.Name) == folded {
			return i
		}
	}
	return -1
}

func (t Table) Canonical(name string) (string, error) {
	if i := t.Index(name); i >= 0 {
		return t.members[i].Name, nil
	}
	return "", fmt.Errorf("%q is not a member; expected %s", name, joinNames(t.Names()))
}

func (t Table) Choice(name string) Choice {
	return Choice{Set: t.typeName, Name: name, Index: t.Index(name)}
}

type Choice struct {
	Set   string
	Name  string
	Index int
}
