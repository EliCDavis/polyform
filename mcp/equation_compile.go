package mcp

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/subgraph"
)

// Exact registered type keys for the math nodes the equation compiler
// targets, verified against a running type registry rather than guessed —
// see mcp/equation_compile_test.go's TestEquationNodeTypesAreRegistered.
const (
	eqFloatParamType = "github.com/EliCDavis/polyform/generator/parameter.Value[float64]"
	eqAddNodeType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.AddNode[float64]]"
	eqSubNodeType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.SubtractNode[float64]]"
	eqMulNodeType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.MultiplyNode[float64]]"
	eqDivNodeType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.DivideNode[float64]]"
	eqNegNodeType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.NegateNode[float64]]"
	eqMinNodeType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.MinNode[float64]]"
	eqMaxNodeType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.MaxNode[float64]]"
	eqSqrtNodeType   = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.SquareRootNode]"
	eqSquareNodeType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.SquareNode]"
	eqHypotNodeType  = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.HypotenuseNode]"
)

// eqPort identifies a specific node's output port within the subgraph
// being built.
type eqPort struct {
	nodeID string
	port   string
}

// equationCompiler walks an eqExpr AST and emits polyform nodes into inst
// (a subgraph's own child Instance), deduplicating repeated variable
// references and repeated numeric literals so e.g. "y = a*x^2 + b*x + c"
// reuses a single boundary input for each of x/a/b/c rather than creating
// one per occurrence.
type equationCompiler struct {
	inst      *graph.Instance
	variables map[string]eqPort
	varOrder  []string
	literals  map[float64]eqPort
}

func newEquationCompiler(inst *graph.Instance) *equationCompiler {
	return &equationCompiler{
		inst:      inst,
		variables: map[string]eqPort{},
		literals:  map[float64]eqPort{},
	}
}

func (c *equationCompiler) literal(v float64) (eqPort, error) {
	if p, ok := c.literals[v]; ok {
		return p, nil
	}
	_, id, err := c.inst.CreateNode(eqFloatParamType)
	if err != nil {
		return eqPort{}, fmt.Errorf("create literal %v: %w", v, err)
	}
	data, err := json.Marshal(v)
	if err != nil {
		return eqPort{}, err
	}
	if _, err := c.inst.UpdateParameter(id, data); err != nil {
		return eqPort{}, fmt.Errorf("set literal %v: %w", v, err)
	}
	p := eqPort{nodeID: id, port: "Value"}
	c.literals[v] = p
	return p, nil
}

func (c *equationCompiler) variable(name string) (eqPort, error) {
	if p, ok := c.variables[name]; ok {
		return p, nil
	}
	_, id, err := c.inst.CreateBoundaryNode(subgraph.InputNodeTypeKey, "float64")
	if err != nil {
		return eqPort{}, fmt.Errorf("create input %q: %w", name, err)
	}
	if err := c.inst.SetBoundaryNodeInfo(id, name); err != nil {
		return eqPort{}, fmt.Errorf("name input %q: %w", name, err)
	}
	p := eqPort{nodeID: id, port: "Value"}
	c.variables[name] = p
	c.varOrder = append(c.varOrder, name)
	return p, nil
}

func (c *equationCompiler) unary(typeKey, inPort, outPort string, a eqPort) (eqPort, error) {
	_, id, err := c.inst.CreateNode(typeKey)
	if err != nil {
		return eqPort{}, err
	}
	c.inst.ConnectNodes(a.nodeID, a.port, id, inPort)
	return eqPort{nodeID: id, port: outPort}, nil
}

func (c *equationCompiler) binary(typeKey, portA, portB, outPort string, a, b eqPort) (eqPort, error) {
	_, id, err := c.inst.CreateNode(typeKey)
	if err != nil {
		return eqPort{}, err
	}
	c.inst.ConnectNodes(a.nodeID, a.port, id, portA)
	c.inst.ConnectNodes(b.nodeID, b.port, id, portB)
	return eqPort{nodeID: id, port: outPort}, nil
}

func (c *equationCompiler) nary(typeKey, inPort, outPort string, args []eqPort) (eqPort, error) {
	_, id, err := c.inst.CreateNode(typeKey)
	if err != nil {
		return eqPort{}, err
	}
	for _, a := range args {
		c.inst.ConnectNodes(a.nodeID, a.port, id, inPort)
	}
	return eqPort{nodeID: id, port: outPort}, nil
}

// integerPower expands base^n (n a non-negative integer) into repeated
// multiplication, since polyform has no general pow node. n==2 uses the
// dedicated SquareNode; other exponents chain MultiplyNode calls.
func (c *equationCompiler) integerPower(base eqPort, n int) (eqPort, error) {
	switch {
	case n == 0:
		return c.literal(1)
	case n == 1:
		return base, nil
	case n == 2:
		return c.unary(eqSquareNodeType, "In", "Out", base)
	}
	acc := base
	var err error
	for i := 1; i < n; i++ {
		acc, err = c.binary(eqMulNodeType, "Values", "Values", "Float", acc, base)
		if err != nil {
			return eqPort{}, err
		}
	}
	return acc, nil
}

