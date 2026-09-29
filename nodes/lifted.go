package nodes

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/EliCDavis/polyform/refutil"
)

// Takes either a T or an array of them, so a node written once against a
// single T works both ways. Only the count is decided at runtime, not the type.
type LiftedPort[T any] interface {
	OutputPort
}

const liftedPortFieldType = "nodes.LiftedPort"

const liftedPortArrayFieldType = "[]nodes.LiftedPort"

const liftedHandleType = "*nodes.Lifted"

// Never deeper than one: a lifted port's element type is already whatever
// the author asked for.
const (
	rankScalar = 0
	rankArray  = 1
)

func arrayOf(elem string) string {
	return "[]" + elem
}

// An array when any lifted input carries one. No rank is stored anywhere;
// it falls out of the connections.
func liftedRank(data any) int {
	for field, elem := range refutil.GenericFieldTypes(liftedPortFieldType, data) {
		port := refutil.FieldValue[OutputPort](data, field)
		if port == nil {
			continue
		}
		if portValueTypeName(port) == arrayOf(elem) {
			return rankArray
		}
	}

	arrays := refutil.GenericFieldTypes(liftedPortArrayFieldType, data)
	if len(arrays) == 0 {
		return rankScalar
	}

	wired := refutil.FieldValuesOfTypeInArray[OutputPort](data)
	for field, elem := range arrays {
		for _, port := range wired[field] {
			if port == nil {
				continue
			}
			if portValueTypeName(port) == arrayOf(elem) {
				return rankArray
			}
		}
	}
	return rankScalar
}

// ============================================================================

// Filled by Zip rather than by the author, so rank alignment is decided in
// one place instead of in every node.
type Lifted[T any] struct {
	rank   int
	scalar T
	array  []T
	report ExecutionReport
}

// 1 when carrying an array, 0 for a single value.
func (l *Lifted[T]) Rank() int { return l.rank }

func (l *Lifted[T]) CaptureError(err error) {
	if err == nil {
		return
	}
	l.report.Errors = append(l.report.Errors, err.Error())
}

func (l *Lifted[T]) CaptureTiming(title string, timing time.Duration) {
	l.report.Steps = append(l.report.Steps, StepTiming{Label: title, Duration: timing})
}

func (l *Lifted[T]) build(Node, *structOutputCache, any, string, string, *sync.Mutex) OutputPort {
	panic("a *nodes.Lifted output is built by Struct.Outputs, not through outputPortBuilder")
}

// The handle is its own port factory, so a lifted output never consults the
// type registry: its element type is known at compile time.
func (l *Lifted[T]) BuildDynamicOutput(source DynamicSource) OutputPort {
	return NewDynamicOutput[T](source)
}

func (l *Lifted[T]) BuildDynamicArrayOutput(source DynamicSource) OutputPort {
	return NewDynamicOutput[[]T](source)
}

func (l *Lifted[T]) attachRank(rank int) { l.rank = rank }

func (l *Lifted[T]) drain() liftedResult {
	if l.rank == rankArray {
		if l.array == nil {
			l.array = []T{}
		}
		return liftedResult{value: l.array, report: l.report}
	}
	return liftedResult{value: l.scalar, report: l.report}
}

// liftedHandle is every instantiation of Lifted, so Struct.Outputs can
// drive one without naming its type parameter.
type liftedHandle interface {
	attachRank(rank int)
	drain() liftedResult
}

type liftedResult struct {
	value  any
	report ExecutionReport
}

// ============================================================================

type liftedInput struct {
	node        Node
	data        dataProvider
	structField string
	displayName string
	elem        string
}

func (i *liftedInput) Node() Node   { return i.node }
func (i *liftedInput) Name() string { return i.displayName }

// Empty while nothing is connected, when either rank is still allowed.
func (i *liftedInput) Type() string {
	port := i.Value()
	if port == nil {
		return ""
	}
	return portValueTypeName(port)
}

// Both ranks, so an editor can refuse a wrong connection rather than error later.
func (i *liftedInput) AcceptedTypes() []string {
	return []string{i.elem, arrayOf(i.elem)}
}

func (i *liftedInput) Description() string {
	return refutil.GetStructTag(i.data.Data(), i.structField, "description")
}

func (i *liftedInput) Value() OutputPort {
	return refutil.FieldValue[OutputPort](i.data.Data(), i.structField)
}

func (i *liftedInput) Clear() {
	refutil.SetStructField(i.data.Data(), i.structField, nil)
	Touch()
}

func (i *liftedInput) Set(port OutputPort) error {
	if port != nil {
		concrete := portValueTypeName(port)
		if concrete != "" && concrete != i.elem && concrete != arrayOf(i.elem) {
			return fmt.Errorf("input %q takes %s or %s, not %s", i.displayName, i.elem, arrayOf(i.elem), concrete)
		}
		rememberPortType(concrete, port)
	}

	refutil.SetStructField(i.data.Data(), i.structField, port)
	Touch()
	return nil
}

// ============================================================================

type liftedArrayInput struct {
	node        Node
	data        dataProvider
	structField string
	displayName string
	elem        string
}

func (i *liftedArrayInput) Node() Node   { return i.node }
func (i *liftedArrayInput) Name() string { return i.displayName }

// One array anywhere in the input makes the whole node's output an array.
func (i *liftedArrayInput) Type() string {
	found := ""
	for _, port := range i.Value() {
		if port == nil {
			continue
		}
		name := portValueTypeName(port)
		if name == arrayOf(i.elem) {
			return name
		}
		if name != "" {
			found = name
		}
	}
	return found
}

