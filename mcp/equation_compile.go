package mcp

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
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
	eqRoundNodeType  = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.RoundNode]"
	eqAbsNodeType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.AbsNode]"
	eqModNodeType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.ModuloNode]"

	eqSinNodeType   = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/trig.SinNode]"
	eqCosNodeType   = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/trig.CosNode]"
	eqTanNodeType   = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/trig.TanNode]"
	eqAsinNodeType  = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/trig.ArcSinNode]"
	eqAcosNodeType  = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/trig.ArcCosNode]"
	eqAtanNodeType  = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/trig.ArcTanNode]"
	eqAtan2NodeType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/trig.ArcTan2Node]"
	eqRadNodeType   = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/trig.DegreesToRadiansNode]"
	eqDegNodeType   = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/trig.RadiansToDegreesNode]"

	eqCompareNodeType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.CompareNode[float64]]"
	eqNotNodeType     = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.NotNode]"
	eqAndNodeType     = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/nodes/logic.AndNode]"
	eqOrNodeType      = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/nodes/logic.OrNode]"
	eqSelectNodeType  = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/nodes/logic.SelectNode]"
)

var eqComparePorts = map[string]string{
	"<":  "Less",
	"<=": "Less Or Equal",
	">":  "Greater",
	">=": "Greater Or Equal",
	"==": "Equal",
	"!=": "Not Equal",
}

// eqPort identifies a specific node's output port within the subgraph
// being built.
type eqPort struct {
	nodeID string
	port   string
	cond   bool
}

type eqWire struct {
	port string
	from eqPort
}

// equationCompiler walks an eqExpr AST and emits polyform nodes into inst
// (a subgraph's own child Instance), deduplicating repeated variable
// references and repeated numeric literals so e.g. "y = a*x^2 + b*x + c"
// reuses a single boundary input for each of x/a/b/c rather than creating
// one per occurrence.
type equationCompiler struct {
	inst      *graph.Graph
	variables map[string]eqPort
	varOrder  []string
	literals  map[float64]eqPort
	constants []string

	// Boundary input nodes already in inst, reused by name instead of created.
	existing map[string]string
	// Input names typed int; the boundary is int and converted inside.
	ints map[string]bool
	// Outputs of earlier statements, read by name in later ones.
	results map[string]eqPort
	// Input names typed bool.
	bools map[string]bool
	// Boundary input node id by input name.
	inputIDs map[string]string
	// Node id by node type and the ports feeding it.
	shared map[string]string
}

func newEquationCompiler(inst *graph.Graph) *equationCompiler {
	return &equationCompiler{
		inst:      inst,
		variables: map[string]eqPort{},
		literals:  map[float64]eqPort{},
		existing:  map[string]string{},
		ints:      map[string]bool{},
		bools:     map[string]bool{},
		results:   map[string]eqPort{},
		inputIDs:  map[string]string{},
		shared:    map[string]string{},
	}
}

func (c *equationCompiler) compileStatements(statements []eqStatement) ([]eqPort, error) {
	ports := make([]eqPort, len(statements))
	for i, statement := range statements {
		port, err := c.compile(statement.expr)
		if err != nil {
			return nil, err
		}
		ports[i] = port
		c.results[statement.output] = port
	}
	return ports, nil
}

func (c *equationCompiler) intOutput(result eqPort) (eqPort, error) {
	if result.cond {
		return eqPort{}, fmt.Errorf("an int output can't hold a condition")
	}
	return c.unary(eqRoundNodeType, "In", "Int", result)
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
	if p, ok := c.results[name]; ok {
		return p, nil
	}
	if p, ok := c.variables[name]; ok {
		return p, nil
	}
	id, ok := c.existing[name]
	if !ok {
		portType := "float64"
		if c.ints[name] {
			portType = "int"
		}
		if c.bools[name] {
			portType = "bool"
		}
		var err error
		if _, id, err = c.inst.CreateBoundaryNode(subgraph.InputNodeTypeKey, portType); err != nil {
			return eqPort{}, fmt.Errorf("create input %q: %w", name, err)
		}
		if err := c.inst.SetBoundaryNodeInfo(id, name); err != nil {
			return eqPort{}, fmt.Errorf("name input %q: %w", name, err)
		}
	}
	p := eqPort{nodeID: id, port: "Value", cond: c.bools[name]}
	c.variables[name] = p
	c.inputIDs[name] = id
	c.varOrder = append(c.varOrder, name)
	return p, nil
}

