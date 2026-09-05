package mcp

import (
	"context"
	"fmt"

	"github.com/EliCDavis/polyform/generator/subgraph"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type CreateEquationSubgraphInput struct {
	Id          string `json:"id" jsonschema:"unique id for the new subgraph"`
	Name        string `json:"name,omitempty" jsonschema:"human-readable display name; defaults to the equation's output variable name"`
	Description string `json:"description,omitempty"`
	Equation    string `json:"equation" jsonschema:"an equation of the form 'output = expression', e.g. 'c = sqrt(a^2 + b^2)'. Supports + - * / ^ (numeric-literal exponents only, including negative and 0.5 for sqrt), unary minus, parentheses, the functions sqrt(x), hypot(a,b)/hypotenuse(a,b), min(a,b,...), max(a,b,...), and the constants pi and e (lowercase only). Every other bare identifier becomes a float64 boundary input, deduplicated by name if it appears more than once. There is no general pow(base,exponent), sin/cos/tan, or abs node in polyform, so those are not supported — you'll get a clear error naming the unsupported piece rather than a silent wrong answer."`
}

type CreateEquationSubgraphOutput struct {
	SubgraphId string   `json:"subgraphId"`
	Inputs     []string `json:"inputs" jsonschema:"boundary input names, in first-appearance order in the equation"`
	Output     string   `json:"output" jsonschema:"boundary output name (the equation's left-hand side)"`
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
