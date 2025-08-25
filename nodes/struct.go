package nodes

import (
	"fmt"
	"strings"
	"time"

	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/polyform/utils"
)

type inputVersions interface {
	inputVersions() string
}

// ============================================================================
type outputPortBuilder interface {
	build(node Node, cache *structOutputCache, data any, functionName, displayName string) OutputPort
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

	version := val.version
	if soc.Outdated(key) {
		version++
	}

	return version
}

func (soc structOutputCache) Outdated(key string) bool {
	val, ok := soc.cache[key]

	if !ok {
		return true
	}

	newVersion := soc.versioner.inputVersions()
	return val.nodeInputVersions != newVersion
}

func (soc structOutputCache) InputString() string {
	return soc.versioner.inputVersions()
}

func (soc *structOutputCache) Cache(key string, val any) {
	version := soc.Version(key)
	newVersion := version + 1

	// subtract one because we add one when we're outdated
	// So we're kinda just saying:
	// "heres the value to the version we where telling you about"
	if version > -1 && soc.Outdated(key) {
		newVersion--
	}

	soc.cache[key] = cachedStructOutput{
		nodeInputVersions: soc.versioner.inputVersions(),
		val:               val,
		version:           newVersion,
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
	displayName  string
	functionName string
	node         Node
	data         any
	val          T
	cache        *structOutputCache
	report       ExecutionReport
}

func (so StructOutput[T]) Name() string {
	return so.displayName
}

func (so StructOutput[T]) Node() Node {
	return so.node
}

func (so *StructOutput[T]) Value() T {
	var val StructOutput[T]
	if !so.cache.Outdated(so.functionName) {
		val = so.cache.Get(so.functionName).(StructOutput[T])
	} else {
		start := time.Now()
		refutil.CallStructMethod(so.data, so.functionName, &val)
		val.report.TotalTime = time.Since(start)
		self := val.report.TotalTime
		for _, v := range val.report.Steps {
			self -= v.Duration
		}
		val.report.SelfTime = &self
		// val.report.Errors = append(val.report.Errors, fmt.Sprintf("Version: %d", so.cache.Version(so.functionName)))
		// val.report.Errors = append(val.report.Errors, fmt.Sprintf("Input: %s", so.cache.InputString()))
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

func (so StructOutput[T]) Description() string {
	name := so.functionName + "Description"
	if refutil.HasMethod(so.data, name) {
		return refutil.CallStructMethod(so.data, name)[0].(string)
	}
	return ""
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// Implementing ObservableExecution
// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

// The report of executing the output function
func (so StructOutput[T]) ExecutionReport() ExecutionReport {

	// More song and dance of function return vs node return
	if !so.cache.Outdated(so.functionName) {
		val := so.cache.Get(so.functionName).(StructOutput[T])
		return val.report
	}
	return so.report
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// Methods called by the actual function that builds the thing
//
// We do this circus act where the StructOutput returned from the function
// isn't the StructOutput we pass around to other nodes to use.
// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

// Set the result of the output
func (so *StructOutput[T]) Set(v T) {
	so.val = v
}

func (so *StructOutput[T]) CaptureError(err error) {
	if err == nil {
		return
	}

	so.report.Errors = append(so.report.Errors, err.Error())
}

func (so *StructOutput[T]) CaptureTiming(title string, timing time.Duration) {
	so.report.Steps = append(so.report.Steps, StepTiming{
		Label:    title,
		Duration: timing,
	})
}

// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

func (so *StructOutput[T]) build(node Node, cache *structOutputCache, data any, functionName, displayName string) OutputPort {
	return &StructOutput[T]{
		node:         node,
		data:         data,
		functionName: functionName,
		displayName:  displayName,
		cache:        cache,
	}
}

// ============================================================================

type structArrayInput struct {
	node        Node
	data        dataProvider
	structField string
	displayName string
	datatype    string
}

func (si *structArrayInput) Clear() {
	refutil.SetStructField(si.data.Data(), si.structField, nil)
}

func (si structArrayInput) Node() Node {
	return si.node
}

func (si structArrayInput) Name() string {
	return si.displayName
}

func (so structArrayInput) Type() string {
	return so.datatype
}

func (si structArrayInput) Value() []OutputPort {
	return refutil.FieldValuesOfTypeInArray[OutputPort](si.data.Data())[si.structField]
	// return refutil.FieldValue[[]OutputPort](si.data.Data(), si.port)
}

func (si structArrayInput) Add(port OutputPort) error {
	refutil.AddToStructFieldArray(si.data.Data(), si.structField, port)
	return nil
}

func (si structArrayInput) Remove(port OutputPort) error {
	vals := refutil.FieldValuesOfTypeInArray[OutputPort](si.data.Data())[si.structField]
	for i, v := range vals {
		if v == port {
			refutil.RemoveFromStructFieldArray(si.data.Data(), si.structField, i)
			return nil
		}
	}
	return fmt.Errorf("array input port %s does not contain a reference to output port %s", si.Name(), port.Name())
}

func (si structArrayInput) Description() string {
	return refutil.GetStructTag(si.data.Data(), si.structField, "description")
}

// ============================================================================

type structInput struct {
	node        Node
	data        dataProvider
	structField string
	displayName string
	datatype    string
}

func (si *structInput) Clear() {
	refutil.SetStructField(si.data.Data(), si.structField, nil)
}

func (si structInput) Node() Node {
	return si.node
}

func (si structInput) Name() string {
	return si.displayName
}

func (so structInput) Type() string {
	return so.datatype
}

func (si structInput) Value() OutputPort {
	return refutil.FieldValue[OutputPort](si.data.Data(), si.structField)
}

func (si structInput) Set(port OutputPort) error {
	refutil.SetStructField(si.data.Data(), si.structField, port)
	return nil
}

func (si structInput) Description() string {
	return refutil.GetStructTag(si.data.Data(), si.structField, "description")
}

// ============================================================================

// Hack to give input ports the ability to modify the Data field on a Struct
// without the Data field being a pointer.
type structDataProvider[T any] struct {
	Node *Struct[T]
}

func (sdp structDataProvider[T]) Data() any {
	return &sdp.Node.Data
}

type dataProvider interface {
	Data() any
}

// ============================================================================

type Struct[T any] struct {
	Data T

	outputCache *structOutputCache
}

func (s *Struct[T]) Outputs() map[string]OutputPort {
	// if s.Data == nil {
	// 	var v T
	// 	s.Data = &v
	// }

	funcs := refutil.FuncArgumentsOfType[outputPortBuilder](s.Data)
	out := make(map[string]OutputPort)

	if s.outputCache == nil {
		s.outputCache = &structOutputCache{
			versioner: s,
			cache:     make(map[string]cachedStructOutput),
		}
	}

	for functionName, zero := range funcs {
		portName := utils.CamelCaseToSpaceCase(functionName)
		out[portName] = zero.build(s, s.outputCache, &s.Data, functionName, portName)
	}

	return out
}

func (s *Struct[T]) Inputs() map[string]InputPort {
	nodeInputs := make(map[string]InputPort)

	// if s.Data == nil {
	// 	var v T
	// 	s.Data = &v
	// }

	refInput := refutil.GenericFieldTypes("nodes.Output", s.Data)
	for name, dataType := range refInput {
		portName := utils.CamelCaseToSpaceCase(name)
		nodeInputs[portName] = &structInput{
			node:        s,
			data:        &structDataProvider[T]{Node: s},
			displayName: portName,
			structField: name,
			datatype:    dataType,
		}
	}

	refArrInput := refutil.GenericFieldTypes("[]nodes.Output", s.Data)
	for name, dataType := range refArrInput {
		// nodeInputs = append(nodeInputs, Input{Name: name, Type: inputType, Array: true})
		portName := utils.CamelCaseToSpaceCase(name)
		nodeInputs[portName] = &structArrayInput{
			node:        s,
			data:        &structDataProvider[T]{Node: s},
			structField: name,
			displayName: portName,
			datatype:    dataType,
		}
	}

	return nodeInputs
}

func (s *Struct[T]) inputVersions() string {
	builder := strings.Builder{}
	inputs := utils.SortMapByKey(s.Inputs())

	for _, input := range inputs {

		switch v := input.Val.(type) {
		case SingleValueInputPort:
			val := v.Value()
			if val != nil {
				builder.WriteString(fmt.Sprintf("%p:%d", val.Node(), val.Version()))
			} else {
				builder.WriteString("nil")
			}

		case ArrayValueInputPort:
			builder.WriteString("{")

			for _, val := range v.Value() {
				if val != nil {
					builder.WriteString(fmt.Sprintf("%p:%d", val.Node(), val.Version()))
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

func collapseCommonPackages(dirty string) string {
	result := dirty

	commonPackages := []string{
		"github.com/EliCDavis/vector/",
		"github.com/EliCDavis/polyform/drawing/",
	}

	for _, pack := range commonPackages {
		vectorStart := strings.Index(dirty, pack)
		if vectorStart == -1 {
			continue
		}
		return result[:vectorStart] + dirty[vectorStart+len(pack):]
	}

	return result
}

func (sn Struct[T]) Name() string {
	name := refutil.GetTypeNameWithoutPackage(sn.Data)

	genericType := ""
	startGeneric := strings.Index(name, "[")
	if startGeneric != -1 && name[len(name)-1:] == "]" {
		genericType = collapseCommonPackages(name[startGeneric:])

		name = name[0:startGeneric]
	}

	i := strings.LastIndex(name, "Node")
	if i != -1 && i == len(name)-8 {
		name = name[0 : len(name)-8]
	} else {
		i = strings.LastIndex(name, "Node")
		if i != -1 && i == len(name)-4 {
			name = name[0 : len(name)-4]
		}
	}

	return utils.CamelCaseToSpaceCase(name) + genericType
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
