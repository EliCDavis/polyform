package nodes_test

import (
	"testing"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type onlyElementRegistered struct{ v int }

type onlyElementRegisteredNode struct{}

func (onlyElementRegisteredNode) Out(out *nodes.StructOutput[*onlyElementRegistered]) {
	out.Set(&onlyElementRegistered{v: 1})
}

type fixedProxySource struct {
	port nodes.OutputPort
	typ  string
}

func (s fixedProxySource) Node() nodes.Node                { return nil }
func (s fixedProxySource) Name() string                    { return "Out" }
func (s fixedProxySource) Version() int                    { return 0 }
func (s fixedProxySource) Type() string                    { return s.typ }
func (s fixedProxySource) CurrentSource() nodes.OutputPort { return s.port }

func TestAListTypeIsKnownFromItsElementAlone(t *testing.T) {
	nodes.DiscoverNodePortTypes(&nodes.Struct[onlyElementRegisteredNode]{})
	listType := "[]*" + pkg + "onlyElementRegistered"

	builder, ok := nodes.LookupPortTypeProxy(listType)
	require.True(t, ok)
	list := nodes.ConstOutput[[]*onlyElementRegistered]{Val: []*onlyElementRegistered{{v: 1}, {v: 2}}}
	proxy := builder.BuildProxyOutput(fixedProxySource{port: list, typ: listType})
	out, ok := proxy.(nodes.Output[[]*onlyElementRegistered])
	require.True(t, ok, "%T", proxy)
	assert.Len(t, out.Value(), 2)
}
