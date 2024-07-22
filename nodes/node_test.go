package nodes_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

type CombineNode = nodes.StructNode[modeling.Mesh, CombineData]

type CombineData struct {
	Meshes []nodes.NodeOutput[modeling.Mesh]
}

func (cn CombineData) Process() (modeling.Mesh, error) {
	finalMesh := modeling.EmptyMesh(modeling.TriangleTopology)

	for _, n := range cn.Meshes {
		finalMesh = finalMesh.Append(n.Value())
	}

	return finalMesh, nil
}

func TestNodes(t *testing.T) {

	times := nodes.Value(5)

	repeated := &repeat.CircleNode{
		Data: repeat.CircleNodeData{
			Radius: nodes.Value(15.),
			Times:  nodes.Value(5),
			Mesh: (&repeat.CircleNode{
				Data: repeat.CircleNodeData{
					Radius: nodes.Value(5.),
					Times:  times,
					Mesh:   nodes.Value(primitives.UVSphere(1, 10, 10)),
				},
			}).Out(),
		},
	}

	repeated.
		Out().
		Node().
		SetInput("Times", nodes.Output{NodeOutput: nodes.Value(30)})

	combined := CombineNode{
		Data: CombineData{
			Meshes: []nodes.NodeOutput[modeling.Mesh]{
				repeated.Out(),
				(&repeat.CircleNode{
					Data: repeat.CircleNodeData{
						Radius: nodes.Value(5.),
						Times:  times,
						Mesh:   nodes.Value(primitives.UVSphere(1, 10, 10)),
					},
				}).Out(),
			},
		},
	}

	combinedInputs := combined.Inputs()
	assert.Len(t, combinedInputs, 1)
	assert.Equal(t, "[]github.com/EliCDavis/polyform/modeling.Mesh", combinedInputs[0].Type)

	combinedDeps := combined.Out().Node().Dependencies()
	assert.Len(t, combinedDeps, 2)

	// Stage changes
	out := combined.Out()

	out.Value()
	times.Set(13)
	out.Value()

	deps := repeated.Out().Node().Dependencies()
	assert.Len(t, deps, 3)
	// assert.Equal(t, []nodes.Output{{
	// 	// Name: "Out",
	// 	Type: "github.com/EliCDavis/polyform/modeling.Mesh",
	// }}, combined.Out().Node().Outputs())

	assert.Equal(t, "nodes_test", combined.Path())
	assert.Equal(t, "modeling/repeat", repeated.Path())
	// obj.Save("test.obj", repeat.Value())
}
