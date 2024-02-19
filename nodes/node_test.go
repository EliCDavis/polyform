package nodes_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

type CombineNode struct {
	nodes.StructData[modeling.Mesh]

	Meshes []nodes.NodeOutput[modeling.Mesh]
}

func (cn *CombineNode) Out() nodes.NodeOutput[modeling.Mesh] {
	return nodes.StructNodeOutput[modeling.Mesh]{Definition: cn}
}

func (cn CombineNode) Process() (modeling.Mesh, error) {
	finalMesh := modeling.EmptyMesh(modeling.TriangleTopology)

	for _, n := range cn.Meshes {
		finalMesh = finalMesh.Append(n.Data())
	}

	return finalMesh, nil
}

func TestNodes(t *testing.T) {

	times := nodes.Value(5)

	repeated := &repeat.CircleNode{
		Radius: nodes.Value(15.),
		Times:  nodes.Value(5),
		Mesh: (&repeat.CircleNode{
			Radius: nodes.Value(5.),
			Times:  times,
			Mesh:   nodes.Value(primitives.UVSphere(1, 10, 10)),
		}).Out(),
	}

	combined := CombineNode{
		Meshes: []nodes.NodeOutput[modeling.Mesh]{
			repeated.Out(),
			(&repeat.CircleNode{
				Radius: nodes.Value(5.),
				Times:  times,
				Mesh:   nodes.Value(primitives.UVSphere(1, 10, 10)),
			}).Out(),
		},
	}

	combinedDeps := combined.Out().Node().Dependencies()
	assert.Len(t, combinedDeps, 2)

	// Stage changes
	out := combined.Out()

	out.Data()
	times.Set(13)
	out.Data()

	deps := repeated.Out().Node().Dependencies()
	assert.Len(t, deps, 3)
	// obj.Save("test.obj", repeat.Data())
}
