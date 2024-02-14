package nodes_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

func TestNodes(t *testing.T) {

	times := nodes.Value(5)

	repeat := repeat.CircleNode{
		Radius: nodes.Value(15.),
		Times:  nodes.Value(5),
		Mesh: (&repeat.CircleNode{
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
