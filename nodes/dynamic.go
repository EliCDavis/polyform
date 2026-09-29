package nodes

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/EliCDavis/polyform/refutil"
)

// A node declares one per independent type it carries: type SelectType nodes.DynamicType
type DynamicType any

// Value type decided at runtime, so the node is registered once for every
// type. Ports sharing a variable settle on one type together.
type DynamicPort[V any] interface {
	OutputPort
}

// How a dynamic port's field reads back from reflection, which is what
// Struct.Inputs matches on to tell it apart from a nodes.Output[T] field.
const (
	dynamicPortFieldType      = "nodes.DynamicPort"
	dynamicPortArrayFieldType = "[]nodes.DynamicPort"
	dynamicHandleType         = "*nodes.Dynamic"
)

// dynamicPattern is the type argument a port was declared with: the type
// variable itself, or some number of slices of it.
type dynamicPattern string

func (p dynamicPattern) variable() string {
	s := string(p)
	for strings.HasPrefix(s, "[]") {
		s = s[2:]
	}
	return s
}

func (p dynamicPattern) prefix() string {
	return string(p)[:len(p)-len(p.variable())]
}

// resolve turns the pattern into a concrete type string once the variable
// is known.
func (p dynamicPattern) resolve(bound string) string {
	if bound == "" {
		return ""
	}
	return p.prefix() + bound
}

// solve reads the variable back out of a concrete type a port was connected
// to.
func (p dynamicPattern) solve(concrete string) (string, bool) {
	prefix := p.prefix()
	if concrete == "" || !strings.HasPrefix(concrete, prefix) {
		return "", false
	}
	return concrete[len(prefix):], true
}

// DynamicVariable is the type variable a port's pattern is written in terms
// of, which is the key its binding appears under in DynamicTypes.
func DynamicVariable(pattern string) string {
	return dynamicPattern(pattern).variable()
}

// patternOfHandle reads the pattern off an output method's argument, which
// reflection renders as "*nodes.Dynamic[pkg.Variable]".
func patternOfHandle(handle reflect.Type) (dynamicPattern, bool) {
	name := handle.String()
	inner, ok := strings.CutPrefix(name, dynamicHandleType+"[")
	if !ok {
		return "", false
	}
	inner, ok = strings.CutSuffix(inner, "]")
	if !ok {
		return "", false
	}
	return dynamicPattern(inner), true
}

// ============================================================================

// The first connection to any port using a variable sets it, and every other
// port on that variable reports and enforces the same type.
type dynamicBindings struct {
	mu    sync.RWMutex
	bound map[string]string
}

