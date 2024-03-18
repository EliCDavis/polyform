package nodes

import (
	"fmt"

	"github.com/EliCDavis/polyform/refutil"
)

type TransformerNode[Tin any, Tout any] struct {
	nodeData

	value       Tout
	err         error
	depVersions []int

	name      string
	in        Tin
	transform func(in Tin) (Tout, error)
}

func (tn TransformerNode[Tin, Tout]) Outdated() bool {

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

func (tn *TransformerNode[Tin, Tout]) updateUsedDependencyVersions() {
	deps := tn.Dependencies()
	tn.depVersions = make([]int, len(deps))
	for i, dep := range deps {
		tn.depVersions[i] = dep.Dependency().Version()
	}
}

type transformerNodeDependency struct {
	name string
	dep  Node
}

func (tnd transformerNodeDependency) Name() string {
	return tnd.name
}

func (tnd transformerNodeDependency) Dependency() Node {
	return tnd.dep
}

func (in *TransformerNode[Tin, Tout]) Node() Node {
	return in
}

func (tn TransformerNode[Tin, Tout]) Inputs() []Input {
	return []Input{{
		Name: "In",
		Type: fmt.Sprintf("%T", *new(Tin)),
	}}
}

func (tn TransformerNode[Tin, Tout]) Outputs() []Output {
	return []Output{
		{
			Name: "Data",
			Type: fmt.Sprintf("%T", *new(Tout)),
		},
	}
}

func (tn TransformerNode[Tin, Tout]) Dependencies() []NodeDependency {

	// The input for the transformer is a node itself,
	if dep, ok := any(tn.in).(Node); ok {
		return []NodeDependency{
			transformerNodeDependency{
				name: "Input",
				dep:  dep,
			},
		}
	}

	if dep, ok := any(tn.in).(ReferencesNode); ok {
		return []NodeDependency{
			transformerNodeDependency{
				name: "Input",
				dep:  dep.Node(),
			},
		}
	}

	data := refutil.FieldValuesOfType[ReferencesNode](tn.in)

	output := make([]NodeDependency, 0)
	for key, val := range data {
		output = append(output, transformerNodeDependency{
			name: key,
			dep:  val.Node(),
		})
	}
	return output
	// return refutil.FieldValuesOfType[Dependency](tn.in)
}

func (tn TransformerNode[Tin, Tout]) Data() Tout {
	if tn.Outdated() {
		tn.process()
	}
	return tn.value
}

func (tn *TransformerNode[Tin, Tout]) process() {
	tn.value, tn.err = tn.transform(tn.in)
	tn.version++
	tn.state = Processed
	tn.updateUsedDependencyVersions()
}

func (tn *TransformerNode[Tin, Tout]) Name() string {
	return tn.name
}

func Transformer[Tin any, Tout any](name string, in Tin, trasnformer func(in Tin) (Tout, error)) *TransformerNode[Tin, Tout] {
	return &TransformerNode[Tin, Tout]{
		nodeData: nodeData{
			version: 0,
			state:   Stale,
			subs:    make([]Alertable, 0),
		},
		name:        name,
		in:          in,
		transform:   trasnformer,
		depVersions: nil,
	}
}
