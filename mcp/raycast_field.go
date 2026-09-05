package mcp

import (
	"context"
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/vector/vector3"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Ray is one query: where does the surface sit along this direction.
type Ray struct {
	Origin    Vector3Input `json:"origin" jsonschema:"world-space point to start from. Somewhere clearly inside or clearly outside the shape is fine; the march works from either side."`
	Direction Vector3Input `json:"direction" jsonschema:"direction to travel, any length - it is normalized for you"`
}

// RayHit is where the surface was found, and which way it faces.
type RayHit struct {
	Hit      bool         `json:"hit" jsonschema:"false when the ray reached maxDistance without crossing the surface"`
	Point    Vector3Input `json:"point,omitempty" jsonschema:"the surface point, ready to use as a part's Translation"`
	Distance float64      `json:"distance,omitempty" jsonschema:"how far along the ray the surface was, in world units"`
	Normal   Vector3Input `json:"normal,omitempty" jsonschema:"unit surface normal pointing out of the shape. Use it to orient a part that should sit against the surface, and to offset along it - Point + Normal*h puts a part of half-extent h flush, Point - Normal*d embeds it by d."`
	Note     string       `json:"note,omitempty" jsonschema:"set when the answer needs qualifying rather than trusting outright"`
}

type RaycastFieldInput struct {
	NodeId      string  `json:"nodeId" jsonschema:"id of a node whose output is an SDF field"`
	Port        string  `json:"port,omitempty" jsonschema:"output port name; omit to use the node's single field output"`
	Scope       string  `json:"scope,omitempty" jsonschema:"the root graph if omitted, or a subgraph id. On its own this marches against the subgraph DEFINITION, whose boundary inputs have nothing wired into them - pass instanceId as well to hit a real instance."`
	InstanceId  string  `json:"instanceId,omitempty" jsonschema:"id of a subgraph instance node in the root graph. Marches that instance's own copy, with the values wired into its boundary inputs applied."`
	Rays        []Ray   `json:"rays" jsonschema:"the rays to march, answered in the order given"`
	MaxDistance float64 `json:"maxDistance,omitempty" jsonschema:"how far to travel before giving up. Defaults to 10."`
}

type RaycastFieldOutput struct {
	Port string   `json:"port" jsonschema:"the output port actually sampled"`
	Hits []RayHit `json:"hits" jsonschema:"one entry per ray, in the order given"`

	EvaluationContext string `json:"evaluationContext,omitempty" jsonschema:"present when these hits came from a subgraph definition rather than a live instance, which makes them unreliable. Read it before drawing any conclusion from them."`
}

// marchRay sphere-traces field from origin along dir. A signed distance
// field says how far it is safe to step, so stepping by the value itself
// converges on the surface without needing a fixed step size fine enough
// to avoid skipping thin features.
func marchRay(field sample.Vec3ToFloat, origin, dir vector3.Float64, maxDistance float64) (float64, bool) {
	const (
		epsilon = 1e-5
		maxStep = 512
	)

	start := field(origin)
	if math.IsNaN(start) || math.IsInf(start, 0) {
		return 0, false
	}
	// Marching from inside is just as valid; travel against the sign so
	// the distance stays a step we can take.
	sign := 1.
	if start < 0 {
		sign = -1
	}

	travelled := 0.
	for i := 0; i < maxStep && travelled < maxDistance; i++ {
		at := origin.Add(dir.Scale(travelled))
		d := field(at) * sign
		if math.IsNaN(d) || math.IsInf(d, 0) {
			return 0, false
		}
		if d < epsilon {
			return travelled, true
		}
		travelled += d
	}
	return 0, false
}

// fieldNormal is the gradient by central differences, which for a
// distance field is the outward surface normal.
func fieldNormal(field sample.Vec3ToFloat, at vector3.Float64) vector3.Float64 {
	const h = 1e-4
	grad := vector3.New(
		field(at.Add(vector3.New(h, 0., 0.)))-field(at.Sub(vector3.New(h, 0., 0.))),
		field(at.Add(vector3.New(0., h, 0.)))-field(at.Sub(vector3.New(0., h, 0.))),
		field(at.Add(vector3.New(0., 0., h)))-field(at.Sub(vector3.New(0., 0., h))),
	)
	if grad.Length() == 0 {
		return vector3.Zero[float64]()
	}
	return grad.Normalized()
}

// raycastField answers "where is the surface, and which way does it face".
//
// Attaching anything to a blended SDF body - a tooth on a jaw, an eye in a
// socket, a fin on a flank - needs a point on a surface that has no closed
// form, because it is the smooth union of several primitives with a cavity
// subtracted out of it. Builds were hand-tuning those coordinates by
// render feedback and shipping teeth that floated off the snout and fins
// that sat proud of the body. The flush and sphere-surface tools only
// cover a flat or spherical reference; this covers the general case.
func (s *Server) raycastField(ctx context.Context, req *mcpsdk.CallToolRequest, in RaycastFieldInput) (*mcpsdk.CallToolResult, RaycastFieldOutput, error) {
	var out RaycastFieldOutput
	var err error
	s.atomic(&err, func() error {
		if len(in.Rays) == 0 {
			return fmt.Errorf("\"rays\" is empty; pass at least one ray")
		}

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
		out.Port = portName
		out.EvaluationContext = contextWarning
		field := fieldOut.Value()

		maxDistance := in.MaxDistance
		if maxDistance <= 0 {
			maxDistance = 10
		}

		out.Hits = make([]RayHit, len(in.Rays))
		for i, ray := range in.Rays {
			dir := vector3.New(ray.Direction.X, ray.Direction.Y, ray.Direction.Z)
			if dir.Length() == 0 {
				out.Hits[i] = RayHit{Note: "direction has zero length, so there is nowhere to march"}
				continue
			}
			dir = dir.Normalized()
			origin := vector3.New(ray.Origin.X, ray.Origin.Y, ray.Origin.Z)

			distance, hit := marchRay(field, origin, dir, maxDistance)
			if !hit {
				out.Hits[i] = RayHit{Note: fmt.Sprintf(
					"no surface within %g units; the ray may be pointing away from the shape, or the field may be degenerate along it", maxDistance)}
				continue
			}

			point := origin.Add(dir.Scale(distance))
			normal := fieldNormal(field, point)
			out.Hits[i] = RayHit{
				Hit:      true,
				Point:    Vector3Input{X: point.X(), Y: point.Y(), Z: point.Z()},
				Distance: distance,
				Normal:   Vector3Input{X: normal.X(), Y: normal.Y(), Z: normal.Z()},
			}
		}
		return nil
	})
	return nil, out, err
}
