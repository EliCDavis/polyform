package trs_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformPointNodeScalesRotatesThenTranslates(t *testing.T) {
	transform := trs.New(
		vector3.New(10., 0., 0.),
		quaternion.FromTheta(math.Pi/2, vector3.Up[float64]()),
		vector3.New(2., 2., 2.),
	)

	got := nodes.GetNodeOutputPort[vector3.Float64](&nodes.Struct[trs.TransformPointNode]{
		Data: trs.TransformPointNode{
			TRS:   nodes.ConstOutput[trs.TRS]{Val: transform},
			Point: nodes.ConstOutput[vector3.Float64]{Val: vector3.New(1., 0., 0.)},
		},
	}, "Out").Value()

	// (1,0,0) scaled to (2,0,0), spun 90° about Y to (0,0,-2), moved to (10,0,-2).
	assert.InDelta(t, 10, got.X(), 1e-9)
	assert.InDelta(t, 0, got.Y(), 1e-9)
	assert.InDelta(t, -2, got.Z(), 1e-9)
}

func TestTransformPointNodeWithoutATransformPassesThePointThrough(t *testing.T) {
	got := nodes.GetNodeOutputPort[vector3.Float64](&nodes.Struct[trs.TransformPointNode]{
		Data: trs.TransformPointNode{
			Point: nodes.ConstOutput[vector3.Float64]{Val: vector3.New(1., 2., 3.)},
		},
	}, "Out").Value()
	assert.Equal(t, vector3.New(1., 2., 3.), got)
}

func TestTransformPointNodeMovesEveryPointOfAnArray(t *testing.T) {
	transform := trs.Position(vector3.New(0., 1., 0.))
	got := nodes.GetNodeOutputPort[[]vector3.Float64](&nodes.Struct[trs.TransformPointNode]{
		Data: trs.TransformPointNode{
			TRS:   nodes.ConstOutput[trs.TRS]{Val: transform},
			Point: nodes.ConstOutput[[]vector3.Float64]{Val: []vector3.Float64{vector3.Zero[float64](), vector3.One[float64]()}},
		},
	}, "Out").Value()
	assert.Equal(t, []vector3.Float64{vector3.New(0., 1., 0.), vector3.New(1., 2., 1.)}, got)
}

func TestTransformPointNodePairsUpTwoArrays(t *testing.T) {
	got := nodes.GetNodeOutputPort[[]vector3.Float64](&nodes.Struct[trs.TransformPointNode]{
		Data: trs.TransformPointNode{
			TRS: nodes.ConstOutput[[]trs.TRS]{Val: []trs.TRS{
				trs.Position(vector3.New(0., 1., 0.)),
				trs.Position(vector3.New(5., 0., 0.)),
			}},
			Point: nodes.ConstOutput[[]vector3.Float64]{Val: []vector3.Float64{
				vector3.Zero[float64](), vector3.One[float64](),
			}},
		},
	}, "Out").Value()
	assert.Equal(t, []vector3.Float64{vector3.New(0., 1., 0.), vector3.New(6., 1., 1.)}, got)
}

func TestLookAtNodeAimsLocalZAtTheTarget(t *testing.T) {
	got := nodes.GetNodeOutputPort[trs.TRS](&nodes.Struct[trs.LookAtNode]{
		Data: trs.LookAtNode{
			Position: nodes.ConstOutput[vector3.Float64]{Val: vector3.New(0.2, 0., 0.2)},
			Target:   nodes.ConstOutput[vector3.Float64]{Val: vector3.New(-0.1, 0., 0.2)},
		},
	}, "Out").Value()

	nose := got.Transform(vector3.New(0., 0., 1.))
	assert.InDelta(t, -0.8, nose.X(), 1e-9, "local +Z lands one unit toward -x")
	assert.InDelta(t, 0, nose.Y(), 1e-9)
	assert.InDelta(t, 0.2, nose.Z(), 1e-9)

	up := got.RotateDirection(vector3.Up[float64]())
	assert.InDelta(t, 1, up.Y(), 1e-9, "Y stays up for a level look")
}

func TestLookAtNodeScaleDefaultsToOneOnlyWhenUnwired(t *testing.T) {
	look := func(scale nodes.Output[vector3.Float64]) trs.TRS {
		return nodes.GetNodeOutputPort[trs.TRS](&nodes.Struct[trs.LookAtNode]{
			Data: trs.LookAtNode{
				Target: nodes.ConstOutput[vector3.Float64]{Val: vector3.New(0., 0., 1.)},
				Scale:  scale,
			},
		}, "Out").Value()
	}

	assert.Equal(t, vector3.One[float64](), look(nil).Scale())
	assert.Equal(t, vector3.Zero[float64](), look(nodes.ConstOutput[vector3.Float64]{Val: vector3.Zero[float64]()}).Scale())
}

func TestLookAtNodeAtItsOwnTargetIsIdentityRotation(t *testing.T) {
	p := vector3.New(1., 2., 3.)
	got := nodes.GetNodeOutputPort[trs.TRS](&nodes.Struct[trs.LookAtNode]{
		Data: trs.LookAtNode{
			Position: nodes.ConstOutput[vector3.Float64]{Val: p},
			Target:   nodes.ConstOutput[vector3.Float64]{Val: p},
		},
	}, "Out").Value()
	assert.Equal(t, p, got.Position())
	assert.Equal(t, vector3.New(0., 0., 1.), got.RotateDirection(vector3.New(0., 0., 1.)))
}

