package geometry_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntersect(t *testing.T) {
	l1 := geometry.NewLine2D(vector2.New(0., 0.), vector2.New(1., 1.))
	l2 := geometry.NewLine2D(vector2.New(0., 1.), vector2.New(1., 0.))
	if l1.Intersects(l2) == false {
		t.Error("Lines should have interesected")
	}
}

func TestDoesntIntersect(t *testing.T) {
	l1 := geometry.NewLine2D(vector2.New(0., 0.), vector2.New(0., .4))
	l2 := geometry.NewLine2D(vector2.New(0., 1.), vector2.New(1., .0))
	if l1.Intersects(l2) == true {
		t.Error("Lines should have not interesected")
	}
}

func TestTrimLines2DCutsBothEndsAndDropsShortLines(t *testing.T) {
	node := nodes.Struct[geometry.TrimLines2DNode]{Data: geometry.TrimLines2DNode{
		Lines:  nodes.ConstOutput[[]geometry.Line2D]{Val: []geometry.Line2D{geometry.NewLine2D(vector2.New(0., 0.), vector2.New(4., 0.)), geometry.NewLine2D(vector2.New(0., 0.), vector2.New(1., 0.))}},
		Amount: nodes.ConstOutput[float64]{Val: 1},
	}}

	lines := nodes.GetNodeOutputPort[[]geometry.Line2D](&node, "Trimmed").Value()

	require.Len(t, lines, 1)
	assert.Equal(t, vector2.New(1., 0.), lines[0].GetStartPoint())
	assert.Equal(t, vector2.New(3., 0.), lines[0].GetEndPoint())
}
