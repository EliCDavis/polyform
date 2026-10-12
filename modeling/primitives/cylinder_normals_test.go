package primitives_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestFrustumSideNormalsArePerpendicularToTheSlant(t *testing.T) {
	const sides = 12
	frustum := primitives.TruncatedCone{
		Sides:        sides,
		Height:       2,
		TopRadius:    0.25,
		BottomRadius: 1,
		NoTop:        true,
		NoBottom:     true,
	}.ToMesh()

	normals := requireUnitNormals(t, frustum)
	positions := frustum.Float3Attribute(modeling.PositionAttribute)

	for side := 0; side <= sides; side++ {
		top, bottom := positions.At(side*2), positions.At(side*2+1)
		slant := top.Sub(bottom)
		radial := vector3.New(bottom.X(), 0., bottom.Z())

		for _, i := range []int{side * 2, side*2 + 1} {
			assert.InDeltaf(t, 0, normals[i].Dot(slant), 1e-9, "normal %d is not perpendicular to the slant", i)
			assert.Greaterf(t, normals[i].Dot(radial), 0., "normal %d faces inward", i)
			assert.Greaterf(t, normals[i].Y(), 0., "normal %d of a narrowing frustum should tilt up", i)
		}
	}
}

func TestCylinderSideNormalsAreHorizontal(t *testing.T) {
	cylinder := primitives.Cylinder{Sides: 8, Height: 3, Radius: 0.5, NoTop: true, NoBottom: true}.ToMesh()

	for i, n := range requireUnitNormals(t, cylinder) {
		assert.InDeltaf(t, 0, n.Y(), 1e-9, "normal %d", i)
	}
}
