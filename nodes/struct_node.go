package nodes

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/EliCDavis/polyform/refutil"
)

// ============================================================================

// type StructOutputDefinition[T any] interface {
// 	StructProcesor[T]
// 	IStructData[T]
// }

type StructOutput[T any, G StructProcesor[T]] struct {
	Struct *Struct[T, G]
	Name   string
}

func (sno StructOutput[T, G]) Value() T {
	return sno.Struct.Value()
}

func (sno StructOutput[T, G]) Node() Node {
	return sno.Struct
}

func (sno StructOutput[T, G]) Port() string {
	return sno.Name
}

// ============================================================================

// type IStructData[T any] interface {
// 	node(p StructProcesor[T]) *Struct[T]
// }

// type StructData[T any] struct {
// 	n *Struct[T]
// }

// func (bd *StructData[T]) node(p StructProcesor[T]) *Struct[T] {
// 	if bd.n == nil {
// 		bd.n = Struct(p)
// 	}

// 	return bd.n
// }

// ============================================================================

type StructProcesor[T any] interface {
	Process() (T, error)
}

type Struct[T any, G StructProcesor[T]] struct {
	Data G

	value                        T
	err                          error
	depVersions                  []int
	inputChangedSinceLastProcess bool

	version int
	subs    []Alertable
}

func (sn *Struct[T, G]) SetInput(input string, output Output) {
	// HHHHACK for array types atm.
	// I don't have enough knowledge yet for a proper refactor. I need to add
	// more features to see where things go fucky
	if index := strings.Index(input, "."); index != -1 {
		inputName := input[:index]
		if output.NodeOutput == nil {
			index, err := strconv.Atoi(input[index+1:])
			if err != nil {
				panic(err)
			}

			refutil.RemoveFromStructFieldArray(&sn.Data, inputName, index)
		} else {
			refutil.AddToStructFieldArray(&sn.Data, inputName, output.NodeOutput)
		}
		sn.inputChangedSinceLastProcess = true
		return
	}

	refutil.SetStructField(&sn.Data, input, output.NodeOutput)
	sn.inputChangedSinceLastProcess = true
}

func (sn Struct[T, G]) Outdated() bool {

	// Nil versions means we've never processed before
	if sn.depVersions == nil {
		return true
	}

	if sn.inputChangedSinceLastProcess {
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

func (sn *Struct[T, G]) updateUsedDependencyVersions() {
	deps := sn.Dependencies()
	sn.depVersions = make([]int, len(deps))
	for i, dep := range deps {
		sn.depVersions[i] = dep.Dependency().Version()
	}
}

type StructDependency struct {
	name           string
	dep            Node
	dependencyPort string
}

func (snd StructDependency) Name() string {
	return snd.name
}

func (snd StructDependency) Dependency() Node {
	return snd.dep
}

func (snd StructDependency) DependencyPort() string {
	return snd.dependencyPort
}

func (sn Struct[T, G]) Version() int {
	return sn.version
}

func (sn *Struct[T, G]) Out() StructOutput[T, G] {
	return StructOutput[T, G]{
		Struct: sn,
		Name:   "Out",
	}
}

func (sn *Struct[T, G]) Outputs() []Output {
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
			NodeOutput: StructOutput[T, G]{
				Name:   "Out",
				Struct: sn,
			},
		},
	}
}

func (sn Struct[T, G]) Inputs() []Input {
	nodeInputs := make([]Input, 0)

	refInput := refutil.GenericFieldValues("nodes.NodeOutput", sn.Data)
	for name, inputType := range refInput {
		nodeInputs = append(nodeInputs, Input{Name: name, Type: inputType})
	}

	refArrInput := refutil.GenericFieldValues("[]nodes.NodeOutput", sn.Data)
	for name, inputType := range refArrInput {
		nodeInputs = append(nodeInputs, Input{Name: name, Type: inputType, Array: true})
	}

	return nodeInputs
}

func (sn Struct[T, G]) Dependencies() []NodeDependency {
	output := make([]NodeDependency, 0)

	basicData := refutil.FieldValuesOfType[NodeOutputReference](sn.Data)
	for key, val := range basicData {
		output = append(output, StructDependency{
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

			output = append(output, StructDependency{
				name:           fmt.Sprintf("%s.%d", key, i),
				dep:            e.Node(),
				dependencyPort: e.Port(),
			})
		}
	}
	return output
}

func (sn *Struct[T, G]) Value() T {
	if sn.Outdated() {
		sn.process()
	}
	return sn.value
}

func (sn *Struct[T, G]) Node() Node {
	return sn
}

func (sn *Struct[T, G]) Port() string {
	return "Out"
}

func (sn *Struct[T, G]) process() {
	// tn.value, tn.err = tn.transform(tn.in)
	sn.value, sn.err = sn.Data.Process()
	sn.version++
	sn.updateUsedDependencyVersions()
	sn.inputChangedSinceLastProcess = false
}

func (sn Struct[T, G]) Name() string {
	return refutil.GetTypeNameWithoutPackage(sn.Data)
}

func (sn Struct[T, G]) Type() string {
	return refutil.GetTypeNameWithoutPackage(sn.Data)
}

func (sn Struct[T, G]) Path() string {
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

func (sn *Struct[T, G]) State() NodeState {
	if sn.Outdated() {
		return Stale
	}
	return Processed
}

func (sn *Struct[T, G]) AddSubscription(a Alertable) {
	if a == nil {
		panic("attempting to subscribe with nil alertable")
	}
	sn.subs = append(sn.subs, a)
}

func NewStruct[T StructProcesor[G], G any](p T) *Struct[G, T] {
	return &Struct[G, T]{
		Data:        p,
		version:     0,
		subs:        make([]Alertable, 0),
		depVersions: nil,
	}
}
