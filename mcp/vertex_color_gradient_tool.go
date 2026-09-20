package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/subgraph"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Exact registered type/port-type keys the vertex-color-gradient subgraph
// wires together, verified against a running type registry rather than
// guessed — see mcp/vertex_color_gradient_tool_test.go.
const (
	vcgSelectFromMeshType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling.SelectFromMeshNode]"
	vcgSelectArrayType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/vector3.SelectArray[float64]]"
	vcgMinArrayType       = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.MinArrayNode[float64]]"
	vcgMaxArrayType       = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.MaxArrayNode[float64]]"
	vcgRemapArrayType     = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.RemapToArrayNode[float64]]"
	vcgInterpolateType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/drawing/coloring.InterpolateToArrayNode]"
	vcgToVectorArrayType  = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/drawing/coloring.ToVectorArrayNode]"
	vcgSetAttribute3DType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling.SetAttribute3DNode]"

	vcgMeshPortType  = "github.com/EliCDavis/polyform/modeling.Mesh"
	vcgColorPortType = "github.com/EliCDavis/polyform/drawing/coloring.Color"
	vcgBoolParamType = "github.com/EliCDavis/polyform/generator/parameter.Value[bool]"
	vcgStrParamType  = "github.com/EliCDavis/polyform/generator/parameter.Value[string]"
)

type CreateVertexColorGradientSubgraphInput struct {
	Id          string `json:"id" jsonschema:"unique id for the new subgraph"`
	Name        string `json:"name,omitempty" jsonschema:"human-readable display name; defaults to a generic name"`
	Description string `json:"description,omitempty"`
	Axis        string `json:"axis" jsonschema:"which world-space axis the gradient runs along - 'X', 'Y', or 'Z' (case-insensitive). 'Y' is the usual choice for a belly-to-back fade on a standing creature."`
}

type CreateVertexColorGradientSubgraphOutput struct {
	SubgraphId string   `json:"subgraphId"`
	Inputs     []string `json:"inputs" jsonschema:"boundary input names, in the order they must be wired: Mesh (the mesh to color), Color A (the color at the Axis's minimum, e.g. the belly), Color B (the color at the Axis's maximum, e.g. the back)"`
	Output     string   `json:"output" jsonschema:"boundary output name: Mesh (the same mesh, with a per-vertex \"Color\" attribute set)"`
}

