package sdf_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

// straddlingField has real, different content on both sides of every
// axis, so fold vs union mirroring gives different answers.
func straddlingField() sample.Vec3ToFloat {
	right := sdf.Sphere(vector3.New(2., 0., 0.), 0.5) // real content at x=+2
	left := sdf.Sphere(vector3.New(-5., 0., 0.), 0.5) // real, different content at x=-5
	return sdf.Union(right, left)
}

func mirrorXPort(t *testing.T, field sample.Vec3ToFloat, union *bool) sample.Vec3ToFloat {
	t.Helper()
	data := sdf.MirrorNode{
		Field: nodes.ConstOutput[sample.Vec3ToFloat]{Val: field},
	}
	if union != nil {
		data.Union = nodes.ConstOutput[bool]{Val: *union}
	}
	node := &nodes.Struct[sdf.MirrorNode]{Data: data}
	return nodes.GetNodeOutputPort[sample.Vec3ToFloat](node, "X").Value()
}

func TestMirrorNodeXDefaultUnionPreservesStraddlingContent(t *testing.T) {
	f := straddlingField()
	mirrored := mirrorXPort(t, f, nil) // Union left unset -> should default true

	atMinus5 := vector3.New(-5., 0., 0.)
	assert.InDelta(t, f(atMinus5), mirrored(atMinus5), 1e-9,
		"default (Union unset) should preserve the real content at x=-5, not fold it away")
}

func TestMirrorNodeXUnionFalseUsesCheapFold(t *testing.T) {
	f := straddlingField()
	no := false
	mirrored := mirrorXPort(t, f, &no)

	atMinus5 := vector3.New(-5., 0., 0.)
	folded := sdf.MirrorX(f)(atMinus5)
	assert.InDelta(t, folded, mirrored(atMinus5), 1e-9,
		"Union: false should reproduce the plain MirrorX fold exactly")
	assert.NotEqual(t, f(atMinus5), mirrored(atMinus5),
		"the fold is expected to lose the real content at x=-5 - that's the documented tradeoff")
}

func TestMirrorNodeUnionFalseIsAPureReflection(t *testing.T) {
	ear := sdf.Sphere(vector3.New(0.3, 2., 1.3), 0.15)

	no := false
	reflected := mirrorXPort(t, ear, &no)

	original := vector3.New(0.3, 2., 1.3)
	mirrored := vector3.New(-0.3, 2., 1.3)

	assert.NotEqual(t, ear(original), reflected(original),
		"Union: false is a pure reflection - it does not preserve the original")
	assert.InDelta(t, ear(original), reflected(mirrored), 1e-9,
		"the mirrored point should see the original ear's value")
}

func TestMirrorNodeUnionTrueIncludesBothOriginalAndReflection(t *testing.T) {
	ear := sdf.Sphere(vector3.New(0.3, 2., 1.3), 0.15)

	yes := true
	both := mirrorXPort(t, ear, &yes)

	original := vector3.New(0.3, 2., 1.3)
	mirrored := vector3.New(-0.3, 2., 1.3)

	assert.InDelta(t, ear(original), both(original), 1e-9,
		"Union: true (default) should still preserve the original")
	assert.InDelta(t, ear(original), both(mirrored), 1e-9,
		"Union: true (default) should also give the reflection")
}

func TestMirrorNodeXYUnionPreservesAllFourQuadrants(t *testing.T) {
	// Four distinct spheres, one per quadrant of the XY plane.
	q1 := sdf.Sphere(vector3.New(3., 3., 0.), 0.4)
	q2 := sdf.Sphere(vector3.New(-7., 3., 0.), 0.4)
	q3 := sdf.Sphere(vector3.New(-7., -9., 0.), 0.4)
	q4 := sdf.Sphere(vector3.New(3., -9., 0.), 0.4)
	f := sdf.Union(q1, sdf.Union(q2, sdf.Union(q3, q4)))

	data := sdf.MirrorNode{Field: nodes.ConstOutput[sample.Vec3ToFloat]{Val: f}}
	node := &nodes.Struct[sdf.MirrorNode]{Data: data} // Union defaults true
	mirroredXY := nodes.GetNodeOutputPort[sample.Vec3ToFloat](node, "XY").Value()

	for name, p := range map[string]vector3.Float64{
		"q1 (already positive/positive)": vector3.New(3., 3., 0.),
		"q2 (negative/positive)":         vector3.New(-7., 3., 0.),
		"q3 (negative/negative)":         vector3.New(-7., -9., 0.),
		"q4 (positive/negative)":         vector3.New(3., -9., 0.),
	} {
		t.Run(name, func(t *testing.T) {
			assert.InDelta(t, f(p), mirroredXY(p), 1e-9,
				"each quadrant's own distinct sphere should be preserved under the default union behavior")
		})
	}
}
