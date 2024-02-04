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

	times := nodes.Value(5)

	repeat := nodes.Struct(Repeat{
		Radius: nodes.Value(15.),
		Times:  nodes.Value(5),
		Mesh: nodes.Struct(Repeat{
			Radius: nodes.Value(5.),
			Times:  times,
			Mesh:   nodes.Value(primitives.UVSphere(1, 10, 10)),
		}),
	})

	// Stage changes
	times.Set(13)

	repeat.Data()
	repeat.Dependencies()

	// obj.Save("test.obj", repeat.Data())
}
