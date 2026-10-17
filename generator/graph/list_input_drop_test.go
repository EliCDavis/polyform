package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/modeling/sketch"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type centersNode struct{}

func (centersNode) Out(out *nodes.StructOutput[[]vector2.Float64]) {
	out.Set([]vector2.Float64{vector2.New(0., 0.), vector2.New(5., 0.)})
}

const (
	centersType          = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.centersNode]"
	roundedRectangleType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/sketch.RoundedRectangleNode]"
	fillType             = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/sketch.FillNode]"
)

func TestTheEdgeDroppedFromAListInputIsTheOneThatDoesNotFit(t *testing.T) {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[centersNode]](factory)
	refutil.RegisterType[nodes.Struct[sketch.RoundedRectangleNode]](factory)
	refutil.RegisterType[nodes.Struct[sketch.FillNode]](factory)
	instance := graph.New(graph.Config{TypeFactory: factory})

	_, one, err := instance.CreateNode(roundedRectangleType)
	require.NoError(t, err)
	_, centers, err := instance.CreateNode(centersType)
	require.NoError(t, err)
	_, many, err := instance.CreateNode(roundedRectangleType)
	require.NoError(t, err)
	require.NoError(t, instance.ConnectNodes(centers, "Out", many, "Center"))
	fill, fillID, err := instance.CreateNode(fillType)
	require.NoError(t, err)

	err = instance.Batch(func() error {
		require.NoError(t, instance.ConnectNodes(one, "Out", fillID, "Outlines"))
		require.NoError(t, instance.ConnectNodes(many, "Out", fillID, "Outlines"))
		return nil
	})

	var unfit *graph.UnfitEdges
	require.ErrorAs(t, err, &unfit)
	require.Len(t, unfit.Edges, 1)
	assert.Equal(t, many, unfit.Edges[0].From)
	kept := fill.Inputs()["Outlines"].(nodes.ArrayValueInputPort).Value()
	require.Len(t, kept, 1)
	assert.Equal(t, instance.Node(one), kept[0].Node())
}
