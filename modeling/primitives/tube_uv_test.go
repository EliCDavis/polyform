package primitives_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func uvsOnSurface(t *testing.T, mesh modeling.Mesh, keep func(pos, normal vector3.Float64) bool) []vector2.Float64 {
	t.Helper()

	require.True(t, mesh.HasFloat2Attribute(modeling.TexCoordAttribute))
	positions := mesh.Float3Attribute(modeling.PositionAttribute)
	normals := mesh.Float3Attribute(modeling.NormalAttribute)
	uvs := mesh.Float2Attribute(modeling.TexCoordAttribute)

	out := []vector2.Float64{}
	for i := 0; i < positions.Len(); i++ {
		if keep(positions.At(i), normals.At(i)) {
			out = append(out, uvs.At(i))
		}
	}
	require.NotEmpty(t, out, "no vertices matched")
	return out
}

func facesUp(_, n vector3.Float64) bool   { return n.Y() > 0.5 }
func facesDown(_, n vector3.Float64) bool { return n.Y() < -0.5 }
func facesOutward(p, n vector3.Float64) bool {
	return math.Abs(n.Y()) < 0.5 && vector3.New(p.X(), 0., p.Z()).Normalized().Dot(n) > 0.5
}
func facesInward(p, n vector3.Float64) bool {
	return math.Abs(n.Y()) < 0.5 && vector3.New(p.X(), 0., p.Z()).Normalized().Dot(n) < -0.5
}

func bounds(uvs []vector2.Float64) (min, max vector2.Float64) {
	min = vector2.New(math.Inf(1), math.Inf(1))
	max = vector2.New(math.Inf(-1), math.Inf(-1))
	for _, uv := range uvs {
		min = vector2.New(math.Min(min.X(), uv.X()), math.Min(min.Y(), uv.Y()))
		max = vector2.New(math.Max(max.X(), uv.X()), math.Max(max.Y(), uv.Y()))
	}
	return min, max
}

func TestTubeWithoutUVsHasNoTexCoordAttribute(t *testing.T) {
	mesh := primitives.Tube{Sides: 12, Height: 1, InnerRadius: 0.4, OuterRadius: 1}.ToMesh()
	assert.False(t, mesh.HasFloat2Attribute(modeling.TexCoordAttribute))
}

func TestTubeDefaultUVsCoverTheUnitSquare(t *testing.T) {
	mesh := primitives.Tube{
		Sides: 24, Height: 2, InnerRadius: 0.4, OuterRadius: 1,
		UVs: &primitives.TubeUVs{},
	}.ToMesh()

	for _, tc := range []struct {
		name string
		keep func(pos, normal vector3.Float64) bool
	}{
		{"outer wall", facesOutward},
		{"inner wall", facesInward},
	} {
		t.Run(tc.name, func(t *testing.T) {
			min, max := bounds(uvsOnSurface(t, mesh, tc.keep))
			assert.InDelta(t, 0, min.X(), 1e-9)
			assert.InDelta(t, 1, max.X(), 1e-9)
			assert.InDelta(t, 0, min.Y(), 1e-9)
			assert.InDelta(t, 1, max.Y(), 1e-9)
		})
	}

	for _, tc := range []struct {
		name string
		keep func(pos, normal vector3.Float64) bool
	}{
		{"top cap", facesUp},
		{"bottom cap", facesDown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			min, max := bounds(uvsOnSurface(t, mesh, tc.keep))
			assert.InDelta(t, 0, min.X(), 1e-9)
			assert.InDelta(t, 1, max.X(), 1e-9)
			assert.InDelta(t, 0, min.Y(), 1e-9)
			assert.InDelta(t, 1, max.Y(), 1e-9)
		})
	}
}