func TestPivotNodeLeavesThePivotPointFixed(t *testing.T) {
	pivot := vector3.New(0.055, 0.165, 0.078)
	rotation := quaternion.FromTheta(-1.2, vector3.Right[float64]())
	got := nodes.GetNodeOutputPort[trs.TRS](&nodes.Struct[trs.PivotNode]{
		Data: trs.PivotNode{
			Pivot:    nodes.ConstOutput[vector3.Float64]{Val: pivot},
			Rotation: nodes.ConstOutput[quaternion.Quaternion]{Val: rotation},
		},
	}, "Out").Value()

	moved := got.Transform(pivot)
	assert.InDelta(t, pivot.X(), moved.X(), 1e-9)
	assert.InDelta(t, pivot.Y(), moved.Y(), 1e-9)
	assert.InDelta(t, pivot.Z(), moved.Z(), 1e-9)

	// A point below the pivot swings the same way a plain rotation about the pivot would.
	paw := pivot.Add(vector3.New(0., -0.13, 0.))
	want := pivot.Add(rotation.Rotate(paw.Sub(pivot)))
	have := got.Transform(paw)
	assert.InDelta(t, want.Y(), have.Y(), 1e-9)
	assert.InDelta(t, want.Z(), have.Z(), 1e-9)
}

func TestPivotNodeScalesAboutThePivotToo(t *testing.T) {
	pivot := vector3.New(1., 2., 3.)
	got := nodes.GetNodeOutputPort[trs.TRS](&nodes.Struct[trs.PivotNode]{
		Data: trs.PivotNode{
			Pivot: nodes.ConstOutput[vector3.Float64]{Val: pivot},
			Scale: nodes.ConstOutput[vector3.Float64]{Val: vector3.New(2., 2., 2.)},
		},
	}, "Out").Value()
	assert.Equal(t, pivot, got.Transform(pivot))
	assert.Equal(t, vector3.New(3., 2., 3.), got.Transform(vector3.New(2., 2., 3.)), "one unit from the pivot becomes two")
}

func turned() trs.TRS {
	return trs.New(
		vector3.New(1., 2., 3.),
		quaternion.FromTheta(0.7, vector3.New(1., 2., 0.5)),
		vector3.New(1., 2., 0.5),
	)
}

func assertSameTransform(t *testing.T, want, got trs.TRS) {
	t.Helper()
	for _, p := range []vector3.Float64{vector3.New(1., 0., 0.), vector3.New(0., 1., 0.), vector3.New(0., 0., 1.), vector3.New(0.3, -2., 1.5)} {
		assert.InDelta(t, 0, want.Transform(p).Distance(got.Transform(p)), 1e-9, "point %v", p)
	}
}

func TestScaleLocalScalesAlongTheTransformsOwnAxes(t *testing.T) {
	start := turned()
	amount := vector3.New(3., 0.5, 2.)
	want, err := trs.FromMatrix(start.Multiply(trs.Scale(amount)))
	require.NoError(t, err)
	assertSameTransform(t, want, start.ScaleBy(amount))
}

func TestRotateLocalTurnsAboutTheTransformsOwnAxis(t *testing.T) {
	start := turned().SetScale(vector3.Fill(2.))
	axis := vector3.Forward[float64]()
	want, err := trs.FromMatrix(start.Multiply(trs.Rotation(quaternion.FromTheta(1.1, axis))))
	require.NoError(t, err)
	assertSameTransform(t, want, start.RotateAroundAxis(axis, 1.1))
}

func TestRotateLocalTurnsAnUnevenScaleWithIt(t *testing.T) {
	oval := trs.Scale(vector3.New(2., 1., 1.)).RotateAroundAxis(vector3.Forward[float64](), math.Pi/2)
	assert.InDelta(t, 0, oval.Transform(vector3.New(1., 0., 0.)).Distance(vector3.New(0., 2., 0.)), 1e-9)
}

func TestRotateLocalAboutNoAxisLeavesTheTransform(t *testing.T) {
	start := turned()
	assert.Equal(t, start, start.RotateAroundAxis(vector3.Zero[float64](), 1))
}

func TestScaleNodeLiftsOverAListOfTransforms(t *testing.T) {
	frames := []trs.TRS{trs.Identity(), trs.Position(vector3.New(1., 0., 0.))}
	out := nodes.GetNodeOutputPort[[]trs.TRS](&nodes.Struct[trs.ScaleNode]{Data: trs.ScaleNode{
		TRS:     nodes.ConstOutput[[]trs.TRS]{Val: frames},
		Uniform: nodes.ConstOutput[[]float64]{Val: []float64{2, 0.5}},
	}}, "Out").Value()

	require.Len(t, out, 2)
	assert.Equal(t, vector3.New(2., 2., 2.), out[0].Scale())
	assert.Equal(t, vector3.New(0.5, 0.5, 0.5), out[1].Scale())
	assert.Equal(t, vector3.New(1., 0., 0.), out[1].Position())
}

func TestRotateAroundAxisNodeRollsEachFrameByItsOwnAngle(t *testing.T) {
	frames := []trs.TRS{trs.Identity(), trs.Identity()}
	out := nodes.GetNodeOutputPort[[]trs.TRS](&nodes.Struct[trs.RotateAroundAxisNode]{Data: trs.RotateAroundAxisNode{
		TRS:   nodes.ConstOutput[[]trs.TRS]{Val: frames},
		Angle: nodes.ConstOutput[[]float64]{Val: []float64{0, math.Pi / 2}},
	}}, "Out").Value()

	require.Len(t, out, 2)
	assert.InDelta(t, 0, out[0].RotateDirection(vector3.Up[float64]()).Distance(vector3.Up[float64]()), 1e-9)
	assert.InDelta(t, 0, out[1].RotateDirection(vector3.Up[float64]()).Distance(vector3.Left[float64]()), 1e-9, "the default axis is +Z")
}
