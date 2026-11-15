package mcp

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"math"
	"slices"
	"strings"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/subgraph"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func constantValues(names []string) []float64 {
	values := make([]float64, len(names))
	for i := range names {
		values[i] = math.Pi
	}
	return values
}

type CreateEquationSubgraphInput struct {
	Id          string             `json:"id" jsonschema:"unique id for the new subgraph"`
	Name        string             `json:"name,omitempty" jsonschema:"human-readable display name; defaults to the equation's output variable name"`
	Description string             `json:"description,omitempty"`
	Equation    string             `json:"equation" jsonschema:"an equation of the form 'output = expression', e.g. 'c = sqrt(a^2 + b^2)'. Supports + - * / ^ (numeric-literal exponents only, including negative and 0.5 for sqrt), unary minus, parentheses, the functions sqrt(x), hypot(a,b)/hypotenuse(a,b), min(a,b,...), max(a,b,...), floor(x), ceil(x), round(x), abs(x), mod(a,b) (remainder taking the divisor's sign, so it never goes negative over a repeating range — which copy of a repeating pattern an index lands on), sin(x), cos(x), tan(x), asin(x), acos(x), atan(x), atan2(y,x), radians(deg), degrees(rad) (angles in radians, matching quaternion.FromEulerAngleNode), comparisons < <= > >= == != (a condition, not a number), if(cond, a, b), and(c, ...), or(c, ...), not(c), and the constant pi (lowercase only; e is an ordinary input - write 2.718281828459045 for Euler's number). Every other bare identifier becomes a float64 boundary input, deduplicated by name if it appears more than once. Several statements separated by ';' or newlines ('r = sqrt(x^2+y^2); h = r*2') give one subgraph with one output per statement, sharing inputs; a later statement may read an earlier output by name. A statement whose expression is a condition is a bool output. There is no general pow(base,exponent) node, so that is not supported — you'll get a clear error naming the unsupported piece rather than a silent wrong answer."`
	Types       map[string]string  `json:"types,omitempty" jsonschema:"{input or output name: 'int' or 'bool'} to type that boundary port int or bool instead of float64: an int variable or a count wires straight in; an int output rounds the result (for Times, Segments, Count ports); a bool input is a condition, for if(), and(), or(), not(). Any name not listed stays float64."`
	Constants   map[string]float64 `json:"constants,omitempty" jsonschema:"{input name: default value}: the named input gets that value as its default, so instances read it unless something is wired in, and set_parameter/update_equation_subgraph can tune it later. Write the constant as a named identifier in the equation (k, not 0.52) and give its value here instead of freezing a literal into the definition."`
}

type CreateEquationSubgraphOutput struct {
	SubgraphId string   `json:"subgraphId"`
	Inputs     []string `json:"inputs" jsonschema:"boundary input names, in first-appearance order in the equation"`
	Output     string   `json:"output" jsonschema:"boundary output name (the first statement's left-hand side)"`
	Outputs    []string `json:"outputs,omitempty" jsonschema:"every boundary output name, when the equation has more than one statement"`
	Warning    string   `json:"warning,omitempty"`
}

type equationSpec struct {
	statements      []eqStatement
	types           map[string]string
	constants       map[string]float64
	existingInputs  map[string]string
	existingOutputs map[string]string
}

type equationBuild struct {
	inputs    []string
	outputs   []string
	constants []string
}

func equationPortType(types map[string]string, name string) string {
	switch types[name] {
	case "int", "bool":
		return types[name]
	}
	return "float64"
}

func outputPortType(types map[string]string, conditions map[string]bool, name string) string {
	if conditions[name] {
		return "bool"
	}
	return equationPortType(types, name)
}

