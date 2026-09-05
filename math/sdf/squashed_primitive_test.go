package sdf_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The documented way to build a fin, a blade or a leaf: a tapered capsule
// flattened on one axis. Every build so far has used a RoundCubeNode
// instead and shipped rectangular tabs, because a box's single Roundness
// applies to all three axes at once and a thin box therefore can't have a
// soft edge. This recipe can, and it only works because Transform now
// rescales the distance it returns - before that it produced a field with
// a gradient of 1/scale, which marched as speckle.
func TestSquashedTaperedCapsuleIsAValidBlade(t *testing.T) {
	fin := sdf.Transform(
		// Wide at the base, narrow at the tip.
		sdf.RoundedCone(
			vector3.New(0., 0., 0.),
			vector3.New(0., 0.6, 0.),
			0.25, 0.04,
		),
		trs.New(
			vector3.Zero[float64](),
			trs.Identity().Rotation(),
			vector3.New(0.18, 1., 1.), // flattened across the blade
		),
	)

	assert.LessOrEqual(t, worstGradient(fin), 1.0001,
		"a flattened blade must still be a valid distance field")

	// It really is thin across, and tall along the taper.
	across := surfaceAlong(fin, vector3.New(1., 0., 0.))
	along := surfaceAlong(fin, vector3.New(0., 1., 0.))
	require.Greater(t, along, across*3, "the blade should be far longer than it is thick")
	assert.Less(t, across, 0.06, "and genuinely thin")

	// The taper is real: it is narrower near the tip than the base.
	baseWidth := surfaceAlong(
		sdf.Transform(sdf.RoundedCone(vector3.New(0., 0., 0.), vector3.New(0., 0.6, 0.), 0.25, 0.04),
			trs.New(vector3.New(0., -0.5, 0.), trs.Identity().Rotation(), vector3.New(1., 1., 1.))),
		vector3.New(0., 1., 0.))
	assert.Positive(t, baseWidth)
}
