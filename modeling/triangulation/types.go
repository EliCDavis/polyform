package triangulation

import (
	"errors"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector2"
)

func init() {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[BowyerWatsonNode]](factory)
	generator.RegisterTypes(factory)
}

type BowyerWatsonNode struct {
	Points      nodes.Output[[]vector2.Float64]
	Constraints nodes.Output[[]vector2.Float64] `description:"Closed boundary the triangulation must respect. Every edge of it survives in the result, and triangles outside it are discarded. Fewer than 3 points leaves the triangulation unconstrained."`
}

func (node BowyerWatsonNode) Description() string {
	return "Triangulates a set of 2D points into a mesh."
}

func (node BowyerWatsonNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
	if node.Points == nil {
		return
	}

	val := nodes.TryGetOutputValue(out, node.Points, nil)
	if len(val) < 3 {
		out.CaptureError(errors.New("require atleast 3 points to run"))
		return
	}

	constraints := nodes.TryGetOutputValue(out, node.Constraints, nil)
	if len(constraints) < 3 {
		out.Set(BowyerWatson(val))
		return
	}

	mesh, err := ConstrainedDelaunay(val, []Constraint{NewConstraint(constraints)})
	if err != nil {
		out.CaptureError(err)
		return
	}
	out.Set(mesh)
}
