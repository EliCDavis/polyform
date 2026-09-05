package mcp

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type Vector3Input struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type SampleFieldInput struct {
	NodeId     string         `json:"nodeId" jsonschema:"id of a node whose output is an SDF field (any math/sdf node - a primitive, or the result of a Union/SmoothUnion/etc.)"`
	Port       string         `json:"port,omitempty" jsonschema:"output port name. Usually omit it: the node's single field output is found for you, whatever it is called (a primitive outputs \"Field\", a union \"Union\", a subtraction \"Subtract\", a transform \"Result\"). Only needed on a node with more than one field output, such as MirrorNode."`
	Scope      string         `json:"scope,omitempty" jsonschema:"the root graph if omitted, or a subgraph id. On its own this evaluates the subgraph DEFINITION, whose boundary inputs have nothing wired into them and so read as zero - pass instanceId as well to measure a real instance."`
	InstanceId string         `json:"instanceId,omitempty" jsonschema:"id of a subgraph instance node in the root graph. Evaluates that instance's own copy, with the values wired into its boundary inputs applied. This is what you want whenever the thing you are probing depends on a boundary input."`
	Points     []Vector3Input `json:"points" jsonschema:"world-space points to evaluate the field at"`
}

// FieldSample is one evaluated point. Value and NonFinite are mutually
// exclusive: a field that blows up at a point reports it in NonFinite
// rather than failing the whole call, since "this point is NaN" is usually
// the answer being looked for, not an obstacle to reporting one.
type FieldSample struct {
	Point     Vector3Input `json:"point"`
	Value     float64      `json:"value" jsonschema:"the field's signed distance here - negative is inside the surface, positive is outside, zero is exactly on it. Meaningless when nonFinite is set."`
	NonFinite string       `json:"nonFinite,omitempty" jsonschema:"set to NaN, +Inf or -Inf when the field didn't evaluate to a real number here, which means the field is degenerate at this point rather than merely far from the surface"`
}

type SampleFieldOutput struct {
	Port    string        `json:"port" jsonschema:"the output port actually sampled, which matters when you let it pick the node's field output for you"`
	Samples []FieldSample `json:"samples" jsonschema:"one entry per input point, in the order given. Compare two fields at the same point, or check a point that should be in empty space for an unexpectedly negative value (e.g. diagnosing a suspected smooth-union bridge), without needing a render at all."`
	Warning string        `json:"warning,omitempty" jsonschema:"present only when some point failed to evaluate to a real number, with the likely causes"`

	EvaluationContext string `json:"evaluationContext,omitempty" jsonschema:"present when these numbers came from a subgraph definition rather than a live instance, which makes them unreliable. Read it before drawing any conclusion from the values."`
}

// resolveFieldOutput finds the SDF field output to sample. Naming the
// port is optional because the name varies by node and is the detail
// callers get wrong: a primitive outputs "Field", but a union outputs
// "Union", a subtraction "Subtract", an intersection "Intersection", and
// a translate/transform/repeat "Result". Defaulting to the literal name
// "Field" made sample_field fail on exactly the combinator nodes a
// broken-geometry investigation most wants to look at. Nearly every field
// node has exactly one field-typed output whatever it is called, so when
// the caller doesn't name one, take it.
func resolveFieldOutput(node nodes.Node, nodeID, requested string) (nodes.Output[sample.Vec3ToFloat], string, error) {
	outputs := node.Outputs()

	if requested != "" {
		name := resolveOutputPortName(node, requested)
		port, ok := outputs[name]
		if !ok {
			return nil, "", fmt.Errorf("node %q has no output port %q; it has %s",
				nodeID, requested, strings.Join(sortedOutputNames(outputs), ", "))
		}
		field, ok := port.(nodes.Output[sample.Vec3ToFloat])
		if !ok {
			typeName := "unknown"
			if typed, ok := port.(nodes.Typed); ok {
				typeName = typed.Type()
			}
			return nil, "", fmt.Errorf("node %q's %q output is not an SDF field (math/sdf), it's %s", nodeID, name, typeName)
		}
		return field, name, nil
	}

	found := []string{}
	for name, port := range outputs {
		if _, ok := port.(nodes.Output[sample.Vec3ToFloat]); ok {
			found = append(found, name)
		}
	}
	sort.Strings(found)

	switch len(found) {
	case 0:
		return nil, "", fmt.Errorf("node %q has no SDF field output (math/sdf); its outputs are %s",
			nodeID, strings.Join(sortedOutputNames(outputs), ", "))
	case 1:
		return outputs[found[0]].(nodes.Output[sample.Vec3ToFloat]), found[0], nil
	default:
		return nil, "", fmt.Errorf("node %q has several field outputs (%s); pass \"port\" to say which",
			nodeID, strings.Join(found, ", "))
	}
}