func checkEquationTypes(statements []eqStatement, types map[string]string) error {
	named := map[string]bool{}
	inputs := map[string]bool{}
	for _, name := range freeVariables(statements) {
		named[name] = true
		inputs[name] = true
	}
	for _, statement := range statements {
		named[statement.output] = true
	}
	for name, typeName := range types {
		if typeName != "int" && typeName != "float64" && typeName != "bool" {
			return fmt.Errorf("types: %q must be \"int\", \"float64\" or \"bool\", got %q", name, typeName)
		}
		if !named[name] {
			return fmt.Errorf("types: %q is neither an input nor an output of this equation; names are %s", name, strings.Join(slices.Sorted(maps.Keys(named)), ", "))
		}
	}
	conditions := conditionOutputs(statements, types)
	for _, statement := range statements {
		switch typeName := types[statement.output]; {
		case conditions[statement.output] && typeName == "int":
			return fmt.Errorf("types: %q is a condition, so it can't be an int output", statement.output)
		case !conditions[statement.output] && typeName == "bool" && !inputs[statement.output]:
			return fmt.Errorf("types: %q is a number; an output is a bool when its expression is a comparison, and(), or(), not() or an if() choosing between conditions", statement.output)
		}
	}
	return nil
}

func buildEquation(child *graph.Graph, spec equationSpec) (equationBuild, error) {
	var built equationBuild
	if err := checkEquationTypes(spec.statements, spec.types); err != nil {
		return built, err
	}

	compiler := newEquationCompiler(child)
	compiler.existing = spec.existingInputs
	for _, name := range freeVariables(spec.statements) {
		switch equationPortType(spec.types, name) {
		case "int":
			compiler.ints[name] = true
		case "bool":
			compiler.bools[name] = true
		}
	}
	results, err := compiler.compileStatements(spec.statements)
	if err != nil {
		return built, err
	}

	conditions := conditionOutputs(spec.statements, spec.types)
	for i, statement := range spec.statements {
		result := results[i]
		portType := outputPortType(spec.types, conditions, statement.output)
		if portType == "int" {
			if result, err = compiler.intOutput(result); err != nil {
				return built, err
			}
		}
		outID, kept := spec.existingOutputs[statement.output]
		if !kept {
			if outID, err = createBoundary(child, subgraph.OutputNodeTypeKey, statement.output, portType, ""); err != nil {
				return built, err
			}
		}
		if err := child.ConnectNodes(result.nodeID, result.port, outID, subgraph.ValuePortName); err != nil {
			return built, err
		}
		built.outputs = append(built.outputs, statement.output)
	}

	for _, name := range slices.Sorted(maps.Keys(spec.constants)) {
		inputID, isInput := compiler.inputIDs[name]
		if !isInput {
			return built, fmt.Errorf("constants: %q is not an input of this equation; inputs are %s", name, strings.Join(compiler.varOrder, ", "))
		}
		value := spec.constants[name]
		if compiler.bools[name] {
			return built, fmt.Errorf("constants: %q is a bool input; constants hold numbers", name)
		}
		if compiler.ints[name] && value != math.Trunc(value) {
			return built, fmt.Errorf("constants: %q is an int input, so its default must be a whole number, got %v", name, value)
		}
		if err := child.SetLiteral(inputID, subgraph.DefaultPortName, []byte(fmt.Sprintf("%v", value))); err != nil {
			return built, fmt.Errorf("constants: %q: %w", name, err)
		}
	}

	built.inputs = compiler.varOrder
	built.constants = compiler.constants
	return built, nil
}

// A batch drops an edge that doesn't fit where a lone connect would refuse it.
func refuseInteriorDrops(subgraphID string, dropped []graph.DroppedEdge) error {
	for _, drop := range dropped {
		if drop.Scope == subgraphID {
			return fmt.Errorf("%s.%s does not fit %s.%s: %s", drop.From, drop.FromPort, drop.To, drop.ToPort, drop.Reason)
		}
	}
	return nil
}

func (built equationBuild) warning() string {
	if len(built.constants) == 0 {
		return ""
	}
	warning := fmt.Sprintf("%s read as the constant(s) %v, not as input(s); rename (e.g. 'turn', 'p') if a value was meant to be wired in", strings.Join(built.constants, " and "), constantValues(built.constants))
	if len(built.inputs) == 0 {
		warning += " - this equation has no inputs at all, so its output is a fixed number"
	}
	return warning
}