// evalConstantExpr evaluates e in plain Go (not by emitting graph nodes) if
// it's made up entirely of numeric literals, the pi/e constants, and
// +-*/^ — i.e. it contains no variable and no function call. This is what
// lets an exponent like x^(2^3) or x^(1+1) work: the exponent itself is a
// compile-time constant, even though it's not a single bare literal token.
// It returns ok=false the moment it hits anything that depends on a
// runtime value, since that can't be resolved without a graph node.
func evalConstantExpr(e *eqExpr) (float64, bool) {
	switch e.kind {
	case eqExprNum:
		return e.num, true
	case eqExprVar:
		switch e.name {
		case "pi":
			return math.Pi, true
		case "e":
			return math.E, true
		}
		return 0, false
	case eqExprNeg:
		v, ok := evalConstantExpr(e.args[0])
		return -v, ok
	case eqExprBinOp:
		a, ok := evalConstantExpr(e.args[0])
		if !ok {
			return 0, false
		}
		b, ok := evalConstantExpr(e.args[1])
		if !ok {
			return 0, false
		}
		switch e.op {
		case '+':
			return a + b, true
		case '-':
			return a - b, true
		case '*':
			return a * b, true
		case '/':
			if b == 0 {
				return 0, false
			}
			return a / b, true
		case '^':
			return math.Pow(a, b), true
		}
	}
	return 0, false
}

func (c *equationCompiler) power(baseExpr, expExpr *eqExpr) (eqPort, error) {
	n, ok := evalConstantExpr(expExpr)
	if !ok {
		return eqPort{}, fmt.Errorf("exponent must be a compile-time constant (e.g. x^2, x^0.5, x^(1+1)) — polyform has no general pow(base, exponent) node to fall back on for an exponent that depends on a variable")
	}

	base, err := c.compile(baseExpr)
	if err != nil {
		return eqPort{}, err
	}

	switch {
	case n == 0.5:
		return c.unary(eqSqrtNodeType, "In", "Out", base)
	case n == math.Trunc(n) && n >= 0:
		return c.integerPower(base, int(n))
	case n == math.Trunc(n) && n < 0:
		positive, err := c.integerPower(base, int(-n))
		if err != nil {
			return eqPort{}, err
		}
		one, err := c.literal(1)
		if err != nil {
			return eqPort{}, err
		}
		return c.binary(eqDivNodeType, "Dividend", "Divisor", "Float", one, positive)
	default:
		return eqPort{}, fmt.Errorf("unsupported exponent %v — only integer powers and 0.5 (square root) are supported, since polyform has no general pow node", n)
	}
}

func (c *equationCompiler) call(name string, argExprs []*eqExpr) (eqPort, error) {
	args := make([]eqPort, len(argExprs))
	for i, a := range argExprs {
		p, err := c.compile(a)
		if err != nil {
			return eqPort{}, err
		}
		args[i] = p
	}

	switch strings.ToLower(name) {
	case "sqrt":
		if len(args) != 1 {
			return eqPort{}, fmt.Errorf("sqrt() takes exactly 1 argument, got %d", len(args))
		}
		return c.unary(eqSqrtNodeType, "In", "Out", args[0])

	case "hypot", "hypotenuse":
		if len(args) != 2 {
			return eqPort{}, fmt.Errorf("%s() takes exactly 2 arguments, got %d", name, len(args))
		}
		return c.binary(eqHypotNodeType, "P", "Q", "Out", args[0], args[1])

	case "min":
		if len(args) < 2 {
			return eqPort{}, fmt.Errorf("min() takes at least 2 arguments, got %d", len(args))
		}
		return c.nary(eqMinNodeType, "In", "Float 64", args)

	case "max":
		if len(args) < 2 {
			return eqPort{}, fmt.Errorf("max() takes at least 2 arguments, got %d", len(args))
		}
		return c.nary(eqMaxNodeType, "In", "Float 64", args)

	default:
		return eqPort{}, fmt.Errorf("unsupported function %q — supported: sqrt(x), hypot(a,b)/hypotenuse(a,b), min(a,b,...), max(a,b,...). polyform has no scalar sin/cos/tan/abs/pow node yet", name)
	}
}

func (c *equationCompiler) compile(e *eqExpr) (eqPort, error) {
	switch e.kind {
	case eqExprNum:
		return c.literal(e.num)

	case eqExprVar:
		switch e.name {
		case "pi":
			return c.literal(math.Pi)
		case "e":
			return c.literal(math.E)
		}
		return c.variable(e.name)

	case eqExprNeg:
		inner, err := c.compile(e.args[0])
		if err != nil {
			return eqPort{}, err
		}
		return c.unary(eqNegNodeType, "In", "Out", inner)

	case eqExprCall:
		return c.call(e.name, e.args)

	case eqExprBinOp:
		switch e.op {
		case '+':
			a, err := c.compile(e.args[0])
			if err != nil {
				return eqPort{}, err
			}
			b, err := c.compile(e.args[1])
			if err != nil {
				return eqPort{}, err
			}
			return c.binary(eqAddNodeType, "Values", "Values", "Float", a, b)
		case '-':
			a, err := c.compile(e.args[0])
			if err != nil {
				return eqPort{}, err
			}
			b, err := c.compile(e.args[1])
			if err != nil {
				return eqPort{}, err
			}
			return c.binary(eqSubNodeType, "A", "B", "Float", a, b)
		case '*':
			a, err := c.compile(e.args[0])
			if err != nil {
				return eqPort{}, err
			}
			b, err := c.compile(e.args[1])
			if err != nil {
				return eqPort{}, err
			}
			return c.binary(eqMulNodeType, "Values", "Values", "Float", a, b)
		case '/':
			a, err := c.compile(e.args[0])
			if err != nil {
				return eqPort{}, err
			}
			b, err := c.compile(e.args[1])
			if err != nil {
				return eqPort{}, err
			}
			return c.binary(eqDivNodeType, "Dividend", "Divisor", "Float", a, b)
		case '^':
			return c.power(e.args[0], e.args[1])
		}
	}

	return eqPort{}, fmt.Errorf("internal error: unhandled expression kind %v", e.kind)
}
