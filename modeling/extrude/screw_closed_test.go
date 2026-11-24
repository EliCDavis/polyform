package extrude_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/csg"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func turned(t *testing.T, profile []vector3.Float64, revolutions float64) modeling.Mesh {
	t.Helper()
	return nodes.GetNodeOutputPort[modeling.Mesh](&nodes.Struct[extrude.ScrewNode]{
		Data: extrude.ScrewNode{
			Line:        nodes.ConstOutput[[]vector3.Float64]{Val: profile},
			Segments:    nodes.ConstOutput[int]{Val: 48},
			Revolutions: nodes.ConstOutput[float64]{Val: revolutions},
		},
	}, "Out").Value()
}

func TestFullTurnLatheFromAxisToAxisIsAClosedSolid(t *testing.T) {
	pawn := []vector3.Float64{
		vector3.New(0., 0., 0.),
		vector3.New(0.3, 0., 0.),
		vector3.New(0.3, 0.1, 0.),
		vector3.New(0.12, 0.3, 0.),
		vector3.New(0.15, 0.6, 0.),
		vector3.New(0., 0.7, 0.),
	}
	mesh := turned(t, pawn, 1)
	require.NoError(t, csg.CheckClosed(mesh))
	require.NoError(t, csg.CheckClosed(meshops.Weld(mesh, modeling.PositionAttribute, 0)))

	notch := primitives.Cube{Width: 0.5, Height: 0.2, Depth: 0.04}.Welded().Translate(vector3.New(0., 0.65, 0.))
	cut, err := csg.Subtract(mesh, notch)
	require.NoError(t, err)
	assert.NotZero(t, cut.PrimitiveCount())
}

func TestPartialTurnLatheStaysOpen(t *testing.T) {
	mesh := turned(t, []vector3.Float64{vector3.New(0., 0., 0.), vector3.New(0.3, 0., 0.), vector3.New(0., 0.5, 0.)}, 0.5)
	assert.Error(t, csg.CheckClosed(mesh))
}