func (s *Server) createEquationSubgraph(ctx context.Context, req *mcpsdk.CallToolRequest, in CreateEquationSubgraphInput) (*mcpsdk.CallToolResult, CreateEquationSubgraphOutput, error) {
	var out CreateEquationSubgraphOutput
	var err error
	s.atomic(&err, func() error {
		statements, e := parseEquations(in.Equation)
		if e != nil {
			return fmt.Errorf("parsing equation %q: %w", in.Equation, e)
		}

		name := in.Name
		if name == "" {
			name = statements[0].output
		}

		if e := s.graph.CreateSubGraph(in.Id, name, in.Description); e != nil {
			return e
		}
		child, e := s.graph.SubGraphInstance(in.Id)
		if e != nil {
			return e
		}

		var built equationBuild
		dropped, e := s.graph.Batch(func() (e error) {
			built, e = buildEquation(child, equationSpec{statements: statements, types: in.Types, constants: in.Constants})
			return e
		})
		if e == nil {
			e = refuseInteriorDrops(in.Id, dropped)
		}
		if e != nil {
			return fmt.Errorf("compiling equation %q: %w", in.Equation, e)
		}
		storeEquation(child, statements)

		out.SubgraphId = in.Id
		out.Inputs = built.inputs
		out.Output = built.outputs[0]
		if len(built.outputs) > 1 {
			out.Outputs = built.outputs
		}
		out.Warning = built.warning()
		return nil
	})
	return nil, out, err
}

type UpdateEquationSubgraphInput struct {
	Id         string             `json:"id" jsonschema:"the equation subgraph to recompile; its id and instances stay. Pass only id to read the current equation back without changing anything"`
	Equation   string             `json:"equation,omitempty" jsonschema:"the whole new equation, same grammar as create_equation_subgraph. Boundary ports whose name (and type) still appear stay, with their outside wiring; a new free variable becomes a new, unwired input on every instance; an input no longer used is removed and its outside wires dropped. A single output whose name changed is renamed, keeping its consumers. For a change to a few statements of a long equation, use statements/remove instead"`
	Statements map[string]string  `json:"statements,omitempty" jsonschema:"{output name: expression} to change or add single statements of the stored equation, e.g. {\"cgZ\": \"min(bz - 2, cz)\"}; every other statement stays as it is. New ones are added, and statements are ordered so each output is computed before it is read"`
	Remove     []string           `json:"remove,omitempty" jsonschema:"output names whose statements to delete from the stored equation"`
	Types      map[string]string  `json:"types,omitempty" jsonschema:"as create_equation_subgraph, merged over the current port types: a port not listed keeps its type. A port whose type changes is recreated, dropping its wires"`
	Constants  map[string]float64 `json:"constants,omitempty" jsonschema:"as create_equation_subgraph; a kept input's default is updated in place, instances still holding the old default follow, and defaults not listed stay"`
}

type UpdateEquationSubgraphOutput struct {
	SubgraphId     string               `json:"subgraphId"`
	Inputs         []string             `json:"inputs" jsonschema:"boundary input names now, in first-appearance order"`
	Output         string               `json:"output"`
	Outputs        []string             `json:"outputs,omitempty"`
	AddedInputs    []string             `json:"addedInputs,omitempty" jsonschema:"inputs that did not exist before: unwired on every instance until fed"`
	RemovedInputs  []string             `json:"removedInputs,omitempty"`
	RemovedOutputs []string             `json:"removedOutputs,omitempty"`
	RenamedOutput  string               `json:"renamedOutput,omitempty" jsonschema:"'old -> new' when the single output changed name; its consumers followed"`
	Dropped        []DroppedEdgeSummary `json:"dropped,omitempty" jsonschema:"outside wires removed with the ports that went: 'scope' is the graph holding the instance"`
	Instances      []string             `json:"instances,omitempty" jsonschema:"'scope/nodeId' of every instance, each of which needs any addedInputs wired"`
	Warning        string               `json:"warning,omitempty"`
	Equation       string               `json:"equation,omitempty" jsonschema:"the current equation, one statement per line; returned when only id was passed"`
}

const equationMetadataKey = "equation"

func storeEquation(child *graph.Graph, statements []eqStatement) {
	lines := make([]any, len(statements))
	for i, statement := range statements {
		lines[i] = statement.text
	}
	child.SetMetadata(equationMetadataKey, lines)
}

