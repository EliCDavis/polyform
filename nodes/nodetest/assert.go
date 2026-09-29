package nodetest

import (
	"strings"
	"testing"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

type Assertion interface {
	Assert(t *testing.T, node nodes.Node)
}

// ============================================================================

type AssertOutputPortValue[T any] struct {
	Port            string
	Value           T
	ExecutionReport *nodes.ExecutionReport
}

func (apv AssertOutputPortValue[T]) Assert(t *testing.T, node nodes.Node) {
	out := nodes.GetNodeOutputPort[T](node, apv.Port)
	assert.Equal(t, apv.Value, out.Value())

	if apv.ExecutionReport == nil {
		return
	}

	obvervable, ok := out.(nodes.ObservableExecution)
	if !ok {
		t.Error("node output does not have expected execution report")
	}

	report := obvervable.ExecutionReport()
	assert.Equal(t, apv.ExecutionReport.Logs, report.Logs)
	assert.Equal(t, apv.ExecutionReport.Errors, report.Errors)
}

func AssertOutput[T any](port string, value T) AssertOutputPortValue[T] {
	return AssertOutputPortValue[T]{
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
	inputs := node.Inputs()

	port, ok := inputs[apv.Port]
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

func NewNode[T any](data T) nodes.Node {
	return &nodes.Struct[T]{
		Data: data,
	}
}

func NewPortValue[T any](data T) nodes.Output[T] {
	return nodes.ConstOutput[T]{Val: data}
}

// ============================================================================

type AssertOutputPortType struct {
	Port string
	Type string
}

func (apt AssertOutputPortType) Assert(t *testing.T, node nodes.Node) {
	port, ok := node.Outputs()[apt.Port]
	if !ok {
		t.Error("node does not contain output port", apt.Port)
		return
	}

	typed, ok := port.(nodes.Typed)
	if !ok {
		t.Error("node output port does not report a type", apt.Port)
		return
	}

	assert.Equal(t, apt.Type, typed.Type())
}

// AssertOutputType pins what an output is carrying, which for a lifted port
// is the rank its inputs settled on rather than anything fixed by the type.
func AssertOutputType(port, portType string) AssertOutputPortType {
	return AssertOutputPortType{Port: port, Type: portType}
}

// ============================================================================

type AssertInputPortRanks struct {
	Port string
	Elem string
}

func (apr AssertInputPortRanks) Assert(t *testing.T, node nodes.Node) {
	port, ok := node.Inputs()[apr.Port]
	if !ok {
		t.Error("node does not contain input port", apr.Port)
		return
	}

	options, ok := port.(nodes.TypeOptions)
	if !ok {
		t.Error("node input port takes only one type", apr.Port)
		return
	}

	assert.Equal(t, []string{apr.Elem, "[]" + apr.Elem}, options.AcceptedTypes())
}

// AssertLiftedInput checks a port really does take both a single value and an
// array of them, which is what stops an editor refusing one of the two.
func AssertLiftedInput(port, elem string) AssertInputPortRanks {
	return AssertInputPortRanks{Port: port, Elem: elem}
}

// ============================================================================

type AssertOutputPortError struct {
	Port     string
	Contains string
}

func (ape AssertOutputPortError) Assert(t *testing.T, node nodes.Node) {
	port, ok := node.Outputs()[ape.Port]
	if !ok {
		t.Error("node does not contain output port", ape.Port)
		return
	}

	observable, ok := port.(nodes.ObservableExecution)
	if !ok {
		t.Error("node output port reports no execution", ape.Port)
		return
	}

	errors := observable.ExecutionReport().Errors
	if len(errors) == 0 {
		t.Errorf("output port %q reported no errors, wanted one containing %q", ape.Port, ape.Contains)
		return
	}

	for _, err := range errors {
		if strings.Contains(err, ape.Contains) {
			return
		}
	}
	t.Errorf("output port %q errors %v, none containing %q", ape.Port, errors, ape.Contains)
}

func AssertOutputError(port, contains string) AssertOutputPortError {
	return AssertOutputPortError{Port: port, Contains: contains}
}

// ============================================================================

// NewPortValues wires several constants into one variadic lifted input.
func NewPortValues[T any](values ...T) []nodes.LiftedPort[T] {
	ports := make([]nodes.LiftedPort[T], len(values))
	for i, v := range values {
		ports[i] = nodes.ConstOutput[T]{Val: v}
	}
	return ports
}

// ============================================================================

type AssertNodeOutputPortDescription struct {
	Port        string
	Description string
}

func (apv AssertNodeOutputPortDescription) Assert(t *testing.T, node nodes.Node) {
	inputs := node.Outputs()

	port, ok := inputs[apv.Port]
	if !ok {
		t.Error("node does not contain output port", apv.Port)
		return
	}

	describable, ok := port.(nodes.Describable)
	if !ok {
		t.Error("node output port does not contain a description", apv.Port)
		return
	}

	assert.Equal(t, apv.Description, describable.Description())
}
