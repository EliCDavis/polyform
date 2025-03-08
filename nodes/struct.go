package nodes

import (
	"strconv"
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

type inputVersions interface {
	inputVersions() string
}

// ============================================================================
type outputPortBuilder interface {
	build(node Node, cache *structOutputCache, data any, functionName string) OutputPort
}

func NewStructOutput[T any](val T) StructOutput[T] {
	return StructOutput[T]{
		val: val,
	}
}

type structOutputCache struct {
	versioner inputVersions
	cache     map[string]cachedStructOutput
}

func (soc structOutputCache) Version(key string) int {
	val, ok := soc.cache[key]

	if !ok {
		return -1
	}

	return val.version

}

func (soc structOutputCache) Outdated(key string) bool {
	val, ok := soc.cache[key]

	if !ok {
		return true
	}

	return val.nodeInputVersions != soc.versioner.inputVersions()
}

func (soc structOutputCache) Cache(key string, val any) {
	soc.cache[key] = cachedStructOutput{
		nodeInputVersions: soc.versioner.inputVersions(),
		val:               val,
		version:           soc.Version(key) + 1,
	}
}

func (soc structOutputCache) Get(key string) any {
	return soc.cache[key].val
}

type cachedStructOutput struct {
	nodeInputVersions string
	version           int
	val               any
}

type StructOutput[T any] struct {
	functionName string
	node         Node
	data         any
	val          T
	cache        *structOutputCache
}

func (so StructOutput[T]) Name() string {
	return so.functionName
}

func (so StructOutput[T]) Node() Node {
	return so.node
}

func (so *StructOutput[T]) Value() T {
	var val StructOutput[T]
	if !so.cache.Outdated(so.functionName) {
		val = so.cache.Get(so.functionName).(StructOutput[T])
	} else {
		val = refutil.CallStructMethod(so.data, so.functionName)[0].(StructOutput[T])
		so.cache.Cache(so.functionName, val)
	}
	return val.val
}

func (so StructOutput[T]) Version() int {
	return so.cache.Version(so.functionName)
}

func (so StructOutput[T]) Type() string {
	return refutil.GetTypeWithPackage(new(T))
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// Methods called by the actual function that builds the thing
//
// We do this circus act where the StructOutput returned from the function
// isn't the StructOutput we pass around to other nodes to use.
// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
func (so *StructOutput[T]) Set(v T) {
	so.val = v
}

func (so StructOutput[T]) LogError(err error) {
	if err == nil {
		return
	}

	// Do capture
}

// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

func (so StructOutput[T]) build(node Node, cache *structOutputCache, data any, functionName string) OutputPort {
	return &StructOutput[T]{
		node:         node,
		data:         data,
		functionName: functionName,
		cache:        cache,
	}
}

// ============================================================================

type structInput struct {
	node     Node
	data     any
	port     string
	datatype string
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

func (so structInput) Type() string {
	return so.datatype
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

	outputCache *structOutputCache
}

func (s *Struct[T]) Outputs() map[string]OutputPort {
	funcs := refutil.FuncValuesOfType[outputPortBuilder](s.Data)
	out := make(map[string]OutputPort)

	if s.outputCache == nil {
		s.outputCache = &structOutputCache{
			versioner: s,
			cache:     make(map[string]cachedStructOutput),
		}
	}

	for functionName, zero := range funcs {
		out[functionName] = zero.build(s, s.outputCache, s.Data, functionName)
	}

	return out
}

func (s *Struct[T]) Inputs() map[string]InputPort {
	nodeInputs := make(map[string]InputPort)

	refInput := refutil.GenericFieldTypes("nodes.Output", s.Data)
	for name, dataType := range refInput {
		nodeInputs[name] = &structInput{
			node:     s,
			data:     &s.Data,
			port:     name,
			datatype: dataType,
		}
	}

	// refArrInput := refutil.GenericFieldTypes("[]nodes.NodeOutput", s.Data)
	// for name, inputType := range refArrInput {
	// 	nodeInputs = append(nodeInputs, Input{Name: name, Type: inputType, Array: true})
	// }

	return nodeInputs
}

func (s *Struct[T]) inputVersions() string {
	builder := strings.Builder{}

	inputs := s.Inputs()
	for _, input := range inputs {

		switch v := input.(type) {
		case SingleValueInputPort:
			val := v.Value()
			if val != nil {
				builder.WriteString(strconv.Itoa(val.Version()))
			} else {
				builder.WriteString("nil")
			}

		case ArrayValueInputPort:
			builder.WriteString("{")

			for _, val := range v.Value() {
				if val != nil {
					builder.WriteString(strconv.Itoa(val.Version()))
					builder.WriteString(",")
				} else {
					builder.WriteString("nil,")
				}
			}

			builder.WriteString("}")
		default:
		}

		builder.WriteString(";")
	}

	return builder.String()
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