// createVertexColorGradientSubgraph builds a reusable subgraph that colors
// a mesh with a two-color gradient along one axis in one call:
// SelectFromMeshNode -> SelectArray (pick the axis) -> Min/MaxArrayNode
// (the mesh's own actual extent on that axis, not a guessed range) ->
// RemapToArrayNode (-> 0..1) -> InterpolateToArrayNode -> ToVectorArrayNode
// -> SetAttribute3DNode. This exists because every organic build this
// session hand-wired this exact 6-node chain from scratch, and because
// hand-picking InMin/InMax (instead of deriving them from the mesh) is
// exactly the documented bug where a gradient tuned for the torso clips
// newly-added limbs to a solid color.
func (s *Server) createVertexColorGradientSubgraph(ctx context.Context, req *mcpsdk.CallToolRequest, in CreateVertexColorGradientSubgraphInput) (*mcpsdk.CallToolResult, CreateVertexColorGradientSubgraphOutput, error) {
	var out CreateVertexColorGradientSubgraphOutput
	var err error
	s.atomic(&err, func() error {
		axisPort, e := vcgAxisPort(in.Axis)
		if e != nil {
			return e
		}

		name := in.Name
		if name == "" {
			name = "Vertex Color Gradient"
		}

		if e := s.graph.CreateSubGraph(in.Id, name, in.Description); e != nil {
			return e
		}
		child, e := s.graph.SubGraphInstance(in.Id)
		if e != nil {
			return e
		}

		_, meshID, e := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, vcgMeshPortType)
		if e != nil {
			return fmt.Errorf("create Mesh input: %w", e)
		}
		if e := child.SetBoundaryNodeInfo(meshID, "Mesh"); e != nil {
			return e
		}

		_, colorAID, e := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, vcgColorPortType)
		if e != nil {
			return fmt.Errorf("create Color A input: %w", e)
		}
		if e := child.SetBoundaryNodeInfo(colorAID, "Color A"); e != nil {
			return e
		}

		_, colorBID, e := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, vcgColorPortType)
		if e != nil {
			return fmt.Errorf("create Color B input: %w", e)
		}
		if e := child.SetBoundaryNodeInfo(colorBID, "Color B"); e != nil {
			return e
		}

		// SelectFromMeshNode(Mesh) -> Position
		_, selectMeshID, e := child.CreateNode(vcgSelectFromMeshType)
		if e != nil {
			return e
		}
		child.ConnectNodes(meshID, "Value", selectMeshID, "Mesh")

		// SelectArray[float64](Position) -> X/Y/Z
		_, selectAxisID, e := child.CreateNode(vcgSelectArrayType)
		if e != nil {
			return e
		}
		child.ConnectNodes(selectMeshID, "Position", selectAxisID, "In")

		// Min/MaxArrayNode(axis array) -> the mesh's own actual range, not a
		// hand-picked one.
		_, minID, e := child.CreateNode(vcgMinArrayType)
		if e != nil {
			return e
		}
		child.ConnectNodes(selectAxisID, axisPort, minID, "In")

		_, maxID, e := child.CreateNode(vcgMaxArrayType)
		if e != nil {
			return e
		}
		child.ConnectNodes(selectAxisID, axisPort, maxID, "In")

		outMinID, e := createFloatLiteral(child, 0)
		if e != nil {
			return e
		}
		outMaxID, e := createFloatLiteral(child, 1)
		if e != nil {
			return e
		}
		clampID, e := createBoolLiteral(child, true)
		if e != nil {
			return e
		}

		// RemapToArrayNode(axis array, mesh's own min/max -> 0..1)
		_, remapID, e := child.CreateNode(vcgRemapArrayType)
		if e != nil {
			return e
		}
		child.ConnectNodes(selectAxisID, axisPort, remapID, "Value")
		child.ConnectNodes(minID, "Float 64", remapID, "In Min")
		child.ConnectNodes(maxID, "Float 64", remapID, "In Max")
		child.ConnectNodes(outMinID, "Value", remapID, "Out Min")
		child.ConnectNodes(outMaxID, "Value", remapID, "Out Max")
		child.ConnectNodes(clampID, "Value", remapID, "Clamp")

		// InterpolateToArrayNode(Color A, Color B, 0..1 factors) -> []Color
		_, interpID, e := child.CreateNode(vcgInterpolateType)
		if e != nil {
			return e
		}
		child.ConnectNodes(colorAID, "Value", interpID, "A")
		child.ConnectNodes(colorBID, "Value", interpID, "B")
		child.ConnectNodes(remapID, "Out", interpID, "Time")

		// ToVectorArrayNode([]Color) -> []vector3.Float64
		_, toVecID, e := child.CreateNode(vcgToVectorArrayType)
		if e != nil {
			return e
		}
		child.ConnectNodes(interpID, "Out", toVecID, "In")

		attrID, e := createStringLiteral(child, "Color")
		if e != nil {
			return e
		}

		// SetAttribute3DNode(Mesh, "Color", []vector3.Float64) -> colored Mesh
		_, setAttrID, e := child.CreateNode(vcgSetAttribute3DType)
		if e != nil {
			return e
		}
		child.ConnectNodes(meshID, "Value", setAttrID, "Mesh")
		child.ConnectNodes(attrID, "Value", setAttrID, "Attribute")
		child.ConnectNodes(toVecID, "Vector 3", setAttrID, "Data")

		_, meshOutID, e := child.CreateBoundaryNode(subgraph.OutputNodeTypeKey, vcgMeshPortType)
		if e != nil {
			return fmt.Errorf("create Mesh output: %w", e)
		}
		if e := child.SetBoundaryNodeInfo(meshOutID, "Mesh"); e != nil {
			return e
		}
		child.ConnectNodes(setAttrID, "Out", meshOutID, "Value")

		out.SubgraphId = in.Id
		out.Inputs = []string{"Mesh", "Color A", "Color B"}
		out.Output = "Mesh"
		return nil
	})
	return nil, out, err
}

// vcgAxisPort maps a case-insensitive axis letter to the SelectArray
// node's corresponding output port name.
func vcgAxisPort(axis string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(axis)) {
	case "X":
		return "X", nil
	case "Y":
		return "Y", nil
	case "Z":
		return "Z", nil
	default:
		return "", fmt.Errorf("axis must be one of \"X\", \"Y\", \"Z\" (got %q)", axis)
	}
}

// createBoolLiteral creates a bool parameter node preset to v inside inst.
func createBoolLiteral(inst *graph.Instance, v bool) (string, error) {
	_, id, err := inst.CreateNode(vcgBoolParamType)
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

// createStringLiteral creates a string parameter node preset to v inside
// inst.
func createStringLiteral(inst *graph.Instance, v string) (string, error) {
	_, id, err := inst.CreateNode(vcgStrParamType)
	if err != nil {
		return "", fmt.Errorf("create literal %q: %w", v, err)
	}
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	if _, err := inst.UpdateParameter(id, data); err != nil {
		return "", fmt.Errorf("set literal %q: %w", v, err)
	}
	return id, nil
}

func (s *Server) registerVertexColorGradientTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "create_vertex_color_gradient_subgraph",
		Description: "Creates a reusable subgraph that colors a mesh with a two-color gradient along one axis in one call - the standard way to color a marched/SDF mesh, which has no UVs to texture. Automatically derives the gradient's range from the mesh's own actual extent on that axis (not a guessed min/max), so the gradient always covers the whole mesh correctly even if geometry is added or resized later - a hand-picked range silently clips newly-added parts to a solid color instead of shading them. After creating it, instantiate_subgraph and wire its three boundary inputs (Mesh, Color A at the axis's minimum, Color B at the axis's maximum), then use its Mesh output like the original mesh.",
	}, s.createVertexColorGradientSubgraph)
}
