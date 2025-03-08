package nodes

import (
	"fmt"
	"strings"

	"github.com/EliCDavis/polyform/refutil"
)

// ============================================================================

// func (sn *Struct[T]) Outputs() []TypedPort {
// 	// outputs := refutil.FuncValuesOfType[ReferencesNode](tn.Data)

// 	// outs := make([]Output, len(outputs))
// 	// var v *T = new(T)
// 	// for i, o := range outputs {
// 	// 	outs[i] = Output{
// 	// 		Name: o,
// 	// 		// Type: fmt.Sprintf("%T", *new(T)),
// 	// 		Type: refutil.GetTypeWithPackage(v),
// 	// 	}
// 	// }
// 	// return outs

// 	// TODO: This is wrong for nodes with more than one output
// 	return []TypedPort{
// 		{
// 			Type: refutil.GetTypeWithPackage(new(T)),
// 			Port: StructOutput[T, G]{
// 				Name:   "Out",
// 				Struct: sn,
// 			},
// 		},
// 	}
// }

// func (sn Struct[T]) Inputs() []Input {
// 	nodeInputs := make([]Input, 0)

// 	refInput := refutil.GenericFieldTypes("nodes.NodeOutput", sn.Data)
// 	for name, inputType := range refInput {
// 		nodeInputs = append(nodeInputs, Input{Name: name, Type: inputType})
// 	}

// 	refArrInput := refutil.GenericFieldTypes("[]nodes.NodeOutput", sn.Data)
// 	for name, inputType := range refArrInput {
// 		nodeInputs = append(nodeInputs, Input{Name: name, Type: inputType, Array: true})
// 	}

// 	return nodeInputs
// }

// func (sn Struct[T]) Dependencies() []NodeDependency {
// 	output := make([]NodeDependency, 0)

// 	basicData := refutil.FieldValuesOfType[PortReference](sn.Data)
// 	for key, val := range basicData {
// 		output = append(output, StructDependency{
// 			name:           key,
// 			dep:            val.Node(),
// 			dependencyPort: val.Port(),
// 		})
// 	}

// 	arrayData := refutil.FieldValuesOfTypeInArray[PortReference](sn.Data)
// 	for key, field := range arrayData {
// 		for i, e := range field {
// 			if e == nil {
// 				continue
// 			}

// 			output = append(output, StructDependency{
// 				name:           fmt.Sprintf("%s.%d", key, i),
// 				dep:            e.Node(),
// 				dependencyPort: e.Port(),
// 			})
// 		}
// 	}
// 	return output
// }

// func (sn *Struct[T]) process() {
// 	// tn.value, tn.err = tn.transform(tn.in)
// 	sn.value, sn.err = sn.Data.Process()
// 	sn.version++
// 	sn.updateUsedDependencyVersions()
// 	sn.inputChangedSinceLastProcess = false
// }

// ============================================================================
type outputPortBuilder interface {
	build(node Node, data any, functionName string) OutputPort
}

func NewStructOutput[T any](val T) StructOutput[T] {
	return StructOutput[T]{
		val: val,
	}
}

type StructOutput[T any] struct {
	functionName string
	node         Node
	data         any
	val          T
	verion       int
}

func (so StructOutput[T]) LogError(err error) {
	if err == nil {
		return
	}

	// Do capture
}

func (so StructOutput[T]) Name() string {
	return so.functionName
}

func (so StructOutput[T]) Node() Node {
	return so.node
}

func (so StructOutput[T]) Value() T {
	return refutil.CallStructMethod(so.data, so.functionName)[0].(StructOutput[T]).val
}

func (so *StructOutput[T]) Set(v T) {
	so.val = v
}

func (so StructOutput[T]) Version() int {
	return so.verion
}

func (so StructOutput[T]) build(node Node, data any, functionName string) OutputPort {
	return StructOutput[T]{
		node:         node,
		data:         data,
		functionName: functionName,
	}
}

// ============================================================================

type structOutput[T any] struct {
	version int
	name    string
	node    Node
	data    any
}

func (si *structOutput[T]) Node() Node {
	return si.node
}

func (si *structOutput[T]) Name() string {
	return si.name
}

func (so *structOutput[T]) Value() T {
	vals := refutil.CallStructMethod(so.data, so.name, nil)

	if len(vals) != 1 {
		panic(fmt.Errorf("output function %s had %d return values", so.name, len(vals)))
	}

	return vals[0].(T)
}

func (so structOutput[T]) Version() int {
	return so.version
}

// ============================================================================

type structInput struct {
	node Node
	data any
	port string
}

func (si *structInput) Clear() {
	refutil.SetStructField(si.data, si.port, nil)
}

func (si structInput) Node() Node {
	return si.node
}

func (si structInput) Name() string {
	return si.port
}

func (si structInput) Value() OutputPort {
	return refutil.FieldValue[OutputPort](si.data, si.port)
}

func (si structInput) Set(port OutputPort) error {
	refutil.SetStructField(si.data, si.port, port)
	return nil
}

// ============================================================================

type Struct[T any] struct {
	Data T

	err                          error
	depVersions                  []int
	inputChangedSinceLastProcess bool

	version int
}

func (s *Struct[T]) Outputs() map[string]OutputPort {
	funcs := refutil.FuncValuesOfType[outputPortBuilder](s.Data)
	out := make(map[string]OutputPort)

	for functionName, zero := range funcs {
		out[functionName] = zero.build(s, s.Data, functionName)
	}

	return out
}

func (s *Struct[T]) Inputs() map[string]InputPort {
	nodeInputs := make(map[string]InputPort)

	refInput := refutil.GenericFieldTypes("node.Output", s.Data)
	for name := range refInput {
		nodeInputs[name] = &structInput{
			node: s,
			data: s.Data,
			port: name,
		}
	}

	return nodeInputs
}

func (sn Struct[T]) Name() string {
	return refutil.GetTypeNameWithoutPackage(sn.Data)
}

func (sn Struct[T]) Description() string {
	if described, ok := any(sn.Data).(Describable); ok {
		return described.Description()
	}
	return ""
}

func (sn Struct[T]) Type() string {
	return refutil.GetTypeNameWithoutPackage(sn.Data)
}

func (sn Struct[T]) Path() string {
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
