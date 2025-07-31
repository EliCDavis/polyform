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
	Constraints nodes.Output[[]vector2.Float64]
}

func (node BowyerWatsonNode) Out() nodes.StructOutput[modeling.Mesh] {
	if node.Points == nil {
		return nodes.NewStructOutput(modeling.EmptyMesh(modeling.TriangleTopology))
	}

	out := nodes.StructOutput[modeling.Mesh]{}

	val := nodes.TryGetOutputValue(&out, node.Points, nil)
	if len(val) < 3 {
		out.CaptureError(errors.New("require atleast 3 points to run"))
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return out
	}

	contraints := nodes.TryGetOutputValue(&out, node.Constraints, nil)
	if len(contraints) < 3 {
		out.Set(BowyerWatson(val))
		return out
	}

	out.Set(ConstrainedBowyerWatson(
		val,
		[]Constraint{NewConstraint(contraints)},
	))
	return out
}