// A node of the same type fed by the same ports is reused, not rebuilt.
func (c *equationCompiler) node(typeKey, outPort string, orderFree bool, wires ...eqWire) (eqPort, error) {
	parts := make([]string, len(wires))
	for i, w := range wires {
		parts[i] = w.port + "<" + w.from.nodeID + "." + w.from.port
	}
	if orderFree {
		slices.Sort(parts)
	}
	key := typeKey + "|" + strings.Join(parts, "|")
	if id, ok := c.shared[key]; ok {
		return eqPort{nodeID: id, port: outPort}, nil
	}

	_, id, err := c.inst.CreateNode(typeKey)
	if err != nil {
		return eqPort{}, err
	}
	for _, w := range wires {
		if err := c.inst.ConnectNodes(w.from.nodeID, w.from.port, id, w.port); err != nil {
			return eqPort{}, err
		}
	}
	c.shared[key] = id
	return eqPort{nodeID: id, port: outPort}, nil
}

func (c *equationCompiler) unary(typeKey, inPort, outPort string, a eqPort) (eqPort, error) {
	return c.node(typeKey, outPort, false, eqWire{inPort, a})
}

func (c *equationCompiler) binary(typeKey, portA, portB, outPort string, a, b eqPort) (eqPort, error) {
	return c.node(typeKey, outPort, portA == portB, eqWire{portA, a}, eqWire{portB, b})
}

func (c *equationCompiler) nary(typeKey, inPort, outPort string, args []eqPort) (eqPort, error) {
	wires := make([]eqWire, len(args))
	for i, a := range args {
		wires[i] = eqWire{inPort, a}
	}
	return c.node(typeKey, outPort, true, wires...)
}

func asCondition(p eqPort, err error) (eqPort, error) {
	p.cond = true
	return p, err
}

func requireNumbers(what string, args ...eqPort) error {
	for _, a := range args {
		if a.cond {
			return fmt.Errorf("%s takes numbers, not a condition; turn a condition into a number with if(cond, 1, 0)", what)
		}
	}
	return nil
}

