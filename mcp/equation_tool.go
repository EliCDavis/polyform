package mcp

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/EliCDavis/polyform/generator/subgraph"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func constantValues(names []string) []float64 {
	values := make([]float64, len(names))
	for i, n := range names {
		if n == "pi" {
			values[i] = math.Pi
		} else {
			values[i] = math.E
		}
	}
	return values
}

type CreateEquationSubgraphInput struct {
	Id          string `json:"id" jsonschema:"unique id for the new subgraph"`
	Name        string `json:"name,omitempty" jsonschema:"human-readable display name; defaults to the equation's output variable name"`
	Description string `json:"description,omitempty"`
	Equation    string `json:"equation" jsonschema:"an equation of the form 'output = expression', e.g. 'c = sqrt(a^2 + b^2)'. Supports + - * / ^ (numeric-literal exponents only, including negative and 0.5 for sqrt), unary minus, parentheses, the functions sqrt(x), hypot(a,b)/hypotenuse(a,b), min(a,b,...), max(a,b,...), floor(x), ceil(x), round(x), abs(x), mod(a,b) (remainder taking the divisor's sign, so it never goes negative over a repeating range — which board of a wall, which tooth of a gear), sin(x), cos(x), tan(x), asin(x), acos(x), atan(x), atan2(y,x), radians(deg), degrees(rad) (angles in radians, matching quaternion.FromEulerAngleNode), and the constants pi and e (lowercase only). Every other bare identifier becomes a float64 boundary input, deduplicated by name if it appears more than once. There is no general pow(base,exponent) node, so that is not supported — you'll get a clear error naming the unsupported piece rather than a silent wrong answer."`
}

type CreateEquationSubgraphOutput struct {
	SubgraphId string   `json:"subgraphId"`
	Inputs     []string `json:"inputs" jsonschema:"boundary input names, in first-appearance order in the equation"`
	Output     string   `json:"output" jsonschema:"boundary output name (the equation's left-hand side)"`
	Warning    string   `json:"warning,omitempty"`
}

func (s *Server) createEquationSubgraph(ctx context.Context, req *mcpsdk.CallToolRequest, in CreateEquationSubgraphInput) (*mcpsdk.CallToolResult, CreateEquationSubgraphOutput, error) {
	var out CreateEquationSubgraphOutput
	var err error
	s.atomic(&err, func() error {
		outputName, ast, e := parseEquation(in.Equation)
		if e != nil {
			return fmt.Errorf("parsing equation %q: %w", in.Equation, e)
		}

		name := in.Name
		if name == "" {
			name = outputName
		}

		if e := s.graph.CreateSubGraph(in.Id, name, in.Description); e != nil {
			return e
		}
		child, e := s.graph.SubGraphInstance(in.Id)
		if e != nil {
			return e
		}

		compiler := newEquationCompiler(child)
		result, e := compiler.compile(ast)
		if e != nil {
			return fmt.Errorf("compiling equation %q: %w", in.Equation, e)
		}

		_, outID, e := child.CreateBoundaryNode(subgraph.OutputNodeTypeKey, "float64")
		if e != nil {
			return e
		}
		if e := child.SetBoundaryNodeInfo(outID, outputName); e != nil {
			return e
		}
		child.ConnectNodes(result.nodeID, result.port, outID, "Value")

		out.SubgraphId = in.Id
		out.Inputs = compiler.varOrder
		out.Output = outputName
		if len(compiler.constants) > 0 {
			out.Warning = fmt.Sprintf("%s read as the constant(s) %v, not as input(s); rename (e.g. 'turn', 'p') if a value was meant to be wired in", strings.Join(compiler.constants, " and "), constantValues(compiler.constants))
			if len(compiler.varOrder) == 0 {
				out.Warning += " - this equation has no inputs at all, so its output is a fixed number"
			}
		}
		return nil
	})
	return nil, out, err
}

func (s *Server) registerEquationTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "create_equation_subgraph",
		Description: "Compile a math equation (e.g. \"c = sqrt(a^2 + b^2)\") directly into a new subgraph with one float64 boundary input per free variable and a boundary output for the result — instead of hand-wiring several AddNode/SubtractNode/MultiplyNode/DivideNode/etc. one at a time and connecting them together. Use this for an expression combining two or more operations (distances, ratios, derived dimensions, easing). For a single operation - one multiply, one add, one subtract, one divide - use the corresponding plain node (MultiplyNode, AddNode, SubtractNode, DivideNode) directly instead: creating and naming a whole subgraph for one operation is unnecessary overhead a single node avoids.",
	}, s.createEquationSubgraph)
}
