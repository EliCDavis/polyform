package triangulation

import (
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

	val := node.Points.Value()
	if len(val) < 3 {
		return nodes.NewStructOutput(modeling.EmptyMesh(modeling.TriangleTopology))
	}

	contraints := nodes.TryGetOutputValue(node.Constraints, nil)
	if len(contraints) < 3 {
		return nodes.NewStructOutput(BowyerWatson(val))
	}

	return nodes.NewStructOutput(ConstrainedBowyerWatson(
		val,
		[]Constraint{NewConstraint(contraints)},
	))
}
