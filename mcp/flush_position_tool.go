package mcp

import (
	"context"
	"fmt"

	"github.com/EliCDavis/polyform/generator/subgraph"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type CreateFlushPositionSubgraphInput struct {
	Id          string `json:"id" jsonschema:"unique id for the new subgraph"`
	Name        string `json:"name,omitempty" jsonschema:"human-readable display name; defaults to a generic name"`
	Description string `json:"description,omitempty"`
}

type CreateFlushPositionSubgraphOutput struct {
	SubgraphId string   `json:"subgraphId"`
	Inputs     []string `json:"inputs" jsonschema:"boundary input names, in the order they must be wired: A Position (float64, A's position on the axis B is being placed along), A Half Size (float64, A's half-extent on that same axis), B Half Size (float64, B's half-extent on that same axis), Direction (float64, +1 if B sits on A's positive side along this axis, -1 if on the negative side)"`
	Output     string   `json:"output" jsonschema:"boundary output name: Position (float64) - wire this into the matching X/Y/Z component of B's own translation; the other two axes are whatever else positions B on them, computed or set independently"`
}

// createFlushPositionSubgraph builds a reusable subgraph computing
// A Position + Direction*(A Half Size + B Half Size) - the "never freehand
// a relative position" formula every part placed flush against another
// part needs, on one axis at a time. This exists because that formula was
// being hand-derived via create_equation_subgraph every time a part
// attached to another, and "forgetting B's own half-size" is the
// documented easy mistake - it still looks plausible in a render.
func (s *Server) createFlushPositionSubgraph(ctx context.Context, req *mcpsdk.CallToolRequest, in CreateFlushPositionSubgraphInput) (*mcpsdk.CallToolResult, CreateFlushPositionSubgraphOutput, error) {
	var out CreateFlushPositionSubgraphOutput
	var err error
	s.atomic(&err, func() error {
		name := in.Name
		if name == "" {
			name = "Flush Position"
		}

		if e := s.graph.CreateSubGraph(in.Id, name, in.Description); e != nil {
			return e
		}
		child, e := s.graph.SubGraphInstance(in.Id)
		if e != nil {
			return e
		}

		_, aPosID, e := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, tcFloatPortType)
		if e != nil {
			return fmt.Errorf("create A Position input: %w", e)
		}
		if e := child.SetBoundaryNodeInfo(aPosID, "A Position"); e != nil {
			return e
		}

		_, aHalfID, e := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, tcFloatPortType)
		if e != nil {
			return fmt.Errorf("create A Half Size input: %w", e)
		}
		if e := child.SetBoundaryNodeInfo(aHalfID, "A Half Size"); e != nil {
			return e
		}

		_, bHalfID, e := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, tcFloatPortType)
		if e != nil {
			return fmt.Errorf("create B Half Size input: %w", e)
		}
		if e := child.SetBoundaryNodeInfo(bHalfID, "B Half Size"); e != nil {
			return e
		}

		_, dirID, e := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, tcFloatPortType)
		if e != nil {
			return fmt.Errorf("create Direction input: %w", e)
		}
		if e := child.SetBoundaryNodeInfo(dirID, "Direction"); e != nil {
			return e
		}

		// AddNode(A Half Size, B Half Size) -> sum of both halves
		_, sumID, e := child.CreateNode(eqAddNodeType)
		if e != nil {
			return e
		}
		child.ConnectNodes(aHalfID, "Value", sumID, "Values")
		child.ConnectNodes(bHalfID, "Value", sumID, "Values")

		// MultiplyNode(Direction, sum of halves) -> signed offset
		_, offsetID, e := child.CreateNode(eqMulNodeType)
		if e != nil {
			return e
		}
		child.ConnectNodes(dirID, "Value", offsetID, "Values")
		child.ConnectNodes(sumID, "Float", offsetID, "Values")

		// AddNode(A Position, signed offset) -> final position
		_, finalID, e := child.CreateNode(eqAddNodeType)
		if e != nil {
			return e
		}
		child.ConnectNodes(aPosID, "Value", finalID, "Values")
		child.ConnectNodes(offsetID, "Float", finalID, "Values")

		_, posOutID, e := child.CreateBoundaryNode(subgraph.OutputNodeTypeKey, tcFloatPortType)
		if e != nil {
			return fmt.Errorf("create Position output: %w", e)
		}
		if e := child.SetBoundaryNodeInfo(posOutID, "Position"); e != nil {
			return e
		}
		child.ConnectNodes(finalID, "Float", posOutID, "Value")

		out.SubgraphId = in.Id
		out.Inputs = []string{"A Position", "A Half Size", "B Half Size", "Direction"}
		out.Output = "Position"
		return nil
	})
	return nil, out, err
}

func (s *Server) registerFlushPositionTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "create_flush_position_subgraph",
		Description: "Creates a reusable subgraph computing A Position + Direction*(A Half Size + B Half Size) - the position B needs on one axis to sit exactly flush against reference part A (a headlight on the front of a car body, a doorknob's edge against a door's edge), accounting for both parts' half-sizes so B doesn't overlap or gap. Works one axis at a time: after creating it, instantiate_subgraph, wire the four boundary inputs (A Position/A Half Size/B Half Size on the relevant axis, Direction +1 or -1 for which side of A that B sits on), then wire the Position output into the matching component of B's own translation. Use this instead of hand-deriving the formula via create_equation_subgraph each time a part is placed relative to another.",
	}, s.createFlushPositionSubgraph)
}
