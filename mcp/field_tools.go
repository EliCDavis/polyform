package mcp

import (
	"context"
	"fmt"

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
	NodeId string         `json:"nodeId" jsonschema:"id of a node whose output is an SDF field (any math/sdf node - a primitive, or the result of a Union/SmoothUnion/etc.)"`
	Port   string         `json:"port,omitempty" jsonschema:"output port name; defaults to \"Field\", the standard output port name every math/sdf node uses"`
	Scope  string         `json:"scope,omitempty" jsonschema:"the root graph if omitted, or a subgraph id"`
	Points []Vector3Input `json:"points" jsonschema:"world-space points to evaluate the field at"`
}

type SampleFieldOutput struct {
	Values []float64 `json:"values" jsonschema:"the field's signed distance at each input point, same order as points - negative means inside the surface, positive means outside, zero is exactly on it. Compare two fields' values at the same point, or check a point that should be in empty space for an unexpectedly negative value (e.g. diagnosing a suspected smooth-union bridge), without needing a render at all."`
}

func (s *Server) sampleField(ctx context.Context, req *mcpsdk.CallToolRequest, in SampleFieldInput) (*mcpsdk.CallToolResult, SampleFieldOutput, error) {
	var out SampleFieldOutput
	var err error
	s.atomic(&err, func() error {
		inst, e := s.resolveScope(in.Scope)
		if e != nil {
			return e
		}

		node := inst.Node(in.NodeId)

		portName := in.Port
		if portName == "" {
			portName = "Field"
		}
		portName = resolveOutputPortName(node, portName)

		port, ok := node.Outputs()[portName]
		if !ok {
			return fmt.Errorf("node %q has no output port %q", in.NodeId, portName)
		}

		fieldOut, ok := port.(nodes.Output[sample.Vec3ToFloat])
		if !ok {
			typeName := "unknown"
			if typed, ok := port.(nodes.Typed); ok {
				typeName = typed.Type()
			}
			return fmt.Errorf("node %q's %q output is not an SDF field (math/sdf), it's %s", in.NodeId, portName, typeName)
		}

		field := fieldOut.Value()
		values := make([]float64, len(in.Points))
		for i, p := range in.Points {
			values[i] = field(vector3.New(p.X, p.Y, p.Z))
		}
		out.Values = values
		return nil
	})
	return nil, out, err
}

func (s *Server) registerFieldTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "sample_field",
		Description: "Evaluate a math/sdf node's distance field at specific world-space points, returning raw numbers instantly - no marching, no rasterizing, no image to read. The precise, cheap alternative to a render-and-look cycle when debugging SDF geometry: is this point inside or outside the shape, does a suspected smooth-union bridge actually reach into this empty-space region, does field A's value here match field B's (the exact condition that drives smooth-blend behavior). Works on any node whose output is a field - a single primitive, or the combined result of a Union/SmoothUnion/etc.",
	}, s.sampleField)
}