// resolveFieldGraph picks the graph a field query should be evaluated in.
//
// A `scope` alone names the subgraph *definition*, which is shared by every
// instance and has nothing wired into its boundary inputs - so anything
// downstream of one evaluates to a type zero value. That reads as a real
// measurement rather than a missing one (a distance of 0 looks like "on the
// surface"), and it has now twice convinced a build that a working node was
// broken. Naming an instance evaluates its own copy, with its inputs
// applied; without one, the caller is told what they are actually looking
// at.
func (s *Server) resolveFieldGraph(scope, instanceID string) (*graph.Instance, string, error) {
	if instanceID != "" {
		node := s.graph.Node(instanceID)
		if node == nil {
			return nil, "", fmt.Errorf("no node exists with id %q in the root graph; instanceId names a subgraph instance placed there, not a node inside the subgraph", instanceID)
		}
		instance, ok := node.(*graph.SubgraphInstanceNode)
		if !ok {
			return nil, "", fmt.Errorf("node %q is not a subgraph instance, so it has no inputs to evaluate through", instanceID)
		}
		if scope != "" && scope != instance.SubGraphID() {
			return nil, "", fmt.Errorf("instance %q is an instance of subgraph %q, not %q", instanceID, instance.SubGraphID(), scope)
		}
		return instance.LiveGraph(), "", nil
	}

	inst, err := s.resolveScope(scope)
	if err != nil {
		return nil, "", err
	}
	if scope == "" {
		return inst, "", nil
	}

	// Warn only when there is actually something better to look at.
	var instances []string
	for _, id := range s.graph.NodeIds() {
		if sub, ok := s.graph.Node(id).(*graph.SubgraphInstanceNode); ok && sub.SubGraphID() == scope {
			instances = append(instances, id)
		}
	}
	if len(instances) == 0 {
		return inst, "", nil
	}
	sort.Strings(instances)
	return inst, fmt.Sprintf(
		"These numbers come from the %q definition, where nothing is wired into the boundary inputs, so anything downstream of one evaluated as a zero value rather than the value a real instance supplies - a distance of 0 here means 'no input', not 'on the surface'. Pass instanceId to evaluate a live instance instead: %s.",
		scope, strings.Join(instances, ", ")), nil
}

func sortedOutputNames(outputs map[string]nodes.OutputPort) []string {
	names := make([]string, 0, len(outputs))
	for name := range outputs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (s *Server) sampleField(ctx context.Context, req *mcpsdk.CallToolRequest, in SampleFieldInput) (*mcpsdk.CallToolResult, SampleFieldOutput, error) {
	var out SampleFieldOutput
	var err error
	s.atomic(&err, func() error {
		inst, contextWarning, e := s.resolveFieldGraph(in.Scope, in.InstanceId)
		if e != nil {
			return e
		}

		node := inst.Node(in.NodeId)
		if node == nil {
			return fmt.Errorf("no node exists with id %q", in.NodeId)
		}

		fieldOut, portName, e := resolveFieldOutput(node, in.NodeId, in.Port)
		if e != nil {
			return e
		}

		field := fieldOut.Value()
		samples := make([]FieldSample, len(in.Points))
		nonFinite := 0
		for i, p := range in.Points {
			v := field(vector3.New(p.X, p.Y, p.Z))
			sample := FieldSample{Point: p}

			switch {
			case math.IsNaN(v):
				sample.NonFinite = "NaN"
			case math.IsInf(v, 1):
				sample.NonFinite = "+Inf"
			case math.IsInf(v, -1):
				sample.NonFinite = "-Inf"
			default:
				sample.Value = v
			}

			if sample.NonFinite != "" {
				nonFinite++
			}
			samples[i] = sample
		}

		out.Port = portName
		out.Samples = samples
		out.EvaluationContext = contextWarning
		if nonFinite > 0 {
			out.Warning = fmt.Sprintf(
				"%d of %d points did not evaluate to a real number. The field is degenerate there, not merely far from the surface - common causes are a zero or negative radius/scale on a primitive feeding it, a division by zero somewhere in the chain building its position, or a NaN already arriving from an upstream field. Sample the inputs of this node one at a time to find which one is producing it.",
				nonFinite, len(in.Points))
		}
		return nil
	})
	return nil, out, err
}

func (s *Server) registerFieldTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "sample_field",
		Description: "Evaluate a math/sdf node's distance field at world-space points, returning raw numbers - no marching or image to read. The cheap, precise alternative to a render-and-look cycle when debugging SDF geometry: is this point inside the shape, does a smooth-union bridge reach this region, does field A match field B here. Works on any field-output node. NaN/Inf is reported per-point as nonFinite rather than raised, which is usually what a broken-geometry investigation wants.",
	}, s.sampleField)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "raycast_field",
		Description: "Find where an SDF surface actually is, and which way it faces, by marching rays at it. Answers the question that comes up whenever something has to sit ON a blended body - a tooth on a jaw rim, an eye in a socket, a fin on a flank - where the surface is the smooth union of several primitives with a cavity subtracted out and so has no formula to solve. Each hit returns the surface point, the distance, and the outward unit normal: use the point as the part's Translation, and the normal both to orient it and to offset it (point + normal*h sits a part of half-extent h flush, point - normal*d embeds it by d). Use this instead of nudging a hand-typed coordinate between renders; it is exact, and it costs one call rather than a render cycle.",
	}, s.raycastField)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "describe_mesh",
		Description: "Report what a mesh actually is, in numbers: triangle count, bounding box, vertex attributes, and how many separate connected pieces it is in. Answers the questions a render is worst at - is the part there at all, and is it attached - because an absent part among many present ones looks like nothing, and a tooth floating just off a jaw looks fine from every angle that isn't edge-on. A part that failed to wire doesn't change the triangle count; a part that failed to union is its own piece. Also the cheapest proportion check there is: compare the bounding box against the ratios you intended instead of eyeballing a picture. Costs the same evaluation as a render but skips the rasterizing, the PNG and the image tokens, so prefer it whenever the question is factual rather than aesthetic.",
	}, s.describeMesh)
}
