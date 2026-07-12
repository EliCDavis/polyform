package nodes

type ProxySource interface {
	Port
	Version() int
	Type() string

	// CurrentSource returns the port that values should be read from, or nil
	// if nothing is currently connected.
	CurrentSource() OutputPort
}

// ProxyOutputBuilder is implemented by typed output ports that can
// manufacture a proxy output port carrying the same value type. Because the
// implementations are generic, every value type produced by any registered
// node automatically has a proxy implementation compiled into the binary,
// which lets runtime-typed ports (like sub-graph boundaries) expose a
// strongly typed nodes.Output[T] without per-type registration.
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

func (p proxyOutput[T]) Type() string {
	return p.source.Type()
}

func (p proxyOutput[T]) Value() T {
	if src := p.source.CurrentSource(); src != nil {
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
