package sdf_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestVarryingThicknessLine(t *testing.T) {
	p0 := vector3.New(0., 0., 0.)
	p1 := vector3.New(0., 1., 0.)
	p2 := vector3.New(0., 2., 0.)

	field := sdf.VarryingThicknessLine([]sdf.LinePoint{
		{Point: p0, Radius: 0.2},
		{Point: p1, Radius: 0.1},
		{Point: p2, Radius: 0.05},
	})

	assert.InDelta(t, -0.1, field(p1), 1e-9)

	seg1 := sdf.RoundedCone(p0, p1, 0.2, 0.1)
	seg2 := sdf.RoundedCone(p1, p2, 0.1, 0.05)
	probes := []vector3.Float64{
		vector3.New(0.09, 1., 0.),
		vector3.New(0.11, 1., 0.),
		vector3.New(0.05, 0.7, 0.02),
		vector3.New(0.03, 1.4, -0.01),
	}
	for _, p := range probes {
		want := min(seg1(p), seg2(p))
		assert.InDelta(t, want, field(p), 1e-9)
	}
}