func requireConditions(what string, args ...eqPort) error {
	for _, a := range args {
		if !a.cond {
			return fmt.Errorf("%s takes conditions, such as a < b, not numbers", what)
		}
	}
	return nil
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
// it's made up entirely of numeric literals, the pi constant, and
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
		if e.name == "pi" {
			return math.Pi, true
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
	if err := requireNumbers("'^'", base); err != nil {
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
	case "if":
		if len(args) != 3 {
			return eqPort{}, fmt.Errorf("if() takes 3 arguments (condition, value when true, value when false), got %d", len(args))
		}
		if err := requireConditions("if()'s first argument", args[0]); err != nil {
			return eqPort{}, err
		}
		if args[1].cond != args[2].cond {
			return eqPort{}, fmt.Errorf("if()'s two values must both be numbers or both be conditions")
		}
		p, err := c.node(eqSelectNodeType, "Out", false, eqWire{"Condition", args[0]}, eqWire{"A", args[1]}, eqWire{"B", args[2]})
		p.cond = args[1].cond
		return p, err

	case "and", "or":
		if len(args) < 2 {
			return eqPort{}, fmt.Errorf("%s() takes at least 2 conditions, got %d", name, len(args))
		}
		if err := requireConditions(name+"()", args...); err != nil {
			return eqPort{}, err
		}
		typeKey := eqAndNodeType
		if strings.ToLower(name) == "or" {
			typeKey = eqOrNodeType
		}
		return asCondition(c.nary(typeKey, "Values", "Out", args))

	case "not":
		if len(args) != 1 {
			return eqPort{}, fmt.Errorf("not() takes exactly 1 condition, got %d", len(args))
		}
		if err := requireConditions("not()", args...); err != nil {
			return eqPort{}, err
		}
		return asCondition(c.unary(eqNotNodeType, "In", "Out", args[0]))
	}

	if err := requireNumbers(name+"()", args...); err != nil {
		return eqPort{}, err
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

	case "floor", "ceil", "round", "abs":
		if len(args) != 1 {
			return eqPort{}, fmt.Errorf("%s() takes exactly 1 argument, got %d", name, len(args))
		}
		if strings.ToLower(name) == "abs" {
			return c.unary(eqAbsNodeType, "In", "Out", args[0])
		}
		ports := map[string]string{"floor": "Down", "ceil": "Up", "round": "Float"}
		return c.unary(eqRoundNodeType, "In", ports[strings.ToLower(name)], args[0])

	case "mod":
		if len(args) != 2 {
			return eqPort{}, fmt.Errorf("mod() takes exactly 2 arguments, got %d", len(args))
		}
		return c.binary(eqModNodeType, "A", "B", "Out", args[0], args[1])

	// Trig. Angles are radians throughout, matching what
	// quaternion.FromEulerAngleNode expects; radians()/degrees() convert.
	case "atan2":
		if len(args) != 2 {
			return eqPort{}, fmt.Errorf("atan2() takes exactly 2 arguments (y, x), got %d", len(args))
		}
		return c.binary(eqAtan2NodeType, "Y", "X", "Out", args[0], args[1])

	case "sin", "cos", "tan", "asin", "arcsin", "acos", "arccos", "atan", "arctan", "radians", "degrees":
		if len(args) != 1 {
			return eqPort{}, fmt.Errorf("%s() takes exactly 1 argument, got %d", name, len(args))
		}
		unaryTypes := map[string]string{
			"sin": eqSinNodeType, "cos": eqCosNodeType, "tan": eqTanNodeType,
			"asin": eqAsinNodeType, "arcsin": eqAsinNodeType,
			"acos": eqAcosNodeType, "arccos": eqAcosNodeType,
			"atan": eqAtanNodeType, "arctan": eqAtanNodeType,
			"radians": eqRadNodeType, "degrees": eqDegNodeType,
		}
		return c.unary(unaryTypes[strings.ToLower(name)], "Input", "Out", args[0])

	default:
		return eqPort{}, fmt.Errorf("unsupported function %q — supported: sqrt(x), hypot(a,b)/hypotenuse(a,b), min(a,b,...), max(a,b,...), floor(x), ceil(x), round(x), abs(x), mod(a,b), sin(x), cos(x), tan(x), asin(x), acos(x), atan(x), atan2(y,x), radians(deg), degrees(rad), if(cond,a,b), and(c,...), or(c,...), not(c). polyform has no pow node yet — use x*x or sqrt()", name)
	}
}

func (c *equationCompiler) operands(e *eqExpr, op string) (eqPort, eqPort, error) {
	a, err := c.compile(e.args[0])
	if err != nil {
		return eqPort{}, eqPort{}, err
	}
	b, err := c.compile(e.args[1])
	if err != nil {
		return eqPort{}, eqPort{}, err
	}
	return a, b, requireNumbers("'"+op+"'", a, b)
}

func (c *equationCompiler) compile(e *eqExpr) (eqPort, error) {
	switch e.kind {
	case eqExprNum:
		return c.literal(e.num)

	case eqExprVar:
		if e.name == "pi" {
			if !slices.Contains(c.constants, "pi") {
				c.constants = append(c.constants, "pi")
			}
			return c.literal(math.Pi)
		}
		return c.variable(e.name)

	case eqExprNeg:
		inner, err := c.compile(e.args[0])
		if err != nil {
			return eqPort{}, err
		}
		if err := requireNumbers("unary minus", inner); err != nil {
			return eqPort{}, err
		}
		return c.unary(eqNegNodeType, "In", "Out", inner)

	case eqExprCall:
		return c.call(e.name, e.args)

	case eqExprCompare:
		a, b, err := c.operands(e, e.cmp)
		if err != nil {
			return eqPort{}, err
		}
		return asCondition(c.node(eqCompareNodeType, eqComparePorts[e.cmp], false, eqWire{"A", a}, eqWire{"B", b}))

	case eqExprBinOp:
		if e.op == '^' {
			return c.power(e.args[0], e.args[1])
		}
		a, b, err := c.operands(e, string(e.op))
		if err != nil {
			return eqPort{}, err
		}
		switch e.op {
		case '+':
			return c.binary(eqAddNodeType, "Values", "Values", "Float", a, b)
		case '-':
			return c.binary(eqSubNodeType, "A", "B", "Float", a, b)
		case '*':
			return c.binary(eqMulNodeType, "Values", "Values", "Float", a, b)
		case '/':
			return c.binary(eqDivNodeType, "Dividend", "Divisor", "Float", a, b)
		}
	}

	return eqPort{}, fmt.Errorf("internal error: unhandled expression kind %v", e.kind)
}
