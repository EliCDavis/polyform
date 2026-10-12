package primitives_test

import (
	"testing"

	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestConeHeightSegmentsAddRings(t *testing.T) {
	const sides, segments, height = 8, 3, 2.
	cone := primitives.Cone{Height: height, Radius: 0.5, Sides: sides, HeightSegments: segments}.ToMesh()

	assert.Equal(t, sides*segments+1, cone.AttributeLength())
	assert.Equal(t, sides*(2*segments-1), cone.PrimitiveCount())

	normals := requireUnitNormals(t, cone)
	positions := cone.Float3Attribute(modeling.PositionAttribute)
	apex := vector3.New(0., height, 0.)
	for i := 0; i < positions.Len()-1; i++ {
		p := positions.At(i)
		assert.InDeltaf(t, 0, normals[i].Dot(apex.Sub(p)), 1e-9, "normal %d is not perpendicular to the slant", i)
		assert.Greaterf(t, normals[i].Dot(vector3.New(p.X(), 0., p.Z())), 0., "normal %d faces inward", i)
	}

	for i := 0; i < cone.PrimitiveCount(); i++ {
		tri := cone.Tri(i)
		a, b, c := tri.P1Vec3Attr(modeling.PositionAttribute), tri.P2Vec3Attr(modeling.PositionAttribute), tri.P3Vec3Attr(modeling.PositionAttribute)
		face := b.Sub(a).Cross(c.Sub(a))
		centre := a.Add(b).Add(c).Scale(1. / 3.)
		assert.Greaterf(t, face.Dot(vector3.New(centre.X(), 0., centre.Z())), 0., "triangle %d winds inward", i)
	}
}

func TestConeDefaultsToASingleSegment(t *testing.T) {
	plain := primitives.Cone{Height: 1, Radius: 0.5, Sides: 6}.ToMesh()
	one := primitives.Cone{Height: 1, Radius: 0.5, Sides: 6, HeightSegments: 1}.ToMesh()

	assert.Equal(t, 7, plain.AttributeLength())
	assert.Equal(t, 6, plain.PrimitiveCount())
	assert.Equal(t, iter.ReadFull(plain.Indices()), iter.ReadFull(one.Indices()))
}
