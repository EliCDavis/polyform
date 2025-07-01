package nodetest

import (
	"testing"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

type Assertion interface {
	Assert(t *testing.T, node nodes.Node)
}

// ============================================================================

type assertPortValue[T any] struct {
	Port  string
	Value T
}

func (apv assertPortValue[T]) Assert(t *testing.T, node nodes.Node) {
	out := nodes.GetNodeOutputPort[T](node, apv.Port).Value()
	assert.Equal(t, apv.Value, out)
}

func AssertOutput[T any](port string, value T) Assertion {
	return assertPortValue[T]{
		Port:  port,
		Value: value,
	}
}

// ============================================================================

type AssertNodeDescription struct {
	Description string
}

func (apv AssertNodeDescription) Assert(t *testing.T, node nodes.Node) {
	if describable, ok := node.(nodes.Describable); ok {
		assert.Equal(t, apv.Description, describable.Description())
		return
	}
	t.Error("node does not contain a description")
}

// ============================================================================

type AssertNodeInputPortDescription struct {
	Port        string
	Description string
}

func (apv AssertNodeInputPortDescription) Assert(t *testing.T, node nodes.Node) {
	outputs := node.Inputs()

	port, ok := outputs[apv.Port]
	if !ok {
		t.Error("node does not contain input port", apv.Port)
		return
	}

	describable, ok := port.(nodes.Describable)
	if !ok {
		t.Error("node input port does not contain a description", apv.Port)
		return
	}

	assert.Equal(t, apv.Description, describable.Description())
}

func NewAssertInputPortDescription(port, description string) AssertNodeInputPortDescription {
	return AssertNodeInputPortDescription{
		Port:        port,
		Description: description,
	}
}

// ============================================================================

func NewNode[T any](data T) nodes.Node {
	return &nodes.Struct[T]{
		Data: data,
	}
}

func NewPortValue[T any](data T) nodes.Output[T] {
	return nodes.ConstOutput[T]{Val: data}
}
