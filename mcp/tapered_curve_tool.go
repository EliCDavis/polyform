package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/subgraph"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Exact registered type/port-type keys the tapered-curve subgraph wires
// together, verified against a running type registry rather than guessed —
// see mcp/tapered_curve_tool_test.go.
const (
	tcSplineNodeType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/curves.CatmullRomSplineNode]"
	tcLengthNodeType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/curves.LengthNode]"
	tcLinearNodeType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sequence.LinearNode]"
	tcPositionsNodeType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/curves.PositionsForArrayNode]"
	tcTaperedLineType   = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sdf.VaryingRadiusLinesNode]"

	tcPointsArrayPortType = "[]github.com/EliCDavis/vector/vector3.Vector[float64]"
	tcFloatPortType       = "float64"
	tcIntPortType         = "int"
	tcFieldPortType       = "github.com/EliCDavis/polyform/math/sample.Vec3ToFloat"
)

type CreateTaperedCurveSubgraphInput struct {
	Id          string `json:"id" jsonschema:"unique id for the new subgraph"`
	Name        string `json:"name,omitempty" jsonschema:"human-readable display name; defaults to a generic name"`
	Description string `json:"description,omitempty"`
}

type CreateTaperedCurveSubgraphOutput struct {
	SubgraphId string   `json:"subgraphId"`
	Inputs     []string `json:"inputs" jsonschema:"boundary input names, in the order they must be wired: Points ([]vector3, the few control points to pass through), Base Radius (float64, radius at Points[0]), Tip Radius (float64, radius at the last point), Samples (int, how many points to resample the curve into before tapering it - use a number well above len(Points), e.g. 12-20, not equal to it)"`
	Output     string   `json:"output" jsonschema:"boundary output name: Field (a math/sdf field, ready to wire into a Union/SmoothUnionNode like any other math/sdf node output)"`
}

