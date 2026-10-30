package nodes

type ProxySource interface {
	Port
	Version() int
	Type() string

	// CurrentSource returns the port that values should be read from, or nil
	// if nothing is currently connected.
	CurrentSource() OutputPort
}

// Every typed output can build a proxy of its own type, so runtime-typed
// ports get a real Output[T] with no per-type registration.
type ProxyOutputBuilder interface {
	BuildProxyOutput(source ProxySource) OutputPort
}

func NewProxyOutput[T any](source ProxySource) OutputPort {
	return proxyOutput[T]{source: source}
}

type proxyOutput[T any] struct {
	source ProxySource
}

func (p proxyOutput[T]) Node() Node {
	return p.source.Node()
}

func (p proxyOutput[T]) Name() string {
	return p.source.Name()
}

func (p proxyOutput[T]) Version() int {
	return p.source.Version()
}

func (p proxyOutput[T]) Label() string {
	return labelOf(p.source)
}

func (p proxyOutput[T]) InstanceType() string {
	if instance, ok := p.source.(InstanceTyped); ok {
		return instance.InstanceType()
	}
	return ""
}

func labelOf(port Port) string {
	if labeled, ok := port.(Labeled); ok {
		return labeled.Label()
	}
	return port.Name()
}

func (p proxyOutput[T]) CurrentSource() OutputPort {
	return p.source.CurrentSource()
}

func (p proxyOutput[T]) Type() string {
	return p.source.Type()
}

// LazyProxySource has nothing behind it until a value is asked for, so
// CurrentSource can be nil where SourceForValue is not.
type LazyProxySource interface {
	SourceForValue() OutputPort
}

func (p proxyOutput[T]) Value() T {
	src := p.source.CurrentSource()
	if lazy, ok := p.source.(LazyProxySource); ok {
		src = lazy.SourceForValue()
	}
	if src != nil {
		if typed, ok := src.(Output[T]); ok {
			return typed.Value()
		}
	}
	var zero T
	return zero
}

func (p proxyOutput[T]) BuildProxyOutput(source ProxySource) OutputPort {
	return proxyOutput[T]{source: source}
}

func (p proxyOutput[T]) BuildDynamicOutput(source DynamicSource) OutputPort {
	return NewDynamicOutput[T](source)
}
