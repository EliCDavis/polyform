package nodes

import (
	"fmt"
	"strings"

	"github.com/EliCDavis/polyform/refutil"
)

// ============================================================================

// type StructNodeOutputDefinition[T any] interface {
// 	StructNodeProcesor[T]
// 	IStructData[T]
// }

type StructNodeOutput[T any, G StructNodeProcesor[T]] struct {
	Struct *StructNode[T, G]
	Name   string
}

func (sno StructNodeOutput[T, G]) Value() T {
	return sno.Struct.Value()
}

func (sno StructNodeOutput[T, G]) Node() Node {
	return sno.Struct
}

func (sno StructNodeOutput[T, G]) Port() string {
	return sno.Name
}

// ============================================================================

// type IStructData[T any] interface {
// 	node(p StructNodeProcesor[T]) *StructNode[T]
// }

// type StructData[T any] struct {
// 	n *StructNode[T]
// }

// func (bd *StructData[T]) node(p StructNodeProcesor[T]) *StructNode[T] {
// 	if bd.n == nil {
// 		bd.n = Struct(p)
// 	}

// 	return bd.n
// }

// ============================================================================

type StructNodeProcesor[T any] interface {
	Process() (T, error)
}

type StructNode[T any, G StructNodeProcesor[T]] struct {
	Data G

	value       T
	err         error
	depVersions []int

	version int
	subs    []Alertable
}

func (sn *StructNode[T, G]) SetInput(input string, output Output) {
	refutil.SetStructField(&sn.Data, input, output.NodeOutput)
}

func (sn StructNode[T, G]) Outdated() bool {

	// Nil versions means we've never processed before
	if sn.depVersions == nil {
		return true
	}

	deps := sn.Dependencies()

	// No dependencies, we can never be outdated
	if len(deps) == 0 {
		return false
	}

	for i, nodeDep := range deps {
		dep := nodeDep.Dependency()
		if dep.Version() != sn.depVersions[i] || dep.State() != Processed {
			return true
		}
	}

	return false
}

func (sn *StructNode[T, G]) updateUsedDependencyVersions() {
	deps := sn.Dependencies()
	sn.depVersions = make([]int, len(deps))
	for i, dep := range deps {
		sn.depVersions[i] = dep.Dependency().Version()
	}
}

type StructNodeDependency struct {
	name           string
	dep            Node
	dependencyPort string
}

func (snd StructNodeDependency) Name() string {
	return snd.name
}

func (snd StructNodeDependency) Dependency() Node {
	return snd.dep
}

func (snd StructNodeDependency) DependencyPort() string {
	return snd.dependencyPort
}

func (sn StructNode[T, G]) Version() int {
	return sn.version
}

func (sn *StructNode[T, G]) Out() StructNodeOutput[T, G] {
	return StructNodeOutput[T, G]{
		Struct: sn,
		Name:   "Out",
	}
}

func (sn *StructNode[T, G]) Outputs() []Output {
	// outputs := refutil.FuncValuesOfType[ReferencesNode](tn.Data)

	// outs := make([]Output, len(outputs))
	// var v *T = new(T)
	// for i, o := range outputs {
	// 	outs[i] = Output{
	// 		Name: o,
	// 		// Type: fmt.Sprintf("%T", *new(T)),
	// 		Type: refutil.GetTypeWithPackage(v),
	// 	}
	// }
	// return outs

	// TODO: This is wrong for nodes with more than one output
	return []Output{
		{
			Type: refutil.GetTypeWithPackage(new(T)),
			NodeOutput: StructNodeOutput[T, G]{
				Name:   "Out",
				Struct: sn,
			},
		},
	}
}

func (sn StructNode[T, G]) Inputs() []Input {
	nodeInputs := make([]Input, 0)

	refInput := refutil.GenericFieldValues("nodes.NodeOutput", sn.Data)
	for name, inputType := range refInput {
		nodeInputs = append(nodeInputs, Input{Name: name, Type: inputType})
	}

	refArrInput := refutil.GenericFieldValues("[]nodes.NodeOutput", sn.Data)
	for name, inputType := range refArrInput {
		nodeInputs = append(nodeInputs, Input{Name: name, Type: "[]" + inputType})
	}

	return nodeInputs
}

func (sn StructNode[T, G]) Dependencies() []NodeDependency {
	output := make([]NodeDependency, 0)

	basicData := refutil.FieldValuesOfType[NodeOutputReference](sn.Data)
	for key, val := range basicData {
		output = append(output, StructNodeDependency{
			name:           key,
			dep:            val.Node(),
			dependencyPort: val.Port(),
		})
	}

	arrayData := refutil.FieldValuesOfTypeInArray[NodeOutputReference](sn.Data)
	for key, field := range arrayData {
		for i, e := range field {
			if e == nil {
				continue
			}

			output = append(output, StructNodeDependency{
				name:           fmt.Sprintf("%s.%d", key, i),
				dep:            e.Node(),
				dependencyPort: e.Port(),
			})
		}
	}
	return output
}

func (sn *StructNode[T, G]) Value() T {
	if sn.Outdated() {
		sn.process()
	}
	return sn.value
}

func (sn *StructNode[T, G]) Node() Node {
	return sn
}

func (sn *StructNode[T, G]) Port() string {
	return "Out"
}

func (sn *StructNode[T, G]) process() {
	// tn.value, tn.err = tn.transform(tn.in)
	sn.value, sn.err = sn.Data.Process()
	sn.version++
	sn.updateUsedDependencyVersions()
}

func (sn StructNode[T, G]) Name() string {
	return refutil.GetTypeNameWithoutPackage(sn.Data)
}

func (sn StructNode[T, G]) Type() string {
	return refutil.GetTypeNameWithoutPackage(sn.Data)
}

func (sn StructNode[T, G]) Path() string {
	packagePath := refutil.GetPackagePath(sn.Data)
	if !strings.Contains(packagePath, "/") {
		return packagePath
	}

	path := strings.Split(packagePath, "/")
	path = path[1:]
	if path[0] == "EliCDavis" {
		path = path[1:]
	}

	if path[0] == "polyform" {
		path = path[1:]
	}
	return strings.Join(path, "/")
}

func (sn *StructNode[T, G]) State() NodeState {
	if sn.Outdated() {
		return Stale
	}
	return Processed
}

func (sn *StructNode[T, G]) AddSubscription(a Alertable) {
	if a == nil {
		panic("attempting to subscribe with nil alertable")
	}
	sn.subs = append(sn.subs, a)
}

func Struct[T StructNodeProcesor[G], G any](p T) *StructNode[G, T] {
	return &StructNode[G, T]{
		Data:        p,
		version:     0,
		subs:        make([]Alertable, 0),
		depVersions: nil,
	}
}
