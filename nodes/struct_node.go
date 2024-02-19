package nodes

import (
	"fmt"

	"github.com/EliCDavis/polyform/refutil"
)

// ============================================================================

type StructNodeOutputDefinition[T any] interface {
	StructNodeProcesor[T]
	IStructData[T]
}

type StructNodeOutput[T any] struct {
	Definition StructNodeOutputDefinition[T]
}

func (sno StructNodeOutput[T]) Data() T {
	return sno.Definition.node(sno.Definition).Data()
}

func (sno StructNodeOutput[T]) Node() Node {
	return sno.Definition.node(sno.Definition)
}

// ============================================================================

type IStructData[T any] interface {
	node(p StructNodeProcesor[T]) *StructNode[T]
}

type StructData[T any] struct {
	n *StructNode[T]
}

func (bd *StructData[T]) node(p StructNodeProcesor[T]) *StructNode[T] {
	if bd.n == nil {
		bd.n = Struct(p)
	}

	return bd.n
}

// ============================================================================

type StructNodeProcesor[T any] interface {
	Process() (T, error)
}

type StructNode[T any] struct {
	processir StructNodeProcesor[T]

	value       T
	err         error
	depVersions []int

	version int
	subs    []Alertable
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

	for i, nodeDep := range deps {
		dep := nodeDep.Dependency()
		if dep.Version() != tn.depVersions[i] || dep.State() != Processed {
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
	dep  Node
}

func (tnd StructNodeDependency) Name() string {
	return tnd.name
}

func (tnd StructNodeDependency) Dependency() Node {
	return tnd.dep
}

func (tn StructNode[T]) Version() int {
	return tn.version
}

func (tn StructNode[T]) Dependencies() []NodeDependency {
	output := make([]NodeDependency, 0)

	basicData := refutil.FieldValuesOfType[ReferencesNode](tn.processir)
	for key, val := range basicData {
		output = append(output, StructNodeDependency{
			name: key,
			dep:  val.Node(),
		})
	}

	arrayData := refutil.FieldValuesOfTypeInArray[ReferencesNode](tn.processir)
	for key, field := range arrayData {
		for i, e := range field {
			if e == nil {
				continue
			}

			output = append(output, StructNodeDependency{
				name: fmt.Sprintf("%s.%d", key, i),
				dep:  e.Node(),
			})
		}
	}
	return output
}

func (tn *StructNode[T]) Data() T {
	if tn.Outdated() {
		tn.process()
	}
	return tn.value
}

func (tn *StructNode[T]) process() {
	// tn.value, tn.err = tn.transform(tn.in)
	tn.value, tn.err = tn.processir.Process()
	tn.version++
	tn.updateUsedDependencyVersions()
}

func (tn *StructNode[T]) Name() string {
	return refutil.GetName(tn.processir)
}

func (tn *StructNode[T]) State() NodeState {
	if tn.Outdated() {
		return Stale
	}
	return Processed
}

func (tn *StructNode[T]) AddSubscription(a Alertable) {
	if a == nil {
		panic("attempting to subribe with nil alertable")
	}
	tn.subs = append(tn.subs, a)
}

func Struct[T any](p StructNodeProcesor[T]) *StructNode[T] {
	return &StructNode[T]{
		version:     0,
		subs:        make([]Alertable, 0),
		processir:   p,
		depVersions: nil,
	}
}
