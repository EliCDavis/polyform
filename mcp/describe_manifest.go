package mcp

import (
	"context"
	"fmt"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type DescribeManifestInput struct {
	NodeId string `json:"nodeId" jsonschema:"id of a gltf.ManifestNode (or any node with a 'Models' array input)"`
	Scope  string `json:"scope,omitempty"`
}

type ManifestModelSummary struct {
	Path      string `json:"path" jsonschema:"model names from the top-level entry down, joined with /"`
	Triangles int    `json:"triangles" jsonschema:"this model's own mesh, not counting children; times its Gpu Instances count"`
	Instances int    `json:"instances,omitempty" jsonschema:"Gpu Instances count when the model is instanced"`
	Empty     bool   `json:"empty,omitempty" jsonschema:"true when the model carries no mesh (a pure group, or a disabled gate)"`

	WorldPosition Vector3Input `json:"worldPosition" jsonschema:"where this model's origin lands after every parent transform - the frame a child part (a collar under a posed head) is actually placed in"`
	WorldRotation [4]float64   `json:"worldRotation" jsonschema:"composed rotation as a quaternion (x, y, z, w)"`
	WorldScale    Vector3Input `json:"worldScale"`
}

type DescribeManifestOutput struct {
	Models         []ManifestModelSummary `json:"models" jsonschema:"every model in the tree, depth first"`
	TotalTriangles int                    `json:"totalTriangles" jsonschema:"what a viewer or exporter will draw for this manifest, instances included"`
	EmptyModels    int                    `json:"emptyModels" jsonschema:"models with no mesh; a group is expected to be one, a part is not"`
}

func (s *Server) describeManifest(ctx context.Context, req *mcpsdk.CallToolRequest, in DescribeManifestInput) (*mcpsdk.CallToolResult, DescribeManifestOutput, error) {
	var out DescribeManifestOutput
	var err error
	s.atomic(&err, func() error {
		inst, e := s.resolveScope(in.Scope)
		if e != nil {
			return e
		}
		node := inst.Node(in.NodeId)
		if node == nil {
			return fmt.Errorf("no node exists with id %q", in.NodeId)
		}
		modelsInput, ok := node.Inputs()["Models"]
		if !ok {
			return fmt.Errorf("node %q has no \"Models\" input", in.NodeId)
		}
		arrInput, ok := modelsInput.(nodes.ArrayValueInputPort)
		if !ok {
			return fmt.Errorf("node %q's \"Models\" input is not an array port", in.NodeId)
		}

		out.Models = []ManifestModelSummary{}
		for _, port := range arrInput.Value() {
			modelOut, ok := port.(nodes.Output[*gltf.PolyformModel])
			if !ok || modelOut.Value() == nil {
				continue
			}
			summarizeModel(modelOut.Value(), "", trs.Identity(), &out)
		}
		return nil
	})
	return nil, out, err
}

func summarizeModel(model *gltf.PolyformModel, parentPath string, parent trs.TRS, out *DescribeManifestOutput) {
	path := model.Name
	if parentPath != "" {
		path = parentPath + "/" + model.Name
	}
	world := parent
	if model.TRS != nil {
		if composed, err := trs.FromMatrix(parent.Multiply(*model.TRS)); err == nil {
			world = composed
		}
	}
	rot := world.Rotation()
	summary := ManifestModelSummary{
		Path:          path,
		WorldPosition: Vector3Input{world.Position().X(), world.Position().Y(), world.Position().Z()},
		WorldRotation: [4]float64{rot.Dir().X(), rot.Dir().Y(), rot.Dir().Z(), rot.W()},
		WorldScale:    Vector3Input{world.Scale().X(), world.Scale().Y(), world.Scale().Z()},
	}
	if model.Mesh != nil && model.Mesh.HasFloat3Attribute(modeling.PositionAttribute) {
		summary.Triangles = model.Mesh.PrimitiveCount()
		if n := len(model.GpuInstances); n > 0 {
			summary.Instances = n
			summary.Triangles *= n
		}
	} else {
		summary.Empty = true
		out.EmptyModels++
	}
	out.TotalTriangles += summary.Triangles
	out.Models = append(out.Models, summary)
	for _, child := range model.Children {
		summarizeModel(child, path, world, out)
	}
}

func (s *Server) registerManifestTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "describe_manifest",
		Description: "Per-model triangle counts for everything a manifest will export, with the total, without rendering anything. The way to compare a high-poly and a low-poly manifest, find which part dominates the budget, or confirm a gated/disabled part really dropped out (it shows as empty). Instanced models count once per instance, as a viewer draws them.",
	}, s.describeManifest)
}
