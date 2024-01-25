package nodes_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/nodes"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/repeat"
)

type RepeatNodeParameters struct {
	Mesh   nodes.Node[modeling.Mesh]
	Radius nodes.Node[float64]
	Times  nodes.Node[int]
}

type RepeatNode = *nodes.TransformerNode[RepeatNodeParameters, modeling.Mesh]

func Repeat(parameters RepeatNodeParameters) RepeatNode {
	return nodes.Transformer(
		parameters,
		func(in RepeatNodeParameters) (modeling.Mesh, error) {
			return repeat.Circle(
				in.Mesh.Data(),
				in.Times.Data(),
				in.Radius.Data(),
			), nil
		},
	)
}

func TestNodes(t *testing.T) {

	times := nodes.Input(5)

	repeat := Repeat(RepeatNodeParameters{
		Radius: nodes.Input(15.),
		Times:  nodes.Input(5),
		Mesh: Repeat(RepeatNodeParameters{
			Radius: nodes.Input(5.),
			Times:  times,
			Mesh:   nodes.Input(primitives.UVSphere(1, 10, 10)),
		}),
	})

	g := nodes.NewProcessManager()
	g.AddProcessNode(repeat)

	// Stage changes
	times.Set(3)

	// Kick off
	g.Process()

	// obj.Save("test.obj", repeat.Data())
}
