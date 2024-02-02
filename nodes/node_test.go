package nodes_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/polyform/nodes"
)

type Repeat struct {
	Mesh   nodes.Node[modeling.Mesh]
	Radius nodes.Node[float64]
	Times  nodes.Node[int]
}

func (r Repeat) Process() (modeling.Mesh, error) {
	return repeat.Circle(
		r.Mesh.Data(),
		r.Times.Data(),
		r.Radius.Data(),
	), nil
}

func TestNodes(t *testing.T) {

	times := nodes.Input(5)

	repeat := nodes.Struct(Repeat{
		Radius: nodes.Input(15.),
		Times:  nodes.Input(5),
		Mesh: nodes.Struct(Repeat{
			Radius: nodes.Input(5.),
			Times:  times,
			Mesh:   nodes.Input(primitives.UVSphere(1, 10, 10)),
		}),
	})

	// Stage changes
	times.Set(13)

	repeat.Data()
	repeat.Dependencies()

	// obj.Save("test.obj", repeat.Data())
}
