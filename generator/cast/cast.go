package cast

import (
	"cmp"
	"errors"
	"slices"
	"sync"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

type Pair struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Table struct {
	casts map[Pair]func(nodes.OutputPort) nodes.OutputPort
}

// Also registers []From to []To, converting element by element.
func Register[From, To any](table *Table, convert func(From) To) {
	RegisterFallible(table, func(v From) (To, error) { return convert(v), nil })
}

// The error is reported by the node reading the cast.
func RegisterFallible[From, To any](table *Table, convert func(From) (To, error)) {
	if table.casts == nil {
		table.casts = make(map[Pair]func(nodes.OutputPort) nodes.OutputPort)
	}

	table.casts[Pair{From: typeName[From](), To: typeName[To]()}] = func(src nodes.OutputPort) nodes.OutputPort {
		return &output[From, To]{src: src, convert: convert}
	}

	convertAll := func(in []From) ([]To, error) {
		out := make([]To, len(in))
		errs := make([]error, len(in))
		for i, v := range in {
			out[i], errs[i] = convert(v)
		}
		return out, errors.Join(errs...)
	}
	table.casts[Pair{From: typeName[[]From](), To: typeName[[]To]()}] = func(src nodes.OutputPort) nodes.OutputPort {
		return &output[[]From, []To]{src: src, convert: convertAll}
	}
}

func (t *Table) Output(port nodes.OutputPort, to string) (nodes.OutputPort, bool) {
	if t == nil {
		return nil, false
	}
	typed, ok := port.(nodes.Typed)
	if !ok {
		return nil, false
	}
	wrap, ok := t.casts[Pair{From: typed.Type(), To: to}]
	if !ok {
		return nil, false
	}
	return wrap(port), true
}

func (t *Table) Pairs() []Pair {
	pairs := make([]Pair, 0)
	if t == nil {
		return pairs
	}
	for pair := range t.casts {
		pairs = append(pairs, pair)
	}
	slices.SortFunc(pairs, func(a, b Pair) int {
		return cmp.Or(cmp.Compare(a.From, b.From), cmp.Compare(a.To, b.To))
	})
	return pairs
}

func typeName[T any]() string {
	resolver := refutil.TypeResolution{
		IncludePackage:     true,
		IncludePointer:     true,
		StripSinglePointer: true,
	}
	return resolver.Resolve(new(T))
}

// Everything but Type and Value must pass through: wiring, caching and timing identify the producer's port by them.
type output[From, To any] struct {
	src     nodes.OutputPort
	convert func(From) (To, error)

	mu      sync.Mutex
	cached  bool
	version int
	value   To
	err     error
}

func (c *output[From, To]) Node() nodes.Node { return c.src.Node() }
func (c *output[From, To]) Name() string     { return c.src.Name() }
func (c *output[From, To]) Version() int     { return c.src.Version() }
func (c *output[From, To]) Type() string     { return typeName[To]() }

func (c *output[From, To]) Label() string {
	if labeled, ok := c.src.(nodes.Labeled); ok {
		return labeled.Label()
	}
	return c.src.Name()
}

func (c *output[From, To]) Value() To {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cached && c.version == c.src.Version() {
		return c.value
	}
	c.value, c.err = c.convert(c.read())
	c.version = c.src.Version()
	c.cached = c.version >= 0
	return c.value
}

func (c *output[From, To]) Err() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.err
}

func (c *output[From, To]) read() From {
	if typed, ok := c.src.(nodes.Output[From]); ok {
		return typed.Value()
	}
	for _, value := range nodes.DynamicAnyValues(nil, c.src) {
		if v, ok := value.(From); ok {
			return v
		}
	}
	var zero From
	return zero
}

func (c *output[From, To]) Spreads() bool { return nodes.IsSpread(c.src) }

func (c *output[From, To]) SpreadValues() ([]To, bool) {
	values, ok := nodes.SpreadValues[From](c.src)
	if !ok {
		return nil, false
	}
	out := make([]To, len(values))
	errs := make([]error, len(values))
	for i, v := range values {
		out[i], errs[i] = c.convert(v)
	}
	c.mu.Lock()
	c.err = errors.Join(errs...)
	c.mu.Unlock()
	return out, true
}

func (c *output[From, To]) SpreadAny() any {
	values, _ := c.SpreadValues()
	return values
}
