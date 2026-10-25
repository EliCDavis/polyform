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
	unionType            = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/sketch.UnionNode]"
)

func TestTheEdgeDroppedFromAListInputIsTheOneThatDoesNotFit(t *testing.T) {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[centersNode]](factory)
	refutil.RegisterType[nodes.Struct[sketch.RoundedRectangleNode]](factory)
	refutil.RegisterType[nodes.Struct[sketch.FillNode]](factory)
	refutil.RegisterType[nodes.Struct[sketch.UnionNode]](factory)
	instance := graph.New(graph.Config{TypeFactory: factory})

	_, oneOutline, err := instance.CreateNode(roundedRectangleType)
	require.NoError(t, err)
	_, centers, err := instance.CreateNode(centersType)
	require.NoError(t, err)
	_, manyOutlines, err := instance.CreateNode(roundedRectangleType)
	require.NoError(t, err)
	require.NoError(t, instance.ConnectNodes(centers, "Out", manyOutlines, "Center"))

	_, one, err := instance.CreateNode(fillType)
	require.NoError(t, err)
	require.NoError(t, instance.ConnectNodes(oneOutline, "Out", one, "Outlines"))
	_, many, err := instance.CreateNode(fillType)
	require.NoError(t, err)
	require.NoError(t, instance.ConnectNodes(manyOutlines, "Out", many, "Outlines"))

	union, unionID, err := instance.CreateNode(unionType)
	require.NoError(t, err)

	dropped, err := instance.Batch(func() error {
		require.NoError(t, instance.ConnectNodes(one, "Out", unionID, "Areas"))
		require.NoError(t, instance.ConnectNodes(many, "Out", unionID, "Areas"))
		return nil
	})
	require.NoError(t, err)
	require.Len(t, dropped, 1)
	assert.Equal(t, many, dropped[0].From)
	kept := union.Inputs()["Areas"].(nodes.ArrayValueInputPort).Value()
	require.Len(t, kept, 1)
	assert.Equal(t, instance.Node(one), kept[0].Node())
}