func storedEquation(child *graph.Graph) ([]string, bool) {
	raw, ok := child.Metadata(equationMetadataKey).([]any)
	if !ok {
		return nil, false
	}
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		text, ok := line.(string)
		if !ok {
			return nil, false
		}
		lines = append(lines, text)
	}
	return lines, len(lines) > 0
}

func statementOutput(line string) string {
	lhs, _, _ := strings.Cut(line, "=")
	return strings.TrimSpace(lhs)
}

func patchEquation(stored []string, set map[string]string, remove []string) (string, error) {
	removing := map[string]bool{}
	for _, name := range remove {
		if _, alsoSet := set[name]; alsoSet {
			return "", fmt.Errorf("%q is both in statements and remove", name)
		}
		removing[name] = true
	}

	expressions := map[string]string{}
	for name, expr := range set {
		if lhs, rhs, full := strings.Cut(expr, "="); full && isValidEqIdent(strings.TrimSpace(lhs)) && !strings.HasPrefix(rhs, "=") {
			if strings.TrimSpace(lhs) != name {
				return "", fmt.Errorf("statements[%q] assigns %q instead", name, strings.TrimSpace(lhs))
			}
			expr = rhs
		}
		expressions[name] = strings.TrimSpace(expr)
	}

	var lines []string
	present := map[string]bool{}
	for _, line := range stored {
		name := statementOutput(line)
		present[name] = true
		if removing[name] {
			continue
		}
		if expr, ok := expressions[name]; ok {
			line = name + " = " + expr
		}
		lines = append(lines, line)
	}
	for _, name := range remove {
		if !present[name] {
			return "", fmt.Errorf("remove: no statement defines %q", name)
		}
	}
	for _, name := range slices.Sorted(maps.Keys(expressions)) {
		if !present[name] {
			lines = append(lines, name+" = "+expressions[name])
		}
	}
	return strings.Join(lines, "\n"), nil
}

func equationTypes(statements []eqStatement, requested map[string]string, oldInputs, oldOutputs map[string]boundaryPort) map[string]string {
	types := map[string]string{}
	for _, name := range freeVariables(statements) {
		if old, ok := oldInputs[name]; ok && (old.portType == "int" || old.portType == "bool") {
			types[name] = old.portType
		}
	}
	for _, statement := range statements {
		if old, ok := oldOutputs[statement.output]; ok && old.portType == "int" {
			types[statement.output] = "int"
		}
	}
	maps.Copy(types, requested)
	return types
}

type boundaryPort struct {
	id       string
	portType string
}

func boundaryPorts(child *graph.Graph) (inputs, outputs map[string]boundaryPort) {
	inputs, outputs = map[string]boundaryPort{}, map[string]boundaryPort{}
	for _, id := range child.NodeIds() {
		switch boundary := child.Node(id).(type) {
		case *subgraph.InputNode:
			inputs[boundary.PortName] = boundaryPort{id: id, portType: boundary.PortType}
		case *subgraph.OutputNode:
			outputs[boundary.PortName] = boundaryPort{id: id, portType: boundary.PortType}
		}
	}
	return inputs, outputs
}

