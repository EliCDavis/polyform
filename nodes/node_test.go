package nodes_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

type Repeat struct {
	nodes.StructData[modeling.Mesh]

	Mesh   nodes.NodeOutput[modeling.Mesh]
	Radius nodes.NodeOutput[float64]
	Times  nodes.NodeOutput[int]
}

func (r Repeat) Process() (modeling.Mesh, error) {
	return repeat.Circle(
		r.Mesh.Data(),
		r.Times.Data(),
		r.Radius.Data(),
	), nil
}

func (r *Repeat) Out() nodes.NodeOutput[modeling.Mesh] {
	return &nodes.StructNodeOutput[modeling.Mesh]{r}
}

func TestNodes(t *testing.T) {

	times := nodes.Value(5)

	repeat := Repeat{
		Radius: nodes.Value(15.),
		Times:  nodes.Value(5),
		Mesh: (&Repeat{
			Radius: nodes.Value(5.),
			Times:  times,
			Mesh:   nodes.Value(primitives.UVSphere(1, 10, 10)),
		}).Out(),
	}

	// Stage changes
	out := repeat.Out()

	out.Data()
	times.Set(13)
	out.Data()

	deps := out.Node().Dependencies()
	assert.Len(t, deps, 3)
	// obj.Save("test.obj", repeat.Data())
}
