// Package enum describes a Go type's fixed set of values by name so the
// graph can save, edit and list them. The type stays an ordinary defined
// type with iota constants; Define adds the table Go cannot produce itself.
package enum

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/EliCDavis/polyform/refutil"
)

type Entry[T comparable] struct {
	value       T
	name        string
	description string
}

func Value[T comparable](v T, name, description string) Entry[T] {
	return Entry[T]{value: v, name: name, description: description}
}

type Set[T comparable] struct {
	typeName string
	entries  []Entry[T]
	byValue  map[T]int
	byName   map[string]int
}

type Definition interface {
	TypeName() string
	Names() []string
	Describe(name string) string
	NameOf(value any) (string, bool)
	ValueOf(name string) (any, error)
}

var (
	registryMu sync.RWMutex
	byTypeName = map[string]Definition{}
)

func Define[T comparable](entries ...Entry[T]) Set[T] {
	var zero T
	s := Set[T]{
		typeName: refutil.TypeResolution{IncludePackage: true}.Resolve(zero),
		entries:  entries,
		byValue:  make(map[T]int, len(entries)),
		byName:   make(map[string]int, len(entries)),
	}
	short := shortName[T]()
	if len(entries) == 0 {
		panic(fmt.Errorf("enum %s: no values listed", short))
	}
	for i, e := range entries {
		if strings.TrimSpace(e.name) == "" {
			panic(fmt.Errorf("enum %s: entry %d has no name", short, i))
		}
		if _, dup := s.byValue[e.value]; dup {
			panic(fmt.Errorf("enum %s: value %v listed twice", short, e.value))
		}
		if _, dup := s.byName[fold(e.name)]; dup {
			panic(fmt.Errorf("enum %s: name %q listed twice", short, e.name))
		}
		s.byValue[e.value] = i
		s.byName[fold(e.name)] = i
	}
	if _, ok := s.byValue[zero]; !ok {
		panic(fmt.Errorf("enum %s: the zero value %v must be listed; it is what an unwired port holds", short, zero))
	}

	registryMu.Lock()
	defer registryMu.Unlock()
	if _, dup := byTypeName[s.typeName]; dup {
		panic(fmt.Errorf("enum %s: defined twice", short))
	}
	byTypeName[s.typeName] = s
	return s
}

func Lookup(typeName string) (Definition, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	def, ok := byTypeName[typeName]
	return def, ok
}

func Definitions() []Definition {
	registryMu.RLock()
	defer registryMu.RUnlock()
	defs := make([]Definition, 0, len(byTypeName))
	for _, name := range slices.Sorted(maps.Keys(byTypeName)) {
		defs = append(defs, byTypeName[name])
	}
	return defs
}

func (s Set[T]) TypeName() string { return s.typeName }

func (s Set[T]) Values() []T {
	out := make([]T, len(s.entries))
	for i, e := range s.entries {
		out[i] = e.value
	}
	return out
}

func (s Set[T]) Names() []string {
	out := make([]string, len(s.entries))
	for i, e := range s.entries {
		out[i] = e.name
	}
	return out
}

func (s Set[T]) Name(v T) string {
	if i, ok := s.byValue[v]; ok {
		return s.entries[i].name
	}
	return ""
}

func (s Set[T]) Description(v T) string {
	if i, ok := s.byValue[v]; ok {
		return s.entries[i].description
	}
	return ""
}

func (s Set[T]) Describe(name string) string {
	if i, ok := s.byName[fold(name)]; ok {
		return s.entries[i].description
	}
	return ""
}

func (s Set[T]) Parse(name string) (T, error) {
	if i, ok := s.byName[fold(name)]; ok {
		return s.entries[i].value, nil
	}
	var zero T
	return zero, fmt.Errorf("%q is not %s; expected %s", name, withArticle(shortName[T]()), joinNames(s.Names()))
}

func (s Set[T]) NameOf(value any) (string, bool) {
	v, ok := value.(T)
	if !ok {
		return "", false
	}
	i, ok := s.byValue[v]
	if !ok {
		return "", false
	}
	return s.entries[i].name, true
}

func (s Set[T]) ValueOf(name string) (any, error) {
	return s.Parse(name)
}

func fold(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func shortName[T any]() string {
	var zero T
	return refutil.GetTypeNameWithoutPackage(zero)
}

func withArticle(noun string) string {
	if noun != "" && strings.ContainsRune("aeiouAEIOU", rune(noun[0])) {
		return "an " + noun
	}
	return "a " + noun
}

func joinNames(names []string) string {
	switch len(names) {
	case 0:
		return "nothing"
	case 1:
		return names[0]
	}
	return strings.Join(names[:len(names)-1], ", ") + " or " + names[len(names)-1]
}
