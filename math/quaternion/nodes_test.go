package quaternion_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func rotationTo(from, to vector3.Float64) quaternion.Quaternion {
	return nodes.GetNodeOutputPort[quaternion.Quaternion](&nodes.Struct[quaternion.RotationToNode]{
		Data: quaternion.RotationToNode{
			From: nodes.ConstOutput[vector3.Float64]{Val: from},
			To:   nodes.ConstOutput[vector3.Float64]{Val: to},
		},
	}, "Out").Value()
}

func TestRotationToNodeAimsFromAtTo(t *testing.T) {
	tests := map[string]struct {
		from, to vector3.Float64
	}{
		"forward to up":       {vector3.Forward[float64](), vector3.Up[float64]()},
		"forward to diagonal": {vector3.Forward[float64](), vector3.New(1., 1., 1.)},
		"unnormalized inputs": {vector3.New(0., 0., 5.), vector3.New(3., 0., 0.)},
		"opposite":            {vector3.Forward[float64](), vector3.Backwards[float64]()},
		"same":                {vector3.Up[float64](), vector3.Up[float64]()},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := rotationTo(tc.from, tc.to).Rotate(tc.from.Normalized())
			want := tc.to.Normalized()
			assert.InDelta(t, want.X(), got.X(), 1e-9)
			assert.InDelta(t, want.Y(), got.Y(), 1e-9)
			assert.InDelta(t, want.Z(), got.Z(), 1e-9)
		})
	}
}

func TestMultiplyNodeAppliesFirstThenThen(t *testing.T) {
	tiltX := quaternion.FromTheta(math.Pi/2, vector3.Right[float64]())
	spinY := quaternion.FromTheta(math.Pi/2, vector3.Up[float64]())

	got := nodes.GetNodeOutputPort[quaternion.Quaternion](&nodes.Struct[quaternion.MultiplyNode]{
		Data: quaternion.MultiplyNode{
			First: nodes.ConstOutput[quaternion.Quaternion]{Val: tiltX},
			Then:  nodes.ConstOutput[quaternion.Quaternion]{Val: spinY},
		},
	}, "Out").Value().Rotate(vector3.Forward[float64]())

	want := spinY.Rotate(tiltX.Rotate(vector3.Forward[float64]()))
	assert.InDelta(t, want.X(), got.X(), 1e-9)
	assert.InDelta(t, want.Y(), got.Y(), 1e-9)
	assert.InDelta(t, want.Z(), got.Z(), 1e-9)

	// The two orders really differ, so the test is pinning something.
	other := tiltX.Rotate(spinY.Rotate(vector3.Forward[float64]()))
	assert.Greater(t, other.Distance(want), 0.5)
}

func TestMultiplyNodeFollowsEveryEntryOfAnArrayWithThen(t *testing.T) {
	spins := []quaternion.Quaternion{
		quaternion.FromTheta(0, vector3.Up[float64]()),
		quaternion.FromTheta(math.Pi/2, vector3.Up[float64]()),
	}
	tilt := quaternion.FromTheta(math.Pi/3, vector3.Right[float64]())

	got := nodes.GetNodeOutputPort[[]quaternion.Quaternion](&nodes.Struct[quaternion.MultiplyNode]{
		Data: quaternion.MultiplyNode{
			First: nodes.ConstOutput[[]quaternion.Quaternion]{Val: spins},
			Then:  nodes.ConstOutput[quaternion.Quaternion]{Val: tilt},
		},
	}, "Out").Value()

	assert.Len(t, got, 2)
	for i, spin := range spins {
		want := tilt.Rotate(spin.Rotate(vector3.Forward[float64]()))
		have := got[i].Rotate(vector3.Forward[float64]())
		assert.InDelta(t, want.X(), have.X(), 1e-9)
		assert.InDelta(t, want.Y(), have.Y(), 1e-9)
		assert.InDelta(t, want.Z(), have.Z(), 1e-9)
	}
}
