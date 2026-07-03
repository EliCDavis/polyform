package schema

type RuntimeVariable struct {
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

func (vg NestedGroup[T]) Traverse(f func(path string, variable T) bool) {
	vg.traverse("", f)
}