func (i *liftedArrayInput) AcceptedTypes() []string {
	return []string{i.elem, arrayOf(i.elem)}
}

func (i *liftedArrayInput) Description() string {
	return refutil.GetStructTag(i.data.Data(), i.structField, "description")
}

func (i *liftedArrayInput) Value() []OutputPort {
	return refutil.FieldValuesOfTypeInArray[OutputPort](i.data.Data())[i.structField]
}

func (i *liftedArrayInput) Clear() {
	refutil.SetStructField(i.data.Data(), i.structField, nil)
	Touch()
}

func (i *liftedArrayInput) accepts(port OutputPort) error {
	if port == nil {
		return nil
	}
	concrete := portValueTypeName(port)
	if concrete != "" && concrete != i.elem && concrete != arrayOf(i.elem) {
		return fmt.Errorf("input %q takes %s or %s, not %s", i.displayName, i.elem, arrayOf(i.elem), concrete)
	}
	rememberPortType(concrete, port)
	return nil
}

func (i *liftedArrayInput) Add(port OutputPort) error {
	if err := i.accepts(port); err != nil {
		return err
	}
	refutil.AddToStructFieldArray(i.data.Data(), i.structField, port)
	Touch()
	return nil
}

func (i *liftedArrayInput) Remove(port OutputPort) error {
	for index, v := range i.Value() {
		if v == port {
			refutil.RemoveFromStructFieldArray(i.data.Data(), i.structField, index)
			Touch()
			return nil
		}
	}
	return fmt.Errorf("array input port %s does not contain a reference to output port %s", i.Name(), port.Name())
}

func (i *liftedArrayInput) Replace(index int, port OutputPort) error {
	count := len(i.Value())
	if index < 0 || index >= count {
		return fmt.Errorf("array input port %s has %d element(s), so there is no index %d to replace", i.Name(), count, index)
	}
	if err := i.accepts(port); err != nil {
		return err
	}
	refutil.SetStructFieldArrayElement(i.data.Data(), i.structField, index, port)
	Touch()
	return nil
}

// ============================================================================

type liftedOutput struct {
	node         Node
	data         any
	functionName string
	displayName  string
	elem         string
	handleType   reflect.Type
	cache        *structOutputCache
	mutex        *sync.Mutex
}

func (o *liftedOutput) Node() Node   { return o.node }
func (o *liftedOutput) Name() string { return o.displayName }

func (o *liftedOutput) Type() string {
	if liftedRank(o.data) == rankArray {
		return arrayOf(o.elem)
	}
	return o.elem
}

func (o *liftedOutput) Description() string {
	name := o.functionName + "Description"
	if refutil.HasMethod(o.data, name) {
		return refutil.CallStructMethod(o.data, name)[0].(string)
	}
	return ""
}

func (o *liftedOutput) Version() int {
	return o.cache.Version(o.functionName)
}

func (o *liftedOutput) CurrentSource() OutputPort { return nil }

func (o *liftedOutput) AnyValue() any {
	return o.evaluate().value
}

func (o *liftedOutput) ExecutionReport() ExecutionReport {
	return o.evaluate().report
}

func (o *liftedOutput) evaluate() liftedResult {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !o.cache.Outdated(o.functionName) {
		if cached, ok := o.cache.Get(o.functionName).(liftedResult); ok {
			return cached
		}
	}

	handleValue := reflect.New(o.handleType.Elem())
	handle := handleValue.Interface().(liftedHandle)
	handle.attachRank(liftedRank(o.data))

	start := time.Now()
	o.call(handleValue.Interface())

	result := handle.drain()
	result.report.TotalTime = time.Since(start)
	self := result.report.TotalTime
	for _, step := range result.report.Steps {
		self -= step.Duration
	}
	result.report.SelfTime = &self

	o.cache.Cache(o.functionName, result)
	return result
}

func (o *liftedOutput) call(handle any) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		resolver := refutil.TypeResolution{IncludePackage: true, StripSinglePointer: true}
		panic(fmt.Errorf("%s.%s: %v", resolver.Resolve(o.data), o.functionName, r))
	}()
	refutil.CallStructMethod(o.data, o.functionName, handle)
}

// Wraps the output in a real Output[T] for whichever rank the inputs settled
// on, so downstream nodes can hold it in their own typed fields.
func (o *liftedOutput) buildPort() OutputPort {
	builder := reflect.Zero(o.handleType).Interface()

	if liftedRank(o.data) == rankArray {
		if array, ok := builder.(DynamicArrayOutputBuilder); ok {
			return array.BuildDynamicArrayOutput(o)
		}
		return o
	}

	if single, ok := builder.(DynamicOutputBuilder); ok {
		return single.BuildDynamicOutput(o)
	}
	return o
}

// patternOfLiftedHandle reads the element type off an output method's
// argument, which reflection renders as "*nodes.Lifted[pkg.Type]".
func patternOfLiftedHandle(handle reflect.Type) (string, bool) {
	name := handle.String()
	inner, ok := cutBetween(name, liftedHandleType+"[", "]")
	if !ok {
		return "", false
	}
	return inner, true
}

func cutBetween(s, prefix, suffix string) (string, bool) {
	if len(s) < len(prefix)+len(suffix) {
		return "", false
	}
	if s[:len(prefix)] != prefix || s[len(s)-len(suffix):] != suffix {
		return "", false
	}
	return s[len(prefix) : len(s)-len(suffix)], true
}

type TypeOptions interface {
	AcceptedTypes() []string
}

type InstanceTyped interface {
	InstanceType() string
}

func (o *liftedOutput) InstanceType() string { return o.Type() }
