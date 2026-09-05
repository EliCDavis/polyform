package primitives_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// requireUnitNormals checks the mesh carries one usable normal per vertex.
// Cone, Torus and Hemisphere shipped without any normal attribute at all,
// which surfaced far downstream as "mesh is required to have the vector3
// attribute: 'Normal'" the moment anything tried to rotate or light them.
func requireUnitNormals(t *testing.T, m modeling.Mesh) []vector3.Float64 {
	t.Helper()

	require.True(t, m.HasFloat3Attribute(modeling.NormalAttribute), "mesh has no Normal attribute")

	normals := m.Float3Attribute(modeling.NormalAttribute)
	positions := m.Float3Attribute(modeling.PositionAttribute)
	require.Equal(t, positions.Len(), normals.Len(), "expected one normal per vertex")

	out := make([]vector3.Float64, normals.Len())
	for i := 0; i < normals.Len(); i++ {
		n := normals.At(i)
		require.InDeltaf(t, 1, n.Length(), 1e-9, "normal %d is not unit length: %v", i, n)
		out[i] = n
	}
	return out
}

func TestConeNormals(t *testing.T) {
	const height, radius = 2., 0.5
	cone := primitives.Cone{Height: height, Radius: radius, Sides: 16}.ToMesh()

	normals := requireUnitNormals(t, cone)
	positions := cone.Float3Attribute(modeling.PositionAttribute)
	apex := vector3.New(0., height, 0.)

	// Every base vertex's normal must lie flat against the cone's slanted
	// side - perpendicular to the line running up to the apex - and face
	// outward rather than into the cone.
	for i := 0; i < positions.Len()-1; i++ {
		p := positions.At(i)
		slant := apex.Sub(p).Normalized()

		assert.InDeltaf(t, 0, normals[i].Dot(slant), 1e-9,
			"normal %d is not perpendicular to the cone's slant", i)
		assert.Greaterf(t, normals[i].Dot(vector3.New(p.X(), 0., p.Z())), 0.,
			"normal %d faces inward", i)
	}
}

// TestConeNormalsTrackShape guards the part that makes this a real normal
// rather than a filler value: a squat cone's sides face mostly up, a tall
// one's mostly outward.
func TestConeNormalsTrackShape(t *testing.T) {
	squat := primitives.Cone{Height: 0.01, Radius: 1, Sides: 8}.ToMesh()
	tall := primitives.Cone{Height: 100, Radius: 1, Sides: 8}.ToMesh()

	squatY := squat.Float3Attribute(modeling.NormalAttribute).At(0).Y()
	tallY := tall.Float3Attribute(modeling.NormalAttribute).At(0).Y()

	assert.Greater(t, squatY, 0.99, "a nearly flat cone's sides should face nearly straight up")
	assert.Less(t, tallY, 0.02, "a very tall cone's sides should face nearly straight out")
}

func TestTorusNormals(t *testing.T) {
	const major, minor = 1., 0.25
	torus := primitives.Torus{
		MajorRadius:     major,
		MinorRadius:     minor,
		MajorResolution: 12,
		MinorResolution: 8,
	}.ToMesh()

	normals := requireUnitNormals(t, torus)
	positions := torus.Float3Attribute(modeling.PositionAttribute)

	// A torus normal points straight out from the major circle: the vertex
	// sits exactly MinorRadius along it from the ring's centerline.
	for i := 0; i < positions.Len(); i++ {
		p := positions.At(i)
		ringCenter := vector3.New(p.X(), 0., p.Z()).Normalized().Scale(major)
		want := p.Sub(ringCenter).Normalized()

		assert.InDeltaf(t, want.X(), normals[i].X(), 1e-9, "normal %d x", i)
		assert.InDeltaf(t, want.Y(), normals[i].Y(), 1e-9, "normal %d y", i)
		assert.InDeltaf(t, want.Z(), normals[i].Z(), 1e-9, "normal %d z", i)
	}
}

func TestHemisphereNormalsSphericalCase(t *testing.T) {
	const radius = 2.
	// Radius == Height, so the surface is a true hemisphere and every
	// surface normal is just its normalized position.
	hemi := primitives.Hemisphere{Radius: radius, Height: radius, Capped: true}.UV(8, 12)

	normals := requireUnitNormals(t, hemi)
	positions := hemi.Float3Attribute(modeling.PositionAttribute)

	// Index 0 is the flat cap's center, which faces down instead.
	assert.InDelta(t, -1, normals[0].Y(), 1e-9, "bottom cap center should face down")

	for i := 1; i < positions.Len(); i++ {
		want := positions.At(i).Normalized()
		assert.InDeltaf(t, want.X(), normals[i].X(), 1e-9, "normal %d x", i)
		assert.InDeltaf(t, want.Y(), normals[i].Y(), 1e-9, "normal %d y", i)
		assert.InDeltaf(t, want.Z(), normals[i].Z(), 1e-9, "normal %d z", i)
	}
}

// TestHemisphereNormalsSquashed covers the case a normalized position would
// get wrong: with Height != Radius the surface is an ellipsoid, so the
// normal leans differently than the position vector does.
func TestHemisphereNormalsSquashed(t *testing.T) {
	hemi := primitives.Hemisphere{Radius: 4, Height: 1, Capped: true}.UV(8, 12)
	normals := requireUnitNormals(t, hemi)
	positions := hemi.Float3Attribute(modeling.PositionAttribute)

	differs := false
	for i := 1; i < positions.Len(); i++ {
		p := positions.At(i)
		if p.Length() == 0 {
			continue
		}
		if math.Abs(p.Normalized().Dot(normals[i])-1) > 1e-6 {
			differs = true
		}
		// However it leans, a squashed dome still never faces inward.
		assert.GreaterOrEqualf(t, normals[i].Y(), -1e-9, "normal %d dips below the dome", i)
	}
	assert.True(t, differs, "a squashed dome's normals should differ from its normalized positions")
}

// TestRotatingNormalsOnPrimitives reproduces the exact failure from a real
// build: rotating a mesh's positions and then its normals - the correct
// thing to do - blew up on a cone with "mesh is required to have the
// vector3 attribute: 'Normal'".
func TestRotatingNormalsOnPrimitives(t *testing.T) {
	quarterTurn := quaternion.FromEulerAngle(vector3.New(math.Pi/2, 0., 0.))

	meshes := map[string]modeling.Mesh{
		"cone":       primitives.Cone{Height: 1, Radius: 0.5, Sides: 8}.ToMesh(),
		"torus":      primitives.Torus{MajorRadius: 1, MinorRadius: 0.25, MajorResolution: 8, MinorResolution: 6}.ToMesh(),
		"hemisphere": primitives.Hemisphere{Radius: 1, Height: 1, Capped: true}.UV(6, 8),
	}

	for name, m := range meshes {
		t.Run(name, func(t *testing.T) {
			require.NotPanics(t, func() {
				m = meshops.RotateAttribute3D(m, modeling.PositionAttribute, quarterTurn)
				m = meshops.RotateAttribute3D(m, modeling.NormalAttribute, quarterTurn)
			})
		})
	}
}
