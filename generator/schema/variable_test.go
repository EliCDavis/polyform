package schema_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/stretchr/testify/assert"
)

func TestNestedGroupTraverse_empty(t *testing.T) {
	var group schema.NestedGroup[int]

	var visited []string
	group.Traverse(func(path string, value int) bool {
		visited = append(visited, path)
		return true
	})

	assert.Empty(t, visited)
}

func TestNestedGroupTraverse_rootVariables(t *testing.T) {
	group := schema.NestedGroup[int]{
		Variables: map[string]int{
			"a": 1,
			"b": 2,
		},
	}

	visited := make(map[string]int)
	group.Traverse(func(path string, value int) bool {
		visited[path] = value
		return true
	})

	assert.Equal(t, map[string]int{"a": 1, "b": 2}, visited)
}

func TestNestedGroupTraverse_nestedSubgroups(t *testing.T) {
	group := schema.NestedGroup[int]{
		Variables: map[string]int{
			"root": 0,
		},
		SubGroups: map[string]schema.NestedGroup[int]{
			"group1": {
				Variables: map[string]int{
					"var1": 1,
				},
				SubGroups: map[string]schema.NestedGroup[int]{
					"group2": {
						Variables: map[string]int{
							"var2": 2,
						},
					},
				},
			},
		},
	}

	visited := make(map[string]int)
	group.Traverse(func(path string, value int) bool {
		visited[path] = value
		return true
	})

	assert.Equal(t, map[string]int{
		"root":               0,
		"group1/var1":        1,
		"group1/group2/var2": 2,
	}, visited)
}

func TestNestedGroupTraverse_stopsAtRoot(t *testing.T) {
	group := schema.NestedGroup[int]{
		Variables: map[string]int{
			"stop": 1,
		},
		SubGroups: map[string]schema.NestedGroup[int]{
			"child": {
				Variables: map[string]int{
					"nested": 2,
				},
			},
		},
	}

	var visited []string
	group.Traverse(func(path string, value int) bool {
		visited = append(visited, path)
		return false
	})

	assert.Equal(t, []string{"stop"}, visited)
}

func TestNestedGroupTraverse_stopsInSubgroup(t *testing.T) {
	group := schema.NestedGroup[int]{
		SubGroups: map[string]schema.NestedGroup[int]{
			"parent": {
				Variables: map[string]int{
					"first": 1,
				},
				SubGroups: map[string]schema.NestedGroup[int]{
					"child": {
						Variables: map[string]int{
							"deep": 2,
						},
					},
				},
			},
		},
	}

	var visited []string
	group.Traverse(func(path string, value int) bool {
		visited = append(visited, path)
		return path != "parent/first"
	})

	assert.Equal(t, []string{"parent/first"}, visited)
}
