package nodes

import (
	"github.com/EliCDavis/polyform/refutil"
)

type StructNodeProcesor[T any] interface {
	Process() (T, error)
}

type StructNode[T any] struct {
	nodeData

	processir StructNodeProcesor[T]

	value       T
	err         error
	depVersions []int
}

func (tn StructNode[T]) Outdated() bool {

	// Nil versions means we've never processed before
	if tn.depVersions == nil {
		return true
	}

	deps := tn.Dependencies()

	// No dependencies, we can never be outdated
	if len(deps) == 0 {
		return false
	}

	for i, dep := range deps {
		if dep.Dependency().Version() != tn.depVersions[i] {
			return true
		}
	}

	return false
}

func (tn *StructNode[T]) updateUsedDependencyVersions() {
	deps := tn.Dependencies()
	tn.depVersions = make([]int, len(deps))
	for i, dep := range deps {
		tn.depVersions[i] = dep.Dependency().Version()
	}
}

type StructNodeDependency struct {
	name string
	dep  Dependency
}

func (tnd StructNodeDependency) Name() string {
	return tnd.name
}

func (tnd StructNodeDependency) Dependency() Dependency {
	return tnd.dep
}

func (tn StructNode[T]) Dependencies() []NodeDependency {
	data := refutil.FieldValuesOfType[Dependency](tn.processir)

	output := make([]NodeDependency, 0)
	for key, val := range data {
		output = append(output, StructNodeDependency{
			name: key,
			dep:  val,
		})
	}
	return output
}

func (tn StructNode[T]) Data() T {
	if tn.Outdated() {
		tn.process()
	}
	return tn.value
}

func (tn *StructNode[T]) process() {
	// tn.value, tn.err = tn.transform(tn.in)
	tn.value, tn.err = tn.processir.Process()
	tn.version++
	tn.state = Processed
	tn.updateUsedDependencyVersions()
}

func (tn *StructNode[T]) Name() string {
	return refutil.GetName(tn.processir)
}

func Struct[T any](p StructNodeProcesor[T]) *StructNode[T] {
	return &StructNode[T]{
		nodeData: nodeData{
			version: 0,
			state:   Stale,
			subs:    make([]Alertable, 0),
		},
		processir:   p,
		depVersions: nil,
	}
}
