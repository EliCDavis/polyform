package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/modeling"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type EvaluateNodeInput struct {
	NodeId     string `json:"nodeId"`
	Port       string `json:"port,omitempty" jsonschema:"output port to evaluate; omit when the node has exactly one"`
	Scope      string `json:"scope,omitempty" jsonschema:"subgraph id the node lives in; omit for the root graph. Inside a definition, boundary inputs read as zero - pass instanceId to evaluate a live instance's copy instead"`
	InstanceId string `json:"instanceId,omitempty" jsonschema:"a subgraph instance in the root graph (or a path outer/inner into a nested one) whose interior copy of nodeId to evaluate, with that instance's real inputs"`
	MaxItems   int    `json:"maxItems,omitempty" jsonschema:"for an array value, how many leading entries to return; defaults to 64. The full length is always reported."`
}

type EvaluateNodeOutput struct {
	Port    string `json:"port"`
	Type    string `json:"type" jsonschema:"Go type of the value"`
	Value   any    `json:"value,omitempty" jsonschema:"the value as JSON: numbers, strings, booleans, vectors ({x,y,z}), quaternions, transforms, colors, and arrays of those (truncated to maxItems)"`
	Length  int    `json:"length,omitempty" jsonschema:"for an array: its full length, even when value is truncated"`
	Summary string `json:"summary,omitempty" jsonschema:"for a value with no useful JSON form (a distance field, a mesh, a model): what it is and where to look instead"`

	Texture *TextureStats `json:"texture,omitempty" jsonschema:"for a texture: its size and, per channel, the min/max/mean/median and a 10-bucket histogram - how much of the texture a pattern actually covers, which a picture can only be eyeballed for"`

	Warning string `json:"warning,omitempty"`
}

const evaluateDefaultMaxItems = 64

func (s *Server) evaluateNode(ctx context.Context, req *mcpsdk.CallToolRequest, in EvaluateNodeInput) (*mcpsdk.CallToolResult, EvaluateNodeOutput, error) {
	var out EvaluateNodeOutput
	var err error
	s.atomic(&err, func() error {
		inst, contextWarning, e := s.resolveFieldGraph(in.Scope, in.InstanceId)
		if e != nil {
			return e
		}
		out.Warning = contextWarning

		node := inst.Node(in.NodeId)
		if node == nil {
			return s.explainMissingNode(in.Scope, in.NodeId, fmt.Errorf("no node exists with id %q", in.NodeId))
		}

		outputs := node.Outputs()
		portName := in.Port
		if portName == "" {
			if len(outputs) != 1 {
				return fmt.Errorf("node %q has %d output ports (%s); say which with 'port'", in.NodeId, len(outputs), strings.Join(sortedOutputNames(outputs), ", "))
			}
			for name := range outputs {
				portName = name
			}
		} else {
			portName = resolveOutputPortName(node, portName)
		}
		port, ok := outputs[portName]
		if !ok {
			return fmt.Errorf("node %q has no output port %q; it has %s", in.NodeId, in.Port, strings.Join(sortedOutputNames(outputs), ", "))
		}
		out.Port = portName

		method := reflect.ValueOf(port).MethodByName("Value")
		if !method.IsValid() || method.Type().NumIn() != 0 || method.Type().NumOut() != 1 {
			return fmt.Errorf("port %q does not expose a value", portName)
		}
		var value any
		func() {
			defer func() {
				if r := recover(); r != nil {
					e = fmt.Errorf("evaluating %s.%s panicked: %v", in.NodeId, portName, r)
				}
			}()
			value = method.Call(nil)[0].Interface()
		}()
		if e != nil {
			return e
		}
		out.Type = method.Type().Out(0).String()

		maxItems := in.MaxItems
		if maxItems <= 0 {
			maxItems = evaluateDefaultMaxItems
		}
		describeValue(&out, value, maxItems)
		return nil
	})
	return nil, out, err
}

func describeValue(out *EvaluateNodeOutput, value any, maxItems int) {
	switch v := value.(type) {
	case nil:
		out.Summary = "nil - nothing is wired in, or the node produced no value"
		return
	case modeling.Mesh:
		out.Summary = fmt.Sprintf("a mesh with %d triangles; use describe_mesh for its bounds and pieces", v.PrimitiveCount())
		return
	case *gltf.PolyformModel:
		out.Summary = "a glTF model; use describe_mesh on its mesh, or render_preview"
		return
	}

	if stats := textureStats(value); stats != nil {
		out.Texture = stats
		out.Summary = stats.summary()
		return
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Func:
		out.Summary = "a function (a distance or color field); use sample_field / raycast_field to query it at points"
		return
	case reflect.Slice, reflect.Array:
		out.Length = rv.Len()
		if rv.Len() > 0 && rv.Index(0).Kind() == reflect.Func {
			out.Summary = fmt.Sprintf("%d fields; sample them one at a time", rv.Len())
			return
		}
		if rv.Len() > maxItems {
			rv = rv.Slice(0, maxItems)
			value = rv.Interface()
		}
	}

	// Round-trip through JSON so the reply carries plain maps and
	// numbers, not Go types the schema can't describe.
	data, err := json.Marshal(value)
	if err != nil || len(data) > 64*1024 {
		out.Summary = fmt.Sprintf("%.200v", value)
		return
	}
	var generic any
	if json.Unmarshal(data, &generic) != nil {
		out.Summary = fmt.Sprintf("%.200v", value)
		return
	}
	out.Value = generic
}

func (s *Server) registerEvaluateTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "evaluate_node",
		Description: "Read the current value of a node's output port as JSON: a computed position, a chained transform, a spline length, a color, a point array. The exact way to answer 'where did this end up' for anything that is a number or vector rather than geometry - a head's world center after yaw and tilt, a paw anchor after a pose - instead of re-deriving it by hand from the parameters. Fields, meshes and models come back as a summary pointing at sample_field / describe_mesh / render_preview. Evaluating inside a subgraph definition reads its boundary inputs as zero; pass instanceId for a live instance's values.",
	}, s.evaluateNode)
}