func TestTubeWallVRunsUp(t *testing.T) {
	mesh := primitives.Tube{
		Sides: 16, Height: 2, InnerRadius: 0.4, OuterRadius: 1,
		UVs: &primitives.TubeUVs{},
	}.ToMesh()

	positions := mesh.Float3Attribute(modeling.PositionAttribute)
	normals := mesh.Float3Attribute(modeling.NormalAttribute)
	uvs := mesh.Float2Attribute(modeling.TexCoordAttribute)

	for i := 0; i < positions.Len(); i++ {
		p, n := positions.At(i), normals.At(i)
		if !facesOutward(p, n) {
			continue
		}
		if p.Y() > 0 {
			assert.InDelta(t, 1, uvs.At(i).Y(), 1e-9, "top of the wall should be V=1")
		} else {
			assert.InDelta(t, 0, uvs.At(i).Y(), 1e-9, "bottom of the wall should be V=0")
		}
	}
}

func TestTubeCapUVsFormAnAnnulus(t *testing.T) {
	const inner, outer = 0.4, 1.0
	circle := primitives.CircleUVs{Center: vector2.New(0.5, 0.5), Radius: 0.5}
	mesh := primitives.Tube{
		Sides: 32, Height: 1, InnerRadius: inner, OuterRadius: outer,
		UVs: &primitives.TubeUVs{Top: &circle},
	}.ToMesh()

	positions := mesh.Float3Attribute(modeling.PositionAttribute)
	normals := mesh.Float3Attribute(modeling.NormalAttribute)
	uvs := mesh.Float2Attribute(modeling.TexCoordAttribute)

	for i := 0; i < positions.Len(); i++ {
		p, n := positions.At(i), normals.At(i)
		if !facesUp(p, n) {
			continue
		}
		geometric := vector3.New(p.X(), 0., p.Z()).Length()
		uvRadius := uvs.At(i).Sub(circle.Center).Length()
		assert.InDelta(t, (geometric/outer)*circle.Radius, uvRadius, 1e-9,
			"uv radius should track geometric radius")
	}
}

func TestTubeUVsAreIndependentPerSurface(t *testing.T) {
	corner := primitives.StripUVs{Start: vector2.New(0., 0.1), End: vector2.New(0.25, 0.1), Width: 0.2}
	mesh := primitives.Tube{
		Sides: 16, Height: 1, InnerRadius: 0.4, OuterRadius: 1,
		UVs: &primitives.TubeUVs{Inner: &corner},
	}.ToMesh()

	innerMin, innerMax := bounds(uvsOnSurface(t, mesh, facesInward))
	assert.InDelta(t, 0, innerMin.X(), 1e-9)
	assert.InDelta(t, 0.25, innerMax.X(), 1e-9)
	assert.InDelta(t, 0.2, innerMax.Y()-innerMin.Y(), 1e-9, "the strip's width spans the height")

	outerMin, outerMax := bounds(uvsOnSurface(t, mesh, facesOutward))
	assert.InDelta(t, 0, outerMin.X(), 1e-9)
	assert.InDelta(t, 1, outerMax.X(), 1e-9)
	assert.InDelta(t, 1, outerMax.Y()-outerMin.Y(), 1e-9)
}

func TestTubeUVsCoverEveryVertex(t *testing.T) {
	mesh := primitives.Tube{
		Sides: 9, Height: 1, InnerRadius: 0.3, OuterRadius: 0.8,
		UVs: &primitives.TubeUVs{},
	}.ToMesh()

	assert.Equal(t,
		mesh.Float3Attribute(modeling.PositionAttribute).Len(),
		mesh.Float2Attribute(modeling.TexCoordAttribute).Len())
}

func TestTubeZeroRadiusUVsStayFinite(t *testing.T) {
	mesh := primitives.Tube{
		Sides: 8, Height: 1, InnerRadius: 0, OuterRadius: 0,
		UVs: &primitives.TubeUVs{},
	}.ToMesh()

	uvs := mesh.Float2Attribute(modeling.TexCoordAttribute)
	for i := 0; i < uvs.Len(); i++ {
		uv := uvs.At(i)
		require.False(t, math.IsNaN(uv.X()) || math.IsNaN(uv.Y()), "vertex %d is NaN", i)
	}
}
