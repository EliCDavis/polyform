package nodes_test

import (
	"testing"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

type settablePort struct {
	value   int
	version int
}

func (p *settablePort) Node() nodes.Node { return nil }
func (p *settablePort) Name() string     { return "Value" }
func (p *settablePort) Version() int     { return p.version }
func (p *settablePort) Value() int       { return p.value }
func (p *settablePort) Type() string     { return "int" }

type promiseDoubleNode struct {
	In nodes.Output[int]
}

func (n promiseDoubleNode) Out(out *nodes.StructOutput[int]) {
	out.Set(n.In.Value() * 2)
}

var afterPromiseRead func()

type promiseReadNode struct {
	In nodes.Output[int]
}

func (n promiseReadNode) Out(out *nodes.StructOutput[int]) {
	v := n.In.Value()
	if afterPromiseRead != nil {
		afterPromiseRead()
	}
	out.Set(v)
}

func TestAVersionReportedWhileOutdatedIsNeverReusedForAnotherValue(t *testing.T) {
	src := &settablePort{value: 1}
	doubled := &nodes.Struct[promiseDoubleNode]{Data: promiseDoubleNode{In: src}}
	reader := &nodes.Struct[promiseReadNode]{Data: promiseReadNode{In: doubled.Outputs()["Out"].(nodes.Output[int])}}
	out := reader.Outputs()["Out"].(nodes.Output[int])

	afterPromiseRead = func() {
		src.version++
		nodes.Touch()
	}
	assert.Equal(t, 2, out.Value())
	afterPromiseRead = nil

	src.value = 5
	src.version++
	nodes.Touch()
	assert.Equal(t, 10, out.Value())
}
