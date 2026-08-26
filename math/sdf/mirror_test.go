package sdf_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

// straddlingField has real, distinct content on both sides of every axis -
// not the usual "one limb built on the positive side only" case - so a
// fold-based mirror and a union-based mirror give genuinely different
// answers.
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

func TestMirrorNodeOneSidedFieldIdenticalEitherWay(t *testing.T) {
	// A single ear/limb built entirely on the positive side - the common
	// real case. Union true/false must agree here, or this would be a
	// behavior change for every existing build.
	ear := sdf.Sphere(vector3.New(0.3, 2., 1.3), 0.15)

	yes, no := true, false
	withUnion := mirrorXPort(t, ear, &yes)
	withoutUnion := mirrorXPort(t, ear, &no)

	pts := []vector3.Float64{
		vector3.New(0.3, 2., 1.3),
		vector3.New(-0.3, 2., 1.3),
		vector3.New(0., 2., 1.3),
		vector3.New(5., 5., 5.),
	}
	for _, p := range pts {
		assert.InDelta(t, withoutUnion(p), withUnion(p), 1e-9,
			"one-sided field: Union true/false must give identical results at %v", p)
	}
}

func TestMirrorNodeXYUnionPreservesAllFourQuadrants(t *testing.T) {
	// Four distinct spheres, one per quadrant of the XY plane - the
	// multi-axis analog of the single-axis straddling test.
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
