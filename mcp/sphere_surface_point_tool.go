package mcp

import (
	"context"
	"fmt"

	"github.com/EliCDavis/polyform/generator/subgraph"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Exact registered type keys the sphere-surface-point subgraph wires
// together, verified against a running type registry rather than guessed —
// see mcp/sphere_surface_point_tool_test.go.
const (
	sspNormalizeType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/vector3.Normalize]"
	sspScaleType     = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/vector3.Scale[float64]]"
	sspSumType       = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/vector3.SumNode[float64]]"

	sspVectorPortType = "github.com/EliCDavis/vector/vector3.Vector[float64]"
)

type CreateSphereSurfacePointSubgraphInput struct {
	Id          string `json:"id" jsonschema:"unique id for the new subgraph"`
	Name        string `json:"name,omitempty" jsonschema:"human-readable display name; defaults to a generic name"`
	Description string `json:"description,omitempty"`
}

type CreateSphereSurfacePointSubgraphOutput struct {
	SubgraphId string   `json:"subgraphId"`
	Inputs     []string `json:"inputs" jsonschema:"boundary input names, in the order they must be wired: Center (vector3, the sphere's center), Radius (float64), Direction (vector3, which way from Center to place the point - any length, normalized internally), Embed Fraction (float64, how far toward the surface as a fraction of Radius - 1.0 lands exactly on the surface, less than 1 sits embedded toward the center, more than 1 floats outside)"`
	Output     string   `json:"output" jsonschema:"boundary output name: Position (vector3)"`
}

// createSphereSurfacePointSubgraph builds a reusable subgraph computing a
// point positioned relative to a sphere's surface along a given direction
// - Center + Normalize(Direction) * (Radius * Embed Fraction). This exists
// because a small part meant to sit embedded in a larger round one (an eye
// on a skull, a nose on a muzzle) keeps getting positioned with hand-
// derived offset coefficients that miss the actual surface - e.g. floating
// visibly outside the sphere - needing a render-and-fix cycle to catch.
func (s *Server) createSphereSurfacePointSubgraph(ctx context.Context, req *mcpsdk.CallToolRequest, in CreateSphereSurfacePointSubgraphInput) (*mcpsdk.CallToolResult, CreateSphereSurfacePointSubgraphOutput, error) {
	var out CreateSphereSurfacePointSubgraphOutput
	var err error
	s.atomic(&err, func() error {
		name := in.Name
		if name == "" {
			name = "Sphere Surface Point"
		}

		if e := s.graph.CreateSubGraph(in.Id, name, in.Description); e != nil {
			return e
		}
		child, e := s.graph.SubGraphInstance(in.Id)
		if e != nil {
			return e
		}

		_, centerID, e := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, sspVectorPortType)
		if e != nil {
			return fmt.Errorf("create Center input: %w", e)
		}
		if e := child.SetBoundaryNodeInfo(centerID, "Center"); e != nil {
			return e
		}

		_, radiusID, e := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, tcFloatPortType)
		if e != nil {
			return fmt.Errorf("create Radius input: %w", e)
		}
		if e := child.SetBoundaryNodeInfo(radiusID, "Radius"); e != nil {
			return e
		}

		_, dirID, e := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, sspVectorPortType)
		if e != nil {
			return fmt.Errorf("create Direction input: %w", e)
		}
		if e := child.SetBoundaryNodeInfo(dirID, "Direction"); e != nil {
			return e
		}

		_, embedID, e := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, tcFloatPortType)
		if e != nil {
			return fmt.Errorf("create Embed Fraction input: %w", e)
		}
		if e := child.SetBoundaryNodeInfo(embedID, "Embed Fraction"); e != nil {
			return e
		}

		// Normalize(Direction) -> unit direction
		_, normID, e := child.CreateNode(sspNormalizeType)
		if e != nil {
			return e
		}
		child.ConnectNodes(dirID, "Value", normID, "In")

		// MultiplyNode(Radius, Embed Fraction) -> offset magnitude
		_, magID, e := child.CreateNode(eqMulNodeType)
		if e != nil {
			return e
		}
		child.ConnectNodes(radiusID, "Value", magID, "Values")
		child.ConnectNodes(embedID, "Value", magID, "Values")

		// Scale(unit direction, offset magnitude) -> offset vector
		_, scaleID, e := child.CreateNode(sspScaleType)
		if e != nil {
			return e
		}
		child.ConnectNodes(normID, "Normalized", scaleID, "Vector")
		child.ConnectNodes(magID, "Float", scaleID, "Amount")

		// SumNode(Center, offset vector) -> final position
		_, sumID, e := child.CreateNode(sspSumType)
		if e != nil {
			return e
		}
		child.ConnectNodes(centerID, "Value", sumID, "Values")
		child.ConnectNodes(scaleID, "Float 64", sumID, "Values")

		_, posOutID, e := child.CreateBoundaryNode(subgraph.OutputNodeTypeKey, sspVectorPortType)
		if e != nil {
			return fmt.Errorf("create Position output: %w", e)
		}
		if e := child.SetBoundaryNodeInfo(posOutID, "Position"); e != nil {
			return e
		}
		child.ConnectNodes(sumID, "Out", posOutID, "Value")

		out.SubgraphId = in.Id
		out.Inputs = []string{"Center", "Radius", "Direction", "Embed Fraction"}
		out.Output = "Position"
		return nil
	})
	return nil, out, err
}

func (s *Server) registerSphereSurfacePointTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "create_sphere_surface_point_subgraph",
		Description: "Creates a reusable subgraph computing a point positioned relative to a sphere's surface along a direction - Center + Normalize(Direction) * (Radius * Embed Fraction). Use this for a small part meant to sit embedded in or on a larger round reference part (an eye/nose on a skull, a bolt head on a rounded housing) instead of hand-deriving offset coefficients, which easily misses the actual surface (floats outside, or sinks too far in) and needs a render-and-fix cycle to catch. After creating it, instantiate_subgraph and wire its four boundary inputs (Center, Radius, Direction - any length, normalized internally - and Embed Fraction, where 1.0 lands exactly on the surface), then use its Position output as the part's translation.",
	}, s.createSphereSurfacePointSubgraph)
}
