package schema

import "strings"

type Variable struct {
	Description string `json:"description"`
	Type        string `json:"type"`
	Value       any    `json:"value"`
}

type NestedGroup[T any] struct {
	Variables map[string]T              `json:"variables"`
	SubGroups map[string]NestedGroup[T] `json:"subgroups"`
}

func (vg NestedGroup[T]) traverse(prepend string, f func(path string, variable T) bool) bool {
	for name, variable := range vg.Variables {
		if !f(prepend+name, variable) {
			return false
		}
	}
	for name, subgroup := range vg.SubGroups {
		if !subgroup.traverse(prepend+name+"/", f) {
			return false
		}
	}

	return true
}

// Nest files each entry under the groups its "/"-separated path names; the
// last segment is the entry's own name.
func Nest[T any](flat map[string]T) NestedGroup[T] {
	root := emptyGroup[T]()
	for path, value := range flat {
		segments := strings.Split(path, "/")
		group := root
		for _, segment := range segments[:len(segments)-1] {
			child, ok := group.SubGroups[segment]
			if !ok {
				child = emptyGroup[T]()
				group.SubGroups[segment] = child
			}
			group = child
		}
		group.Variables[segments[len(segments)-1]] = value
	}
	return root
}

// Both maps are allocated up front so a group copied out of a map still
// writes into the same storage.
func emptyGroup[T any]() NestedGroup[T] {
	return NestedGroup[T]{Variables: map[string]T{}, SubGroups: map[string]NestedGroup[T]{}}
}

func (vg NestedGroup[T]) Traverse(f func(path string, variable T) bool) {
	vg.traverse("", f)
}