func (b *dynamicBindings) boundFor(variable string) string {
	if b == nil {
		return ""
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.bound[variable]
}

func (b *dynamicBindings) bind(variable, to string) error {
	if to == "" {
		return nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if existing := b.bound[variable]; existing != "" {
		if existing == to {
			return nil
		}
		return fmt.Errorf("this node is already carrying %s; disconnect the ports using it before wiring %s in", existing, to)
	}
	if b.bound == nil {
		b.bound = make(map[string]string)
	}
	b.bound[variable] = to
	Touch()
	return nil
}

func (b *dynamicBindings) unbind(variable string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.bound[variable]; !ok {
		return
	}
	delete(b.bound, variable)
	Touch()
}

func (b *dynamicBindings) all() map[string]string {
	if b == nil {
		return nil
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	if len(b.bound) == 0 {
		return nil
	}
	out := make(map[string]string, len(b.bound))
	for variable, concrete := range b.bound {
		out[variable] = concrete
	}
	return out
}

// Keyed by the type variable each one belongs to.
type DynamicallyTyped interface {
	DynamicTypes() map[string]string
}

type DynamicallyTypedPort interface {
	// The port's type written in terms of its variable, which is what to
	// show while nothing is connected.
	DynamicPattern() string

	// The graph offers the peer's type before connecting, so an unbound
	// port can take its type from either end of the wire.
	BindType(concrete string) error

	// Released once the last connection holding the type goes, so a node
	// wired up wrong can be rewired rather than thrown away.
	ReleaseType()
}

// ============================================================================

// The counterpart to StructOutput[T]; its type parameter names the variable
// the output carries.
type Dynamic[V any] struct {
	typ       string
	forwarded OutputPort
	value     any
	hasValue  bool
	report    ExecutionReport
}

// The value is never copied or boxed, so forwarding costs nothing whatever
// is flowing through.
func (d *Dynamic[V]) Forward(port OutputPort) {
	d.forwarded = port
	d.value = nil
	d.hasValue = false
}

// Must be the node's bound type; anything else reads as that type's zero
// downstream.
func (d *Dynamic[V]) Set(value any) {
	d.value = value
	d.hasValue = true
	d.forwarded = nil
}

// Type is the concrete type this output settled on, empty while unbound.
func (d *Dynamic[V]) Type() string {
	return d.typ
}

func (d *Dynamic[V]) CaptureError(err error) {
	if err == nil {
		return
	}
	d.report.Errors = append(d.report.Errors, err.Error())
}

func (d *Dynamic[V]) CaptureTiming(title string, timing time.Duration) {
	d.report.Steps = append(d.report.Steps, StepTiming{Label: title, Duration: timing})
}

func (d *Dynamic[V]) build(Node, *structOutputCache, any, string, string, *sync.Mutex) OutputPort {
	panic("a *nodes.Dynamic output is built by Struct.Outputs, not through outputPortBuilder")
}

func (d *Dynamic[V]) attachType(typ string) { d.typ = typ }

func (d *Dynamic[V]) drain() dynamicResult {
	return dynamicResult{forwarded: d.forwarded, value: d.value, report: d.report}
}

// dynamicHandle is every instantiation of Dynamic, so Struct.Outputs can
// recognise and drive one without naming its type parameter.
type dynamicHandle interface {
	attachType(typ string)
	drain() dynamicResult
}

// ============================================================================

type dynamicResult struct {
	forwarded OutputPort
	value     any
	report    ExecutionReport
}

// DynamicSource is what a runtime-typed output port reads from: either a
// port to pass through, or a computed value.
type DynamicSource interface {
	Port
	Version() int
	Type() string

	// CurrentSource is the port being passed through, nil when the node
	// computed a value instead.
	CurrentSource() OutputPort

	AnyValue() any
}

// Every typed output port is a factory for a runtime-typed port of the same
// value type, so every type a node produces is covered with no registration.
type DynamicOutputBuilder interface {
	BuildDynamicOutput(source DynamicSource) OutputPort
}

// The port it returns deliberately cannot build a further array of itself:
// a method returning the slice of its own parameter is an instantiation cycle.
type DynamicArrayOutputBuilder interface {
	BuildDynamicArrayOutput(source DynamicSource) OutputPort
}

func NewDynamicOutput[T any](source DynamicSource) OutputPort {
	return dynamicTyped[T]{source: source}
}

type dynamicTyped[T any] struct {
	source DynamicSource
}

func (d dynamicTyped[T]) Node() Node                { return d.source.Node() }
func (d dynamicTyped[T]) Name() string              { return d.source.Name() }
func (d dynamicTyped[T]) Version() int              { return d.source.Version() }
func (d dynamicTyped[T]) Type() string              { return d.source.Type() }
func (d dynamicTyped[T]) CurrentSource() OutputPort { return d.source.CurrentSource() }
func (d dynamicTyped[T]) BuildProxyOutput(s ProxySource) OutputPort {
	return proxyOutput[T]{source: s}
}
func (d dynamicTyped[T]) BuildDynamicOutput(s DynamicSource) OutputPort {
	return dynamicTyped[T]{source: s}
}

func (d dynamicTyped[T]) Value() T {
	if src := d.source.CurrentSource(); src != nil {
		if typed, ok := src.(Output[T]); ok {
			return typed.Value()
		}
		var zero T
		return zero
	}
	if v, ok := d.source.AnyValue().(T); ok {
		return v
	}
	var zero T
	return zero
}

func (d dynamicTyped[T]) ExecutionReport() ExecutionReport {
	if observable, ok := d.source.(ObservableExecution); ok {
		return observable.ExecutionReport()
	}
	return ExecutionReport{}
}

// ============================================================================

type dynamicOutput struct {
	node         Node
	data         any
	functionName string
	displayName  string
	pattern      dynamicPattern
	handleType   reflect.Type
	cache        *structOutputCache
	mutex        *sync.Mutex
	bindings     *dynamicBindings
}

func (d *dynamicOutput) Node() Node   { return d.node }
func (d *dynamicOutput) Name() string { return d.displayName }

func (d *dynamicOutput) Description() string {
	name := d.functionName + "Description"
	if refutil.HasMethod(d.data, name) {
		return refutil.CallStructMethod(d.data, name)[0].(string)
	}
	return ""
}

func (d *dynamicOutput) Type() string {
	return d.pattern.resolve(d.bindings.boundFor(d.pattern.variable()))
}

func (d *dynamicOutput) Version() int {
	return d.cache.Version(d.functionName)
}

func (d *dynamicOutput) CurrentSource() OutputPort {
	return d.evaluate().forwarded
}

func (d *dynamicOutput) AnyValue() any {
	return d.evaluate().value
}

func (d *dynamicOutput) ExecutionReport() ExecutionReport {
	return d.evaluate().report
}

func (d *dynamicOutput) BindType(concrete string) error {
	bound, ok := d.pattern.solve(concrete)
	if !ok {
		return fmt.Errorf("output %q carries %s, which %s cannot be read out of", d.displayName, d.pattern, concrete)
	}
	return d.bindings.bind(d.pattern.variable(), bound)
}

func (d *dynamicOutput) ReleaseType() {
	d.bindings.unbind(d.pattern.variable())
}

func (d *dynamicOutput) evaluate() dynamicResult {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if !d.cache.Outdated(d.functionName) {
		if cached, ok := d.cache.Get(d.functionName).(dynamicResult); ok {
			return cached
		}
	}

	handleValue := reflect.New(d.handleType.Elem())
	handle := handleValue.Interface().(dynamicHandle)
	handle.attachType(d.Type())

	start := time.Now()
	d.call(handleValue.Interface())

	result := handle.drain()
	result.report.TotalTime = time.Since(start)
	self := result.report.TotalTime
	for _, step := range result.report.Steps {
		self -= step.Duration
	}
	result.report.SelfTime = &self

	d.cache.Cache(d.functionName, result)
	return result
}

func (d *dynamicOutput) call(handle any) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		resolver := refutil.TypeResolution{IncludePackage: true, StripSinglePointer: true}
		panic(fmt.Errorf("%s.%s: %v", resolver.Resolve(d.data), d.functionName, r))
	}()
	refutil.CallStructMethod(d.data, d.functionName, handle)
}

// Wraps the source in a real Output[T] for the bound type. Unbound it stays
// untyped, which is what lets anything connect to it.
func (d *dynamicOutput) buildPort() OutputPort {
	concrete := d.Type()
	if concrete == "" {
		return d
	}
	if builder, ok := LookupPortTypeDynamic(concrete); ok {
		return builder.BuildDynamicOutput(d)
	}
	if elem, isArray := strings.CutPrefix(concrete, "[]"); isArray {
		if builder, ok := LookupPortTypeDynamicArray(elem); ok {
			return builder.BuildDynamicArrayOutput(d)
		}
	}
	return d
}

// ============================================================================

type dynamicInput struct {
	node        Node
	data        dataProvider
	structField string
	displayName string
	pattern     dynamicPattern
	bindings    *dynamicBindings
}

func (i *dynamicInput) Node() Node   { return i.node }
func (i *dynamicInput) Name() string { return i.displayName }

func (i *dynamicInput) Type() string {
	return i.pattern.resolve(i.bindings.boundFor(i.pattern.variable()))
}

func (i *dynamicInput) Description() string {
	return refutil.GetStructTag(i.data.Data(), i.structField, "description")
}

func (i *dynamicInput) Value() OutputPort {
	return refutil.FieldValue[OutputPort](i.data.Data(), i.structField)
}

func (i *dynamicInput) Clear() {
	refutil.SetStructField(i.data.Data(), i.structField, nil)
	Touch()
}

func (i *dynamicInput) Set(port OutputPort) error {
	if err := bindPortToVariable(i.pattern, i.bindings, port, i.displayName); err != nil {
		return err
	}
	refutil.SetStructField(i.data.Data(), i.structField, port)
	Touch()
	return nil
}

func (i *dynamicInput) BindType(concrete string) error {
	bound, ok := i.pattern.solve(concrete)
	if !ok {
		return fmt.Errorf("input %q carries %s, which %s cannot be read out of", i.displayName, i.pattern, concrete)
	}
	return i.bindings.bind(i.pattern.variable(), bound)
}

func (i *dynamicInput) ReleaseType() {
	i.bindings.unbind(i.pattern.variable())
}

// ============================================================================

type dynamicArrayInput struct {
	node        Node
	data        dataProvider
	structField string
	displayName string
	pattern     dynamicPattern
	bindings    *dynamicBindings
}

func (i *dynamicArrayInput) Node() Node   { return i.node }
func (i *dynamicArrayInput) Name() string { return i.displayName }

func (i *dynamicArrayInput) Type() string {
	return i.pattern.resolve(i.bindings.boundFor(i.pattern.variable()))
}

func (i *dynamicArrayInput) Description() string {
	return refutil.GetStructTag(i.data.Data(), i.structField, "description")
}

func (i *dynamicArrayInput) Value() []OutputPort {
	return refutil.FieldValuesOfTypeInArray[OutputPort](i.data.Data())[i.structField]
}

func (i *dynamicArrayInput) Clear() {
	refutil.SetStructField(i.data.Data(), i.structField, nil)
	Touch()
}

func (i *dynamicArrayInput) Add(port OutputPort) error {
	if err := bindPortToVariable(i.pattern, i.bindings, port, i.displayName); err != nil {
		return err
	}
	refutil.AddToStructFieldArray(i.data.Data(), i.structField, port)
	Touch()
	return nil
}

func (i *dynamicArrayInput) Remove(port OutputPort) error {
	for index, existing := range i.Value() {
		if existing == port {
			refutil.RemoveFromStructFieldArray(i.data.Data(), i.structField, index)
			Touch()
			return nil
		}
	}
	return fmt.Errorf("array input port %s does not contain a reference to output port %s", i.displayName, port.Name())
}

func (i *dynamicArrayInput) Replace(index int, port OutputPort) error {
	count := len(i.Value())
	if index < 0 || index >= count {
		return fmt.Errorf("array input port %s has %d element(s), so there is no index %d to replace", i.displayName, count, index)
	}
	if err := bindPortToVariable(i.pattern, i.bindings, port, i.displayName); err != nil {
		return err
	}
	refutil.SetStructFieldArrayElement(i.data.Data(), i.structField, index, port)
	Touch()
	return nil
}

func (i *dynamicArrayInput) BindType(concrete string) error {
	bound, ok := i.pattern.solve(concrete)
	if !ok {
		return fmt.Errorf("input %q carries %s, which %s cannot be read out of", i.displayName, i.pattern, concrete)
	}
	return i.bindings.bind(i.pattern.variable(), bound)
}

func (i *dynamicArrayInput) ReleaseType() {
	i.bindings.unbind(i.pattern.variable())
}

// ============================================================================

func bindPortToVariable(pattern dynamicPattern, bindings *dynamicBindings, port OutputPort, portName string) error {
	if port == nil {
		return nil
	}

	concrete := portValueTypeName(port)
	if concrete == "" {
		return nil
	}

	// Learned while a real port of this type is in hand, so a port derived
	// from it later can be built even if no registered node produces one.
	rememberPortType(concrete, port)

	bound, ok := pattern.solve(concrete)
	if !ok {
		return fmt.Errorf("port %q carries %s, which %s cannot be read out of", portName, pattern, concrete)
	}
	return bindings.bind(pattern.variable(), bound)
}

func (d *dynamicOutput) DynamicPattern() string     { return string(d.pattern) }
func (i *dynamicInput) DynamicPattern() string      { return string(i.pattern) }
func (i *dynamicArrayInput) DynamicPattern() string { return string(i.pattern) }

func (d dynamicTyped[T]) DynamicPattern() string {
	if pattern, ok := d.source.(DynamicallyTypedPort); ok {
		return pattern.DynamicPattern()
	}
	return ""
}

func (d dynamicTyped[T]) Description() string {
	if described, ok := d.source.(Describable); ok {
		return described.Description()
	}
	return ""
}

func (d dynamicTyped[T]) InstanceType() string {
	if instance, ok := d.source.(InstanceTyped); ok {
		return instance.InstanceType()
	}
	return ""
}

func (d dynamicTyped[T]) BindType(concrete string) error {
	if bindable, ok := d.source.(DynamicallyTypedPort); ok {
		return bindable.BindType(concrete)
	}
	return nil
}

func (d dynamicTyped[T]) ReleaseType() {
	if releasable, ok := d.source.(DynamicallyTypedPort); ok {
		releasable.ReleaseType()
	}
}