// carryDefault moves every placement still holding a boundary's old default
// to its current one: a placement's literal is its own once placed.
func (s *Server) carryDefault(subgraphID, boundaryID string, was []byte) error {
	definition, err := s.graph.SubGraphInstance(subgraphID)
	if err != nil {
		return err
	}
	now, err := definition.LiteralValue(boundaryID, subgraph.DefaultPortName)
	if err != nil {
		return err
	}
	graphs := []*graph.Graph{s.graph.Graph}
	for id := range s.graph.Schema().SubGraphs {
		inner, err := s.graph.SubGraphInstance(id)
		if err != nil {
			return err
		}
		graphs = append(graphs, inner)
	}
	for _, holder := range graphs {
		for _, id := range holder.NodeIds() {
			placement, ok := holder.Node(id).(*graph.SubgraphInstanceNode)
			if !ok || placement.SubGraphID() != subgraphID {
				continue
			}
			held, err := holder.LiteralValue(id, boundaryID)
			if err != nil || !bytes.Equal(held, was) {
				continue
			}
			if err := holder.SetLiteral(id, boundaryID, now); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Server) readEquation(id string) (UpdateEquationSubgraphOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := UpdateEquationSubgraphOutput{SubgraphId: id}
	child, err := s.graph.SubGraphInstance(id)
	if err != nil {
		return out, err
	}
	lines, ok := storedEquation(child)
	if !ok {
		return out, fmt.Errorf("subgraph %q has no stored equation", id)
	}
	out.Equation = strings.Join(lines, "\n")
	return out, nil
}

func (s *Server) updateEquationSubgraph(ctx context.Context, req *mcpsdk.CallToolRequest, in UpdateEquationSubgraphInput) (*mcpsdk.CallToolResult, UpdateEquationSubgraphOutput, error) {
	patching := len(in.Statements) > 0 || len(in.Remove) > 0
	if in.Equation == "" && !patching && len(in.Types) == 0 && len(in.Constants) == 0 {
		out, err := s.readEquation(in.Id)
		return nil, out, err
	}
	if in.Equation != "" && patching {
		return nil, UpdateEquationSubgraphOutput{}, fmt.Errorf("pass equation or statements/remove, not both")
	}

	var out UpdateEquationSubgraphOutput
	var err error
	s.atomic(&err, func() error {
		child, e := s.graph.SubGraphInstance(in.Id)
		if e != nil {
			return e
		}
		source := in.Equation
		if source == "" {
			stored, ok := storedEquation(child)
			if !ok {
				return fmt.Errorf("subgraph %q has no stored equation", in.Id)
			}
			if source, e = patchEquation(stored, in.Statements, in.Remove); e != nil {
				return e
			}
		}
		statements, e := parseEquations(source)
		if e != nil {
			return fmt.Errorf("parsing equation %q: %w", source, e)
		}

		oldInputs, oldOutputs := boundaryPorts(child)
		types := equationTypes(statements, in.Types, oldInputs, oldOutputs)
		if e := checkEquationTypes(statements, types); e != nil {
			return e
		}
		spec := equationSpec{
			statements:      statements,
			types:           types,
			constants:       in.Constants,
			existingInputs:  map[string]string{},
			existingOutputs: map[string]string{},
		}

		named := map[string]bool{}
		for _, name := range freeVariables(statements) {
			named[name] = true
		}
		for _, statement := range statements {
			named[statement.output] = true
		}
		// Deleting before the rebuild re-types kept outputs and drops wires downstream.
		var clashing, doomed []string
		// A dropped wire names the port by its boundary id once the boundary is gone.
		removedNames := map[string]string{}
		doom := func(name, id string) {
			removedNames[id] = name
			if named[name] {
				clashing = append(clashing, id)
			} else {
				doomed = append(doomed, id)
			}
		}

		for _, name := range freeVariables(statements) {
			if old, ok := oldInputs[name]; ok && old.portType == equationPortType(types, name) {
				spec.existingInputs[name] = old.id
			} else {
				out.AddedInputs = append(out.AddedInputs, name)
			}
		}
		for _, name := range slices.Sorted(maps.Keys(oldInputs)) {
			if _, kept := spec.existingInputs[name]; !kept {
				out.RemovedInputs = append(out.RemovedInputs, name)
				doom(name, oldInputs[name].id)
			}
		}

		conditions := conditionOutputs(statements, types)
		for _, statement := range statements {
			if old, ok := oldOutputs[statement.output]; ok && old.portType == outputPortType(types, conditions, statement.output) {
				spec.existingOutputs[statement.output] = old.id
			}
		}
		if len(oldOutputs) == 1 && len(statements) == 1 && len(spec.existingOutputs) == 0 {
			oldName := slices.Collect(maps.Keys(oldOutputs))[0]
			if old := oldOutputs[oldName]; old.portType == outputPortType(types, conditions, statements[0].output) {
				if e := child.RenameBoundaryPorts(map[string]string{old.id: statements[0].output}); e != nil {
					return e
				}
				spec.existingOutputs[statements[0].output] = old.id
				out.RenamedOutput = oldName + " -> " + statements[0].output
				delete(oldOutputs, oldName)
			}
		}
		for _, name := range slices.Sorted(maps.Keys(oldOutputs)) {
			if _, kept := spec.existingOutputs[name]; !kept {
				out.RemovedOutputs = append(out.RemovedOutputs, name)
				doom(name, oldOutputs[name].id)
			}
		}

		var interior []string
		for _, id := range slices.Sorted(slices.Values(child.NodeIds())) {
			if _, isBoundary := subgraph.IsBoundaryNode(child.Node(id)); !isBoundary {
				interior = append(interior, id)
			}
		}
		deleteNodes := func(ids []string) error {
			for _, id := range ids {
				if !child.HasNodeWithId(id) {
					continue
				}
				if _, e := child.DeleteNodeById(id); e != nil {
					return e
				}
			}
			return nil
		}

		oldDefaults := map[string][]byte{}
		for name := range in.Constants {
			if id, kept := spec.existingInputs[name]; kept {
				if held, e := child.LiteralValue(id, subgraph.DefaultPortName); e == nil {
					oldDefaults[id] = held
				}
			}
		}
		var built equationBuild
		dropped, e := s.graph.Batch(func() (e error) {
			if e = deleteNodes(clashing); e != nil {
				return e
			}
			if built, e = buildEquation(child, spec); e != nil {
				return e
			}
			return deleteNodes(append(interior, doomed...))
		})
		if e == nil {
			e = refuseInteriorDrops(in.Id, dropped)
		}
		if e != nil {
			return fmt.Errorf("compiling equation %q: %w", source, e)
		}
		for _, drop := range dropped {
			if name, ok := removedNames[drop.ToPort]; ok {
				drop.ToPort = name
			}
			out.Dropped = append(out.Dropped, summarizeDrops([]graph.DroppedEdge{drop})...)
		}
		for id, was := range oldDefaults {
			if e := s.carryDefault(in.Id, id, was); e != nil {
				return e
			}
		}
		storeEquation(child, statements)

		out.SubgraphId = in.Id
		out.Inputs = built.inputs
		out.Output = built.outputs[0]
		if len(built.outputs) > 1 {
			out.Outputs = built.outputs
		}
		out.Warning = built.warning()
		if len(out.AddedInputs) > 0 {
			out.Instances = subgraphInstances(s.graph.Schema(), in.Id)
		}
		return nil
	})
	return nil, out, err
}

func (s *Server) registerEquationTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "create_equation_subgraph",
		Description: "Compile a math equation (e.g. \"c = sqrt(a^2 + b^2)\") directly into a new subgraph with one float64 boundary input per free variable and a boundary output for the result — instead of hand-wiring several AddNode/SubtractNode/MultiplyNode/DivideNode/etc. one at a time and connecting them together. Use this for an expression combining two or more operations (distances, ratios, derived dimensions, easing). Several statements ('a = ...; b = ...') make one helper with several outputs over shared inputs. Name tunable constants as identifiers and pass 'constants' so they become inputs with defaults, not literals buried in the definition. For a single operation - one multiply, one add, one subtract, one divide - use the corresponding plain node (MultiplyNode, AddNode, SubtractNode, DivideNode) directly instead: creating and naming a whole subgraph for one operation is unnecessary overhead a single node avoids. To change the formula later, update_equation_subgraph, not a new subgraph.",
	}, s.createEquationSubgraph)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "update_equation_subgraph",
		Description: "Recompile an equation subgraph in place: same id, same instances, same wiring on every boundary port whose name still appears in the new equation. New free variables become new (unwired) inputs; inputs the equation no longer uses are removed and their outside wires reported in 'dropped'. The one call for a sign fix, a clamp, a changed constant or an extra term — never create a new subgraph, rewire its consumers and delete the old one. To change a few statements of a long equation, pass statements {name: expression} and remove [names] rather than resending the whole equation; pass only id to read it back. Port types and input defaults carry over unless named. The interior is rebuilt from the equation, so nodes added to it by hand are lost.",
	}, s.updateEquationSubgraph)
}