// createTaperedCurveSubgraph builds a reusable subgraph that turns a
// handful of control points into a smoothly tapering SDF field in one
// call: CatmullRomSplineNode -> LengthNode -> LinearNode (dense arc-length
// distances) -> PositionsForArrayNode (resample) -> VaryingRadiusLinesNode
// (tapered union), with a second LinearNode building the matching Radii
// array off the same Samples count. This exists because agents kept
// wiring VaryingRadiusLinesNode.Points directly from a few raw control
// points instead of resampling them first, producing a visibly
// straight-segmented/faceted chain instead of a smooth curve.
func (s *Server) createTaperedCurveSubgraph(ctx context.Context, req *mcpsdk.CallToolRequest, in CreateTaperedCurveSubgraphInput) (*mcpsdk.CallToolResult, CreateTaperedCurveSubgraphOutput, error) {
	var out CreateTaperedCurveSubgraphOutput
	var err error
	s.atomic(&err, func() error {
		name := in.Name
		if name == "" {
			name = "Tapered Curve"
		}

		if e := s.graph.CreateSubGraph(in.Id, name, in.Description); e != nil {
			return e
		}
		child, e := s.graph.SubGraphInstance(in.Id)
		if e != nil {
			return e
		}

		// Boundary inputs.
		_, pointsID, e := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, tcPointsArrayPortType)
		if e != nil {
			return fmt.Errorf("create Points input: %w", e)
		}
		if e := child.SetBoundaryNodeInfo(pointsID, "Points"); e != nil {
			return e
		}

		_, baseRadiusID, e := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, tcFloatPortType)
		if e != nil {
			return fmt.Errorf("create Base Radius input: %w", e)
		}
		if e := child.SetBoundaryNodeInfo(baseRadiusID, "Base Radius"); e != nil {
			return e
		}

		_, tipRadiusID, e := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, tcFloatPortType)
		if e != nil {
			return fmt.Errorf("create Tip Radius input: %w", e)
		}
		if e := child.SetBoundaryNodeInfo(tipRadiusID, "Tip Radius"); e != nil {
			return e
		}

		_, samplesID, e := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, tcIntPortType)
		if e != nil {
			return fmt.Errorf("create Samples input: %w", e)
		}
		if e := child.SetBoundaryNodeInfo(samplesID, "Samples"); e != nil {
			return e
		}

		// CatmullRomSplineNode(Points) -> Spline
		_, splineID, e := child.CreateNode(tcSplineNodeType)
		if e != nil {
			return e
		}
		child.ConnectNodes(pointsID, "Value", splineID, "Points")

		// LengthNode(Spline) -> arc length
		_, lengthID, e := child.CreateNode(tcLengthNodeType)
		if e != nil {
			return e
		}
		child.ConnectNodes(splineID, "Out", lengthID, "Spline")

		// LinearNode(0, length, Samples) -> dense Distances
		zeroID, e := createFloatLiteral(child, 0)
		if e != nil {
			return e
		}
		_, distancesID, e := child.CreateNode(tcLinearNodeType)
		if e != nil {
			return e
		}
		child.ConnectNodes(zeroID, "Value", distancesID, "Start")
		child.ConnectNodes(lengthID, "Out", distancesID, "End")
		child.ConnectNodes(samplesID, "Value", distancesID, "Samples")

		// PositionsForArrayNode(Spline, Distances) -> dense resampled points
		_, positionsID, e := child.CreateNode(tcPositionsNodeType)
		if e != nil {
			return e
		}
		child.ConnectNodes(splineID, "Out", positionsID, "Spline")
		child.ConnectNodes(distancesID, "Out", positionsID, "Distances")

		// LinearNode(Base Radius, Tip Radius, Samples) -> matching Radii,
		// same length as the resampled points for free since it shares the
		// same Samples input.
		_, radiiID, e := child.CreateNode(tcLinearNodeType)
		if e != nil {
			return e
		}
		child.ConnectNodes(baseRadiusID, "Value", radiiID, "Start")
		child.ConnectNodes(tipRadiusID, "Value", radiiID, "End")
		child.ConnectNodes(samplesID, "Value", radiiID, "Samples")

		// VaryingRadiusLinesNode(Points, Radii) -> tapered field
		_, lineID, e := child.CreateNode(tcTaperedLineType)
		if e != nil {
			return e
		}
		child.ConnectNodes(positionsID, "Position", lineID, "Points")
		child.ConnectNodes(radiiID, "Out", lineID, "Radii")

		// Boundary output.
		_, fieldOutID, e := child.CreateBoundaryNode(subgraph.OutputNodeTypeKey, tcFieldPortType)
		if e != nil {
			return fmt.Errorf("create Field output: %w", e)
		}
		if e := child.SetBoundaryNodeInfo(fieldOutID, "Field"); e != nil {
			return e
		}
		child.ConnectNodes(lineID, "Field", fieldOutID, "Value")

		out.SubgraphId = in.Id
		out.Inputs = []string{"Points", "Base Radius", "Tip Radius", "Samples"}
		out.Output = "Field"
		return nil
	})
	return nil, out, err
}

// createFloatLiteral creates a float64 parameter node preset to v inside
// inst, returning its node id.
func createFloatLiteral(inst *graph.Instance, v float64) (string, error) {
	_, id, err := inst.CreateNode(eqFloatParamType)
	if err != nil {
		return "", fmt.Errorf("create literal %v: %w", v, err)
	}
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	if _, err := inst.UpdateParameter(id, data); err != nil {
		return "", fmt.Errorf("set literal %v: %w", v, err)
	}
	return id, nil
}

func (s *Server) registerTaperedCurveTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "create_tapered_curve_subgraph",
		Description: "Creates a reusable subgraph that turns a handful of control points into a smoothly tapering math/sdf field in one call - a tail, tentacle, horn, or any part that should read as a smooth curved line getting thinner (or thicker) along its length. Internally resamples the control points through a spline before tapering, so the result is a smooth curve, not a straight-segmented/faceted chain. After creating it, instantiate_subgraph and wire its four boundary inputs (Points, Base Radius, Tip Radius, Samples), then use its Field output like any other math/sdf field. Prefer this over wiring VaryingRadiusLinesNode directly whenever the input is a handful of hand-placed or posable control points rather than an already-dense point array.",
	}, s.createTaperedCurveSubgraph)
}
